package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type JobDB struct {
	bun.BaseModel `bun:"table:jobs,alias:j"`

	JobID     string    `bun:"job_id,pk" json:"job_id"`
	TotalRows int       `bun:"total_rows,notnull" json:"total_rows"`
	StartedAt time.Time `bun:"started_at,notnull,default:current_timestamp" json:"started_at"`
	Status    JobStatus `bun:"status,notnull,default:'RUNNING'" json:"status"`
	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt time.Time `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
}

type RowStateDB struct {
	bun.BaseModel `bun:"table:row_states,alias:rs"`

	ID            uuid.UUID              `bun:"id,pk,type:uuid,default:gen_random_uuid()" json:"id"`
	JobID         string                 `bun:"job_id,notnull,unique:job_key" json:"job_id"`
	Job           *JobDB                 `bun:"rel:belongs-to,join:job_id=job_id,on_delete:CASCADE"`
	Key           string                 `bun:"key,notnull,unique:job_key" json:"key"`
	Stage         RowStage               `bun:"stage,notnull" json:"stage"`
	SerpData      *SerpData              `bun:"serp_data,type:jsonb" json:"serp_data,omitempty"`
	Decision      *Decision              `bun:"decision,type:jsonb" json:"decision,omitempty"`
	CrawlResults  *CrawlResults          `bun:"crawl_results,type:jsonb" json:"crawl_results,omitempty"`
	ExtractedData map[string]interface{} `bun:"extracted_data,type:jsonb" json:"extracted_data,omitempty"`
	Error         *string                `bun:"error" json:"error,omitempty"`
	CreatedAt     time.Time              `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt     time.Time              `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
}

func (r *RowStateDB) ToRowState() *RowState {
	return &RowState{
		Key:           r.Key,
		Stage:         r.Stage,
		SerpData:      r.SerpData,
		Decision:      r.Decision,
		CrawlResults:  r.CrawlResults,
		ExtractedData: r.ExtractedData,
		Error:         r.Error,
		CreatedAt:     r.CreatedAt,
		UpdatedAt:     r.UpdatedAt,
	}
}

func RowStateFromApp(jobID string, state *RowState) *RowStateDB {
	return &RowStateDB{
		JobID:         jobID,
		Key:           state.Key,
		Stage:         state.Stage,
		SerpData:      state.SerpData,
		Decision:      state.Decision,
		CrawlResults:  state.CrawlResults,
		ExtractedData: state.ExtractedData,
		Error:         state.Error,
		CreatedAt:     state.CreatedAt,
		UpdatedAt:     state.UpdatedAt,
	}
}
