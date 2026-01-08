package feedback

import (
	"maps"
	"sync"

	"github.com/blagoySimandov/ampledata/go/internal/models"
)

type Attempt struct {
	Number        int
	TargetColumns []string
	PatternsUsed  []string
	URLsCrawled   []string
	Results       map[string]any
	Confidences   map[string]float64
	Assessment    *QualityAssessment
}

type AttemptHistory struct {
	RowKey   string
	Attempts []Attempt
}

func NewAttemptHistory(rowKey string) *AttemptHistory {
	return &AttemptHistory{
		RowKey:   rowKey,
		Attempts: make([]Attempt, 0),
	}
}

func (h *AttemptHistory) Record(attempt Attempt) {
	h.Attempts = append(h.Attempts, attempt)
}

func (h *AttemptHistory) Count() int {
	return len(h.Attempts)
}

func (h *AttemptHistory) LastAttempt() *Attempt {
	if len(h.Attempts) == 0 {
		return nil
	}
	return &h.Attempts[len(h.Attempts)-1]
}

func (h *AttemptHistory) LastWeakColumns() []WeakColumn {
	last := h.LastAttempt()
	if last == nil || last.Assessment == nil {
		return nil
	}
	return last.Assessment.WeakColumns
}

func (h *AttemptHistory) AllPatternsUsed() []string {
	seen := make(map[string]struct{})
	var patterns []string
	for _, a := range h.Attempts {
		for _, p := range a.PatternsUsed {
			if _, ok := seen[p]; !ok {
				seen[p] = struct{}{}
				patterns = append(patterns, p)
			}
		}
	}
	return patterns
}

func (h *AttemptHistory) AllURLsCrawled() []string {
	seen := make(map[string]struct{})
	var urls []string
	for _, a := range h.Attempts {
		for _, u := range a.URLsCrawled {
			if _, ok := seen[u]; !ok {
				seen[u] = struct{}{}
				urls = append(urls, u)
			}
		}
	}
	return urls
}

func (h *AttemptHistory) BestConfidences() map[string]float64 {
	best := make(map[string]float64)
	for _, a := range h.Attempts {
		for col, conf := range a.Confidences {
			if conf > best[col] {
				best[col] = conf
			}
		}
	}
	return best
}

func (h *AttemptHistory) BestResults() map[string]any {
	best := make(map[string]any)
	bestConf := make(map[string]float64)

	for _, a := range h.Attempts {
		for col, conf := range a.Confidences {
			if conf > bestConf[col] {
				bestConf[col] = conf
				if val, ok := a.Results[col]; ok {
					best[col] = val
				}
			}
		}
	}
	return best
}

type AttemptStore interface {
	GetOrCreate(rowKey string) *AttemptHistory
	Save(history *AttemptHistory) error
	Delete(rowKey string)
}

type InMemoryAttemptStore struct {
	mu       sync.RWMutex
	histories map[string]*AttemptHistory
}

func NewInMemoryAttemptStore() *InMemoryAttemptStore {
	return &InMemoryAttemptStore{
		histories: make(map[string]*AttemptHistory),
	}
}

func (s *InMemoryAttemptStore) GetOrCreate(rowKey string) *AttemptHistory {
	s.mu.Lock()
	defer s.mu.Unlock()

	if h, ok := s.histories[rowKey]; ok {
		return h
	}

	h := NewAttemptHistory(rowKey)
	s.histories[rowKey] = h
	return h
}

func (s *InMemoryAttemptStore) Save(history *AttemptHistory) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.histories[history.RowKey] = history
	return nil
}

func (s *InMemoryAttemptStore) Delete(rowKey string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.histories, rowKey)
}

type PartialResult struct {
	ExtractedData map[string]any
	Confidences   map[string]*models.FieldConfidenceInfo
	Sources       []string
}

func MergePartialResults(existing, new *PartialResult) *PartialResult {
	if existing == nil {
		return new
	}
	if new == nil {
		return existing
	}

	merged := &PartialResult{
		ExtractedData: make(map[string]any),
		Confidences:   make(map[string]*models.FieldConfidenceInfo),
		Sources:       make([]string, 0),
	}

	maps.Copy(merged.ExtractedData, existing.ExtractedData)
	maps.Copy(merged.Confidences, existing.Confidences)

	for k, v := range new.ExtractedData {
		existingConf := float64(0)
		if ec, ok := existing.Confidences[k]; ok && ec != nil {
			existingConf = ec.Score
		}

		newConf := float64(0)
		if nc, ok := new.Confidences[k]; ok && nc != nil {
			newConf = nc.Score
		}

		if newConf > existingConf {
			merged.ExtractedData[k] = v
			if nc, ok := new.Confidences[k]; ok {
				merged.Confidences[k] = nc
			}
		}
	}

	seen := make(map[string]struct{})
	for _, s := range existing.Sources {
		if _, ok := seen[s]; !ok {
			seen[s] = struct{}{}
			merged.Sources = append(merged.Sources, s)
		}
	}
	for _, s := range new.Sources {
		if _, ok := seen[s]; !ok {
			seen[s] = struct{}{}
			merged.Sources = append(merged.Sources, s)
		}
	}

	return merged
}
