package readmodels

import (
	"context"
	"database/sql"
	"time"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
	"github.com/lib/pq"
)

type RealizationDTO struct {
	ID                   string            `json:"id"`
	CapabilityID         string            `json:"capabilityId"`
	ComponentID          string            `json:"componentId"`
	ComponentName        string            `json:"componentName,omitempty"`
	RealizationLevel     string            `json:"realizationLevel"`
	Notes                string            `json:"notes,omitempty"`
	Origin               string            `json:"origin"`
	SourceRealizationID  string            `json:"sourceRealizationId,omitempty"`
	SourceCapabilityID   string            `json:"sourceCapabilityId,omitempty"`
	SourceCapabilityName string            `json:"sourceCapabilityName,omitempty"`
	LinkedAt             time.Time         `json:"linkedAt"`
	Links                map[string]string `json:"_links,omitempty"`
}

type RealizationReadModel struct {
	db *database.TenantAwareDB
}

func NewRealizationReadModel(db *database.TenantAwareDB) *RealizationReadModel {
	return &RealizationReadModel{db: db}
}

func (rm *RealizationReadModel) Insert(ctx context.Context, dto RealizationDTO) error {
	return rm.insertRealization(ctx, dto, "INSERT INTO capability_realizations (id, tenant_id, capability_id, component_id, realization_level, notes, origin, source_realization_id, linked_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)", true)
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

	_, err = rm.db.ExecContext(ctx, "DELETE FROM capability_realizations WHERE tenant_id = $1 AND id = $2", tenantID.Value(), id)
	return err
}

func (rm *RealizationReadModel) GetByCapabilityID(ctx context.Context, capabilityID string) ([]RealizationDTO, error) {
	return rm.queryRealizations(ctx, "SELECT id, capability_id, component_id, realization_level, notes, origin, COALESCE(source_realization_id, ''), linked_at FROM capability_realizations WHERE tenant_id = $1 AND capability_id = $2 ORDER BY linked_at DESC", capabilityID)
}

func (rm *RealizationReadModel) GetByComponentID(ctx context.Context, componentID string) ([]RealizationDTO, error) {
	return rm.queryRealizations(ctx, "SELECT id, capability_id, component_id, realization_level, notes, origin, COALESCE(source_realization_id, ''), linked_at FROM capability_realizations WHERE tenant_id = $1 AND component_id = $2 ORDER BY linked_at DESC", componentID)
}

func (rm *RealizationReadModel) queryRealizations(ctx context.Context, query, param string) ([]RealizationDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var realizations []RealizationDTO
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, query, tenantID.Value(), param)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			dto, err := rm.scanRealizationRow(rows)
			if err != nil {
				return err
			}
			realizations = append(realizations, dto)
		}

		return rows.Err()
	})

	return realizations, err
}

func (rm *RealizationReadModel) scanRealizationRow(rows *sql.Rows) (RealizationDTO, error) {
	var dto RealizationDTO
	err := rows.Scan(&dto.ID, &dto.CapabilityID, &dto.ComponentID, &dto.RealizationLevel, &dto.Notes, &dto.Origin, &dto.SourceRealizationID, &dto.LinkedAt)
	return dto, err
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
			"SELECT id, capability_id, component_id, realization_level, notes, origin, COALESCE(source_realization_id, ''), linked_at FROM capability_realizations WHERE tenant_id = $1 AND id = $2",
			tenantID.Value(), id,
		).Scan(&dto.ID, &dto.CapabilityID, &dto.ComponentID, &dto.RealizationLevel, &dto.Notes, &dto.Origin, &dto.SourceRealizationID, &dto.LinkedAt)

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

func (rm *RealizationReadModel) DeleteBySourceRealizationID(ctx context.Context, sourceRealizationID string) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx, "DELETE FROM capability_realizations WHERE tenant_id = $1 AND source_realization_id = $2", tenantID.Value(), sourceRealizationID)
	return err
}

func (rm *RealizationReadModel) InsertInherited(ctx context.Context, dto RealizationDTO) error {
	return rm.insertRealization(ctx, dto, `INSERT INTO capability_realizations (id, tenant_id, capability_id, component_id, realization_level, notes, origin, source_realization_id, linked_at)
		 VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8)
		 ON CONFLICT (tenant_id, capability_id, component_id) DO NOTHING`, false)
}

func (rm *RealizationReadModel) insertRealization(ctx context.Context, dto RealizationDTO, query string, includeID bool) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	args := rm.buildInsertArgs(tenantID.Value(), dto, includeID)
	_, err = rm.db.ExecContext(ctx, query, args...)
	return err
}

func (rm *RealizationReadModel) buildInsertArgs(tenantID string, dto RealizationDTO, includeID bool) []interface{} {
	commonArgs := []interface{}{
		tenantID, dto.CapabilityID, dto.ComponentID, dto.RealizationLevel,
		dto.Notes, dto.Origin, rm.toNullableString(dto.SourceRealizationID), dto.LinkedAt,
	}

	if includeID {
		return append([]interface{}{dto.ID}, commonArgs...)
	}
	return commonArgs
}

func (rm *RealizationReadModel) toNullableString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func (rm *RealizationReadModel) GetByCapabilityIDs(ctx context.Context, capabilityIDs []string) ([]RealizationDTO, error) {
	if len(capabilityIDs) == 0 {
		return []RealizationDTO{}, nil
	}

	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var realizations []RealizationDTO
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		query := `
			SELECT
				cr.id, cr.capability_id, cr.component_id, cr.realization_level, cr.notes,
				cr.origin, COALESCE(cr.source_realization_id, ''), cr.linked_at,
				COALESCE(ac.name, ''),
				COALESCE(source_r.capability_id, ''),
				COALESCE(source_cap.name, '')
			FROM capability_realizations cr
			LEFT JOIN application_components ac ON cr.component_id = ac.id AND ac.tenant_id = cr.tenant_id AND ac.is_deleted = FALSE
			LEFT JOIN capability_realizations source_r ON cr.source_realization_id = source_r.id AND source_r.tenant_id = cr.tenant_id
			LEFT JOIN capabilities source_cap ON source_r.capability_id = source_cap.id AND source_cap.tenant_id = cr.tenant_id
			WHERE cr.tenant_id = $1 AND cr.capability_id = ANY($2)
			ORDER BY cr.linked_at DESC
		`

		rows, err := tx.QueryContext(ctx, query, tenantID.Value(), pq.Array(capabilityIDs))
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var dto RealizationDTO
			err := rows.Scan(
				&dto.ID, &dto.CapabilityID, &dto.ComponentID, &dto.RealizationLevel, &dto.Notes,
				&dto.Origin, &dto.SourceRealizationID, &dto.LinkedAt,
				&dto.ComponentName, &dto.SourceCapabilityID, &dto.SourceCapabilityName,
			)
			if err != nil {
				return err
			}
			realizations = append(realizations, dto)
		}

		return rows.Err()
	})

	return realizations, err
}
