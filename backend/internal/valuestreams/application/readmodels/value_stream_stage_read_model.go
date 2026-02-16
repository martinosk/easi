package readmodels

import (
	"context"

	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/types"
)

type ValueStreamStageDTO struct {
	ID            string      `json:"id"`
	ValueStreamID string      `json:"valueStreamId"`
	Name          string      `json:"name"`
	Description   string      `json:"description,omitempty"`
	Position      int         `json:"position"`
	Links         types.Links `json:"_links,omitempty"`
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

type StageNameQuery struct {
	ValueStreamID string
	Name          string
	ExcludeID     string
}

func (rm *ValueStreamReadModel) InsertStage(ctx context.Context, dto ValueStreamStageDTO) error {
	return rm.idempotentInsert(ctx,
		"DELETE FROM valuestreams.value_stream_stages WHERE tenant_id = $1 AND id = $2",
		"INSERT INTO valuestreams.value_stream_stages (id, tenant_id, value_stream_id, name, description, position) VALUES ($1, $2, $3, $4, $5, $6)",
		func(tid string) []interface{} { return []interface{}{tid, dto.ID} },
		func(tid string) []interface{} { return []interface{}{dto.ID, tid, dto.ValueStreamID, dto.Name, dto.Description, dto.Position} },
	)
}

func (rm *ValueStreamReadModel) UpdateStage(ctx context.Context, update StageUpdate) error {
	return rm.execTenantQuery(ctx,
		"UPDATE valuestreams.value_stream_stages SET name = $1, description = $2, updated_at = CURRENT_TIMESTAMP WHERE tenant_id = $3 AND id = $4",
		func(tid string) []interface{} { return []interface{}{update.Name, update.Description, tid, update.StageID} },
	)
}

func (rm *ValueStreamReadModel) DeleteStage(ctx context.Context, stageID string) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}
	return rm.cascadeDeleteStages(ctx,
		"DELETE FROM valuestreams.value_stream_stage_capabilities WHERE tenant_id = $1 AND stage_id = $2",
		"DELETE FROM valuestreams.value_stream_stages WHERE tenant_id = $1 AND id = $2",
		tenantID.Value(), stageID,
	)
}

func (rm *ValueStreamReadModel) DeleteStagesByValueStreamID(ctx context.Context, valueStreamID string) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}
	return rm.cascadeDeleteStages(ctx,
		"DELETE FROM valuestreams.value_stream_stage_capabilities WHERE tenant_id = $1 AND stage_id IN (SELECT id FROM valuestreams.value_stream_stages WHERE tenant_id = $1 AND value_stream_id = $2)",
		"DELETE FROM valuestreams.value_stream_stages WHERE tenant_id = $1 AND value_stream_id = $2",
		tenantID.Value(), valueStreamID,
	)
}

func (rm *ValueStreamReadModel) UpdateStagePositions(ctx context.Context, updates []StagePositionUpdate) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}
	for _, u := range updates {
		_, err = rm.db.ExecContext(ctx,
			"UPDATE valuestreams.value_stream_stages SET position = $1, updated_at = CURRENT_TIMESTAMP WHERE tenant_id = $2 AND id = $3",
			u.Position, tenantID.Value(), u.StageID,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (rm *ValueStreamReadModel) GetStagesByValueStreamID(ctx context.Context, valueStreamID string) ([]ValueStreamStageDTO, error) {
	return queryList(rm, ctx,
		"SELECT id, value_stream_id, name, COALESCE(description, ''), position FROM valuestreams.value_stream_stages WHERE tenant_id = $1 AND value_stream_id = $2 ORDER BY position",
		valueStreamID,
		func(rows scanner) (ValueStreamStageDTO, error) {
			var dto ValueStreamStageDTO
			err := rows.Scan(&dto.ID, &dto.ValueStreamID, &dto.Name, &dto.Description, &dto.Position)
			return dto, err
		},
	)
}

func (rm *ValueStreamReadModel) AdjustStageCount(ctx context.Context, valueStreamID string, delta int) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}
	_, err = rm.db.ExecContext(ctx,
		"UPDATE valuestreams.value_streams SET stage_count = GREATEST(stage_count + $1, 0) WHERE tenant_id = $2 AND id = $3",
		delta, tenantID.Value(), valueStreamID,
	)
	return err
}

func (rm *ValueStreamReadModel) StageNameExists(ctx context.Context, query StageNameQuery) (bool, error) {
	return rm.nameExistsForTenant(ctx,
		"SELECT COUNT(*) FROM valuestreams.value_stream_stages WHERE tenant_id = $1 AND value_stream_id = $2 AND name = $3",
		"SELECT COUNT(*) FROM valuestreams.value_stream_stages WHERE tenant_id = $1 AND value_stream_id = $2 AND name = $3 AND id != $4",
		query.ExcludeID, query.ValueStreamID, query.Name,
	)
}

func (rm *ValueStreamReadModel) cascadeDeleteStages(ctx context.Context, capQuery, stageQuery string, args ...interface{}) error {
	if _, err := rm.db.ExecContext(ctx, capQuery, args...); err != nil {
		return err
	}
	_, err := rm.db.ExecContext(ctx, stageQuery, args...)
	return err
}
