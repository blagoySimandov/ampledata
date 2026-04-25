package activities

import (
	"context"
	"fmt"
	"strings"

	"github.com/blagoySimandov/ampledata/go/internal/logger"
	"github.com/blagoySimandov/ampledata/go/internal/models"
	"github.com/blagoySimandov/ampledata/go/internal/services"
	"github.com/blagoySimandov/ampledata/go/internal/sheets"
	"github.com/blagoySimandov/ampledata/go/internal/state"
	"github.com/google/uuid"
)

type Activities struct {
	stateManager     *state.StateManager
	webSearcher      services.WebSearcher
	decisionMaker    services.DecisionMaker
	crawler          services.WebCrawler
	contentExtractor services.IContentExtractor
	patternGenerator services.QueryPatternGenerator
	billingService   services.BillingService
	sheetsClient     *sheets.Client
}

func NewActivities(
	stateManager *state.StateManager,
	webSearcher services.WebSearcher,
	decisionMaker services.DecisionMaker,
	crawler services.WebCrawler,
	contentExtractor services.IContentExtractor,
	patternGenerator services.QueryPatternGenerator,
	billingService services.BillingService,
	sheetsClient *sheets.Client,
) *Activities {
	return &Activities{
		stateManager:     stateManager,
		webSearcher:      webSearcher,
		decisionMaker:    decisionMaker,
		crawler:          crawler,
		contentExtractor: contentExtractor,
		patternGenerator: patternGenerator,
		billingService:   billingService,
		sheetsClient:     sheetsClient,
	}
}

type GeneratePatternsInput struct {
	JobID           string
	ColumnsMetadata []*models.ColumnMetadata
}

type GeneratePatternsWithFeedbackInput struct {
	JobID            string
	ColumnsMetadata  []*models.ColumnMetadata
	PreviousAttempts []*models.EnrichmentAttempt
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
	JobID            string
	RowKey           string
	SerpData         *models.SerpData
	ColumnsMetadata  []*models.ColumnMetadata
	KeyColumnDescription      string
	PreviousAttempts []*models.EnrichmentAttempt
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
	KeyColumnDescription      string
}

type ExtractOutput struct {
	ExtractedData map[string]interface{}
	Confidence    map[string]*models.FieldConfidenceInfo
	Reasoning     string
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
	ctx = services.ContextWithJobID(ctx, input.JobID)
	event := logger.NewActivityEvent("generate_patterns", input.JobID)

	patterns, err := a.patternGenerator.GeneratePatterns(ctx, input.ColumnsMetadata)
	if err != nil {
		logger.Log.Warn("pattern generation failed, using fallback", "error", err, "job_id", input.JobID)
		patterns = []string{"%entity"}
		event.SetMetadata("fallback_used", true)
	}

	event.EmitActivitySuccess(ctx, map[string]interface{}{
		"pattern_count": len(patterns),
	})

	return &GeneratePatternsOutput{
		Patterns: patterns,
	}, nil
}

func (a *Activities) GeneratePatternsWithFeedback(ctx context.Context, input GeneratePatternsWithFeedbackInput) (*GeneratePatternsOutput, error) {
	ctx = services.ContextWithJobID(ctx, input.JobID)
	event := logger.NewActivityEvent("generate_patterns_with_feedback", input.JobID)
	event.SetMetadata("attempt_count", len(input.PreviousAttempts))

	patterns, err := a.patternGenerator.GeneratePatternsWithFeedback(ctx, input.ColumnsMetadata, input.PreviousAttempts)
	if err != nil {
		logger.Log.Warn("pattern generation with feedback failed, using fallback", "error", err, "job_id", input.JobID)
		patterns = []string{"%entity"}
		event.SetMetadata("fallback_used", true)
	}

	event.EmitActivitySuccess(ctx, map[string]interface{}{
		"pattern_count": len(patterns),
	})

	return &GeneratePatternsOutput{
		Patterns: patterns,
	}, nil
}

func (a *Activities) SerpFetch(ctx context.Context, input SerpFetchInput) (*SerpFetchOutput, error) {
	event := logger.NewActivityEvent("serp_fetch", input.JobID)
	event.RowKey = input.RowKey

	queryBuilder := services.NewPatternQueryBuilder(input.QueryPatterns, input.ColumnsMetadata)
	queries := queryBuilder.Build(input.RowKey)
	event.SetMetadata("query_count", len(queries))

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
		event.EmitActivityError(ctx, fmt.Errorf("all SERP queries failed: %w", lastErr))
		return nil, fmt.Errorf("all SERP queries failed: %w", lastErr)
	}

	serpData := &models.SerpData{
		Queries: queries,
		Results: allResults,
	}

	event.EmitActivitySuccess(ctx, map[string]interface{}{
		"result_count": len(allResults),
		"query_count":  len(queries),
	})

	return &SerpFetchOutput{
		SerpData: serpData,
	}, nil
}

func mergeSerpResults(results []*models.GoogleSearchResults) *models.GoogleSearchResults {
	merged := results[0]
	for i := 1; i < len(results); i++ {
		merged.Organic = append(merged.Organic, results[i].Organic...)
		if merged.KnowledgeGraph == nil {
			merged.KnowledgeGraph = results[i].KnowledgeGraph
		}
		merged.PeopleAlsoAsk = append(merged.PeopleAlsoAsk, results[i].PeopleAlsoAsk...)
	}
	return merged
}

func (a *Activities) MakeDecision(ctx context.Context, input DecisionInput) (*DecisionOutput, error) {
	event := logger.NewActivityEvent("make_decision", input.JobID)
	event.RowKey = input.RowKey

	if input.SerpData == nil || len(input.SerpData.Results) == 0 {
		err := fmt.Errorf("no SERP data available for decision making")
		event.EmitActivityError(ctx, err)
		return nil, err
	}

	mergedResults := mergeSerpResults(input.SerpData.Results)
	crawlDecision, err := a.decisionMaker.MakeDecision(ctx, mergedResults, input.RowKey, 3, input.ColumnsMetadata, input.KeyColumnDescription, input.PreviousAttempts)
	if err != nil {
		event.EmitActivityError(ctx, fmt.Errorf("decision making failed: %w", err))
		return nil, fmt.Errorf("decision making failed: %w", err)
	}

	decision := &models.Decision{
		URLsToCrawl:    crawlDecision.URLsToCrawl,
		ExtractedData:  crawlDecision.ExtractedData,
		Confidence:     crawlDecision.Confidence,
		Reasoning:      crawlDecision.Reasoning,
		SourceURLs:     crawlDecision.SourceURLs,
		MissingColumns: crawlDecision.MissingColumns,
	}

	event.EmitActivitySuccess(ctx, map[string]interface{}{
		"urls_to_crawl":   len(decision.URLsToCrawl),
		"missing_columns": len(decision.MissingColumns),
		"extracted_count": len(decision.ExtractedData),
	})

	return &DecisionOutput{
		Decision: decision,
	}, nil
}

func (a *Activities) Crawl(ctx context.Context, input CrawlInput) (*CrawlOutput, error) {
	event := logger.NewActivityEvent("crawl", input.JobID)
	event.RowKey = input.RowKey

	if input.Decision == nil {
		err := fmt.Errorf("no decision data available for crawling")
		event.EmitActivityError(ctx, err)
		return nil, err
	}

	if len(input.Decision.URLsToCrawl) == 0 {
		event.EmitActivitySuccess(ctx, map[string]interface{}{
			"sources": 0,
			"skipped": true,
		})
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
		event.EmitActivityError(ctx, fmt.Errorf("crawling failed: %w", err))
		return nil, fmt.Errorf("crawling failed: %w", err)
	}

	crawlResults := &models.CrawlResults{
		Content: &content,
		Sources: input.Decision.URLsToCrawl,
	}

	event.EmitActivitySuccess(ctx, map[string]interface{}{
		"sources":       len(input.Decision.URLsToCrawl),
		"content_bytes": len(content),
	})

	return &CrawlOutput{
		CrawlResults: crawlResults,
	}, nil
}

func filterMissingColumnsMetadata(missingColumns []string, allColumns []*models.ColumnMetadata) []*models.ColumnMetadata {
	result := []*models.ColumnMetadata{}
	for _, colName := range missingColumns {
		for _, col := range allColumns {
			if col.Name == colName {
				result = append(result, col)
				break
			}
		}
	}
	return result
}

func (a *Activities) extractFromContent(ctx context.Context, content, rowKey string, metadata []*models.ColumnMetadata, entityType string) (map[string]interface{}, map[string]*models.FieldConfidenceInfo, string, error) {
	result, err := a.contentExtractor.Extract(ctx, content, rowKey, metadata, entityType)
	if err != nil {
		return nil, nil, "", fmt.Errorf("content extraction failed: %w", err)
	}

	confidence := result.Confidence
	if confidence == nil {
		confidence = make(map[string]*models.FieldConfidenceInfo)
	}
	return result.ExtractedData, confidence, result.Reasoning, nil
}

func mergeDecisionData(extractedData map[string]interface{}, confidence map[string]*models.FieldConfidenceInfo, decision *models.Decision) {
	if decision == nil || decision.ExtractedData == nil {
		return
	}

	for k, v := range decision.ExtractedData {
		// Skip null values - if data wasn't extracted, don't add it with false confidence
		if v == nil {
			continue
		}

		if _, exists := extractedData[k]; !exists {
			extractedData[k] = v
			if confidence[k] == nil {
				if decision.Confidence != nil {
					if conf, ok := decision.Confidence[k]; ok && conf != nil {
						confidence[k] = conf
						continue
					}
				}
				confidence[k] = &models.FieldConfidenceInfo{
					Score:  0.6,
					Reason: "Extracted from SERP results (no per-field confidence provided by LLM)",
				}
			}
		}
	}
}

func (a *Activities) Extract(ctx context.Context, input ExtractInput) (*ExtractOutput, error) {
	ctx = services.ContextWithJobID(ctx, input.JobID)
	event := logger.NewActivityEvent("extract", input.JobID)
	event.RowKey = input.RowKey

	var extractedData map[string]interface{}
	var confidence map[string]*models.FieldConfidenceInfo
	var reasoning string

	if input.CrawlResults != nil && input.CrawlResults.Content != nil && *input.CrawlResults.Content != "" {
		missingColsMetadata := filterMissingColumnsMetadata(input.Decision.MissingColumns, input.ColumnsMetadata)

		if len(missingColsMetadata) > 0 {
			var err error
			extractedData, confidence, reasoning, err = a.extractFromContent(ctx, *input.CrawlResults.Content, input.RowKey, missingColsMetadata, input.KeyColumnDescription)
			if err != nil {
				event.EmitActivityError(ctx, err)
				return nil, err
			}
		} else {
			extractedData = make(map[string]interface{})
			confidence = make(map[string]*models.FieldConfidenceInfo)
		}
	} else {
		extractedData = make(map[string]interface{})
		confidence = make(map[string]*models.FieldConfidenceInfo)
	}

	mergeDecisionData(extractedData, confidence, input.Decision)

	// Calculate average confidence
	var totalConfidence float64
	for _, conf := range confidence {
		totalConfidence += conf.Score
	}
	avgConfidence := 0.0
	if len(confidence) > 0 {
		avgConfidence = totalConfidence / float64(len(confidence))
	}

	event.EmitActivitySuccess(ctx, map[string]interface{}{
		"fields_extracted":   len(extractedData),
		"average_confidence": avgConfidence,
	})

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
		Reasoning:     reasoning,
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

	confidenceThreshold := 0.75
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
	event := logger.NewActivityEvent("initialize_job", jobID)
	event.SetMetadata("row_count", len(rowKeys))

	err := a.stateManager.InitializeJob(ctx, jobID, rowKeys)
	if err != nil {
		event.EmitActivityError(ctx, fmt.Errorf("job initialization failed: %w", err))
		return fmt.Errorf("job initialization failed: %w", err)
	}

	event.EmitActivitySuccess(ctx, map[string]interface{}{
		"rows_initialized": len(rowKeys),
	})

	return nil
}

func (a *Activities) InitializeJobRow(ctx context.Context, jobID, rowKey string) error {
	return a.stateManager.Store().BulkCreateRows(ctx, jobID, []string{rowKey})
}

func (a *Activities) CompleteJob(ctx context.Context, jobID string) error {
	event := logger.NewActivityEvent("complete_job", jobID)

	err := a.stateManager.Complete(ctx, jobID)
	if err != nil {
		event.EmitActivityError(ctx, fmt.Errorf("job completion failed: %w", err))
		return fmt.Errorf("job completion failed: %w", err)
	}

	event.EmitActivitySuccess(ctx, nil)

	return nil
}

type ReportUsageInput struct {
	StripeCustomerID string
	Credits          int
}

func (a *Activities) ReportUsage(ctx context.Context, input ReportUsageInput) error {
	return a.billingService.ReportUsage(ctx, input.StripeCustomerID, input.Credits)
}

type IncrementJobCreditsInput struct {
	JobID   string
	Credits int
}

func (a *Activities) IncrementJobCredits(ctx context.Context, input IncrementJobCreditsInput) error {
	return a.stateManager.Store().IncrementJobCost(ctx, input.JobID, 0, input.Credits)
}

type WriteResultsToSheetInput struct {
	JobID           string
	SourceID        string
	UserID          string
	ColumnsMetadata []*models.ColumnMetadata
}

func (a *Activities) WriteResultsToSheet(ctx context.Context, input WriteResultsToSheetInput) error {
	sourceID, err := uuid.Parse(input.SourceID)
	if err != nil {
		return fmt.Errorf("invalid source id: %w", err)
	}
	source, err := a.stateManager.Store().GetSource(ctx, sourceID)
	if err != nil {
		return fmt.Errorf("failed to get source: %w", err)
	}
	sheetsMeta, ok := source.Metadata.(*models.GoogleSheetsSourceMetadata)
	if !ok {
		return fmt.Errorf("source is not a google sheets source")
	}
	job, err := a.stateManager.Store().GetJob(ctx, input.JobID)
	if err != nil {
		return fmt.Errorf("failed to get job: %w", err)
	}
	existing, err := a.sheetsClient.ReadSheetData(ctx, input.UserID, sheetsMeta.SpreadsheetID, sheetsMeta.SheetName)
	if err != nil {
		return fmt.Errorf("failed to read sheet: %w", err)
	}
	dataByKey, err := a.buildEnrichedDataMap(ctx, input.JobID)
	if err != nil {
		return err
	}
	enrichedColNames := columnNames(input.ColumnsMetadata)
	return a.sheetsClient.WriteResults(ctx, input.UserID, sheetsMeta.SpreadsheetID, sheetsMeta.SheetName, existing, job.KeyColumns, enrichedColNames, dataByKey)
}

func (a *Activities) buildEnrichedDataMap(ctx context.Context, jobID string) (map[string]map[string]interface{}, error) {
	result := map[string]map[string]interface{}{}
	offset := 0
	for {
		rows, err := a.stateManager.Store().GetRowsAtStage(ctx, jobID, models.StageCompleted, offset, 500)
		if err != nil {
			return nil, fmt.Errorf("failed to get completed rows: %w", err)
		}
		for _, row := range rows {
			result[row.Key] = row.ExtractedData
		}
		if len(rows) < 500 {
			break
		}
		offset += 500
	}
	return result, nil
}

func columnNames(cols []*models.ColumnMetadata) []string {
	names := make([]string, len(cols))
	for i, c := range cols {
		names[i] = c.Name
	}
	return names
}
