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

// sseLines writes SSE-formatted data lines to a ResponseWriter.
func writeSSELines(w http.ResponseWriter, lines ...string) {
	for _, line := range lines {
		_, _ = fmt.Fprintf(w, "data: %s\n\n", line)
	}
}

func TestResponsesAPIClient_StreamChat_TextTokens(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))

		var body map[string]interface{}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		// must use "input", not "messages"
		assert.Contains(t, body, "input")
		assert.NotContains(t, body, "messages")
		// max_output_tokens, not max_tokens
		assert.Contains(t, body, "max_output_tokens")
		assert.NotContains(t, body, "max_tokens")

		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		writeSSELines(w,
			`{"type":"response.output_text.delta","delta":"Hello"}`,
			`{"type":"response.output_text.delta","delta":" world"}`,
			`{"type":"response.completed","response":{"usage":{"total_tokens":15}}}`,
		)
	}))
	defer server.Close()

	client := llm.NewResponsesAPIClient(server.URL, "test-key")
	ch, err := client.StreamChat(context.Background(), []llm.Message{
		{Role: llm.RoleSystem, Content: "You are helpful."},
		{Role: llm.RoleUser, Content: "Hi"},
	}, llm.Options{Model: "gpt-4o", MaxTokens: 100, Temperature: 0.5})

	require.NoError(t, err)

	var tokens []string
	var doneEvent *llm.StreamEvent
	for e := range ch {
		switch e.Type {
		case llm.EventToken:
			tokens = append(tokens, e.Content)
		case llm.EventDone:
			ev := e
			doneEvent = &ev
		}
	}

	assert.Equal(t, []string{"Hello", " world"}, tokens)
	require.NotNil(t, doneEvent)
	assert.Equal(t, 15, doneEvent.TokensUsed)
}

func TestResponsesAPIClient_StreamChat_NonOKStatusReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	client := llm.NewResponsesAPIClient(server.URL, "bad-key")
	_, err := client.StreamChat(context.Background(), []llm.Message{
		{Role: llm.RoleUser, Content: "Hi"},
	}, llm.Options{Model: "gpt-4o", MaxTokens: 10})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "401")
}

func TestResponsesAPIClient_StreamChat_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		flusher := w.(http.Flusher)
		_, _ = fmt.Fprint(w, `data: {"type":"response.output_text.delta","delta":"Hi"}`+"\n\n")
		flusher.Flush()
		<-r.Context().Done()
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	client := llm.NewResponsesAPIClient(server.URL, "test-key")
	ch, err := client.StreamChat(ctx, []llm.Message{
		{Role: llm.RoleUser, Content: "Hi"},
	}, llm.Options{Model: "gpt-4o", MaxTokens: 100})
	require.NoError(t, err)

	event := <-ch
	assert.Equal(t, llm.EventToken, event.Type)
	cancel()
	for range ch {
	}
}

func TestResponsesAPIClient_StreamChat_BuildsInputCorrectly(t *testing.T) {
	var receivedBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedBody, _ = readAll(r.Body)
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		writeSSELines(w, `{"type":"response.completed","response":{}}`)
	}))
	defer server.Close()

	client := llm.NewResponsesAPIClient(server.URL, "test-key")
	ch, err := client.StreamChat(context.Background(), []llm.Message{
		{Role: llm.RoleSystem, Content: "Be helpful."},
		{Role: llm.RoleUser, Content: "Hello"},
	}, llm.Options{Model: "gpt-4o", MaxTokens: 50, Temperature: 1.0})
	require.NoError(t, err)
	for range ch {
	}

	var body map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(receivedBody, &body))

	// system message becomes instructions field
	var instructions string
	require.NoError(t, json.Unmarshal(body["instructions"], &instructions))
	assert.Equal(t, "Be helpful.", instructions)

	// user message is in input array
	var input []map[string]interface{}
	require.NoError(t, json.Unmarshal(body["input"], &input))
	require.Len(t, input, 1)
	assert.Equal(t, "user", input[0]["role"])
	assert.Equal(t, "Hello", input[0]["content"])
}

func TestResponsesAPIClient_StreamChat_ToolCallEvent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		writeSSELines(w,
			`{"type":"response.output_item.added","output_index":0,"item":{"type":"function_call","id":"item_abc","call_id":"call_xyz","name":"list_apps"}}`,
			`{"type":"response.function_call_arguments.delta","item_id":"item_abc","delta":"{\"na"}`,
			`{"type":"response.function_call_arguments.delta","item_id":"item_abc","delta":"me\":\"test\"}"}`,
			`{"type":"response.function_call_arguments.done","item_id":"item_abc","name":"list_apps","arguments":"{\"name\":\"test\"}"}`,
			`{"type":"response.completed","response":{"usage":{"total_tokens":30}}}`,
		)
	}))
	defer server.Close()

	client := llm.NewResponsesAPIClient(server.URL, "test-key")
	ch, err := client.StreamChat(context.Background(), []llm.Message{
		{Role: llm.RoleUser, Content: "List apps"},
	}, llm.Options{
		Model:     "gpt-4o",
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
	for e := range ch {
		switch e.Type {
		case llm.EventToolCall:
			toolCallEvents = append(toolCallEvents, e)
		case llm.EventDone:
			ev := e
			doneEvent = &ev
		}
	}

	require.Len(t, toolCallEvents, 1)
	require.Len(t, toolCallEvents[0].ToolCalls, 1)
	tc := toolCallEvents[0].ToolCalls[0]
	// The ID should be the call_id from output_item.added, not item_id.
	assert.Equal(t, "call_xyz", tc.ID)
	assert.Equal(t, "list_apps", tc.Name)
	assert.Equal(t, `{"name":"test"}`, tc.Arguments)
	require.NotNil(t, doneEvent)
	assert.Equal(t, 30, doneEvent.TokensUsed)
}

func TestResponsesAPIClient_StreamChat_MultipleToolCalls(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		writeSSELines(w,
			`{"type":"response.output_item.added","output_index":0,"item":{"type":"function_call","id":"item_1","call_id":"call_1","name":"get_app"}}`,
			`{"type":"response.function_call_arguments.done","item_id":"item_1","name":"get_app","arguments":"{\"id\":1}"}`,
			`{"type":"response.output_item.added","output_index":1,"item":{"type":"function_call","id":"item_2","call_id":"call_2","name":"get_vendor"}}`,
			`{"type":"response.function_call_arguments.done","item_id":"item_2","name":"get_vendor","arguments":"{\"id\":2}"}`,
			`{"type":"response.completed","response":{}}`,
		)
	}))
	defer server.Close()

	client := llm.NewResponsesAPIClient(server.URL, "test-key")
	ch, err := client.StreamChat(context.Background(), []llm.Message{
		{Role: llm.RoleUser, Content: "Get info"},
	}, llm.Options{Model: "gpt-4o", MaxTokens: 100})
	require.NoError(t, err)

	var toolCallEvents []llm.StreamEvent
	for e := range ch {
		if e.Type == llm.EventToolCall {
			toolCallEvents = append(toolCallEvents, e)
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

func TestResponsesAPIClient_StreamChat_ToolsInRequest(t *testing.T) {
	var receivedBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedBody, _ = readAll(r.Body)
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		writeSSELines(w, `{"type":"response.completed","response":{}}`)
	}))
	defer server.Close()

	client := llm.NewResponsesAPIClient(server.URL, "test-key")
	ch, err := client.StreamChat(context.Background(), []llm.Message{
		{Role: llm.RoleUser, Content: "Hi"},
	}, llm.Options{
		Model:     "gpt-4o",
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
	// Responses API format has flat name/description/parameters (no nested "function" key)
	assert.Equal(t, "function", tool["type"])
	assert.Equal(t, "list_apps", tool["name"])
	assert.Equal(t, "Lists applications", tool["description"])
	assert.Contains(t, tool, "parameters")
}

func TestResponsesAPIClient_StreamChat_ToolCallInputInRequest(t *testing.T) {
	var receivedBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedBody, _ = readAll(r.Body)
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		writeSSELines(w, `{"type":"response.completed","response":{}}`)
	}))
	defer server.Close()

	client := llm.NewResponsesAPIClient(server.URL, "test-key")
	ch, err := client.StreamChat(context.Background(), []llm.Message{
		{Role: llm.RoleUser, Content: "List apps"},
		{Role: llm.RoleAssistant, ToolCalls: []llm.ToolCall{{ID: "call_1", Name: "list_apps", Arguments: `{"name":"x"}`}}},
		{Role: llm.RoleTool, Content: `[{"name":"App1"}]`, ToolCallID: "call_1"},
	}, llm.Options{Model: "gpt-4o", MaxTokens: 100})
	require.NoError(t, err)
	for range ch {
	}

	var body map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(receivedBody, &body))

	var input []json.RawMessage
	require.NoError(t, json.Unmarshal(body["input"], &input))
	require.Len(t, input, 3)

	// First: user message
	var userMsg map[string]interface{}
	require.NoError(t, json.Unmarshal(input[0], &userMsg))
	assert.Equal(t, "user", userMsg["role"])

	// Second: function_call item (assistant tool call)
	var fcItem map[string]interface{}
	require.NoError(t, json.Unmarshal(input[1], &fcItem))
	assert.Equal(t, "function_call", fcItem["type"])
	assert.Equal(t, "call_1", fcItem["call_id"])
	assert.Equal(t, "list_apps", fcItem["name"])
	assert.Equal(t, `{"name":"x"}`, fcItem["arguments"])

	// Third: function_call_output item (tool result)
	var fcoItem map[string]interface{}
	require.NoError(t, json.Unmarshal(input[2], &fcoItem))
	assert.Equal(t, "function_call_output", fcoItem["type"])
	assert.Equal(t, "call_1", fcoItem["call_id"])
	assert.Equal(t, `[{"name":"App1"}]`, fcoItem["output"])
}

// TestResponsesAPIClient_StreamChat_ToolCallNameFromOutputItem verifies that
// when response.function_call_arguments.done omits the "name" field (as Azure
// AI Foundry does), the name is resolved from the earlier output_item.added event.
func TestResponsesAPIClient_StreamChat_ToolCallNameFromOutputItem(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		writeSSELines(w,
			// name present in output_item.added
			`{"type":"response.output_item.added","output_index":0,"item":{"type":"function_call","id":"item_1","call_id":"call_xyz","name":"list_apps"}}`,
			// name intentionally absent from arguments.done (Azure pattern)
			`{"type":"response.function_call_arguments.done","item_id":"item_1","arguments":"{\"id\":1}"}`,
			`{"type":"response.completed","response":{}}`,
		)
	}))
	defer server.Close()

	client := llm.NewResponsesAPIClient(server.URL, "test-key")
	ch, err := client.StreamChat(context.Background(), []llm.Message{
		{Role: llm.RoleUser, Content: "List apps"},
	}, llm.Options{Model: "gpt-4o", MaxTokens: 100})
	require.NoError(t, err)

	var toolCallEvents []llm.StreamEvent
	for e := range ch {
		if e.Type == llm.EventToolCall {
			toolCallEvents = append(toolCallEvents, e)
		}
	}

	require.Len(t, toolCallEvents, 1)
	require.Len(t, toolCallEvents[0].ToolCalls, 1)
	assert.Equal(t, "call_xyz", toolCallEvents[0].ToolCalls[0].ID)
	assert.Equal(t, "list_apps", toolCallEvents[0].ToolCalls[0].Name, "name should be resolved from output_item.added")
	assert.Equal(t, `{"id":1}`, toolCallEvents[0].ToolCalls[0].Arguments)
}

func TestResponsesAPIClient_StreamChat_ErrorEvent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		writeSSELines(w,
			`{"type":"response.output_text.delta","delta":"Hello"}`,
			`{"type":"error","message":"something went wrong"}`,
		)
	}))
	defer server.Close()

	client := llm.NewResponsesAPIClient(server.URL, "test-key")
	ch, err := client.StreamChat(context.Background(), []llm.Message{
		{Role: llm.RoleUser, Content: "Hi"},
	}, llm.Options{Model: "gpt-4o", MaxTokens: 100})
	require.NoError(t, err)

	var errorEvent *llm.StreamEvent
	for e := range ch {
		if e.Type == llm.EventError {
			ev := e
			errorEvent = &ev
		}
	}

	require.NotNil(t, errorEvent)
	assert.Contains(t, errorEvent.Error.Error(), "something went wrong")
}

func TestResponsesToolCallAccumulator_NormalFlow(t *testing.T) {
	acc := llm.NewResponsesToolCallAccumulator()

	acc.RegisterItem("item_1", "call_abc", "list_apps")
	acc.Finalize("item_1", "list_apps", `{"name":"x"}`)

	ch := make(chan llm.StreamEvent, 4)
	acc.Emit(ch)
	close(ch)

	var events []llm.StreamEvent
	for e := range ch {
		events = append(events, e)
	}

	require.Len(t, events, 1)
	require.Len(t, events[0].ToolCalls, 1)
	assert.Equal(t, "call_abc", events[0].ToolCalls[0].ID)
	assert.Equal(t, "list_apps", events[0].ToolCalls[0].Name)
	assert.Equal(t, `{"name":"x"}`, events[0].ToolCalls[0].Arguments)
}

func TestResponsesToolCallAccumulator_NameFallback(t *testing.T) {
	// Azure omits "name" from function_call_arguments.done; name must come from registerItem.
	acc := llm.NewResponsesToolCallAccumulator()

	acc.RegisterItem("item_1", "call_abc", "list_apps")
	acc.Finalize("item_1", "", `{"name":"x"}`) // name intentionally blank

	ch := make(chan llm.StreamEvent, 4)
	acc.Emit(ch)
	close(ch)

	var events []llm.StreamEvent
	for e := range ch {
		events = append(events, e)
	}

	require.Len(t, events, 1)
	assert.Equal(t, "list_apps", events[0].ToolCalls[0].Name, "name must fall back to value from output_item.added")
}

func TestResponsesToolCallAccumulator_MultipleItems(t *testing.T) {
	acc := llm.NewResponsesToolCallAccumulator()

	acc.RegisterItem("item_1", "call_1", "get_app")
	acc.RegisterItem("item_2", "call_2", "get_vendor")
	acc.Finalize("item_1", "get_app", `{"id":1}`)
	acc.Finalize("item_2", "get_vendor", `{"id":2}`)

	ch := make(chan llm.StreamEvent, 4)
	acc.Emit(ch)
	close(ch)

	var events []llm.StreamEvent
	for e := range ch {
		events = append(events, e)
	}

	require.Len(t, events, 1)
	require.Len(t, events[0].ToolCalls, 2)
	assert.Equal(t, "call_1", events[0].ToolCalls[0].ID)
	assert.Equal(t, "call_2", events[0].ToolCalls[1].ID)
}

func TestResponsesToolCallAccumulator_EmitWithoutCalls(t *testing.T) {
	acc := llm.NewResponsesToolCallAccumulator()
	ch := make(chan llm.StreamEvent, 4)
	acc.Emit(ch) // should not send anything
	close(ch)
	assert.Empty(t, ch)
}

func TestIsResponsesAPIEndpoint(t *testing.T) {
	cases := []struct {
		url      string
		expected bool
	}{
		{"https://cog-easi.cognitiveservices.azure.com/openai/responses?api-version=2025-04-01-preview", true},
		{"https://api.openai.com/v1/responses", true},
		{"https://api.openai.com", false},
		{"https://api.openai.com/v1/chat/completions", false},
		{"https://api.anthropic.com/v1/messages", false},
		{"not-a-url", false},
	}

	for _, tc := range cases {
		t.Run(tc.url, func(t *testing.T) {
			assert.Equal(t, tc.expected, llm.IsResponsesAPIEndpoint(tc.url))
		})
	}
}

func TestNewClientFactory_RoutesToResponsesAPIClient(t *testing.T) {
	// Verifies the factory selects ResponsesAPIClient for /responses endpoints.
	var receivedBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedBody, _ = readAll(r.Body)
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		writeSSELines(w, `{"type":"response.completed","response":{}}`)
	}))
	defer server.Close()

	endpoint := server.URL + "/openai/responses?api-version=2025-04-01-preview"
	client, err := llm.NewClient("openai", endpoint, "key")
	require.NoError(t, err)

	ch, err := client.StreamChat(context.Background(), []llm.Message{
		{Role: llm.RoleUser, Content: "Hello"},
	}, llm.Options{Model: "gpt-4o", MaxTokens: 10})
	require.NoError(t, err)
	for range ch {
	}

	var body map[string]interface{}
	require.NoError(t, json.Unmarshal(receivedBody, &body))
	// Responses API uses "input", not "messages"
	assert.Contains(t, body, "input")
	assert.NotContains(t, body, "messages")
}
