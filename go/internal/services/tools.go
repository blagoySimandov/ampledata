package services

import (
	"context"
	"fmt"
)

// Tool repository — canonical tool definitions and handler factories.

var FetchPageTool = ToolDefinition{
	Name:        "fetch_page",
	Description: "Fetch the content of a web page by URL. Use this when you see a link in the content that likely contains additional information about the target entity (e.g. an about page, a detailed profile, or a data source referenced in the text).",
	Parameters: []ToolParameter{
		{Name: "url", Type: ToolParamString, Description: "The URL of the page to fetch"},
	},
	Required: []string{"url"},
}

// NewFetchPageHandler returns a ToolCallHandler that uses the given crawler
// to fetch page content.
func NewFetchPageHandler(crawler WebCrawler) ToolCallHandler {
	return func(ctx context.Context, name string, args map[string]any) (map[string]any, error) {
		if name != FetchPageTool.Name {
			return map[string]any{"error": fmt.Sprintf("unknown tool: %s", name)}, nil
		}
		url, ok := args["url"].(string)
		if !ok {
			return map[string]any{"error": "url parameter must be a string"}, nil
		}
		content, err := crawler.Crawl(ctx, []string{url}, "")
		if err != nil {
			return map[string]any{"error": fmt.Sprintf("failed to fetch page: %v", err)}, nil
		}
		return map[string]any{"content": content}, nil
	}
}
