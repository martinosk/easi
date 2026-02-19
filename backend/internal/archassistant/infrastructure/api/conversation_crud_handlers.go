package api

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"easi/backend/internal/archassistant/domain"
	"easi/backend/internal/archassistant/domain/aggregates"
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/types"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

const maxConversationsPerUser = 100

var ErrConversationNotFound = domain.ErrConversationNotFound

func NewCRUDHandlers(repo domain.ConversationRepository) *ConversationHandlers {
	return &ConversationHandlers{
		convRepo: repo,
		links:    sharedAPI.NewLinkBuilder("/assistant/conversations"),
	}
}

type ConversationListItem struct {
	ID            string      `json:"id"`
	Title         string      `json:"title"`
	CreatedAt     time.Time   `json:"createdAt"`
	LastMessageAt time.Time   `json:"lastMessageAt"`
	Links         types.Links `json:"_links"`
}

type ConversationDetailResponse struct {
	ID            string            `json:"id"`
	Title         string            `json:"title"`
	CreatedAt     time.Time         `json:"createdAt"`
	LastMessageAt time.Time         `json:"lastMessageAt"`
	Messages      []MessageResponse `json:"messages"`
	Links         types.Links       `json:"_links"`
}

type MessageResponse struct {
	ID        string    `json:"id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}

// ListConversations godoc
// @Summary List user conversations
// @Description Returns a paginated list of the authenticated user's conversations.
// @Tags assistant
// @Produce json
// @Param limit query int false "Page size (default 20, max 50)"
// @Param offset query int false "Offset (default 0)"
// @Success 200 {object} sharedAPI.CollectionResponse
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /assistant/conversations [get]
func (h *ConversationHandlers) ListConversations(w http.ResponseWriter, r *http.Request) {
	actor, ok := sharedctx.GetActor(r.Context())
	if !ok {
		sharedAPI.RespondError(w, http.StatusUnauthorized, nil, "Unauthorized")
		return
	}

	limit, offset := parseListParams(r)

	conversations, total, err := h.convRepo.ListByUser(r.Context(), domain.ListConversationsParams{
		UserID: actor.ID,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to list conversations")
		return
	}

	items := make([]ConversationListItem, 0, len(conversations))
	for _, conv := range conversations {
		items = append(items, toConversationListItem(conv, h.links))
	}

	collectionLinks := sharedAPI.NewResourceLinks().
		Self("/assistant/conversations").
		Add("create", h.links.Collection(), "POST").
		Build()

	sharedAPI.RespondCollectionWithTotal(w, sharedAPI.CollectionWithTotalParams{
		StatusCode: http.StatusOK,
		Data:       items,
		Total:      total,
		Links:      collectionLinks,
	})
}

// GetConversation godoc
// @Summary Get a conversation with messages
// @Description Returns a conversation and its message history.
// @Tags assistant
// @Produce json
// @Param id path string true "Conversation ID"
// @Success 200 {object} ConversationDetailResponse
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /assistant/conversations/{id} [get]
func (h *ConversationHandlers) GetConversation(w http.ResponseWriter, r *http.Request) {
	actor, ok := sharedctx.GetActor(r.Context())
	if !ok {
		sharedAPI.RespondError(w, http.StatusUnauthorized, nil, "Unauthorized")
		return
	}

	convID := chi.URLParam(r, "id")
	if _, err := uuid.Parse(convID); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, nil, "Invalid conversation ID format")
		return
	}

	conv, err := h.convRepo.GetByIDAndUser(r.Context(), convID, actor.ID)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to get conversation")
		return
	}
	if conv == nil {
		sharedAPI.RespondError(w, http.StatusNotFound, nil, "Conversation not found")
		return
	}

	messages, err := h.convRepo.GetMessages(r.Context(), convID)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to load messages")
		return
	}

	resp := toConversationDetail(conv, messages, h.links)
	sharedAPI.RespondJSON(w, http.StatusOK, resp)
}

// DeleteConversation godoc
// @Summary Delete a conversation
// @Description Deletes a conversation and all its messages.
// @Tags assistant
// @Param id path string true "Conversation ID"
// @Success 204
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /assistant/conversations/{id} [delete]
func (h *ConversationHandlers) DeleteConversation(w http.ResponseWriter, r *http.Request) {
	actor, ok := sharedctx.GetActor(r.Context())
	if !ok {
		sharedAPI.RespondError(w, http.StatusUnauthorized, nil, "Unauthorized")
		return
	}

	convID := chi.URLParam(r, "id")
	if _, err := uuid.Parse(convID); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, nil, "Invalid conversation ID format")
		return
	}

	err := h.convRepo.Delete(r.Context(), convID, actor.ID)
	if err != nil {
		if errors.Is(err, domain.ErrConversationNotFound) {
			sharedAPI.RespondError(w, http.StatusNotFound, nil, "Conversation not found")
			return
		}
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to delete conversation")
		return
	}

	sharedAPI.RespondDeleted(w)
}

// CreateConversationWithLimit godoc
// @Summary Create a new conversation with limit enforcement
// @Description Creates a new assistant conversation, enforcing max 100 per user.
// @Tags assistant
// @Produce json
// @Success 201 {object} CreateConversationResponse
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 409 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /assistant/conversations [post]
func (h *ConversationHandlers) CreateConversationWithLimit(w http.ResponseWriter, r *http.Request) {
	actor, ok := sharedctx.GetActor(r.Context())
	if !ok {
		sharedAPI.RespondError(w, http.StatusUnauthorized, nil, "Unauthorized")
		return
	}

	tenantID, err := sharedctx.GetTenant(r.Context())
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to get tenant")
		return
	}

	count, err := h.convRepo.CountByUser(r.Context(), actor.ID)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to check conversation count")
		return
	}
	if count >= maxConversationsPerUser {
		sharedAPI.RespondError(w, http.StatusConflict, nil, "You have reached the maximum number of conversations (100). Please delete some before creating new ones.")
		return
	}

	conv := aggregates.NewConversation(tenantID.Value(), actor.ID)

	if err := h.convRepo.Create(r.Context(), conv); err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to create conversation")
		return
	}

	resp := CreateConversationResponse{
		ID:        conv.ID(),
		Title:     conv.Title(),
		CreatedAt: conv.CreatedAt(),
		Links:     h.conversationLinks(conv.ID()),
	}

	location := h.links.Self(sharedAPI.ResourceID(conv.ID()))
	sharedAPI.RespondCreated(w, location, resp)
}

func parseListParams(r *http.Request) (int, int) {
	return clamp(queryInt(r, "limit", 20), 1, 50),
		clamp(queryInt(r, "offset", 0), 0, 10000)
}

func queryInt(r *http.Request, name string, fallback int) int {
	raw := r.URL.Query().Get(name)
	if raw == "" {
		return fallback
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return v
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func toConversationListItem(conv *aggregates.Conversation, links *sharedAPI.LinkBuilder) ConversationListItem {
	id := sharedAPI.ResourceID(conv.ID())
	return ConversationListItem{
		ID:            conv.ID(),
		Title:         conv.Title(),
		CreatedAt:     conv.CreatedAt(),
		LastMessageAt: conv.LastMessageAt(),
		Links: sharedAPI.NewResourceLinks().
			SelfWithID("/assistant/conversations", id).
			Add("delete", links.Delete(id), "DELETE").
			Build(),
	}
}

func toConversationDetail(conv *aggregates.Conversation, messages []*aggregates.Message, links *sharedAPI.LinkBuilder) ConversationDetailResponse {
	id := sharedAPI.ResourceID(conv.ID())
	msgResponses := make([]MessageResponse, 0, len(messages))
	for _, msg := range messages {
		msgResponses = append(msgResponses, MessageResponse{
			ID:        msg.ID(),
			Role:      msg.Role().String(),
			Content:   msg.Content(),
			CreatedAt: msg.CreatedAt(),
		})
	}

	return ConversationDetailResponse{
		ID:            conv.ID(),
		Title:         conv.Title(),
		CreatedAt:     conv.CreatedAt(),
		LastMessageAt: conv.LastMessageAt(),
		Messages:      msgResponses,
		Links: sharedAPI.NewResourceLinks().
			SelfWithID("/assistant/conversations", id).
			Add("messages", links.SubResource(id, "/messages"), "POST").
			Add("delete", links.Delete(id), "DELETE").
			Build(),
	}
}
