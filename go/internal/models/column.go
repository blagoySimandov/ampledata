package models

type ColumnType string

const (
	ColumnTypeString  ColumnType = "string"
	ColumnTypeNumber  ColumnType = "number"
	ColumnTypeBoolean ColumnType = "boolean"
	ColumnTypeDate    ColumnType = "date"
)

type ColumnMetadata struct {
	Name        string      `json:"name"`
	Type        ColumnType  `json:"type"`
	Description *string     `json:"description,omitempty"`
}
