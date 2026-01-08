package services

import (
	"context"

	"github.com/blagoySimandov/ampledata/go/internal/cache"
	"github.com/blagoySimandov/ampledata/go/internal/feedback"
	"github.com/blagoySimandov/ampledata/go/internal/models"
	"github.com/rs/zerolog/log"
)

type CachedPatternGenerator struct {
	underlying QueryPatternGenerator
	cache      cache.PatternCache
}

func NewCachedPatternGenerator(underlying QueryPatternGenerator, patternCache cache.PatternCache) *CachedPatternGenerator {
	return &CachedPatternGenerator{
		underlying: underlying,
		cache:      patternCache,
	}
}

func (c *CachedPatternGenerator) GeneratePatterns(ctx context.Context, columnsMetadata []*models.ColumnMetadata) ([]string, error) {
	return c.GeneratePatternsWithFeedback(ctx, columnsMetadata, nil)
}

func (c *CachedPatternGenerator) GeneratePatternsWithFeedback(ctx context.Context, columnsMetadata []*models.ColumnMetadata, fb *feedback.EnrichmentFeedback) ([]string, error) {
	if fb != nil && fb.IsRetry() {
		log.Debug().
			Int("attemptNumber", fb.AttemptNumber).
			Msg("Skipping cache for retry attempt")
		return c.underlying.GeneratePatternsWithFeedback(ctx, columnsMetadata, fb)
	}

	cacheKey := cache.GenerateCacheKey(columnsMetadata)

	if patterns, ok := c.cache.Get(ctx, cacheKey); ok {
		log.Debug().
			Str("cacheKey", cacheKey[:16]).
			Int("numPatterns", len(patterns)).
			Msg("Pattern cache hit")
		return patterns, nil
	}

	log.Debug().
		Str("cacheKey", cacheKey[:16]).
		Msg("Pattern cache miss, generating patterns")

	patterns, err := c.underlying.GeneratePatternsWithFeedback(ctx, columnsMetadata, fb)
	if err != nil {
		return patterns, err
	}

	if err := c.cache.Set(ctx, cacheKey, patterns); err != nil {
		log.Warn().
			Err(err).
			Str("cacheKey", cacheKey[:16]).
			Msg("Failed to cache patterns")
	}

	return patterns, nil
}
