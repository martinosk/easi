package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/cqrs"

	"github.com/go-chi/chi/v5"
)

type RealizationHandlers struct {
	commandBus cqrs.CommandBus
	readModel  *readmodels.RealizationReadModel
	hateoas    *sharedAPI.HATEOASLinks
}

func NewRealizationHandlers(
	commandBus cqrs.CommandBus,
	readModel *readmodels.RealizationReadModel,
	hateoas *sharedAPI.HATEOASLinks,
) *RealizationHandlers {
	return &RealizationHandlers{
		commandBus: commandBus,
		readModel:  readModel,
		hateoas:    hateoas,
	}
}

type LinkSystemRequest struct {
	ComponentID      string `json:"componentId"`
	RealizationLevel string `json:"realizationLevel"`
	Notes            string `json:"notes,omitempty"`
}

// LinkSystemToCapability godoc
// @Summary Link a system to a capability
// @Description Associates an application component with a capability to indicate realization
// @Tags capabilities
// @Accept json
// @Produce json
// @Param id path string true "Capability ID"
// @Param system body LinkSystemRequest true "System link data"
// @Success 201 {object} easi_backend_internal_capabilitymapping_application_readmodels.RealizationDTO
// @Failure 400 {object} easi_backend_internal_shared_api.ErrorResponse
// @Failure 404 {object} easi_backend_internal_shared_api.ErrorResponse
// @Failure 500 {object} easi_backend_internal_shared_api.ErrorResponse
// @Router /capabilities/{id}/systems [post]
func (h *RealizationHandlers) LinkSystemToCapability(w http.ResponseWriter, r *http.Request) {
	capabilityID := chi.URLParam(r, "id")

	var req LinkSystemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	_, err := valueobjects.NewCapabilityIDFromString(capabilityID)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid capability ID")
		return
	}

	_, err = valueobjects.NewComponentIDFromString(req.ComponentID)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid component ID")
		return
	}

	_, err = valueobjects.NewRealizationLevel(req.RealizationLevel)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "")
		return
	}

	cmd := &commands.LinkSystemToCapability{
		CapabilityID:     capabilityID,
		ComponentID:      req.ComponentID,
		RealizationLevel: req.RealizationLevel,
		Notes:            req.Notes,
	}

	result, err := h.commandBus.Dispatch(r.Context(), cmd)
	if err != nil {
		sharedAPI.HandleError(w, err)
		return
	}

	realization, err := h.readModel.GetByID(r.Context(), result.CreatedID)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve created realization")
		return
	}

	if realization == nil {
		location := fmt.Sprintf("/api/capability-realizations/%s", result.CreatedID)
		w.Header().Set("Location", location)
		sharedAPI.RespondJSON(w, http.StatusCreated, map[string]string{
			"id":      result.CreatedID,
			"message": "Realization created, processing",
		})
		return
	}

	realization.Links = h.hateoas.RealizationLinks(realization.ID, realization.CapabilityID, realization.ComponentID)

	location := fmt.Sprintf("/api/capability-realizations/%s", realization.ID)
	w.Header().Set("Location", location)
	sharedAPI.RespondJSON(w, http.StatusCreated, realization)
}

// GetSystemsByCapability godoc
// @Summary Get systems linked to a capability
// @Description Retrieves all application components that realize a specific capability
// @Tags capabilities
// @Produce json
// @Param id path string true "Capability ID"
// @Success 200 {object} easi_backend_internal_shared_api.CollectionResponse{data=[]easi_backend_internal_capabilitymapping_application_readmodels.RealizationDTO}
// @Failure 500 {object} easi_backend_internal_shared_api.ErrorResponse
// @Router /capabilities/{id}/systems [get]
func (h *RealizationHandlers) GetSystemsByCapability(w http.ResponseWriter, r *http.Request) {
	capabilityID := chi.URLParam(r, "id")

	realizations, err := h.readModel.GetByCapabilityID(r.Context(), capabilityID)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve systems")
		return
	}

	for i := range realizations {
		realizations[i].Links = h.hateoas.RealizationLinks(realizations[i].ID, realizations[i].CapabilityID, realizations[i].ComponentID)
	}

	links := map[string]string{
		"self":       "/api/v1/capabilities/" + capabilityID + "/systems",
		"capability": "/api/v1/capabilities/" + capabilityID,
	}

	sharedAPI.RespondCollection(w, http.StatusOK, realizations, links)
}

// GetCapabilitiesByComponent godoc
// @Summary Get capabilities realized by a component
// @Description Retrieves all capabilities that are realized by a specific application component
// @Tags capability-realizations
// @Produce json
// @Param componentId path string true "Component ID"
// @Success 200 {object} easi_backend_internal_shared_api.CollectionResponse{data=[]easi_backend_internal_capabilitymapping_application_readmodels.RealizationDTO}
// @Failure 500 {object} easi_backend_internal_shared_api.ErrorResponse
// @Router /capability-realizations/by-component/{componentId} [get]
func (h *RealizationHandlers) GetCapabilitiesByComponent(w http.ResponseWriter, r *http.Request) {
	componentID := chi.URLParam(r, "componentId")

	realizations, err := h.readModel.GetByComponentID(r.Context(), componentID)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve capabilities")
		return
	}

	for i := range realizations {
		realizations[i].Links = h.hateoas.RealizationLinks(realizations[i].ID, realizations[i].CapabilityID, realizations[i].ComponentID)
	}

	links := map[string]string{
		"self":      "/api/v1/capability-realizations/by-component/" + componentID,
		"component": "/api/v1/components/" + componentID,
	}

	sharedAPI.RespondCollection(w, http.StatusOK, realizations, links)
}

type UpdateRealizationRequest struct {
	RealizationLevel string `json:"realizationLevel"`
	Notes            string `json:"notes,omitempty"`
}

// UpdateRealization godoc
// @Summary Update a system realization
// @Description Updates the realization level and notes for a system-capability link
// @Tags capability-realizations
// @Accept json
// @Produce json
// @Param id path string true "Realization ID"
// @Param realization body UpdateRealizationRequest true "Updated realization data"
// @Success 200 {object} easi_backend_internal_capabilitymapping_application_readmodels.RealizationDTO
// @Failure 400 {object} easi_backend_internal_shared_api.ErrorResponse
// @Failure 404 {object} easi_backend_internal_shared_api.ErrorResponse
// @Failure 500 {object} easi_backend_internal_shared_api.ErrorResponse
// @Router /capability-realizations/{id} [put]
func (h *RealizationHandlers) UpdateRealization(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req UpdateRealizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	_, err := valueobjects.NewRealizationLevel(req.RealizationLevel)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "")
		return
	}

	cmd := &commands.UpdateSystemRealization{
		ID:               id,
		RealizationLevel: req.RealizationLevel,
		Notes:            req.Notes,
	}

	if _, err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		sharedAPI.HandleError(w, err)
		return
	}

	realization, err := h.readModel.GetByID(r.Context(), id)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve updated realization")
		return
	}

	if realization == nil {
		sharedAPI.RespondError(w, http.StatusNotFound, nil, "Realization not found")
		return
	}

	realization.Links = h.hateoas.RealizationLinks(realization.ID, realization.CapabilityID, realization.ComponentID)

	sharedAPI.RespondJSON(w, http.StatusOK, realization)
}

// DeleteRealization godoc
// @Summary Delete a system realization
// @Description Removes the link between a system and a capability
// @Tags capability-realizations
// @Param id path string true "Realization ID"
// @Success 204 "No Content"
// @Failure 404 {object} easi_backend_internal_shared_api.ErrorResponse
// @Failure 500 {object} easi_backend_internal_shared_api.ErrorResponse
// @Router /capability-realizations/{id} [delete]
func (h *RealizationHandlers) DeleteRealization(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	cmd := &commands.DeleteSystemRealization{
		ID: id,
	}

	result, err := h.commandBus.Dispatch(r.Context(), cmd)
	sharedAPI.HandleCommandResult(w, result, err, func(_ string) {
		w.WriteHeader(http.StatusNoContent)
	})
}
