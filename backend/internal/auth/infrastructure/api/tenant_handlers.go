package api

import (
	"net/http"
	"slices"

	"easi/backend/internal/auth/application/readmodels"
	"easi/backend/internal/auth/domain/valueobjects"
	"easi/backend/internal/auth/infrastructure/repositories"
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
)

type TenantHandlers struct {
	tenantRepo    *repositories.TenantRepository
	userReadModel *readmodels.UserReadModel
}

func NewTenantHandlers(
	tenantRepo *repositories.TenantRepository,
	userReadModel *readmodels.UserReadModel,
) *TenantHandlers {
	return &TenantHandlers{
		tenantRepo:    tenantRepo,
		userReadModel: userReadModel,
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
	ctx := r.Context()
	actor, ok := sharedctx.GetActor(ctx)
	if !ok {
		sharedAPI.RespondError(w, http.StatusUnauthorized, nil, "Not authenticated")
		return
	}

	tenantIDVO, err := sharedctx.GetTenant(ctx)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusUnauthorized, err, "Not authenticated")
		return
	}
	tenantID := tenantIDVO.Value()

	tenant, err := h.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve tenant")
		return
	}

	domains, err := h.tenantRepo.GetDomains(ctx, tenantID)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve tenant domains")
		return
	}

	user, err := h.userReadModel.GetByEmail(ctx, actor.Email)
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

	hasUsersRead := slices.Contains(permissions, valueobjects.PermUsersRead.String())
	hasInvitationsManage := slices.Contains(permissions, valueobjects.PermInvitationsManage.String())

	if userRole == "admin" || hasUsersRead {
		links["users"] = "/api/v1/users"
	}

	if userRole == "admin" || hasInvitationsManage {
		links["invitations"] = "/api/v1/invitations"
	}

	return links
}
