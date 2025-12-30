package api

import (
	"fmt"
	"net/http"
	"strconv"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/application/readmodels"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/cqrs"
)

type BusinessDomainReadModels struct {
	Domain      *readmodels.BusinessDomainReadModel
	Assignment  *readmodels.DomainCapabilityAssignmentReadModel
	Capability  *readmodels.CapabilityReadModel
	Realization *readmodels.RealizationReadModel
}

type BusinessDomainHandlers struct {
	commandBus cqrs.CommandBus
	readModels *BusinessDomainReadModels
	hateoas    *sharedAPI.HATEOASLinks
}

func NewBusinessDomainHandlers(
	commandBus cqrs.CommandBus,
	readModels *BusinessDomainReadModels,
	hateoas *sharedAPI.HATEOASLinks,
) *BusinessDomainHandlers {
	return &BusinessDomainHandlers{
		commandBus: commandBus,
		readModels: readModels,
		hateoas:    hateoas,
	}
}

type CreateBusinessDomainRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type UpdateBusinessDomainRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type AssignCapabilityRequest struct {
	CapabilityID string `json:"capabilityId"`
}

type AssignmentResponse struct {
	BusinessDomainID string            `json:"businessDomainId"`
	CapabilityID     string            `json:"capabilityId"`
	AssignedAt       string            `json:"assignedAt"`
	Links            map[string]string `json:"_links"`
}

type CapabilityInDomainDTO struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Level       string            `json:"level"`
	AssignedAt  string            `json:"assignedAt"`
	Links       map[string]string `json:"_links"`
}

type DomainForCapabilityDTO struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	AssignedAt  string            `json:"assignedAt"`
	Links       map[string]string `json:"_links"`
}

// CreateBusinessDomain godoc
// @Summary Create a new business domain
// @Description Creates a new business domain for organizing capabilities
// @Tags business-domains
// @Accept json
// @Produce json
// @Param domain body CreateBusinessDomainRequest true "Domain data"
// @Success 201 {object} easi_backend_internal_capabilitymapping_application_readmodels.BusinessDomainDTO
// @Failure 400 {object} easi_backend_internal_shared_api.ErrorResponse
// @Failure 500 {object} easi_backend_internal_shared_api.ErrorResponse
// @Router /business-domains [post]
func (h *BusinessDomainHandlers) CreateBusinessDomain(w http.ResponseWriter, r *http.Request) {
	req, ok := sharedAPI.DecodeRequestOrFail[CreateBusinessDomainRequest](w, r)
	if !ok {
		return
	}

	cmd := &commands.CreateBusinessDomain{
		Name:        req.Name,
		Description: req.Description,
	}

	result, err := h.commandBus.Dispatch(r.Context(), cmd)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "")
		return
	}

	h.respondWithDomain(w, r, result.CreatedID, http.StatusCreated)
}

// GetAllBusinessDomains godoc
// @Summary List all business domains
// @Description Returns all business domains with their capability counts
// @Tags business-domains
// @Produce json
// @Success 200 {object} easi_backend_internal_shared_api.CollectionResponse{data=[]easi_backend_internal_capabilitymapping_application_readmodels.BusinessDomainDTO}
// @Failure 500 {object} easi_backend_internal_shared_api.ErrorResponse
// @Router /business-domains [get]
func (h *BusinessDomainHandlers) GetAllBusinessDomains(w http.ResponseWriter, r *http.Request) {
	domains, err := h.readModels.Domain.GetAll(r.Context())
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve domains")
		return
	}

	for i := range domains {
		domains[i].Links = h.hateoas.BusinessDomainLinks(domains[i].ID, domains[i].CapabilityCount > 0)
	}

	links := map[string]string{
		"self": "/api/v1/business-domains",
	}

	sharedAPI.RespondCollection(w, http.StatusOK, domains, links)
}

func (h *BusinessDomainHandlers) getDomainOrNotFound(w http.ResponseWriter, r *http.Request, id string) *readmodels.BusinessDomainDTO {
	domain, err := h.readModels.Domain.GetByID(r.Context(), id)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve domain")
		return nil
	}
	if domain == nil {
		sharedAPI.RespondError(w, http.StatusNotFound, nil, "Domain not found")
		return nil
	}
	return domain
}

// GetBusinessDomainByID godoc
// @Summary Get a business domain by ID
// @Description Returns a single business domain with its details
// @Tags business-domains
// @Produce json
// @Param id path string true "Business Domain ID"
// @Success 200 {object} easi_backend_internal_capabilitymapping_application_readmodels.BusinessDomainDTO
// @Failure 404 {object} easi_backend_internal_shared_api.ErrorResponse
// @Failure 500 {object} easi_backend_internal_shared_api.ErrorResponse
// @Router /business-domains/{id} [get]
func (h *BusinessDomainHandlers) GetBusinessDomainByID(w http.ResponseWriter, r *http.Request) {
	domain := h.getDomainOrNotFound(w, r, sharedAPI.GetPathParam(r, "id"))
	if domain == nil {
		return
	}
	domain.Links = h.hateoas.BusinessDomainLinks(domain.ID, domain.CapabilityCount > 0)
	sharedAPI.RespondJSON(w, http.StatusOK, domain)
}

// UpdateBusinessDomain godoc
// @Summary Update a business domain
// @Description Updates an existing business domain's name and description
// @Tags business-domains
// @Accept json
// @Produce json
// @Param id path string true "Business Domain ID"
// @Param domain body UpdateBusinessDomainRequest true "Updated domain data"
// @Success 200 {object} easi_backend_internal_capabilitymapping_application_readmodels.BusinessDomainDTO
// @Failure 400 {object} easi_backend_internal_shared_api.ErrorResponse
// @Failure 404 {object} easi_backend_internal_shared_api.ErrorResponse
// @Failure 409 {object} easi_backend_internal_shared_api.ErrorResponse "Domain name already exists"
// @Failure 500 {object} easi_backend_internal_shared_api.ErrorResponse
// @Router /business-domains/{id} [put]
func (h *BusinessDomainHandlers) UpdateBusinessDomain(w http.ResponseWriter, r *http.Request) {
	id := sharedAPI.GetPathParam(r, "id")

	req, ok := sharedAPI.DecodeRequestOrFail[UpdateBusinessDomainRequest](w, r)
	if !ok {
		return
	}

	cmd := &commands.UpdateBusinessDomain{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
	}

	result, err := h.commandBus.Dispatch(r.Context(), cmd)
	sharedAPI.HandleCommandResult(w, result, err, func(_ string) {
		h.respondWithDomain(w, r, id, http.StatusOK)
	})
}

func (h *BusinessDomainHandlers) respondWithDomain(w http.ResponseWriter, r *http.Request, domainID string, statusCode int) {
	domain, err := h.readModels.Domain.GetByID(r.Context(), domainID)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve domain")
		return
	}

	location := sharedAPI.BuildResourceLink(sharedAPI.ResourcePath("/business-domains"), sharedAPI.ResourceID(domainID))

	if domain == nil {
		if statusCode == http.StatusCreated {
			sharedAPI.RespondCreated(w, location, map[string]string{
				"id":      domainID,
				"message": "Domain created, processing",
			})
			return
		}
		sharedAPI.RespondError(w, http.StatusNotFound, nil, "Domain not found")
		return
	}

	domain.Links = h.hateoas.BusinessDomainLinks(domain.ID, domain.CapabilityCount > 0)
	if statusCode == http.StatusCreated {
		sharedAPI.RespondCreated(w, location, domain)
	} else {
		sharedAPI.RespondJSON(w, statusCode, domain)
	}
}

// DeleteBusinessDomain godoc
// @Summary Delete a business domain
// @Description Deletes a business domain (only if it has no assigned capabilities)
// @Tags business-domains
// @Param id path string true "Business Domain ID"
// @Success 204 "No Content"
// @Failure 404 {object} easi_backend_internal_shared_api.ErrorResponse
// @Failure 409 {object} easi_backend_internal_shared_api.ErrorResponse "Domain has assigned capabilities"
// @Failure 500 {object} easi_backend_internal_shared_api.ErrorResponse
// @Router /business-domains/{id} [delete]
func (h *BusinessDomainHandlers) DeleteBusinessDomain(w http.ResponseWriter, r *http.Request) {
	id := sharedAPI.GetPathParam(r, "id")
	if h.getDomainOrNotFound(w, r, id) == nil {
		return
	}

	result, err := h.commandBus.Dispatch(r.Context(), &commands.DeleteBusinessDomain{ID: id})
	sharedAPI.HandleCommandResult(w, result, err, func(_ string) {
		sharedAPI.RespondDeleted(w)
	})
}

// GetCapabilitiesInDomain godoc
// @Summary List capabilities assigned to a domain
// @Description Returns all capabilities assigned to a specific business domain
// @Tags business-domains
// @Produce json
// @Param id path string true "Business Domain ID"
// @Success 200 {object} easi_backend_internal_shared_api.CollectionResponse{data=[]api.CapabilityInDomainDTO}
// @Failure 404 {object} easi_backend_internal_shared_api.ErrorResponse
// @Failure 500 {object} easi_backend_internal_shared_api.ErrorResponse
// @Router /business-domains/{id}/capabilities [get]
func (h *BusinessDomainHandlers) GetCapabilitiesInDomain(w http.ResponseWriter, r *http.Request) {
	domainID := sharedAPI.GetPathParam(r, "id")

	domain, err := h.readModels.Domain.GetByID(r.Context(), domainID)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve domain")
		return
	}

	if domain == nil {
		sharedAPI.RespondError(w, http.StatusNotFound, nil, "Domain not found")
		return
	}

	assignments, err := h.readModels.Assignment.GetByDomainID(r.Context(), domainID)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve capabilities")
		return
	}

	capabilities := make([]CapabilityInDomainDTO, len(assignments))
	for i, a := range assignments {
		capabilities[i] = CapabilityInDomainDTO{
			ID:          a.CapabilityID,
			Name:        a.CapabilityName,
			Description: a.CapabilityDescription,
			Level:       a.CapabilityLevel,
			AssignedAt:  a.AssignedAt.Format("2006-01-02T15:04:05Z07:00"),
			Links:       h.hateoas.CapabilityInDomainLinks(a.CapabilityID, domainID),
		}
	}

	links := map[string]string{
		"self":   fmt.Sprintf("/api/v1/business-domains/%s/capabilities", domainID),
		"domain": fmt.Sprintf("/api/v1/business-domains/%s", domainID),
	}

	sharedAPI.RespondCollection(w, http.StatusOK, capabilities, links)
}

// AssignCapabilityToDomain godoc
// @Summary Assign a capability to a domain
// @Description Assigns an L1 capability to a business domain
// @Tags business-domains
// @Accept json
// @Produce json
// @Param id path string true "Business Domain ID"
// @Param assignment body AssignCapabilityRequest true "Capability assignment"
// @Success 201 {object} AssignmentResponse
// @Failure 400 {object} easi_backend_internal_shared_api.ErrorResponse
// @Failure 404 {object} easi_backend_internal_shared_api.ErrorResponse "Capability not found"
// @Failure 409 {object} easi_backend_internal_shared_api.ErrorResponse "Capability already assigned"
// @Failure 500 {object} easi_backend_internal_shared_api.ErrorResponse
// @Router /business-domains/{id}/capabilities [post]
func (h *BusinessDomainHandlers) AssignCapabilityToDomain(w http.ResponseWriter, r *http.Request) {
	domainID := sharedAPI.GetPathParam(r, "id")

	req, ok := sharedAPI.DecodeRequestOrFail[AssignCapabilityRequest](w, r)
	if !ok {
		return
	}

	cmd := &commands.AssignCapabilityToDomain{
		BusinessDomainID: domainID,
		CapabilityID:     req.CapabilityID,
	}

	result, err := h.commandBus.Dispatch(r.Context(), cmd)
	sharedAPI.HandleCommandResult(w, result, err, func(_ string) {
		h.respondWithAssignment(w, r, domainID, req.CapabilityID)
	})
}

func (h *BusinessDomainHandlers) respondWithAssignment(w http.ResponseWriter, r *http.Request, domainID, capabilityID string) {
	assignment, err := h.readModels.Assignment.GetByDomainAndCapability(r.Context(), domainID, capabilityID)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve assignment")
		return
	}

	if assignment == nil {
		sharedAPI.RespondJSON(w, http.StatusCreated, map[string]string{
			"message": "Assignment created, processing",
		})
		return
	}

	location := fmt.Sprintf("/api/v1/business-domains/%s/capabilities/%s", domainID, capabilityID)
	w.Header().Set("Location", location)

	response := AssignmentResponse{
		BusinessDomainID: assignment.BusinessDomainID,
		CapabilityID:     assignment.CapabilityID,
		AssignedAt:       assignment.AssignedAt.Format("2006-01-02T15:04:05Z07:00"),
		Links:            h.hateoas.AssignmentLinks(domainID, capabilityID),
	}

	sharedAPI.RespondJSON(w, http.StatusCreated, response)
}

// RemoveCapabilityFromDomain godoc
// @Summary Remove a capability from a domain
// @Description Removes the assignment between a capability and a business domain
// @Tags business-domains
// @Param domainId path string true "Business Domain ID"
// @Param capabilityId path string true "Capability ID"
// @Success 204 "No Content"
// @Failure 404 {object} easi_backend_internal_shared_api.ErrorResponse
// @Failure 500 {object} easi_backend_internal_shared_api.ErrorResponse
// @Router /business-domains/{domainId}/capabilities/{capabilityId} [delete]
func (h *BusinessDomainHandlers) RemoveCapabilityFromDomain(w http.ResponseWriter, r *http.Request) {
	domainID := sharedAPI.GetPathParam(r, "domainId")
	capabilityID := sharedAPI.GetPathParam(r, "capabilityId")

	assignment, err := h.readModels.Assignment.GetByDomainAndCapability(r.Context(), domainID, capabilityID)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve assignment")
		return
	}

	if assignment == nil {
		sharedAPI.RespondError(w, http.StatusNotFound, nil, "Assignment not found")
		return
	}

	cmd := &commands.UnassignCapabilityFromDomain{
		AssignmentID: assignment.AssignmentID,
	}

	if _, err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to remove capability from domain")
		return
	}

	sharedAPI.RespondDeleted(w)
}

// GetDomainsForCapability godoc
// @Summary List business domains for a capability
// @Description Returns all business domains that have a specific capability assigned
// @Tags capabilities
// @Produce json
// @Param id path string true "Capability ID"
// @Success 200 {object} easi_backend_internal_shared_api.CollectionResponse{data=[]api.DomainForCapabilityDTO}
// @Failure 404 {object} easi_backend_internal_shared_api.ErrorResponse
// @Failure 500 {object} easi_backend_internal_shared_api.ErrorResponse
// @Router /capabilities/{id}/business-domains [get]
func (h *BusinessDomainHandlers) GetDomainsForCapability(w http.ResponseWriter, r *http.Request) {
	capabilityID := sharedAPI.GetPathParam(r, "id")

	capability, err := h.readModels.Capability.GetByID(r.Context(), capabilityID)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve capability")
		return
	}

	if capability == nil {
		sharedAPI.RespondError(w, http.StatusNotFound, nil, "Capability not found")
		return
	}

	assignments, err := h.readModels.Assignment.GetByCapabilityID(r.Context(), capabilityID)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve domains")
		return
	}

	domains := make([]DomainForCapabilityDTO, len(assignments))
	for i, a := range assignments {
		domains[i] = DomainForCapabilityDTO{
			ID:         a.BusinessDomainID,
			Name:       a.BusinessDomainName,
			AssignedAt: a.AssignedAt.Format("2006-01-02T15:04:05Z07:00"),
			Links:      h.hateoas.DomainForCapabilityLinks(a.BusinessDomainID, capabilityID),
		}
	}

	links := map[string]string{
		"self":       fmt.Sprintf("/api/v1/capabilities/%s/business-domains", capabilityID),
		"capability": fmt.Sprintf("/api/v1/capabilities/%s", capabilityID),
	}

	sharedAPI.RespondCollection(w, http.StatusOK, domains, links)
}

type CapabilityRealizationsGroupDTO struct {
	CapabilityID   string                      `json:"capabilityId"`
	CapabilityName string                      `json:"capabilityName"`
	Level          string                      `json:"level"`
	Realizations   []readmodels.RealizationDTO `json:"realizations"`
}

func parseDepthParam(r *http.Request) (int, error) {
	depthStr := r.URL.Query().Get("depth")
	if depthStr == "" {
		return 4, nil
	}
	depth, err := strconv.Atoi(depthStr)
	if err != nil {
		return 0, fmt.Errorf("depth must be between 1 and 4")
	}
	if depth < 1 || depth > 4 {
		return 0, fmt.Errorf("depth must be between 1 and 4")
	}
	return depth, nil
}

func (h *BusinessDomainHandlers) toRealizationGroupDTOs(groups []readmodels.CapabilityRealizationsGroup) []CapabilityRealizationsGroupDTO {
	result := make([]CapabilityRealizationsGroupDTO, len(groups))
	for i, g := range groups {
		result[i] = CapabilityRealizationsGroupDTO{
			CapabilityID:   g.CapabilityID,
			CapabilityName: g.CapabilityName,
			Level:          g.Level,
			Realizations:   h.addRealizationLinks(g.Realizations),
		}
	}
	return result
}

func (h *BusinessDomainHandlers) addRealizationLinks(realizations []readmodels.RealizationDTO) []readmodels.RealizationDTO {
	result := make([]readmodels.RealizationDTO, len(realizations))
	for i, r := range realizations {
		r.Links = h.hateoas.RealizationLinks(r.ID, r.CapabilityID, r.ComponentID)
		result[i] = r
	}
	return result
}

// GetCapabilityRealizationsByDomain godoc
// @Summary Get all capability realizations for a business domain
// @Description Returns all application components that realize capabilities in a business domain, up to the specified depth level
// @Tags business-domains
// @Produce json
// @Param id path string true "Business Domain ID"
// @Param depth query int false "Maximum capability depth level (1-4)" default(4)
// @Success 200 {object} easi_backend_internal_shared_api.CollectionResponse{data=[]api.CapabilityRealizationsGroupDTO}
// @Failure 400 {object} easi_backend_internal_shared_api.ErrorResponse
// @Failure 404 {object} easi_backend_internal_shared_api.ErrorResponse
// @Failure 500 {object} easi_backend_internal_shared_api.ErrorResponse
// @Router /business-domains/{id}/capability-realizations [get]
func (h *BusinessDomainHandlers) GetCapabilityRealizationsByDomain(w http.ResponseWriter, r *http.Request) {
	domainID := sharedAPI.GetPathParam(r, "id")

	domain, err := h.readModels.Domain.GetByID(r.Context(), domainID)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve domain")
		return
	}
	if domain == nil {
		sharedAPI.RespondError(w, http.StatusNotFound, nil, "Domain not found")
		return
	}

	depth, err := parseDepthParam(r)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, nil, err.Error())
		return
	}

	groups, err := h.readModels.Realization.GetByBusinessDomainAndDepth(r.Context(), domainID, depth)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve realizations")
		return
	}

	links := map[string]string{
		"self":   fmt.Sprintf("/api/v1/business-domains/%s/capability-realizations?depth=%d", domainID, depth),
		"domain": fmt.Sprintf("/api/v1/business-domains/%s", domainID),
	}

	sharedAPI.RespondCollection(w, http.StatusOK, h.toRealizationGroupDTOs(groups), links)
}
