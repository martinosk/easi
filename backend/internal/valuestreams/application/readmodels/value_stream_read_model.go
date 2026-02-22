package readmodels

import (
	"context"
	"database/sql"
	"time"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/types"
)

type ValueStreamDTO struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	StageCount  int         `json:"stageCount"`
	CreatedAt   time.Time   `json:"createdAt"`
	UpdatedAt   *time.Time  `json:"updatedAt,omitempty"`
	Links       types.Links `json:"_links,omitempty"`
}

type ValueStreamUpdate struct {
	Name        string
	Description string
}

type ValueStreamDetailDTO struct {
	ValueStreamDTO
	Stages            []ValueStreamStageDTO       `json:"stages"`
	StageCapabilities []StageCapabilityMappingDTO `json:"stageCapabilities"`
}

type ValueStreamReadModel struct {
	db *database.TenantAwareDB
}

func NewValueStreamReadModel(db *database.TenantAwareDB) *ValueStreamReadModel {
	return &ValueStreamReadModel{db: db}
}

func (rm *ValueStreamReadModel) Insert(ctx context.Context, dto ValueStreamDTO) error {
	return rm.idempotentInsert(ctx,
		"DELETE FROM valuestreams.value_streams WHERE tenant_id = $1 AND id = $2",
		"INSERT INTO valuestreams.value_streams (id, tenant_id, name, description, stage_count, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
		func(tid string) []interface{} { return []interface{}{tid, dto.ID} },
		func(tid string) []interface{} {
			return []interface{}{dto.ID, tid, dto.Name, dto.Description, 0, dto.CreatedAt}
		},
	)
}

func (rm *ValueStreamReadModel) Update(ctx context.Context, id string, update ValueStreamUpdate) error {
	return rm.execTenantQuery(ctx,
		"UPDATE valuestreams.value_streams SET name = $1, description = $2, updated_at = CURRENT_TIMESTAMP WHERE tenant_id = $3 AND id = $4",
		func(tid string) []interface{} { return []interface{}{update.Name, update.Description, tid, id} },
	)
}

func (rm *ValueStreamReadModel) Delete(ctx context.Context, id string) error {
	return rm.execTenantQuery(ctx,
		"DELETE FROM valuestreams.value_streams WHERE tenant_id = $1 AND id = $2",
		func(tid string) []interface{} { return []interface{}{tid, id} },
	)
}

func (rm *ValueStreamReadModel) GetAll(ctx context.Context) ([]ValueStreamDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var streams []ValueStreamDTO
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx,
			"SELECT id, name, description, stage_count, created_at, updated_at FROM valuestreams.value_streams WHERE tenant_id = $1 ORDER BY name",
			tenantID.Value(),
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			dto, err := scanValueStream(rows)
			if err != nil {
				return err
			}
			streams = append(streams, *dto)
		}

		return rows.Err()
	})

	return streams, err
}

func (rm *ValueStreamReadModel) GetByID(ctx context.Context, id string) (*ValueStreamDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var dto *ValueStreamDTO
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		row := tx.QueryRowContext(ctx,
			"SELECT id, name, description, stage_count, created_at, updated_at FROM valuestreams.value_streams WHERE tenant_id = $1 AND id = $2",
			tenantID.Value(), id,
		)

		result, err := scanValueStream(row)
		if err == sql.ErrNoRows {
			return nil
		}
		if err != nil {
			return err
		}
		dto = result
		return nil
	})

	return dto, err
}

func (rm *ValueStreamReadModel) NameExists(ctx context.Context, name, excludeID string) (bool, error) {
	return rm.nameExistsForTenant(ctx,
		"SELECT COUNT(*) FROM valuestreams.value_streams WHERE tenant_id = $1 AND name = $2",
		"SELECT COUNT(*) FROM valuestreams.value_streams WHERE tenant_id = $1 AND name = $2 AND id != $3",
		excludeID, name,
	)
}

func (rm *ValueStreamReadModel) GetValueStreamDetail(ctx context.Context, id string) (*ValueStreamDetailDTO, error) {
	vs, err := rm.GetByID(ctx, id)
	if err != nil || vs == nil {
		return nil, err
	}

	stages, err := rm.GetStagesByValueStreamID(ctx, id)
	if err != nil {
		return nil, err
	}
	if stages == nil {
		stages = []ValueStreamStageDTO{}
	}

	caps, err := rm.GetCapabilitiesByValueStreamID(ctx, id)
	if err != nil {
		return nil, err
	}
	if caps == nil {
		caps = []StageCapabilityMappingDTO{}
	}

	return &ValueStreamDetailDTO{
		ValueStreamDTO:    *vs,
		Stages:            stages,
		StageCapabilities: caps,
	}, nil
}

type scanner interface {
	Scan(dest ...interface{}) error
}

func scanValueStream(s scanner) (*ValueStreamDTO, error) {
	var dto ValueStreamDTO
	var updatedAt sql.NullTime

	if err := s.Scan(&dto.ID, &dto.Name, &dto.Description, &dto.StageCount, &dto.CreatedAt, &updatedAt); err != nil {
		return nil, err
	}

	if updatedAt.Valid {
		dto.UpdatedAt = &updatedAt.Time
	}

	return &dto, nil
}
