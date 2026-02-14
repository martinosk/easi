package handlers

import (
	"context"

	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/valuestreams/application/commands"
	"easi/backend/internal/valuestreams/domain/aggregates"
	"easi/backend/internal/valuestreams/domain/valueobjects"
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
		return cqrs.EmptyResult(), mapRepositoryError(err)
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
		return cqrs.EmptyResult(), mapStageError(err)
	}

	if err := h.repository.Save(ctx, vs); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.EmptyResult(), nil
}
