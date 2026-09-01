package handlers

import (
	"context"

	"easi/backend/internal/architecturemodeling/application/commands"
	"easi/backend/internal/architecturemodeling/domain/aggregates"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	"easi/backend/internal/shared/cqrs"
)

type ClassifyApplicationHostingHandler struct {
	repository ComponentRepository
}

func NewClassifyApplicationHostingHandler(repository ComponentRepository) *ClassifyApplicationHostingHandler {
	return &ClassifyApplicationHostingHandler{repository: repository}
}

func (h *ClassifyApplicationHostingHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.ClassifyApplicationHosting)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	classification, err := valueobjects.NewHostingClassification(command.Hosting)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	return mutateComponent(ctx, h.repository, command.ComponentID, func(component *aggregates.ApplicationComponent) error {
		return component.ClassifyHosting(classification)
	})
}
