package services

import (
	"context"
	"os"
	"strings"
	"testing"
)

func TestWithCaveman_SetsSystemPrompt(t *testing.T) {
	c := &GeminiAIClient{}
	if err := WithCaveman()(c); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(c.systemPrompt, "caveman") {
		t.Errorf("systemPrompt missing caveman instruction, got: %q", c.systemPrompt)
	}
}

func TestWithModel_SetsModel(t *testing.T) {
	c := &GeminiAIClient{}
	if err := WithModel("gemini-2.0-flash")(c); err != nil {
		t.Fatal(err)
	}
	if c.model != "gemini-2.0-flash" {
		t.Errorf("expected model %q, got %q", "gemini-2.0-flash", c.model)
	}
}

func skipIfNotLive(t *testing.T) {
	t.Helper()
	if os.Getenv("AI_LIVE") == "" {
		t.Skip("set AI_LIVE=1 to run live AI tests")
	}
}

func TestLive_GenerateContent(t *testing.T) {
	skipIfNotLive(t)
	client, err := NewGeminiAIClient()
	if err != nil {
		t.Fatal(err)
	}
	out, err := client.GenerateContent(context.Background(), "Explain what is python in one sentance")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("response:", out)
}

func TestLive_WithCaveman_GenerateContent(t *testing.T) {
	skipIfNotLive(t)
	client, err := NewGeminiAIClient(WithCaveman())
	if err != nil {
		t.Fatal(err)
	}
	out, err := client.GenerateContent(context.Background(), "Explain what is python in one sentance")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("caveman response:", out)
}
