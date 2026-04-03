package services

import (
	"context"
	"fmt"

	"google.golang.org/genai"
)

type IAIClient interface {
	GenerateContent(ctx context.Context, prompt string) (string, error)
}

// ToolCallHandler processes a function call from the model and returns the result.
type ToolCallHandler func(ctx context.Context, name string, args map[string]any) (map[string]any, error)

// ToolDefinition describes a tool that the model can invoke.
type ToolDefinition struct {
	Name        string
	Description string
	Parameters  []ToolParameter
	Required    []string
}

// ToolParameter describes a single parameter of a tool.
type ToolParameter struct {
	Name        string
	Type        string // genai type string: "STRING", "NUMBER", "BOOLEAN", etc.
	Description string
}

// IToolAIClient extends IAIClient with support for function-calling tools.
type IToolAIClient interface {
	IAIClient
	GenerateContentWithTools(ctx context.Context, prompt string, toolDefs []ToolDefinition, handler ToolCallHandler, maxSteps int) (string, error)
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

func (g *GeminiAIClient) GenerateContentWithTools(ctx context.Context, prompt string, toolDefs []ToolDefinition, handler ToolCallHandler, maxSteps int) (string, error) {
	// Convert ToolDefinitions to genai tool declarations.
	var funcDecls []*genai.FunctionDeclaration
	for _, td := range toolDefs {
		props := make(map[string]*genai.Schema)
		for _, p := range td.Parameters {
			props[p.Name] = &genai.Schema{
				Type:        genai.Type(p.Type),
				Description: p.Description,
			}
		}
		funcDecls = append(funcDecls, &genai.FunctionDeclaration{
			Name:        td.Name,
			Description: td.Description,
			Parameters: &genai.Schema{
				Type:       genai.TypeObject,
				Properties: props,
				Required:   td.Required,
			},
		})
	}

	config := &genai.GenerateContentConfig{
		Tools: []*genai.Tool{
			{FunctionDeclarations: funcDecls},
		},
	}

	contents := genai.Text(prompt)

	for step := 0; step <= maxSteps; step++ {
		result, err := g.client.Models.GenerateContent(ctx, g.model, contents, config)
		if err != nil {
			return "", fmt.Errorf("failed to generate content (step %d): %w", step, err)
		}

		um := Deref(result.UsageMetadata)
		g.TrackCost(ctx, um)

		if len(result.Candidates) == 0 || result.Candidates[0].Content == nil {
			return result.Text(), nil
		}

		modelContent := result.Candidates[0].Content

		// Collect any function calls from the model's response.
		var functionCalls []*genai.FunctionCall
		for _, part := range modelContent.Parts {
			if part.FunctionCall != nil {
				functionCalls = append(functionCalls, part.FunctionCall)
			}
		}

		if len(functionCalls) == 0 {
			return result.Text(), nil
		}

		// Append the model's response to the conversation.
		contents = append(contents, modelContent)

		// Execute each function call and build response parts.
		var responseParts []*genai.Part
		for _, fc := range functionCalls {
			resp, err := handler(ctx, fc.Name, fc.Args)
			if err != nil {
				resp = map[string]any{"error": err.Error()}
			}
			responseParts = append(responseParts, &genai.Part{
				FunctionResponse: &genai.FunctionResponse{
					ID:       fc.ID,
					Name:     fc.Name,
					Response: resp,
				},
			})
		}

		contents = append(contents, &genai.Content{
			Role:  "user",
			Parts: responseParts,
		})
	}

	return "", fmt.Errorf("max tool call steps (%d) exceeded", maxSteps)
}
