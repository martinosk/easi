package api

import (
	"net/http"

	"easi/backend/internal/auth/infrastructure/session"
	"easi/backend/internal/metamodel/application/commands"
	"easi/backend/internal/metamodel/application/readmodels"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/cqrs"

	"github.com/go-chi/chi/v5"
)

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
}

// GetMaturityScale godoc
// @Summary Get the maturity scale configuration
// @Description Retrieves the maturity scale configuration for the current tenant
// @Tags metamodel
// @Produce json
// @Success 200 {object} readmodels.MetaModelConfigurationDTO
// @Failure 404 {object} easi_backend_internal_shared_api.ErrorResponse
// @Failure 500 {object} easi_backend_internal_shared_api.ErrorResponse
// @Router /metamodel/maturity-scale [get]
func (h *MetaModelHandlers) GetMaturityScale(w http.ResponseWriter, r *http.Request) {
	config, err := h.readModel.GetByTenantID(r.Context())
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve maturity scale configuration")
		return
	}

	if config == nil {
		sharedAPI.RespondError(w, http.StatusNotFound, nil, "Maturity scale configuration not found")
		return
	}

	config.Links = h.hateoas.MaturityScaleLinks()

	sharedAPI.RespondJSON(w, http.StatusOK, config)
}

// UpdateMaturityScale godoc
// @Summary Update the maturity scale configuration
// @Description Updates the maturity scale sections for the current tenant
// @Tags metamodel
// @Accept json
// @Produce json
// @Param scale body UpdateMaturityScaleRequest true "Maturity scale configuration"
// @Success 200 {object} readmodels.MetaModelConfigurationDTO
// @Failure 400 {object} easi_backend_internal_shared_api.ErrorResponse
// @Failure 401 {object} easi_backend_internal_shared_api.ErrorResponse
// @Failure 403 {object} easi_backend_internal_shared_api.ErrorResponse
// @Failure 404 {object} easi_backend_internal_shared_api.ErrorResponse
// @Failure 500 {object} easi_backend_internal_shared_api.ErrorResponse
// @Router /metamodel/maturity-scale [put]
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

	config, err := h.readModel.GetByTenantID(r.Context())
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve configuration")
		return
	}

	if config == nil {
		sharedAPI.RespondError(w, http.StatusNotFound, nil, "Configuration not found")
		return
	}

	var sections [4]commands.MaturitySectionInput
	for i, s := range req.Sections {
		sections[i] = commands.MaturitySectionInput{
			Order:    s.Order,
			Name:     s.Name,
			MinValue: s.MinValue,
			MaxValue: s.MaxValue,
		}
	}

	cmd := &commands.UpdateMaturityScale{
		ID:         config.ID,
		Sections:   sections,
		ModifiedBy: authSession.UserEmail(),
	}

	if err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Failed to update maturity scale")
		return
	}

	updatedConfig, err := h.readModel.GetByID(r.Context(), config.ID)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve updated configuration")
		return
	}

	updatedConfig.Links = h.hateoas.MaturityScaleLinks()

	sharedAPI.RespondJSON(w, http.StatusOK, updatedConfig)
}

// ResetMaturityScale godoc
// @Summary Reset maturity scale to defaults
// @Description Resets the maturity scale configuration to default values
// @Tags metamodel
// @Produce json
// @Success 200 {object} readmodels.MetaModelConfigurationDTO
// @Failure 401 {object} easi_backend_internal_shared_api.ErrorResponse
// @Failure 403 {object} easi_backend_internal_shared_api.ErrorResponse
// @Failure 404 {object} easi_backend_internal_shared_api.ErrorResponse
// @Failure 500 {object} easi_backend_internal_shared_api.ErrorResponse
// @Router /metamodel/maturity-scale/reset [put]
func (h *MetaModelHandlers) ResetMaturityScale(w http.ResponseWriter, r *http.Request) {
	authSession, err := h.sessionManager.LoadAuthenticatedSession(r.Context())
	if err != nil {
		sharedAPI.RespondError(w, http.StatusUnauthorized, err, "Authentication required")
		return
	}

	config, err := h.readModel.GetByTenantID(r.Context())
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve configuration")
		return
	}

	if config == nil {
		sharedAPI.RespondError(w, http.StatusNotFound, nil, "Configuration not found")
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

	updatedConfig, err := h.readModel.GetByID(r.Context(), config.ID)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve reset configuration")
		return
	}

	updatedConfig.Links = h.hateoas.MaturityScaleLinks()

	sharedAPI.RespondJSON(w, http.StatusOK, updatedConfig)
}

// GetMaturityScaleByID godoc
// @Summary Get maturity scale by configuration ID
// @Description Retrieves a specific maturity scale configuration by ID
// @Tags metamodel
// @Produce json
// @Param id path string true "Configuration ID"
// @Success 200 {object} readmodels.MetaModelConfigurationDTO
// @Failure 404 {object} easi_backend_internal_shared_api.ErrorResponse
// @Failure 500 {object} easi_backend_internal_shared_api.ErrorResponse
// @Router /metamodel/configurations/{id} [get]
func (h *MetaModelHandlers) GetMaturityScaleByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	config, err := h.readModel.GetByID(r.Context(), id)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve configuration")
		return
	}

	if config == nil {
		sharedAPI.RespondError(w, http.StatusNotFound, nil, "Configuration not found")
		return
	}

	config.Links = h.hateoas.MetaModelConfigLinks(id)

	sharedAPI.RespondJSON(w, http.StatusOK, config)
}
