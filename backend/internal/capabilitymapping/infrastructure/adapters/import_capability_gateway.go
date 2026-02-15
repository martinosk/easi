package adapters

import (
	"context"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/shared/cqrs"
)

type ImportCapabilityGateway struct {
	commandBus cqrs.CommandBus
}

func NewImportCapabilityGateway(bus cqrs.CommandBus) *ImportCapabilityGateway {
	return &ImportCapabilityGateway{commandBus: bus}
}

func (g *ImportCapabilityGateway) CreateCapability(ctx context.Context, name, description, parentID, level string) (string, error) {
	result, err := g.commandBus.Dispatch(ctx, &commands.CreateCapability{
		Name: name, Description: description,
		ParentID: parentID, Level: level,
	})
	if err != nil {
		return "", err
	}
	return result.CreatedID, nil
}

func (g *ImportCapabilityGateway) UpdateMetadata(ctx context.Context, id, eaOwner, status string) error {
	_, err := g.commandBus.Dispatch(ctx, &commands.UpdateCapabilityMetadata{
		ID: id, EAOwner: eaOwner, Status: status,
	})
	return err
}

func (g *ImportCapabilityGateway) LinkSystem(ctx context.Context, capabilityID, componentID, realizationLevel, notes string) (string, error) {
	result, err := g.commandBus.Dispatch(ctx, &commands.LinkSystemToCapability{
		CapabilityID: capabilityID, ComponentID: componentID,
		RealizationLevel: realizationLevel, Notes: notes,
	})
	if err != nil {
		return "", err
	}
	return result.CreatedID, nil
}

func (g *ImportCapabilityGateway) AssignToDomain(ctx context.Context, capabilityID, businessDomainID string) error {
	_, err := g.commandBus.Dispatch(ctx, &commands.AssignCapabilityToDomain{
		CapabilityID: capabilityID, BusinessDomainID: businessDomainID,
	})
	return err
}
