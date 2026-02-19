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
	"easi/backend/internal/archassistant/domain"
	"easi/backend/internal/archassistant/domain/aggregates"
	"easi/backend/internal/archassistant/infrastructure/adapters"
	vo "easi/backend/internal/archassistant/domain/valueobjects"
	"easi/backend/internal/archassistant/infrastructure/sse"
	"easi/backend/internal/archassistant/publishedlanguage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockConversationRepo struct {
	conversation *aggregates.Conversation
	messages     []*aggregates.Message
	savedMsgs    []*aggregates.Message
	updatedConvs []*aggregates.Conversation
	createdConvs []*aggregates.Conversation
	convCount    int
}

func (m *mockConversationRepo) Create(ctx context.Context, conv *aggregates.Conversation) error {
	m.createdConvs = append(m.createdConvs, conv)
	return nil
}

func (m *mockConversationRepo) GetByIDAndUser(ctx context.Context, id, userID string) (*aggregates.Conversation, error) {
	return m.conversation, nil
}

func (m *mockConversationRepo) SaveMessage(ctx context.Context, msg *aggregates.Message) error {
	m.savedMsgs = append(m.savedMsgs, msg)
	return nil
}

func (m *mockConversationRepo) GetMessages(ctx context.Context, conversationID string) ([]*aggregates.Message, error) {
	return m.messages, nil
}

func (m *mockConversationRepo) UpdateConversation(ctx context.Context, conv *aggregates.Conversation) error {
	m.updatedConvs = append(m.updatedConvs, conv)
	return nil
}

func (m *mockConversationRepo) ListByUser(_ context.Context, _ domain.ListConversationsParams) ([]*aggregates.Conversation, int, error) {
	return nil, 0, nil
}

func (m *mockConversationRepo) Delete(_ context.Context, _, _ string) error {
	return nil
}

func (m *mockConversationRepo) CountByUser(_ context.Context, _ string) (int, error) {
	return m.convCount, nil
}

type flushRecorder struct {
	*httptest.ResponseRecorder
}

func (f *flushRecorder) Flush() {
	f.ResponseRecorder.Flush()
}

func newFlushRecorder() *flushRecorder {
	return &flushRecorder{ResponseRecorder: httptest.NewRecorder()}
}

func setupTestOrchestrator(t *testing.T) (*orchestrator.Orchestrator, *mockConversationRepo, *sse.Writer, *flushRecorder) {
	t.Helper()
	repo := &mockConversationRepo{}
	factory := adapters.NewLLMClientFactory()
	orch := orchestrator.New(repo, factory)
	rec := newFlushRecorder()
	writer, err := sse.NewWriter(rec)
	require.NoError(t, err)
	return orch, repo, writer, rec
}

func testConfig(endpoint string) *publishedlanguage.AIConfigInfo {
	return &publishedlanguage.AIConfigInfo{
		Provider:    "openai",
		Endpoint:    endpoint,
		APIKey:      "test-key",
		Model:       "gpt-4",
		MaxTokens:   4096,
		Temperature: 0.3,
	}
}

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

func toolCallResponse(id, name, args string) string {
	return fmt.Sprintf(
		`data: {"choices":[{"delta":{"tool_calls":[{"index":0,"id":"%s","function":{"name":"%s","arguments":"%s"}}]}}]}`+"\n\n"+
			"data: [DONE]\n\n",
		id, name, args,
	)
}

func textResponse(content string) string {
	return fmt.Sprintf(
		`data: {"choices":[{"delta":{"content":"%s"}}]}`+"\n\n"+
			"data: [DONE]\n\n",
		content,
	)
}

func TestOrchestrator_SendMessage_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("data: {\"choices\":[{\"delta\":{\"content\":\"Hello\"}}]}\n\n"))
		w.Write([]byte("data: {\"choices\":[{\"delta\":{\"content\":\" there\"}}]}\n\n"))
		w.Write([]byte("data: [DONE]\n\n"))
	}))
	defer server.Close()

	orch, repo, writer, rec := setupTestOrchestrator(t)

	conv := aggregates.NewConversation("tenant-1", "user-1")
	repo.conversation = conv

	err := orch.SendMessage(context.Background(), writer, orchestrator.SendMessageParams{
		ConversationID: conv.ID(),
		UserID:         "user-1",
		Content:        "Hi there",
		TenantID:       "tenant-1",
		UserRole:       "architect",
		Config:         testConfig(server.URL),
	})
	require.NoError(t, err)

	body := rec.Body.String()
	assert.Contains(t, body, "event: token")
	assert.Contains(t, body, `"content":"Hello"`)
	assert.Contains(t, body, `"content":" there"`)
	assert.Contains(t, body, "event: done")

	require.Len(t, repo.savedMsgs, 2)
	assert.Equal(t, vo.MessageRoleUser, repo.savedMsgs[0].Role())
	assert.Equal(t, vo.MessageRoleAssistant, repo.savedMsgs[1].Role())
	assert.Equal(t, "Hello there", repo.savedMsgs[1].Content())

	require.Len(t, repo.updatedConvs, 1)
}

func TestOrchestrator_SendMessage_WithHistory(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"response\"}}]}\n\n")
		fmt.Fprint(w, "data: [DONE]\n\n")
	}))
	defer server.Close()

	orch, repo, writer, _ := setupTestOrchestrator(t)

	conv := aggregates.ReconstructConversation(aggregates.ReconstructConversationParams{
		ID:       "conv-1",
		TenantID: "tenant-1",
		UserID:   "user-1",
		Title:    "Test",
	})
	repo.conversation = conv
	repo.messages = []*aggregates.Message{
		aggregates.ReconstructMessage(aggregates.ReconstructMessageParams{
			ID:             "msg-1",
			ConversationID: "conv-1",
			Role:           vo.MessageRoleUser,
			Content:        "First question",
		}),
		aggregates.ReconstructMessage(aggregates.ReconstructMessageParams{
			ID:             "msg-2",
			ConversationID: "conv-1",
			Role:           vo.MessageRoleAssistant,
			Content:        "First answer",
		}),
	}

	err := orch.SendMessage(context.Background(), writer, orchestrator.SendMessageParams{
		ConversationID: "conv-1",
		UserID:         "user-1",
		Content:        "Follow-up question",
		TenantID:       "tenant-1",
		UserRole:       "architect",
		Config:         testConfig(server.URL),
	})
	require.NoError(t, err)

	require.Len(t, repo.savedMsgs, 2)
	assert.Equal(t, "response", repo.savedMsgs[1].Content())
}

func TestOrchestrator_SendMessage_LLMError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	orch, repo, writer, _ := setupTestOrchestrator(t)

	conv := aggregates.NewConversation("tenant-1", "user-1")
	repo.conversation = conv

	err := orch.SendMessage(context.Background(), writer, orchestrator.SendMessageParams{
		ConversationID: conv.ID(),
		UserID:         "user-1",
		Content:        "Test",
		TenantID:       "tenant-1",
		UserRole:       "architect",
		Config:         testConfig(server.URL),
	})
	assert.Error(t, err)

	var llmErr *orchestrator.LLMError
	assert.ErrorAs(t, err, &llmErr)
}

func TestOrchestrator_SendMessage_ConversationNotFound(t *testing.T) {
	orch, repo, writer, _ := setupTestOrchestrator(t)
	repo.conversation = nil

	err := orch.SendMessage(context.Background(), writer, orchestrator.SendMessageParams{
		ConversationID: "nonexistent",
		UserID:         "user-1",
		Content:        "Test",
		TenantID:       "tenant-1",
		UserRole:       "architect",
		Config:         testConfig("http://localhost"),
	})
	assert.ErrorIs(t, err, orchestrator.ErrConversationNotFound)
}

func TestOrchestrator_SendMessage_ValidationError(t *testing.T) {
	orch, repo, writer, _ := setupTestOrchestrator(t)

	conv := aggregates.NewConversation("tenant-1", "user-1")
	repo.conversation = conv

	err := orch.SendMessage(context.Background(), writer, orchestrator.SendMessageParams{
		ConversationID: conv.ID(),
		UserID:         "user-1",
		Content:        "",
		TenantID:       "tenant-1",
		UserRole:       "architect",
		Config:         testConfig("http://localhost"),
	})
	assert.Error(t, err)

	var valErr *orchestrator.ValidationError
	assert.ErrorAs(t, err, &valErr)
}

func TestOrchestrator_SendMessage_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		flusher := w.(http.Flusher)
		fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"Hello\"}}]}\n\n")
		flusher.Flush()
		<-r.Context().Done()
	}))
	defer server.Close()

	orch, repo, writer, _ := setupTestOrchestrator(t)

	conv := aggregates.NewConversation("tenant-1", "user-1")
	repo.conversation = conv

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := orch.SendMessage(ctx, writer, orchestrator.SendMessageParams{
		ConversationID: conv.ID(),
		UserID:         "user-1",
		Content:        "Test",
		TenantID:       "tenant-1",
		UserRole:       "architect",
		Config:         testConfig(server.URL),
	})

	assert.Error(t, err)
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

func toolThenTextHandler(toolID, toolName, toolArgs, finalText string) http.Handler {
	var callCount atomic.Int32
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		if callCount.Add(1) == 1 && !hasToolCallMessages(body) {
			w.Write([]byte(toolCallResponse(toolID, toolName, toolArgs)))
			return
		}
		w.Write([]byte(textResponse(finalText)))
	})
}

func TestOrchestrator_SendMessage_WithToolCalls(t *testing.T) {
	registry := newTestRegistry(map[string]tools.ToolResult{
		"list_apps": {Content: `[{"name":"App1"},{"name":"App2"}]`},
	})
	fix, server := setupToolTest(t, registry, toolThenTextHandler("call-1", "list_apps", "{}", "Here are the apps."))

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
	fix, server := setupToolTest(t, registry, toolThenTextHandler("call-1", "list_apps", "{}", "Done"))

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
		w.Write([]byte(toolCallResponse("call-1", "list_apps", "{}")))
	})

	registry := newTestRegistry(map[string]tools.ToolResult{
		"list_apps": {Content: `[]`},
	})
	fix, server := setupToolTest(t, registry, alwaysToolCall)

	err := fix.sendMessage(t, server.URL, "Keep calling tools", allPermissions())

	require.Error(t, err)
	var timeoutErr *orchestrator.TimeoutError
	assert.ErrorAs(t, err, &timeoutErr)
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
	var callCount atomic.Int32
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		n := callCount.Add(1)
		if n <= 4 {
			w.Write([]byte(toolCallResponse(fmt.Sprintf("call-%d", n), "list_apps", "{}")))
			return
		}
		w.Write([]byte(textResponse("Final answer")))
	})

	registry := newTestRegistry(map[string]tools.ToolResult{
		"list_apps": {Content: `[]`},
	})
	fix, server := setupToolTest(t, registry, handler)

	err := fix.sendMessage(t, server.URL, "Call the same tool many times", allPermissions())
	require.NoError(t, err)

	var errorResults int
	for _, r := range fix.writer.toolResults {
		if strings.Contains(r.ResultPreview, "limit exceeded") {
			errorResults++
		}
	}
	assert.GreaterOrEqual(t, errorResults, 1)
}

func TestOrchestrator_SendMessage_ToolCallError(t *testing.T) {
	registry := newTestRegistry(map[string]tools.ToolResult{
		"failing_tool": {Content: "something went wrong", IsError: true},
	})
	fix, server := setupToolTest(t, registry, toolThenTextHandler("call-1", "failing_tool", "{}", "Recovered"))

	err := fix.sendMessage(t, server.URL, "Do something", allPermissions())
	require.NoError(t, err)

	assert.Equal(t, []string{"Recovered"}, fix.writer.tokens)
	require.Len(t, fix.writer.toolResults, 1)
}
