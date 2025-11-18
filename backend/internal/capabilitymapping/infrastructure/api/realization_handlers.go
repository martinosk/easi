package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/application/handlers"
	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	"easi/backend/internal/capabilitymapping/infrastructure/repositories"
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

	if err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		if errors.Is(err, handlers.ErrCapabilityNotFoundForRealization) {
			sharedAPI.RespondError(w, http.StatusNotFound, err, "Capability not found")
			return
		}
		if errors.Is(err, handlers.ErrComponentNotFound) {
			sharedAPI.RespondError(w, http.StatusNotFound, err, "Component not found")
			return
		}
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to link system to capability")
		return
	}

	realization, err := h.readModel.GetByID(r.Context(), cmd.ID)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve created realization")
		return
	}

	if realization == nil {
		location := fmt.Sprintf("/api/capability-realizations/%s", cmd.ID)
		w.Header().Set("Location", location)
		sharedAPI.RespondJSON(w, http.StatusCreated, map[string]string{
			"id":      cmd.ID,
			"message": "Realization created, processing",
		})
		return
	}

	realization.Links = h.hateoas.RealizationLinks(realization.ID, realization.CapabilityID, realization.ComponentID)

	location := fmt.Sprintf("/api/capability-realizations/%s", realization.ID)
	sharedAPI.RespondCreated(w, location, realization, nil)
}

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

	sharedAPI.RespondJSON(w, http.StatusOK, realizations)
}

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

	sharedAPI.RespondJSON(w, http.StatusOK, realizations)
}

type UpdateRealizationRequest struct {
	RealizationLevel string `json:"realizationLevel"`
	Notes            string `json:"notes,omitempty"`
}

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

	if err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		if errors.Is(err, repositories.ErrRealizationNotFound) {
			sharedAPI.RespondError(w, http.StatusNotFound, err, "Realization not found")
			return
		}
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to update realization")
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

func (h *RealizationHandlers) DeleteRealization(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	cmd := &commands.DeleteSystemRealization{
		ID: id,
	}

	if err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		if errors.Is(err, repositories.ErrRealizationNotFound) {
			sharedAPI.RespondError(w, http.StatusNotFound, err, "Realization not found")
			return
		}
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to delete realization")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
