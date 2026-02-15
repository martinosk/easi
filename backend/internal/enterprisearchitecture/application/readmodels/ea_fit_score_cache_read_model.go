package readmodels

import (
	"context"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
)

type FitScoreEntry struct {
	ComponentID string
	PillarID    string
	Score       int
	Rationale   string
}

type EAFitScoreCacheReadModel struct {
	db *database.TenantAwareDB
}

func NewEAFitScoreCacheReadModel(db *database.TenantAwareDB) *EAFitScoreCacheReadModel {
	return &EAFitScoreCacheReadModel{db: db}
}

func (rm *EAFitScoreCacheReadModel) Upsert(ctx context.Context, entry FitScoreEntry) error {
	return rm.execForTenant(ctx,
		`INSERT INTO ea_fit_score_cache (tenant_id, component_id, pillar_id, score, rationale)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (tenant_id, component_id, pillar_id) DO UPDATE SET
		 score = EXCLUDED.score,
		 rationale = EXCLUDED.rationale`,
		entry.ComponentID, entry.PillarID, entry.Score, entry.Rationale,
	)
}

func (rm *EAFitScoreCacheReadModel) Delete(ctx context.Context, componentID, pillarID string) error {
	return rm.execForTenant(ctx,
		"DELETE FROM ea_fit_score_cache WHERE tenant_id = $1 AND component_id = $2 AND pillar_id = $3",
		componentID, pillarID,
	)
}

func (rm *EAFitScoreCacheReadModel) execForTenant(ctx context.Context, query string, args ...any) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}
	_, err = rm.db.ExecContext(ctx, query, append([]any{tenantID.Value()}, args...)...)
	return err
}
