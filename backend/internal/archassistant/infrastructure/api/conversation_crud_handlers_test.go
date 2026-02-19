package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"easi/backend/internal/archassistant/domain"
	"easi/backend/internal/archassistant/domain/aggregates"
	vo "easi/backend/internal/archassistant/domain/valueobjects"
	assistantAPI "easi/backend/internal/archassistant/infrastructure/api"
	sharedctx "easi/backend/internal/shared/context"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
	"easi/backend/internal/shared/types"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type crudMockRepo struct {
	conversations []*aggregates.Conversation
	conversation  *aggregates.Conversation
	messages      []*aggregates.Message
	total         int
	count         int
	createErr     error
	getErr        error
	listErr       error
	deleteErr     error
	countErr      error

	createdConvs []*aggregates.Conversation
	deletedID    string
	deletedUser  string
}

func (m *crudMockRepo) Create(_ context.Context, conv *aggregates.Conversation) error {
	m.createdConvs = append(m.createdConvs, conv)
	return m.createErr
}

func (m *crudMockRepo) GetByIDAndUser(_ context.Context, id, userID string) (*aggregates.Conversation, error) {
	return m.conversation, m.getErr
}

func (m *crudMockRepo) SaveMessage(_ context.Context, _ *aggregates.Message) error {
	return nil
}

func (m *crudMockRepo) GetMessages(_ context.Context, _ string) ([]*aggregates.Message, error) {
	return m.messages, nil
}

func (m *crudMockRepo) UpdateConversation(_ context.Context, _ *aggregates.Conversation) error {
	return nil
}

func (m *crudMockRepo) ListByUser(_ context.Context, _ domain.ListConversationsParams) ([]*aggregates.Conversation, int, error) {
	return m.conversations, m.total, m.listErr
}

func (m *crudMockRepo) Delete(_ context.Context, id, userID string) error {
	m.deletedID = id
	m.deletedUser = userID
	return m.deleteErr
}

func (m *crudMockRepo) CountByUser(_ context.Context, _ string) (int, error) {
	return m.count, m.countErr
}

func newCRUDHandlers(repo domain.ConversationRepository) *assistantAPI.ConversationHandlers {
	return assistantAPI.NewCRUDHandlers(repo)
}

func withCRUDActorAndTenant(r *http.Request) *http.Request {
	ctx := r.Context()
	actor := sharedctx.NewActor("user-1", "user@example.com", sharedctx.RoleArchitect)
	ctx = sharedctx.WithActor(ctx, actor)
	tenantID, _ := sharedvo.NewTenantID("tenant-1")
	ctx = sharedctx.WithTenant(ctx, tenantID)
	return r.WithContext(ctx)
}

func withRouteParam(r *http.Request, name, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(name, value)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func makeConversation(p aggregates.ReconstructConversationParams) *aggregates.Conversation {
	return aggregates.ReconstructConversation(p)
}

func makeMessage(p aggregates.ReconstructMessageParams) *aggregates.Message {
	return aggregates.ReconstructMessage(p)
}

func TestListConversations_ReturnsUserConversations(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	repo := &crudMockRepo{
		conversations: []*aggregates.Conversation{
			makeConversation(aggregates.ReconstructConversationParams{ID: "conv-1", TenantID: "tenant-1", UserID: "user-1", Title: "First chat", CreatedAt: now, LastMessageAt: now}),
			makeConversation(aggregates.ReconstructConversationParams{ID: "conv-2", TenantID: "tenant-1", UserID: "user-1", Title: "Second chat", CreatedAt: now.Add(-time.Hour), LastMessageAt: now.Add(-30 * time.Minute)}),
		},
		total: 2,
	}
	handlers := newCRUDHandlers(repo)

	req := httptest.NewRequest("GET", "/assistant/conversations?limit=20&offset=0", nil)
	req = withCRUDActorAndTenant(req)
	rec := httptest.NewRecorder()

	handlers.ListConversations(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp struct {
		Data []struct {
			ID            string      `json:"id"`
			Title         string      `json:"title"`
			CreatedAt     time.Time   `json:"createdAt"`
			LastMessageAt time.Time   `json:"lastMessageAt"`
			Links         types.Links `json:"_links"`
		} `json:"data"`
		Links types.Links `json:"_links"`
		Meta  struct {
			Total *int `json:"total"`
		} `json:"meta"`
	}
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)

	require.Len(t, resp.Data, 2)
	assert.Equal(t, "conv-1", resp.Data[0].ID)
	assert.Equal(t, "First chat", resp.Data[0].Title)
	assert.NotNil(t, resp.Data[0].Links["self"])
	assert.NotNil(t, resp.Data[0].Links["delete"])

	require.NotNil(t, resp.Meta.Total)
	assert.Equal(t, 2, *resp.Meta.Total)

	assert.NotNil(t, resp.Links["self"])
	assert.NotNil(t, resp.Links["create"])
}

func TestListConversations_EmptyList(t *testing.T) {
	repo := &crudMockRepo{
		conversations: nil,
		total:         0,
	}
	handlers := newCRUDHandlers(repo)

	req := httptest.NewRequest("GET", "/assistant/conversations", nil)
	req = withCRUDActorAndTenant(req)
	rec := httptest.NewRecorder()

	handlers.ListConversations(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp struct {
		Data []interface{} `json:"data"`
	}
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Empty(t, resp.Data)
}

func TestListConversations_NoActor(t *testing.T) {
	handlers := newCRUDHandlers(&crudMockRepo{})

	req := httptest.NewRequest("GET", "/assistant/conversations", nil)
	rec := httptest.NewRecorder()

	handlers.ListConversations(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestGetConversation_ReturnsConversationWithMessages(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	convID := "00000000-0000-0000-0000-000000000001"
	conv := makeConversation(aggregates.ReconstructConversationParams{ID: convID, TenantID: "tenant-1", UserID: "user-1", Title: "My chat", CreatedAt: now, LastMessageAt: now})
	messages := []*aggregates.Message{
		makeMessage(aggregates.ReconstructMessageParams{ID: "msg-1", ConversationID: convID, Role: vo.MessageRoleUser, Content: "Hello", CreatedAt: now}),
		makeMessage(aggregates.ReconstructMessageParams{ID: "msg-2", ConversationID: convID, Role: vo.MessageRoleAssistant, Content: "Hi there!", CreatedAt: now.Add(time.Second)}),
	}

	repo := &crudMockRepo{
		conversation: conv,
		messages:     messages,
	}
	handlers := newCRUDHandlers(repo)

	req := httptest.NewRequest("GET", "/assistant/conversations/"+convID, nil)
	req = withCRUDActorAndTenant(req)
	req = withRouteParam(req, "id", convID)
	rec := httptest.NewRecorder()

	handlers.GetConversation(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp struct {
		ID            string      `json:"id"`
		Title         string      `json:"title"`
		CreatedAt     time.Time   `json:"createdAt"`
		LastMessageAt time.Time   `json:"lastMessageAt"`
		Messages      []struct {
			ID        string    `json:"id"`
			Role      string    `json:"role"`
			Content   string    `json:"content"`
			CreatedAt time.Time `json:"createdAt"`
		} `json:"messages"`
		Links types.Links `json:"_links"`
	}
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, convID, resp.ID)
	assert.Equal(t, "My chat", resp.Title)
	require.Len(t, resp.Messages, 2)
	assert.Equal(t, "msg-1", resp.Messages[0].ID)
	assert.Equal(t, "user", resp.Messages[0].Role)
	assert.Equal(t, "Hello", resp.Messages[0].Content)
	assert.Equal(t, "msg-2", resp.Messages[1].ID)
	assert.Equal(t, "assistant", resp.Messages[1].Role)
	assert.NotNil(t, resp.Links["self"])
	assert.NotNil(t, resp.Links["messages"])
	assert.NotNil(t, resp.Links["delete"])
}

func TestGetConversation_NotFound(t *testing.T) {
	repo := &crudMockRepo{conversation: nil}
	handlers := newCRUDHandlers(repo)

	convID := "00000000-0000-0000-0000-000000000001"
	req := httptest.NewRequest("GET", "/assistant/conversations/"+convID, nil)
	req = withCRUDActorAndTenant(req)
	req = withRouteParam(req, "id", convID)
	rec := httptest.NewRecorder()

	handlers.GetConversation(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestGetConversation_InvalidUUID(t *testing.T) {
	handlers := newCRUDHandlers(&crudMockRepo{})

	req := httptest.NewRequest("GET", "/assistant/conversations/not-a-uuid", nil)
	req = withCRUDActorAndTenant(req)
	req = withRouteParam(req, "id", "not-a-uuid")
	rec := httptest.NewRecorder()

	handlers.GetConversation(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestDeleteConversation_Success(t *testing.T) {
	convID := "00000000-0000-0000-0000-000000000001"
	repo := &crudMockRepo{}
	handlers := newCRUDHandlers(repo)

	req := httptest.NewRequest("DELETE", "/assistant/conversations/"+convID, nil)
	req = withCRUDActorAndTenant(req)
	req = withRouteParam(req, "id", convID)
	rec := httptest.NewRecorder()

	handlers.DeleteConversation(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Equal(t, convID, repo.deletedID)
	assert.Equal(t, "user-1", repo.deletedUser)
}

func TestDeleteConversation_NotFound(t *testing.T) {
	convID := "00000000-0000-0000-0000-000000000001"
	repo := &crudMockRepo{deleteErr: assistantAPI.ErrConversationNotFound}
	handlers := newCRUDHandlers(repo)

	req := httptest.NewRequest("DELETE", "/assistant/conversations/"+convID, nil)
	req = withCRUDActorAndTenant(req)
	req = withRouteParam(req, "id", convID)
	rec := httptest.NewRecorder()

	handlers.DeleteConversation(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestDeleteConversation_InvalidUUID(t *testing.T) {
	handlers := newCRUDHandlers(&crudMockRepo{})

	req := httptest.NewRequest("DELETE", "/assistant/conversations/not-a-uuid", nil)
	req = withCRUDActorAndTenant(req)
	req = withRouteParam(req, "id", "not-a-uuid")
	rec := httptest.NewRecorder()

	handlers.DeleteConversation(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateConversation_RejectsWhenMaxReached(t *testing.T) {
	repo := &crudMockRepo{count: 100}
	handlers := newCRUDHandlers(repo)

	req := httptest.NewRequest("POST", "/assistant/conversations", nil)
	req = withCRUDActorAndTenant(req)
	rec := httptest.NewRecorder()

	handlers.CreateConversationWithLimit(rec, req)

	assert.Equal(t, http.StatusConflict, rec.Code)
	assert.Contains(t, rec.Body.String(), "maximum")
}

func TestCreateConversation_AllowsWhenUnderMax(t *testing.T) {
	repo := &crudMockRepo{count: 99}
	handlers := newCRUDHandlers(repo)

	req := httptest.NewRequest("POST", "/assistant/conversations", nil)
	req = withCRUDActorAndTenant(req)
	rec := httptest.NewRecorder()

	handlers.CreateConversationWithLimit(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	require.Len(t, repo.createdConvs, 1)
}
