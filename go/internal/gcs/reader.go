package gcs

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"

	"cloud.google.com/go/storage"
)

type CSVReader struct {
	client     *storage.Client
	bucketName string
}

func NewCSVReader(bucketName string) (*CSVReader, error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCS client: %w", err)
	}
	return &CSVReader{
		client:     client,
		bucketName: bucketName,
	}, nil
}

func (r *CSVReader) Close() error {
	return r.client.Close()
}

type CSVResult struct {
	Headers []string
	Rows    [][]string
}

func (r *CSVReader) ReadCSV(ctx context.Context, objectName string) (*CSVResult, error) {
	bucket := r.client.Bucket(r.bucketName)
	obj := bucket.Object(objectName)

	reader, err := obj.NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create object reader: %w", err)
	}
	defer reader.Close()

	csvReader := csv.NewReader(reader)

	headers, err := csvReader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV headers: %w", err)
	}

	var rows [][]string
	for {
		row, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read CSV row: %w", err)
		}
		rows = append(rows, row)
	}

	return &CSVResult{
		Headers: headers,
		Rows:    rows,
	}, nil
}

func (r *CSVReader) ExtractColumn(result *CSVResult, columnName string) ([]string, error) {
	columnIndex := -1
	for i, header := range result.Headers {
		if header == columnName {
			columnIndex = i
			break
		}
	}

	if columnIndex == -1 {
		return nil, fmt.Errorf("column '%s' not found in CSV headers: %v", columnName, result.Headers)
	}

	values := make([]string, 0, len(result.Rows))
	for _, row := range result.Rows {
		if columnIndex < len(row) {
			values = append(values, row[columnIndex])
		}
	}

	return values, nil
}

func (r *CSVReader) ReadColumnFromFile(ctx context.Context, objectName string, columnName string) ([]string, error) {
	result, err := r.ReadCSV(ctx, objectName)
	if err != nil {
		return nil, err
	}
	return r.ExtractColumn(result, columnName)
}

const CompositeKeyDelimiter = "||"

func (r *CSVReader) ExtractCompositeKey(result *CSVResult, columnNames []string) ([]string, error) {
	if len(columnNames) == 0 {
		return nil, fmt.Errorf("at least one column name is required")
	}

	columnIndices := make([]int, len(columnNames))
	for i, columnName := range columnNames {
		columnIndex := -1
		for j, header := range result.Headers {
			if header == columnName {
				columnIndex = j
				break
			}
		}
		if columnIndex == -1 {
			return nil, fmt.Errorf("column '%s' not found in CSV headers: %v", columnName, result.Headers)
		}
		columnIndices[i] = columnIndex
	}

	values := make([]string, 0, len(result.Rows))
	for _, row := range result.Rows {
		keyParts := make([]string, len(columnIndices))
		for i, columnIndex := range columnIndices {
			if columnIndex < len(row) {
				keyParts[i] = row[columnIndex]
			}
		}
		compositeKey := joinKeyParts(keyParts)
		values = append(values, compositeKey)
	}

	return values, nil
}

func joinKeyParts(parts []string) string {
	if len(parts) == 1 {
		return parts[0]
	}
	result := parts[0]
	for i := 1; i < len(parts); i++ {
		result += CompositeKeyDelimiter + parts[i]
	}
	return result
}

func (r *CSVReader) ReadCompositeKeyFromFile(ctx context.Context, objectName string, columnNames []string) ([]string, error) {
	result, err := r.ReadCSV(ctx, objectName)
	if err != nil {
		return nil, err
	}
	return r.ExtractCompositeKey(result, columnNames)
}

func (r *CSVReader) ReadCompositeKeyFromFileFiltered(ctx context.Context, objectName string, keyColumns []string, filterColumns []string) ([]string, error) {
	result, err := r.ReadCSV(ctx, objectName)
	if err != nil {
		return nil, err
	}
	return ExtractCompositeKeyFiltered(result, keyColumns, filterColumns)
}

func ExtractCompositeKeyFiltered(result *CSVResult, keyColumns []string, filterColumns []string) ([]string, error) {
	keyIndices, err := findColumnIndices(result.Headers, keyColumns)
	if err != nil {
		return nil, err
	}

	filterIndices, err := findColumnIndices(result.Headers, filterColumns)
	if err != nil {
		return nil, err
	}

	values := make([]string, 0, len(result.Rows))
	for _, row := range result.Rows {
		if !hasAnyEmptyColumn(row, filterIndices) {
			continue
		}
		keyParts := make([]string, len(keyIndices))
		for i, idx := range keyIndices {
			if idx < len(row) {
				keyParts[i] = row[idx]
			}
		}
		values = append(values, joinKeyParts(keyParts))
	}

	return values, nil
}

func findColumnIndices(headers []string, columnNames []string) ([]int, error) {
	indices := make([]int, len(columnNames))
	for i, name := range columnNames {
		idx := -1
		for j, header := range headers {
			if header == name {
				idx = j
				break
			}
		}
		if idx == -1 {
			return nil, fmt.Errorf("column '%s' not found in CSV headers: %v", name, headers)
		}
		indices[i] = idx
	}
	return indices, nil
}

func hasAnyEmptyColumn(row []string, indices []int) bool {
	for _, idx := range indices {
		if idx >= len(row) || row[idx] == "" {
			return true
		}
	}
	return false
}
