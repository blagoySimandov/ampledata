package models

import (
	"time"

	"github.com/google/uuid"
)

type EnrichmentRequest struct {
	RowKeys              []string          `json:"row_keys"`
	ColumnsMetadata      []*ColumnMetadata `json:"columns_metadata"`
	KeyColumnDescription *string           `json:"key_column_description,omitempty"`
}

type StartJobRequest struct {
	KeyColumns           []string          `json:"key_columns"`
	ColumnsMetadata      []*ColumnMetadata `json:"columns_metadata"`
	KeyColumnDescription *string           `json:"key_column_description,omitempty"`
}

type StartJobResponse struct {
	JobID   string `json:"job_id"`
	Message string `json:"message"`
}

type JobSummary struct {
	JobID     string     `json:"job_id"`
	Status    JobStatus  `json:"status"`
	TotalRows int        `json:"total_rows"`
	SourceID  uuid.UUID  `json:"source_id"`
	CreatedAt time.Time  `json:"created_at"`
	StartedAt *time.Time `json:"started_at,omitempty"`
}

type JobListResponse struct {
	Jobs       []*JobSummary `json:"jobs"`
	TotalCount int           `json:"total_count"`
}

type EnrichmentResult struct {
	Key           string                          `json:"key"`
	ExtractedData map[string]interface{}          `json:"extracted_data"`
	Confidence    map[string]*FieldConfidenceInfo `json:"confidence,omitempty"`
	Sources       []string                        `json:"sources"`
	Error         *string                         `json:"error,omitempty"`
}

type EnrichmentResponse struct {
	JobID   string `json:"job_id"`
	Message string `json:"message"`
}

type JobProgressResponse struct {
	JobID       string           `json:"job_id"`
	TotalRows   int              `json:"total_rows"`
	RowsByStage map[RowStage]int `json:"rows_by_stage"`
	StartedAt   string           `json:"started_at"`
	Status      JobStatus        `json:"status"`
}

func ToJobSummary(job *Job) *JobSummary {
	return &JobSummary{
		JobID:     job.JobID,
		Status:    job.Status,
		TotalRows: job.TotalRows,
		SourceID:  job.SourceID,
		CreatedAt: job.CreatedAt,
		StartedAt: job.StartedAt,
	}
}

type SelectKeyRequest struct {
	JobID           string            `json:"job_id"`
	ColumnsMetadata []*ColumnMetadata `json:"columns_metadata,omitempty"`
}

type SelectKeyResponse struct {
	SelectedKey string   `json:"selected_key"`
	AllKeys     []string `json:"all_keys"`
	Reasoning   string   `json:"reasoning"`
}

func ToEnrichmentResult(row *RowState) *EnrichmentResult {
	sources := row.Sources
	if sources == nil {
		sources = []string{}
	}

	return &EnrichmentResult{
		Key:           row.Key,
		ExtractedData: row.ExtractedData,
		Confidence:    row.Confidence,
		Sources:       sources,
		Error:         row.Error,
	}
}

type RowProgressItem struct {
	Key           string                          `json:"key"`
	Stage         RowStage                        `json:"stage"`
	ExtractedData map[string]interface{}          `json:"extracted_data,omitempty"`
	Confidence    map[string]*FieldConfidenceInfo `json:"confidence,omitempty"`
	Sources       []string                        `json:"sources,omitempty"`
	Error         *string                         `json:"error,omitempty"`
	UpdatedAt     time.Time                       `json:"updated_at"`
}

type PaginationInfo struct {
	Total   int  `json:"total"`
	Offset  int  `json:"offset"`
	Limit   int  `json:"limit"`
	HasMore bool `json:"has_more"`
}

type RowsProgressResponse struct {
	Rows       []*RowProgressItem `json:"rows"`
	Pagination *PaginationInfo    `json:"pagination"`
}

func ToRowProgressItem(row *RowState) *RowProgressItem {
	sources := row.Sources
	if sources == nil {
		sources = []string{}
	}

	return &RowProgressItem{
		Key:           row.Key,
		Stage:         row.Stage,
		ExtractedData: row.ExtractedData,
		Confidence:    row.Confidence,
		Sources:       sources,
		Error:         row.Error,
		UpdatedAt:     row.UpdatedAt,
	}
}
