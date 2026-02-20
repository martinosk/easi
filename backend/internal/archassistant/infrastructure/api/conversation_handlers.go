package api

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"easi/backend/internal/archassistant/application/orchestrator"
	"easi/backend/internal/archassistant/application/tools"
	"easi/backend/internal/archassistant/domain"
	"easi/backend/internal/archassistant/domain/aggregates"
	"easi/backend/internal/archassistant/infrastructure/agenthttp"
	"easi/backend/internal/archassistant/infrastructure/ratelimit"
	"easi/backend/internal/archassistant/infrastructure/sse"
	"easi/backend/internal/archassistant/infrastructure/toolimpls"
	"easi/backend/internal/archassistant/publishedlanguage"
	"easi/backend/internal/shared/agenttoken"
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/types"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

const maxRequestBodySize = 8 * 1024

type ConversationHandlersDeps struct {
	ConfigProvider  publishedlanguage.AIConfigProvider
	RateLimiter     *ratelimit.Limiter
	Orchestrator    *orchestrator.Orchestrator
	ConvRepo        domain.ConversationRepository
	LoopbackBaseURL string
}

type ConversationHandlers struct {
	configProvider  publishedlanguage.AIConfigProvider
	rateLimiter     *ratelimit.Limiter
	orchestrator    *orchestrator.Orchestrator
	convRepo        domain.ConversationRepository
	links           *sharedAPI.LinkBuilder
	loopbackBaseURL string
}

func NewConversationHandlers(deps ConversationHandlersDeps) *ConversationHandlers {
	return &ConversationHandlers{
		configProvider:  deps.ConfigProvider,
		rateLimiter:     deps.RateLimiter,
		orchestrator:    deps.Orchestrator,
		convRepo:        deps.ConvRepo,
		links:           sharedAPI.NewLinkBuilder("/assistant/conversations"),
		loopbackBaseURL: deps.LoopbackBaseURL,
	}
}

type CreateConversationResponse struct {
	ID        string     `json:"id"`
	Title     string     `json:"title"`
	CreatedAt time.Time  `json:"createdAt"`
	Links     types.Links `json:"_links"`
}

// CreateConversation godoc
// @Summary Create a new conversation
// @Description Creates a new assistant conversation for the authenticated user.
// @Tags assistant
// @Produce json
// @Success 201 {object} CreateConversationResponse
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /assistant/conversations [post]
func (h *ConversationHandlers) CreateConversation(w http.ResponseWriter, r *http.Request) {
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

	conv := aggregates.NewConversation(tenantID.Value(), actor.ID)

	if err := h.orchestrator.CreateConversation(r.Context(), conv); err != nil {
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

type SendMessageRequest struct {
	Content              string `json:"content"`
	AllowWriteOperations bool   `json:"allowWriteOperations"`
}

// SendMessage godoc
// @Summary Send a message to the assistant
// @Description Sends a user message and streams the assistant response as Server-Sent Events.
// @Tags assistant
// @Accept json
// @Produce text/event-stream
// @Param id path string true "Conversation ID"
// @Param request body SendMessageRequest true "Message content"
// @Success 200 {string} string "SSE stream of token/ping/done/error events"
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 429 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /assistant/conversations/{id}/messages [post]
func (h *ConversationHandlers) SendMessage(w http.ResponseWriter, r *http.Request) {
	input, herr := h.parseSendMessageInput(r)
	if herr != nil {
		herr.respond(w)
		return
	}

	if err := h.rateLimiter.AcquireStream(input.actor.ID); err != nil {
		respondError(w, http.StatusTooManyRequests, err.Error())
		return
	}
	defer h.rateLimiter.ReleaseStream(input.actor.ID)

	if err := h.rateLimiter.AllowMessage(input.actor.ID, input.tenantID); err != nil {
		respondError(w, http.StatusTooManyRequests, err.Error())
		return
	}

	config, err := h.configProvider.GetDecryptedConfig(r.Context())
	if err != nil {
		if errors.Is(err, publishedlanguage.ErrNotConfigured) {
			respondError(w, http.StatusBadRequest, "AI assistant is not configured. Ask an admin to configure it in Settings.")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to load AI configuration")
		return
	}

	h.streamAssistantResponse(w, r.Context(), input, config)
}

type parsedInput struct {
	actor                sharedctx.Actor
	tenantID             string
	convID               string
	content              string
	allowWriteOperations bool
}

func (h *ConversationHandlers) parseSendMessageInput(r *http.Request) (*parsedInput, *handlerError) {
	actor, ok := sharedctx.GetActor(r.Context())
	if !ok {
		return nil, &handlerError{http.StatusUnauthorized, "Unauthorized"}
	}

	tenantID, err := sharedctx.GetTenant(r.Context())
	if err != nil {
		return nil, &handlerError{http.StatusInternalServerError, "Failed to get tenant"}
	}

	convID := chi.URLParam(r, "id")
	if convID == "" {
		return nil, &handlerError{http.StatusBadRequest, "Missing conversation ID"}
	}
	if _, err := uuid.Parse(convID); err != nil {
		return nil, &handlerError{http.StatusBadRequest, "Invalid conversation ID format"}
	}

	r.Body = http.MaxBytesReader(nil, r.Body, maxRequestBodySize)
	var req SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, &handlerError{http.StatusBadRequest, "Invalid request body"}
	}

	return &parsedInput{
		actor:                actor,
		tenantID:             tenantID.Value(),
		convID:               convID,
		content:              req.Content,
		allowWriteOperations: req.AllowWriteOperations,
	}, nil
}

func (h *ConversationHandlers) streamAssistantResponse(w http.ResponseWriter, ctx context.Context, input *parsedInput, config *publishedlanguage.AIConfigInfo) {
	sse.SetSSEHeaders(w)
	w.WriteHeader(http.StatusOK)

	sseWriter, err := sse.NewWriter(w)
	if err != nil {
		log.Printf("failed to create SSE writer: %v", err)
		return
	}

	_ = sseWriter.WritePing()

	ctx, cancelPing := context.WithCancel(ctx)
	pingDone := startPingLoop(ctx, sseWriter)
	defer func() {
		cancelPing()
		<-pingDone
	}()

	registry, permissions := h.buildToolContext(input)
	if registry != nil {
		log.Printf("[archassistant] tools enabled: %d tools registered", len(registry.ToolNames()))
	} else {
		log.Printf("[archassistant] tools NOT available (registry=nil)")
	}

	streamErr := h.orchestrator.SendMessage(ctx, sseWriter, orchestrator.SendMessageParams{
		ConversationID:       input.convID,
		UserID:               input.actor.ID,
		Content:              input.content,
		TenantID:             input.tenantID,
		UserRole:             string(input.actor.Role),
		AllowWriteOperations: input.allowWriteOperations,
		SystemPromptOverride: nil,
		Config:               config,
		Permissions:          permissions,
		ToolRegistry:         registry,
	})

	if streamErr != nil && ctx.Err() == nil {
		code := classifyOrchestratorError(streamErr)
		_ = sseWriter.WriteError(code, sanitizeErrorMessage(streamErr))
	}
}

func (h *ConversationHandlers) buildToolContext(input *parsedInput) (*tools.Registry, tools.PermissionChecker) {
	if h.loopbackBaseURL == "" {
		return nil, nil
	}

	token, err := agenttoken.Mint(input.actor.ID, input.tenantID, agenttoken.DefaultTTL)
	if err != nil {
		log.Printf("failed to mint agent token: %v", err)
		return nil, nil
	}

	client := agenthttp.NewClient(h.loopbackBaseURL, token)
	registry := tools.NewRegistry()
	toolimpls.RegisterQueryTools(registry, client)
	toolimpls.RegisterMutationTools(registry, client)

	return registry, &actorPermissions{actor: input.actor}
}

type actorPermissions struct {
	actor sharedctx.Actor
}

func (a *actorPermissions) HasPermission(permission string) bool {
	return a.actor.HasPermission(permission)
}

func startPingLoop(ctx context.Context, writer *sse.Writer) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		defer close(done)
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := writer.WritePing(); err != nil {
					return
				}
			}
		}
	}()
	return done
}

func classifyOrchestratorError(err error) string {
	var iterErr *orchestrator.IterationLimitError
	if errors.As(err, &iterErr) {
		return "iteration_limit"
	}
	var timeoutErr *orchestrator.TimeoutError
	if errors.As(err, &timeoutErr) {
		return "timeout"
	}
	var valErr *orchestrator.ValidationError
	if errors.As(err, &valErr) {
		return "validation_error"
	}
	return "llm_error"
}

func sanitizeErrorMessage(err error) string {
	var iterErr *orchestrator.IterationLimitError
	if errors.As(err, &iterErr) {
		return "The assistant reached its processing limit for this message. Try breaking your request into smaller steps."
	}
	var timeoutErr *orchestrator.TimeoutError
	if errors.As(err, &timeoutErr) {
		return "The request timed out. Please try again."
	}
	var valErr *orchestrator.ValidationError
	if errors.As(err, &valErr) {
		return "Invalid request: " + valErr.Error()
	}
	return "The AI service returned an error. Please try again later."
}

func (h *ConversationHandlers) conversationLinks(convID string) types.Links {
	id := sharedAPI.ResourceID(convID)
	return sharedAPI.NewResourceLinks().
		SelfWithID("/assistant/conversations", id).
		Add("messages", h.links.SubResource(id, "/messages"), "POST").
		Add("delete", h.links.Delete(id), "DELETE").
		Build()
}

type handlerError struct {
	status  int
	message string
}

func (e *handlerError) respond(w http.ResponseWriter) {
	sharedAPI.RespondError(w, e.status, nil, e.message)
}

func respondError(w http.ResponseWriter, status int, message string) {
	sharedAPI.RespondError(w, status, nil, message)
}
