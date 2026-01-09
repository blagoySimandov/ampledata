package activities

import (
	"context"
	"fmt"
	"strings"

	"github.com/blagoySimandov/ampledata/go/internal/logger"
	"github.com/blagoySimandov/ampledata/go/internal/models"
	"github.com/blagoySimandov/ampledata/go/internal/services"
	"github.com/blagoySimandov/ampledata/go/internal/state"
)

type Activities struct {
	stateManager     *state.StateManager
	webSearcher      services.WebSearcher
	decisionMaker    services.DecisionMaker
	crawler          services.WebCrawler
	contentExtractor services.ContentExtractor
	patternGenerator services.QueryPatternGenerator
}

func NewActivities(
	stateManager *state.StateManager,
	webSearcher services.WebSearcher,
	decisionMaker services.DecisionMaker,
	crawler services.WebCrawler,
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
	SerpData        *models.SerpData
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
	Data   *models.StateUpdate
}

type FeedbackAnalysisInput struct {
	JobID           string
	RowKey          string
	ExtractedData   map[string]interface{}
	Confidence      map[string]*models.FieldConfidenceInfo
	ColumnsMetadata []*models.ColumnMetadata
}

type FeedbackAnalysisOutput struct {
	NeedsFeedback        bool
	LowConfidenceColumns []string
	MissingColumns       []string
	AverageConfidence    float64
}

func (a *Activities) GeneratePatterns(ctx context.Context, input GeneratePatternsInput) (*GeneratePatternsOutput, error) {
	patterns, err := a.patternGenerator.GeneratePatterns(ctx, input.ColumnsMetadata)
	if err != nil {
		logger.Log.Warn("pattern generation failed, using fallback", "error", err, "job_id", input.JobID)
		patterns = []string{"%entity"}
	}

	logger.Log.Info("patterns generated", "job_id", input.JobID, "count", len(patterns))

	return &GeneratePatternsOutput{
		Patterns: patterns,
	}, nil
}

func (a *Activities) SerpFetch(ctx context.Context, input SerpFetchInput) (*SerpFetchOutput, error) {
	queryBuilder := services.NewPatternQueryBuilder(input.QueryPatterns, input.ColumnsMetadata)
	queries := queryBuilder.Build(input.RowKey)

	allResults := []*models.GoogleSearchResults{}
	var lastErr error

	for _, query := range queries {
		serp, err := a.webSearcher.Search(ctx, query)
		if err != nil {
			lastErr = err
			continue
		}
		allResults = append(allResults, serp)
	}

	if len(allResults) == 0 {
		return nil, fmt.Errorf("all SERP queries failed: %w", lastErr)
	}

	serpData := &models.SerpData{
		Queries: queries,
		Results: allResults,
	}

	logger.Log.Info("serp fetch completed",
		"job_id", input.JobID,
		"row_key", input.RowKey,
		"results", len(allResults),
		"queries", len(queries))

	return &SerpFetchOutput{
		SerpData: serpData,
	}, nil
}

func (a *Activities) MakeDecision(ctx context.Context, input DecisionInput) (*DecisionOutput, error) {
	if input.SerpData == nil || len(input.SerpData.Results) == 0 {
		return nil, fmt.Errorf("no SERP data available for decision making")
	}

	mergedResults := input.SerpData.Results[0]
	for i := 1; i < len(input.SerpData.Results); i++ {
		mergedResults.Organic = append(mergedResults.Organic, input.SerpData.Results[i].Organic...)
		if mergedResults.KnowledgeGraph == nil {
			mergedResults.KnowledgeGraph = input.SerpData.Results[i].KnowledgeGraph
		}
		mergedResults.PeopleAlsoAsk = append(mergedResults.PeopleAlsoAsk, input.SerpData.Results[i].PeopleAlsoAsk...)
	}

	crawlDecision, err := a.decisionMaker.MakeDecision(ctx, mergedResults, input.RowKey, 3, input.ColumnsMetadata)
	if err != nil {
		return nil, fmt.Errorf("decision making failed: %w", err)
	}

	decision := &models.Decision{
		URLsToCrawl:    crawlDecision.URLsToCrawl,
		ExtractedData:  crawlDecision.ExtractedData,
		Reasoning:      crawlDecision.Reasoning,
		MissingColumns: crawlDecision.MissingColumns,
	}

	logger.Log.Info("decision made",
		"job_id", input.JobID,
		"row_key", input.RowKey,
		"urls_to_crawl", len(decision.URLsToCrawl),
		"missing_columns", len(decision.MissingColumns))

	return &DecisionOutput{
		Decision: decision,
	}, nil
}

func (a *Activities) Crawl(ctx context.Context, input CrawlInput) (*CrawlOutput, error) {
	if input.Decision == nil {
		return nil, fmt.Errorf("no decision data available for crawling")
	}

	if len(input.Decision.URLsToCrawl) == 0 {
		return &CrawlOutput{
			CrawlResults: &models.CrawlResults{
				Content: nil,
				Sources: nil,
			},
		}, nil
	}

	query := ""
	if input.SerpData != nil && len(input.SerpData.Queries) > 0 {
		query = strings.Join(input.SerpData.Queries, " ")
	}

	content, err := a.crawler.Crawl(ctx, input.Decision.URLsToCrawl, query)
	if err != nil {
		return nil, fmt.Errorf("crawling failed: %w", err)
	}

	crawlResults := &models.CrawlResults{
		Content: &content,
		Sources: input.Decision.URLsToCrawl,
	}

	logger.Log.Info("crawl completed",
		"job_id", input.JobID,
		"row_key", input.RowKey,
		"sources", len(input.Decision.URLsToCrawl))

	return &CrawlOutput{
		CrawlResults: crawlResults,
	}, nil
}

func (a *Activities) Extract(ctx context.Context, input ExtractInput) (*ExtractOutput, error) {
	var extractedData map[string]interface{}
	var confidence map[string]*models.FieldConfidenceInfo

	if input.CrawlResults != nil && input.CrawlResults.Content != nil && *input.CrawlResults.Content != "" {
		missingColsMetadata := []*models.ColumnMetadata{}
		for _, colName := range input.Decision.MissingColumns {
			for _, col := range input.ColumnsMetadata {
				if col.Name == colName {
					missingColsMetadata = append(missingColsMetadata, col)
					break
				}
			}
		}

		if len(missingColsMetadata) > 0 {
			result, err := a.contentExtractor.Extract(
				ctx,
				*input.CrawlResults.Content,
				input.RowKey,
				missingColsMetadata,
			)
			if err != nil {
				return nil, fmt.Errorf("content extraction failed: %w", err)
			}

			extractedData = result.ExtractedData
			// No conversion needed - ExtractionResult already uses models.FieldConfidenceInfo
			confidence = result.Confidence
			if confidence == nil {
				confidence = make(map[string]*models.FieldConfidenceInfo)
			}
		} else {
			extractedData = make(map[string]interface{})
			confidence = make(map[string]*models.FieldConfidenceInfo)
		}
	} else {
		extractedData = make(map[string]interface{})
		confidence = make(map[string]*models.FieldConfidenceInfo)
	}

	if input.Decision != nil && input.Decision.ExtractedData != nil {
		for k, v := range input.Decision.ExtractedData {
			if _, exists := extractedData[k]; !exists {
				extractedData[k] = v
				if confidence[k] == nil {
					confidence[k] = &models.FieldConfidenceInfo{
						Score:  0.8,
						Reason: "Extracted from SERP results",
					}
				}
			}
		}
	}

	logger.Log.Info("extraction completed",
		"job_id", input.JobID,
		"row_key", input.RowKey,
		"fields_extracted", len(extractedData))

	var finalExtractedData map[string]interface{}
	if len(extractedData) > 0 {
		finalExtractedData = extractedData
	}

	var finalConfidence map[string]*models.FieldConfidenceInfo
	if len(confidence) > 0 {
		finalConfidence = confidence
	}

	return &ExtractOutput{
		ExtractedData: finalExtractedData,
		Confidence:    finalConfidence,
	}, nil
}

func (a *Activities) UpdateState(ctx context.Context, input StateUpdateInput) error {
	err := a.stateManager.Transition(ctx, input.JobID, input.RowKey, input.Stage, input.Data)
	if err != nil {
		return fmt.Errorf("state transition failed: %w", err)
	}

	return nil
}

func (a *Activities) AnalyzeFeedback(ctx context.Context, input FeedbackAnalysisInput) (*FeedbackAnalysisOutput, error) {
	output := &FeedbackAnalysisOutput{
		NeedsFeedback:        false,
		LowConfidenceColumns: []string{},
		MissingColumns:       []string{},
		AverageConfidence:    1.0,
	}

	for _, col := range input.ColumnsMetadata {
		if _, exists := input.ExtractedData[col.Name]; !exists {
			output.MissingColumns = append(output.MissingColumns, col.Name)
		}
	}

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

	if len(output.MissingColumns) > 0 || len(output.LowConfidenceColumns) > 0 {
		output.NeedsFeedback = true
	}

	return output, nil
}

func (a *Activities) InitializeJob(ctx context.Context, jobID string, rowKeys []string) error {
	logger.Log.Info("initializing job", "job_id", jobID, "rows", len(rowKeys))

	err := a.stateManager.InitializeJob(ctx, jobID, rowKeys)
	if err != nil {
		return fmt.Errorf("job initialization failed: %w", err)
	}

	return nil
}

func (a *Activities) CompleteJob(ctx context.Context, jobID string) error {
	logger.Log.Info("completing job", "job_id", jobID)

	err := a.stateManager.Complete(ctx, jobID)
	if err != nil {
		return fmt.Errorf("job completion failed: %w", err)
	}

	return nil
}
