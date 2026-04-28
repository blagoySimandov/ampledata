package api

import (
	"time"

	"github.com/blagoySimandov/ampledata/go/internal/models"
	"github.com/blagoySimandov/ampledata/go/internal/services"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func toAPITemplate(t *models.TemplateDB) Template {
	cols := make([]TemplateColumnMetadata, 0, len(t.ColumnsMetadata))
	for _, c := range t.ColumnsMetadata {
		col := TemplateColumnMetadata{
			Name:      c.Name,
			Type:      ColumnType(c.Type),
			Operation: c.Operation,
		}
		if c.Description != nil {
			desc := *c.Description
			col.Description = &desc
		}
		cols = append(cols, col)
	}
	return Template{
		Id:              openapi_types.UUID(t.ID),
		Name:            t.Name,
		Description:     t.Description,
		EntityType:      t.EntityType,
		Type:            TemplateType(t.Type),
		KeyColumns:      t.KeyColumns,
		ColumnsMetadata: cols,
		OwnedBy:         t.OwnedBy,
	}
}

func toAPITemplateList(templates []*models.TemplateDB) TemplateListResponse {
	result := make([]Template, 0, len(templates))
	for _, t := range templates {
		result = append(result, toAPITemplate(t))
	}
	return TemplateListResponse{Templates: result, TotalCount: len(result)}
}

func toModelColumnMetadataSlice(cols []ColumnMetadata) []*models.ColumnMetadata {
	result := make([]*models.ColumnMetadata, len(cols))
	for i, c := range cols {
		result[i] = &models.ColumnMetadata{
			Name:        c.Name,
			Type:        models.ColumnType(c.Type),
			JobType:     models.JobType(c.JobType),
			Description: c.Description,
		}
	}
	return result
}

func toModelColumnMetadataSlicePtr(cols *[]ColumnMetadata) []*models.ColumnMetadata {
	if cols == nil {
		return nil
	}
	return toModelColumnMetadataSlice(*cols)
}

func toAPIJobProgress(p *models.JobProgress) JobProgressResponse {
	startedAt := ""
	if !p.StartedAt.IsZero() {
		startedAt = p.StartedAt.Format(time.RFC3339)
	}
	return JobProgressResponse{
		JobId:       p.JobID,
		TotalRows:   p.TotalRows,
		RowsByStage: toAPIRowsByStage(p.RowsByStage),
		StartedAt:   startedAt,
		Status:      JobStatus(p.Status),
	}
}

func toAPIRowsByStage(m map[models.RowStage]int) map[string]int {
	result := make(map[string]int, len(m))
	for k, v := range m {
		result[string(k)] = v
	}
	return result
}

func toAPIEnrichmentResult(r *models.EnrichmentResult) EnrichmentResult {
	return EnrichmentResult{
		Key:               r.Key,
		ExtractedData:     r.ExtractedData,
		Confidence:        toAPIConfidence(r.Confidence),
		Sources:           r.Sources,
		ExtractionHistory: toAPIExtractionHistory(r.ExtractionHistory),
		Error:             r.Error,
	}
}

func toAPIRowProgressItem(r *models.RowProgressItem) RowProgressItem {
	item := RowProgressItem{
		Key:               r.Key,
		Stage:             RowStage(r.Stage),
		Confidence:        toAPIConfidence(r.Confidence),
		ExtractionHistory: toAPIExtractionHistory(r.ExtractionHistory),
		Error:             r.Error,
		UpdatedAt:         r.UpdatedAt,
	}
	if r.ExtractedData != nil {
		item.ExtractedData = &r.ExtractedData
	}
	if r.Sources != nil {
		item.Sources = &r.Sources
	}
	return item
}

func toAPIRowsProgressResponse(r *models.RowsProgressResponse) RowsProgressResponse {
	rows := make([]RowProgressItem, len(r.Rows))
	for i, row := range r.Rows {
		rows[i] = toAPIRowProgressItem(row)
	}
	return RowsProgressResponse{
		Rows:       rows,
		Pagination: toAPIPagination(r.Pagination),
	}
}

func toAPIPagination(p *models.PaginationInfo) PaginationInfo {
	return PaginationInfo{
		Total:   p.Total,
		Offset:  p.Offset,
		Limit:   p.Limit,
		HasMore: p.HasMore,
	}
}

func toAPIExtractionHistory(h []*models.ExtractionHistoryEntry) []ExtractionHistoryEntry {
	result := make([]ExtractionHistoryEntry, len(h))
	for i, e := range h {
		sources := e.Sources
		if sources == nil {
			sources = []string{}
		}
		extractedData := e.ExtractedData
		if extractedData == nil {
			extractedData = map[string]interface{}{}
		}
		result[i] = ExtractionHistoryEntry{
			AttemptNumber: e.AttemptNumber,
			ExtractedData: extractedData,
			Confidence:    toAPIConfidence(e.Confidence),
			Sources:       sources,
			Reasoning:     e.Reasoning,
		}
	}
	return result
}

func toAPIConfidence(c map[string]*models.FieldConfidenceInfo) *map[string]FieldConfidenceInfo {
	if c == nil {
		return nil
	}
	m := make(map[string]FieldConfidenceInfo, len(c))
	for k, v := range c {
		if v != nil {
			m[k] = FieldConfidenceInfo{Score: v.Score, Reason: v.Reason}
		}
	}
	return &m
}

func toAPISourceJobSummary(j *models.Job) SourceJobSummary {
	summary := SourceJobSummary{
		JobId:     j.JobID,
		Status:    JobStatus(j.Status),
		TotalRows: j.TotalRows,
		CreatedAt: j.CreatedAt,
		StartedAt: j.StartedAt,
	}
	if j.KeyColumns != nil {
		summary.KeyColumns = &j.KeyColumns
	}
	if j.KeyColumnDescription != nil {
		summary.KeyColumnDescription = j.KeyColumnDescription
	}
	if j.ColumnsMetadata != nil {
		apiCols := toAPIColumnMetadataSlice(j.ColumnsMetadata)
		summary.ColumnsMetadata = &apiCols
	}
	return summary
}

func toAPIColumnMetadataSlice(cols []*models.ColumnMetadata) []ColumnMetadata {
	result := make([]ColumnMetadata, len(cols))
	for i, c := range cols {
		result[i] = ColumnMetadata{
			Name:        c.Name,
			Type:        ColumnType(c.Type),
			JobType:     JobType(c.JobType),
			Description: c.Description,
		}
	}
	return result
}

func toAPISourceSummaries(results []*services.SourceWithJobs) []SourceSummary {
	summaries := make([]SourceSummary, len(results))
	for i, r := range results {
		summaries[i] = toAPISourceSummary(r)
	}
	return summaries
}

func toAPISourceSummary(r *services.SourceWithJobs) SourceSummary {
	summary := SourceSummary{
		SourceId:  openapi_types.UUID(r.Source.ID),
		Type:      string(r.Source.Type),
		CreatedAt: r.Source.CreatedAt,
		JobCount:  len(r.Jobs),
		Name:      sourceMetaName(r.Source),
	}
	if len(r.Jobs) > 0 {
		status := JobStatus(r.Jobs[0].Status)
		summary.LatestJobStatus = &status
	}
	return summary
}

func sourceMetaName(s *models.Source) *string {
	if meta, ok := s.Metadata.(*models.CSVSourceMetadata); ok && meta.Name != "" {
		return &meta.Name
	}
	return nil
}

func toAPISourceDetail(source *models.Source, jobs []*models.Job) SourceDetail {
	jobSummaries := make([]SourceJobSummary, len(jobs))
	for i, j := range jobs {
		jobSummaries[i] = toAPISourceJobSummary(j)
	}
	return SourceDetail{
		SourceId:  openapi_types.UUID(source.ID),
		Type:      string(source.Type),
		CreatedAt: source.CreatedAt,
		Jobs:      jobSummaries,
	}
}
