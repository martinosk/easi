package api

import (
	"net/http"
	"strings"

	"easi/backend/internal/enterprisearchitecture/application/commands"
	"easi/backend/internal/enterprisearchitecture/application/readmodels"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/types"
)

type LinkCapabilityRequest struct {
	DomainCapabilityID string `json:"domainCapabilityId"`
}

type DomainCapabilityEnterpriseResponse struct {
	Linked                   bool        `json:"linked"`
	EnterpriseCapabilityID   *string     `json:"enterpriseCapabilityId"`
	EnterpriseCapabilityName *string     `json:"enterpriseCapabilityName,omitempty"`
	LinkID                   *string     `json:"linkId,omitempty"`
	Links                    types.Links `json:"_links,omitempty"`
}

// GetLinkedCapabilities godoc
// @Summary Get linked domain capabilities
// @Description Retrieves all domain capabilities linked to an enterprise capability
// @Tags enterprise-capabilities
// @Produce json
// @Param id path string true "Enterprise capability ID"
// @Success 200 {object} easi_backend_internal_shared_api.CollectionResponse{data=[]easi_backend_internal_enterprisearchitecture_application_readmodels.EnterpriseCapabilityLinkDTO}
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /enterprise-capabilities/{id}/links [get]
func (h *EnterpriseCapabilityHandlers) GetLinkedCapabilities(w http.ResponseWriter, r *http.Request) {
	respondScopedCollection(w, r, h, scopedCollection[readmodels.EnterpriseCapabilityLinkDTO]{
		fetch: h.readModels.Link.GetByEnterpriseCapabilityID,
		decorate: func(_ *http.Request, ecID string, items []readmodels.EnterpriseCapabilityLinkDTO) {
			for i := range items {
				items[i].Links = h.hateoas.EnterpriseCapabilityLinkLinks(ecID, items[i].ID)
			}
		},
		collectionLinks: h.hateoas.EnterpriseCapabilityLinksCollectionLinks,
	})
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
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 409 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /enterprise-capabilities/{id}/links [post]
func (h *EnterpriseCapabilityHandlers) LinkCapability(w http.ResponseWriter, r *http.Request) {
	enterpriseCapabilityID := sharedAPI.GetPathParam(r, "id")

	req, ok := sharedAPI.DecodeRequestOrFail[LinkCapabilityRequest](w, r)
	if !ok {
		return
	}

	linkedBy, err := h.sessionProvider.GetCurrentUserEmail(r.Context())
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
		location := sharedAPI.BuildSubResourceLink(sharedAPI.ResourcePath("/enterprise-capabilities"), sharedAPI.ResourceID(enterpriseCapabilityID), sharedAPI.ResourcePath("/links/"+createdID))
		link, err := h.readModels.Link.GetByID(r.Context(), createdID)
		if err != nil || link == nil {
			sharedAPI.RespondCreatedNoBody(w, location)
			return
		}
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
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /enterprise-capabilities/{id}/links/{linkId} [delete]
func (h *EnterpriseCapabilityHandlers) UnlinkCapability(w http.ResponseWriter, r *http.Request) {
	linkID := sharedAPI.GetPathParam(r, "linkId")
	if h.getLinkOrNotFound(w, r, linkID) == nil {
		return
	}
	h.dispatchDelete(w, r, &commands.UnlinkCapability{LinkID: linkID})
}

// GetEnterpriseCapabilityForDomainCapability godoc
// @Summary Get enterprise capability for a domain capability
// @Description Retrieves the enterprise capability linked to a specific domain capability
// @Tags enterprise-capabilities
// @Produce json
// @Param domainCapabilityId path string true "Domain capability ID"
// @Success 200 {object} DomainCapabilityEnterpriseResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /domain-capabilities/{domainCapabilityId}/enterprise-capability [get]
func (h *EnterpriseCapabilityHandlers) GetEnterpriseCapabilityForDomainCapability(w http.ResponseWriter, r *http.Request) {
	domainCapabilityID := sharedAPI.GetPathParam(r, "domainCapabilityId")

	link, err := h.readModels.Link.GetByDomainCapabilityID(r.Context(), domainCapabilityID)
	if err != nil {
		sharedAPI.HandleError(w, err)
		return
	}

	if link == nil {
		sharedAPI.RespondJSON(w, http.StatusOK, DomainCapabilityEnterpriseResponse{
			Linked: false,
			Links:  h.hateoas.DomainCapabilityEnterpriseLinks(domainCapabilityID),
		})
		return
	}

	capability, err := h.readModels.Capability.GetByID(r.Context(), link.EnterpriseCapabilityID)
	if err != nil {
		sharedAPI.HandleError(w, err)
		return
	}

	sharedAPI.RespondJSON(w, http.StatusOK, DomainCapabilityEnterpriseResponse{
		Linked:                   true,
		EnterpriseCapabilityID:   &link.EnterpriseCapabilityID,
		EnterpriseCapabilityName: &capability.Name,
		LinkID:                   &link.ID,
		Links:                    h.hateoas.DomainCapabilityEnterpriseLinkedLinks(domainCapabilityID, link.EnterpriseCapabilityID, link.ID),
	})
}

// GetCapabilityLinkStatus godoc
// @Summary Get link eligibility status for a domain capability
// @Description Checks if a domain capability can be linked to an enterprise capability
// @Tags enterprise-capabilities
// @Produce json
// @Param domainCapabilityId path string true "Domain capability ID"
// @Success 200 {object} easi_backend_internal_enterprisearchitecture_application_readmodels.CapabilityLinkStatusDTO
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /domain-capabilities/{domainCapabilityId}/enterprise-link-status [get]
func (h *EnterpriseCapabilityHandlers) GetCapabilityLinkStatus(w http.ResponseWriter, r *http.Request) {
	capabilityID := sharedAPI.GetPathParam(r, "domainCapabilityId")

	status, err := h.readModels.Link.GetLinkStatus(r.Context(), capabilityID)
	if err != nil {
		sharedAPI.HandleError(w, err)
		return
	}

	h.addLinkStatusLinks(status)
	sharedAPI.RespondJSON(w, http.StatusOK, status)
}

// GetBatchCapabilityLinkStatus godoc
// @Summary Get link eligibility status for multiple domain capabilities
// @Description Batch check link eligibility for domain capabilities, optionally filtered by business domain
// @Tags enterprise-capabilities
// @Produce json
// @Param capabilityIds query string false "Comma-separated list of capability IDs"
// @Success 200 {object} easi_backend_internal_shared_api.CollectionResponse{data=[]easi_backend_internal_enterprisearchitecture_application_readmodels.CapabilityLinkStatusDTO}
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /domain-capabilities/enterprise-link-status [get]
func (h *EnterpriseCapabilityHandlers) GetBatchCapabilityLinkStatus(w http.ResponseWriter, r *http.Request) {
	capabilityIDs, ok := parseCapabilityIDs(w, r.URL.Query().Get("capabilityIds"))
	if !ok {
		return
	}

	statuses, err := h.readModels.Link.GetBatchLinkStatus(r.Context(), capabilityIDs)
	if err != nil {
		sharedAPI.HandleError(w, err)
		return
	}

	for i := range statuses {
		h.addLinkStatusLinks(&statuses[i])
	}

	sharedAPI.RespondCollection(w, http.StatusOK, statuses, nil)
}

func parseCapabilityIDs(w http.ResponseWriter, raw string) ([]string, bool) {
	if raw == "" {
		sharedAPI.RespondError(w, http.StatusBadRequest, nil, "capabilityIds query parameter is required")
		return nil, false
	}
	ids := splitAndTrim(raw)
	switch {
	case len(ids) == 0:
		sharedAPI.RespondError(w, http.StatusBadRequest, nil, "at least one capability ID is required")
		return nil, false
	case len(ids) > 100:
		sharedAPI.RespondError(w, http.StatusBadRequest, nil, "maximum 100 capability IDs allowed per request")
		return nil, false
	}
	return ids, true
}

func splitAndTrim(s string) []string {
	parts := make([]string, 0)
	for _, part := range strings.Split(s, ",") {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return parts
}

func (h *EnterpriseCapabilityHandlers) addLinkStatusLinks(status *readmodels.CapabilityLinkStatusDTO) {
	var linkedToID *string
	if status.LinkedTo != nil {
		linkedToID = &status.LinkedTo.ID
	}

	var blockingCapabilityID *string
	if status.BlockingCapability != nil {
		blockingCapabilityID = &status.BlockingCapability.ID
	}

	status.Links = h.hateoas.CapabilityLinkStatusLinks(LinkStatusParams{
		CapabilityID:  status.CapabilityID,
		Status:        string(status.Status),
		LinkedToID:    linkedToID,
		BlockingCapID: blockingCapabilityID,
		BlockingEcID:  status.BlockingEnterpriseCapID,
	})
}

func (h *EnterpriseCapabilityHandlers) getLinkOrNotFound(w http.ResponseWriter, r *http.Request, id string) *readmodels.EnterpriseCapabilityLinkDTO {
	return getOrNotFound(w, func() (*readmodels.EnterpriseCapabilityLinkDTO, error) {
		return h.readModels.Link.GetByID(r.Context(), id)
	}, "Link")
}
