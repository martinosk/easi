package handlers

import (
	"context"
	"errors"

	"easi/backend/internal/valuestreams/application/commands"
	"easi/backend/internal/valuestreams/application/gateways"
	"easi/backend/internal/valuestreams/domain/aggregates"
	"easi/backend/internal/valuestreams/domain/valueobjects"
	"easi/backend/internal/valuestreams/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

var ErrCapabilityNotFound = errors.New("capability not found")

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
		if errors.Is(err, repositories.ErrValueStreamNotFound) {
			return cqrs.EmptyResult(), ErrValueStreamNotFound
		}
		return cqrs.EmptyResult(), err
	}

	exists, err := h.capabilityGateway.CapabilityExists(ctx, command.CapabilityID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}
	if !exists {
		return cqrs.EmptyResult(), ErrCapabilityNotFound
	}

	stageID, err := valueobjects.NewStageIDFromString(command.StageID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	capRef, err := valueobjects.NewCapabilityRef(command.CapabilityID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := vs.AddCapabilityToStage(stageID, capRef); err != nil {
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
