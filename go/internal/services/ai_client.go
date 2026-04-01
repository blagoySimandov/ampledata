package services

import (
	"context"
	"fmt"

	"google.golang.org/genai"
)

type IAIClient interface {
	GenerateContent(ctx context.Context, prompt string) (string, error)
	GenerateStructuredContent(ctx context.Context, prompt string, schema *genai.Schema) (string, error)
	GenerateWithTools(ctx context.Context, prompt string, tools []*genai.Tool, handler func(string, map[string]any) (any, error)) (string, error)
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
		model:  "gemini-2.5-flash",
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

// GenerateStructuredContent calls Gemini with a response schema, enforcing the
// exact JSON shape at the model level. This eliminates silent parse failures
// that occur when the model produces free-form text instead of valid JSON.
func (g *GeminiAIClient) GenerateStructuredContent(ctx context.Context, prompt string, schema *genai.Schema) (string, error) {
	config := &genai.GenerateContentConfig{
		ResponseMIMEType: "application/json",
		ResponseSchema:   schema,
	}
	result, err := g.client.Models.GenerateContent(ctx, g.model, genai.Text(prompt), config)
	if err != nil {
		return "", fmt.Errorf("failed to generate structured content: %w", err)
	}
	um := Deref(result.UsageMetadata)
	g.TrackCost(ctx, um)

	return result.Text(), nil
}

// GenerateWithTools runs a multi-turn agentic loop. The model may call tools
// (declared in tools) zero or more times; each call is dispatched to handler,
// and the result is fed back before the next turn. The loop ends when the
// model produces a plain text response with no further function calls.
func (g *GeminiAIClient) GenerateWithTools(ctx context.Context, prompt string, tools []*genai.Tool, handler func(string, map[string]any) (any, error)) (string, error) {
	config := &genai.GenerateContentConfig{
		Tools: tools,
	}

	contents := genai.Text(prompt)

	const maxIterations = 10
	for i := 0; i < maxIterations; i++ {
		result, err := g.client.Models.GenerateContent(ctx, g.model, contents, config)
		if err != nil {
			return "", fmt.Errorf("failed to generate content (iteration %d): %w", i+1, err)
		}
		um := Deref(result.UsageMetadata)
		g.TrackCost(ctx, um)

		if len(result.Candidates) == 0 {
			return "", fmt.Errorf("no candidates returned (iteration %d)", i+1)
		}
		candidate := result.Candidates[0]

		// Collect any function calls from this turn.
		var functionResponses []*genai.Part
		for _, part := range candidate.Content.Parts {
			if part.FunctionCall == nil {
				continue
			}
			fc := part.FunctionCall
			toolResult, toolErr := handler(fc.Name, fc.Args)
			var response map[string]any
			if toolErr != nil {
				response = map[string]any{"error": toolErr.Error()}
			} else {
				response = map[string]any{"result": toolResult}
			}
			functionResponses = append(functionResponses, &genai.Part{
				FunctionResponse: &genai.FunctionResponse{
					ID:       fc.ID,
					Name:     fc.Name,
					Response: response,
				},
			})
		}

		if len(functionResponses) == 0 {
			// No tool calls → model has given its final answer.
			return result.Text(), nil
		}

		// Extend the conversation: append model turn then tool-result turn.
		contents = append(contents, candidate.Content)
		contents = append(contents, &genai.Content{
			Role:  "user",
			Parts: functionResponses,
		})
	}

	return "", fmt.Errorf("agentic loop exceeded maximum iterations (%d)", maxIterations)
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
