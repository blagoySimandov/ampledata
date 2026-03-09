package api

import (
	"time"

	"github.com/blagoySimandov/ampledata/go/internal/models"
)

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

func toAPIJobSummary(j *models.JobSummary) JobSummary {
	return JobSummary{
		JobId:     j.JobID,
		Status:    JobStatus(j.Status),
		TotalRows: j.TotalRows,
		FilePath:  j.FilePath,
		CreatedAt: j.CreatedAt,
		StartedAt: j.StartedAt,
	}
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
		Key:           r.Key,
		ExtractedData: r.ExtractedData,
		Confidence:    toAPIConfidence(r.Confidence),
		Sources:       r.Sources,
		Error:         r.Error,
	}
}

func toAPIRowProgressItem(r *models.RowProgressItem) RowProgressItem {
	item := RowProgressItem{
		Key:        r.Key,
		Stage:      RowStage(r.Stage),
		Confidence: toAPIConfidence(r.Confidence),
		Error:      r.Error,
		UpdatedAt:  r.UpdatedAt,
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
