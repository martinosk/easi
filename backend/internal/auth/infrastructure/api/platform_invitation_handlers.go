package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"easi/backend/internal/auth/application/commands"
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/cqrs"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

const defaultInvitationRole = "admin"

type TenantCatalog interface {
	ExistsByID(ctx context.Context, tenantID string) (bool, error)
}

type PlatformInvitationHandlers struct {
	commandBus   cqrs.CommandBus
	tenants      TenantCatalog
	errorHandler *sharedAPI.ErrorHandler
}

func NewPlatformInvitationHandlers(commandBus cqrs.CommandBus, tenants TenantCatalog) *PlatformInvitationHandlers {
	return &PlatformInvitationHandlers{
		commandBus:   commandBus,
		tenants:      tenants,
		errorHandler: sharedAPI.NewErrorHandler(),
	}
}

type PlatformInvitationRequest struct {
	TenantID string `json:"tenantId"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

// CreateInvitation godoc
// @Summary Invite a user into a tenant
// @Description Creates an invitation in the named tenant. Guarded by the platform admin API key.
// @Tags invitations
// @Accept json
// @Produce json
// @Param request body PlatformInvitationRequest true "Tenant, email and role"
// @Success 201 {object} map[string]string "Created invitation"
// @Failure 400 {object} sharedAPI.ErrorResponse "Invalid request body, tenant ID or email"
// @Failure 401 {object} sharedAPI.ErrorResponse "Missing or invalid platform admin API key"
// @Failure 404 {object} sharedAPI.ErrorResponse "Tenant not found"
// @Failure 500 {object} sharedAPI.ErrorResponse "Internal server error"
// @Router /auth/invitations [post]
func (h *PlatformInvitationHandlers) CreateInvitation(w http.ResponseWriter, r *http.Request) {
	tenantID, req, ok := h.parseRequest(w, r)
	if !ok {
		return
	}

	known, err := h.tenants.ExistsByID(r.Context(), tenantID.Value())
	if err != nil {
		h.errorHandler.HandleError(w, err, "Failed to look up tenant")
		return
	}
	if !known {
		sharedAPI.RespondError(w, http.StatusNotFound, nil, "Tenant not found")
		return
	}

	tenantCtx := sharedctx.WithTenant(r.Context(), tenantID)
	cmd := &commands.CreateInvitation{Email: req.Email, Role: req.Role}

	result, err := h.commandBus.Dispatch(tenantCtx, cmd)
	if err != nil {
		h.errorHandler.HandleError(w, fmt.Errorf("invite %s into tenant %s: %w", req.Email, tenantID.Value(), err), "Failed to create invitation")
		return
	}

	sharedAPI.RespondJSON(w, http.StatusCreated, map[string]string{
		"id":       result.CreatedID,
		"tenantId": tenantID.Value(),
		"email":    req.Email,
		"role":     req.Role,
	})
}

func (h *PlatformInvitationHandlers) parseRequest(w http.ResponseWriter, r *http.Request) (sharedvo.TenantID, PlatformInvitationRequest, bool) {
	var req PlatformInvitationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid request body")
		return sharedvo.TenantID{}, req, false
	}

	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" {
		sharedAPI.RespondError(w, http.StatusBadRequest, nil, "Email is required")
		return sharedvo.TenantID{}, req, false
	}

	if req.Role == "" {
		req.Role = defaultInvitationRole
	}

	tenantID, err := sharedvo.NewTenantID(req.TenantID)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid tenant ID")
		return sharedvo.TenantID{}, req, false
	}

	return tenantID, req, true
}
