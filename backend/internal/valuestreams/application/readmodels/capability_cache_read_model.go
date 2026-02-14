package readmodels

import (
	"context"
	"database/sql"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
)

type CapabilityCacheDTO struct {
	ID   string
	Name string
}

type CapabilityCacheReadModel struct {
	db *database.TenantAwareDB
}

func NewCapabilityCacheReadModel(db *database.TenantAwareDB) *CapabilityCacheReadModel {
	return &CapabilityCacheReadModel{db: db}
}

func (rm *CapabilityCacheReadModel) GetByID(ctx context.Context, id string) (*CapabilityCacheDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var dto CapabilityCacheDTO
	var notFound bool

	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		err := tx.QueryRowContext(ctx,
			"SELECT id, name FROM value_stream_capability_cache WHERE tenant_id = $1 AND id = $2",
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

func (rm *CapabilityCacheReadModel) Upsert(ctx context.Context, id, name string) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx, `
		INSERT INTO value_stream_capability_cache (tenant_id, id, name)
		VALUES ($1, $2, $3)
		ON CONFLICT (tenant_id, id) DO UPDATE SET name = EXCLUDED.name
	`, tenantID.Value(), id, name)
	return err
}

func (rm *CapabilityCacheReadModel) Delete(ctx context.Context, id string) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"DELETE FROM value_stream_capability_cache WHERE tenant_id = $1 AND id = $2",
		tenantID.Value(), id,
	)
	return err
}
