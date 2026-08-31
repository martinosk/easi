package adapters

import (
	"context"

	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/valuestreams/publishedlanguage"
)

type ImportValueStreamGateway struct {
	commandDispatcher
}

func NewImportValueStreamGateway(bus cqrs.CommandBus) *ImportValueStreamGateway {
	return &ImportValueStreamGateway{commandDispatcher{commandBus: bus}}
}

func (g *ImportValueStreamGateway) CreateValueStream(ctx context.Context, cmd publishedlanguage.CreateValueStream) (string, error) {
	return g.createdID(ctx, &cmd, cmd.Name)
}

func (g *ImportValueStreamGateway) AddStage(ctx context.Context, cmd publishedlanguage.AddStage) (string, error) {
	return g.createdID(ctx, &cmd, "value stream "+cmd.ValueStreamID)
}

func (g *ImportValueStreamGateway) MapCapabilityToStage(ctx context.Context, cmd publishedlanguage.AddStageCapability) error {
	return g.dispatch(ctx, &cmd, "value stream "+cmd.ValueStreamID+" stage "+cmd.StageID+" capability "+cmd.CapabilityID)
}
