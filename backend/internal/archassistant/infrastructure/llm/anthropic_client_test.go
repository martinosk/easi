package llm_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"easi/backend/internal/archassistant/infrastructure/llm"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnthropicClient_StreamChat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/messages", r.URL.Path)
		assert.Equal(t, "test-key", r.Header.Get("x-api-key"))
		assert.Equal(t, "2023-06-01", r.Header.Get("anthropic-version"))

		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		fmt.Fprint(w, "event: content_block_start\ndata: {\"type\":\"content_block_start\"}\n\n")
		fmt.Fprint(w, "event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"delta\":{\"type\":\"text_delta\",\"text\":\"Hello\"}}\n\n")
		fmt.Fprint(w, "event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"delta\":{\"type\":\"text_delta\",\"text\":\" world\"}}\n\n")
		fmt.Fprint(w, "event: message_delta\ndata: {\"type\":\"message_delta\",\"usage\":{\"input_tokens\":10,\"output_tokens\":5}}\n\n")
		fmt.Fprint(w, "event: message_stop\ndata: {\"type\":\"message_stop\"}\n\n")
	}))
	defer server.Close()

	client := llm.NewAnthropicClient(server.URL, "test-key")
	ch, err := client.StreamChat(context.Background(), []llm.Message{
		{Role: llm.RoleSystem, Content: "You are helpful."},
		{Role: llm.RoleUser, Content: "Hi"},
	}, llm.Options{Model: "claude-3-opus", MaxTokens: 100, Temperature: 0.3})

	require.NoError(t, err)

	var tokens []string
	var doneEvent *llm.StreamEvent
	for event := range ch {
		switch event.Type {
		case llm.EventToken:
			tokens = append(tokens, event.Content)
		case llm.EventDone:
			e := event
			doneEvent = &e
		}
	}

	assert.Equal(t, []string{"Hello", " world"}, tokens)
	require.NotNil(t, doneEvent)
	assert.Equal(t, 15, doneEvent.TokensUsed)
}

func TestAnthropicClient_NonOKStatusReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	client := llm.NewAnthropicClient(server.URL, "bad-key")
	_, err := client.StreamChat(context.Background(), []llm.Message{
		{Role: llm.RoleUser, Content: "Hi"},
	}, llm.Options{Model: "claude-3-opus", MaxTokens: 100})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "401")
}

func TestAnthropicClient_SystemMessageExtracted(t *testing.T) {
	var receivedBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedBody, _ = readAll(r.Body)
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "event: message_stop\ndata: {\"type\":\"message_stop\"}\n\n")
	}))
	defer server.Close()

	client := llm.NewAnthropicClient(server.URL, "test-key")
	ch, err := client.StreamChat(context.Background(), []llm.Message{
		{Role: llm.RoleSystem, Content: "System prompt here"},
		{Role: llm.RoleUser, Content: "Hi"},
	}, llm.Options{Model: "claude-3-opus", MaxTokens: 100})
	require.NoError(t, err)
	for range ch {
	}

	assert.Contains(t, string(receivedBody), `"system":"System prompt here"`)
	assert.NotContains(t, string(receivedBody), `"role":"system"`)
}

func TestClientFactory(t *testing.T) {
	t.Run("OpenAI", func(t *testing.T) {
		client, err := llm.NewClient("openai", "http://localhost", "key")
		require.NoError(t, err)
		assert.NotNil(t, client)
	})

	t.Run("Anthropic", func(t *testing.T) {
		client, err := llm.NewClient("anthropic", "http://localhost", "key")
		require.NoError(t, err)
		assert.NotNil(t, client)
	})

	t.Run("Unknown", func(t *testing.T) {
		_, err := llm.NewClient("unknown", "http://localhost", "key")
		assert.Error(t, err)
	})
}
