package readmodels

import (
	"context"
	"database/sql"

	"easi/backend/internal/capabilitymapping/infrastructure/architecturemodeling"
	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
)

type ComponentCacheReadModel struct {
	db *database.TenantAwareDB
}

func NewComponentCacheReadModel(db *database.TenantAwareDB) *ComponentCacheReadModel {
	return &ComponentCacheReadModel{db: db}
}

func (rm *ComponentCacheReadModel) GetByID(ctx context.Context, id string) (*architecturemodeling.ComponentDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var dto architecturemodeling.ComponentDTO
	var notFound bool

	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		err := tx.QueryRowContext(ctx,
			"SELECT id, name FROM capabilitymapping.capability_component_cache WHERE tenant_id = $1 AND id = $2",
			tenantID.Value(), id,
		).Scan(&dto.ID, &dto.Name)

		if err == sql.ErrNoRows {
			notFound = true
			return nil
		}
		return err
	})

	if err != nil {
		return nil, err
	}
	if notFound {
		return nil, nil
	}

	return &dto, nil
}

func (rm *ComponentCacheReadModel) Upsert(ctx context.Context, id, name string) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx, `
		INSERT INTO capabilitymapping.capability_component_cache (tenant_id, id, name)
		VALUES ($1, $2, $3)
		ON CONFLICT (tenant_id, id) DO UPDATE SET name = EXCLUDED.name
	`, tenantID.Value(), id, name)
	return err
}

func (rm *ComponentCacheReadModel) Delete(ctx context.Context, id string) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"DELETE FROM capabilitymapping.capability_component_cache WHERE tenant_id = $1 AND id = $2",
		tenantID.Value(), id,
	)
	return err
}
