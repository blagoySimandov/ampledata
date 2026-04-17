package services

import (
	"context"
	"errors"
)

// ErrToolsNotSupported is returned when GenerateContentWithTools is called on a wrapper
// whose inner client does not implement IToolAIClient.
var ErrToolsNotSupported = errors.New("inner AI client does not support tool calls")

// CavemanAIClient wraps any IAIClient and prepends a terse-response instruction to every
// prompt, cutting output token usage ~75%. Drop-in replacement: same interfaces, lower cost.
// Useful when response verbosity doesn't matter (structured JSON extractions, short decisions).
const cavemanInstruction = `Respond terse like smart caveman. All technical substance stay. Only fluff die.

Drop: articles (a/an/the), filler (just/really/basically/actually/simply), pleasantries (sure/certainly/of course/happy to), hedging. Fragments OK. Short synonyms (big not extensive, fix not "implement a solution for"). Technical terms exact. Code blocks unchanged. Errors quoted exact.

Pattern: [thing] [action] [reason]. [next step].

`

// CavemanAIClient wraps an IAIClient to reduce response verbosity and token cost.
type CavemanAIClient struct {
	inner IAIClient
}

// NewCavemanAIClient wraps client with caveman terse-response mode.
func NewCavemanAIClient(client IAIClient) *CavemanAIClient {
	return &CavemanAIClient{inner: client}
}

func (c *CavemanAIClient) GenerateContent(ctx context.Context, prompt string) (string, error) {
	return c.inner.GenerateContent(ctx, cavemanInstruction+prompt)
}

// GenerateContentWithTools delegates to the inner IToolAIClient if supported, otherwise errors.
func (c *CavemanAIClient) GenerateContentWithTools(ctx context.Context, prompt string, tools []Tool, maxSteps int) (string, error) {
	tc, ok := c.inner.(IToolAIClient)
	if !ok {
		return "", ErrToolsNotSupported
	}
	return tc.GenerateContentWithTools(ctx, cavemanInstruction+prompt, tools, maxSteps)
}
