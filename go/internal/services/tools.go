package services

import (
	"context"
	"fmt"
)

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
