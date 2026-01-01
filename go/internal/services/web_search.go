package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/blagoySimandov/ampledata/go/internal/models"
)

type WebSearcher interface {
	Search(ctx context.Context, query string) (*models.GoogleSearchResults, error)
}

type SerperClient struct {
	apiKey     string
	httpClient *http.Client
	baseURL    string
}

func NewSerperClient(apiKey string) *SerperClient {
	return &SerperClient{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://google.serper.dev/search",
	}
}

func (c *SerperClient) Search(ctx context.Context, query string) (*models.GoogleSearchResults, error) {
	payload := map[string]string{"q": query}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL, bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-API-KEY", c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("serper API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var result models.GoogleSearchResults
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}
