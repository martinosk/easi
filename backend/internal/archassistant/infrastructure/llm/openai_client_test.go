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

func TestOpenAIClient_StreamChat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/chat/completions", r.URL.Path)
		assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"Hello\"}}]}\n\n")
		fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\" world\"}}]}\n\n")
		fmt.Fprint(w, "data: [DONE]\n\n")
	}))
	defer server.Close()

	client := llm.NewOpenAIClient(server.URL, "test-key")
	ch, err := client.StreamChat(context.Background(), []llm.Message{
		{Role: llm.RoleSystem, Content: "You are helpful."},
		{Role: llm.RoleUser, Content: "Hi"},
	}, llm.Options{Model: "gpt-4", MaxTokens: 100, Temperature: 0.3})

	require.NoError(t, err)

	var tokens []string
	var done bool
	for event := range ch {
		switch event.Type {
		case llm.EventToken:
			tokens = append(tokens, event.Content)
		case llm.EventDone:
			done = true
		}
	}

	assert.Equal(t, []string{"Hello", " world"}, tokens)
	assert.True(t, done)
}

func TestOpenAIClient_NonOKStatusReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	client := llm.NewOpenAIClient(server.URL, "bad-key")
	_, err := client.StreamChat(context.Background(), []llm.Message{
		{Role: llm.RoleUser, Content: "Hi"},
	}, llm.Options{Model: "gpt-4", MaxTokens: 100})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "401")
}

func TestOpenAIClient_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		flusher := w.(http.Flusher)

		fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"Hello\"}}]}\n\n")
		flusher.Flush()

		<-r.Context().Done()
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	client := llm.NewOpenAIClient(server.URL, "test-key")
	ch, err := client.StreamChat(ctx, []llm.Message{
		{Role: llm.RoleUser, Content: "Hi"},
	}, llm.Options{Model: "gpt-4", MaxTokens: 100})
	require.NoError(t, err)

	event := <-ch
	assert.Equal(t, llm.EventToken, event.Type)
	assert.Equal(t, "Hello", event.Content)

	cancel()

	for range ch {
	}
}
