package orchestrator_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"easi/backend/internal/archassistant/application/orchestrator"
	"easi/backend/internal/archassistant/application/tools"
	"easi/backend/internal/archassistant/domain/aggregates"
	vo "easi/backend/internal/archassistant/domain/valueobjects"
	"easi/backend/internal/archassistant/infrastructure/adapters"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockStreamWriter struct {
	mu          sync.Mutex
	tokens      []string
	toolStarts  []orchestrator.ToolCallStartEvent
	toolResults []orchestrator.ToolCallResultEvent
	thinkings   []orchestrator.ThinkingEvent
	doneMsg     string
	doneTokens  int
}

func (m *mockStreamWriter) WriteToken(content string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tokens = append(m.tokens, content)
	return nil
}

func (m *mockStreamWriter) WriteDone(messageID string, tokensUsed int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.doneMsg = messageID
	m.doneTokens = tokensUsed
	return nil
}

func (m *mockStreamWriter) WriteToolCallStart(event orchestrator.ToolCallStartEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.toolStarts = append(m.toolStarts, event)
	return nil
}

func (m *mockStreamWriter) WriteToolCallResult(event orchestrator.ToolCallResultEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.toolResults = append(m.toolResults, event)
	return nil
}

func (m *mockStreamWriter) WriteThinking(event orchestrator.ThinkingEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.thinkings = append(m.thinkings, event)
	return nil
}

type mockPermissions struct {
	allowed map[string]bool
}

func (m *mockPermissions) HasPermission(permission string) bool {
	return m.allowed[permission]
}

func allPermissions() *mockPermissions {
	return &mockPermissions{allowed: map[string]bool{
		"arch:read":  true,
		"arch:write": true,
	}}
}

type mockToolExecutor struct {
	result tools.ToolResult
}

func (m *mockToolExecutor) Execute(_ context.Context, _ map[string]interface{}) tools.ToolResult {
	return m.result
}

func newTestRegistry(executors map[string]tools.ToolResult) *tools.Registry {
	registry := tools.NewRegistry()
	for name, result := range executors {
		registry.Register(tools.ToolDefinition{
			Name:       name,
			Permission: "arch:read",
			Access:     tools.AccessRead,
		}, &mockToolExecutor{result: result})
	}
	return registry
}

func hasToolCallMessages(body []byte) bool {
	type requestBody struct {
		Messages []json.RawMessage `json:"messages"`
	}
	var req requestBody
	if err := json.Unmarshal(body, &req); err != nil {
		return false
	}
	for _, m := range req.Messages {
		if strings.Contains(string(m), `"tool"`) {
			return true
		}
	}
	return false
}

type toolCallSpec struct {
	id   string
	name string
	args string
}

func toolCallResponse(tc toolCallSpec) string {
	return fmt.Sprintf(
		`data: {"choices":[{"delta":{"tool_calls":[{"index":0,"id":"%s","function":{"name":"%s","arguments":"%s"}}]}}]}`+"\n\n"+
			"data: [DONE]\n\n",
		tc.id, tc.name, tc.args,
	)
}

func textResponse(content string) string {
	return fmt.Sprintf(
		`data: {"choices":[{"delta":{"content":"%s"}}]}`+"\n\n"+
			"data: [DONE]\n\n",
		content,
	)
}

type toolTestFixture struct {
	repo     *mockConversationRepo
	orch     *orchestrator.Orchestrator
	writer   *mockStreamWriter
	conv     *aggregates.Conversation
	registry *tools.Registry
}

func setupToolTest(t *testing.T, registry *tools.Registry, handler http.Handler) (*toolTestFixture, *httptest.Server) {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	repo := &mockConversationRepo{}
	factory := adapters.NewLLMClientFactory()
	orch := orchestrator.New(repo, factory)
	writer := &mockStreamWriter{}
	conv := aggregates.NewConversation("tenant-1", "user-1")
	repo.conversation = conv

	return &toolTestFixture{repo: repo, orch: orch, writer: writer, conv: conv, registry: registry}, server
}

func (f *toolTestFixture) sendMessage(t *testing.T, serverURL, content string, perms tools.PermissionChecker) error {
	t.Helper()
	return f.orch.SendMessage(context.Background(), f.writer, orchestrator.SendMessageParams{
		ConversationID: f.conv.ID(),
		UserID:         "user-1",
		Content:        content,
		TenantID:       "tenant-1",
		UserRole:       "architect",
		Config:         testConfig(serverURL),
		Permissions:    perms,
		ToolRegistry:   f.registry,
	})
}

func toolThenTextHandler(tc toolCallSpec, finalText string) http.Handler {
	var callCount atomic.Int32
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		if callCount.Add(1) == 1 && !hasToolCallMessages(body) {
			w.Write([]byte(toolCallResponse(tc)))
			return
		}
		w.Write([]byte(textResponse(finalText)))
	})
}

func repeatingToolHandler(callsBeforeText int32) http.Handler {
	var callCount atomic.Int32
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		n := callCount.Add(1)
		if n <= callsBeforeText {
			w.Write([]byte(toolCallResponse(toolCallSpec{fmt.Sprintf("call-%d", n), "list_apps", "{}"})))
			return
		}
		w.Write([]byte(textResponse("Final answer")))
	})
}

func countLimitExceeded(results []orchestrator.ToolCallResultEvent) int {
	var count int
	for _, r := range results {
		if strings.Contains(r.ResultPreview, "limit exceeded") {
			count++
		}
	}
	return count
}

func TestOrchestrator_SendMessage_WithToolCalls(t *testing.T) {
	registry := newTestRegistry(map[string]tools.ToolResult{
		"list_apps": {Content: `[{"name":"App1"},{"name":"App2"}]`},
	})
	fix, server := setupToolTest(t, registry, toolThenTextHandler(toolCallSpec{"call-1", "list_apps", "{}"}, "Here are the apps."))

	err := fix.sendMessage(t, server.URL, "List all apps", allPermissions())
	require.NoError(t, err)

	assert.Equal(t, []string{"Here are the apps."}, fix.writer.tokens)
	assert.NotEmpty(t, fix.writer.doneMsg)

	require.Len(t, fix.repo.savedMsgs, 2)
	assert.Equal(t, vo.MessageRoleUser, fix.repo.savedMsgs[0].Role())
	assert.Equal(t, vo.MessageRoleAssistant, fix.repo.savedMsgs[1].Role())
	assert.Equal(t, "Here are the apps.", fix.repo.savedMsgs[1].Content())
}

func TestOrchestrator_SendMessage_ToolCallStreamsEvents(t *testing.T) {
	registry := newTestRegistry(map[string]tools.ToolResult{
		"list_apps": {Content: `[{"name":"App1"}]`},
	})
	fix, server := setupToolTest(t, registry, toolThenTextHandler(toolCallSpec{"call-1", "list_apps", "{}"}, "Done"))

	err := fix.sendMessage(t, server.URL, "List apps", allPermissions())
	require.NoError(t, err)

	require.Len(t, fix.writer.toolStarts, 1)
	assert.Equal(t, "call-1", fix.writer.toolStarts[0].ToolCallID)
	assert.Equal(t, "list_apps", fix.writer.toolStarts[0].Name)

	require.Len(t, fix.writer.toolResults, 1)
	assert.Equal(t, "call-1", fix.writer.toolResults[0].ToolCallID)
	assert.Equal(t, "list_apps", fix.writer.toolResults[0].Name)
	assert.NotEmpty(t, fix.writer.toolResults[0].ResultPreview)

	require.Len(t, fix.writer.thinkings, 1)
}

func TestOrchestrator_SendMessage_MaxIterations(t *testing.T) {
	alwaysToolCall := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(toolCallResponse(toolCallSpec{"call-1", "list_apps", "{}"})))
	})

	registry := newTestRegistry(map[string]tools.ToolResult{
		"list_apps": {Content: `[]`},
	})
	fix, server := setupToolTest(t, registry, alwaysToolCall)

	err := fix.sendMessage(t, server.URL, "Keep calling tools", allPermissions())

	require.Error(t, err)
	var iterErr *orchestrator.IterationLimitError
	assert.ErrorAs(t, err, &iterErr)
}

func TestOrchestrator_SendMessage_NoToolsWhenRegistryNil(t *testing.T) {
	var requestBody []byte
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(textResponse("Hello")))
	})

	fix, server := setupToolTest(t, nil, handler)

	err := fix.sendMessage(t, server.URL, "Hello", nil)
	require.NoError(t, err)

	assert.NotContains(t, string(requestBody), `"tools"`)
	assert.Equal(t, []string{"Hello"}, fix.writer.tokens)
	assert.NotEmpty(t, fix.writer.doneMsg)
}

func TestOrchestrator_SendMessage_MaxSameToolCalls(t *testing.T) {
	tests := []struct {
		name            string
		callsBeforeText int32
		toolResult      string
		expectedCalls   int
	}{
		{"within limit", 8, `[]`, 8},
		{"exact boundary", 5, `[{"name":"App1"}]`, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := newTestRegistry(map[string]tools.ToolResult{
				"list_apps": {Content: tt.toolResult},
			})
			fix, server := setupToolTest(t, registry, repeatingToolHandler(tt.callsBeforeText))

			err := fix.sendMessage(t, server.URL, "Call the same tool many times", allPermissions())
			require.NoError(t, err)

			assert.Equal(t, 0, countLimitExceeded(fix.writer.toolResults), "calls should be within maxSameToolCalls=10")
			assert.Equal(t, tt.expectedCalls, len(fix.writer.toolResults))
		})
	}
}

func newRegistryWithAccess(name string, access tools.AccessClass) *tools.Registry {
	registry := tools.NewRegistry()
	registry.Register(tools.ToolDefinition{
		Name:       name,
		Permission: "arch:write",
		Access:     access,
	}, &mockToolExecutor{result: tools.ToolResult{Content: "ok"}})
	return registry
}

func repeatingNamedToolHandler(toolName string, callsBeforeText int32) http.Handler {
	var callCount atomic.Int32
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		n := callCount.Add(1)
		if n <= callsBeforeText {
			w.Write([]byte(toolCallResponse(toolCallSpec{fmt.Sprintf("call-%d", n), toolName, "{}"})))
			return
		}
		w.Write([]byte(textResponse("Final answer")))
	})
}

func TestOrchestrator_SendMessage_DeleteToolCappedAt5(t *testing.T) {
	registry := newRegistryWithAccess("delete_app", tools.AccessDelete)
	fix, server := setupToolTest(t, registry, repeatingNamedToolHandler("delete_app", 10))

	err := fix.sendMessage(t, server.URL, "Delete many apps", allPermissions())
	require.NoError(t, err)

	exceeded := countLimitExceeded(fix.writer.toolResults)
	assert.Equal(t, 5, exceeded, "calls 6-10 should be rejected")

	successful := len(fix.writer.toolResults) - exceeded
	assert.Equal(t, 5, successful, "only first 5 delete calls should succeed")
}

func TestOrchestrator_SendMessage_UpdateToolCappedAt100(t *testing.T) {
	registry := newRegistryWithAccess("update_app", tools.AccessUpdate)
	fix, server := setupToolTest(t, registry, repeatingNamedToolHandler("update_app", 8))

	err := fix.sendMessage(t, server.URL, "Update many apps", allPermissions())
	require.NoError(t, err)

	assert.Equal(t, 0, countLimitExceeded(fix.writer.toolResults), "8 update calls within 100 limit")
	assert.Equal(t, 8, len(fix.writer.toolResults))
}

func TestOrchestrator_SendMessage_ReadToolAllows500(t *testing.T) {
	registry := newTestRegistry(map[string]tools.ToolResult{
		"list_apps": {Content: `[]`},
	})
	fix, server := setupToolTest(t, registry, repeatingToolHandler(8))

	err := fix.sendMessage(t, server.URL, "Read many times", allPermissions())
	require.NoError(t, err)

	assert.Equal(t, 0, countLimitExceeded(fix.writer.toolResults), "8 read calls should all succeed within 500 limit")
}

func TestOrchestrator_SendMessage_ToolCallError(t *testing.T) {
	registry := newTestRegistry(map[string]tools.ToolResult{
		"failing_tool": {Content: "something went wrong", IsError: true},
	})
	fix, server := setupToolTest(t, registry, toolThenTextHandler(toolCallSpec{"call-1", "failing_tool", "{}"}, "Recovered"))

	err := fix.sendMessage(t, server.URL, "Do something", allPermissions())
	require.NoError(t, err)

	assert.Equal(t, []string{"Recovered"}, fix.writer.tokens)
	require.Len(t, fix.writer.toolResults, 1)
}

func TestOrchestrator_SendMessage_NoToolsWhenPartialConfig(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(textResponse("No tools response")))
	})

	tests := []struct {
		name     string
		registry *tools.Registry
		perms    tools.PermissionChecker
	}{
		{
			name:     "registry but nil permissions",
			registry: newTestRegistry(map[string]tools.ToolResult{"list_apps": {Content: `[]`}}),
			perms:    nil,
		},
		{
			name:     "permissions but nil registry",
			registry: nil,
			perms:    allPermissions(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fix, server := setupToolTest(t, tt.registry, handler)
			err := fix.sendMessage(t, server.URL, "Hello", tt.perms)
			require.NoError(t, err)
			assert.Equal(t, []string{"No tools response"}, fix.writer.tokens)
		})
	}
}

func TestOrchestrator_SendMessage_ToolResultTruncation(t *testing.T) {
	longResult := strings.Repeat("x", 20000)
	registry := newTestRegistry(map[string]tools.ToolResult{
		"big_tool": {Content: longResult},
	})
	fix, server := setupToolTest(t, registry, toolThenTextHandler(toolCallSpec{"call-1", "big_tool", "{}"}, "Summary"))

	err := fix.sendMessage(t, server.URL, "Get big data", allPermissions())
	require.NoError(t, err)

	require.Len(t, fix.writer.toolResults, 1)
	assert.LessOrEqual(t, len(fix.writer.toolResults[0].ResultPreview), 203)
}

func TestOrchestrator_SendMessage_PreviewTruncation(t *testing.T) {
	longResult := strings.Repeat("a", 300)
	registry := newTestRegistry(map[string]tools.ToolResult{
		"verbose_tool": {Content: longResult},
	})
	fix, server := setupToolTest(t, registry, toolThenTextHandler(toolCallSpec{"call-1", "verbose_tool", "{}"}, "Done"))

	err := fix.sendMessage(t, server.URL, "Get verbose data", allPermissions())
	require.NoError(t, err)

	require.Len(t, fix.writer.toolResults, 1)
	preview := fix.writer.toolResults[0].ResultPreview
	assert.LessOrEqual(t, len(preview), 203, "preview should be truncated to ~200 chars + ellipsis")
	assert.True(t, strings.HasSuffix(preview, "..."), "truncated preview should end with ...")
}

func textToolCallResponse(preamble, toolName, toolArgs string) string {
	content := fmt.Sprintf(
		"%s\n<tool_call>\n{\"name\": \"%s\", \"arguments\": %s}\n</tool_call>\n<tool_response>\n{\"fake\": \"data\"}\n</tool_response>",
		preamble, toolName, toolArgs,
	)
	encoded, _ := json.Marshal(content)
	raw := string(encoded[1 : len(encoded)-1])
	return fmt.Sprintf(
		`data: {"choices":[{"delta":{"content":"%s"}}]}`+"\n\n"+
			"data: [DONE]\n\n",
		raw,
	)
}

func TestOrchestrator_SendMessage_TextToolCallFallback(t *testing.T) {
	var callCount atomic.Int32
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		n := callCount.Add(1)
		if n == 1 && !hasToolCallMessages(body) {
			w.Write([]byte(textToolCallResponse(
				"I'll look up the applications for you.",
				"list_apps",
				`{}`,
			)))
			return
		}
		w.Write([]byte(textResponse("You have 2 real applications: Word and Excel.")))
	})

	registry := newTestRegistry(map[string]tools.ToolResult{
		"list_apps": {Content: `[{"name":"Word"},{"name":"Excel"}]`},
	})
	fix, server := setupToolTest(t, registry, handler)

	err := fix.sendMessage(t, server.URL, "List applications in my portfolio", allPermissions())
	require.NoError(t, err)

	require.Len(t, fix.writer.toolStarts, 1)
	assert.Equal(t, "list_apps", fix.writer.toolStarts[0].Name)
	assert.True(t, strings.HasPrefix(fix.writer.toolStarts[0].ToolCallID, "text-tc-"))

	require.Len(t, fix.writer.toolResults, 1)
	assert.Equal(t, "list_apps", fix.writer.toolResults[0].Name)

	lastSaved := fix.repo.savedMsgs[len(fix.repo.savedMsgs)-1]
	assert.Contains(t, lastSaved.Content(), "Word")
	assert.Contains(t, lastSaved.Content(), "Excel")
	assert.NotContains(t, lastSaved.Content(), "tool_call")

	for _, tok := range fix.writer.tokens {
		assert.NotContains(t, tok, "tool_call", "hallucinated XML must not reach the client")
		assert.NotContains(t, tok, "tool_response", "hallucinated response must not reach the client")
	}
}

func functionCallsResponse(preamble, toolName string, params map[string]string) string {
	var paramXML strings.Builder
	for k, v := range params {
		fmt.Fprintf(&paramXML, `<parameter name="%s">%s</parameter> `, k, v)
	}
	content := fmt.Sprintf(
		`%s`+"\n"+`<function_calls> <invoke name="%s"> %s</invoke> </function_calls> `+
			`<function_result> <invoke name="%s"> {"fake":"data"} </invoke> </function_calls>`,
		preamble, toolName, paramXML.String(), toolName,
	)
	encoded, _ := json.Marshal(content)
	raw := string(encoded[1 : len(encoded)-1])
	return fmt.Sprintf(
		`data: {"choices":[{"delta":{"content":"%s"}}]}`+"\n\n"+
			"data: [DONE]\n\n",
		raw,
	)
}

func TestOrchestrator_SendMessage_FunctionCallsFallback(t *testing.T) {
	var callCount atomic.Int32
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		n := callCount.Add(1)
		if n == 1 && !hasToolCallMessages(body) {
			w.Write([]byte(functionCallsResponse(
				"I'll look up the applications in your enterprise right away.",
				"get_applications",
				map[string]string{"tenant": "acme"},
			)))
			return
		}
		w.Write([]byte(textResponse("You have 2 real applications: Word and Excel.")))
	})

	registry := newTestRegistry(map[string]tools.ToolResult{
		"list_applications": {Content: `Found 2 applications:\n1. Word\n2. Excel`},
	})
	fix, server := setupToolTest(t, registry, handler)

	err := fix.sendMessage(t, server.URL, "What applications are in my enterprise?", allPermissions())
	require.NoError(t, err)

	require.Len(t, fix.writer.toolStarts, 1)
	assert.Equal(t, "list_applications", fix.writer.toolStarts[0].Name, "hallucinated 'get_applications' should resolve to 'list_applications'")

	require.Len(t, fix.writer.toolResults, 1)

	lastSaved := fix.repo.savedMsgs[len(fix.repo.savedMsgs)-1]
	assert.Contains(t, lastSaved.Content(), "Word")
	assert.NotContains(t, lastSaved.Content(), "function_calls")

	for _, tok := range fix.writer.tokens {
		assert.NotContains(t, tok, "function_calls", "hallucinated XML must not reach the client")
		assert.NotContains(t, tok, "function_result", "hallucinated result must not reach the client")
		assert.NotContains(t, tok, "get_applications", "hallucinated tool name must not reach the client")
	}
}

func TestOrchestrator_SendMessage_TextToolCallFallback_NoToolCallsPassesThrough(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(textResponse("Just a plain answer, no tools needed.")))
	})

	registry := newTestRegistry(map[string]tools.ToolResult{
		"list_apps": {Content: `[]`},
	})
	fix, server := setupToolTest(t, registry, handler)

	err := fix.sendMessage(t, server.URL, "Hello", allPermissions())
	require.NoError(t, err)

	assert.Empty(t, fix.writer.toolStarts)
	assert.Equal(t, []string{"Just a plain answer, no tools needed."}, fix.writer.tokens)
}
