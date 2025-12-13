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
	ID         string              `json:"id"`
	Name       string              `json:"name"`
	Status     string              `json:"status"`
	Domains    []string            `json:"domains"`
	OIDCConfig *OIDCConfigResponse `json:"oidcConfig,omitempty"`
	CreatedAt  time.Time           `json:"createdAt"`
	Links      map[string]string   `json:"_links,omitempty"`
	Warnings   []string            `json:"_warnings,omitempty"`
}

type OIDCConfigResponse struct {
	DiscoveryURL      string `json:"discoveryUrl"`
	ClientID          string `json:"clientId"`
	AuthMethod        string `json:"authMethod"`
	Scopes            string `json:"scopes"`
	SecretProvisioned bool   `json:"secretProvisioned"`
}

type TenantListItem struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Status    string            `json:"status"`
	Domains   []string          `json:"domains"`
	CreatedAt time.Time         `json:"createdAt"`
	Links     map[string]string `json:"_links,omitempty"`
}

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

	location := fmt.Sprintf("/api/platform/v1/tenants/%s", req.ID)
	w.Header().Set("Location", location)
	sharedAPI.RespondJSON(w, http.StatusCreated, response)
}

func (h *TenantHandlers) GetTenantByID(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "id")

	record, err := h.repository.GetByID(r.Context(), tenantID)
	if err != nil {
		if errors.Is(err, repositories.ErrTenantNotFound) {
			sharedAPI.RespondError(w, http.StatusNotFound, err, "Tenant not found")
			return
		}
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve tenant")
		return
	}

	response := h.mapRecordToResponse(r.Context(), record)
	sharedAPI.RespondJSON(w, http.StatusOK, response)
}

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
			Links: map[string]string{
				"self": fmt.Sprintf("/api/platform/v1/tenants/%s", record.ID),
			},
		}
	}

	sharedAPI.RespondPaginated(w, sharedAPI.PaginatedResponseParams{
		StatusCode: http.StatusOK,
		Data:       items,
		HasMore:    false,
		NextCursor: "",
		Limit:      50,
		SelfLink:   "/api/platform/v1/tenants",
		BaseLink:   "/api/platform/v1/tenants",
	})
}

func (h *TenantHandlers) mapRecordToResponse(ctx context.Context, record *repositories.TenantRecord) TenantResponse {
	secretProvisioned := h.secretProvider.IsProvisioned(ctx, record.ID)

	var warnings []string
	if !secretProvisioned {
		warnings = append(warnings, "OIDC secret not provisioned")
	}

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
		Links: map[string]string{
			"self":       fmt.Sprintf("/api/platform/v1/tenants/%s", record.ID),
			"domains":    fmt.Sprintf("/api/platform/v1/tenants/%s/domains", record.ID),
			"oidcConfig": fmt.Sprintf("/api/platform/v1/tenants/%s/oidc-config", record.ID),
			"suspend":    fmt.Sprintf("/api/platform/v1/tenants/%s/suspend", record.ID),
			"users":      fmt.Sprintf("/api/v1/users?tenant=%s", record.ID),
		},
		Warnings: warnings,
	}
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
