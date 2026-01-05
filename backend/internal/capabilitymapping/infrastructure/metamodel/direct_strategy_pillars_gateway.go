package metamodel

import (
	"context"
	"database/sql"
	"encoding/json"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
)

type directStrategyPillarsGateway struct {
	db *database.TenantAwareDB
}

func NewDirectStrategyPillarsGateway(db *database.TenantAwareDB) StrategyPillarsGateway {
	return &directStrategyPillarsGateway{db: db}
}

func (g *directStrategyPillarsGateway) GetStrategyPillars(ctx context.Context) (*StrategyPillarsConfigDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	pillars, err := g.queryPillars(ctx, tenantID.Value())
	if err != nil {
		return nil, err
	}

	if len(pillars) == 0 {
		return DefaultStrategyPillarsConfig(), nil
	}

	return &StrategyPillarsConfigDTO{Pillars: pillars}, nil
}

func (g *directStrategyPillarsGateway) GetActivePillar(ctx context.Context, pillarID string) (*StrategyPillarDTO, error) {
	config, err := g.GetStrategyPillars(ctx)
	if err != nil {
		return nil, err
	}

	for _, pillar := range config.Pillars {
		if pillar.ID == pillarID && pillar.Active {
			return &pillar, nil
		}
	}

	return nil, nil
}

func (g *directStrategyPillarsGateway) InvalidateCache(tenantID string) {
}

func (g *directStrategyPillarsGateway) queryPillars(ctx context.Context, tenantID string) ([]StrategyPillarDTO, error) {
	query := `
		SELECT strategy_pillars
		FROM meta_model_configurations
		WHERE tenant_id = $1
	`

	var pillarsJSON []byte
	var notFound bool

	err := g.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		err := tx.QueryRowContext(ctx, query, tenantID).Scan(&pillarsJSON)
		if err != nil {
			if err == sql.ErrNoRows {
				notFound = true
				return nil
			}
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	if notFound || pillarsJSON == nil {
		return nil, nil
	}

	var pillars []StrategyPillarDTO
	if err := json.Unmarshal(pillarsJSON, &pillars); err != nil {
		return nil, err
	}

	return pillars, nil
}
