package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type WebCrawler interface {
	Crawl(ctx context.Context, urls []string, query string) (string, error)
}

type Crawl4aiClient struct {
	httpClient *http.Client
	baseURL    string
}

type CrawlRequest struct {
	URLs  []string `json:"urls"`
	Query string   `json:"query"`
}

type CrawlResponse struct {
	Content string `json:"content"`
	Success bool   `json:"success"`
}

func NewCrawl4aiClient(baseURL string) *Crawl4aiClient {
	return &Crawl4aiClient{
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
		baseURL: baseURL,
	}
}

func (c *Crawl4aiClient) Crawl(ctx context.Context, urls []string, query string) (string, error) {
	reqBody := CrawlRequest{
		URLs:  urls,
		Query: query,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/crawl", bytes.NewReader(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("crawl4ai service error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var result CrawlResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if !result.Success {
		return "", fmt.Errorf("crawl failed")
	}

	return result.Content, nil
}
