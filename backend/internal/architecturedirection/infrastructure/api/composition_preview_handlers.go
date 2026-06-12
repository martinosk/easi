package api

import (
	"context"
	"net/http"

	sharedAPI "easi/backend/internal/shared/api"
)

type CarvedOutByDTO struct {
	EnterpriseCapabilityID   string `json:"enterpriseCapabilityId"`
	EnterpriseCapabilityName string `json:"enterpriseCapabilityName"`
}

type PreviewIncludedCapabilityDTO struct {
	CapabilityID       string          `json:"capabilityId"`
	Name               string          `json:"name"`
	Level              string          `json:"level"`
	BusinessDomainID   *string         `json:"businessDomainId"`
	BusinessDomainName *string         `json:"businessDomainName"`
	Role               string          `json:"role"`
	CarvedOutBy        *CarvedOutByDTO `json:"carvedOutBy"`
}

type ConflictingECDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type SourceEligibilityDTO struct {
	CapabilityID                    string            `json:"capabilityId"`
	Eligible                        bool              `json:"eligible"`
	IneligibilityReason             *string           `json:"ineligibilityReason"`
	ConflictingEnterpriseCapability *ConflictingECDTO `json:"conflictingEnterpriseCapability"`
}

type CompositionPreviewMetaDTO struct {
	SourceCount    int `json:"sourceCount"`
	IncludedCount  int `json:"includedCount"`
	CarvedOutCount int `json:"carvedOutCount"`
}

type CompositionPreviewData struct {
	IncludedCapabilities []PreviewIncludedCapabilityDTO
	SourceEligibility    []SourceEligibilityDTO
	Meta                 CompositionPreviewMetaDTO
}

type CompositionPreviewProvider interface {
	PreviewComposition(ctx context.Context, enterpriseCapabilityID string, sourceCapabilityIDs []string) (*CompositionPreviewData, error)
}

type CompositionPreviewRequest struct {
	SourceCapabilityIDs []string `json:"sourceCapabilityIds"`
}

type CompositionPreviewResponse struct {
	IncludedCapabilities []PreviewIncludedCapabilityDTO `json:"includedCapabilities"`
	SourceEligibility    []SourceEligibilityDTO         `json:"sourceEligibility"`
	Meta                 CompositionPreviewMetaDTO      `json:"meta"`
	Links                sharedAPI.Links                `json:"_links,omitempty"`
}

type CompositionPreviewHandlers struct {
	provider CompositionPreviewProvider
	hateoas  *sharedAPI.HATEOASLinks
}

func NewCompositionPreviewHandlers(provider CompositionPreviewProvider, hateoas *sharedAPI.HATEOASLinks) *CompositionPreviewHandlers {
	return &CompositionPreviewHandlers{provider: provider, hateoas: hateoas}
}

// PreviewComposition godoc
// @Summary Preview the composition for a proposed source set
// @Description Stateless preview: resolves what the enterprise capability's composition would be for the given source set without persisting anything (R2 carve-out preview and R1 eligibility pre-flight).
// @Tags architecturedirection
// @Accept json
// @Produce json
// @Security CookieAuth
// @Param id path string true "Enterprise capability ID"
// @Param body body CompositionPreviewRequest true "Proposed source capability IDs"
// @Success 200 {object} CompositionPreviewResponse
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /enterprise-capabilities/{id}/direction/composition-preview [post]
func (h *CompositionPreviewHandlers) PreviewComposition(w http.ResponseWriter, r *http.Request) {
	ecID := sharedAPI.GetPathParam(r, "id")
	req, ok := sharedAPI.DecodeRequestOrFail[CompositionPreviewRequest](w, r)
	if !ok {
		return
	}
	sourceIDs := req.SourceCapabilityIDs
	if sourceIDs == nil {
		sourceIDs = []string{}
	}
	preview, err := h.provider.PreviewComposition(r.Context(), ecID, sourceIDs)
	if err != nil {
		sharedAPI.HandleError(w, err)
		return
	}
	if preview == nil {
		sharedAPI.RespondError(w, http.StatusNotFound, nil, "Enterprise capability not found")
		return
	}
	sharedAPI.RespondJSON(w, http.StatusOK, CompositionPreviewResponse{
		IncludedCapabilities: preview.IncludedCapabilities,
		SourceEligibility:    preview.SourceEligibility,
		Meta:                 preview.Meta,
		Links: sharedAPI.Links{
			"self": h.hateoas.Post(directionResourcePath(ecID) + "/composition-preview"),
		},
	})
}
