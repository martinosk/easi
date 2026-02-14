package handlers

import (
	"context"

	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/valuestreams/application/commands"
	"easi/backend/internal/valuestreams/application/gateways"
	"easi/backend/internal/valuestreams/domain/aggregates"
	"easi/backend/internal/valuestreams/domain/valueobjects"
)

type AddStageCapabilityRepository interface {
	GetByID(ctx context.Context, id string) (*aggregates.ValueStream, error)
	Save(ctx context.Context, vs *aggregates.ValueStream) error
}

type AddStageCapabilityHandler struct {
	repository        AddStageCapabilityRepository
	capabilityGateway gateways.CapabilityGateway
}

func NewAddStageCapabilityHandler(
	repository AddStageCapabilityRepository,
	capabilityGateway gateways.CapabilityGateway,
) *AddStageCapabilityHandler {
	return &AddStageCapabilityHandler{
		repository:        repository,
		capabilityGateway: capabilityGateway,
	}
}

func (h *AddStageCapabilityHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.AddStageCapability)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	vs, err := h.repository.GetByID(ctx, command.ValueStreamID)
	if err != nil {
		return cqrs.EmptyResult(), mapRepositoryError(err)
	}

	capRef, err := h.resolveCapability(ctx, command.CapabilityID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	stageID, err := valueobjects.NewStageIDFromString(command.StageID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := vs.AddCapabilityToStage(stageID, capRef); err != nil {
		return cqrs.EmptyResult(), mapStageError(err)
	}

	if err := h.repository.Save(ctx, vs); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.EmptyResult(), nil
}

func (h *AddStageCapabilityHandler) resolveCapability(ctx context.Context, capabilityID string) (valueobjects.CapabilityRef, error) {
	exists, err := h.capabilityGateway.CapabilityExists(ctx, capabilityID)
	if err != nil {
		return valueobjects.CapabilityRef{}, err
	}
	if !exists {
		return valueobjects.CapabilityRef{}, ErrCapabilityNotFound
	}
	return valueobjects.NewCapabilityRef(capabilityID)
}
