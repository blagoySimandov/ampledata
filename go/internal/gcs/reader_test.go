package gcs

import (
	"testing"
)

func TestHasAnyEmptyColumn(t *testing.T) {
	tests := []struct {
		name    string
		row     []string
		indices []int
		want    bool
	}{
		{"all filled", []string{"a", "b", "c"}, []int{1, 2}, false},
		{"one empty", []string{"a", "", "c"}, []int{1, 2}, true},
		{"all empty", []string{"a", "", ""}, []int{1, 2}, true},
		{"index out of range", []string{"a"}, []int{0, 2}, true},
		{"no indices", []string{"a", "b"}, []int{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasAnyEmptyColumn(tt.row, tt.indices)
			if got != tt.want {
				t.Errorf("hasAnyEmptyColumn() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractCompositeKeyFiltered(t *testing.T) {
	result := &CSVResult{
		Headers: []string{"id", "name", "email", "phone"},
		Rows: [][]string{
			{"1", "Alice", "alice@example.com", "555-0001"},
			{"2", "Bob", "", "555-0002"},
			{"3", "Charlie", "charlie@example.com", ""},
			{"4", "Diana", "diana@example.com", "555-0004"},
		},
	}

	keys, err := ExtractCompositeKeyFiltered(result, []string{"id"}, []string{"email", "phone"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{"2", "3"}
	if len(keys) != len(expected) {
		t.Fatalf("got %d keys, want %d", len(keys), len(expected))
	}
	for i, key := range keys {
		if key != expected[i] {
			t.Errorf("key[%d] = %q, want %q", i, key, expected[i])
		}
	}
}

func TestExtractCompositeKeyFiltered_CompositeKey(t *testing.T) {
	result := &CSVResult{
		Headers: []string{"first", "last", "email"},
		Rows: [][]string{
			{"John", "Doe", "john@example.com"},
			{"Jane", "Doe", ""},
		},
	}

	keys, err := ExtractCompositeKeyFiltered(result, []string{"first", "last"}, []string{"email"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(keys) != 1 || keys[0] != "Jane"+CompositeKeyDelimiter+"Doe" {
		t.Errorf("got %v, want [%q]", keys, "Jane"+CompositeKeyDelimiter+"Doe")
	}
}

func TestExtractCompositeKeyFiltered_AllFilled(t *testing.T) {
	result := &CSVResult{
		Headers: []string{"id", "email"},
		Rows: [][]string{
			{"1", "a@b.com"},
			{"2", "c@d.com"},
		},
	}

	keys, err := ExtractCompositeKeyFiltered(result, []string{"id"}, []string{"email"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(keys) != 0 {
		t.Errorf("expected 0 keys, got %d", len(keys))
	}
}

func TestExtractCompositeKeyFiltered_InvalidColumn(t *testing.T) {
	result := &CSVResult{
		Headers: []string{"id", "email"},
		Rows:    [][]string{{"1", "a@b.com"}},
	}

	_, err := ExtractCompositeKeyFiltered(result, []string{"id"}, []string{"nonexistent"})
	if err == nil {
		t.Error("expected error for invalid filter column")
	}

	_, err = ExtractCompositeKeyFiltered(result, []string{"nonexistent"}, []string{"email"})
	if err == nil {
		t.Error("expected error for invalid key column")
	}
}
