package adapters

import (
	"context"

	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/valuestreams/application/commands"
)

type ImportValueStreamGateway struct {
	commandBus cqrs.CommandBus
}

func NewImportValueStreamGateway(bus cqrs.CommandBus) *ImportValueStreamGateway {
	return &ImportValueStreamGateway{commandBus: bus}
}

func (g *ImportValueStreamGateway) CreateValueStream(ctx context.Context, name, description string) (string, error) {
	result, err := g.commandBus.Dispatch(ctx, &commands.CreateValueStream{
		Name: name, Description: description,
	})
	if err != nil {
		return "", err
	}
	return result.CreatedID, nil
}

func (g *ImportValueStreamGateway) AddStage(ctx context.Context, valueStreamID, name, description string) (string, error) {
	result, err := g.commandBus.Dispatch(ctx, &commands.AddStage{
		ValueStreamID: valueStreamID, Name: name, Description: description,
	})
	if err != nil {
		return "", err
	}
	return result.CreatedID, nil
}

func (g *ImportValueStreamGateway) MapCapabilityToStage(ctx context.Context, valueStreamID, stageID, capabilityID string) error {
	_, err := g.commandBus.Dispatch(ctx, &commands.AddStageCapability{
		ValueStreamID: valueStreamID, StageID: stageID, CapabilityID: capabilityID,
	})
	return err
}
