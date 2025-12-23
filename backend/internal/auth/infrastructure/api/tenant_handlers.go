package api

import (
	"net/http"

	"easi/backend/internal/auth/application/readmodels"
	"easi/backend/internal/auth/domain/valueobjects"
	"easi/backend/internal/auth/infrastructure/repositories"
	"easi/backend/internal/auth/infrastructure/session"
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

type TenantHandlers struct {
	tenantRepo     *repositories.TenantRepository
	userReadModel  *readmodels.UserReadModel
	sessionManager *session.SessionManager
}

func NewTenantHandlers(
	tenantRepo *repositories.TenantRepository,
	userReadModel *readmodels.UserReadModel,
	sessionManager *session.SessionManager,
) *TenantHandlers {
	return &TenantHandlers{
		tenantRepo:     tenantRepo,
		userReadModel:  userReadModel,
		sessionManager: sessionManager,
	}
}

type CurrentTenantResponse struct {
	ID      string            `json:"id"`
	Name    string            `json:"name"`
	Domains []string          `json:"domains"`
	Links   map[string]string `json:"_links"`
}

// GetCurrentTenant godoc
// @Summary Get current tenant
// @Description Returns information about the current user's tenant including registered domains
// @Tags tenants
// @Accept json
// @Produce json
// @Success 200 {object} CurrentTenantResponse "Tenant details with HATEOAS links"
// @Failure 401 {object} sharedAPI.ErrorResponse "Not authenticated"
// @Failure 500 {object} sharedAPI.ErrorResponse "Internal server error"
// @Router /tenants/current [get]
func (h *TenantHandlers) GetCurrentTenant(w http.ResponseWriter, r *http.Request) {
	authSession, err := h.sessionManager.LoadAuthenticatedSession(r.Context())
	if err != nil {
		sharedAPI.RespondError(w, http.StatusUnauthorized, err, "Not authenticated")
		return
	}

	tenantID := authSession.TenantID()
	tenant, err := h.tenantRepo.GetByID(r.Context(), tenantID)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve tenant")
		return
	}

	domains, err := h.tenantRepo.GetDomains(r.Context(), tenantID)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve tenant domains")
		return
	}

	tenantIDVO, err := sharedvo.NewTenantID(tenantID)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Invalid tenant ID")
		return
	}
	ctx := sharedctx.WithTenant(r.Context(), tenantIDVO)

	userEmail := authSession.UserEmail()
	user, err := h.userReadModel.GetByEmail(ctx, userEmail)
	if err != nil || user == nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to get user")
		return
	}

	role, err := valueobjects.RoleFromString(user.Role)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Invalid user role")
		return
	}
	permissions := valueobjects.PermissionsToStrings(role.Permissions())

	response := CurrentTenantResponse{
		ID:      tenant.ID,
		Name:    tenant.Name,
		Domains: domains,
		Links:   h.tenantLinks(user.Role, permissions),
	}

	sharedAPI.RespondJSON(w, http.StatusOK, response)
}

func (h *TenantHandlers) tenantLinks(userRole string, permissions []string) map[string]string {
	links := map[string]string{
		"self": "/api/v1/tenants/current",
	}

	hasUsersRead := hasPermission(permissions, valueobjects.PermUsersRead.String())
	hasInvitationsManage := hasPermission(permissions, valueobjects.PermInvitationsManage.String())

	if userRole == "admin" || hasUsersRead {
		links["users"] = "/api/v1/users"
	}

	if userRole == "admin" || hasInvitationsManage {
		links["invitations"] = "/api/v1/invitations"
	}

	return links
}

func hasPermission(permissions []string, perm string) bool {
	for _, p := range permissions {
		if p == perm {
			return true
		}
	}
	return false
}
