package readmodels

import (
	"context"
	"database/sql"
	"time"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
)

type RealizationDTO struct {
	ID               string            `json:"id"`
	CapabilityID     string            `json:"capabilityId"`
	ComponentID      string            `json:"componentId"`
	RealizationLevel string            `json:"realizationLevel"`
	Notes            string            `json:"notes,omitempty"`
	LinkedAt         time.Time         `json:"linkedAt"`
	Links            map[string]string `json:"_links,omitempty"`
}

type RealizationReadModel struct {
	db *database.TenantAwareDB
}

func NewRealizationReadModel(db *database.TenantAwareDB) *RealizationReadModel {
	return &RealizationReadModel{db: db}
}

func (rm *RealizationReadModel) Insert(ctx context.Context, dto RealizationDTO) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"INSERT INTO capability_realizations (id, tenant_id, capability_id, component_id, realization_level, notes, linked_at) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		dto.ID, tenantID.Value(), dto.CapabilityID, dto.ComponentID, dto.RealizationLevel, dto.Notes, dto.LinkedAt,
	)
	return err
}

func (rm *RealizationReadModel) Update(ctx context.Context, id, realizationLevel, notes string) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"UPDATE capability_realizations SET realization_level = $1, notes = $2, updated_at = CURRENT_TIMESTAMP WHERE tenant_id = $3 AND id = $4",
		realizationLevel, notes, tenantID.Value(), id,
	)
	return err
}

func (rm *RealizationReadModel) Delete(ctx context.Context, id string) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"DELETE FROM capability_realizations WHERE tenant_id = $1 AND id = $2",
		tenantID.Value(), id,
	)
	return err
}

func (rm *RealizationReadModel) GetByCapabilityID(ctx context.Context, capabilityID string) ([]RealizationDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var realizations []RealizationDTO
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx,
			"SELECT id, capability_id, component_id, realization_level, notes, linked_at FROM capability_realizations WHERE tenant_id = $1 AND capability_id = $2 ORDER BY linked_at DESC",
			tenantID.Value(), capabilityID,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var dto RealizationDTO
			if err := rows.Scan(&dto.ID, &dto.CapabilityID, &dto.ComponentID, &dto.RealizationLevel, &dto.Notes, &dto.LinkedAt); err != nil {
				return err
			}
			realizations = append(realizations, dto)
		}

		return rows.Err()
	})

	return realizations, err
}

func (rm *RealizationReadModel) GetByComponentID(ctx context.Context, componentID string) ([]RealizationDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var realizations []RealizationDTO
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx,
			"SELECT id, capability_id, component_id, realization_level, notes, linked_at FROM capability_realizations WHERE tenant_id = $1 AND component_id = $2 ORDER BY linked_at DESC",
			tenantID.Value(), componentID,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var dto RealizationDTO
			if err := rows.Scan(&dto.ID, &dto.CapabilityID, &dto.ComponentID, &dto.RealizationLevel, &dto.Notes, &dto.LinkedAt); err != nil {
				return err
			}
			realizations = append(realizations, dto)
		}

		return rows.Err()
	})

	return realizations, err
}

func (rm *RealizationReadModel) GetByID(ctx context.Context, id string) (*RealizationDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var dto RealizationDTO
	var notFound bool

	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		err := tx.QueryRowContext(ctx,
			"SELECT id, capability_id, component_id, realization_level, notes, linked_at FROM capability_realizations WHERE tenant_id = $1 AND id = $2",
			tenantID.Value(), id,
		).Scan(&dto.ID, &dto.CapabilityID, &dto.ComponentID, &dto.RealizationLevel, &dto.Notes, &dto.LinkedAt)

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
