package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLLMEndpointURL_BaseURL_AppendsPath(t *testing.T) {
	tests := []struct {
		name        string
		endpoint    string
		defaultPath string
		want        string
	}{
		{
			name:        "openai base URL gets default path appended",
			endpoint:    "https://api.openai.com",
			defaultPath: "/v1/chat/completions",
			want:        "https://api.openai.com/v1/chat/completions",
		},
		{
			name:        "anthropic base URL gets default path appended",
			endpoint:    "https://api.anthropic.com",
			defaultPath: "/v1/messages",
			want:        "https://api.anthropic.com/v1/messages",
		},
		{
			name:        "trailing slash on base URL gets default path appended",
			endpoint:    "https://api.openai.com/",
			defaultPath: "/v1/chat/completions",
			want:        "https://api.openai.com//v1/chat/completions",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := llmEndpointURL(tc.endpoint, tc.defaultPath)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestLLMEndpointURL_FullPathURL_UsedAsIs(t *testing.T) {
	tests := []struct {
		name        string
		endpoint    string
		defaultPath string
	}{
		{
			name:        "Azure AI Foundry full URL with query string",
			endpoint:    "https://cog-easi-prd01.cognitiveservices.azure.com/openai/responses?api-version=2025-04-01-preview",
			defaultPath: "/v1/chat/completions",
		},
		{
			name:        "custom host with explicit path",
			endpoint:    "https://my-llm.example.com/llm/v2/chat",
			defaultPath: "/v1/chat/completions",
		},
		{
			name:        "localhost with non-root path",
			endpoint:    "http://localhost:11434/api/chat",
			defaultPath: "/v1/chat/completions",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := llmEndpointURL(tc.endpoint, tc.defaultPath)
			assert.Equal(t, tc.endpoint, got, "full-path endpoint should be returned unchanged")
		})
	}
}
