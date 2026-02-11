package handlers

import (
	"context"
	"errors"

	"easi/backend/internal/valuestreams/application/commands"
	"easi/backend/internal/valuestreams/domain/aggregates"
	"easi/backend/internal/valuestreams/domain/valueobjects"
	"easi/backend/internal/valuestreams/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type RemoveStageCapabilityRepository interface {
	GetByID(ctx context.Context, id string) (*aggregates.ValueStream, error)
	Save(ctx context.Context, vs *aggregates.ValueStream) error
}

type RemoveStageCapabilityHandler struct {
	repository RemoveStageCapabilityRepository
}

func NewRemoveStageCapabilityHandler(repository RemoveStageCapabilityRepository) *RemoveStageCapabilityHandler {
	return &RemoveStageCapabilityHandler{repository: repository}
}

func (h *RemoveStageCapabilityHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.RemoveStageCapability)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	vs, err := h.repository.GetByID(ctx, command.ValueStreamID)
	if err != nil {
		if errors.Is(err, repositories.ErrValueStreamNotFound) {
			return cqrs.EmptyResult(), ErrValueStreamNotFound
		}
		return cqrs.EmptyResult(), err
	}

	stageID, err := valueobjects.NewStageIDFromString(command.StageID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	capRef, err := valueobjects.NewCapabilityRef(command.CapabilityID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := vs.RemoveCapabilityFromStage(stageID, capRef); err != nil {
		if errors.Is(err, aggregates.ErrStageNotFound) {
			return cqrs.EmptyResult(), ErrStageNotFound
		}
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, vs); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.EmptyResult(), nil
}
