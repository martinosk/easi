package adapters

import (
	"context"
	"fmt"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/importing/publishedlanguage"
	"easi/backend/internal/shared/cqrs"
)

type ImportCapabilityGateway struct {
	commandBus cqrs.CommandBus
}

func NewImportCapabilityGateway(bus cqrs.CommandBus) *ImportCapabilityGateway {
	return &ImportCapabilityGateway{commandBus: bus}
}

func (g *ImportCapabilityGateway) CreateCapability(ctx context.Context, input publishedlanguage.CreateCapabilityInput) (string, error) {
	result, err := g.commandBus.Dispatch(ctx, &commands.CreateCapability{
		Name:        input.Name,
		Description: input.Description,
		ParentID:    input.ParentID,
		Level:       input.Level,
	})
	if err != nil {
		return "", fmt.Errorf("dispatch create capability command for %s: %w", input.Name, err)
	}
	return result.CreatedID, nil
}

func (g *ImportCapabilityGateway) UpdateMetadata(ctx context.Context, id, eaOwner, status string) error {
	_, err := g.commandBus.Dispatch(ctx, &commands.UpdateCapabilityMetadata{
		ID:      id,
		EAOwner: eaOwner,
		Status:  status,
	})
	if err != nil {
		return fmt.Errorf("dispatch update capability metadata command for capability %s: %w", id, err)
	}
	return nil
}

func (g *ImportCapabilityGateway) LinkSystem(ctx context.Context, input publishedlanguage.LinkSystemInput) (string, error) {
	result, err := g.commandBus.Dispatch(ctx, &commands.LinkSystemToCapability{
		CapabilityID:     input.CapabilityID,
		ComponentID:      input.ComponentID,
		RealizationLevel: input.RealizationLevel,
		Notes:            input.Notes,
	})
	if err != nil {
		return "", fmt.Errorf("dispatch link system to capability command for capability %s component %s: %w", input.CapabilityID, input.ComponentID, err)
	}
	return result.CreatedID, nil
}

func (g *ImportCapabilityGateway) AssignToDomain(ctx context.Context, capabilityID, businessDomainID string) error {
	_, err := g.commandBus.Dispatch(ctx, &commands.AssignCapabilityToDomain{
		CapabilityID:     capabilityID,
		BusinessDomainID: businessDomainID,
	})
	if err != nil {
		return fmt.Errorf("dispatch assign capability to domain command for capability %s domain %s: %w", capabilityID, businessDomainID, err)
	}
	return nil
}
