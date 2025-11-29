package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/application/readmodels"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/cqrs"
	"github.com/go-chi/chi/v5"
)

type BusinessDomainReadModels struct {
	Domain      *readmodels.BusinessDomainReadModel
	Assignment  *readmodels.DomainCapabilityAssignmentReadModel
	Capability  *readmodels.CapabilityReadModel
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
	Code        string            `json:"code"`
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

func (h *BusinessDomainHandlers) CreateBusinessDomain(w http.ResponseWriter, r *http.Request) {
	var req CreateBusinessDomainRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	cmd := &commands.CreateBusinessDomain{
		Name:        req.Name,
		Description: req.Description,
	}

	if err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "")
		return
	}

	h.respondWithDomain(w, r, cmd.ID, http.StatusCreated)
}

func (h *BusinessDomainHandlers) GetAllBusinessDomains(w http.ResponseWriter, r *http.Request) {
	domains, err := h.readModels.Domain.GetAll(r.Context())
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve domains")
		return
	}

	for i := range domains {
		domains[i].Links = h.hateoas.BusinessDomainLinks(domains[i].ID, domains[i].CapabilityCount > 0)
	}

	params := sharedAPI.ParsePaginationParams(r)
	selfLink := fmt.Sprintf("/api/v1/business-domains?limit=%d", params.Limit)
	baseLink := "/api/v1/business-domains"

	sharedAPI.RespondPaginated(w, sharedAPI.PaginatedResponseParams{
		StatusCode: http.StatusOK,
		Data:       domains,
		HasMore:    false,
		NextCursor: "",
		Limit:      params.Limit,
		SelfLink:   selfLink,
		BaseLink:   baseLink,
	})
}

func (h *BusinessDomainHandlers) GetBusinessDomainByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	domain, err := h.readModels.Domain.GetByID(r.Context(), id)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve domain")
		return
	}

	if domain == nil {
		sharedAPI.RespondError(w, http.StatusNotFound, nil, "Domain not found")
		return
	}

	domain.Links = h.hateoas.BusinessDomainLinks(domain.ID, domain.CapabilityCount > 0)
	sharedAPI.RespondJSON(w, http.StatusOK, domain)
}

func (h *BusinessDomainHandlers) UpdateBusinessDomain(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req UpdateBusinessDomainRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	cmd := &commands.UpdateBusinessDomain{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
	}

	if err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "")
		return
	}

	h.respondWithDomain(w, r, id, http.StatusOK)
}

func (h *BusinessDomainHandlers) respondWithDomain(w http.ResponseWriter, r *http.Request, domainID string, statusCode int) {
	domain, err := h.readModels.Domain.GetByID(r.Context(), domainID)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve domain")
		return
	}

	if domain == nil {
		if statusCode == http.StatusCreated {
			sharedAPI.RespondJSON(w, http.StatusCreated, map[string]string{
				"id":      domainID,
				"message": "Domain created, processing",
			})
			return
		}
		sharedAPI.RespondError(w, http.StatusNotFound, nil, "Domain not found")
		return
	}

	if statusCode == http.StatusCreated {
		location := fmt.Sprintf("/api/v1/business-domains/%s", domainID)
		w.Header().Set("Location", location)
	}

	domain.Links = h.hateoas.BusinessDomainLinks(domain.ID, domain.CapabilityCount > 0)
	sharedAPI.RespondJSON(w, statusCode, domain)
}

func (h *BusinessDomainHandlers) DeleteBusinessDomain(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	domain, err := h.readModels.Domain.GetByID(r.Context(), id)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve domain")
		return
	}

	if domain == nil {
		sharedAPI.RespondError(w, http.StatusNotFound, nil, "Domain not found")
		return
	}

	cmd := &commands.DeleteBusinessDomain{
		ID: id,
	}

	if err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *BusinessDomainHandlers) GetCapabilitiesInDomain(w http.ResponseWriter, r *http.Request) {
	domainID := chi.URLParam(r, "id")

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
			Code:        a.CapabilityCode,
			Name:        a.CapabilityName,
			Level:       a.CapabilityLevel,
			AssignedAt:  a.AssignedAt.Format("2006-01-02T15:04:05Z07:00"),
			Links:       h.hateoas.CapabilityInDomainLinks(a.CapabilityID, domainID),
		}
	}

	params := sharedAPI.ParsePaginationParams(r)
	selfLink := fmt.Sprintf("/api/v1/business-domains/%s/capabilities?limit=%d", domainID, params.Limit)
	baseLink := fmt.Sprintf("/api/v1/business-domains/%s/capabilities", domainID)

	sharedAPI.RespondPaginated(w, sharedAPI.PaginatedResponseParams{
		StatusCode: http.StatusOK,
		Data:       capabilities,
		HasMore:    false,
		NextCursor: "",
		Limit:      params.Limit,
		SelfLink:   selfLink,
		BaseLink:   baseLink,
	})
}

func (h *BusinessDomainHandlers) AssignCapabilityToDomain(w http.ResponseWriter, r *http.Request) {
	domainID := chi.URLParam(r, "id")

	var req AssignCapabilityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	cmd := &commands.AssignCapabilityToDomain{
		BusinessDomainID: domainID,
		CapabilityID:     req.CapabilityID,
	}

	if err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "")
		return
	}

	assignment, err := h.readModels.Assignment.GetByDomainAndCapability(r.Context(), domainID, req.CapabilityID)
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

	location := fmt.Sprintf("/api/v1/business-domains/%s/capabilities/%s", domainID, req.CapabilityID)
	w.Header().Set("Location", location)

	response := AssignmentResponse{
		BusinessDomainID: assignment.BusinessDomainID,
		CapabilityID:     assignment.CapabilityID,
		AssignedAt:       assignment.AssignedAt.Format("2006-01-02T15:04:05Z07:00"),
		Links:            h.hateoas.AssignmentLinks(domainID, req.CapabilityID),
	}

	sharedAPI.RespondJSON(w, http.StatusCreated, response)
}

func (h *BusinessDomainHandlers) RemoveCapabilityFromDomain(w http.ResponseWriter, r *http.Request) {
	domainID := chi.URLParam(r, "domainId")
	capabilityID := chi.URLParam(r, "capabilityId")

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

	if err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to remove capability from domain")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *BusinessDomainHandlers) GetDomainsForCapability(w http.ResponseWriter, r *http.Request) {
	capabilityID := chi.URLParam(r, "id")

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
