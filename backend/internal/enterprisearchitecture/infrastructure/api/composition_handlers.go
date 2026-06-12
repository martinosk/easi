package api

import (
	"context"
	"net/http"
	"strconv"

	appservices "easi/backend/internal/enterprisearchitecture/application/services"
	domainservices "easi/backend/internal/enterprisearchitecture/domain/services"
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/types"
)

const defaultSourceCandidatesLimit = 20

type CompositionQueries interface {
	CompositionForEC(ctx context.Context, enterpriseCapabilityID string) (appservices.CompositionResult, error)
	SourceCandidates(ctx context.Context, query appservices.SourceCandidatesQuery) (appservices.SourceCandidatesResult, error)
}

type CarvedOutByDTO struct {
	EnterpriseCapabilityID   string `json:"enterpriseCapabilityId"`
	EnterpriseCapabilityName string `json:"enterpriseCapabilityName"`
}

type IncludedCapabilityItemDTO struct {
	CapabilityID       string          `json:"capabilityId"`
	Name               string          `json:"name"`
	Level              string          `json:"level"`
	BusinessDomainID   *string         `json:"businessDomainId"`
	BusinessDomainName *string         `json:"businessDomainName"`
	Role               string          `json:"role"`
	CarvedOutBy        *CarvedOutByDTO `json:"carvedOutBy"`
	Links              types.Links     `json:"_links,omitempty"`
}

type CompositionDomainGroupDTO struct {
	BusinessDomainID   *string                     `json:"businessDomainId"`
	BusinessDomainName *string                     `json:"businessDomainName"`
	Items              []IncludedCapabilityItemDTO `json:"items"`
}

type CompositionMetaDTO struct {
	SourceCount    int `json:"sourceCount"`
	IncludedCount  int `json:"includedCount"`
	CarvedOutCount int `json:"carvedOutCount"`
	DomainCount    int `json:"domainCount"`
}

type CompositionResponseDTO struct {
	Data  []CompositionDomainGroupDTO `json:"data"`
	Meta  CompositionMetaDTO          `json:"meta"`
	Links types.Links                 `json:"_links,omitempty"`
}

type ConflictingECDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type SourceCandidateDTO struct {
	CapabilityID                    string            `json:"capabilityId"`
	Name                            string            `json:"name"`
	Level                           string            `json:"level"`
	ParentID                        *string           `json:"parentId"`
	BusinessDomainID                *string           `json:"businessDomainId"`
	BusinessDomainName              *string           `json:"businessDomainName"`
	Eligible                        bool              `json:"eligible"`
	IneligibilityReason             *string           `json:"ineligibilityReason"`
	ConflictingEnterpriseCapability *ConflictingECDTO `json:"conflictingEnterpriseCapability"`
	Links                           types.Links       `json:"_links,omitempty"`
}

type SourceCandidatesPaginationDTO struct {
	HasMore bool   `json:"hasMore"`
	Limit   int    `json:"limit"`
	Cursor  string `json:"cursor"`
}

type SourceCandidatesResponseDTO struct {
	Data       []SourceCandidateDTO          `json:"data"`
	Pagination SourceCandidatesPaginationDTO `json:"pagination"`
	Links      types.Links                   `json:"_links,omitempty"`
}

type CompositionHandlers struct {
	queries      CompositionQueries
	capabilities EnterpriseCapabilityQueries
	hateoas      *EnterpriseArchLinks
}

func NewCompositionHandlers(queries CompositionQueries, capabilities EnterpriseCapabilityQueries, hateoas *EnterpriseArchLinks) *CompositionHandlers {
	return &CompositionHandlers{queries: queries, capabilities: capabilities, hateoas: hateoas}
}

// GetComposition godoc
// @Summary Get the composition of an enterprise capability
// @Description Returns every domain capability included via the active direction's sources and their subtrees, grouped by business domain, with per-item role classification and carve-out attribution (R2).
// @Tags enterprisearchitecture
// @Produce json
// @Security CookieAuth
// @Param id path string true "Enterprise capability ID"
// @Success 200 {object} CompositionResponseDTO
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /enterprise-capabilities/{id}/composition [get]
func (h *CompositionHandlers) GetComposition(w http.ResponseWriter, r *http.Request) {
	ecID := sharedAPI.GetPathParam(r, "id")
	capability, err := h.capabilities.GetByID(r.Context(), ecID)
	if err != nil {
		sharedAPI.HandleError(w, err)
		return
	}
	if capability == nil {
		sharedAPI.RespondError(w, http.StatusNotFound, nil, "Enterprise capability not found")
		return
	}
	composition, err := h.queries.CompositionForEC(r.Context(), ecID)
	if err != nil {
		sharedAPI.HandleError(w, err)
		return
	}

	actor, _ := sharedctx.GetActor(r.Context())
	sharedAPI.RespondJSON(w, http.StatusOK, CompositionResponseDTO{
		Data: h.groupByDomain(ecID, composition, actor),
		Meta: CompositionMetaDTO{
			SourceCount:    composition.Counts.SourceCount,
			IncludedCount:  composition.Counts.IncludedCount,
			CarvedOutCount: composition.Counts.CarvedOutCount,
			DomainCount:    composition.Counts.DomainCount,
		},
		Links: h.compositionLinks(ecID, composition.HasActiveDirection, actor),
	})
}

func (h *CompositionHandlers) groupByDomain(ecID string, composition appservices.CompositionResult, actor sharedctx.Actor) []CompositionDomainGroupDTO {
	groups := []CompositionDomainGroupDTO{}
	indexByDomain := map[string]int{}
	for _, resolved := range composition.Resolved {
		key := resolved.Node.BusinessDomainID
		index, exists := indexByDomain[key]
		if !exists {
			groups = append(groups, CompositionDomainGroupDTO{
				BusinessDomainID:   nilIfEmpty(resolved.Node.BusinessDomainID),
				BusinessDomainName: nilIfEmpty(resolved.Node.BusinessDomainName),
				Items:              []IncludedCapabilityItemDTO{},
			})
			index = len(groups) - 1
			indexByDomain[key] = index
		}
		groups[index].Items = append(groups[index].Items, h.includedItem(ecID, composition.DirectionStatus, resolved, actor))
	}
	return groups
}

func (h *CompositionHandlers) includedItem(ecID, directionStatus string, resolved domainservices.ResolvedCapability, actor sharedctx.Actor) IncludedCapabilityItemDTO {
	item := IncludedCapabilityItemDTO{
		CapabilityID:       resolved.Node.ID,
		Name:               resolved.Node.Name,
		Level:              resolved.Node.Level,
		BusinessDomainID:   nilIfEmpty(resolved.Node.BusinessDomainID),
		BusinessDomainName: nilIfEmpty(resolved.Node.BusinessDomainName),
		Role:               string(resolved.Role),
		Links: types.Links{
			"self": h.hateoas.Get("/capabilities/" + resolved.Node.ID),
		},
	}
	if canExcludeSource(resolved.Role, directionStatus, actor) {
		item.Links["x-exclude"] = h.hateoas.Del("/enterprise-capabilities/" + ecID + "/direction/sources/" + resolved.Node.ID)
	}
	if resolved.CarvedOutBy != nil {
		item.CarvedOutBy = &CarvedOutByDTO{
			EnterpriseCapabilityID:   resolved.CarvedOutBy.EnterpriseCapabilityID,
			EnterpriseCapabilityName: resolved.CarvedOutBy.EnterpriseCapabilityName,
		}
		item.Links["x-owning-ec"] = h.hateoas.Get("/enterprise-capabilities/" + resolved.CarvedOutBy.EnterpriseCapabilityID)
	}
	return item
}

func canExcludeSource(role domainservices.CompositionRole, directionStatus string, actor sharedctx.Actor) bool {
	if role != domainservices.RoleSource || directionStatus != "draft" {
		return false
	}
	return actor.CanWrite("architecture-direction")
}

func (h *CompositionHandlers) compositionLinks(ecID string, hasActiveDirection bool, actor sharedctx.Actor) types.Links {
	base := "/enterprise-capabilities/" + ecID
	links := types.Links{
		"self":        h.hateoas.Get(base + "/composition"),
		"up":          h.hateoas.Get(base),
		"x-direction": h.hateoas.Get(base + "/direction"),
	}
	if !hasActiveDirection && actor.CanWrite("architecture-direction") {
		links["x-capture-direction"] = h.hateoas.Post(base + "/direction")
	}
	return links
}

// GetSourceCandidates godoc
// @Summary Search domain capabilities as direction source candidates
// @Description Case-insensitive name search over domain capabilities with per-candidate R1 eligibility against active directions on other enterprise capabilities.
// @Tags enterprisearchitecture
// @Produce json
// @Security CookieAuth
// @Param q query string true "Search term (case-insensitive substring match on capability name)"
// @Param ecId query string true "Enterprise capability the sources are being searched for"
// @Param domainId query string false "Filter to capabilities in this business domain"
// @Param limit query int false "Max results to return (default 20)"
// @Param cursor query string false "Opaque pagination cursor (reserved; not yet populated)"
// @Success 200 {object} SourceCandidatesResponseDTO
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /capabilities/source-candidates [get]
func (h *CompositionHandlers) GetSourceCandidates(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	search := params.Get("q")
	ecID := params.Get("ecId")
	if search == "" || ecID == "" {
		sharedAPI.RespondJSON(w, http.StatusBadRequest, sharedAPI.ErrorResponse{
			Error:   "BadRequest",
			Message: "q and ecId are required",
		})
		return
	}
	limit := defaultSourceCandidatesLimit
	if raw := params.Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	result, err := h.queries.SourceCandidates(r.Context(), appservices.SourceCandidatesQuery{
		EnterpriseCapabilityID: ecID,
		Search:                 search,
		BusinessDomainID:       params.Get("domainId"),
		Limit:                  limit,
	})
	if err != nil {
		sharedAPI.HandleError(w, err)
		return
	}

	data := make([]SourceCandidateDTO, len(result.Candidates))
	for i, candidate := range result.Candidates {
		data[i] = h.sourceCandidateDTO(candidate)
	}
	sharedAPI.RespondJSON(w, http.StatusOK, SourceCandidatesResponseDTO{
		Data:       data,
		Pagination: SourceCandidatesPaginationDTO{HasMore: result.HasMore, Limit: limit, Cursor: ""},
		Links: types.Links{
			"self": h.hateoas.Get("/capabilities/source-candidates?" + params.Encode()),
		},
	})
}

func (h *CompositionHandlers) sourceCandidateDTO(candidate appservices.SourceCandidate) SourceCandidateDTO {
	dto := SourceCandidateDTO{
		CapabilityID:        candidate.Node.ID,
		Name:                candidate.Node.Name,
		Level:               candidate.Node.Level,
		ParentID:            nilIfEmpty(candidate.Node.ParentID),
		BusinessDomainID:    nilIfEmpty(candidate.Node.BusinessDomainID),
		BusinessDomainName:  nilIfEmpty(candidate.Node.BusinessDomainName),
		Eligible:            candidate.Eligible,
		IneligibilityReason: candidate.IneligibilityReason,
		Links: types.Links{
			"self": h.hateoas.Get("/capabilities/" + candidate.Node.ID),
		},
	}
	if candidate.ConflictingEnterpriseCapability != nil {
		dto.ConflictingEnterpriseCapability = &ConflictingECDTO{
			ID:   candidate.ConflictingEnterpriseCapability.EnterpriseCapabilityID,
			Name: candidate.ConflictingEnterpriseCapability.EnterpriseCapabilityName,
		}
		dto.Links["x-conflicting-ec"] = h.hateoas.Get("/enterprise-capabilities/" + candidate.ConflictingEnterpriseCapability.EnterpriseCapabilityID)
	}
	return dto
}

func nilIfEmpty(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
