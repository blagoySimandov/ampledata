package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"google.golang.org/genai"

	"github.com/blagoySimandov/ampledata/go/internal/models"
)

// AgenticContentExtractor is an IContentExtractor that, instead of doing a
// single one-shot extraction from pre-fetched content, gives the LLM two tools:
//
//   - search_web(query)  – calls the Serper search API for targeted lookups
//   - fetch_page(url)    – crawls a URL via Crawl4ai
//
// This lets the model iteratively seek out missing fields rather than relying
// on the outer Temporal retry loop, which re-runs the full pipeline at significant
// latency and cost. It also sidesteps the hard 8 000-character content truncation
// that AIContentExtractor applies.
type AgenticContentExtractor struct {
	client        IAIClient
	promptService IPromptService
	webSearcher   WebSearcher
	crawler       WebCrawler
}

func NewAgenticContentExtractor(
	client IAIClient,
	promptService IPromptService,
	webSearcher WebSearcher,
	crawler WebCrawler,
) *AgenticContentExtractor {
	return &AgenticContentExtractor{
		client:        client,
		promptService: promptService,
		webSearcher:   webSearcher,
		crawler:       crawler,
	}
}

// Extract implements IContentExtractor. The content parameter is used as the
// initial context; the model may call search_web / fetch_page to supplement it.
func (a *AgenticContentExtractor) Extract(
	ctx context.Context,
	content string,
	entityKey string,
	columnsMetadata []*models.ColumnMetadata,
	keyColumnDescription string,
) (*ExtractionResult, error) {
	tools := a.buildTools()
	prompt := a.buildPrompt(content, entityKey, keyColumnDescription, columnsMetadata)

	handler := func(name string, args map[string]any) (any, error) {
		return a.handleToolCall(ctx, entityKey, name, args)
	}

	raw, err := a.client.GenerateWithTools(ctx, prompt, tools, handler)
	if err != nil {
		return nil, fmt.Errorf("agentic extraction failed: %w", err)
	}

	er, err := parseResponse(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to parse agentic extraction response: %w", err)
	}

	er.ExtractedData = ValidateAndCoerceTypes(er.ExtractedData, columnsMetadata, er.Confidence)
	return er, nil
}

// buildTools declares the two tools the model may invoke.
func (a *AgenticContentExtractor) buildTools() []*genai.Tool {
	return []*genai.Tool{
		{
			FunctionDeclarations: []*genai.FunctionDeclaration{
				{
					Name:        "search_web",
					Description: "Search the web for information about the entity. Use targeted queries to find specific missing fields.",
					Parameters: &genai.Schema{
						Type: genai.TypeObject,
						Properties: map[string]*genai.Schema{
							"query": {
								Type:        genai.TypeString,
								Description: "The search query to send to Google. Be specific and include the entity name.",
							},
						},
						Required: []string{"query"},
					},
				},
				{
					Name:        "fetch_page",
					Description: "Fetch and read the full content of a specific web page. Use when a URL is known to contain the required data.",
					Parameters: &genai.Schema{
						Type: genai.TypeObject,
						Properties: map[string]*genai.Schema{
							"url": {
								Type:        genai.TypeString,
								Description: "The URL of the page to fetch.",
							},
						},
						Required: []string{"url"},
					},
				},
			},
		},
	}
}

// handleToolCall dispatches a tool call to the appropriate service.
func (a *AgenticContentExtractor) handleToolCall(ctx context.Context, entityKey, name string, args map[string]any) (any, error) {
	switch name {
	case "search_web":
		query, _ := args["query"].(string)
		if query == "" {
			return nil, fmt.Errorf("search_web: missing query argument")
		}
		results, err := a.webSearcher.Search(ctx, query)
		if err != nil {
			return nil, fmt.Errorf("search_web: %w", err)
		}
		return formatSearchResults(results), nil

	case "fetch_page":
		url, _ := args["url"].(string)
		if url == "" {
			return nil, fmt.Errorf("fetch_page: missing url argument")
		}
		content, err := a.crawler.Crawl(ctx, []string{url}, entityKey)
		if err != nil {
			return nil, fmt.Errorf("fetch_page: %w", err)
		}
		return content, nil

	default:
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
}

// buildPrompt constructs the initial prompt for the agentic loop.
func (a *AgenticContentExtractor) buildPrompt(
	initialContent string,
	entityKey string,
	keyColumnDescription string,
	columns []*models.ColumnMetadata,
) string {
	entityContext := formatEntityContext(entityKey, keyColumnDescription)
	colLines := columnsText(columns)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(`You are a data extraction agent. Your goal is to find and extract specific fields about the entity %s.

Fields to extract:
%s

You have been given initial web content below. Read it carefully and extract any fields you can find.
For any fields that are missing or have low confidence, use the available tools:
- search_web(query): Search Google for targeted information
- fetch_page(url): Fetch the full content of a specific URL

Guidelines:
- Only extract data that clearly refers to the target entity %s
- Use confidence scores (0.0–1.0) to reflect certainty; prefer explicit statements over inferences
- Call search_web or fetch_page only when needed; stop when all fields are found or further searching is unlikely to help
- After gathering enough information, respond with a JSON object (no markdown fences):

{
    "extracted_data": {"field_name": value_or_null},
    "confidence": {"field_name": {"score": 0.95, "reason": "one-sentence explanation"}},
    "reasoning": "overall summary of what was found and how"
}

Every field listed above must appear in both extracted_data and confidence, even if the value is null.

--- Initial content ---
`, entityContext, colLines, entityContext))

	// Include the initial content without hard-truncating; the model can request
	// more via fetch_page if needed.
	if len(initialContent) > 20000 {
		sb.WriteString(initialContent[:20000])
		sb.WriteString("\n[content truncated — use fetch_page to retrieve more if needed]")
	} else {
		sb.WriteString(initialContent)
	}

	return sb.String()
}

// formatSearchResults serialises SERP results into a compact text representation
// that the model can read in a tool response.
func formatSearchResults(results *models.GoogleSearchResults) string {
	if results == nil {
		return "no results"
	}
	b, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return "failed to format results"
	}
	return string(b)
}
