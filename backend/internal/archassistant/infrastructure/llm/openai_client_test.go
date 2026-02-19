package llm_test

import (
	"context"
	"encoding/json"
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

func TestOpenAIClient_StreamChat_WithToolCalls(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		fmt.Fprint(w, `data: {"choices":[{"delta":{"tool_calls":[{"index":0,"id":"call_abc","type":"function","function":{"name":"list_apps","arguments":""}}]}}]}`+"\n\n")
		fmt.Fprint(w, `data: {"choices":[{"delta":{"tool_calls":[{"index":0,"function":{"arguments":"{\"na"}}]}}]}`+"\n\n")
		fmt.Fprint(w, `data: {"choices":[{"delta":{"tool_calls":[{"index":0,"function":{"arguments":"me\":\""}}]}}]}`+"\n\n")
		fmt.Fprint(w, `data: {"choices":[{"delta":{"tool_calls":[{"index":0,"function":{"arguments":"test\"}"}}]}}]}`+"\n\n")
		fmt.Fprint(w, `data: {"choices":[{"finish_reason":"tool_calls"}],"usage":{"total_tokens":42}}`+"\n\n")
		fmt.Fprint(w, "data: [DONE]\n\n")
	}))
	defer server.Close()

	client := llm.NewOpenAIClient(server.URL, "test-key")
	ch, err := client.StreamChat(context.Background(), []llm.Message{
		{Role: llm.RoleUser, Content: "List apps"},
	}, llm.Options{
		Model:     "gpt-4",
		MaxTokens: 100,
		Tools: []llm.ToolDef{{
			Type: "function",
			Function: llm.ToolFunctionDef{
				Name:        "list_apps",
				Description: "Lists applications",
				Parameters:  map[string]interface{}{"type": "object"},
			},
		}},
	})
	require.NoError(t, err)

	var toolCallEvents []llm.StreamEvent
	var doneEvent *llm.StreamEvent
	for event := range ch {
		switch event.Type {
		case llm.EventToolCall:
			toolCallEvents = append(toolCallEvents, event)
		case llm.EventDone:
			e := event
			doneEvent = &e
		}
	}

	require.Len(t, toolCallEvents, 1)
	require.Len(t, toolCallEvents[0].ToolCalls, 1)
	assert.Equal(t, "call_abc", toolCallEvents[0].ToolCalls[0].ID)
	assert.Equal(t, "list_apps", toolCallEvents[0].ToolCalls[0].Name)
	assert.Equal(t, `{"name":"test"}`, toolCallEvents[0].ToolCalls[0].Arguments)
	require.NotNil(t, doneEvent)
	assert.Equal(t, 42, doneEvent.TokensUsed)
}

func TestOpenAIClient_StreamChat_MultipleToolCalls(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		fmt.Fprint(w, `data: {"choices":[{"delta":{"tool_calls":[{"index":0,"id":"call_1","type":"function","function":{"name":"get_app","arguments":""}}]}}]}`+"\n\n")
		fmt.Fprint(w, `data: {"choices":[{"delta":{"tool_calls":[{"index":0,"function":{"arguments":"{\"id\":1}"}}]}}]}`+"\n\n")
		fmt.Fprint(w, `data: {"choices":[{"delta":{"tool_calls":[{"index":1,"id":"call_2","type":"function","function":{"name":"get_vendor","arguments":""}}]}}]}`+"\n\n")
		fmt.Fprint(w, `data: {"choices":[{"delta":{"tool_calls":[{"index":1,"function":{"arguments":"{\"id\":2}"}}]}}]}`+"\n\n")
		fmt.Fprint(w, `data: {"choices":[{"finish_reason":"tool_calls"}]}`+"\n\n")
		fmt.Fprint(w, "data: [DONE]\n\n")
	}))
	defer server.Close()

	client := llm.NewOpenAIClient(server.URL, "test-key")
	ch, err := client.StreamChat(context.Background(), []llm.Message{
		{Role: llm.RoleUser, Content: "Get info"},
	}, llm.Options{Model: "gpt-4", MaxTokens: 100})
	require.NoError(t, err)

	var toolCallEvents []llm.StreamEvent
	for event := range ch {
		if event.Type == llm.EventToolCall {
			toolCallEvents = append(toolCallEvents, event)
		}
	}

	require.Len(t, toolCallEvents, 1)
	require.Len(t, toolCallEvents[0].ToolCalls, 2)
	assert.Equal(t, "call_1", toolCallEvents[0].ToolCalls[0].ID)
	assert.Equal(t, "get_app", toolCallEvents[0].ToolCalls[0].Name)
	assert.Equal(t, `{"id":1}`, toolCallEvents[0].ToolCalls[0].Arguments)
	assert.Equal(t, "call_2", toolCallEvents[0].ToolCalls[1].ID)
	assert.Equal(t, "get_vendor", toolCallEvents[0].ToolCalls[1].Name)
	assert.Equal(t, `{"id":2}`, toolCallEvents[0].ToolCalls[1].Arguments)
}

func TestOpenAIClient_StreamChat_ToolsInRequest(t *testing.T) {
	var receivedBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedBody, _ = readAll(r.Body)
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "data: [DONE]\n\n")
	}))
	defer server.Close()

	client := llm.NewOpenAIClient(server.URL, "test-key")
	ch, err := client.StreamChat(context.Background(), []llm.Message{
		{Role: llm.RoleUser, Content: "Hi"},
	}, llm.Options{
		Model:     "gpt-4",
		MaxTokens: 100,
		Tools: []llm.ToolDef{{
			Type: "function",
			Function: llm.ToolFunctionDef{
				Name:        "list_apps",
				Description: "Lists applications",
				Parameters:  map[string]interface{}{"type": "object"},
			},
		}},
	})
	require.NoError(t, err)
	for range ch {
	}

	var body map[string]interface{}
	require.NoError(t, json.Unmarshal(receivedBody, &body))
	assert.Contains(t, body, "tools")
	assert.Equal(t, "auto", body["tool_choice"])

	tools := body["tools"].([]interface{})
	require.Len(t, tools, 1)
	tool := tools[0].(map[string]interface{})
	assert.Equal(t, "function", tool["type"])
	fn := tool["function"].(map[string]interface{})
	assert.Equal(t, "list_apps", fn["name"])
}

func TestOpenAIClient_StreamChat_ToolCallMessageInRequest(t *testing.T) {
	var receivedBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedBody, _ = readAll(r.Body)
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "data: [DONE]\n\n")
	}))
	defer server.Close()

	client := llm.NewOpenAIClient(server.URL, "test-key")
	ch, err := client.StreamChat(context.Background(), []llm.Message{
		{Role: llm.RoleUser, Content: "List apps"},
		{Role: llm.RoleAssistant, ToolCalls: []llm.ToolCall{{ID: "call_1", Name: "list_apps", Arguments: `{"name":"x"}`}}},
		{Role: llm.RoleTool, Content: `[{"name":"App1"}]`, ToolCallID: "call_1", Name: "list_apps"},
	}, llm.Options{Model: "gpt-4", MaxTokens: 100})
	require.NoError(t, err)
	for range ch {
	}

	var body map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(receivedBody, &body))

	var messages []map[string]interface{}
	require.NoError(t, json.Unmarshal(body["messages"], &messages))
	require.Len(t, messages, 3)

	assistantMsg := messages[1]
	assert.Equal(t, "assistant", assistantMsg["role"])
	toolCalls := assistantMsg["tool_calls"].([]interface{})
	require.Len(t, toolCalls, 1)
	tc := toolCalls[0].(map[string]interface{})
	assert.Equal(t, "call_1", tc["id"])
	fn := tc["function"].(map[string]interface{})
	assert.Equal(t, "list_apps", fn["name"])
	assert.Equal(t, `{"name":"x"}`, fn["arguments"])

	toolMsg := messages[2]
	assert.Equal(t, "tool", toolMsg["role"])
	assert.Equal(t, "call_1", toolMsg["tool_call_id"])
}
