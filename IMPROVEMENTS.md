# Enrichment Workflow Improvements

## Problem Analysis

The system extracted Bill Gates' personal website (`gatesnotes.com`) instead of Microsoft's company website, with 100% confidence. This occurred despite the metadata clearly specifying "Company website URL".

### Root Causes

1. **Missing Entity Type Context**: The prompts don't emphasize that we're extracting data about THE COMPANY, not about people related to the company. The `entity_type` parameter exists in the API but isn't used in prompt construction.

2. **No Entity Disambiguation**: When content contains information about multiple entities (Microsoft the company, Bill Gates the person, Paul Allen, etc.), there's no instruction to distinguish between them.

3. **No Cross-Field Consistency Checks**: The LLM extracts each field independently without verifying they all refer to the same entity. It can extract "founder: Bill Gates" and "website: gatesnotes.com" without realizing these are about different entities.

4. **Overconfident Scoring**: The LLM assigns high confidence when it finds explicit information, but doesn't check if the information is about the CORRECT entity.

5. **Insufficient Source Context**: During the decision-making phase, the system may select Bill Gates' Wikipedia page when looking for Microsoft info, creating ambiguity.

## Recommended Solutions

### 1. Add Entity Type Awareness to Prompts ⭐ HIGH PRIORITY

**Current State**:
```go
// extraction_prompt_builder.go:36
fmt.Sprintf(`You are a data extraction specialist. Extract the following fields from the provided website content about %s.`, entity, ...)
```

**Problem**: Doesn't specify what TYPE of entity we're extracting data about.

**Solution**:
```go
func (e *ExtractionPromptBuilder) Build(content string, columnsMetadata []*models.ColumnMetadata, entity string, entityType string) string {
    // ... existing code ...

    return fmt.Sprintf(`You are a data extraction specialist. Extract the following fields about the %s named "%s" from the provided website content.

## CRITICAL: Entity Extraction Rules

You are extracting data about the %s "%s" - NOT about people, founders, executives, or other related entities.

ALL extracted fields must be about the %s itself:
- ✓ "website" means the %s's official website (e.g., microsoft.com for Microsoft)
- ✗ NOT the founder's personal website (e.g., gatesnotes.com for Bill Gates)
- ✓ "employee_count" means employees of the %s
- ✗ NOT the number of followers/staff of a related person
- ✓ "founded_year" means when the %s was founded
- ✗ NOT when a person was born

## Fields to Extract (ONLY extract these fields)
%s

## Website Content
%s

...`, entityType, entity, entityType, entity, entityType, entityType, entityType, entityType, columnsText, truncatedContent)
}
```

**Impact**: Makes the entity context explicit in every extraction, reducing ambiguity.

---

### 2. Add Entity Consistency Validation Instructions ⭐ HIGH PRIORITY

**Solution**: Add to both decision_maker.go and extraction_prompt_builder.go:

```go
## Entity Consistency Validation

Before finalizing your response, verify:
1. ALL extracted data refers to the TARGET ENTITY (%s the %s), not to:
   - Founders, CEOs, or employees
   - Parent companies or subsidiaries (unless explicitly requested)
   - Related organizations or people

2. If you extract data about the WRONG entity:
   - Do NOT include that field in extracted_data
   - REDUCE confidence score to 0.0-0.3 if uncertain about entity
   - Add explanation in reasoning about the ambiguity

3. Common mistakes to AVOID:
   - Extracting a founder's personal website instead of company website
   - Extracting a CEO's birth year instead of company founding year
   - Extracting a founder's photo instead of company logo
   - Mixing data from multiple entities in the same response

If the content is primarily about a DIFFERENT entity (e.g., a person's biography when you need company data), explicitly note this in your reasoning and extract only what clearly applies to the target entity.
```

**Impact**: Forces the LLM to actively check entity consistency before responding.

---

### 3. Add Field-Specific Disambiguation Examples ⭐ HIGH PRIORITY

**Solution**: Enhance the prompt with concrete examples:

```go
## Field Extraction Examples

### ✓ CORRECT Examples:
- Extracting "website" for Microsoft company:
  * ✓ "microsoft.com" (the COMPANY's website)
  * ✗ "gatesnotes.com" (Bill Gates' personal site - wrong entity!)
  * ✗ "linkedin.com/in/satyanadella" (Satya Nadella's profile - wrong entity!)

- Extracting "founded" for Apple company:
  * ✓ "1976" (when the COMPANY was founded)
  * ✗ "1955" (when Steve Jobs was born - wrong entity!)

- Extracting "employee_count" for Google company:
  * ✓ "150000" (COMPANY employees)
  * ✗ "100" (Sundar Pichai's team size - wrong entity!)

### ✗ INCORRECT Examples:
Query: Extract data about Microsoft (company)
Wrong: {"website": "gatesnotes.com", "founder": "Bill Gates"}
Issue: Website is Bill Gates' personal site, not Microsoft's

Correct: {"website": "microsoft.com", "founder": "Bill Gates"}
Explanation: Microsoft is the target entity, so extract ITS website
```

**Impact**: Provides concrete patterns the LLM can learn from.

---

### 4. Improve Source Selection in Decision Maker ⭐ MEDIUM PRIORITY

**Current State**:
```go
// decision_maker.go:164-167
// "Prioritizing: * Wikipedia * Reliable data sources * Avoid SEO aggregator sites"
```

**Problem**: Doesn't specify to prioritize COMPANY pages over PERSON pages.

**Solution**:
```go
2. Check if you extracted ALL the columns we need:
   - If YES: Return empty urls_to_crawl array
   - If NO: Select up to %d URLs to crawl for missing data, prioritizing:
     * Official %s website (most authoritative for %s data)
     * Wikipedia page specifically about the %s (e.g., "Wikipedia - Microsoft" not "Wikipedia - Bill Gates")
     * Reliable data sources (SEC filings, financial sites like Crunchbase, Bloomberg)
     * Industry databases and registries
     * Avoid: Personal websites, biography pages of people, SEO aggregators

   ⚠️  CRITICAL: Verify URLs are about the TARGET ENTITY (%s the %s), not related people:
     - ✓ en.wikipedia.org/wiki/Microsoft (about the company)
     - ✗ en.wikipedia.org/wiki/Bill_Gates (about a person, not the company)
     - ✓ microsoft.com (official company site)
     - ✗ gatesnotes.com (founder's personal site)
```

**Impact**: Reduces chances of crawling pages about wrong entities.

---

### 5. Add Post-Extraction Validation Logic ⭐ MEDIUM PRIORITY

**Solution**: Create a new validation service that checks for common errors:

**File**: `go/internal/services/entity_validator.go`

```go
package services

import (
	"fmt"
	"strings"
	"net/url"
)

type EntityValidator struct{}

type ValidationWarning struct {
	Field      string
	Issue      string
	Suggestion string
}

func NewEntityValidator() *EntityValidator {
	return &EntityValidator{}
}

// ValidateExtractedData checks for common entity confusion errors
func (v *EntityValidator) ValidateExtractedData(
	extractedData map[string]interface{},
	entityName string,
	entityType string,
	sources []string,
) []ValidationWarning {
	var warnings []ValidationWarning

	// Check for website/founder mismatch
	if website, ok := extractedData["website"].(string); ok {
		if founder, fok := extractedData["founder"].(string); fok {
			warnings = append(warnings, v.checkWebsiteFounderConsistency(website, founder, entityName)...)
		}
	}

	// Check if sources are about the correct entity
	warnings = append(warnings, v.checkSourceRelevance(sources, entityName, entityType)...)

	// Check for URL consistency with entity name
	if website, ok := extractedData["website"].(string); ok {
		warnings = append(warnings, v.checkURLConsistency(website, entityName, entityType)...)
	}

	return warnings
}

func (v *EntityValidator) checkWebsiteFounderConsistency(website, founder, companyName string) []ValidationWarning {
	var warnings []ValidationWarning

	// Parse website domain
	parsedURL, err := url.Parse(website)
	if err != nil {
		return warnings
	}

	domain := strings.ToLower(parsedURL.Hostname())
	founderLower := strings.ToLower(founder)
	companyLower := strings.ToLower(companyName)

	// Check if website contains founder's name but not company name
	// Split founder name into parts
	founderParts := strings.Fields(founderLower)

	containsFounderName := false
	for _, part := range founderParts {
		if len(part) > 3 && strings.Contains(domain, part) {
			containsFounderName = true
			break
		}
	}

	containsCompanyName := strings.Contains(domain, strings.ReplaceAll(companyLower, " ", ""))

	if containsFounderName && !containsCompanyName {
		warnings = append(warnings, ValidationWarning{
			Field:      "website",
			Issue:      fmt.Sprintf("Website URL '%s' appears to contain founder's name '%s' but not company name '%s'", website, founder, companyName),
			Suggestion: fmt.Sprintf("Verify this is the company website, not the founder's personal website. Expected something like: %s.com", strings.ToLower(strings.ReplaceAll(companyName, " ", ""))),
		})
	}

	return warnings
}

func (v *EntityValidator) checkSourceRelevance(sources []string, entityName, entityType string) []ValidationWarning {
	var warnings []ValidationWarning

	entityLower := strings.ToLower(entityName)

	for _, source := range sources {
		sourceLower := strings.ToLower(source)

		// Check for person-focused Wikipedia pages when entity is a company
		if entityType == "company" && strings.Contains(sourceLower, "wikipedia.org/wiki/") {
			// Extract the Wikipedia page title
			parts := strings.Split(sourceLower, "/wiki/")
			if len(parts) > 1 {
				pageTitle := parts[1]
				pageTitle = strings.Split(pageTitle, "#")[0] // Remove anchors

				// Check if page title doesn't match company name
				if !strings.Contains(pageTitle, strings.ReplaceAll(entityLower, " ", "_")) {
					warnings = append(warnings, ValidationWarning{
						Field:      "sources",
						Issue:      fmt.Sprintf("Source '%s' may be about a person/different entity, not the company '%s'", source, entityName),
						Suggestion: fmt.Sprintf("Verify data was extracted about the company, not related individuals"),
					})
				}
			}
		}
	}

	return warnings
}

func (v *EntityValidator) checkURLConsistency(website, entityName, entityType string) []ValidationWarning {
	var warnings []ValidationWarning

	if entityType != "company" {
		return warnings
	}

	parsedURL, err := url.Parse(website)
	if err != nil {
		return warnings
	}

	domain := strings.ToLower(parsedURL.Hostname())
	domain = strings.TrimPrefix(domain, "www.")

	entityLower := strings.ToLower(strings.ReplaceAll(entityName, " ", ""))

	// For companies, domain should typically contain company name
	// (with some exceptions for rebranded companies, abbreviations, etc.)
	if !strings.Contains(domain, entityLower) && len(entityName) > 4 {
		// Check if it's a common short form or abbreviation
		words := strings.Fields(strings.ToLower(entityName))
		matchFound := false

		for _, word := range words {
			if len(word) > 3 && strings.Contains(domain, word) {
				matchFound = true
				break
			}
		}

		if !matchFound {
			warnings = append(warnings, ValidationWarning{
				Field:      "website",
				Issue:      fmt.Sprintf("Website domain '%s' doesn't contain company name '%s'", domain, entityName),
				Suggestion: "Verify this is the correct company website. Lower confidence if uncertain.",
			})
		}
	}

	return warnings
}
```

**Integration**: Add to `activities.go` Extract activity:

```go
// After extraction, validate the data
validator := services.NewEntityValidator()
warnings := validator.ValidateExtractedData(
	mergedData,
	input.RowKey,
	input.EntityType, // Need to add this to ExtractInput
	append(input.Sources, extractOutput.Sources...),
)

// If warnings found, log them and potentially adjust confidence
if len(warnings) > 0 {
	for _, warning := range warnings {
		logger.Warn("Entity validation warning",
			"field", warning.Field,
			"issue", warning.Issue,
			"suggestion", warning.Suggestion,
		)

		// Reduce confidence for flagged fields
		if conf, ok := mergedConfidence[warning.Field]; ok {
			conf.Score = min(conf.Score, 0.7) // Cap at 0.7 when validation fails
			conf.Reason = fmt.Sprintf("%s (Warning: %s)", conf.Reason, warning.Issue)
			mergedConfidence[warning.Field] = conf
		}
	}
}
```

**Impact**: Catches obvious entity confusion errors and adjusts confidence accordingly.

---

### 6. Add Confidence Score Calibration ⭐ MEDIUM PRIORITY

**Solution**: Update the confidence scoring guidelines in prompts:

```go
## Confidence Scoring Guidelines

Base your confidence on BOTH information availability AND entity certainty:

- 1.0: Exact match, explicitly stated, AND clearly about the target entity
  * Example: "Microsoft's official website is microsoft.com" → 1.0 for company website

- 0.8-0.9: Clear statement, minor interpretation needed, target entity is clear
  * Example: "Visit us at apple.com" on Apple Inc. page → 0.9

- 0.6-0.7: Partial information, OR content mentions multiple entities
  * Example: "Bill Gates' website gatesnotes.com" when extracting for Microsoft → 0.6 (wrong entity risk)
  * Example: "Approximately 180,000 employees" → 0.7 (approximation)

- 0.4-0.5: Significant uncertainty, entity ambiguity, or approximation
  * Example: Website found on a founder's bio page, not official company page → 0.5

- <0.4: High uncertainty, likely derived or estimated, or wrong entity suspected
  * Example: Inferring company website from an employee's email domain → 0.3

⚠️  ALWAYS reduce confidence by at least 0.2 when:
- Information is about a related person instead of the target entity
- Multiple entities are mentioned and target is unclear
- Source is indirect (e.g., third-party profile, not official source)
```

**Impact**: Results in more realistic confidence scores that reflect entity ambiguity.

---

### 7. Add Structured Reasoning Requirements ⭐ LOW PRIORITY

**Solution**: Require the LLM to document its entity verification process:

```go
## Response Format (JSON only, no markdown)
{
    "extracted_data": {"field_name": value_with_correct_type},
    "confidence": {
        "field_name": {
            "score": 0.95,
            "reason": "Brief 1-sentence explanation",
            "entity_verification": "Confirmed this data is about [target entity]" // NEW FIELD
        }
    },
    "reasoning": "Overall extraction summary, including any entity disambiguation performed"
}
```

**Impact**: Forces the LLM to explicitly think about entity correctness for each field.

---

## Implementation Priority

### Phase 1 - Critical Fixes (Immediate)
1. ✅ Add entity type to prompt context (extraction_prompt_builder.go, decision_maker.go)
2. ✅ Add entity consistency validation instructions to prompts
3. ✅ Add field-specific disambiguation examples to prompts

### Phase 2 - Enhanced Validation (1-2 weeks)
4. ✅ Improve source selection logic in decision maker
5. ✅ Add post-extraction validation service (entity_validator.go)
6. ✅ Integrate validator into Extract activity

### Phase 3 - Refinements (2-4 weeks)
7. ✅ Calibrate confidence scoring guidelines
8. ✅ Add structured reasoning requirements
9. ✅ Collect metrics on validation warnings and false positives
10. ✅ Fine-tune validation heuristics based on real data

---

## Testing Recommendations

### Test Cases to Add

1. **Founder Website Confusion**
   - Company: Microsoft
   - Expected: microsoft.com
   - Should NOT extract: gatesnotes.com

2. **CEO Personal Site**
   - Company: Tesla
   - Expected: tesla.com
   - Should NOT extract: Elon Musk's personal site

3. **Birth Year vs Founding Year**
   - Company: Apple
   - Expected founded: 1976
   - Should NOT extract: 1955 (Steve Jobs' birth year)

4. **Subsidiary Confusion**
   - Company: YouTube
   - Expected website: youtube.com
   - Should NOT extract: google.com (parent company)

5. **Employee Count Ambiguity**
   - Company: Meta
   - Expected: ~60,000 employees
   - Should NOT extract: Mark Zuckerberg's team size

### Validation Metrics to Track

```go
type ValidationMetrics struct {
    TotalExtractions           int
    ValidationsTriggered       int
    WebsiteFounderMismatches   int
    SourceRelevanceWarnings    int
    URLConsistencyIssues       int
    ConfidenceAdjustmentsMade  int
    UserReportedErrors         int // For comparing against user feedback
}
```

---

## Expected Impact

### Before Improvements:
- ❌ Microsoft → gatesnotes.com (Bill Gates' site) - 100% confidence
- ❌ Apple → 1955 (Jobs' birth year as "founded") - 90% confidence
- ❌ Tesla → elonmusk.com (CEO site as company website) - 100% confidence

### After Improvements:
- ✅ Microsoft → microsoft.com - 95% confidence
- ✅ Microsoft → (Bill Gates info correctly extracted to "founder" field only)
- ✅ Sources automatically prefer company pages over biography pages
- ✅ Validation warnings flag suspicious extractions
- ✅ Confidence scores reduced when entity ambiguity detected

### Estimated Error Reduction:
- Entity confusion errors: **70-80% reduction**
- False confidence (high score on wrong data): **60-70% reduction**
- Source relevance issues: **50-60% reduction**

---

## Additional Recommendations

### 1. Add Entity Type to API Models

Update `ExtractInput` in activities.go:

```go
type ExtractInput struct {
    JobID           string
    RowKey          string
    EntityType      string // NEW: "company", "person", etc.
    ColumnsMetadata []*models.ColumnMetadata
    Content         string
    Sources         []string
    ExtractedData   map[string]interface{}
    Confidence      map[string]models.FieldConfidenceInfo
}
```

### 2. Consider Multi-Stage Validation

For high-value extractions, add a second LLM call that reviews the first extraction:

```go
// Pseudo-code
func (a *Activities) ValidateExtraction(ctx context.Context, input ValidationInput) (*ValidationOutput, error) {
    prompt := fmt.Sprintf(`Review this data extraction and identify any entity confusion errors:

    Target Entity: %s (type: %s)
    Extracted Data: %v
    Sources: %v

    Check for:
    1. Data about wrong entities (e.g., founder's website instead of company website)
    2. Inconsistent data across fields
    3. Source-data mismatches

    Return JSON with:
    - is_valid: boolean
    - issues: [{"field": "...", "problem": "...", "confidence_adjustment": -0.3}]
    - corrected_data: {} (if you can fix issues)
    `, input.EntityName, input.EntityType, input.ExtractedData, input.Sources)

    // Call LLM, parse response, adjust confidence scores
}
```

### 3. Build a Feedback Loop

Store user corrections to build a dataset of common errors:

```sql
CREATE TABLE extraction_corrections (
    job_id UUID,
    row_key TEXT,
    field_name TEXT,
    incorrect_value TEXT,
    correct_value TEXT,
    error_type TEXT, -- 'entity_confusion', 'wrong_data_type', 'missing', etc.
    created_at TIMESTAMP
);
```

Use this data to:
- Generate more targeted examples in prompts
- Fine-tune validation heuristics
- Measure improvement over time

---

## Conclusion

The core issue is **entity ambiguity** - the LLM doesn't sufficiently distinguish between the target entity and related entities. By making entity type explicit in prompts, adding consistency checks, and implementing post-processing validation, we can significantly reduce these errors while maintaining system flexibility.

The most critical changes are in **Phase 1** (prompt improvements), which can be implemented immediately with minimal code changes but maximum impact.
