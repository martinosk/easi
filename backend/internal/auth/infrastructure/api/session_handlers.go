package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"

	"easi/backend/internal/auth/domain/valueobjects"
	"easi/backend/internal/auth/infrastructure/session"
	sharedAPI "easi/backend/internal/shared/api"
)

type UserDTO struct {
	ID     uuid.UUID
	Email  string
	Name   string
	Role   string
	Status string
}

type TenantDTO struct {
	ID   string
	Name string
}

type UserRepository interface {
	GetByEmail(ctx context.Context, tenantID, email string) (*UserDTO, error)
}

type TenantRepository interface {
	GetByID(ctx context.Context, tenantID string) (*TenantDTO, error)
}

type SessionHandlers struct {
	sessionManager *session.SessionManager
	userRepo       UserRepository
	tenantRepo     TenantRepository
}

func NewSessionHandlers(
	sessionManager *session.SessionManager,
	userRepo UserRepository,
	tenantRepo TenantRepository,
) *SessionHandlers {
	return &SessionHandlers{
		sessionManager: sessionManager,
		userRepo:       userRepo,
		tenantRepo:     tenantRepo,
	}
}

type CurrentSessionResponse struct {
	ID        string                 `json:"id"`
	User      CurrentSessionUser     `json:"user"`
	Tenant    CurrentSessionTenant   `json:"tenant"`
	ExpiresAt time.Time              `json:"expiresAt"`
	Links     map[string]string      `json:"_links"`
}

type CurrentSessionUser struct {
	ID          string   `json:"id"`
	Email       string   `json:"email"`
	Name        string   `json:"name"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
}

type CurrentSessionTenant struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (h *SessionHandlers) GetCurrentSession(w http.ResponseWriter, r *http.Request) {
	authSession, err := h.sessionManager.LoadAuthenticatedSession(r.Context())
	if err != nil || !authSession.IsAuthenticated() {
		sharedAPI.RespondError(w, http.StatusUnauthorized, err, "No valid session")
		return
	}

	user, err := h.userRepo.GetByEmail(r.Context(), authSession.TenantID(), authSession.UserEmail())
	if err != nil {
		sharedAPI.RespondError(w, http.StatusUnauthorized, err, "User not found")
		return
	}

	tenant, err := h.tenantRepo.GetByID(r.Context(), authSession.TenantID())
	if err != nil {
		sharedAPI.RespondError(w, http.StatusUnauthorized, err, "Tenant not found")
		return
	}

	role, err := valueobjects.RoleFromString(user.Role)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Invalid role")
		return
	}

	permissions := valueobjects.PermissionsToStrings(role.Permissions())

	response := CurrentSessionResponse{
		ID: authSession.TenantID(),
		User: CurrentSessionUser{
			ID:          user.ID.String(),
			Email:       user.Email,
			Name:        user.Name,
			Role:        user.Role,
			Permissions: permissions,
		},
		Tenant: CurrentSessionTenant{
			ID:   tenant.ID,
			Name: tenant.Name,
		},
		ExpiresAt: authSession.TokenExpiry(),
		Links: map[string]string{
			"self":   "/auth/sessions/current",
			"logout": "/auth/sessions/current",
			"user":   fmt.Sprintf("/api/v1/users/%s", user.ID),
			"tenant": "/api/v1/tenants/current",
		},
	}

	sharedAPI.RespondJSON(w, http.StatusOK, response)
}

func (h *SessionHandlers) DeleteCurrentSession(w http.ResponseWriter, r *http.Request) {
	if err := h.sessionManager.ClearSession(r.Context()); err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to logout")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
