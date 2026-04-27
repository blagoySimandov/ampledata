package services

import (
	"context"
	"fmt"

	"google.golang.org/genai"
)

type IAIClient interface {
	GenerateContent(ctx context.Context, prompt string, opts ...GenerateOption) (string, error)
}

type generateConfig struct {
	tools    []Tool
	maxSteps int
}

type GenerateOption func(*generateConfig)

func WithTools(tools []Tool, maxSteps int) GenerateOption {
	return func(c *generateConfig) {
		c.tools = tools
		c.maxSteps = maxSteps
	}
}

type GeminiAIClient struct {
	client       *genai.Client
	tracker      ICostTracker
	systemPrompt string
	model        string
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
	if err = applyFuncOptions(&geminiai, opts...); err != nil {
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

func WithCaveman() GeminiAIClientFuncOptions {
	const cavemanInstruction = `Respond terse like smart caveman. All technical substance stay. Only fluff die.

Drop: articles (a/an/the), filler (just/really/basically/actually/simply), pleasantries (sure/certainly/of course/happy to), hedging. Fragments OK. Short synonyms (big not extensive, fix not "implement a solution for"). Technical terms exact. Code blocks unchanged. Errors quoted exact.

Pattern: [thing] [action] [reason]. [next step].
`
	return func(client *GeminiAIClient) error {
		client.systemPrompt += cavemanInstruction
		return nil
	}
}

func WithCostTracker(tracker ICostTracker) GeminiAIClientFuncOptions {
	return func(client *GeminiAIClient) error {
		client.tracker = tracker
		return nil
	}
}

func (g *GeminiAIClient) GenerateContent(ctx context.Context, prompt string, opts ...GenerateOption) (string, error) {
	cfg := &generateConfig{}
	for _, o := range opts {
		o(cfg)
	}

	config := g.baseConfig()
	contents := genai.Text(prompt)

	if len(cfg.tools) == 0 {
		result, err := g.generateOnce(ctx, contents, config)
		if err != nil {
			return "", fmt.Errorf("failed to generate content: %w", err)
		}
		return result.Text(), nil
	}

	config.Tools = []*genai.Tool{{FunctionDeclarations: toFuncDecls(cfg.tools)}}
	for step := range cfg.maxSteps + 1 {
		result, err := g.generateOnce(ctx, contents, config)
		if err != nil {
			return "", fmt.Errorf("step %d: %w", step, err)
		}
		if len(result.Candidates) == 0 || result.Candidates[0].Content == nil {
			return result.Text(), nil
		}
		modelContent := result.Candidates[0].Content
		fcs := collectFunctionCalls(modelContent)
		if len(fcs) == 0 {
			return result.Text(), nil
		}
		contents = append(contents, modelContent)
		contents = append(contents, buildFunctionResponses(ctx, fcs, cfg.tools))
	}

	return "", fmt.Errorf("max tool call steps (%d) exceeded", cfg.maxSteps)
}

func (g *GeminiAIClient) TrackCost(ctx context.Context, um genai.GenerateContentResponseUsageMetadata) {
	if g.tracker == nil {
		return
	}
	tknIn := um.PromptTokenCount
	tknOut := um.TotalTokenCount - tknIn
	g.tracker.AddTokenCost(ctx, int(tknIn), int(tknOut))
}

func (g *GeminiAIClient) baseConfig() *genai.GenerateContentConfig {
	cfg := &genai.GenerateContentConfig{}
	if g.systemPrompt != "" {
		cfg.SystemInstruction = &genai.Content{
			Role:  "system",
			Parts: []*genai.Part{{Text: g.systemPrompt}},
		}
	}
	return cfg
}

func (g *GeminiAIClient) generateOnce(ctx context.Context, contents []*genai.Content, config *genai.GenerateContentConfig) (*genai.GenerateContentResponse, error) {
	result, err := g.client.Models.GenerateContent(ctx, g.model, contents, config)
	if err != nil {
		return nil, err
	}
	g.TrackCost(ctx, Deref(result.UsageMetadata))
	return result, nil
}

func toFuncDecls(tools []Tool) []*genai.FunctionDeclaration {
	decls := make([]*genai.FunctionDeclaration, len(tools))
	for i, t := range tools {
		decls[i] = toFuncDecl(t.Definition)
	}
	return decls
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
