package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"easi/backend/internal/platform/application/commands"
	"easi/backend/internal/platform/domain/aggregates"
	"easi/backend/internal/platform/domain/valueobjects"
	"easi/backend/internal/platform/infrastructure/repositories"
	"easi/backend/internal/platform/infrastructure/secrets"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/cqrs"
	sharedvo "easi/backend/internal/shared/domain/valueobjects"

	"github.com/go-chi/chi/v5"
)

type TenantHandlers struct {
	commandBus     *cqrs.InMemoryCommandBus
	repository     *repositories.TenantRepository
	secretProvider secrets.SecretProvider
}

func NewTenantHandlers(commandBus *cqrs.InMemoryCommandBus, repository *repositories.TenantRepository, secretProvider secrets.SecretProvider) *TenantHandlers {
	return &TenantHandlers{
		commandBus:     commandBus,
		repository:     repository,
		secretProvider: secretProvider,
	}
}

type CreateTenantRequest struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	Domains         []string          `json:"domains"`
	OIDCConfig      OIDCConfigRequest `json:"oidcConfig"`
	FirstAdminEmail string            `json:"firstAdminEmail"`
}

type OIDCConfigRequest struct {
	DiscoveryURL string `json:"discoveryUrl"`
	ClientID     string `json:"clientId"`
	AuthMethod   string `json:"authMethod"`
	Scopes       string `json:"scopes"`
}

func (r *CreateTenantRequest) Validate() error {
	if _, err := sharedvo.NewTenantID(r.ID); err != nil {
		return err
	}

	if _, err := valueobjects.NewTenantName(r.Name); err != nil {
		return err
	}

	if _, err := valueobjects.NewEmailDomainList(r.Domains); err != nil {
		return err
	}

	if _, err := valueobjects.NewOIDCConfig(
		r.OIDCConfig.DiscoveryURL,
		r.OIDCConfig.ClientID,
		valueobjects.OIDCAuthMethod(r.OIDCConfig.AuthMethod),
		r.OIDCConfig.Scopes,
	); err != nil {
		return err
	}

	return nil
}

type TenantResponse struct {
	ID         string                   `json:"id"`
	Name       string                   `json:"name"`
	Status     string                   `json:"status"`
	Domains    []string                 `json:"domains"`
	OIDCConfig *OIDCConfigResponse      `json:"oidcConfig,omitempty"`
	CreatedAt  time.Time                `json:"createdAt"`
	Links      map[string]sharedAPI.Link `json:"_links,omitempty"`
	Warnings   []string                 `json:"_warnings,omitempty"`
}

type OIDCConfigResponse struct {
	DiscoveryURL      string `json:"discoveryUrl"`
	ClientID          string `json:"clientId"`
	AuthMethod        string `json:"authMethod"`
	Scopes            string `json:"scopes"`
	SecretProvisioned bool   `json:"secretProvisioned"`
}

type TenantListItem struct {
	ID        string                   `json:"id"`
	Name      string                   `json:"name"`
	Status    string                   `json:"status"`
	Domains   []string                 `json:"domains"`
	CreatedAt time.Time                `json:"createdAt"`
	Links     map[string]sharedAPI.Link `json:"_links,omitempty"`
}

// CreateTenant godoc
// @Summary Create a new tenant
// @Description Creates a new tenant with OIDC configuration for SSO authentication
// @Tags tenants
// @Accept json
// @Produce json
// @Param request body CreateTenantRequest true "Tenant configuration"
// @Success 201 {object} TenantResponse "Tenant created successfully"
// @Failure 400 {object} sharedAPI.ErrorResponse "Invalid request or validation error"
// @Failure 409 {object} sharedAPI.ErrorResponse "Tenant or domain already exists"
// @Failure 500 {object} sharedAPI.ErrorResponse "Internal server error"
// @Router /platform/tenants [post]
func (h *TenantHandlers) CreateTenant(w http.ResponseWriter, r *http.Request) {
	var req CreateTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	if err := req.Validate(); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, err.Error())
		return
	}

	cmd := &commands.CreateTenant{
		ID:              req.ID,
		Name:            req.Name,
		Domains:         req.Domains,
		DiscoveryURL:    req.OIDCConfig.DiscoveryURL,
		ClientID:        req.OIDCConfig.ClientID,
		AuthMethod:      req.OIDCConfig.AuthMethod,
		Scopes:          req.OIDCConfig.Scopes,
		FirstAdminEmail: req.FirstAdminEmail,
	}

	if err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		statusCode := mapTenantErrorToStatusCode(err)
		sharedAPI.RespondError(w, statusCode, err, err.Error())
		return
	}

	record, err := h.repository.GetByID(r.Context(), req.ID)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve created tenant")
		return
	}

	response := h.mapRecordToResponse(r.Context(), record)

	location := fmt.Sprintf("/api/v1/platform/tenants/%s", req.ID)
	w.Header().Set("Location", location)
	sharedAPI.RespondJSON(w, http.StatusCreated, response)
}

// GetTenantByID godoc
// @Summary Get a tenant by ID
// @Description Retrieves detailed information about a specific tenant
// @Tags tenants
// @Produce json
// @Param id path string true "Tenant ID"
// @Success 200 {object} TenantResponse "Tenant details"
// @Failure 404 {object} sharedAPI.ErrorResponse "Tenant not found"
// @Failure 500 {object} sharedAPI.ErrorResponse "Internal server error"
// @Router /platform/tenants/{id} [get]
func (h *TenantHandlers) GetTenantByID(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "id")

	record, err := h.repository.GetByID(r.Context(), tenantID)
	if err != nil {
		if errors.Is(err, repositories.ErrTenantNotFound) {
			sharedAPI.RespondErrorWithLinks(w, http.StatusNotFound, err, "Tenant not found", map[string]sharedAPI.Link{
				"list":   {Href: "/api/v1/platform/tenants"},
				"create": {Href: "/api/v1/platform/tenants", Method: "POST"},
			})
			return
		}
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve tenant")
		return
	}

	response := h.mapRecordToResponse(r.Context(), record)
	sharedAPI.RespondJSON(w, http.StatusOK, response)
}

// ListTenants godoc
// @Summary List all tenants
// @Description Retrieves a list of all tenants with optional filtering
// @Tags tenants
// @Produce json
// @Param status query string false "Filter by status (e.g., 'active', 'suspended')"
// @Param domain query string false "Filter by email domain"
// @Success 200 {object} sharedAPI.CollectionResponse{data=[]TenantListItem} "List of tenants"
// @Failure 500 {object} sharedAPI.ErrorResponse "Internal server error"
// @Router /platform/tenants [get]
func (h *TenantHandlers) ListTenants(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	domain := r.URL.Query().Get("domain")

	records, err := h.repository.List(r.Context(), status, domain)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve tenants")
		return
	}

	items := make([]TenantListItem, len(records))
	for i, record := range records {
		items[i] = TenantListItem{
			ID:        record.ID,
			Name:      record.Name,
			Status:    record.Status,
			Domains:   record.Domains,
			CreatedAt: record.CreatedAt,
			Links: map[string]sharedAPI.Link{
				"self": {Href: fmt.Sprintf("/api/v1/platform/tenants/%s", record.ID)},
			},
		}
	}

	sharedAPI.RespondCollection(w, http.StatusOK, items, map[string]string{
		"self": "/api/v1/platform/tenants",
	})
}

func (h *TenantHandlers) mapRecordToResponse(ctx context.Context, record *repositories.TenantRecord) TenantResponse {
	secretProvisioned := h.secretProvider.IsProvisioned(ctx, record.ID)

	var warnings []string
	if !secretProvisioned {
		warnings = append(warnings, "OIDC secret not provisioned")
	}

	basePath := fmt.Sprintf("/api/v1/platform/tenants/%s", record.ID)
	return TenantResponse{
		ID:      record.ID,
		Name:    record.Name,
		Status:  record.Status,
		Domains: record.Domains,
		OIDCConfig: &OIDCConfigResponse{
			DiscoveryURL:      record.DiscoveryURL,
			ClientID:          record.ClientID,
			AuthMethod:        record.AuthMethod,
			Scopes:            record.Scopes,
			SecretProvisioned: secretProvisioned,
		},
		CreatedAt: record.CreatedAt,
		Links: map[string]sharedAPI.Link{
			"self":       {Href: basePath},
			"domains":    {Href: basePath + "/domains"},
			"oidcConfig": {Href: basePath + "/oidc-config", Method: "PATCH"},
			"suspend":    {Href: basePath + "/suspend", Method: "POST"},
			"users":      {Href: fmt.Sprintf("/api/v1/users?tenant=%s", record.ID)},
		},
		Warnings: warnings,
	}
}

type CreateInvitationRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

// CreateTenantInvitation godoc
// @Summary Create an invitation for a tenant
// @Description Creates an admin invitation for an existing tenant
// @Tags tenants
// @Accept json
// @Produce json
// @Param id path string true "Tenant ID"
// @Param request body CreateInvitationRequest true "Invitation details"
// @Success 201 {object} map[string]string "Invitation created"
// @Failure 400 {object} sharedAPI.ErrorResponse "Invalid request"
// @Failure 404 {object} sharedAPI.ErrorResponse "Tenant not found"
// @Failure 500 {object} sharedAPI.ErrorResponse "Internal server error"
// @Router /platform/tenants/{id}/invitations [post]
func (h *TenantHandlers) CreateTenantInvitation(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "id")

	exists, err := h.repository.ExistsByID(r.Context(), tenantID)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to check tenant")
		return
	}
	if !exists {
		sharedAPI.RespondError(w, http.StatusNotFound, nil, "Tenant not found")
		return
	}

	var req CreateInvitationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	if req.Email == "" {
		sharedAPI.RespondError(w, http.StatusBadRequest, nil, "Email is required")
		return
	}
	if req.Role == "" {
		req.Role = "admin"
	}

	if err := h.repository.CreateInvitation(r.Context(), tenantID, req.Email, req.Role); err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to create invitation")
		return
	}

	sharedAPI.RespondJSON(w, http.StatusCreated, map[string]string{
		"message": "Invitation created for " + req.Email,
		"tenant":  tenantID,
		"email":   req.Email,
		"role":    req.Role,
	})
}

func mapTenantErrorToStatusCode(err error) int {
	switch {
	case errors.Is(err, repositories.ErrTenantAlreadyExists):
		return http.StatusConflict
	case errors.Is(err, repositories.ErrDomainAlreadyExists):
		return http.StatusConflict
	case errors.Is(err, aggregates.ErrFirstAdminEmailRequired):
		return http.StatusBadRequest
	case errors.Is(err, aggregates.ErrFirstAdminEmailDomainMismatch):
		return http.StatusBadRequest
	case errors.Is(err, sharedvo.ErrInvalidTenantIDFormat):
		return http.StatusBadRequest
	case errors.Is(err, sharedvo.ErrReservedTenantID):
		return http.StatusBadRequest
	case errors.Is(err, valueobjects.ErrTenantNameEmpty):
		return http.StatusBadRequest
	case errors.Is(err, valueobjects.ErrEmailDomainListEmpty):
		return http.StatusBadRequest
	case errors.Is(err, valueobjects.ErrInvalidEmailDomain):
		return http.StatusBadRequest
	case errors.Is(err, valueobjects.ErrDuplicateEmailDomain):
		return http.StatusBadRequest
	case errors.Is(err, valueobjects.ErrOIDCDiscoveryURLEmpty):
		return http.StatusBadRequest
	case errors.Is(err, valueobjects.ErrOIDCDiscoveryURLInvalid):
		return http.StatusBadRequest
	case errors.Is(err, valueobjects.ErrOIDCDiscoveryURLNotHTTPS):
		return http.StatusBadRequest
	case errors.Is(err, valueobjects.ErrOIDCClientIDEmpty):
		return http.StatusBadRequest
	case errors.Is(err, valueobjects.ErrOIDCAuthMethodInvalid):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
