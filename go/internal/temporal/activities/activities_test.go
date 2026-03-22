package activities

import (
	"context"
	"sort"
	"testing"

	"github.com/blagoySimandov/ampledata/go/internal/models"
)

// TestCollectPreviouslyCrawledURLs verifies that URLs are collected from all
// previous enrichment attempts, deduplicated, and empty strings are skipped.
func TestCollectPreviouslyCrawledURLs(t *testing.T) {
	tests := []struct {
		name     string
		attempts []*models.EnrichmentAttempt
		want     []string
	}{
		{
			name:     "nil attempts",
			attempts: nil,
			want:     nil,
		},
		{
			name: "single attempt",
			attempts: []*models.EnrichmentAttempt{
				{CrawledURLs: []string{"https://a.com", "https://b.com"}},
			},
			want: []string{"https://a.com", "https://b.com"},
		},
		{
			name: "multiple attempts with duplicates",
			attempts: []*models.EnrichmentAttempt{
				{CrawledURLs: []string{"https://a.com", "https://b.com"}},
				{CrawledURLs: []string{"https://b.com", "https://c.com"}},
			},
			want: []string{"https://a.com", "https://b.com", "https://c.com"},
		},
		{
			name: "skips empty strings",
			attempts: []*models.EnrichmentAttempt{
				{CrawledURLs: []string{"", "https://a.com", ""}},
			},
			want: []string{"https://a.com"},
		},
		{
			name: "attempt with nil CrawledURLs",
			attempts: []*models.EnrichmentAttempt{
				{CrawledURLs: nil},
				{CrawledURLs: []string{"https://a.com"}},
			},
			want: []string{"https://a.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := collectPreviouslyCrawledURLs(tt.attempts)
			if len(got) != len(tt.want) {
				t.Errorf("len = %d, want %d (got=%v)", len(got), len(tt.want))
				return
			}
			for i, v := range got {
				if v != tt.want[i] {
					t.Errorf("got[%d] = %q, want %q", i, v, tt.want[i])
				}
			}
		})
	}
}

// TestAnalyzeFeedback_DataNotFoundExclusion verifies that columns flagged in
// DataNotFound are excluded from both MissingColumns and LowConfidenceColumns,
// preventing wasteful retries for data that genuinely doesn't exist publicly.
func TestAnalyzeFeedback_DataNotFoundExclusion(t *testing.T) {
	a := &Activities{}

	columns := []*models.ColumnMetadata{
		{Name: "contact_email", Type: "string"},
		{Name: "founded_year", Type: "number"},
		{Name: "headquarters", Type: "string"},
		{Name: "annual_revenue", Type: "number"},
	}

	tests := []struct {
		name                 string
		extractedData        map[string]interface{}
		confidence           map[string]*models.FieldConfidenceInfo
		dataNotFound         map[string]string
		wantNeedsFeedback    bool
		wantMissing          []string
		wantLowConf          []string
	}{
		{
			name: "data_not_found columns excluded from missing",
			extractedData: map[string]interface{}{
				"founded_year": 2021,
				"headquarters": "Portland, OR",
			},
			confidence: map[string]*models.FieldConfidenceInfo{
				"founded_year": {Score: 0.95, Reason: "explicit"},
				"headquarters": {Score: 0.9, Reason: "explicit"},
			},
			dataNotFound: map[string]string{
				"contact_email":  "Company uses contact form only",
				"annual_revenue": "Not publicly available",
			},
			wantNeedsFeedback: false,
			wantMissing:       []string{},
			wantLowConf:       []string{},
		},
		{
			name: "data_not_found columns excluded from low confidence",
			extractedData: map[string]interface{}{
				"contact_email":  nil,
				"founded_year":   2021,
				"headquarters":   "Portland, OR",
				"annual_revenue": nil,
			},
			confidence: map[string]*models.FieldConfidenceInfo{
				"contact_email":  {Score: 0.0, Reason: "not found"},
				"founded_year":   {Score: 0.95, Reason: "explicit"},
				"headquarters":   {Score: 0.9, Reason: "explicit"},
				"annual_revenue": {Score: 0.0, Reason: "not found"},
			},
			dataNotFound: map[string]string{
				"contact_email":  "No public email",
				"annual_revenue": "Private company",
			},
			wantNeedsFeedback: false,
			wantMissing:       []string{},
			wantLowConf:       []string{},
		},
		{
			name: "non-data-not-found missing columns still trigger feedback",
			extractedData: map[string]interface{}{
				"founded_year": 2021,
			},
			confidence: map[string]*models.FieldConfidenceInfo{
				"founded_year": {Score: 0.95, Reason: "explicit"},
			},
			dataNotFound: map[string]string{
				"contact_email": "No public email",
			},
			wantNeedsFeedback: true,
			wantMissing:       []string{"headquarters", "annual_revenue"},
			wantLowConf:       []string{},
		},
		{
			name: "empty data_not_found behaves like before",
			extractedData: map[string]interface{}{
				"founded_year": 2021,
				"headquarters": "Portland",
			},
			confidence: map[string]*models.FieldConfidenceInfo{
				"founded_year": {Score: 0.95, Reason: "explicit"},
				"headquarters": {Score: 0.5, Reason: "partial"},
			},
			dataNotFound:      map[string]string{},
			wantNeedsFeedback: true,
			wantMissing:       []string{"contact_email", "annual_revenue"},
			wantLowConf:       []string{"headquarters"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := a.AnalyzeFeedback(context.Background(), FeedbackAnalysisInput{
				JobID:           "test-job",
				RowKey:          "test-row",
				ExtractedData:   tt.extractedData,
				Confidence:      tt.confidence,
				ColumnsMetadata: columns,
				DataNotFound:    tt.dataNotFound,
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if output.NeedsFeedback != tt.wantNeedsFeedback {
				t.Errorf("NeedsFeedback = %v, want %v", output.NeedsFeedback, tt.wantNeedsFeedback)
			}

			sort.Strings(output.MissingColumns)
			sort.Strings(tt.wantMissing)
			if len(output.MissingColumns) != len(tt.wantMissing) {
				t.Errorf("MissingColumns = %v, want %v", output.MissingColumns, tt.wantMissing)
			} else {
				for i, v := range output.MissingColumns {
					if v != tt.wantMissing[i] {
						t.Errorf("MissingColumns[%d] = %q, want %q", i, v, tt.wantMissing[i])
					}
				}
			}

			sort.Strings(output.LowConfidenceColumns)
			sort.Strings(tt.wantLowConf)
			if len(output.LowConfidenceColumns) != len(tt.wantLowConf) {
				t.Errorf("LowConfidenceColumns = %v, want %v", output.LowConfidenceColumns, tt.wantLowConf)
			} else {
				for i, v := range output.LowConfidenceColumns {
					if v != tt.wantLowConf[i] {
						t.Errorf("LowConfidenceColumns[%d] = %q, want %q", i, v, tt.wantLowConf[i])
					}
				}
			}
		})
	}
}

// TestFilterMissingColumnsMetadata verifies that only columns named in the
// missing list are returned, preserving order and ignoring nonexistent names.
func TestFilterMissingColumnsMetadata(t *testing.T) {
	allColumns := []*models.ColumnMetadata{
		{Name: "a", Type: "string"},
		{Name: "b", Type: "number"},
		{Name: "c", Type: "boolean"},
	}

	tests := []struct {
		name    string
		missing []string
		want    []string
	}{
		{"all missing", []string{"a", "b", "c"}, []string{"a", "b", "c"}},
		{"some missing", []string{"a", "c"}, []string{"a", "c"}},
		{"none missing", []string{}, []string{}},
		{"nonexistent column", []string{"x"}, []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterMissingColumnsMetadata(tt.missing, allColumns)
			if len(got) != len(tt.want) {
				t.Errorf("len = %d, want %d", len(got), len(tt.want))
				return
			}
			for i, col := range got {
				if col.Name != tt.want[i] {
					t.Errorf("got[%d].Name = %q, want %q", i, col.Name, tt.want[i])
				}
			}
		})
	}
}
