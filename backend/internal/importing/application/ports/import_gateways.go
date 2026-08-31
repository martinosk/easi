package ports

import (
	"context"

	amPL "easi/backend/internal/architecturemodeling/publishedlanguage"
	cmPL "easi/backend/internal/capabilitymapping/publishedlanguage"
	vsPL "easi/backend/internal/valuestreams/publishedlanguage"
)

type ComponentGateway interface {
	CreateComponent(ctx context.Context, cmd amPL.CreateApplicationComponent) (string, error)
	CreateRelation(ctx context.Context, cmd amPL.CreateComponentRelation) (string, error)
}

type CapabilityGateway interface {
	CreateCapability(ctx context.Context, cmd cmPL.CreateCapability) (string, error)
	UpdateMetadata(ctx context.Context, cmd cmPL.UpdateCapabilityMetadata) error
	LinkSystem(ctx context.Context, cmd cmPL.LinkSystemToCapability) (string, error)
	AssignToDomain(ctx context.Context, cmd cmPL.AssignCapabilityToDomain) error
}

type ValueStreamGateway interface {
	CreateValueStream(ctx context.Context, cmd vsPL.CreateValueStream) (string, error)
	AddStage(ctx context.Context, cmd vsPL.AddStage) (string, error)
	MapCapabilityToStage(ctx context.Context, cmd vsPL.AddStageCapability) error
}
