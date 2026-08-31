package adapters

import (
	"context"

	"easi/backend/internal/capabilitymapping/publishedlanguage"
	"easi/backend/internal/shared/cqrs"
)

type ImportCapabilityGateway struct {
	commandDispatcher
}

func NewImportCapabilityGateway(bus cqrs.CommandBus) *ImportCapabilityGateway {
	return &ImportCapabilityGateway{commandDispatcher{commandBus: bus}}
}

func (g *ImportCapabilityGateway) CreateCapability(ctx context.Context, cmd publishedlanguage.CreateCapability) (string, error) {
	return g.createdID(ctx, &cmd, cmd.Name)
}

func (g *ImportCapabilityGateway) UpdateMetadata(ctx context.Context, cmd publishedlanguage.UpdateCapabilityMetadata) error {
	return g.dispatch(ctx, &cmd, "capability "+cmd.ID)
}

func (g *ImportCapabilityGateway) LinkSystem(ctx context.Context, cmd publishedlanguage.LinkSystemToCapability) (string, error) {
	return g.createdID(ctx, &cmd, "capability "+cmd.CapabilityID+" component "+cmd.ComponentID)
}

func (g *ImportCapabilityGateway) AssignToDomain(ctx context.Context, cmd publishedlanguage.AssignCapabilityToDomain) error {
	return g.dispatch(ctx, &cmd, "capability "+cmd.CapabilityID+" domain "+cmd.BusinessDomainID)
}
