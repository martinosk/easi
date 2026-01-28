package api

import (
	"net/http"

	"easi/backend/internal/auth/infrastructure/session"
	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/infrastructure/metamodel"
	sharedAPI "easi/backend/internal/shared/api"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type StrategicFitAnalysisHandlers struct {
	analysisRM     *readmodels.StrategicFitAnalysisReadModel
	pillarsGateway metamodel.StrategyPillarsGateway
	sessionManager *session.SessionManager
}

func NewStrategicFitAnalysisHandlers(
	analysisRM *readmodels.StrategicFitAnalysisReadModel,
	pillarsGateway metamodel.StrategyPillarsGateway,
	sessionManager *session.SessionManager,
) *StrategicFitAnalysisHandlers {
	return &StrategicFitAnalysisHandlers{
		analysisRM:     analysisRM,
		pillarsGateway: pillarsGateway,
		sessionManager: sessionManager,
	}
}

type RealizationFitResponse struct {
	RealizationID                  string `json:"realizationId"`
	ComponentID                    string `json:"componentId"`
	ComponentName                  string `json:"componentName"`
	CapabilityID                   string `json:"capabilityId"`
	CapabilityName                 string `json:"capabilityName"`
	BusinessDomainID               string `json:"businessDomainId,omitempty"`
	BusinessDomainName             string `json:"businessDomainName,omitempty"`
	Importance                     int    `json:"importance"`
	ImportanceLabel                string `json:"importanceLabel"`
	ImportanceSourceCapabilityID   string `json:"importanceSourceCapabilityId,omitempty"`
	ImportanceSourceCapabilityName string `json:"importanceSourceCapabilityName,omitempty"`
	IsImportanceInherited          bool   `json:"isImportanceInherited"`
	ImportanceRationale            string `json:"importanceRationale,omitempty"`
	FitScore                       int    `json:"fitScore"`
	FitScoreLabel                  string `json:"fitScoreLabel"`
	Gap                            int    `json:"gap"`
	FitRationale                   string `json:"fitRationale,omitempty"`
	Category                       string `json:"category"`
}

type StrategicFitSummaryResponse struct {
	TotalRealizations  int     `json:"totalRealizations"`
	ScoredRealizations int     `json:"scoredRealizations"`
	LiabilityCount     int     `json:"liabilityCount"`
	ConcernCount       int     `json:"concernCount"`
	AlignedCount       int     `json:"alignedCount"`
	AverageGap         float64 `json:"averageGap"`
}

type StrategicFitAnalysisResponse struct {
	PillarID    string                      `json:"pillarId"`
	PillarName  string                      `json:"pillarName"`
	Summary     StrategicFitSummaryResponse `json:"summary"`
	Liabilities []RealizationFitResponse    `json:"liabilities"`
	Concerns    []RealizationFitResponse    `json:"concerns"`
	Aligned     []RealizationFitResponse    `json:"aligned"`
	Links       map[string]string           `json:"_links,omitempty"`
}

// GetStrategicFitAnalysis godoc
// @Summary Get strategic fit analysis for a pillar
// @Description Analyzes the gap between capability importance and application fit scores for a strategic pillar. Requires enterprise-arch:read permission (Admin, Architect, or Stakeholder role).
// @Tags strategic-fit-analysis
// @Accept json
// @Produce json
// @Param pillarId path string true "Strategy Pillar ID"
// @Success 200 {object} StrategicFitAnalysisResponse
// @Failure 400 {object} sharedAPI.ErrorResponse "Invalid pillar ID format or fit scoring not enabled"
// @Failure 401 {object} sharedAPI.ErrorResponse "Authentication required"
// @Failure 403 {object} sharedAPI.ErrorResponse "Insufficient permissions - requires enterprise-arch:read"
// @Failure 404 {object} sharedAPI.ErrorResponse "Strategy pillar not found"
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Security ApiKeyAuth
// @Router /strategic-fit-analysis/{pillarId} [get]
func (h *StrategicFitAnalysisHandlers) GetStrategicFitAnalysis(w http.ResponseWriter, r *http.Request) {
	_, err := h.sessionManager.LoadAuthenticatedSession(r.Context())
	if err != nil {
		sharedAPI.RespondError(w, http.StatusUnauthorized, err, "Authentication required")
		return
	}

	pillarID := chi.URLParam(r, "pillarId")

	if _, err := uuid.Parse(pillarID); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid pillar ID format")
		return
	}

	pillar, err := h.pillarsGateway.GetActivePillar(r.Context(), pillarID)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to get pillar")
		return
	}

	if pillar == nil {
		sharedAPI.RespondError(w, http.StatusNotFound, nil, "Strategy pillar not found")
		return
	}

	if !pillar.FitScoringEnabled {
		sharedAPI.RespondError(w, http.StatusBadRequest, nil, "Fit scoring is not enabled for this pillar")
		return
	}

	analysis, err := h.analysisRM.GetStrategicFitAnalysis(r.Context(), pillarID, pillar.Name)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to get strategic fit analysis")
		return
	}

	response := h.buildAnalysisResponse(analysis)
	sharedAPI.RespondJSON(w, http.StatusOK, response)
}

func (h *StrategicFitAnalysisHandlers) buildAnalysisResponse(dto *readmodels.StrategicFitAnalysisDTO) StrategicFitAnalysisResponse {
	return StrategicFitAnalysisResponse{
		PillarID:   dto.PillarID,
		PillarName: dto.PillarName,
		Summary: StrategicFitSummaryResponse{
			TotalRealizations:  dto.Summary.TotalRealizations,
			ScoredRealizations: dto.Summary.ScoredRealizations,
			LiabilityCount:     dto.Summary.LiabilityCount,
			ConcernCount:       dto.Summary.ConcernCount,
			AlignedCount:       dto.Summary.AlignedCount,
			AverageGap:         dto.Summary.AverageGap,
		},
		Liabilities: h.buildRealizationFitResponses(dto.Liabilities),
		Concerns:    h.buildRealizationFitResponses(dto.Concerns),
		Aligned:     h.buildRealizationFitResponses(dto.Aligned),
		Links: map[string]string{
			"self": "/api/v1/strategic-fit-analysis/" + dto.PillarID,
		},
	}
}

func (h *StrategicFitAnalysisHandlers) buildRealizationFitResponses(dtos []readmodels.RealizationFitDTO) []RealizationFitResponse {
	responses := make([]RealizationFitResponse, len(dtos))
	for i, dto := range dtos {
		responses[i] = RealizationFitResponse{
			RealizationID:                  dto.RealizationID,
			ComponentID:                    dto.ComponentID,
			ComponentName:                  dto.ComponentName,
			CapabilityID:                   dto.CapabilityID,
			CapabilityName:                 dto.CapabilityName,
			BusinessDomainID:               dto.BusinessDomainID,
			BusinessDomainName:             dto.BusinessDomainName,
			Importance:                     dto.Importance,
			ImportanceLabel:                dto.ImportanceLabel,
			ImportanceSourceCapabilityID:   dto.ImportanceSourceCapabilityID,
			ImportanceSourceCapabilityName: dto.ImportanceSourceCapabilityName,
			IsImportanceInherited:          dto.IsImportanceInherited,
			ImportanceRationale:            dto.ImportanceRationale,
			FitScore:                       dto.FitScore,
			FitScoreLabel:                  dto.FitScoreLabel,
			Gap:                            dto.Gap,
			FitRationale:                   dto.FitRationale,
			Category:                       dto.Category,
		}
	}
	return responses
}
