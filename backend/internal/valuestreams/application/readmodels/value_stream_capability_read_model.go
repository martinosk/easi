package readmodels

import (
	"context"

	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/types"
)

type StageCapabilityMappingDTO struct {
	StageID        string      `json:"stageId"`
	CapabilityID   string      `json:"capabilityId"`
	CapabilityName string      `json:"capabilityName,omitempty"`
	Links          types.Links `json:"_links,omitempty"`
}

type StageCapabilityRef struct {
	StageID        string
	CapabilityID   string
	CapabilityName string
}

type StageCapabilityMapping struct {
	ValueStreamID string
	StageID       string
}

func (rm *ValueStreamReadModel) InsertStageCapability(ctx context.Context, ref StageCapabilityRef) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}
	tid := tenantID.Value()

	_, err = rm.db.ExecContext(ctx,
		"DELETE FROM valuestreams.value_stream_stage_capabilities WHERE tenant_id = $1 AND stage_id = $2 AND capability_id = $3",
		tid, ref.StageID, ref.CapabilityID,
	)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"INSERT INTO valuestreams.value_stream_stage_capabilities (tenant_id, stage_id, capability_id, capability_name) VALUES ($1, $2, $3, $4)",
		tid, ref.StageID, ref.CapabilityID, ref.CapabilityName,
	)
	return err
}

func (rm *ValueStreamReadModel) DeleteStageCapability(ctx context.Context, ref StageCapabilityRef) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"DELETE FROM valuestreams.value_stream_stage_capabilities WHERE tenant_id = $1 AND stage_id = $2 AND capability_id = $3",
		tenantID.Value(), ref.StageID, ref.CapabilityID,
	)
	return err
}

func (rm *ValueStreamReadModel) GetCapabilitiesByValueStreamID(ctx context.Context, valueStreamID string) ([]StageCapabilityMappingDTO, error) {
	return queryList(rm, ctx,
		`SELECT sc.stage_id, sc.capability_id, COALESCE(sc.capability_name, '')
		 FROM valuestreams.value_stream_stage_capabilities sc
		 INNER JOIN valuestreams.value_stream_stages s ON sc.tenant_id = s.tenant_id AND sc.stage_id = s.id
		 WHERE sc.tenant_id = $1 AND s.value_stream_id = $2
		 ORDER BY s.position, sc.capability_id`,
		valueStreamID,
		func(rows scanner) (StageCapabilityMappingDTO, error) {
			var dto StageCapabilityMappingDTO
			err := rows.Scan(&dto.StageID, &dto.CapabilityID, &dto.CapabilityName)
			return dto, err
		},
	)
}

func (rm *ValueStreamReadModel) GetStagesByCapabilityID(ctx context.Context, capabilityID string) ([]StageCapabilityMapping, error) {
	return queryList(rm, ctx,
		`SELECT s.value_stream_id, sc.stage_id
		 FROM valuestreams.value_stream_stage_capabilities sc
		 INNER JOIN valuestreams.value_stream_stages s ON sc.tenant_id = s.tenant_id AND sc.stage_id = s.id
		 WHERE sc.tenant_id = $1 AND sc.capability_id = $2`,
		capabilityID,
		func(rows scanner) (StageCapabilityMapping, error) {
			var m StageCapabilityMapping
			err := rows.Scan(&m.ValueStreamID, &m.StageID)
			return m, err
		},
	)
}

func (rm *ValueStreamReadModel) UpdateStageCapabilityName(ctx context.Context, capabilityID, name string) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"UPDATE valuestreams.value_stream_stage_capabilities SET capability_name = $1 WHERE tenant_id = $2 AND capability_id = $3",
		name, tenantID.Value(), capabilityID,
	)
	return err
}
