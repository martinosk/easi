package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"easi/backend/internal/auth/infrastructure/session"
	"easi/backend/internal/metamodel/application/commands"
	"easi/backend/internal/metamodel/application/readmodels"
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/cqrs"

	"github.com/go-chi/chi/v5"
)

type StrategyPillarsHandlers struct {
	commandBus     cqrs.CommandBus
	readModel      *readmodels.MetaModelConfigurationReadModel
	hateoas        *sharedAPI.HATEOASLinks
	sessionManager *session.SessionManager
}

func NewStrategyPillarsHandlers(
	commandBus cqrs.CommandBus,
	readModel *readmodels.MetaModelConfigurationReadModel,
	hateoas *sharedAPI.HATEOASLinks,
	sessionManager *session.SessionManager,
) *StrategyPillarsHandlers {
	return &StrategyPillarsHandlers{
		commandBus:     commandBus,
		readModel:      readModel,
		hateoas:        hateoas,
		sessionManager: sessionManager,
	}
}

type CreateStrategyPillarRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdateStrategyPillarRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type StrategyPillarResponse struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Active      bool              `json:"active"`
	Links       map[string]string `json:"_links"`
}

// GetStrategyPillars godoc
// @Summary Get strategy pillars
// @Description Retrieves the list of strategy pillars for the current tenant
// @Tags meta-model
// @Accept json
// @Produce json
// @Param includeInactive query bool false "Include inactive pillars"
// @Success 200 {object} sharedAPI.CollectionResponse{data=[]StrategyPillarResponse}
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /meta-model/strategy-pillars [get]
func (h *StrategyPillarsHandlers) GetStrategyPillars(w http.ResponseWriter, r *http.Request) {
	includeInactive := r.URL.Query().Get("includeInactive") == "true"

	config, err := h.readModel.GetByTenantID(r.Context())
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve pillars")
		return
	}

	var pillars []readmodels.StrategyPillarDTO
	var version int
	if config != nil {
		version = config.Version
		if includeInactive {
			pillars = config.StrategyPillars
		} else {
			pillars = filterActivePillars(config.StrategyPillars)
		}
	}

	data := h.buildPillarResponses(pillars)
	links := h.hateoas.StrategyPillarsCollectionLinks()
	w.Header().Set("ETag", fmt.Sprintf(`"%d"`, version))
	sharedAPI.RespondCollection(w, http.StatusOK, data, links)
}

// GetStrategyPillarByID godoc
// @Summary Get strategy pillar by ID
// @Description Retrieves a single strategy pillar by its ID
// @Tags meta-model
// @Accept json
// @Produce json
// @Param id path string true "Pillar ID"
// @Success 200 {object} StrategyPillarResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /meta-model/strategy-pillars/{id} [get]
func (h *StrategyPillarsHandlers) GetStrategyPillarByID(w http.ResponseWriter, r *http.Request) {
	pillarID := chi.URLParam(r, "id")

	config, err := h.readModel.GetByTenantID(r.Context())
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve pillar")
		return
	}

	if config == nil {
		sharedAPI.RespondError(w, http.StatusNotFound, nil, "Pillar not found")
		return
	}

	h.respondWithPillarOrNotFound(w, config, pillarID)
}

func (h *StrategyPillarsHandlers) respondWithPillarOrNotFound(w http.ResponseWriter, config *readmodels.MetaModelConfigurationDTO, pillarID string) {
	pillar, found := findPillarByID(config.StrategyPillars, pillarID)
	if !found {
		sharedAPI.RespondError(w, http.StatusNotFound, nil, "Pillar not found")
		return
	}
	response := h.buildPillarResponse(pillar)
	w.Header().Set("ETag", fmt.Sprintf(`"%d"`, config.Version))
	sharedAPI.RespondJSON(w, http.StatusOK, response)
}

func findPillarByID(pillars []readmodels.StrategyPillarDTO, id string) (readmodels.StrategyPillarDTO, bool) {
	for _, pillar := range pillars {
		if pillar.ID == id {
			return pillar, true
		}
	}
	return readmodels.StrategyPillarDTO{}, false
}

// CreateStrategyPillar godoc
// @Summary Create a new strategy pillar
// @Description Creates a new strategy pillar for the current tenant
// @Tags meta-model
// @Accept json
// @Produce json
// @Param pillar body CreateStrategyPillarRequest true "Pillar to create"
// @Success 201 {object} StrategyPillarResponse
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 409 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /meta-model/strategy-pillars [post]
func (h *StrategyPillarsHandlers) CreateStrategyPillar(w http.ResponseWriter, r *http.Request) {
	authSession, err := h.sessionManager.LoadAuthenticatedSession(r.Context())
	if err != nil {
		sharedAPI.RespondError(w, http.StatusUnauthorized, err, "Authentication required")
		return
	}

	req, ok := sharedAPI.DecodeRequestOrFail[CreateStrategyPillarRequest](w, r)
	if !ok {
		return
	}

	result, ok := h.ensureConfigExists(w, r, authSession.UserEmail())
	if !ok {
		return
	}

	cmd := &commands.AddStrategyPillar{
		ConfigID:    result.config.ID,
		Name:        req.Name,
		Description: req.Description,
		ModifiedBy:  authSession.UserEmail(),
	}

	if _, err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		statusCode := sharedAPI.MapErrorToStatusCode(err, http.StatusBadRequest)
		sharedAPI.RespondError(w, statusCode, err, "Failed to create pillar")
		return
	}

	updatedConfig, err := h.readModel.GetByTenantID(r.Context())
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve created pillar")
		return
	}

	newPillar := updatedConfig.StrategyPillars[len(updatedConfig.StrategyPillars)-1]
	response := h.buildPillarResponse(newPillar)

	w.Header().Set("Location", "/api/v1/meta-model/strategy-pillars/"+newPillar.ID)
	w.Header().Set("ETag", fmt.Sprintf(`"%d"`, updatedConfig.Version))
	sharedAPI.RespondJSON(w, http.StatusCreated, response)
}

// UpdateStrategyPillar godoc
// @Summary Update a strategy pillar
// @Description Updates a strategy pillar's name and description
// @Tags meta-model
// @Accept json
// @Produce json
// @Param id path string true "Pillar ID"
// @Param If-Match header string true "ETag for optimistic locking"
// @Param pillar body UpdateStrategyPillarRequest true "Pillar updates"
// @Success 200 {object} StrategyPillarResponse
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 412 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /meta-model/strategy-pillars/{id} [put]
func (h *StrategyPillarsHandlers) UpdateStrategyPillar(w http.ResponseWriter, r *http.Request) {
	pillarID := chi.URLParam(r, "id")

	authSession, err := h.sessionManager.LoadAuthenticatedSession(r.Context())
	if err != nil {
		sharedAPI.RespondError(w, http.StatusUnauthorized, err, "Authentication required")
		return
	}

	expectedVersion, ok := h.requireETag(w, r)
	if !ok {
		return
	}

	req, ok := sharedAPI.DecodeRequestOrFail[UpdateStrategyPillarRequest](w, r)
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

	cmd := &commands.UpdateStrategyPillar{
		ConfigID:        config.ID,
		PillarID:        pillarID,
		Name:            req.Name,
		Description:     req.Description,
		ModifiedBy:      authSession.UserEmail(),
		ExpectedVersion: &expectedVersion,
	}

	if _, err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		statusCode := sharedAPI.MapErrorToStatusCode(err, http.StatusBadRequest)
		sharedAPI.RespondError(w, statusCode, err, "Failed to update pillar")
		return
	}

	h.respondWithUpdatedPillar(w, r, pillarID)
}

func (h *StrategyPillarsHandlers) respondWithUpdatedPillar(w http.ResponseWriter, r *http.Request, pillarID string) {
	updatedConfig, err := h.readModel.GetByTenantID(r.Context())
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve updated pillar")
		return
	}
	h.respondWithPillarOrNotFound(w, updatedConfig, pillarID)
}

// DeleteStrategyPillar godoc
// @Summary Delete a strategy pillar
// @Description Soft deletes a strategy pillar (sets it as inactive)
// @Tags meta-model
// @Accept json
// @Produce json
// @Param id path string true "Pillar ID"
// @Success 204
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 409 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /meta-model/strategy-pillars/{id} [delete]
func (h *StrategyPillarsHandlers) DeleteStrategyPillar(w http.ResponseWriter, r *http.Request) {
	pillarID := chi.URLParam(r, "id")

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

	cmd := &commands.RemoveStrategyPillar{
		ConfigID:   config.ID,
		PillarID:   pillarID,
		ModifiedBy: authSession.UserEmail(),
	}

	if _, err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		statusCode := sharedAPI.MapErrorToStatusCode(err, http.StatusBadRequest)
		sharedAPI.RespondError(w, statusCode, err, "Failed to delete pillar")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type PillarChangeRequest struct {
	Operation   string `json:"operation"`
	ID          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

type BatchUpdateStrategyPillarsRequest struct {
	Changes []PillarChangeRequest `json:"changes"`
}

type BatchUpdateStrategyPillarsResponse struct {
	Data  []StrategyPillarResponse `json:"data"`
	Links map[string]string        `json:"_links"`
}

// BatchUpdateStrategyPillars godoc
// @Summary Batch update strategy pillars
// @Description Atomically update multiple strategy pillars in a single transaction
// @Tags meta-model
// @Accept json
// @Produce json
// @Param If-Match header string true "ETag for optimistic locking"
// @Param changes body BatchUpdateStrategyPillarsRequest true "Pillar changes"
// @Success 200 {object} BatchUpdateStrategyPillarsResponse
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 412 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /meta-model/strategy-pillars [patch]
func (h *StrategyPillarsHandlers) BatchUpdateStrategyPillars(w http.ResponseWriter, r *http.Request) {
	authSession, err := h.sessionManager.LoadAuthenticatedSession(r.Context())
	if err != nil {
		sharedAPI.RespondError(w, http.StatusUnauthorized, err, "Authentication required")
		return
	}

	expectedVersion, ok := h.requireETag(w, r)
	if !ok {
		return
	}

	req, ok := sharedAPI.DecodeRequestOrFail[BatchUpdateStrategyPillarsRequest](w, r)
	if !ok {
		return
	}

	if len(req.Changes) == 0 {
		sharedAPI.RespondError(w, http.StatusBadRequest, nil, "No changes provided")
		return
	}

	result, ok := h.ensureConfigExists(w, r, authSession.UserEmail())
	if !ok {
		return
	}

	if err := h.dispatchBatchUpdate(r.Context(), w, req, result, expectedVersion, authSession.UserEmail()); err != nil {
		return
	}

	h.respondWithUpdatedPillars(w, r)
}

func (h *StrategyPillarsHandlers) dispatchBatchUpdate(ctx context.Context, w http.ResponseWriter, req BatchUpdateStrategyPillarsRequest, result *ensureConfigResult, expectedVersion int, userEmail string) error {
	changes, err := mapPillarChanges(req.Changes)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, nil, err.Error())
		return err
	}

	cmd := &commands.BatchUpdateStrategyPillars{
		ConfigID:   result.config.ID,
		Changes:    changes,
		ModifiedBy: userEmail,
	}
	if !result.wasCreated {
		cmd.ExpectedVersion = &expectedVersion
	}

	if _, err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		statusCode := sharedAPI.MapErrorToStatusCode(err, http.StatusBadRequest)
		sharedAPI.RespondError(w, statusCode, err, "Failed to update strategy pillars")
		return err
	}
	return nil
}

func (h *StrategyPillarsHandlers) respondWithUpdatedPillars(w http.ResponseWriter, r *http.Request) {
	updatedConfig, err := h.readModel.GetByTenantID(r.Context())
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve updated configuration")
		return
	}

	activePillars := filterActivePillars(updatedConfig.StrategyPillars)
	response := BatchUpdateStrategyPillarsResponse{
		Data:  h.buildPillarResponses(activePillars),
		Links: h.hateoas.StrategyPillarsCollectionLinks(),
	}

	w.Header().Set("ETag", fmt.Sprintf(`"%d"`, updatedConfig.Version))
	sharedAPI.RespondJSON(w, http.StatusOK, response)
}

func (h *StrategyPillarsHandlers) requireETag(w http.ResponseWriter, r *http.Request) (int, bool) {
	ifMatch := r.Header.Get("If-Match")
	if ifMatch == "" {
		sharedAPI.RespondError(w, http.StatusBadRequest, nil, "If-Match header required for optimistic locking")
		return 0, false
	}
	expectedVersion, err := parseETag(ifMatch)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid ETag format")
		return 0, false
	}
	return expectedVersion, true
}

func mapPillarChanges(requests []PillarChangeRequest) ([]commands.PillarChange, error) {
	changes := make([]commands.PillarChange, len(requests))
	for i, change := range requests {
		op, err := mapOperation(change.Operation)
		if err != nil {
			return nil, err
		}
		changes[i] = commands.PillarChange{
			Operation:   op,
			PillarID:    change.ID,
			Name:        change.Name,
			Description: change.Description,
		}
	}
	return changes, nil
}

func mapOperation(operation string) (commands.PillarOperation, error) {
	switch operation {
	case "add":
		return commands.PillarOperationAdd, nil
	case "update":
		return commands.PillarOperationUpdate, nil
	case "remove":
		return commands.PillarOperationRemove, nil
	default:
		return "", fmt.Errorf("invalid operation: %s", operation)
	}
}

type ensureConfigResult struct {
	config     *readmodels.MetaModelConfigurationDTO
	wasCreated bool
}

func (h *StrategyPillarsHandlers) ensureConfigExists(w http.ResponseWriter, r *http.Request, userEmail string) (*ensureConfigResult, bool) {
	config, err := h.readModel.GetByTenantID(r.Context())
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve configuration")
		return nil, false
	}
	if config != nil {
		return &ensureConfigResult{config: config, wasCreated: false}, true
	}

	tenantID, err := sharedctx.GetTenant(r.Context())
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to determine tenant")
		return nil, false
	}

	createCmd := &commands.CreateMetaModelConfiguration{
		TenantID:  tenantID.Value(),
		CreatedBy: userEmail,
	}
	if _, err := h.commandBus.Dispatch(r.Context(), createCmd); err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to initialize configuration")
		return nil, false
	}

	config, err = h.readModel.GetByTenantID(r.Context())
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve created configuration")
		return nil, false
	}

	return &ensureConfigResult{config: config, wasCreated: true}, true
}

func (h *StrategyPillarsHandlers) buildPillarResponse(pillar readmodels.StrategyPillarDTO) StrategyPillarResponse {
	return StrategyPillarResponse{
		ID:          pillar.ID,
		Name:        pillar.Name,
		Description: pillar.Description,
		Active:      pillar.Active,
		Links:       h.hateoas.StrategyPillarLinks(pillar.ID, pillar.Active),
	}
}

func (h *StrategyPillarsHandlers) buildPillarResponses(pillars []readmodels.StrategyPillarDTO) []StrategyPillarResponse {
	data := make([]StrategyPillarResponse, len(pillars))
	for i, pillar := range pillars {
		data[i] = h.buildPillarResponse(pillar)
	}
	return data
}

func filterActivePillars(pillars []readmodels.StrategyPillarDTO) []readmodels.StrategyPillarDTO {
	result := make([]readmodels.StrategyPillarDTO, 0)
	for _, p := range pillars {
		if p.Active {
			result = append(result, p)
		}
	}
	return result
}

func parseETag(etag string) (int, error) {
	if !isValidETagFormat(etag) {
		return 0, errors.New("invalid ETag format")
	}
	versionStr := etag[1 : len(etag)-1]
	return strconv.Atoi(versionStr)
}

func isValidETagFormat(etag string) bool {
	return len(etag) >= 3 && etag[0] == '"' && etag[len(etag)-1] == '"'
}
