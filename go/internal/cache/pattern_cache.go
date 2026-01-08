package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
	"sync"

	"github.com/blagoySimandov/ampledata/go/internal/models"
)

type PatternCache interface {
	Get(ctx context.Context, key string) ([]string, bool)
	Set(ctx context.Context, key string, patterns []string) error
}

type InMemoryPatternCache struct {
	mu    sync.RWMutex
	cache map[string][]string
}

func NewInMemoryPatternCache() *InMemoryPatternCache {
	return &InMemoryPatternCache{
		cache: make(map[string][]string),
	}
}

func (c *InMemoryPatternCache) Get(ctx context.Context, key string) ([]string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	patterns, ok := c.cache[key]
	if !ok {
		return nil, false
	}

	result := make([]string, len(patterns))
	copy(result, patterns)
	return result, true
}

func (c *InMemoryPatternCache) Set(ctx context.Context, key string, patterns []string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	stored := make([]string, len(patterns))
	copy(stored, patterns)
	c.cache[key] = stored
	return nil
}

func GenerateCacheKey(columns []*models.ColumnMetadata) string {
	type columnData struct {
		Name        string  `json:"name"`
		Type        string  `json:"type"`
		Description *string `json:"description,omitempty"`
	}

	data := make([]columnData, len(columns))
	for i, col := range columns {
		data[i] = columnData{
			Name:        col.Name,
			Type:        string(col.Type),
			Description: col.Description,
		}
	}

	sort.Slice(data, func(i, j int) bool {
		return data[i].Name < data[j].Name
	})

	jsonBytes, _ := json.Marshal(data)
	hash := sha256.Sum256(jsonBytes)
	return hex.EncodeToString(hash[:])
}
