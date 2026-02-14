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

type ValueStreamReadModel struct {
	db *database.TenantAwareDB
}

func NewValueStreamReadModel(db *database.TenantAwareDB) *ValueStreamReadModel {
	return &ValueStreamReadModel{db: db}
}

func (rm *ValueStreamReadModel) Insert(ctx context.Context, dto ValueStreamDTO) error {
	return rm.execTenantQuery(ctx,
		"INSERT INTO value_streams (id, tenant_id, name, description, stage_count, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
		func(tid string) []interface{} { return []interface{}{dto.ID, tid, dto.Name, dto.Description, 0, dto.CreatedAt} },
	)
}

func (rm *ValueStreamReadModel) Update(ctx context.Context, id string, update ValueStreamUpdate) error {
	return rm.execTenantQuery(ctx,
		"UPDATE value_streams SET name = $1, description = $2, updated_at = CURRENT_TIMESTAMP WHERE tenant_id = $3 AND id = $4",
		func(tid string) []interface{} { return []interface{}{update.Name, update.Description, tid, id} },
	)
}

func (rm *ValueStreamReadModel) Delete(ctx context.Context, id string) error {
	return rm.execTenantQuery(ctx,
		"DELETE FROM value_streams WHERE tenant_id = $1 AND id = $2",
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
			"SELECT id, name, description, stage_count, created_at, updated_at FROM value_streams WHERE tenant_id = $1 ORDER BY name",
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

func (rm *ValueStreamReadModel) GetByID(ctx context.Context, id string) (*ValueStreamDTO, error) {
	return rm.getByCondition(ctx, "id = $2", id)
}

func (rm *ValueStreamReadModel) getByCondition(ctx context.Context, whereClause string, arg interface{}) (*ValueStreamDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var dto *ValueStreamDTO
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		row := tx.QueryRowContext(ctx,
			"SELECT id, name, description, stage_count, created_at, updated_at FROM value_streams WHERE tenant_id = $1 AND "+whereClause,
			tenantID.Value(), arg,
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
		"SELECT COUNT(*) FROM value_streams WHERE tenant_id = $1 AND name = $2",
		"SELECT COUNT(*) FROM value_streams WHERE tenant_id = $1 AND name = $2 AND id != $3",
		excludeID, name,
	)
}

type ValueStreamStageDTO struct {
	ID            string      `json:"id"`
	ValueStreamID string      `json:"valueStreamId"`
	Name          string      `json:"name"`
	Description   string      `json:"description,omitempty"`
	Position      int         `json:"position"`
	Links         types.Links `json:"_links,omitempty"`
}

type StageCapabilityMappingDTO struct {
	StageID        string      `json:"stageId"`
	CapabilityID   string      `json:"capabilityId"`
	CapabilityName string      `json:"capabilityName,omitempty"`
	Links          types.Links `json:"_links,omitempty"`
}

type ValueStreamDetailDTO struct {
	ValueStreamDTO
	Stages            []ValueStreamStageDTO       `json:"stages"`
	StageCapabilities []StageCapabilityMappingDTO  `json:"stageCapabilities"`
}

type StagePositionUpdate struct {
	StageID  string
	Position int
}

type StageUpdate struct {
	StageID     string
	Name        string
	Description string
}

type StageCapabilityRef struct {
	TenantID     string
	StageID      string
	CapabilityID string
}

type StageNameQuery struct {
	ValueStreamID string
	Name          string
	ExcludeID     string
}

func (rm *ValueStreamReadModel) InsertStage(ctx context.Context, dto ValueStreamStageDTO) error {
	return rm.execTenantQuery(ctx,
		"INSERT INTO value_stream_stages (id, tenant_id, value_stream_id, name, description, position) VALUES ($1, $2, $3, $4, $5, $6)",
		func(tid string) []interface{} { return []interface{}{dto.ID, tid, dto.ValueStreamID, dto.Name, dto.Description, dto.Position} },
	)
}

func (rm *ValueStreamReadModel) UpdateStage(ctx context.Context, update StageUpdate) error {
	return rm.execTenantQuery(ctx,
		"UPDATE value_stream_stages SET name = $1, description = $2, updated_at = CURRENT_TIMESTAMP WHERE tenant_id = $3 AND id = $4",
		func(tid string) []interface{} { return []interface{}{update.Name, update.Description, tid, update.StageID} },
	)
}

func (rm *ValueStreamReadModel) DeleteStage(ctx context.Context, stageID string) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}
	return rm.cascadeDeleteStages(ctx,
		"DELETE FROM value_stream_stage_capabilities WHERE tenant_id = $1 AND stage_id = $2",
		"DELETE FROM value_stream_stages WHERE tenant_id = $1 AND id = $2",
		tenantID.Value(), stageID,
	)
}

func (rm *ValueStreamReadModel) DeleteStagesByValueStreamID(ctx context.Context, valueStreamID string) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}
	return rm.cascadeDeleteStages(ctx,
		"DELETE FROM value_stream_stage_capabilities WHERE tenant_id = $1 AND stage_id IN (SELECT id FROM value_stream_stages WHERE tenant_id = $1 AND value_stream_id = $2)",
		"DELETE FROM value_stream_stages WHERE tenant_id = $1 AND value_stream_id = $2",
		tenantID.Value(), valueStreamID,
	)
}

func (rm *ValueStreamReadModel) cascadeDeleteStages(ctx context.Context, capQuery, stageQuery string, args ...interface{}) error {
	if _, err := rm.db.ExecContext(ctx, capQuery, args...); err != nil {
		return err
	}
	_, err := rm.db.ExecContext(ctx, stageQuery, args...)
	return err
}

func (rm *ValueStreamReadModel) UpdateStagePositions(ctx context.Context, updates []StagePositionUpdate) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}
	for _, u := range updates {
		_, err = rm.db.ExecContext(ctx,
			"UPDATE value_stream_stages SET position = $1, updated_at = CURRENT_TIMESTAMP WHERE tenant_id = $2 AND id = $3",
			u.Position, tenantID.Value(), u.StageID,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (rm *ValueStreamReadModel) GetStagesByValueStreamID(ctx context.Context, valueStreamID string) ([]ValueStreamStageDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var stages []ValueStreamStageDTO
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx,
			"SELECT id, value_stream_id, name, COALESCE(description, ''), position FROM value_stream_stages WHERE tenant_id = $1 AND value_stream_id = $2 ORDER BY position",
			tenantID.Value(), valueStreamID,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var dto ValueStreamStageDTO
			if err := rows.Scan(&dto.ID, &dto.ValueStreamID, &dto.Name, &dto.Description, &dto.Position); err != nil {
				return err
			}
			stages = append(stages, dto)
		}
		return rows.Err()
	})
	return stages, err
}

func (rm *ValueStreamReadModel) AdjustStageCount(ctx context.Context, valueStreamID string, delta int) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}
	_, err = rm.db.ExecContext(ctx,
		"UPDATE value_streams SET stage_count = GREATEST(stage_count + $1, 0) WHERE tenant_id = $2 AND id = $3",
		delta, tenantID.Value(), valueStreamID,
	)
	return err
}

func (rm *ValueStreamReadModel) InsertStageCapability(ctx context.Context, ref StageCapabilityRef) error {
	_, err := rm.db.ExecContext(ctx,
		"INSERT INTO value_stream_stage_capabilities (tenant_id, stage_id, capability_id) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING",
		ref.TenantID, ref.StageID, ref.CapabilityID,
	)
	return err
}

func (rm *ValueStreamReadModel) DeleteStageCapability(ctx context.Context, ref StageCapabilityRef) error {
	_, err := rm.db.ExecContext(ctx,
		"DELETE FROM value_stream_stage_capabilities WHERE tenant_id = $1 AND stage_id = $2 AND capability_id = $3",
		ref.TenantID, ref.StageID, ref.CapabilityID,
	)
	return err
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

func (rm *ValueStreamReadModel) GetCapabilitiesByValueStreamID(ctx context.Context, valueStreamID string) ([]StageCapabilityMappingDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var caps []StageCapabilityMappingDTO
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx,
			`SELECT sc.stage_id, sc.capability_id
			 FROM value_stream_stage_capabilities sc
			 INNER JOIN value_stream_stages s ON sc.tenant_id = s.tenant_id AND sc.stage_id = s.id
			 WHERE sc.tenant_id = $1 AND s.value_stream_id = $2
			 ORDER BY s.position, sc.capability_id`,
			tenantID.Value(), valueStreamID,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var dto StageCapabilityMappingDTO
			if err := rows.Scan(&dto.StageID, &dto.CapabilityID); err != nil {
				return err
			}
			caps = append(caps, dto)
		}
		return rows.Err()
	})
	return caps, err
}

func (rm *ValueStreamReadModel) StageNameExists(ctx context.Context, query StageNameQuery) (bool, error) {
	return rm.nameExistsForTenant(ctx,
		"SELECT COUNT(*) FROM value_stream_stages WHERE tenant_id = $1 AND value_stream_id = $2 AND name = $3",
		"SELECT COUNT(*) FROM value_stream_stages WHERE tenant_id = $1 AND value_stream_id = $2 AND name = $3 AND id != $4",
		query.ExcludeID, query.ValueStreamID, query.Name,
	)
}

func (rm *ValueStreamReadModel) execTenantQuery(ctx context.Context, query string, buildArgs func(tenantID string) []interface{}) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}
	_, err = rm.db.ExecContext(ctx, query, buildArgs(tenantID.Value())...)
	return err
}

func (rm *ValueStreamReadModel) nameExistsForTenant(ctx context.Context, baseQuery, excludeQuery, excludeID string, extraArgs ...interface{}) (bool, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return false, err
	}
	args := append([]interface{}{tenantID.Value()}, extraArgs...)
	var count int
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		if excludeID != "" {
			return tx.QueryRowContext(ctx, excludeQuery, append(args, excludeID)...).Scan(&count)
		}
		return tx.QueryRowContext(ctx, baseQuery, args...).Scan(&count)
	})
	return count > 0, err
}
