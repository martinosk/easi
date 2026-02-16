package adapters

import (
	"context"
	"fmt"

	"easi/backend/internal/architecturemodeling/application/commands"
	"easi/backend/internal/shared/cqrs"
)

type ImportComponentGateway struct {
	commandBus cqrs.CommandBus
}

func NewImportComponentGateway(bus cqrs.CommandBus) *ImportComponentGateway {
	return &ImportComponentGateway{commandBus: bus}
}

func (g *ImportComponentGateway) CreateComponent(ctx context.Context, name, description string) (string, error) {
	result, err := g.commandBus.Dispatch(ctx, &commands.CreateApplicationComponent{
		Name: name, Description: description,
	})
	if err != nil {
		return "", fmt.Errorf("dispatch create application component command for %s: %w", name, err)
	}
	return result.CreatedID, nil
}

func (g *ImportComponentGateway) CreateRelation(ctx context.Context, sourceID, targetID, relationType, name, description string) (string, error) {
	result, err := g.commandBus.Dispatch(ctx, &commands.CreateComponentRelation{
		SourceComponentID: sourceID,
		TargetComponentID: targetID,
		RelationType:      relationType,
		Name:              name,
		Description:       description,
	})
	if err != nil {
		return "", fmt.Errorf("dispatch create component relation command for source %s target %s: %w", sourceID, targetID, err)
	}
	return result.CreatedID, nil
}
