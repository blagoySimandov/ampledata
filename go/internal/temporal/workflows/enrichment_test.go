package workflows

import (
	"testing"

	"github.com/blagoySimandov/ampledata/go/internal/models"
)

func TestMergeBestConfidence(t *testing.T) {
	tests := []struct {
		name      string
		base      map[string]interface{}
		baseConf  map[string]*models.FieldConfidenceInfo
		retry     map[string]interface{}
		retryConf map[string]*models.FieldConfidenceInfo
		wantData  map[string]interface{}
		wantConf  map[string]*models.FieldConfidenceInfo
	}{
		{
			name:     "nil base gets initialized",
			base:     nil,
			baseConf: nil,
			retry:    map[string]interface{}{"a": "val"},
			retryConf: map[string]*models.FieldConfidenceInfo{
				"a": {Score: 0.9, Reason: "good"},
			},
			wantData: map[string]interface{}{"a": "val"},
			wantConf: map[string]*models.FieldConfidenceInfo{
				"a": {Score: 0.9, Reason: "good"},
			},
		},
		{
			name:     "retry with higher confidence wins",
			base:     map[string]interface{}{"a": "old"},
			baseConf: map[string]*models.FieldConfidenceInfo{"a": {Score: 0.5, Reason: "low"}},
			retry:    map[string]interface{}{"a": "new"},
			retryConf: map[string]*models.FieldConfidenceInfo{
				"a": {Score: 0.9, Reason: "high"},
			},
			wantData: map[string]interface{}{"a": "new"},
			wantConf: map[string]*models.FieldConfidenceInfo{
				"a": {Score: 0.9, Reason: "high"},
			},
		},
		{
			name:     "base with higher confidence kept",
			base:     map[string]interface{}{"a": "old"},
			baseConf: map[string]*models.FieldConfidenceInfo{"a": {Score: 0.9, Reason: "high"}},
			retry:    map[string]interface{}{"a": "new"},
			retryConf: map[string]*models.FieldConfidenceInfo{
				"a": {Score: 0.3, Reason: "low"},
			},
			wantData: map[string]interface{}{"a": "old"},
			wantConf: map[string]*models.FieldConfidenceInfo{
				"a": {Score: 0.9, Reason: "high"},
			},
		},
		{
			name:     "new field from retry added",
			base:     map[string]interface{}{"a": "old"},
			baseConf: map[string]*models.FieldConfidenceInfo{"a": {Score: 0.9, Reason: "good"}},
			retry:    map[string]interface{}{"b": "new"},
			retryConf: map[string]*models.FieldConfidenceInfo{
				"b": {Score: 0.7, Reason: "ok"},
			},
			wantData: map[string]interface{}{"a": "old", "b": "new"},
			wantConf: map[string]*models.FieldConfidenceInfo{
				"a": {Score: 0.9, Reason: "good"},
				"b": {Score: 0.7, Reason: "ok"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotData, gotConf := mergeBestConfidence(tt.base, tt.baseConf, tt.retry, tt.retryConf)
			for k, want := range tt.wantData {
				if gotData[k] != want {
					t.Errorf("data[%s] = %v, want %v", k, gotData[k], want)
				}
			}
			for k, want := range tt.wantConf {
				got := gotConf[k]
				if got == nil {
					t.Errorf("conf[%s] is nil, want score=%f", k, want.Score)
				} else if got.Score != want.Score {
					t.Errorf("conf[%s].Score = %f, want %f", k, got.Score, want.Score)
				}
			}
		})
	}
}

func TestMergeSources(t *testing.T) {
	tests := []struct {
		name string
		a, b []string
		want []string
	}{
		{"both empty", nil, nil, nil},
		{"deduplicates", []string{"a", "b"}, []string{"b", "c"}, []string{"a", "b", "c"}},
		{"skips empty strings", []string{"", "a"}, []string{"", "b"}, []string{"a", "b"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergeSources(tt.a, tt.b)
			if len(got) != len(tt.want) {
				t.Errorf("len = %d, want %d", len(got), len(tt.want))
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

func TestMergeDataNotFound(t *testing.T) {
	tests := []struct {
		name     string
		previous []*models.EnrichmentAttempt
		current  map[string]string
		want     map[string]string
	}{
		{
			name:     "empty inputs",
			previous: nil,
			current:  nil,
			want:     map[string]string{},
		},
		{
			name: "previous attempts only",
			previous: []*models.EnrichmentAttempt{
				{DataNotFoundColumns: []string{"col_a", "col_b"}},
			},
			current: nil,
			want: map[string]string{
				"col_a": "Determined not to exist in a previous attempt",
				"col_b": "Determined not to exist in a previous attempt",
			},
		},
		{
			name:     "current only",
			previous: nil,
			current:  map[string]string{"col_x": "No public email on official page"},
			want:     map[string]string{"col_x": "No public email on official page"},
		},
		{
			name: "current overrides previous reason",
			previous: []*models.EnrichmentAttempt{
				{DataNotFoundColumns: []string{"col_a"}},
			},
			current: map[string]string{"col_a": "Confirmed: only contact form available"},
			want:    map[string]string{"col_a": "Confirmed: only contact form available"},
		},
		{
			name: "multiple attempts merged",
			previous: []*models.EnrichmentAttempt{
				{DataNotFoundColumns: []string{"col_a"}},
				{DataNotFoundColumns: []string{"col_b"}},
			},
			current: map[string]string{"col_c": "Not publicly available"},
			want: map[string]string{
				"col_a": "Determined not to exist in a previous attempt",
				"col_b": "Determined not to exist in a previous attempt",
				"col_c": "Not publicly available",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergeDataNotFound(tt.previous, tt.current)
			if len(got) != len(tt.want) {
				t.Errorf("len = %d, want %d (got=%v)", len(got), len(tt.want), got)
				return
			}
			for k, wantV := range tt.want {
				if got[k] != wantV {
					t.Errorf("got[%s] = %q, want %q", k, got[k], wantV)
				}
			}
		})
	}
}
