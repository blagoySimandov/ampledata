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
