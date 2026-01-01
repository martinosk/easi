package readmodels

import (
	"context"
	"database/sql"

	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
)

type FitComparisonDTO struct {
	PillarID        string `json:"pillarId"`
	PillarName      string `json:"pillarName"`
	FitScore        int    `json:"fitScore"`
	FitScoreLabel   string `json:"fitScoreLabel"`
	Importance      int    `json:"importance"`
	ImportanceLabel string `json:"importanceLabel"`
	Gap             int    `json:"gap"`
	Category        string `json:"category"`
	FitRationale    string `json:"fitRationale,omitempty"`
}

type ComponentFitComparisonReadModel struct {
	db *database.TenantAwareDB
}

func NewComponentFitComparisonReadModel(db *database.TenantAwareDB) *ComponentFitComparisonReadModel {
	return &ComponentFitComparisonReadModel{db: db}
}

func (rm *ComponentFitComparisonReadModel) GetByComponentAndCapability(
	ctx context.Context,
	componentID string,
	capabilityID string,
	businessDomainID string,
) ([]FitComparisonDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var results []FitComparisonDTO

	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		query := `
			SELECT
				afs.pillar_id,
				afs.pillar_name,
				afs.score as fit_score,
				afs.score_label as fit_score_label,
				COALESCE(si.importance, 0) as importance,
				COALESCE(si.importance_label, '') as importance_label,
				COALESCE(afs.rationale, '') as fit_rationale
			FROM application_fit_scores afs
			LEFT JOIN strategy_importance si ON afs.tenant_id = si.tenant_id
				AND afs.pillar_id = si.pillar_id
				AND si.capability_id = $3
				AND si.business_domain_id = $4
			WHERE afs.tenant_id = $1
				AND afs.component_id = $2
			ORDER BY afs.pillar_name
		`

		rows, err := tx.QueryContext(ctx, query, tenantID.Value(), componentID, capabilityID, businessDomainID)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var dto FitComparisonDTO
			if err := rows.Scan(
				&dto.PillarID,
				&dto.PillarName,
				&dto.FitScore,
				&dto.FitScoreLabel,
				&dto.Importance,
				&dto.ImportanceLabel,
				&dto.FitRationale,
			); err != nil {
				return err
			}

			if dto.Importance > 0 && dto.FitScore > 0 {
				dto.Gap = dto.Importance - dto.FitScore
				dto.Category = valueobjects.CategorizeGap(dto.Gap, dto.Importance).String()
			}

			results = append(results, dto)
		}
		return rows.Err()
	})

	if err != nil {
		return nil, err
	}

	return results, nil
}
