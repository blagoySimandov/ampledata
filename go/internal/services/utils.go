package services

import (
	"strings"
)

func cleanJSONMarkdown(content string) string {
	content = strings.TrimSpace(content)

	if !strings.Contains(content, "```") {
		return content
	}

	parts := strings.Split(content, "```")
	if len(parts) < 2 {
		return content
	}

	inner := parts[1]

	if strings.HasPrefix(inner, "json") {
		inner = inner[4:]
	} else if strings.HasPrefix(inner, "JSON") {
		inner = inner[4:]
	}

	return strings.TrimSpace(inner)
}

func applyFuncOptions[T any](entity T, opts ...func(entity T) error) error {
	for _, opt := range opts {
		err := opt(entity)
		if err != nil {
			return err
		}
	}
	return nil
}
