package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"easi/backend/internal/auth/application/commands"
	"easi/backend/internal/auth/application/readmodels"
	"easi/backend/internal/auth/domain/aggregates"
	"easi/backend/internal/auth/domain/valueobjects"
	"easi/backend/internal/auth/infrastructure/repositories"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/cqrs"

	"github.com/go-chi/chi/v5"
)

type InvitationHandlers struct {
	commandBus       cqrs.CommandBus
	readModel        *readmodels.InvitationReadModel
	paginationHelper *sharedAPI.PaginationHelper
}

func NewInvitationHandlers(
	commandBus cqrs.CommandBus,
	readModel *readmodels.InvitationReadModel,
) *InvitationHandlers {
	return &InvitationHandlers{
		commandBus:       commandBus,
		readModel:        readModel,
		paginationHelper: sharedAPI.NewPaginationHelper("/api/v1/invitations"),
	}
}

type CreateInvitationRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

func (req *CreateInvitationRequest) Validate() error {
	if _, err := valueobjects.NewEmail(req.Email); err != nil {
		return err
	}
	if _, err := valueobjects.RoleFromString(req.Role); err != nil {
		return err
	}
	return nil
}

func (h *InvitationHandlers) CreateInvitation(w http.ResponseWriter, r *http.Request) {
	req, err := h.parseAndValidateRequest(r)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid request")
		return
	}

	exists, err := h.readModel.ExistsPendingForEmail(r.Context(), req.Email)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to check existing invitations")
		return
	}
	if exists {
		sharedAPI.RespondError(w, http.StatusConflict, nil, "A pending invitation already exists for this email")
		return
	}

	cmd := &commands.CreateInvitation{Email: req.Email, Role: req.Role}
	if err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to create invitation")
		return
	}

	h.respondCreated(w, r, cmd.ID)
}

func (h *InvitationHandlers) parseAndValidateRequest(r *http.Request) (*CreateInvitationRequest, error) {
	var req CreateInvitationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	return &req, nil
}

func (h *InvitationHandlers) respondCreated(w http.ResponseWriter, r *http.Request, id string) {
	location := fmt.Sprintf("/api/v1/invitations/%s", id)
	w.Header().Set("Location", location)

	invitation, err := h.readModel.GetByID(r.Context(), id)
	if err != nil || invitation == nil {
		sharedAPI.RespondJSON(w, http.StatusCreated, map[string]string{"id": id})
		return
	}

	invitation.Links = h.invitationLinks(invitation.ID, invitation.Status)
	sharedAPI.RespondJSON(w, http.StatusCreated, invitation)
}

func (h *InvitationHandlers) GetAllInvitations(w http.ResponseWriter, r *http.Request) {
	params := sharedAPI.ParsePaginationParams(r)

	afterCursor, afterTimestamp, err := h.paginationHelper.ProcessCursor(params.After)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid pagination cursor")
		return
	}

	invitations, hasMore, err := h.readModel.GetAllPaginated(r.Context(), params.Limit, afterCursor, afterTimestamp)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve invitations")
		return
	}

	h.addLinksToInvitations(invitations)

	pageables := h.convertToPageable(invitations)
	nextCursor := h.paginationHelper.GenerateNextCursor(pageables, hasMore)
	selfLink := h.paginationHelper.BuildSelfLink(params)

	sharedAPI.RespondPaginated(w, sharedAPI.PaginatedResponseParams{
		StatusCode: http.StatusOK,
		Data:       invitations,
		HasMore:    hasMore,
		NextCursor: nextCursor,
		Limit:      params.Limit,
		SelfLink:   selfLink,
		BaseLink:   "/api/v1/invitations",
	})
}

func (h *InvitationHandlers) addLinksToInvitations(invitations []readmodels.InvitationDTO) {
	for i := range invitations {
		invitations[i].Links = h.invitationLinks(invitations[i].ID, invitations[i].Status)
	}
}

func (h *InvitationHandlers) GetInvitationByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	invitation, err := h.readModel.GetByID(r.Context(), id)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve invitation")
		return
	}

	if invitation == nil {
		sharedAPI.RespondError(w, http.StatusNotFound, nil, "Invitation not found")
		return
	}

	invitation.Links = h.invitationLinks(invitation.ID, invitation.Status)
	sharedAPI.RespondJSON(w, http.StatusOK, invitation)
}

func (h *InvitationHandlers) RevokeInvitation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	cmd := &commands.RevokeInvitation{ID: id}

	if err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		h.handleRevokeError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *InvitationHandlers) handleRevokeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, repositories.ErrInvitationNotFound):
		sharedAPI.RespondError(w, http.StatusNotFound, err, "Invitation not found")
	case errors.Is(err, aggregates.ErrInvitationAlreadyRevoked):
		sharedAPI.RespondError(w, http.StatusConflict, err, "Invitation already revoked")
	case errors.Is(err, aggregates.ErrInvitationNotPending):
		sharedAPI.RespondError(w, http.StatusConflict, err, "Invitation is not pending")
	default:
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to revoke invitation")
	}
}

func (h *InvitationHandlers) invitationLinks(id, status string) map[string]string {
	links := map[string]string{
		"self": fmt.Sprintf("/api/v1/invitations/%s", id),
	}
	if status == "pending" {
		links["revoke"] = fmt.Sprintf("/api/v1/invitations/%s/revoke", id)
	}
	return links
}

func (h *InvitationHandlers) convertToPageable(invitations []readmodels.InvitationDTO) []sharedAPI.Pageable {
	pageables := make([]sharedAPI.Pageable, len(invitations))
	for i := range invitations {
		pageables[i] = &invitationPageable{dto: &invitations[i]}
	}
	return pageables
}

type invitationPageable struct {
	dto *readmodels.InvitationDTO
}

func (p *invitationPageable) GetID() string {
	return p.dto.ID
}

func (p *invitationPageable) GetTimestamp() time.Time {
	return p.dto.CreatedAt
}
