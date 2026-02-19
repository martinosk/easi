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

func TestAnthropicClient_StreamChat_WithToolCalls(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		fmt.Fprint(w, "event: content_block_start\n")
		fmt.Fprint(w, `data: {"type":"content_block_start","index":0,"content_block":{"type":"tool_use","id":"toolu_abc","name":"list_apps","input":{}}}`+"\n\n")
		fmt.Fprint(w, "event: content_block_delta\n")
		fmt.Fprint(w, `data: {"type":"content_block_delta","index":0,"delta":{"type":"input_json_delta","partial_json":"{\"name\":\""}}`+"\n\n")
		fmt.Fprint(w, "event: content_block_delta\n")
		fmt.Fprint(w, `data: {"type":"content_block_delta","index":0,"delta":{"type":"input_json_delta","partial_json":"test\"}"}}`+"\n\n")
		fmt.Fprint(w, "event: content_block_stop\n")
		fmt.Fprint(w, `data: {"type":"content_block_stop","index":0}`+"\n\n")
		fmt.Fprint(w, "event: message_delta\n")
		fmt.Fprint(w, `data: {"type":"message_delta","usage":{"input_tokens":20,"output_tokens":10}}`+"\n\n")
		fmt.Fprint(w, "event: message_stop\n")
		fmt.Fprint(w, `data: {"type":"message_stop"}`+"\n\n")
	}))
	defer server.Close()

	client := llm.NewAnthropicClient(server.URL, "test-key")
	ch, err := client.StreamChat(context.Background(), []llm.Message{
		{Role: llm.RoleUser, Content: "List apps"},
	}, llm.Options{
		Model:     "claude-3-opus",
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
	assert.Equal(t, "toolu_abc", toolCallEvents[0].ToolCalls[0].ID)
	assert.Equal(t, "list_apps", toolCallEvents[0].ToolCalls[0].Name)
	assert.Equal(t, `{"name":"test"}`, toolCallEvents[0].ToolCalls[0].Arguments)
	require.NotNil(t, doneEvent)
	assert.Equal(t, 30, doneEvent.TokensUsed)
}

func TestAnthropicClient_StreamChat_MultipleToolCalls(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		fmt.Fprint(w, "event: content_block_start\n")
		fmt.Fprint(w, `data: {"type":"content_block_start","index":0,"content_block":{"type":"tool_use","id":"toolu_1","name":"get_app","input":{}}}`+"\n\n")
		fmt.Fprint(w, "event: content_block_delta\n")
		fmt.Fprint(w, `data: {"type":"content_block_delta","index":0,"delta":{"type":"input_json_delta","partial_json":"{\"id\":1}"}}`+"\n\n")
		fmt.Fprint(w, "event: content_block_stop\n")
		fmt.Fprint(w, `data: {"type":"content_block_stop","index":0}`+"\n\n")

		fmt.Fprint(w, "event: content_block_start\n")
		fmt.Fprint(w, `data: {"type":"content_block_start","index":1,"content_block":{"type":"tool_use","id":"toolu_2","name":"get_vendor","input":{}}}`+"\n\n")
		fmt.Fprint(w, "event: content_block_delta\n")
		fmt.Fprint(w, `data: {"type":"content_block_delta","index":1,"delta":{"type":"input_json_delta","partial_json":"{\"id\":2}"}}`+"\n\n")
		fmt.Fprint(w, "event: content_block_stop\n")
		fmt.Fprint(w, `data: {"type":"content_block_stop","index":1}`+"\n\n")

		fmt.Fprint(w, "event: message_stop\n")
		fmt.Fprint(w, `data: {"type":"message_stop"}`+"\n\n")
	}))
	defer server.Close()

	client := llm.NewAnthropicClient(server.URL, "test-key")
	ch, err := client.StreamChat(context.Background(), []llm.Message{
		{Role: llm.RoleUser, Content: "Get info"},
	}, llm.Options{Model: "claude-3-opus", MaxTokens: 100})
	require.NoError(t, err)

	var toolCallEvents []llm.StreamEvent
	for event := range ch {
		if event.Type == llm.EventToolCall {
			toolCallEvents = append(toolCallEvents, event)
		}
	}

	require.Len(t, toolCallEvents, 1)
	require.Len(t, toolCallEvents[0].ToolCalls, 2)
	assert.Equal(t, "toolu_1", toolCallEvents[0].ToolCalls[0].ID)
	assert.Equal(t, "get_app", toolCallEvents[0].ToolCalls[0].Name)
	assert.Equal(t, `{"id":1}`, toolCallEvents[0].ToolCalls[0].Arguments)
	assert.Equal(t, "toolu_2", toolCallEvents[0].ToolCalls[1].ID)
	assert.Equal(t, "get_vendor", toolCallEvents[0].ToolCalls[1].Name)
	assert.Equal(t, `{"id":2}`, toolCallEvents[0].ToolCalls[1].Arguments)
}

func TestAnthropicClient_StreamChat_TextThenToolCall(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		fmt.Fprint(w, "event: content_block_start\n")
		fmt.Fprint(w, `data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}`+"\n\n")
		fmt.Fprint(w, "event: content_block_delta\n")
		fmt.Fprint(w, `data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Let me look that up."}}`+"\n\n")
		fmt.Fprint(w, "event: content_block_stop\n")
		fmt.Fprint(w, `data: {"type":"content_block_stop","index":0}`+"\n\n")

		fmt.Fprint(w, "event: content_block_start\n")
		fmt.Fprint(w, `data: {"type":"content_block_start","index":1,"content_block":{"type":"tool_use","id":"toolu_xyz","name":"search","input":{}}}`+"\n\n")
		fmt.Fprint(w, "event: content_block_delta\n")
		fmt.Fprint(w, `data: {"type":"content_block_delta","index":1,"delta":{"type":"input_json_delta","partial_json":"{\"q\":\"test\"}"}}`+"\n\n")
		fmt.Fprint(w, "event: content_block_stop\n")
		fmt.Fprint(w, `data: {"type":"content_block_stop","index":1}`+"\n\n")

		fmt.Fprint(w, "event: message_stop\n")
		fmt.Fprint(w, `data: {"type":"message_stop"}`+"\n\n")
	}))
	defer server.Close()

	client := llm.NewAnthropicClient(server.URL, "test-key")
	ch, err := client.StreamChat(context.Background(), []llm.Message{
		{Role: llm.RoleUser, Content: "Search for test"},
	}, llm.Options{Model: "claude-3-opus", MaxTokens: 100})
	require.NoError(t, err)

	var tokens []string
	var toolCallEvents []llm.StreamEvent
	for event := range ch {
		switch event.Type {
		case llm.EventToken:
			tokens = append(tokens, event.Content)
		case llm.EventToolCall:
			toolCallEvents = append(toolCallEvents, event)
		}
	}

	assert.Equal(t, []string{"Let me look that up."}, tokens)
	require.Len(t, toolCallEvents, 1)
	require.Len(t, toolCallEvents[0].ToolCalls, 1)
	assert.Equal(t, "toolu_xyz", toolCallEvents[0].ToolCalls[0].ID)
	assert.Equal(t, "search", toolCallEvents[0].ToolCalls[0].Name)
}

func TestAnthropicClient_StreamChat_ToolsInRequest(t *testing.T) {
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
		{Role: llm.RoleUser, Content: "Hi"},
	}, llm.Options{
		Model:     "claude-3-opus",
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

	tools := body["tools"].([]interface{})
	require.Len(t, tools, 1)
	tool := tools[0].(map[string]interface{})
	assert.Equal(t, "list_apps", tool["name"])
}

func TestAnthropicClient_StreamChat_ToolResultMessageInRequest(t *testing.T) {
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
		{Role: llm.RoleUser, Content: "List apps"},
		{Role: llm.RoleAssistant, ToolCalls: []llm.ToolCall{{ID: "toolu_1", Name: "list_apps", Arguments: `{}`}}},
		{Role: llm.RoleTool, Content: `[{"name":"App1"}]`, ToolCallID: "toolu_1"},
	}, llm.Options{Model: "claude-3-opus", MaxTokens: 100})
	require.NoError(t, err)
	for range ch {
	}

	var body map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(receivedBody, &body))

	var messages []json.RawMessage
	require.NoError(t, json.Unmarshal(body["messages"], &messages))
	require.Len(t, messages, 3)

	var assistantMsg map[string]interface{}
	require.NoError(t, json.Unmarshal(messages[1], &assistantMsg))
	assert.Equal(t, "assistant", assistantMsg["role"])
	content := assistantMsg["content"].([]interface{})
	require.Len(t, content, 1)
	block := content[0].(map[string]interface{})
	assert.Equal(t, "tool_use", block["type"])
	assert.Equal(t, "toolu_1", block["id"])
	assert.Equal(t, "list_apps", block["name"])

	var toolMsg map[string]interface{}
	require.NoError(t, json.Unmarshal(messages[2], &toolMsg))
	assert.Equal(t, "user", toolMsg["role"])
	toolContent := toolMsg["content"].([]interface{})
	require.Len(t, toolContent, 1)
	toolBlock := toolContent[0].(map[string]interface{})
	assert.Equal(t, "tool_result", toolBlock["type"])
	assert.Equal(t, "toolu_1", toolBlock["tool_use_id"])
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
