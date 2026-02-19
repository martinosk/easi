package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"easi/backend/internal/archassistant/application/orchestrator"
	"easi/backend/internal/archassistant/domain"
	"easi/backend/internal/archassistant/domain/aggregates"
	vo "easi/backend/internal/archassistant/domain/valueobjects"
	"easi/backend/internal/archassistant/infrastructure/adapters"
	assistantAPI "easi/backend/internal/archassistant/infrastructure/api"
	"easi/backend/internal/archassistant/infrastructure/ratelimit"
	"easi/backend/internal/archassistant/publishedlanguage"
	sharedctx "easi/backend/internal/shared/context"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockConversationRepo struct {
	createdConvs []*aggregates.Conversation
	conversation *aggregates.Conversation
	messages     []*aggregates.Message
	savedMsgs    []*aggregates.Message
	updatedConvs []*aggregates.Conversation
	createErr    error
	getErr       error
}

func (m *mockConversationRepo) Create(ctx context.Context, conv *aggregates.Conversation) error {
	m.createdConvs = append(m.createdConvs, conv)
	return m.createErr
}

func (m *mockConversationRepo) GetByIDAndUser(ctx context.Context, id, userID string) (*aggregates.Conversation, error) {
	return m.conversation, m.getErr
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
	return 0, nil
}

type mockConfigProvider struct {
	config *publishedlanguage.AIConfigInfo
	err    error
}

func (m *mockConfigProvider) GetDecryptedConfig(ctx context.Context) (*publishedlanguage.AIConfigInfo, error) {
	return m.config, m.err
}

func (m *mockConfigProvider) IsConfigured(ctx context.Context) (bool, error) {
	return m.config != nil, nil
}

func defaultAIConfig(endpoint string) *publishedlanguage.AIConfigInfo {
	return &publishedlanguage.AIConfigInfo{
		Provider:  "openai",
		Endpoint:  endpoint,
		APIKey:    "key",
		Model:     "gpt-4",
		MaxTokens: 4096,
	}
}

func withActorAndTenant(r *http.Request) *http.Request {
	ctx := r.Context()
	actor := sharedctx.NewActor("user-1", "user@example.com", sharedctx.RoleArchitect)
	ctx = sharedctx.WithActor(ctx, actor)
	tenantID, _ := sharedvo.NewTenantID("tenant-1")
	ctx = sharedctx.WithTenant(ctx, tenantID)
	return r.WithContext(ctx)
}

func newTestHandlers(repo *mockConversationRepo, cp *mockConfigProvider, limiter *ratelimit.Limiter) *assistantAPI.ConversationHandlers {
	factory := adapters.NewLLMClientFactory()
	orch := orchestrator.New(repo, factory)
	return assistantAPI.NewConversationHandlers(assistantAPI.ConversationHandlersDeps{
		ConfigProvider: cp,
		RateLimiter:    limiter,
		Orchestrator:   orch,
		ConvRepo:       repo,
	})
}

func newSendMessageRequest(t *testing.T, convID string, body interface{}) *http.Request {
	t.Helper()
	b, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/assistant/conversations/"+convID+"/messages", bytes.NewReader(b))
	req = withActorAndTenant(req)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", convID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	return req
}

func TestCreateConversation_Success(t *testing.T) {
	repo := &mockConversationRepo{}
	handlers := newTestHandlers(repo, &mockConfigProvider{}, ratelimit.NewLimiter())

	req := httptest.NewRequest("POST", "/assistant/conversations", nil)
	req = withActorAndTenant(req)
	rec := httptest.NewRecorder()

	handlers.CreateConversation(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.NotEmpty(t, rec.Header().Get("Location"))

	var resp assistantAPI.CreateConversationResponse
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.ID)
	assert.Equal(t, "New conversation", resp.Title)
	assert.NotNil(t, resp.Links["self"])
	assert.NotNil(t, resp.Links["messages"])
	assert.NotNil(t, resp.Links["delete"])

	require.Len(t, repo.createdConvs, 1)
	assert.Equal(t, "tenant-1", repo.createdConvs[0].TenantID())
	assert.Equal(t, "user-1", repo.createdConvs[0].UserID())
}

func TestCreateConversation_NoActor(t *testing.T) {
	handlers := newTestHandlers(&mockConversationRepo{}, &mockConfigProvider{}, ratelimit.NewLimiter())

	req := httptest.NewRequest("POST", "/assistant/conversations", nil)
	rec := httptest.NewRecorder()

	handlers.CreateConversation(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestSendMessage_ErrorCases(t *testing.T) {
	tests := []struct {
		name           string
		conversation   *aggregates.Conversation
		configProvider *mockConfigProvider
		content        string
		convID         string
		preLimiter     func(*ratelimit.Limiter)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "not configured",
			conversation:   aggregates.NewConversation("tenant-1", "user-1"),
			configProvider: &mockConfigProvider{err: publishedlanguage.ErrNotConfigured},
			content:        "Hello",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "not configured",
		},
		{
			name:           "conversation not found",
			conversation:   nil,
			configProvider: &mockConfigProvider{config: defaultAIConfig("http://localhost")},
			content:        "Hello",
			convID:         "00000000-0000-0000-0000-000000000000",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "rate limit concurrent stream",
			conversation:   aggregates.NewConversation("tenant-1", "user-1"),
			configProvider: &mockConfigProvider{config: defaultAIConfig("http://localhost")},
			content:        "Hello",
			preLimiter:     func(l *ratelimit.Limiter) { require.NoError(t, l.AcquireStream("user-1")) },
			expectedStatus: http.StatusTooManyRequests,
		},
		{
			name:           "invalid UUID format",
			conversation:   aggregates.NewConversation("tenant-1", "user-1"),
			configProvider: &mockConfigProvider{config: defaultAIConfig("http://localhost")},
			content:        "Hello",
			convID:         "not-a-uuid",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid conversation ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockConversationRepo{conversation: tt.conversation}
			limiter := ratelimit.NewLimiter()
			if tt.preLimiter != nil {
				tt.preLimiter(limiter)
			}

			handlers := newTestHandlers(repo, tt.configProvider, limiter)

			convID := tt.convID
			if convID == "" && tt.conversation != nil {
				convID = tt.conversation.ID()
			}
			if convID == "" {
				convID = "00000000-0000-0000-0000-000000000000"
			}

			req := newSendMessageRequest(t, convID, map[string]string{"content": tt.content})
			rec := httptest.NewRecorder()
			handlers.SendMessage(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.expectedBody != "" {
				assert.Contains(t, rec.Body.String(), tt.expectedBody)
			}
		})
	}
}

func TestSendMessage_SSEStream(t *testing.T) {
	llmServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("data: {\"choices\":[{\"delta\":{\"content\":\"Hi!\"}}]}\n\n"))
		w.Write([]byte("data: [DONE]\n\n"))
	}))
	defer llmServer.Close()

	conv := aggregates.NewConversation("tenant-1", "user-1")
	repo := &mockConversationRepo{conversation: conv}
	cp := &mockConfigProvider{config: defaultAIConfig(llmServer.URL)}
	handlers := newTestHandlers(repo, cp, ratelimit.NewLimiter())

	req := newSendMessageRequest(t, conv.ID(), map[string]string{"content": "Hello"})
	rec := httptest.NewRecorder()
	handlers.SendMessage(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "text/event-stream", rec.Header().Get("Content-Type"))
	assert.Contains(t, rec.Body.String(), "event: ping")
	assert.Contains(t, rec.Body.String(), "event: token")
	assert.Contains(t, rec.Body.String(), "event: done")

	require.Len(t, repo.savedMsgs, 2)
	assert.Equal(t, vo.MessageRoleUser, repo.savedMsgs[0].Role())
	assert.Equal(t, vo.MessageRoleAssistant, repo.savedMsgs[1].Role())
}

func TestSendMessage_InvalidBody(t *testing.T) {
	handlers := newTestHandlers(&mockConversationRepo{}, &mockConfigProvider{}, ratelimit.NewLimiter())

	req := httptest.NewRequest("POST", "/assistant/conversations/00000000-0000-0000-0000-000000000000/messages", bytes.NewReader([]byte("not json")))
	req = withActorAndTenant(req)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "00000000-0000-0000-0000-000000000000")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rec := httptest.NewRecorder()
	handlers.SendMessage(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
