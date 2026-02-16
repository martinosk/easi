package adapters

import (
	"context"
	"fmt"

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
		return "", fmt.Errorf("dispatch create capability command for %s: %w", name, err)
	}
	return result.CreatedID, nil
}

func (g *ImportCapabilityGateway) UpdateMetadata(ctx context.Context, id, eaOwner, status string) error {
	_, err := g.commandBus.Dispatch(ctx, &commands.UpdateCapabilityMetadata{
		ID: id, EAOwner: eaOwner, Status: status,
	})
	if err != nil {
		return fmt.Errorf("dispatch update capability metadata command for capability %s: %w", id, err)
	}
	return nil
}

func (g *ImportCapabilityGateway) LinkSystem(ctx context.Context, capabilityID, componentID, realizationLevel, notes string) (string, error) {
	result, err := g.commandBus.Dispatch(ctx, &commands.LinkSystemToCapability{
		CapabilityID: capabilityID, ComponentID: componentID,
		RealizationLevel: realizationLevel, Notes: notes,
	})
	if err != nil {
		return "", fmt.Errorf("dispatch link system to capability command for capability %s component %s: %w", capabilityID, componentID, err)
	}
	return result.CreatedID, nil
}

func (g *ImportCapabilityGateway) AssignToDomain(ctx context.Context, capabilityID, businessDomainID string) error {
	_, err := g.commandBus.Dispatch(ctx, &commands.AssignCapabilityToDomain{
		CapabilityID: capabilityID, BusinessDomainID: businessDomainID,
	})
	if err != nil {
		return fmt.Errorf("dispatch assign capability to domain command for capability %s domain %s: %w", capabilityID, businessDomainID, err)
	}
	return nil
}
