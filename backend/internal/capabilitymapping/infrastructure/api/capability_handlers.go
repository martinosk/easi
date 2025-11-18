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

type ErrorResponse = sharedAPI.ErrorResponse

type CapabilityHandlers struct {
	commandBus cqrs.CommandBus
	readModel  *readmodels.CapabilityReadModel
	hateoas    *sharedAPI.HATEOASLinks
}

func NewCapabilityHandlers(
	commandBus cqrs.CommandBus,
	readModel *readmodels.CapabilityReadModel,
	hateoas *sharedAPI.HATEOASLinks,
) *CapabilityHandlers {
	return &CapabilityHandlers{
		commandBus: commandBus,
		readModel:  readModel,
		hateoas:    hateoas,
	}
}

type CreateCapabilityRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	ParentID    string `json:"parentId,omitempty"`
	Level       string `json:"level"`
}

type UpdateCapabilityRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

func (h *CapabilityHandlers) CreateCapability(w http.ResponseWriter, r *http.Request) {
	var req CreateCapabilityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	_, err := valueobjects.NewCapabilityName(req.Name)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "")
		return
	}

	_, err = valueobjects.NewCapabilityLevel(req.Level)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "")
		return
	}

	if req.ParentID != "" {
		_, err = valueobjects.NewCapabilityIDFromString(req.ParentID)
		if err != nil {
			sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid parent ID")
			return
		}
	}

	cmd := &commands.CreateCapability{
		Name:        req.Name,
		Description: req.Description,
		ParentID:    req.ParentID,
		Level:       req.Level,
	}

	if err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to create capability")
		return
	}

	capability, err := h.readModel.GetByID(r.Context(), cmd.ID)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve created capability")
		return
	}

	if capability == nil {
		location := fmt.Sprintf("/api/capabilities/%s", cmd.ID)
		w.Header().Set("Location", location)
		sharedAPI.RespondJSON(w, http.StatusCreated, map[string]string{
			"id":      cmd.ID,
			"message": "Capability created, processing",
		})
		return
	}

	capability.Links = h.hateoas.CapabilityLinks(capability.ID, capability.ParentID)

	location := fmt.Sprintf("/api/capabilities/%s", capability.ID)
	sharedAPI.RespondCreated(w, location, capability, nil)
}

func (h *CapabilityHandlers) GetAllCapabilities(w http.ResponseWriter, r *http.Request) {
	capabilities, err := h.readModel.GetAll(r.Context())
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve capabilities")
		return
	}

	for i := range capabilities {
		capabilities[i].Links = h.hateoas.CapabilityLinks(capabilities[i].ID, capabilities[i].ParentID)
	}

	sharedAPI.RespondJSON(w, http.StatusOK, capabilities)
}

func (h *CapabilityHandlers) GetCapabilityByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	capability, err := h.readModel.GetByID(r.Context(), id)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve capability")
		return
	}

	if capability == nil {
		sharedAPI.RespondError(w, http.StatusNotFound, nil, "Capability not found")
		return
	}

	capability.Links = h.hateoas.CapabilityLinks(capability.ID, capability.ParentID)

	sharedAPI.RespondJSON(w, http.StatusOK, capability)
}

func (h *CapabilityHandlers) GetCapabilityChildren(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	capability, err := h.readModel.GetByID(r.Context(), id)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve capability")
		return
	}

	if capability == nil {
		sharedAPI.RespondError(w, http.StatusNotFound, nil, "Capability not found")
		return
	}

	children, err := h.readModel.GetChildren(r.Context(), id)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve children")
		return
	}

	for i := range children {
		children[i].Links = h.hateoas.CapabilityLinks(children[i].ID, children[i].ParentID)
	}

	sharedAPI.RespondJSON(w, http.StatusOK, children)
}

func (h *CapabilityHandlers) UpdateCapability(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req UpdateCapabilityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	_, err := valueobjects.NewCapabilityName(req.Name)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "")
		return
	}

	cmd := &commands.UpdateCapability{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
	}

	if err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to update capability")
		return
	}

	capability, err := h.readModel.GetByID(r.Context(), id)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve updated capability")
		return
	}

	if capability == nil {
		sharedAPI.RespondError(w, http.StatusNotFound, nil, "Capability not found")
		return
	}

	capability.Links = h.hateoas.CapabilityLinks(capability.ID, capability.ParentID)

	sharedAPI.RespondJSON(w, http.StatusOK, capability)
}
