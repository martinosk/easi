package readmodels

import (
	"context"
	"database/sql"
	"errors"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
)

type RealizationID string

type DirectRealizationDTO struct {
	RealizationID RealizationID
	CapabilityID  CapabilityID
	ComponentID   ComponentID
}

type RealizationCacheReadModel struct {
	db *database.TenantAwareDB
}

func NewRealizationCacheReadModel(db *database.TenantAwareDB) *RealizationCacheReadModel {
	return &RealizationCacheReadModel{db: db}
}

func (rm *RealizationCacheReadModel) SaveDirectRealization(ctx context.Context, dto DirectRealizationDTO) error {
	return rm.execForTenant(ctx,
		`INSERT INTO architecturedirection.realization_cache (tenant_id, realization_id, capability_id, component_id)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (tenant_id, realization_id) DO UPDATE SET
		 capability_id = EXCLUDED.capability_id,
		 component_id = EXCLUDED.component_id`,
		string(dto.RealizationID), string(dto.CapabilityID), string(dto.ComponentID),
	)
}

func (rm *RealizationCacheReadModel) RemoveRealization(ctx context.Context, realizationID RealizationID) error {
	return rm.execForTenant(ctx,
		`DELETE FROM architecturedirection.realization_cache WHERE tenant_id = $1 AND realization_id = $2`,
		string(realizationID),
	)
}

func (rm *RealizationCacheReadModel) RemoveRealizationsOfCapability(ctx context.Context, capabilityID CapabilityID) error {
	return rm.execForTenant(ctx,
		`DELETE FROM architecturedirection.realization_cache WHERE tenant_id = $1 AND capability_id = $2`,
		string(capabilityID),
	)
}

func (rm *RealizationCacheReadModel) RemoveRealizationsOfComponent(ctx context.Context, componentID ComponentID) error {
	return rm.execForTenant(ctx,
		`DELETE FROM architecturedirection.realization_cache WHERE tenant_id = $1 AND component_id = $2`,
		string(componentID),
	)
}

func (rm *RealizationCacheReadModel) DirectRealizationID(ctx context.Context, capabilityID CapabilityID, componentID ComponentID) (RealizationID, bool, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return "", false, err
	}
	var realizationID RealizationID
	var notFound bool
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		err := tx.QueryRowContext(ctx,
			`SELECT realization_id FROM architecturedirection.realization_cache
			 WHERE tenant_id = $1 AND capability_id = $2 AND component_id = $3`,
			tenantID.Value(), string(capabilityID), string(componentID),
		).Scan(&realizationID)
		if errors.Is(err, sql.ErrNoRows) {
			notFound = true
			return nil
		}
		return err
	})
	if err != nil || notFound {
		return "", false, err
	}
	return realizationID, true, nil
}

func (rm *RealizationCacheReadModel) execForTenant(ctx context.Context, query string, args ...any) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}
	_, err = rm.db.ExecContext(ctx, query, append([]any{tenantID.Value()}, args...)...)
	return err
}
