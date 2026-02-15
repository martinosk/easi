package readmodels

import (
	"context"
	"database/sql"

	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
)

type RealizationFitDTO struct {
	RealizationID                  string `json:"realizationId"`
	ComponentID                    string `json:"componentId"`
	ComponentName                  string `json:"componentName"`
	CapabilityID                   string `json:"capabilityId"`
	CapabilityName                 string `json:"capabilityName"`
	BusinessDomainID               string `json:"businessDomainId"`
	BusinessDomainName             string `json:"businessDomainName"`
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

type StrategicFitSummaryDTO struct {
	TotalRealizations  int     `json:"totalRealizations"`
	ScoredRealizations int     `json:"scoredRealizations"`
	LiabilityCount     int     `json:"liabilityCount"`
	ConcernCount       int     `json:"concernCount"`
	AlignedCount       int     `json:"alignedCount"`
	AverageGap         float64 `json:"averageGap"`
}

type StrategicFitAnalysisDTO struct {
	PillarID    string                 `json:"pillarId"`
	PillarName  string                 `json:"pillarName"`
	Summary     StrategicFitSummaryDTO `json:"summary"`
	Liabilities []RealizationFitDTO    `json:"liabilities"`
	Concerns    []RealizationFitDTO    `json:"concerns"`
	Aligned     []RealizationFitDTO    `json:"aligned"`
}

type StrategicFitAnalysisReadModel struct {
	db *database.TenantAwareDB
}

func NewStrategicFitAnalysisReadModel(db *database.TenantAwareDB) *StrategicFitAnalysisReadModel {
	return &StrategicFitAnalysisReadModel{db: db}
}

func (rm *StrategicFitAnalysisReadModel) GetStrategicFitAnalysis(ctx context.Context, pillarID, pillarName string) (*StrategicFitAnalysisDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	results, err := rm.queryRealizationFits(ctx, tenantID.Value(), pillarID)
	if err != nil {
		return nil, err
	}

	return rm.categorizeResults(pillarID, pillarName, results), nil
}

func (rm *StrategicFitAnalysisReadModel) queryRealizationFits(ctx context.Context, tenantID, pillarID string) ([]RealizationFitDTO, error) {
	query := `
		SELECT
			r.id as realization_id,
			r.component_id,
			r.component_name,
			r.capability_id,
			c.name as capability_name,
			COALESCE(cmbd.business_domain_id, '') as business_domain_id,
			COALESCE(cmbd.business_domain_name, '') as business_domain_name,
			COALESCE(eci.effective_importance, 0) as importance,
			COALESCE(eci.importance_label, '') as importance_label,
			COALESCE(eci.source_capability_id, '') as source_capability_id,
			COALESCE(eci.source_capability_name, '') as source_capability_name,
			COALESCE(eci.is_inherited, false) as is_inherited,
			COALESCE(eci.rationale, '') as importance_rationale,
			COALESCE(afs.score, 0) as fit_score,
			COALESCE(afs.score_label, '') as fit_score_label,
			COALESCE(afs.rationale, '') as fit_rationale
		FROM capability_realizations r
		JOIN capabilities c ON r.tenant_id = c.tenant_id AND r.capability_id = c.id
		LEFT JOIN cm_effective_business_domain cmbd ON r.tenant_id = cmbd.tenant_id AND r.capability_id = cmbd.capability_id
		LEFT JOIN effective_capability_importance eci ON r.tenant_id = eci.tenant_id
			AND r.capability_id = eci.capability_id
			AND cmbd.business_domain_id = eci.business_domain_id
			AND eci.pillar_id = $2
		LEFT JOIN application_fit_scores afs ON r.tenant_id = afs.tenant_id
			AND r.component_id = afs.component_id
			AND afs.pillar_id = $2
		WHERE r.tenant_id = $1
			AND r.origin = 'Direct'
			AND (eci.pillar_id = $2 OR afs.pillar_id = $2)
		ORDER BY
			CASE WHEN eci.effective_importance IS NOT NULL AND afs.score IS NOT NULL
				THEN eci.effective_importance - afs.score
				ELSE 0
			END DESC,
			c.name ASC
	`

	var results []RealizationFitDTO
	err := rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, query, tenantID, pillarID)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var dto RealizationFitDTO
			if err := rows.Scan(
				&dto.RealizationID,
				&dto.ComponentID,
				&dto.ComponentName,
				&dto.CapabilityID,
				&dto.CapabilityName,
				&dto.BusinessDomainID,
				&dto.BusinessDomainName,
				&dto.Importance,
				&dto.ImportanceLabel,
				&dto.ImportanceSourceCapabilityID,
				&dto.ImportanceSourceCapabilityName,
				&dto.IsImportanceInherited,
				&dto.ImportanceRationale,
				&dto.FitScore,
				&dto.FitScoreLabel,
				&dto.FitRationale,
			); err != nil {
				return err
			}
			results = append(results, dto)
		}
		return rows.Err()
	})

	return results, err
}

func (rm *StrategicFitAnalysisReadModel) categorizeResults(pillarID, pillarName string, results []RealizationFitDTO) *StrategicFitAnalysisDTO {
	analysis := &StrategicFitAnalysisDTO{
		PillarID:    pillarID,
		PillarName:  pillarName,
		Liabilities: []RealizationFitDTO{},
		Concerns:    []RealizationFitDTO{},
		Aligned:     []RealizationFitDTO{},
	}

	var totalGap int
	var scoredCount int

	for i := range results {
		dto := &results[i]

		if dto.Importance > 0 && dto.FitScore > 0 {
			dto.Gap = dto.Importance - dto.FitScore
			scoredCount++
			totalGap += dto.Gap

			dto.Category = rm.categorizeGap(dto.Gap, dto.Importance)

			switch dto.Category {
			case "liability":
				analysis.Liabilities = append(analysis.Liabilities, *dto)
			case "concern":
				analysis.Concerns = append(analysis.Concerns, *dto)
			case "aligned":
				analysis.Aligned = append(analysis.Aligned, *dto)
			}
		}
	}

	analysis.Summary = StrategicFitSummaryDTO{
		TotalRealizations:  len(results),
		ScoredRealizations: scoredCount,
		LiabilityCount:     len(analysis.Liabilities),
		ConcernCount:       len(analysis.Concerns),
		AlignedCount:       len(analysis.Aligned),
	}

	if scoredCount > 0 {
		analysis.Summary.AverageGap = float64(totalGap) / float64(scoredCount)
	}

	return analysis
}

func (rm *StrategicFitAnalysisReadModel) categorizeGap(gap, importance int) string {
	return valueobjects.CategorizeGap(gap, importance).String()
}
