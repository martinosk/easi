package readmodels

import (
	"context"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
)

type ImportanceEntry struct {
	CapabilityID        string
	BusinessDomainID    string
	PillarID            string
	EffectiveImportance int
}

type EAImportanceCacheReadModel struct {
	db *database.TenantAwareDB
}

func NewEAImportanceCacheReadModel(db *database.TenantAwareDB) *EAImportanceCacheReadModel {
	return &EAImportanceCacheReadModel{db: db}
}

func (rm *EAImportanceCacheReadModel) Upsert(ctx context.Context, entry ImportanceEntry) error {
	return rm.execForTenant(ctx,
		`INSERT INTO enterprisearchitecture.ea_importance_cache (tenant_id, capability_id, business_domain_id, pillar_id, effective_importance)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (tenant_id, capability_id, business_domain_id, pillar_id) DO UPDATE SET
		 effective_importance = EXCLUDED.effective_importance`,
		entry.CapabilityID, entry.BusinessDomainID, entry.PillarID, entry.EffectiveImportance,
	)
}

func (rm *EAImportanceCacheReadModel) execForTenant(ctx context.Context, query string, args ...any) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}
	_, err = rm.db.ExecContext(ctx, query, append([]any{tenantID.Value()}, args...)...)
	return err
}
