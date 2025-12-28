package api

import (
	"fmt"
	"net/http"
	"time"

	"easi/backend/internal/auth/infrastructure/session"
	"easi/backend/internal/metamodel/application/commands"
	"easi/backend/internal/metamodel/application/readmodels"
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/cqrs"

	"github.com/go-chi/chi/v5"
)

func setMaturityScaleCacheHeaders(w http.ResponseWriter, version int) {
	w.Header().Set("Cache-Control", "private, max-age=300")
	w.Header().Set("ETag", fmt.Sprintf(`"v%d"`, version))
}

type MetaModelHandlers struct {
	commandBus     cqrs.CommandBus
	readModel      *readmodels.MetaModelConfigurationReadModel
	hateoas        *sharedAPI.HATEOASLinks
	sessionManager *session.SessionManager
}

func NewMetaModelHandlers(
	commandBus cqrs.CommandBus,
	readModel *readmodels.MetaModelConfigurationReadModel,
	hateoas *sharedAPI.HATEOASLinks,
	sessionManager *session.SessionManager,
) *MetaModelHandlers {
	return &MetaModelHandlers{
		commandBus:     commandBus,
		readModel:      readModel,
		hateoas:        hateoas,
		sessionManager: sessionManager,
	}
}

type MaturitySectionRequest struct {
	Order    int    `json:"order"`
	Name     string `json:"name"`
	MinValue int    `json:"minValue"`
	MaxValue int    `json:"maxValue"`
}

type UpdateMaturityScaleRequest struct {
	Sections [4]MaturitySectionRequest `json:"sections"`
	Version  int                       `json:"version"`
}

func (req UpdateMaturityScaleRequest) toCommandInput() [4]commands.MaturitySectionInput {
	var sections [4]commands.MaturitySectionInput
	for i, s := range req.Sections {
		sections[i] = commands.MaturitySectionInput{
			Order:    s.Order,
			Name:     s.Name,
			MinValue: s.MinValue,
			MaxValue: s.MaxValue,
		}
	}
	return sections
}

type configLoader func() (*readmodels.MetaModelConfigurationDTO, error)
type linkGenerator func(config *readmodels.MetaModelConfigurationDTO) map[string]string

func defaultMaturityScaleConfig() *readmodels.MetaModelConfigurationDTO {
	now := time.Now()
	return &readmodels.MetaModelConfigurationDTO{
		ID:        "",
		TenantID:  "",
		Version:   0,
		IsDefault: true,
		Sections: []readmodels.MaturitySectionDTO{
			{Order: 1, Name: "Genesis", MinValue: 0, MaxValue: 24},
			{Order: 2, Name: "Custom Built", MinValue: 25, MaxValue: 49},
			{Order: 3, Name: "Product", MinValue: 50, MaxValue: 74},
			{Order: 4, Name: "Commodity", MinValue: 75, MaxValue: 99},
		},
		CreatedAt:  now,
		ModifiedAt: now,
		ModifiedBy: "system",
	}
}

func (h *MetaModelHandlers) loadConfigOrFail(w http.ResponseWriter, loader configLoader) *readmodels.MetaModelConfigurationDTO {
	config, err := loader()
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve configuration")
		return nil
	}
	if config == nil {
		sharedAPI.RespondError(w, http.StatusNotFound, nil, "Configuration not found")
		return nil
	}
	return config
}

func (h *MetaModelHandlers) loadConfigOrDefault(loader configLoader) (*readmodels.MetaModelConfigurationDTO, error) {
	config, err := loader()
	if err != nil {
		return nil, err
	}
	if config == nil {
		return defaultMaturityScaleConfig(), nil
	}
	return config, nil
}

func (h *MetaModelHandlers) getAndRespondWithConfig(w http.ResponseWriter, loader configLoader, linkGen linkGenerator) {
	config := h.loadConfigOrFail(w, loader)
	if config == nil {
		return
	}
	setMaturityScaleCacheHeaders(w, config.Version)
	config.Links = linkGen(config)
	sharedAPI.RespondJSON(w, http.StatusOK, config)
}

func (h *MetaModelHandlers) ensureConfigExists(w http.ResponseWriter, r *http.Request, userEmail string) (*readmodels.MetaModelConfigurationDTO, bool) {
	config, err := h.readModel.GetByTenantID(r.Context())
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve configuration")
		return nil, false
	}

	if config != nil {
		return config, true
	}

	tenantID, err := sharedctx.GetTenant(r.Context())
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to get tenant")
		return nil, false
	}

	createCmd := &commands.CreateMetaModelConfiguration{
		TenantID:  tenantID.Value(),
		CreatedBy: userEmail,
	}
	if err := h.commandBus.Dispatch(r.Context(), createCmd); err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to create configuration")
		return nil, false
	}

	config, err = h.readModel.GetByTenantID(r.Context())
	if err != nil || config == nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve created configuration")
		return nil, false
	}
	return config, true
}

// GetMaturityScale godoc
// @Summary Get the maturity scale configuration
// @Description Retrieves the maturity scale configuration for the current tenant. Returns default configuration if none exists.
// @Tags meta-model
// @Produce json
// @Success 200 {object} readmodels.MetaModelConfigurationDTO
// @Failure 500 {object} easi_backend_internal_shared_api.ErrorResponse
// @Router /meta-model/maturity-scale [get]
func (h *MetaModelHandlers) GetMaturityScale(w http.ResponseWriter, r *http.Request) {
	config, err := h.loadConfigOrDefault(func() (*readmodels.MetaModelConfigurationDTO, error) {
		return h.readModel.GetByTenantID(r.Context())
	})
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve configuration")
		return
	}
	setMaturityScaleCacheHeaders(w, config.Version)
	config.Links = h.hateoas.MaturityScaleLinks(config.IsDefault)
	sharedAPI.RespondJSON(w, http.StatusOK, config)
}

// UpdateMaturityScale godoc
// @Summary Update the maturity scale configuration
// @Description Updates the maturity scale sections for the current tenant. Creates config if it doesn't exist.
// @Tags meta-model
// @Accept json
// @Produce json
// @Param scale body UpdateMaturityScaleRequest true "Maturity scale configuration"
// @Success 200 {object} readmodels.MetaModelConfigurationDTO
// @Failure 400 {object} easi_backend_internal_shared_api.ErrorResponse
// @Failure 401 {object} easi_backend_internal_shared_api.ErrorResponse
// @Failure 403 {object} easi_backend_internal_shared_api.ErrorResponse
// @Failure 409 {object} easi_backend_internal_shared_api.ErrorResponse
// @Failure 500 {object} easi_backend_internal_shared_api.ErrorResponse
// @Router /meta-model/maturity-scale [put]
func (h *MetaModelHandlers) UpdateMaturityScale(w http.ResponseWriter, r *http.Request) {
	authSession, err := h.sessionManager.LoadAuthenticatedSession(r.Context())
	if err != nil {
		sharedAPI.RespondError(w, http.StatusUnauthorized, err, "Authentication required")
		return
	}

	req, ok := sharedAPI.DecodeRequestOrFail[UpdateMaturityScaleRequest](w, r)
	if !ok {
		return
	}

	config, ok := h.ensureConfigExists(w, r, authSession.UserEmail())
	if !ok {
		return
	}

	if req.Version != config.Version {
		sharedAPI.RespondError(w, http.StatusConflict, nil, "Configuration was modified by another user. Please refresh and try again.")
		return
	}

	cmd := &commands.UpdateMaturityScale{
		ID:         config.ID,
		Sections:   req.toCommandInput(),
		ModifiedBy: authSession.UserEmail(),
	}

	if err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Failed to update maturity scale")
		return
	}

	h.getAndRespondWithConfig(w,
		func() (*readmodels.MetaModelConfigurationDTO, error) {
			return h.readModel.GetByID(r.Context(), config.ID)
		},
		func(c *readmodels.MetaModelConfigurationDTO) map[string]string {
			return h.hateoas.MaturityScaleLinks(c.IsDefault)
		},
	)
}

// ResetMaturityScale godoc
// @Summary Reset maturity scale to defaults
// @Description Resets the maturity scale configuration to default values. Creates config if it doesn't exist.
// @Tags meta-model
// @Produce json
// @Success 200 {object} readmodels.MetaModelConfigurationDTO
// @Failure 401 {object} easi_backend_internal_shared_api.ErrorResponse
// @Failure 403 {object} easi_backend_internal_shared_api.ErrorResponse
// @Failure 500 {object} easi_backend_internal_shared_api.ErrorResponse
// @Router /meta-model/maturity-scale/reset [post]
func (h *MetaModelHandlers) ResetMaturityScale(w http.ResponseWriter, r *http.Request) {
	authSession, err := h.sessionManager.LoadAuthenticatedSession(r.Context())
	if err != nil {
		sharedAPI.RespondError(w, http.StatusUnauthorized, err, "Authentication required")
		return
	}

	config, ok := h.ensureConfigExists(w, r, authSession.UserEmail())
	if !ok {
		return
	}

	cmd := &commands.ResetMaturityScale{
		ID:         config.ID,
		ModifiedBy: authSession.UserEmail(),
	}

	if err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to reset maturity scale")
		return
	}

	h.getAndRespondWithConfig(w,
		func() (*readmodels.MetaModelConfigurationDTO, error) {
			return h.readModel.GetByID(r.Context(), config.ID)
		},
		func(c *readmodels.MetaModelConfigurationDTO) map[string]string {
			return h.hateoas.MaturityScaleLinks(c.IsDefault)
		},
	)
}

// GetMaturityScaleByID godoc
// @Summary Get maturity scale by configuration ID
// @Description Retrieves a specific maturity scale configuration by ID
// @Tags meta-model
// @Produce json
// @Param id path string true "Configuration ID"
// @Success 200 {object} readmodels.MetaModelConfigurationDTO
// @Failure 404 {object} easi_backend_internal_shared_api.ErrorResponse
// @Failure 500 {object} easi_backend_internal_shared_api.ErrorResponse
// @Router /meta-model/configurations/{id} [get]
func (h *MetaModelHandlers) GetMaturityScaleByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	h.getAndRespondWithConfig(w,
		func() (*readmodels.MetaModelConfigurationDTO, error) { return h.readModel.GetByID(r.Context(), id) },
		func(c *readmodels.MetaModelConfigurationDTO) map[string]string {
			return h.hateoas.MetaModelConfigLinks(id, c.IsDefault)
		},
	)
}
