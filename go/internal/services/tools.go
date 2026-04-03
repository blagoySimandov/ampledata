package services

import (
	"context"
	"fmt"
)

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

// Tool repository — each entry pairs a schema definition with its handler.

func NewFetchPageTool(crawler WebCrawler) Tool {
	return Tool{
		Definition: ToolDefinition{
			Name:        "fetch_page",
			Description: "Fetch the content of a web page by URL. Use this when you see a link in the content that likely contains additional information about the target entity (e.g. an about page, a detailed profile, or a data source referenced in the text).",
			Parameters: []ToolParameter{
				{Name: "url", Type: ToolParamString, Description: "The URL of the page to fetch"},
			},
			Required: []string{"url"},
		},
		Handler: func(ctx context.Context, args map[string]any) (map[string]any, error) {
			url, ok := args["url"].(string)
			if !ok {
				return map[string]any{"error": "url parameter must be a string"}, nil
			}
			content, err := crawler.Crawl(ctx, []string{url}, "")
			if err != nil {
				return map[string]any{"error": fmt.Sprintf("failed to fetch page: %v", err)}, nil
			}
			return map[string]any{"content": content}, nil
		},
	}
}
