package ports

import (
	"context"

	"easi/backend/internal/importing/publishedlanguage"
)

type ComponentGateway interface {
	CreateComponent(ctx context.Context, name, description string) (string, error)
	CreateRelation(ctx context.Context, input publishedlanguage.CreateRelationInput) (string, error)
}

type CapabilityGateway interface {
	CreateCapability(ctx context.Context, input publishedlanguage.CreateCapabilityInput) (string, error)
	UpdateMetadata(ctx context.Context, id, eaOwner, status string) error
	LinkSystem(ctx context.Context, input publishedlanguage.LinkSystemInput) (string, error)
	AssignToDomain(ctx context.Context, capabilityID, businessDomainID string) error
}

type ValueStreamGateway interface {
	CreateValueStream(ctx context.Context, name, description string) (string, error)
	AddStage(ctx context.Context, valueStreamID, name, description string) (string, error)
	MapCapabilityToStage(ctx context.Context, valueStreamID, stageID, capabilityID string) error
}
