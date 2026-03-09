package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type SourceType string

const (
	SourceTypeCSVUpload SourceType = "csv_upload"
)

type SourceMetadata interface {
	SourceType() SourceType
}

type CSVSourceMetadata struct {
	FileURI     string `json:"file_uri"`
	ContentType string `json:"content_type"`
}

func (m *CSVSourceMetadata) SourceType() SourceType { return SourceTypeCSVUpload }

type Source struct {
	ID        uuid.UUID      `json:"id"`
	UserID    string         `json:"user_id"`
	Type      SourceType     `json:"type"`
	Metadata  SourceMetadata `json:"metadata"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

type SourceDB struct {
	bun.BaseModel `bun:"table:sources,alias:src"`

	ID        uuid.UUID       `bun:"id,pk,type:uuid,default:gen_random_uuid()" json:"id"`
	UserID    string          `bun:"user_id,notnull" json:"user_id"`
	Type      SourceType      `bun:"type,notnull" json:"type"`
	Metadata  json.RawMessage `bun:"metadata,type:jsonb" json:"metadata"`
	CreatedAt time.Time       `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt time.Time       `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
}

func (s *SourceDB) ParseMetadata() (SourceMetadata, error) {
	switch s.Type {
	case SourceTypeCSVUpload:
		var m CSVSourceMetadata
		if err := json.Unmarshal(s.Metadata, &m); err != nil {
			return nil, err
		}
		return &m, nil
	default:
		return nil, fmt.Errorf("unknown source type: %s", s.Type)
	}
}

func (s *SourceDB) ToSource() (*Source, error) {
	meta, err := s.ParseMetadata()
	if err != nil {
		return nil, fmt.Errorf("failed to parse source metadata: %w", err)
	}
	return &Source{
		ID:        s.ID,
		UserID:    s.UserID,
		Type:      s.Type,
		Metadata:  meta,
		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
	}, nil
}
