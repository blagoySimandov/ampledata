package models

import (
	"time"
)

type RowStage string

const (
	StagePending      RowStage = "PENDING"
	StageSerpFetched  RowStage = "SERP_FETCHED"
	StageDecisionMade RowStage = "DECISION_MADE"
	StageCrawled      RowStage = "CRAWLED"
	StageEnriched     RowStage = "ENRICHED"
	StageCompleted    RowStage = "COMPLETED"
	StageFailed       RowStage = "FAILED"
	StageCancelled    RowStage = "CANCELLED"
)

type JobStatus string

const (
	JobStatusPending   JobStatus = "PENDING"
	JobStatusRunning   JobStatus = "RUNNING"
	JobStatusPaused    JobStatus = "PAUSED"
	JobStatusCancelled JobStatus = "CANCELLED"
	JobStatusCompleted JobStatus = "COMPLETED"
)

type Job struct {
	JobID           string            `json:"job_id"`
	UserID          string            `json:"user_id"`
	FilePath        string            `json:"file_path"`
	KeyColumn       *string           `json:"key_column"`
	ColumnsMetadata []*ColumnMetadata `json:"columns_metadata"`
	EntityType      *string           `json:"entity_type"`
	TotalRows       int               `json:"total_rows"`
	StartedAt       *time.Time        `json:"started_at"`
	Status          JobStatus         `json:"status"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
}

type SerpData struct {
	Queries []string               `json:"queries"`
	Results []*GoogleSearchResults `json:"results"`
}

type Decision struct {
	URLsToCrawl    []string               `json:"urls_to_crawl"`
	ExtractedData  map[string]interface{} `json:"extracted_data,omitempty"`
	Reasoning      string                 `json:"reasoning"`
	MissingColumns []string               `json:"missing_columns"`
}

type CrawlResults struct {
	Content *string  `json:"content"`
	Sources []string `json:"sources"`
}

type FieldConfidenceInfo struct {
	Score  float64 `json:"score"`
	Reason string  `json:"reason"`
}

type RowState struct {
	Key           string                          `json:"key"`
	Stage         RowStage                        `json:"stage"`
	ExtractedData map[string]interface{}          `json:"extracted_data,omitempty"`
	Confidence    map[string]*FieldConfidenceInfo `json:"confidence,omitempty"`
	Sources       []string                        `json:"sources,omitempty"`
	Error         *string                         `json:"error,omitempty"`
	CreatedAt     time.Time                       `json:"created_at"`
	UpdatedAt     time.Time                       `json:"updated_at"`
}

type StateUpdate struct {
	ExtractedData map[string]interface{}          `json:"extracted_data,omitempty"`
	Confidence    map[string]*FieldConfidenceInfo `json:"confidence,omitempty"`
	Sources       []string                        `json:"sources,omitempty"`
	Error         *string                         `json:"error,omitempty"`
}

func (s *RowState) ApplyUpdate(u *StateUpdate) {
	if u.ExtractedData != nil {
		s.ExtractedData = u.ExtractedData
	}
	if u.Confidence != nil {
		s.Confidence = u.Confidence
	}
	if u.Sources != nil {
		s.Sources = u.Sources
	}
	if u.Error != nil {
		s.Error = u.Error
		s.Stage = StageFailed
	}
}

type JobProgress struct {
	JobID       string           `json:"job_id"`
	TotalRows   int              `json:"total_rows"`
	RowsByStage map[RowStage]int `json:"rows_by_stage"`
	StartedAt   time.Time        `json:"started_at"`
	Status      JobStatus        `json:"status"`
}
