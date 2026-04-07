package api

import (
	"errors"
	"fmt"
	"net/http"

	"encoding/json"
	"io"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/application/handlers"
	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/domain/services"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/cqrs"
)

type CapabilityHandlers struct {
	commandBus  cqrs.CommandBus
	readModel   *readmodels.CapabilityReadModel
	hateoas     *CapabilityMappingLinks
	impactQuery *handlers.DeleteImpactQuery
}

func NewCapabilityHandlers(
	commandBus cqrs.CommandBus,
	readModel *readmodels.CapabilityReadModel,
	hateoas *CapabilityMappingLinks,
	impactQuery *handlers.DeleteImpactQuery,
) *CapabilityHandlers {
	return &CapabilityHandlers{
		commandBus:  commandBus,
		readModel:   readModel,
		hateoas:     hateoas,
		impactQuery: impactQuery,
	}
}

func (h *CapabilityHandlers) addLinksToCapability(cap *readmodels.CapabilityDTO, actor sharedctx.Actor) {
	cap.Links = h.hateoas.CapabilityLinksForActor(cap.ID, cap.ParentID, actor)
	for i := range cap.Experts {
		cap.Experts[i].Links = h.hateoas.CapabilityExpertLinksForActor(sharedAPI.ExpertParams{
			ResourcePath: "/capabilities/" + cap.ID,
			ExpertName:   cap.Experts[i].Name,
			ExpertRole:   cap.Experts[i].Role,
			ContactInfo:  cap.Experts[i].Contact,
		}, actor)
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

	actor, _ := sharedctx.GetActor(r.Context())
	h.addLinksToCapability(capability, actor)
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

	actor, _ := sharedctx.GetActor(r.Context())
	for i := range capabilities {
		h.addLinksToCapability(&capabilities[i], actor)
	}

	links := sharedAPI.Links{
		"self": sharedAPI.NewLink(sharedAPI.BuildLink("/capabilities"), "GET"),
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

	actor, _ := sharedctx.GetActor(r.Context())
	h.addLinksToCapability(capability, actor)
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

	actor, _ := sharedctx.GetActor(r.Context())
	for i := range children {
		h.addLinksToCapability(&children[i], actor)
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

	actor, _ := sharedctx.GetActor(r.Context())
	h.addLinksToCapability(capability, actor)
	sharedAPI.RespondJSON(w, http.StatusOK, capability)
}

type DeleteCapabilityRequest struct {
	Cascade                     bool `json:"cascade"`
	DeleteRealisingApplications bool `json:"deleteRealisingApplications"`
}

// DeleteCapability godoc
// @Summary Delete a capability with optional cascade
// @Description Deletes a business capability. If the capability has descendants, cascade:true must be set in the request body. Optionally deletes realizations by setting deleteRealisingApplications:true.
// @Tags capabilities
// @Accept json
// @Param id path string true "Capability ID"
// @Param body body DeleteCapabilityRequest false "Cascade options"
// @Success 204 "No Content"
// @Failure 400 {object} sharedAPI.ErrorResponse
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

	req := h.parseDeleteBody(r)

	cmd := &commands.CascadeDeleteCapability{
		ID:                          id,
		Cascade:                     req.Cascade,
		DeleteRealisingApplications: req.DeleteRealisingApplications,
	}

	if _, err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		if errors.Is(err, services.ErrCascadeRequiredForChildCapabilities) {
			sharedAPI.RespondErrorWithLinks(w, sharedAPI.ErrorWithLinksParams{
				StatusCode: http.StatusConflict,
				Err:        err,
				Message:    "Capability has descendants. Set cascade:true to confirm cascade deletion.",
				Links: map[string]sharedAPI.Link{
					"x-delete-impact": sharedAPI.NewLink(sharedAPI.BuildLink(sharedAPI.ResourcePath("/capabilities/"+id+"/delete-impact")), "GET"),
				},
			})
			return
		}
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to delete capability")
		return
	}

	sharedAPI.RespondDeleted(w)
}

func (h *CapabilityHandlers) parseDeleteBody(r *http.Request) DeleteCapabilityRequest {
	var req DeleteCapabilityRequest
	if r.Body == nil {
		return req
	}
	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		return req
	}
	_ = json.Unmarshal(body, &req)
	return req
}

// GetDeleteImpact godoc
// @Summary Get delete impact analysis for a capability
// @Description Returns all capabilities and realizations that would be affected by deleting this capability and all descendants.
// @Tags capabilities
// @Produce json
// @Param id path string true "Capability ID"
// @Success 200 {object} DeleteImpactResponse
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /capabilities/{id}/delete-impact [get]
func (h *CapabilityHandlers) GetDeleteImpact(w http.ResponseWriter, r *http.Request) {
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

	impact, err := h.impactQuery.Execute(r.Context(), id)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to compute delete impact")
		return
	}

	actor, _ := sharedctx.GetActor(r.Context())

	affectedCaps := make([]AffectedCapabilityDTO, 0, len(impact.AffectedCapabilities))
	for _, capID := range impact.AffectedCapabilities {
		dto := AffectedCapabilityDTO{
			ID: capID,
			Links: sharedAPI.Links{
				"self": h.hateoas.Get("/capabilities/" + capID),
			},
		}
		if cap, err := h.readModel.GetByID(r.Context(), capID); err == nil && cap != nil {
			dto.Name = cap.Name
			dto.Level = cap.Level
			dto.ParentID = cap.ParentID
		}
		affectedCaps = append(affectedCaps, dto)
	}

	links := sharedAPI.Links{
		"self":         sharedAPI.NewLink(sharedAPI.BuildLink(sharedAPI.ResourcePath("/capabilities/"+id+"/delete-impact")), "GET"),
		"x-capability": sharedAPI.NewLink(sharedAPI.BuildLink(sharedAPI.ResourcePath("/capabilities/"+id)), "GET"),
	}
	if actor.CanDelete("capabilities") {
		links["x-confirm-delete"] = sharedAPI.NewLink(sharedAPI.BuildLink(sharedAPI.ResourcePath("/capabilities/"+id)), "DELETE")
	}

	response := DeleteImpactResponse{
		CapabilityID:                       capability.ID,
		CapabilityName:                     capability.Name,
		HasDescendants:                     impact.HasDescendants,
		AffectedCapabilities:               affectedCaps,
		RealizationsOnDeletedCapabilities:  impact.RealizationsOnDeletedCapabilities,
		RealizationsOnRetainedCapabilities: impact.RealizationsOnRetainedCapabilities,
		Links:                              links,
	}

	sharedAPI.RespondJSON(w, http.StatusOK, response)
}

type AffectedCapabilityDTO struct {
	ID       string          `json:"id"`
	Name     string          `json:"name,omitempty"`
	Level    string          `json:"level,omitempty"`
	ParentID string          `json:"parentId,omitempty"`
	Links    sharedAPI.Links `json:"_links,omitempty"`
}

type DeleteImpactResponse struct {
	CapabilityID                       string                      `json:"capabilityId"`
	CapabilityName                     string                      `json:"capabilityName"`
	HasDescendants                     bool                        `json:"hasDescendants"`
	AffectedCapabilities               []AffectedCapabilityDTO     `json:"affectedCapabilities"`
	RealizationsOnDeletedCapabilities  []readmodels.RealizationDTO `json:"realizationsOnDeletedCapabilities"`
	RealizationsOnRetainedCapabilities []readmodels.RealizationDTO `json:"realizationsOnRetainedCapabilities"`
	Links                              sharedAPI.Links             `json:"_links,omitempty"`
}
