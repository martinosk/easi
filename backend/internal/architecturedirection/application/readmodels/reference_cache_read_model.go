package readmodels

import (
	"context"
	"database/sql"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
)

type ReferenceCacheReadModel struct {
	db *database.TenantAwareDB
}

func NewReferenceCacheReadModel(db *database.TenantAwareDB) *ReferenceCacheReadModel {
	return &ReferenceCacheReadModel{db: db}
}

func (rm *ReferenceCacheReadModel) SaveReference(ctx context.Context, entity ReferenceEntity, entityID, name string) error {
	return rm.execForTenant(ctx,
		`INSERT INTO architecturedirection.reference_name_cache (tenant_id, entity_type, entity_id, name)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (tenant_id, entity_type, entity_id) DO UPDATE SET name = EXCLUDED.name`,
		string(entity), entityID, name,
	)
}

func (rm *ReferenceCacheReadModel) RemoveReference(ctx context.Context, entity ReferenceEntity, entityID string) error {
	return rm.execForTenant(ctx,
		`DELETE FROM architecturedirection.reference_name_cache
		 WHERE tenant_id = $1 AND entity_type = $2 AND entity_id = $3`,
		string(entity), entityID,
	)
}

func (rm *ReferenceCacheReadModel) Exists(ctx context.Context, entity ReferenceEntity, entityID string) (bool, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return false, err
	}
	exists := false
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx,
			`SELECT EXISTS (SELECT 1 FROM architecturedirection.reference_name_cache
			 WHERE tenant_id = $1 AND entity_type = $2 AND entity_id = $3)`,
			tenantID.Value(), string(entity), entityID,
		).Scan(&exists)
	})
	return exists, err
}

func (rm *ReferenceCacheReadModel) execForTenant(ctx context.Context, query string, args ...any) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}
	_, err = rm.db.ExecContext(ctx, query, append([]any{tenantID.Value()}, args...)...)
	return err
}
