package models

type (
	ColumnType string
	JobType    string
)

const (
	ColumnTypeString  ColumnType = "string"
	ColumnTypeNumber  ColumnType = "number"
	ColumnTypeBoolean ColumnType = "boolean"
	ColumnTypeDate    ColumnType = "date"
)

const (
	JobTypeEnrichment JobType = "enrichment"
	JobTypeImputation JobType = "imputation"
)

type ColumnMetadata struct {
	Name        string     `json:"name"`
	Type        ColumnType `json:"type"`
	JobType     *string    `json:"job_type,omitempty"`
	Description *string    `json:"description,omitempty"`
}
