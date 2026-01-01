package api

import (
	"errors"
	"fmt"
	"net/http"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/application/handlers"
	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/cqrs"
)

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

// CreateCapability godoc
// @Summary Create a new business capability
// @Description Creates a new business capability in the capability map
// @Tags capabilities
// @Accept json
// @Produce json
// @Param capability body CreateCapabilityRequest true "Capability data"
// @Success 201 {object} easi_backend_internal_capabilitymapping_application_readmodels.CapabilityDTO
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /capabilities [post]
func (h *CapabilityHandlers) CreateCapability(w http.ResponseWriter, r *http.Request) {
	req, ok := sharedAPI.DecodeRequestOrFail[CreateCapabilityRequest](w, r)
	if !ok {
		return
	}

	if err := h.validateCapabilityRequest(req); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "")
		return
	}

	cmd := &commands.CreateCapability{
		Name:        req.Name,
		Description: req.Description,
		ParentID:    req.ParentID,
		Level:       req.Level,
	}

	result, err := h.commandBus.Dispatch(r.Context(), cmd)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to create capability")
		return
	}

	h.handleCreateCapabilityResponse(w, r, result.CreatedID)
}

func (h *CapabilityHandlers) validateCapabilityRequest(req CreateCapabilityRequest) error {
	if _, err := valueobjects.NewCapabilityName(req.Name); err != nil {
		return err
	}

	if _, err := valueobjects.NewCapabilityLevel(req.Level); err != nil {
		return err
	}

	if req.ParentID != "" {
		if _, err := valueobjects.NewCapabilityIDFromString(req.ParentID); err != nil {
			return fmt.Errorf("invalid parent ID: %w", err)
		}
	}

	return nil
}

func (h *CapabilityHandlers) handleCreateCapabilityResponse(w http.ResponseWriter, r *http.Request, capabilityID string) {
	capability, err := h.readModel.GetByID(r.Context(), capabilityID)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve created capability")
		return
	}

	location := sharedAPI.BuildResourceLink(sharedAPI.ResourcePath("/capabilities"), sharedAPI.ResourceID(capabilityID))

	if capability == nil {
		sharedAPI.RespondCreated(w, location, map[string]string{
			"id":      capabilityID,
			"message": "Capability created, processing",
		})
		return
	}

	capability.Links = h.hateoas.CapabilityLinks(capability.ID, capability.ParentID)
	sharedAPI.RespondCreated(w, location, capability)
}

// GetAllCapabilities godoc
// @Summary Get all business capabilities
// @Description Retrieves all business capabilities in the capability map
// @Tags capabilities
// @Produce json
// @Success 200 {object} easi_backend_internal_shared_api.CollectionResponse{data=[]easi_backend_internal_capabilitymapping_application_readmodels.CapabilityDTO}
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /capabilities [get]
func (h *CapabilityHandlers) GetAllCapabilities(w http.ResponseWriter, r *http.Request) {
	capabilities, err := h.readModel.GetAll(r.Context())
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve capabilities")
		return
	}

	for i := range capabilities {
		capabilities[i].Links = h.hateoas.CapabilityLinks(capabilities[i].ID, capabilities[i].ParentID)
	}

	links := map[string]string{
		"self": "/api/v1/capabilities",
	}

	sharedAPI.RespondCollection(w, http.StatusOK, capabilities, links)
}

// GetCapabilityByID godoc
// @Summary Get a capability by ID
// @Description Retrieves a specific business capability by its ID
// @Tags capabilities
// @Produce json
// @Param id path string true "Capability ID"
// @Success 200 {object} easi_backend_internal_capabilitymapping_application_readmodels.CapabilityDTO
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /capabilities/{id} [get]
func (h *CapabilityHandlers) GetCapabilityByID(w http.ResponseWriter, r *http.Request) {
	id := sharedAPI.GetPathParam(r, "id")

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

// GetCapabilityChildren godoc
// @Summary Get child capabilities
// @Description Retrieves all child capabilities of a specific capability
// @Tags capabilities
// @Produce json
// @Param id path string true "Capability ID"
// @Success 200 {object} easi_backend_internal_shared_api.CollectionResponse{data=[]easi_backend_internal_capabilitymapping_application_readmodels.CapabilityDTO}
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /capabilities/{id}/children [get]
func (h *CapabilityHandlers) GetCapabilityChildren(w http.ResponseWriter, r *http.Request) {
	id := sharedAPI.GetPathParam(r, "id")

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

	links := sharedAPI.NewResourceLinks().
		Self(sharedAPI.ResourcePath("/capabilities/"+id+"/children")).
		Related(sharedAPI.LinkRelation("parent"), sharedAPI.ResourcePath("/capabilities"), sharedAPI.ResourceID(id)).
		Build()

	sharedAPI.RespondCollection(w, http.StatusOK, children, links)
}

// UpdateCapability godoc
// @Summary Update a capability
// @Description Updates the name and description of a business capability
// @Tags capabilities
// @Accept json
// @Produce json
// @Param id path string true "Capability ID"
// @Param capability body UpdateCapabilityRequest true "Updated capability data"
// @Success 200 {object} easi_backend_internal_capabilitymapping_application_readmodels.CapabilityDTO
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /capabilities/{id} [put]
func (h *CapabilityHandlers) UpdateCapability(w http.ResponseWriter, r *http.Request) {
	id := sharedAPI.GetPathParam(r, "id")

	req, ok := sharedAPI.DecodeRequestOrFail[UpdateCapabilityRequest](w, r)
	if !ok {
		return
	}

	if _, err := valueobjects.NewCapabilityName(req.Name); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "")
		return
	}

	cmd := &commands.UpdateCapability{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
	}

	if _, err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
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

// DeleteCapability godoc
// @Summary Delete a capability
// @Description Deletes a business capability. Cannot delete capabilities that have children.
// @Tags capabilities
// @Param id path string true "Capability ID"
// @Success 204 "No Content"
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 409 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /capabilities/{id} [delete]
func (h *CapabilityHandlers) DeleteCapability(w http.ResponseWriter, r *http.Request) {
	id := sharedAPI.GetPathParam(r, "id")

	capability, err := h.readModel.GetByID(r.Context(), id)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve capability")
		return
	}

	if capability == nil {
		sharedAPI.RespondError(w, http.StatusNotFound, nil, "Capability not found")
		return
	}

	cmd := &commands.DeleteCapability{
		ID: id,
	}

	if _, err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		if errors.Is(err, handlers.ErrCapabilityHasChildren) {
			sharedAPI.RespondError(w, http.StatusConflict, err, "Cannot delete capability with children. Delete child capabilities first.")
			return
		}
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to delete capability")
		return
	}

	sharedAPI.RespondDeleted(w)
}
