package services

import (
	"context"
	"fmt"

	"google.golang.org/genai"
)

type IAIClient interface {
	GenerateContent(ctx context.Context, prompt string) (string, error)
}

type GeminiAIClient struct {
	client  *genai.Client
	tracker ICostTracker
	model   string
}
type GeminiAIClientFuncOptions = func(client *GeminiAIClient) error

func NewGeminiAIClient(opts ...GeminiAIClientFuncOptions) (*GeminiAIClient, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create AI client: %w", err)
	}
	geminiai := GeminiAIClient{
		client: client,
		model:  "gemini-3-flash-preview",
	}
	err = applyFuncOptions(&geminiai, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to apply options: %w", err)
	}
	return &geminiai, nil
}

func WithModel(model string) GeminiAIClientFuncOptions {
	return func(client *GeminiAIClient) error {
		client.model = model
		return nil
	}
}

func (g *GeminiAIClient) GenerateContent(ctx context.Context, prompt string) (string, error) {
	result, err := g.client.Models.GenerateContent(ctx, g.model, genai.Text(prompt), nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}
	um := Deref(result.UsageMetadata)
	g.TrackCost(ctx, um)

	return result.Text(), nil
}

func (g *GeminiAIClient) TrackCost(ctx context.Context, um genai.GenerateContentResponseUsageMetadata) {
	tknIn := um.PromptTokenCount
	totalTknCount := um.TotalTokenCount
	tknOut := totalTknCount - tknIn
	g.tracker.AddTokenCost(ctx, int(tknIn), int(tknOut))
}

func WithCostTracker(tracker ICostTracker) GeminiAIClientFuncOptions {
	return func(client *GeminiAIClient) error {
		client.tracker = tracker
		return nil
	}
}
