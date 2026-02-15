package readmodels

import (
	"context"
	"database/sql"

	"easi/backend/internal/infrastructure/database"
)

type StrategyPillarCacheDTO struct {
	ID                string `json:"id"`
	TenantID          string `json:"tenantId"`
	Name              string `json:"name"`
	Description       string `json:"description"`
	Active            bool   `json:"active"`
	FitScoringEnabled bool   `json:"fitScoringEnabled"`
	FitCriteria       string `json:"fitCriteria"`
	FitType           string `json:"fitType"`
}

type StrategyPillarCacheReadModel struct {
	db *database.TenantAwareDB
}

func NewStrategyPillarCacheReadModel(db *database.TenantAwareDB) *StrategyPillarCacheReadModel {
	return &StrategyPillarCacheReadModel{db: db}
}

func (rm *StrategyPillarCacheReadModel) Insert(ctx context.Context, dto StrategyPillarCacheDTO) error {
	query := `
		INSERT INTO capabilitymapping.cm_strategy_pillar_cache (
			id, tenant_id, name, description, active,
			fit_scoring_enabled, fit_criteria, fit_type
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id, tenant_id) 
		DO UPDATE SET
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			active = EXCLUDED.active,
			fit_scoring_enabled = EXCLUDED.fit_scoring_enabled,
			fit_criteria = EXCLUDED.fit_criteria,
			fit_type = EXCLUDED.fit_type
	`

	_, err := rm.db.ExecContext(ctx, query,
		dto.ID, dto.TenantID, dto.Name, dto.Description, dto.Active,
		dto.FitScoringEnabled, dto.FitCriteria, dto.FitType,
	)
	return err
}

func (rm *StrategyPillarCacheReadModel) Delete(ctx context.Context, pillarID string) error {
	query := `DELETE FROM capabilitymapping.cm_strategy_pillar_cache WHERE id = $1`

	_, err := rm.db.ExecContext(ctx, query, pillarID)
	return err
}

func (rm *StrategyPillarCacheReadModel) GetAll(ctx context.Context) ([]StrategyPillarCacheDTO, error) {
	query := `
		SELECT id, tenant_id, name, description, active,
		       fit_scoring_enabled, fit_criteria, fit_type
		FROM capabilitymapping.cm_strategy_pillar_cache
		ORDER BY name
	`

	var pillars []StrategyPillarCacheDTO

	err := rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, query)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var dto StrategyPillarCacheDTO
			var fitCriteria, fitType sql.NullString

			err := rows.Scan(
				&dto.ID, &dto.TenantID, &dto.Name, &dto.Description, &dto.Active,
				&dto.FitScoringEnabled, &fitCriteria, &fitType,
			)
			if err != nil {
				return err
			}

			dto.FitCriteria = fitCriteria.String
			dto.FitType = fitType.String
			pillars = append(pillars, dto)
		}

		return rows.Err()
	})

	if err != nil {
		return nil, err
	}

	return pillars, nil
}

func (rm *StrategyPillarCacheReadModel) GetActivePillar(ctx context.Context, pillarID string) (*StrategyPillarCacheDTO, error) {
	query := `
		SELECT id, tenant_id, name, description, active,
		       fit_scoring_enabled, fit_criteria, fit_type
		FROM capabilitymapping.cm_strategy_pillar_cache
		WHERE id = $1 AND active = true
	`

	var dto StrategyPillarCacheDTO
	var fitCriteria, fitType sql.NullString

	err := rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx, query, pillarID).Scan(
			&dto.ID, &dto.TenantID, &dto.Name, &dto.Description, &dto.Active,
			&dto.FitScoringEnabled, &fitCriteria, &fitType,
		)
	})

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	dto.FitCriteria = fitCriteria.String
	dto.FitType = fitType.String

	return &dto, nil
}
