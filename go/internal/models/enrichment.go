package models

type EnrichmentRequest struct {
	RowKeys         []string          `json:"row_keys"`
	ColumnsMetadata []*ColumnMetadata `json:"columns_metadata"`
	EntityType      *string           `json:"entity_type,omitempty"`
}

type EnrichmentResult struct {
	Key           string                 `json:"key"`
	ExtractedData map[string]interface{} `json:"extracted_data"`
	Sources       []string               `json:"sources"`
	Error         *string                `json:"error,omitempty"`
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
