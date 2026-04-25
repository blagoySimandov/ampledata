package sheets

import (
	"context"
	"fmt"

	"github.com/blagoySimandov/ampledata/go/internal/gcs"
	"github.com/blagoySimandov/ampledata/go/internal/googleoauth"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	googlesheets "google.golang.org/api/sheets/v4"
)

type Spreadsheet struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	ModifiedTime string `json:"modified_time"`
}

type SheetTab struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	RowCount int64  `json:"row_count"`
}

type Client struct {
	oauthSvc *googleoauth.Service
}

func NewClient(oauthSvc *googleoauth.Service) *Client {
	return &Client{oauthSvc: oauthSvc}
}

func (c *Client) sheetsService(ctx context.Context, userID string) (*googlesheets.Service, error) {
	httpClient, err := c.oauthSvc.GetOAuthClient(ctx, userID)
	if err != nil {
		return nil, err
	}
	return googlesheets.NewService(ctx, option.WithHTTPClient(httpClient))
}

func (c *Client) driveService(ctx context.Context, userID string) (*drive.Service, error) {
	httpClient, err := c.oauthSvc.GetOAuthClient(ctx, userID)
	if err != nil {
		return nil, err
	}
	return drive.NewService(ctx, option.WithHTTPClient(httpClient))
}

func (c *Client) ListSpreadsheets(ctx context.Context, userID string) ([]*Spreadsheet, error) {
	svc, err := c.driveService(ctx, userID)
	if err != nil {
		return nil, err
	}
	files, err := svc.Files.List().
		Q("mimeType='application/vnd.google-apps.spreadsheet' and trashed=false").
		Fields("files(id,name,modifiedTime)").
		OrderBy("modifiedTime desc").
		PageSize(100).
		Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list spreadsheets: %w", err)
	}
	result := make([]*Spreadsheet, len(files.Files))
	for i, f := range files.Files {
		result[i] = &Spreadsheet{ID: f.Id, Name: f.Name, ModifiedTime: f.ModifiedTime}
	}
	return result, nil
}

func (c *Client) ListSheetTabs(ctx context.Context, userID, spreadsheetID string) ([]*SheetTab, error) {
	svc, err := c.sheetsService(ctx, userID)
	if err != nil {
		return nil, err
	}
	sp, err := svc.Spreadsheets.Get(spreadsheetID).Fields("sheets(properties(sheetId,title,gridProperties))").Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get spreadsheet: %w", err)
	}
	result := make([]*SheetTab, len(sp.Sheets))
	for i, sh := range sp.Sheets {
		result[i] = &SheetTab{
			ID:       sh.Properties.SheetId,
			Name:     sh.Properties.Title,
			RowCount: sh.Properties.GridProperties.RowCount,
		}
	}
	return result, nil
}

func (c *Client) ReadSheetData(ctx context.Context, userID, spreadsheetID, sheetName string) (*gcs.CSVResult, error) {
	svc, err := c.sheetsService(ctx, userID)
	if err != nil {
		return nil, err
	}
	resp, err := svc.Spreadsheets.Values.Get(spreadsheetID, sheetName).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to read sheet data: %w", err)
	}
	return parseSheetValues(resp.Values), nil
}

func parseSheetValues(values [][]interface{}) *gcs.CSVResult {
	if len(values) == 0 {
		return &gcs.CSVResult{}
	}
	headers := toStringSlice(values[0])
	rows := make([][]string, len(values)-1)
	for i, row := range values[1:] {
		rows[i] = toStringSlice(row)
	}
	return &gcs.CSVResult{Headers: headers, Rows: rows}
}

func toStringSlice(row []interface{}) []string {
	result := make([]string, len(row))
	for i, v := range row {
		if v != nil {
			result[i] = fmt.Sprintf("%v", v)
		}
	}
	return result
}

func (c *Client) WriteResults(ctx context.Context, userID, spreadsheetID, sheetName string, existing *gcs.CSVResult, keyColumns []string, enrichedCols []string, dataByKey map[string]map[string]interface{}) error {
	svc, err := c.sheetsService(ctx, userID)
	if err != nil {
		return err
	}
	values := buildWriteValues(existing, keyColumns, enrichedCols, dataByKey)
	if len(values) == 0 {
		return nil
	}
	startCol := columnLetter(len(existing.Headers) + 1)
	endCol := columnLetter(len(existing.Headers) + len(enrichedCols))
	rangeStr := fmt.Sprintf("%s!%s1:%s%d", sheetName, startCol, endCol, len(values))
	_, err = svc.Spreadsheets.Values.Update(spreadsheetID, rangeStr, &googlesheets.ValueRange{
		Values: values,
	}).ValueInputOption("RAW").Do()
	return err
}

func buildWriteValues(existing *gcs.CSVResult, keyColumns []string, enrichedCols []string, dataByKey map[string]map[string]interface{}) [][]interface{} {
	header := make([]interface{}, len(enrichedCols))
	for i, col := range enrichedCols {
		header[i] = col
	}
	rows := [][]interface{}{header}
	for _, row := range existing.Rows {
		key := buildKey(existing.Headers, row, keyColumns)
		enriched := dataByKey[key]
		rowVals := make([]interface{}, len(enrichedCols))
		for i, col := range enrichedCols {
			if v, ok := enriched[col]; ok && v != nil {
				rowVals[i] = fmt.Sprintf("%v", v)
			}
		}
		rows = append(rows, rowVals)
	}
	return rows
}

func buildKey(headers []string, row []string, keyColumns []string) string {
	headerIdx := map[string]int{}
	for i, h := range headers {
		headerIdx[h] = i
	}
	result := ""
	for i, col := range keyColumns {
		if i > 0 {
			result += "||"
		}
		if idx, ok := headerIdx[col]; ok && idx < len(row) {
			result += row[idx]
		}
	}
	return result
}

func columnLetter(n int) string {
	result := ""
	for n > 0 {
		n--
		result = string(rune('A'+n%26)) + result
		n /= 26
	}
	return result
}
