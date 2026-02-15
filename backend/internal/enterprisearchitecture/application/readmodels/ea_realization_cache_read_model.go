package readmodels

import (
	"context"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
)

type RealizationEntry struct {
	RealizationID string
	CapabilityID  string
	ComponentID   string
	ComponentName string
	Origin        string
}

type EARealizationCacheReadModel struct {
	db *database.TenantAwareDB
}

func NewEARealizationCacheReadModel(db *database.TenantAwareDB) *EARealizationCacheReadModel {
	return &EARealizationCacheReadModel{db: db}
}

func (rm *EARealizationCacheReadModel) Upsert(ctx context.Context, entry RealizationEntry) error {
	return rm.execForTenant(ctx,
		`INSERT INTO ea_realization_cache (tenant_id, realization_id, capability_id, component_id, component_name, origin)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT (tenant_id, realization_id) DO UPDATE SET
		 capability_id = EXCLUDED.capability_id,
		 component_id = EXCLUDED.component_id,
		 component_name = EXCLUDED.component_name,
		 origin = EXCLUDED.origin`,
		entry.RealizationID, entry.CapabilityID, entry.ComponentID, entry.ComponentName, entry.Origin,
	)
}

func (rm *EARealizationCacheReadModel) Delete(ctx context.Context, realizationID string) error {
	return rm.execForTenant(ctx,
		"DELETE FROM ea_realization_cache WHERE tenant_id = $1 AND realization_id = $2",
		realizationID,
	)
}

func (rm *EARealizationCacheReadModel) DeleteByCapabilityID(ctx context.Context, capabilityID string) error {
	return rm.execForTenant(ctx,
		"DELETE FROM ea_realization_cache WHERE tenant_id = $1 AND capability_id = $2",
		capabilityID,
	)
}

func (rm *EARealizationCacheReadModel) UpdateComponentName(ctx context.Context, componentID, componentName string) error {
	return rm.execForTenant(ctx,
		"UPDATE ea_realization_cache SET component_name = $2 WHERE tenant_id = $1 AND component_id = $3",
		componentName, componentID,
	)
}

func (rm *EARealizationCacheReadModel) execForTenant(ctx context.Context, query string, args ...any) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}
	_, err = rm.db.ExecContext(ctx, query, append([]any{tenantID.Value()}, args...)...)
	return err
}
