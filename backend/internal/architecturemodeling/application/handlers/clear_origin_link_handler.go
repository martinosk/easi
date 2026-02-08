package handlers

import (
	"context"

	"easi/backend/internal/architecturemodeling/application/commands"
	"easi/backend/internal/architecturemodeling/domain/aggregates"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	"easi/backend/internal/architecturemodeling/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type ClearOriginLinkHandler struct {
	repository *repositories.ComponentOriginLinkRepository
}

func NewClearOriginLinkHandler(repository *repositories.ComponentOriginLinkRepository) *ClearOriginLinkHandler {
	return &ClearOriginLinkHandler{repository: repository}
}

func (h *ClearOriginLinkHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.ClearOriginLink)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	componentID, err := valueobjects.NewComponentIDFromString(command.ComponentID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	originType, err := valueobjects.NewOriginType(command.OriginType)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	aggregateID := aggregates.BuildOriginLinkAggregateID(originType.String(), componentID.String())
	link, err := h.repository.GetByID(ctx, aggregateID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := link.Clear(); err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, link); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.NewResult(componentID.String()), nil
}
