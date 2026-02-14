package handlers

import (
	"context"
	"log"

	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/valuestreams/application/commands"
	"easi/backend/internal/valuestreams/application/readmodels"
	"easi/backend/internal/valuestreams/domain/aggregates"
	"easi/backend/internal/valuestreams/domain/valueobjects"
)

type RemoveDeletedCapabilityReadModel interface {
	GetStagesByCapabilityID(ctx context.Context, capabilityID string) ([]readmodels.StageCapabilityMapping, error)
}

type RemoveDeletedCapabilityRepository interface {
	GetByID(ctx context.Context, id string) (*aggregates.ValueStream, error)
	Save(ctx context.Context, vs *aggregates.ValueStream) error
}

type RemoveDeletedCapabilityHandler struct {
	repository RemoveDeletedCapabilityRepository
	readModel  RemoveDeletedCapabilityReadModel
}

func NewRemoveDeletedCapabilityHandler(
	repository RemoveDeletedCapabilityRepository,
	readModel RemoveDeletedCapabilityReadModel,
) *RemoveDeletedCapabilityHandler {
	return &RemoveDeletedCapabilityHandler{
		repository: repository,
		readModel:  readModel,
	}
}

func (h *RemoveDeletedCapabilityHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.RemoveDeletedCapability)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	mappings, err := h.readModel.GetStagesByCapabilityID(ctx, command.CapabilityID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	capRef, err := valueobjects.NewCapabilityRef(command.CapabilityID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	for _, mapping := range mappings {
		if err := h.removeFromValueStream(ctx, mapping.ValueStreamID, mapping.StageID, capRef); err != nil {
			log.Printf("Failed to remove deleted capability %s from value stream %s stage %s: %v",
				command.CapabilityID, mapping.ValueStreamID, mapping.StageID, err)
			return cqrs.EmptyResult(), err
		}
	}

	return cqrs.EmptyResult(), nil
}

func (h *RemoveDeletedCapabilityHandler) removeFromValueStream(
	ctx context.Context,
	valueStreamID, stageID string,
	capRef valueobjects.CapabilityRef,
) error {
	vs, err := h.repository.GetByID(ctx, valueStreamID)
	if err != nil {
		return mapRepositoryError(err)
	}

	sid, err := valueobjects.NewStageIDFromString(stageID)
	if err != nil {
		return err
	}

	if err := vs.RemoveCapabilityFromStage(sid, capRef); err != nil {
		return mapStageError(err)
	}

	return h.repository.Save(ctx, vs)
}
