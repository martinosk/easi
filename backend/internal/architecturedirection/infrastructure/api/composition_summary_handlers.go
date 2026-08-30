package api

import (
	"context"
	"net/http"
	"sort"

	appservices "easi/backend/internal/architecturedirection/application/services"
	domainservices "easi/backend/internal/architecturedirection/domain/services"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/types"
)

type CompositionSummarySource interface {
	SummariesForAll(ctx context.Context) (appservices.CompositionSummaries, error)
}

type CompositionSummaryDTO struct {
	EnterpriseCapabilityID string      `json:"enterpriseCapabilityId"`
	SourceCount            int         `json:"sourceCount"`
	IncludedCount          int         `json:"includedCount"`
	CarvedOutCount         int         `json:"carvedOutCount"`
	DomainCount            int         `json:"domainCount"`
	HasActiveDirection     bool        `json:"hasActiveDirection"`
	DirectionStatus        string      `json:"directionStatus,omitempty"`
	Links                  types.Links `json:"_links,omitempty"`
}

type CompositionSummariesResponse struct {
	Data  []CompositionSummaryDTO `json:"data"`
	Links types.Links             `json:"_links,omitempty"`
}

type CompositionSummaryHandlers struct {
	compositions CompositionSummarySource
	hateoas      *sharedAPI.HATEOASLinks
}

func NewCompositionSummaryHandlers(compositions CompositionSummarySource, hateoas *sharedAPI.HATEOASLinks) *CompositionSummaryHandlers {
	return &CompositionSummaryHandlers{compositions: compositions, hateoas: hateoas}
}

// GetCompositionSummaries godoc
// @Summary List composition summaries of all enterprise capabilities
// @Description One summary per active enterprise capability: source, included, carved-out and domain counts derived from the active direction, plus the direction status. Enterprise capabilities without an active direction report zero counts. Every item links to its enterprise capability, its composition and its direction.
// @Tags architecturedirection
// @Produce json
// @Success 200 {object} CompositionSummariesResponse
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Security CookieAuth
// @Router /enterprise-capability-compositions [get]
func (h *CompositionSummaryHandlers) GetCompositionSummaries(w http.ResponseWriter, r *http.Request) {
	summaries, err := h.compositions.SummariesForAll(r.Context())
	if err != nil {
		sharedAPI.HandleError(w, err)
		return
	}
	data := make([]CompositionSummaryDTO, 0, len(summaries.EnterpriseCapabilityIDs))
	for _, ecID := range summaries.EnterpriseCapabilityIDs {
		data = append(data, h.summary(ecID, summaries.Counts[ecID], summaries.Statuses[ecID]))
	}
	sort.Slice(data, func(i, j int) bool { return data[i].EnterpriseCapabilityID < data[j].EnterpriseCapabilityID })
	sharedAPI.RespondJSON(w, http.StatusOK, CompositionSummariesResponse{
		Data:  data,
		Links: types.Links{"self": h.hateoas.Get("/enterprise-capability-compositions")},
	})
}

func (h *CompositionSummaryHandlers) summary(ecID string, counts domainservices.CompositionCounts, status string) CompositionSummaryDTO {
	base := "/enterprise-capabilities/" + ecID
	return CompositionSummaryDTO{
		EnterpriseCapabilityID: ecID,
		SourceCount:            counts.SourceCount,
		IncludedCount:          counts.IncludedCount,
		CarvedOutCount:         counts.CarvedOutCount,
		DomainCount:            counts.DomainCount,
		HasActiveDirection:     status != "",
		DirectionStatus:        status,
		Links: types.Links{
			"x-enterprise-capability": h.hateoas.Get(base),
			"x-composition":           h.hateoas.Get(base + "/composition"),
			"x-direction":             h.hateoas.Get(base + "/direction"),
		},
	}
}
