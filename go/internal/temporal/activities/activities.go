package activities

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/blagoySimandov/ampledata/go/internal/models"
	"github.com/blagoySimandov/ampledata/go/internal/services"
	"github.com/blagoySimandov/ampledata/go/internal/state"
)

// Activities contains all the dependencies needed for enrichment activities
type Activities struct {
	stateManager       *state.StateManager
	webSearcher        services.WebSearcher
	decisionMaker      services.DecisionMaker
	crawler            services.Crawler
	contentExtractor   services.ContentExtractor
	patternGenerator   services.QueryPatternGenerator
}

// NewActivities creates a new Activities instance with all required dependencies
func NewActivities(
	stateManager *state.StateManager,
	webSearcher services.WebSearcher,
	decisionMaker services.DecisionMaker,
	crawler services.Crawler,
	contentExtractor services.ContentExtractor,
	patternGenerator services.QueryPatternGenerator,
) *Activities {
	return &Activities{
		stateManager:     stateManager,
		webSearcher:      webSearcher,
		decisionMaker:    decisionMaker,
		crawler:          crawler,
		contentExtractor: contentExtractor,
		patternGenerator: patternGenerator,
	}
}

// Activity Input/Output structures

type GeneratePatternsInput struct {
	JobID           string
	ColumnsMetadata []*models.ColumnMetadata
}

type GeneratePatternsOutput struct {
	Patterns []string
}

type SerpFetchInput struct {
	JobID           string
	RowKey          string
	ColumnsMetadata []*models.ColumnMetadata
	QueryPatterns   []string
}

type SerpFetchOutput struct {
	SerpData *models.SerpData
}

type DecisionInput struct {
	JobID           string
	RowKey          string
	SerpData        *models.SerpData
	ColumnsMetadata []*models.ColumnMetadata
}

type DecisionOutput struct {
	Decision *models.Decision
}

type CrawlInput struct {
	JobID           string
	RowKey          string
	Decision        *models.Decision
	ColumnsMetadata []*models.ColumnMetadata
}

type CrawlOutput struct {
	CrawlResults *models.CrawlResults
}

type ExtractInput struct {
	JobID           string
	RowKey          string
	Decision        *models.Decision
	CrawlResults    *models.CrawlResults
	ColumnsMetadata []*models.ColumnMetadata
}

type ExtractOutput struct {
	ExtractedData map[string]interface{}
	Confidence    map[string]*models.FieldConfidenceInfo
}

type StateUpdateInput struct {
	JobID  string
	RowKey string
	Stage  models.RowStage
	Data   map[string]interface{}
}

// FeedbackAnalysisInput is used for analyzing enrichment results to determine if feedback is needed
type FeedbackAnalysisInput struct {
	JobID           string
	RowKey          string
	ExtractedData   map[string]interface{}
	Confidence      map[string]*models.FieldConfidenceInfo
	ColumnsMetadata []*models.ColumnMetadata
}

// FeedbackAnalysisOutput contains feedback information
type FeedbackAnalysisOutput struct {
	NeedsFeedback      bool
	LowConfidenceColumns []string
	MissingColumns     []string
	AverageConfidence  float64
}

// Activities Implementation

// GeneratePatterns generates query patterns for the enrichment job
func (a *Activities) GeneratePatterns(ctx context.Context, input GeneratePatternsInput) (*GeneratePatternsOutput, error) {
	log.Printf("[Activity] Generating patterns for job %s", input.JobID)

	patterns, err := a.patternGenerator.GeneratePatterns(ctx, input.ColumnsMetadata)
	if err != nil {
		log.Printf("Warning: pattern generation failed: %v. Using fallback patterns.", err)
		// Use fallback pattern
		patterns = []string{"%entity"}
	}

	log.Printf("[Activity] Generated %d patterns for job %s", len(patterns), input.JobID)

	return &GeneratePatternsOutput{
		Patterns: patterns,
	}, nil
}

// SerpFetch performs web search for a row
func (a *Activities) SerpFetch(ctx context.Context, input SerpFetchInput) (*SerpFetchOutput, error) {
	log.Printf("[Activity] SERP fetch for job %s, row %s", input.JobID, input.RowKey)

	// Build queries from patterns
	queryBuilder := services.NewPatternQueryBuilder(input.QueryPatterns, input.ColumnsMetadata)
	queries := queryBuilder.Build(input.RowKey)

	allResults := []*models.GoogleSearchResults{}
	var lastErr error

	// Execute all queries
	for _, query := range queries {
		serp, err := a.webSearcher.Search(ctx, query)
		if err != nil {
			log.Printf("[Activity] SERP fetch error for query '%s': %v", query, err)
			lastErr = err
			continue
		}
		allResults = append(allResults, serp)
	}

	// If all queries failed, return error
	if len(allResults) == 0 {
		return nil, fmt.Errorf("all SERP queries failed: %w", lastErr)
	}

	serpData := &models.SerpData{
		Queries: queries,
		Results: allResults,
	}

	log.Printf("[Activity] SERP fetch completed for job %s, row %s: %d results from %d queries",
		input.JobID, input.RowKey, len(allResults), len(queries))

	return &SerpFetchOutput{
		SerpData: serpData,
	}, nil
}

// MakeDecision analyzes SERP results and decides what to crawl
func (a *Activities) MakeDecision(ctx context.Context, input DecisionInput) (*DecisionOutput, error) {
	log.Printf("[Activity] Making decision for job %s, row %s", input.JobID, input.RowKey)

	if input.SerpData == nil || len(input.SerpData.Results) == 0 {
		return nil, fmt.Errorf("no SERP data available for decision making")
	}

	// Merge all SERP results
	mergedResults := input.SerpData.Results[0]
	for i := 1; i < len(input.SerpData.Results); i++ {
		mergedResults.Organic = append(mergedResults.Organic, input.SerpData.Results[i].Organic...)
		if mergedResults.KnowledgeGraph == nil {
			mergedResults.KnowledgeGraph = input.SerpData.Results[i].KnowledgeGraph
		}
		mergedResults.PeopleAlsoAsk = append(mergedResults.PeopleAlsoAsk, input.SerpData.Results[i].PeopleAlsoAsk...)
	}

	// Use decision maker to analyze results
	decision, err := a.decisionMaker.MakeDecision(ctx, mergedResults, input.ColumnsMetadata, input.RowKey)
	if err != nil {
		return nil, fmt.Errorf("decision making failed: %w", err)
	}

	log.Printf("[Activity] Decision made for job %s, row %s: %d URLs to crawl, %d missing columns",
		input.JobID, input.RowKey, len(decision.URLsToCrawl), len(decision.MissingColumns))

	return &DecisionOutput{
		Decision: decision,
	}, nil
}

// Crawl fetches content from selected URLs
func (a *Activities) Crawl(ctx context.Context, input CrawlInput) (*CrawlOutput, error) {
	log.Printf("[Activity] Crawling for job %s, row %s", input.JobID, input.RowKey)

	if input.Decision == nil {
		return nil, fmt.Errorf("no decision data available for crawling")
	}

	// If no URLs to crawl, return empty result
	if len(input.Decision.URLsToCrawl) == 0 {
		log.Printf("[Activity] No URLs to crawl for job %s, row %s", input.JobID, input.RowKey)
		return &CrawlOutput{
			CrawlResults: &models.CrawlResults{
				Content: nil,
				Sources: []string{},
			},
		}, nil
	}

	// Crawl the URLs
	content, err := a.crawler.Crawl(ctx, input.Decision.URLsToCrawl)
	if err != nil {
		return nil, fmt.Errorf("crawling failed: %w", err)
	}

	crawlResults := &models.CrawlResults{
		Content: &content,
		Sources: input.Decision.URLsToCrawl,
	}

	log.Printf("[Activity] Crawling completed for job %s, row %s: %d sources",
		input.JobID, input.RowKey, len(input.Decision.URLsToCrawl))

	return &CrawlOutput{
		CrawlResults: crawlResults,
	}, nil
}

// Extract extracts structured data from crawled content
func (a *Activities) Extract(ctx context.Context, input ExtractInput) (*ExtractOutput, error) {
	log.Printf("[Activity] Extracting data for job %s, row %s", input.JobID, input.RowKey)

	var extractedData map[string]interface{}
	var confidence map[string]*models.FieldConfidenceInfo

	// If we have crawled content, extract from it
	if input.CrawlResults != nil && input.CrawlResults.Content != nil && *input.CrawlResults.Content != "" {
		extracted, conf, err := a.contentExtractor.ExtractContent(
			ctx,
			*input.CrawlResults.Content,
			input.Decision.MissingColumns,
			input.ColumnsMetadata,
		)
		if err != nil {
			return nil, fmt.Errorf("content extraction failed: %w", err)
		}
		extractedData = extracted
		confidence = conf
	} else {
		// No crawled content, use only decision stage data
		extractedData = make(map[string]interface{})
		confidence = make(map[string]*models.FieldConfidenceInfo)
	}

	// Merge with data extracted during decision stage
	if input.Decision != nil && input.Decision.ExtractedData != nil {
		for k, v := range input.Decision.ExtractedData {
			if _, exists := extractedData[k]; !exists {
				extractedData[k] = v
				// Add confidence score for decision-stage extracted data
				if confidence[k] == nil {
					confidence[k] = &models.FieldConfidenceInfo{
						Score:  0.8, // Decision stage extraction gets 0.8 confidence
						Reason: "Extracted from SERP results",
					}
				}
			}
		}
	}

	log.Printf("[Activity] Extraction completed for job %s, row %s: %d fields extracted",
		input.JobID, input.RowKey, len(extractedData))

	return &ExtractOutput{
		ExtractedData: extractedData,
		Confidence:    confidence,
	}, nil
}

// UpdateState updates the row state in the database
func (a *Activities) UpdateState(ctx context.Context, input StateUpdateInput) error {
	log.Printf("[Activity] Updating state for job %s, row %s to %s", input.JobID, input.RowKey, input.Stage)

	err := a.stateManager.Transition(ctx, input.JobID, input.RowKey, input.Stage, input.Data)
	if err != nil {
		return fmt.Errorf("state transition failed: %w", err)
	}

	return nil
}

// AnalyzeFeedback analyzes the enrichment results to determine if feedback is needed
// This supports the future feedback loop capability
func (a *Activities) AnalyzeFeedback(ctx context.Context, input FeedbackAnalysisInput) (*FeedbackAnalysisOutput, error) {
	log.Printf("[Activity] Analyzing feedback for job %s, row %s", input.JobID, input.RowKey)

	output := &FeedbackAnalysisOutput{
		NeedsFeedback:        false,
		LowConfidenceColumns: []string{},
		MissingColumns:       []string{},
		AverageConfidence:    1.0,
	}

	// Check for missing columns
	for _, col := range input.ColumnsMetadata {
		if _, exists := input.ExtractedData[col.Name]; !exists {
			output.MissingColumns = append(output.MissingColumns, col.Name)
		}
	}

	// Check confidence scores
	confidenceThreshold := 0.6
	var totalConfidence float64
	var confidenceCount int

	for colName, confInfo := range input.Confidence {
		totalConfidence += confInfo.Score
		confidenceCount++

		if confInfo.Score < confidenceThreshold {
			output.LowConfidenceColumns = append(output.LowConfidenceColumns, colName)
		}
	}

	if confidenceCount > 0 {
		output.AverageConfidence = totalConfidence / float64(confidenceCount)
	}

	// Determine if feedback is needed
	// Feedback is needed if there are missing columns or low confidence columns
	if len(output.MissingColumns) > 0 || len(output.LowConfidenceColumns) > 0 {
		output.NeedsFeedback = true
	}

	log.Printf("[Activity] Feedback analysis for job %s, row %s: needs_feedback=%v, avg_confidence=%.2f, missing=%d, low_confidence=%d",
		input.JobID, input.RowKey, output.NeedsFeedback, output.AverageConfidence,
		len(output.MissingColumns), len(output.LowConfidenceColumns))

	return output, nil
}

// InitializeJob creates initial row states in the database
func (a *Activities) InitializeJob(ctx context.Context, jobID string, rowKeys []string) error {
	log.Printf("[Activity] Initializing job %s with %d rows", jobID, len(rowKeys))

	err := a.stateManager.InitializeJob(ctx, jobID, rowKeys)
	if err != nil {
		return fmt.Errorf("job initialization failed: %w", err)
	}

	return nil
}

// CompleteJob marks the job as completed
func (a *Activities) CompleteJob(ctx context.Context, jobID string) error {
	log.Printf("[Activity] Completing job %s", jobID)

	err := a.stateManager.Complete(ctx, jobID)
	if err != nil {
		return fmt.Errorf("job completion failed: %w", err)
	}

	return nil
}
