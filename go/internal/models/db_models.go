package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type JobDB struct {
	bun.BaseModel `bun:"table:jobs,alias:j"`

	JobID           string            `bun:"job_id,pk" json:"job_id"`
	UserID          string            `bun:"user_id,notnull" json:"user_id"`
	FilePath        string            `bun:"file_path" json:"file_path"`
	KeyColumn       *string           `bun:"key_column" json:"key_column"`
	ColumnsMetadata []*ColumnMetadata `bun:"columns_metadata,type:jsonb" json:"columns_metadata"`
	EntityType      *string           `bun:"entity_type" json:"entity_type"`
	TotalRows       int               `bun:"total_rows,notnull" json:"total_rows"`
	StartedAt       *time.Time        `bun:"started_at" json:"started_at"`
	Status          JobStatus         `bun:"status,notnull,default:'PENDING'" json:"status"`
	CostDollars     int               `bun:"cost_dollars,notnull,default:0" json:"cost_dollars"`
	CostCredits     int               `bun:"cost_credits,notnull,default:0" json:"cost_credits"`
	CreatedAt       time.Time         `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt       time.Time         `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
}

func (j *JobDB) ToJob() *Job {
	return &Job{
		JobID:           j.JobID,
		UserID:          j.UserID,
		FilePath:        j.FilePath,
		KeyColumn:       j.KeyColumn,
		ColumnsMetadata: j.ColumnsMetadata,
		EntityType:      j.EntityType,
		TotalRows:       j.TotalRows,
		StartedAt:       j.StartedAt,
		Status:          j.Status,
		CreatedAt:       j.CreatedAt,
		UpdatedAt:       j.UpdatedAt,
	}
}

func JobFromDomain(job *Job) *JobDB {
	return &JobDB{
		JobID:           job.JobID,
		UserID:          job.UserID,
		FilePath:        job.FilePath,
		KeyColumn:       job.KeyColumn,
		ColumnsMetadata: job.ColumnsMetadata,
		EntityType:      job.EntityType,
		TotalRows:       job.TotalRows,
		StartedAt:       job.StartedAt,
		Status:          job.Status,
		CreatedAt:       job.CreatedAt,
		UpdatedAt:       job.UpdatedAt,
	}
}

type RowStateDB struct {
	bun.BaseModel `bun:"table:row_states,alias:rs"`

	ID            uuid.UUID                       `bun:"id,pk,type:uuid,default:gen_random_uuid()" json:"id"`
	JobID         string                          `bun:"job_id,notnull,unique:job_key" json:"job_id"`
	Job           *JobDB                          `bun:"rel:belongs-to,join:job_id=job_id,on_delete:CASCADE"`
	Key           string                          `bun:"key,notnull,unique:job_key" json:"key"`
	Stage         RowStage                        `bun:"stage,notnull" json:"stage"`
	ExtractedData map[string]interface{}          `bun:"extracted_data,type:jsonb" json:"extracted_data,omitempty"`
	Confidence    map[string]*FieldConfidenceInfo `bun:"confidence,type:jsonb" json:"confidence,omitempty"`
	Sources       []string                        `bun:"sources,type:jsonb" json:"sources,omitempty"`
	Error         *string                         `bun:"error" json:"error,omitempty"`
	CreatedAt     time.Time                       `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt     time.Time                       `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
}

func (r *RowStateDB) ToRowState() *RowState {
	return &RowState{
		Key:           r.Key,
		Stage:         r.Stage,
		ExtractedData: r.ExtractedData,
		Confidence:    r.Confidence,
		Sources:       r.Sources,
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
		ExtractedData: state.ExtractedData,
		Confidence:    state.Confidence,
		Sources:       state.Sources,
		Error:         state.Error,
		CreatedAt:     state.CreatedAt,
		UpdatedAt:     state.UpdatedAt,
	}
}

type UserDB struct {
	bun.BaseModel `bun:"table:users,alias:u"`

	ID               string    `bun:"id,pk" json:"user_id"`
	Email            string    `bun:"email,notnull" json:"email"`
	FirstName        string    `bun:"first_name" json:"first_name"`
	LastName         string    `bun:"last_name" json:"last_name"`
	StripeCustomerID *string   `bun:"stripe_customer_id" json:"stripe_customer_id"`
	TokensUsed       int64     `bun:"tokens_used,notnull,default:0" json:"tokens_used"`
	TokensPurchased  int64     `bun:"tokens_purchased,notnull,default:0" json:"tokens_purchased"`
	CreatedAt        time.Time `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt        time.Time `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
}

func (u *UserDB) ToUser() *User {
	return &User{
		ID:               u.ID,
		Email:            u.Email,
		FirstName:        u.FirstName,
		LastName:         u.LastName,
		StripeCustomerID: u.StripeCustomerID,
		TokensUsed:       u.TokensUsed,
		TokensPurchased:  u.TokensPurchased,
		CreatedAt:        u.CreatedAt,
		UpdatedAt:        u.UpdatedAt,
	}
}

func UserFromDomain(user *User) *UserDB {
	return &UserDB{
		ID:               user.ID,
		Email:            user.Email,
		FirstName:        user.FirstName,
		LastName:         user.LastName,
		StripeCustomerID: user.StripeCustomerID,
		TokensUsed:       user.TokensUsed,
		TokensPurchased:  user.TokensPurchased,
		CreatedAt:        user.CreatedAt,
		UpdatedAt:        user.UpdatedAt,
	}
}
