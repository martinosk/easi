package adapters

import (
	"context"

	"easi/backend/internal/architecturemodeling/publishedlanguage"
	"easi/backend/internal/shared/cqrs"
)

type ImportComponentGateway struct {
	commandDispatcher
}

func NewImportComponentGateway(bus cqrs.CommandBus) *ImportComponentGateway {
	return &ImportComponentGateway{commandDispatcher{commandBus: bus}}
}

func (g *ImportComponentGateway) CreateComponent(ctx context.Context, cmd publishedlanguage.CreateApplicationComponent) (string, error) {
	return g.createdID(ctx, &cmd, cmd.Name)
}

func (g *ImportComponentGateway) CreateRelation(ctx context.Context, cmd publishedlanguage.CreateComponentRelation) (string, error) {
	return g.createdID(ctx, &cmd, "source "+cmd.SourceComponentID+" target "+cmd.TargetComponentID)
}
