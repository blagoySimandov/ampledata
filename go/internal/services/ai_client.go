package services

import (
	"context"
	"fmt"

	"google.golang.org/genai"
)

type IAIClient interface {
	GenerateContent(ctx context.Context, prompt string) (string, error)
}

// ToolCallHandler executes a tool call and returns the result.
// The framework dispatches to the correct handler by name — the handler
// only needs to act on the args it was registered with.
type ToolCallHandler func(ctx context.Context, args map[string]any) (map[string]any, error)

// ToolParamType represents the type of a tool parameter.
type ToolParamType string

const (
	ToolParamString  ToolParamType = "STRING"
	ToolParamNumber  ToolParamType = "NUMBER"
	ToolParamInteger ToolParamType = "INTEGER"
	ToolParamBoolean ToolParamType = "BOOLEAN"
)

// ToolDefinition describes the schema of a tool the model can invoke.
type ToolDefinition struct {
	Name        string
	Description string
	Parameters  []ToolParameter
	Required    []string
}

// ToolParameter describes a single parameter of a tool.
type ToolParameter struct {
	Name        string
	Type        ToolParamType
	Description string
}

// Tool pairs a definition (what the model sees) with its handler (what runs when called).
type Tool struct {
	Definition ToolDefinition
	Handler    ToolCallHandler
}

// IToolAIClient extends IAIClient with support for function-calling tools.
type IToolAIClient interface {
	IAIClient
	GenerateContentWithTools(ctx context.Context, prompt string, tools []Tool, maxSteps int) (string, error)
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

func (g *GeminiAIClient) GenerateContentWithTools(ctx context.Context, prompt string, tools []Tool, maxSteps int) (string, error) {
	funcDecls := make([]*genai.FunctionDeclaration, len(tools))
	for i, t := range tools {
		funcDecls[i] = toFuncDecl(t.Definition)
	}

	config := &genai.GenerateContentConfig{
		Tools: []*genai.Tool{{FunctionDeclarations: funcDecls}},
	}

	contents := genai.Text(prompt)

	for step := range maxSteps + 1 {
		result, err := g.client.Models.GenerateContent(ctx, g.model, contents, config)
		if err != nil {
			return "", fmt.Errorf("step %d: %w", step, err)
		}
		g.TrackCost(ctx, Deref(result.UsageMetadata))

		if len(result.Candidates) == 0 || result.Candidates[0].Content == nil {
			return result.Text(), nil
		}

		modelContent := result.Candidates[0].Content
		fcs := collectFunctionCalls(modelContent)
		if len(fcs) == 0 {
			return result.Text(), nil
		}

		contents = append(contents, modelContent)
		contents = append(contents, buildFunctionResponses(ctx, fcs, tools))
	}

	return "", fmt.Errorf("max tool call steps (%d) exceeded", maxSteps)
}

func toFuncDecl(td ToolDefinition) *genai.FunctionDeclaration {
	props := make(map[string]*genai.Schema, len(td.Parameters))
	for _, p := range td.Parameters {
		props[p.Name] = &genai.Schema{Type: genai.Type(p.Type), Description: p.Description}
	}
	return &genai.FunctionDeclaration{
		Name:        td.Name,
		Description: td.Description,
		Parameters:  &genai.Schema{Type: genai.TypeObject, Properties: props, Required: td.Required},
	}
}

func collectFunctionCalls(c *genai.Content) []*genai.FunctionCall {
	var out []*genai.FunctionCall
	for _, p := range c.Parts {
		if p.FunctionCall != nil {
			out = append(out, p.FunctionCall)
		}
	}
	return out
}

func buildFunctionResponses(ctx context.Context, fcs []*genai.FunctionCall, tools []Tool) *genai.Content {
	handlers := make(map[string]ToolCallHandler, len(tools))
	for _, t := range tools {
		handlers[t.Definition.Name] = t.Handler
	}

	parts := make([]*genai.Part, len(fcs))
	for i, fc := range fcs {
		var resp map[string]any
		if h, ok := handlers[fc.Name]; ok {
			var err error
			resp, err = h(ctx, fc.Args)
			if err != nil {
				resp = map[string]any{"error": err.Error()}
			}
		} else {
			resp = map[string]any{"error": fmt.Sprintf("unknown tool: %s", fc.Name)}
		}
		parts[i] = &genai.Part{FunctionResponse: &genai.FunctionResponse{ID: fc.ID, Name: fc.Name, Response: resp}}
	}
	return &genai.Content{Role: "user", Parts: parts}
}
