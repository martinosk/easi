package orchestrator_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"easi/backend/internal/archassistant/application/orchestrator"
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

	require.Error(t, err)
	var llmErr *orchestrator.LLMError
	var timeoutErr *orchestrator.TimeoutError
	isTyped := assert.ErrorAs(t, err, &llmErr) || assert.ErrorAs(t, err, &timeoutErr)
	assert.True(t, isTyped, "context cancellation should be wrapped as LLMError or TimeoutError, got: %T", err)
}
