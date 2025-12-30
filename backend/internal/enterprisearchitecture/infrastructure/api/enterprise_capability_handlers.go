package api

import (
	"net/http"

	"easi/backend/internal/auth/infrastructure/session"
	"easi/backend/internal/enterprisearchitecture/application/commands"
	"easi/backend/internal/enterprisearchitecture/application/readmodels"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/cqrs"
)

type EnterpriseCapabilityReadModels struct {
	Capability *readmodels.EnterpriseCapabilityReadModel
	Link       *readmodels.EnterpriseCapabilityLinkReadModel
	Importance *readmodels.EnterpriseStrategicImportanceReadModel
}

type EnterpriseCapabilityHandlers struct {
	commandBus     cqrs.CommandBus
	readModels     *EnterpriseCapabilityReadModels
	sessionManager *session.SessionManager
	hateoas        *sharedAPI.HATEOASLinks
}

func NewEnterpriseCapabilityHandlers(
	commandBus cqrs.CommandBus,
	readModels *EnterpriseCapabilityReadModels,
	sessionManager *session.SessionManager,
) *EnterpriseCapabilityHandlers {
	return &EnterpriseCapabilityHandlers{
		commandBus:     commandBus,
		readModels:     readModels,
		sessionManager: sessionManager,
		hateoas:        sharedAPI.NewHATEOASLinks(""),
	}
}

type CreateEnterpriseCapabilityRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Category    string `json:"category,omitempty"`
}

type UpdateEnterpriseCapabilityRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Category    string `json:"category,omitempty"`
}

type LinkCapabilityRequest struct {
	DomainCapabilityID string `json:"domainCapabilityId"`
}

type SetStrategicImportanceRequest struct {
	PillarID   string `json:"pillarId"`
	PillarName string `json:"pillarName"`
	Importance int    `json:"importance"`
	Rationale  string `json:"rationale,omitempty"`
}

type UpdateStrategicImportanceRequest struct {
	Importance int    `json:"importance"`
	Rationale  string `json:"rationale,omitempty"`
}

type DomainCapabilityEnterpriseResponse struct {
	Linked                   bool              `json:"linked"`
	EnterpriseCapabilityID   *string           `json:"enterpriseCapabilityId"`
	EnterpriseCapabilityName *string           `json:"enterpriseCapabilityName,omitempty"`
	LinkID                   *string           `json:"linkId,omitempty"`
	Links                    map[string]string `json:"_links,omitempty"`
}

// CreateEnterpriseCapability godoc
// @Summary Create a new enterprise capability
// @Description Creates a new enterprise capability for grouping domain capabilities
// @Tags enterprise-capabilities
// @Accept json
// @Produce json
// @Param capability body CreateEnterpriseCapabilityRequest true "Enterprise capability data"
// @Success 201 {object} easi_backend_internal_enterprisearchitecture_application_readmodels.EnterpriseCapabilityDTO
// @Failure 400 {object} easi_backend_internal_shared_api.ErrorResponse
// @Failure 409 {object} easi_backend_internal_shared_api.ErrorResponse
// @Failure 500 {object} easi_backend_internal_shared_api.ErrorResponse
// @Router /enterprise-capabilities [post]
func (h *EnterpriseCapabilityHandlers) CreateEnterpriseCapability(w http.ResponseWriter, r *http.Request) {
	req, ok := sharedAPI.DecodeRequestOrFail[CreateEnterpriseCapabilityRequest](w, r)
	if !ok {
		return
	}

	cmd := &commands.CreateEnterpriseCapability{
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
	}

	result, err := h.commandBus.Dispatch(r.Context(), cmd)
	sharedAPI.HandleCommandResult(w, result, err, func(createdID string) {
		h.respondWithCapability(w, r, createdID, http.StatusCreated)
	})
}

// GetAllEnterpriseCapabilities godoc
// @Summary Get all enterprise capabilities
// @Description Retrieves all active enterprise capabilities with optional pagination
// @Tags enterprise-capabilities
// @Produce json
// @Param limit query int false "Maximum number of results (default 20, max 100)"
// @Param cursor query string false "Pagination cursor for next page"
// @Success 200 {object} easi_backend_internal_shared_api.CollectionResponse{data=[]easi_backend_internal_enterprisearchitecture_application_readmodels.EnterpriseCapabilityDTO}
// @Failure 500 {object} easi_backend_internal_shared_api.ErrorResponse
// @Router /enterprise-capabilities [get]
func (h *EnterpriseCapabilityHandlers) GetAllEnterpriseCapabilities(w http.ResponseWriter, r *http.Request) {
	capabilities, err := h.readModels.Capability.GetAll(r.Context())
	if err != nil {
		sharedAPI.HandleError(w, err)
		return
	}

	for i := range capabilities {
		capabilities[i].Links = h.hateoas.EnterpriseCapabilityLinks(capabilities[i].ID)
	}

	sharedAPI.RespondCollection(w, http.StatusOK, capabilities, h.hateoas.EnterpriseCapabilityCollectionLinks())
}

// GetEnterpriseCapabilityByID godoc
// @Summary Get an enterprise capability by ID
// @Description Retrieves a specific enterprise capability by its ID
// @Tags enterprise-capabilities
// @Produce json
// @Param id path string true "Enterprise capability ID"
// @Success 200 {object} easi_backend_internal_enterprisearchitecture_application_readmodels.EnterpriseCapabilityDTO
// @Failure 404 {object} easi_backend_internal_shared_api.ErrorResponse
// @Failure 500 {object} easi_backend_internal_shared_api.ErrorResponse
// @Router /enterprise-capabilities/{id} [get]
func (h *EnterpriseCapabilityHandlers) GetEnterpriseCapabilityByID(w http.ResponseWriter, r *http.Request) {
	capability := h.getCapabilityOrNotFound(w, r, sharedAPI.GetPathParam(r, "id"))
	if capability == nil {
		return
	}
	capability.Links = h.hateoas.EnterpriseCapabilityLinks(capability.ID)
	sharedAPI.RespondJSON(w, http.StatusOK, capability)
}

// UpdateEnterpriseCapability godoc
// @Summary Update an enterprise capability
// @Description Updates the name, description, and category of an enterprise capability
// @Tags enterprise-capabilities
// @Accept json
// @Produce json
// @Param id path string true "Enterprise capability ID"
// @Param capability body UpdateEnterpriseCapabilityRequest true "Updated capability data"
// @Success 200 {object} easi_backend_internal_enterprisearchitecture_application_readmodels.EnterpriseCapabilityDTO
// @Failure 400 {object} easi_backend_internal_shared_api.ErrorResponse
// @Failure 404 {object} easi_backend_internal_shared_api.ErrorResponse
// @Failure 409 {object} easi_backend_internal_shared_api.ErrorResponse
// @Failure 500 {object} easi_backend_internal_shared_api.ErrorResponse
// @Router /enterprise-capabilities/{id} [put]
func (h *EnterpriseCapabilityHandlers) UpdateEnterpriseCapability(w http.ResponseWriter, r *http.Request) {
	id := sharedAPI.GetPathParam(r, "id")

	req, ok := sharedAPI.DecodeRequestOrFail[UpdateEnterpriseCapabilityRequest](w, r)
	if !ok {
		return
	}

	cmd := &commands.UpdateEnterpriseCapability{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
	}

	result, err := h.commandBus.Dispatch(r.Context(), cmd)
	sharedAPI.HandleCommandResult(w, result, err, func(_ string) {
		h.respondWithCapability(w, r, id, http.StatusOK)
	})
}

// DeleteEnterpriseCapability godoc
// @Summary Delete an enterprise capability
// @Description Soft deletes an enterprise capability (marks as inactive)
// @Tags enterprise-capabilities
// @Param id path string true "Enterprise capability ID"
// @Success 204 "No Content"
// @Failure 404 {object} easi_backend_internal_shared_api.ErrorResponse
// @Failure 500 {object} easi_backend_internal_shared_api.ErrorResponse
// @Router /enterprise-capabilities/{id} [delete]
func (h *EnterpriseCapabilityHandlers) DeleteEnterpriseCapability(w http.ResponseWriter, r *http.Request) {
	id := sharedAPI.GetPathParam(r, "id")
	if h.getCapabilityOrNotFound(w, r, id) == nil {
		return
	}

	result, err := h.commandBus.Dispatch(r.Context(), &commands.DeleteEnterpriseCapability{ID: id})
	sharedAPI.HandleCommandResult(w, result, err, func(_ string) {
		sharedAPI.RespondDeleted(w)
	})
}

// GetLinkedCapabilities godoc
// @Summary Get linked domain capabilities
// @Description Retrieves all domain capabilities linked to an enterprise capability
// @Tags enterprise-capabilities
// @Produce json
// @Param id path string true "Enterprise capability ID"
// @Success 200 {object} easi_backend_internal_shared_api.CollectionResponse{data=[]easi_backend_internal_enterprisearchitecture_application_readmodels.EnterpriseCapabilityLinkDTO}
// @Failure 404 {object} easi_backend_internal_shared_api.ErrorResponse
// @Failure 500 {object} easi_backend_internal_shared_api.ErrorResponse
// @Router /enterprise-capabilities/{id}/links [get]
func (h *EnterpriseCapabilityHandlers) GetLinkedCapabilities(w http.ResponseWriter, r *http.Request) {
	enterpriseCapabilityID := sharedAPI.GetPathParam(r, "id")

	capability := h.getCapabilityOrNotFound(w, r, enterpriseCapabilityID)
	if capability == nil {
		return
	}

	links, err := h.readModels.Link.GetByEnterpriseCapabilityID(r.Context(), enterpriseCapabilityID)
	if err != nil {
		sharedAPI.HandleError(w, err)
		return
	}

	for i := range links {
		links[i].Links = h.hateoas.EnterpriseCapabilityLinkLinks(enterpriseCapabilityID, links[i].ID)
	}

	sharedAPI.RespondCollection(w, http.StatusOK, links, h.hateoas.EnterpriseCapabilityLinksCollectionLinks(enterpriseCapabilityID))
}

// LinkCapability godoc
// @Summary Link a domain capability
// @Description Links a domain capability to an enterprise capability
// @Tags enterprise-capabilities
// @Accept json
// @Produce json
// @Param id path string true "Enterprise capability ID"
// @Param link body LinkCapabilityRequest true "Link data"
// @Success 201 {object} easi_backend_internal_enterprisearchitecture_application_readmodels.EnterpriseCapabilityLinkDTO
// @Failure 400 {object} easi_backend_internal_shared_api.ErrorResponse
// @Failure 404 {object} easi_backend_internal_shared_api.ErrorResponse
// @Failure 409 {object} easi_backend_internal_shared_api.ErrorResponse
// @Failure 500 {object} easi_backend_internal_shared_api.ErrorResponse
// @Router /enterprise-capabilities/{id}/links [post]
func (h *EnterpriseCapabilityHandlers) LinkCapability(w http.ResponseWriter, r *http.Request) {
	enterpriseCapabilityID := sharedAPI.GetPathParam(r, "id")

	req, ok := sharedAPI.DecodeRequestOrFail[LinkCapabilityRequest](w, r)
	if !ok {
		return
	}

	linkedBy, err := h.getCurrentUserEmail(r)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusUnauthorized, err, "Authentication required")
		return
	}

	cmd := &commands.LinkCapability{
		EnterpriseCapabilityID: enterpriseCapabilityID,
		DomainCapabilityID:     req.DomainCapabilityID,
		LinkedBy:               linkedBy,
	}

	result, err := h.commandBus.Dispatch(r.Context(), cmd)
	sharedAPI.HandleCommandResult(w, result, err, func(createdID string) {
		link, err := h.readModels.Link.GetByID(r.Context(), createdID)
		if err != nil || link == nil {
			location := sharedAPI.BuildSubResourceLink(sharedAPI.ResourcePath("/enterprise-capabilities"), sharedAPI.ResourceID(enterpriseCapabilityID), sharedAPI.ResourcePath("/links/"+createdID))
			sharedAPI.RespondCreatedNoBody(w, location)
			return
		}

		location := sharedAPI.BuildSubResourceLink(sharedAPI.ResourcePath("/enterprise-capabilities"), sharedAPI.ResourceID(enterpriseCapabilityID), sharedAPI.ResourcePath("/links/"+createdID))
		link.Links = h.hateoas.EnterpriseCapabilityLinkLinks(enterpriseCapabilityID, link.ID)
		sharedAPI.RespondCreated(w, location, link)
	})
}

// UnlinkCapability godoc
// @Summary Unlink a domain capability
// @Description Removes the link between a domain capability and an enterprise capability
// @Tags enterprise-capabilities
// @Param id path string true "Enterprise capability ID"
// @Param linkId path string true "Link ID"
// @Success 204 "No Content"
// @Failure 404 {object} easi_backend_internal_shared_api.ErrorResponse
// @Failure 500 {object} easi_backend_internal_shared_api.ErrorResponse
// @Router /enterprise-capabilities/{id}/links/{linkId} [delete]
func (h *EnterpriseCapabilityHandlers) UnlinkCapability(w http.ResponseWriter, r *http.Request) {
	linkID := sharedAPI.GetPathParam(r, "linkId")
	if h.getLinkOrNotFound(w, r, linkID) == nil {
		return
	}

	result, err := h.commandBus.Dispatch(r.Context(), &commands.UnlinkCapability{LinkID: linkID})
	sharedAPI.HandleCommandResult(w, result, err, func(_ string) {
		sharedAPI.RespondDeleted(w)
	})
}

// GetStrategicImportance godoc
// @Summary Get strategic importance ratings
// @Description Retrieves all strategic importance ratings for an enterprise capability
// @Tags enterprise-capabilities
// @Produce json
// @Param id path string true "Enterprise capability ID"
// @Success 200 {object} easi_backend_internal_shared_api.CollectionResponse{data=[]easi_backend_internal_enterprisearchitecture_application_readmodels.EnterpriseStrategicImportanceDTO}
// @Failure 404 {object} easi_backend_internal_shared_api.ErrorResponse
// @Failure 500 {object} easi_backend_internal_shared_api.ErrorResponse
// @Router /enterprise-capabilities/{id}/strategic-importance [get]
func (h *EnterpriseCapabilityHandlers) GetStrategicImportance(w http.ResponseWriter, r *http.Request) {
	enterpriseCapabilityID := sharedAPI.GetPathParam(r, "id")

	capability := h.getCapabilityOrNotFound(w, r, enterpriseCapabilityID)
	if capability == nil {
		return
	}

	ratings, err := h.readModels.Importance.GetByEnterpriseCapabilityID(r.Context(), enterpriseCapabilityID)
	if err != nil {
		sharedAPI.HandleError(w, err)
		return
	}

	for i := range ratings {
		ratings[i].Links = h.hateoas.EnterpriseStrategicImportanceLinks(enterpriseCapabilityID, ratings[i].ID)
	}

	sharedAPI.RespondCollection(w, http.StatusOK, ratings, h.hateoas.EnterpriseStrategicImportanceCollectionLinks(enterpriseCapabilityID))
}

// SetStrategicImportance godoc
// @Summary Set strategic importance
// @Description Sets the strategic importance of an enterprise capability for a specific strategy pillar
// @Tags enterprise-capabilities
// @Accept json
// @Produce json
// @Param id path string true "Enterprise capability ID"
// @Param importance body SetStrategicImportanceRequest true "Strategic importance data"
// @Success 201 {object} easi_backend_internal_enterprisearchitecture_application_readmodels.EnterpriseStrategicImportanceDTO
// @Failure 400 {object} easi_backend_internal_shared_api.ErrorResponse
// @Failure 404 {object} easi_backend_internal_shared_api.ErrorResponse
// @Failure 409 {object} easi_backend_internal_shared_api.ErrorResponse
// @Failure 500 {object} easi_backend_internal_shared_api.ErrorResponse
// @Router /enterprise-capabilities/{id}/strategic-importance [post]
func (h *EnterpriseCapabilityHandlers) SetStrategicImportance(w http.ResponseWriter, r *http.Request) {
	enterpriseCapabilityID := sharedAPI.GetPathParam(r, "id")

	req, ok := sharedAPI.DecodeRequestOrFail[SetStrategicImportanceRequest](w, r)
	if !ok {
		return
	}

	cmd := &commands.SetEnterpriseStrategicImportance{
		EnterpriseCapabilityID: enterpriseCapabilityID,
		PillarID:               req.PillarID,
		PillarName:             req.PillarName,
		Importance:             req.Importance,
		Rationale:              req.Rationale,
	}

	result, err := h.commandBus.Dispatch(r.Context(), cmd)
	sharedAPI.HandleCommandResult(w, result, err, func(createdID string) {
		rating, err := h.readModels.Importance.GetByID(r.Context(), createdID)
		if err != nil || rating == nil {
			location := sharedAPI.BuildSubResourceLink(sharedAPI.ResourcePath("/enterprise-capabilities"), sharedAPI.ResourceID(enterpriseCapabilityID), sharedAPI.ResourcePath("/strategic-importance/"+createdID))
			sharedAPI.RespondCreatedNoBody(w, location)
			return
		}

		location := sharedAPI.BuildSubResourceLink(sharedAPI.ResourcePath("/enterprise-capabilities"), sharedAPI.ResourceID(enterpriseCapabilityID), sharedAPI.ResourcePath("/strategic-importance/"+createdID))
		rating.Links = h.hateoas.EnterpriseStrategicImportanceLinks(enterpriseCapabilityID, rating.ID)
		sharedAPI.RespondCreated(w, location, rating)
	})
}

// UpdateStrategicImportance godoc
// @Summary Update strategic importance
// @Description Updates the strategic importance rating for a specific pillar
// @Tags enterprise-capabilities
// @Accept json
// @Produce json
// @Param id path string true "Enterprise capability ID"
// @Param importanceId path string true "Strategic importance ID"
// @Param importance body UpdateStrategicImportanceRequest true "Updated importance data"
// @Success 200 {object} easi_backend_internal_enterprisearchitecture_application_readmodels.EnterpriseStrategicImportanceDTO
// @Failure 400 {object} easi_backend_internal_shared_api.ErrorResponse
// @Failure 404 {object} easi_backend_internal_shared_api.ErrorResponse
// @Failure 500 {object} easi_backend_internal_shared_api.ErrorResponse
// @Router /enterprise-capabilities/{id}/strategic-importance/{importanceId} [put]
func (h *EnterpriseCapabilityHandlers) UpdateStrategicImportance(w http.ResponseWriter, r *http.Request) {
	enterpriseCapabilityID := sharedAPI.GetPathParam(r, "id")
	importanceID := sharedAPI.GetPathParam(r, "importanceId")

	req, ok := sharedAPI.DecodeRequestOrFail[UpdateStrategicImportanceRequest](w, r)
	if !ok {
		return
	}

	cmd := &commands.UpdateEnterpriseStrategicImportance{
		ID:         importanceID,
		Importance: req.Importance,
		Rationale:  req.Rationale,
	}

	result, err := h.commandBus.Dispatch(r.Context(), cmd)
	sharedAPI.HandleCommandResult(w, result, err, func(_ string) {
		rating, err := h.readModels.Importance.GetByID(r.Context(), importanceID)
		if err != nil || rating == nil {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		rating.Links = h.hateoas.EnterpriseStrategicImportanceLinks(enterpriseCapabilityID, rating.ID)
		sharedAPI.RespondJSON(w, http.StatusOK, rating)
	})
}

// RemoveStrategicImportance godoc
// @Summary Remove strategic importance
// @Description Removes the strategic importance rating for a specific pillar
// @Tags enterprise-capabilities
// @Param id path string true "Enterprise capability ID"
// @Param importanceId path string true "Strategic importance ID"
// @Success 204 "No Content"
// @Failure 404 {object} easi_backend_internal_shared_api.ErrorResponse
// @Failure 500 {object} easi_backend_internal_shared_api.ErrorResponse
// @Router /enterprise-capabilities/{id}/strategic-importance/{importanceId} [delete]
func (h *EnterpriseCapabilityHandlers) RemoveStrategicImportance(w http.ResponseWriter, r *http.Request) {
	importanceID := sharedAPI.GetPathParam(r, "importanceId")
	if h.getImportanceOrNotFound(w, r, importanceID) == nil {
		return
	}

	result, err := h.commandBus.Dispatch(r.Context(), &commands.RemoveEnterpriseStrategicImportance{ID: importanceID})
	sharedAPI.HandleCommandResult(w, result, err, func(_ string) {
		sharedAPI.RespondDeleted(w)
	})
}

// GetEnterpriseCapabilityForDomainCapability godoc
// @Summary Get enterprise capability for a domain capability
// @Description Retrieves the enterprise capability linked to a specific domain capability
// @Tags enterprise-capabilities
// @Produce json
// @Param domainCapabilityId path string true "Domain capability ID"
// @Success 200 {object} DomainCapabilityEnterpriseResponse
// @Failure 500 {object} easi_backend_internal_shared_api.ErrorResponse
// @Router /domain-capabilities/{domainCapabilityId}/enterprise-capability [get]
func (h *EnterpriseCapabilityHandlers) GetEnterpriseCapabilityForDomainCapability(w http.ResponseWriter, r *http.Request) {
	domainCapabilityID := sharedAPI.GetPathParam(r, "domainCapabilityId")

	link, err := h.readModels.Link.GetByDomainCapabilityID(r.Context(), domainCapabilityID)
	if err != nil {
		sharedAPI.HandleError(w, err)
		return
	}

	if link == nil {
		response := DomainCapabilityEnterpriseResponse{
			Linked:                 false,
			EnterpriseCapabilityID: nil,
			Links:                  h.hateoas.DomainCapabilityEnterpriseLinks(domainCapabilityID),
		}
		sharedAPI.RespondJSON(w, http.StatusOK, response)
		return
	}

	capability, err := h.readModels.Capability.GetByID(r.Context(), link.EnterpriseCapabilityID)
	if err != nil {
		sharedAPI.HandleError(w, err)
		return
	}

	response := DomainCapabilityEnterpriseResponse{
		Linked:                   true,
		EnterpriseCapabilityID:   &link.EnterpriseCapabilityID,
		EnterpriseCapabilityName: &capability.Name,
		LinkID:                   &link.ID,
		Links:                    h.hateoas.DomainCapabilityEnterpriseLinkedLinks(domainCapabilityID, link.EnterpriseCapabilityID, link.ID),
	}
	sharedAPI.RespondJSON(w, http.StatusOK, response)
}

func getOrNotFound[T any](w http.ResponseWriter, fetchFn func() (*T, error), resourceName string) *T {
	result, err := fetchFn()
	if err != nil {
		sharedAPI.HandleError(w, err)
		return nil
	}
	if result == nil {
		sharedAPI.RespondError(w, http.StatusNotFound, nil, resourceName+" not found")
		return nil
	}
	return result
}

func (h *EnterpriseCapabilityHandlers) getCapabilityOrNotFound(w http.ResponseWriter, r *http.Request, id string) *readmodels.EnterpriseCapabilityDTO {
	return getOrNotFound(w, func() (*readmodels.EnterpriseCapabilityDTO, error) {
		return h.readModels.Capability.GetByID(r.Context(), id)
	}, "Enterprise capability")
}

func (h *EnterpriseCapabilityHandlers) getLinkOrNotFound(w http.ResponseWriter, r *http.Request, id string) *readmodels.EnterpriseCapabilityLinkDTO {
	return getOrNotFound(w, func() (*readmodels.EnterpriseCapabilityLinkDTO, error) {
		return h.readModels.Link.GetByID(r.Context(), id)
	}, "Link")
}

func (h *EnterpriseCapabilityHandlers) getImportanceOrNotFound(w http.ResponseWriter, r *http.Request, id string) *readmodels.EnterpriseStrategicImportanceDTO {
	return getOrNotFound(w, func() (*readmodels.EnterpriseStrategicImportanceDTO, error) {
		return h.readModels.Importance.GetByID(r.Context(), id)
	}, "Importance rating")
}

func (h *EnterpriseCapabilityHandlers) respondWithCapability(w http.ResponseWriter, r *http.Request, capabilityID string, statusCode int) {
	capability, err := h.readModels.Capability.GetByID(r.Context(), capabilityID)
	if err != nil {
		sharedAPI.HandleError(w, err)
		return
	}

	location := sharedAPI.BuildResourceLink(sharedAPI.ResourcePath("/enterprise-capabilities"), sharedAPI.ResourceID(capabilityID))

	if capability == nil {
		if statusCode == http.StatusCreated {
			sharedAPI.RespondCreatedNoBody(w, location)
			return
		}
		sharedAPI.RespondError(w, http.StatusNotFound, nil, "Enterprise capability not found")
		return
	}

	capability.Links = h.hateoas.EnterpriseCapabilityLinks(capability.ID)
	if statusCode == http.StatusCreated {
		sharedAPI.RespondCreated(w, location, capability)
	} else {
		sharedAPI.RespondJSON(w, statusCode, capability)
	}
}

func (h *EnterpriseCapabilityHandlers) getCurrentUserEmail(r *http.Request) (string, error) {
	authSession, err := h.sessionManager.LoadAuthenticatedSession(r.Context())
	if err != nil {
		return "", err
	}
	return authSession.UserEmail(), nil
}
