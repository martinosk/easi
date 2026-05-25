package handlers

import (
	"context"
	"errors"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/domain/aggregates"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	"easi/backend/internal/capabilitymapping/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

var (
	ErrSourceCapabilityNotFound = errors.New("source capability not found")
	ErrTargetCapabilityNotFound = errors.New("target capability not found")
)

type CreateDependencyRepository interface {
	Save(ctx context.Context, dependency *aggregates.CapabilityDependency) error
}

type CreateDependencyCapabilityRepository interface {
	GetByID(ctx context.Context, id string) (*aggregates.Capability, error)
}

type CreateCapabilityDependencyHandler struct {
	dependencyRepository CreateDependencyRepository
	capabilityRepository CreateDependencyCapabilityRepository
}

func NewCreateCapabilityDependencyHandler(
	dependencyRepository CreateDependencyRepository,
	capabilityRepository CreateDependencyCapabilityRepository,
) *CreateCapabilityDependencyHandler {
	return &CreateCapabilityDependencyHandler{
		dependencyRepository: dependencyRepository,
		capabilityRepository: capabilityRepository,
	}
}

func (h *CreateCapabilityDependencyHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.CreateCapabilityDependency)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	sourceCapabilityID, err := valueobjects.NewCapabilityIDFromString(command.SourceCapabilityID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	targetCapabilityID, err := valueobjects.NewCapabilityIDFromString(command.TargetCapabilityID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	_, err = h.capabilityRepository.GetByID(ctx, sourceCapabilityID.Value())
	if err != nil {
		if errors.Is(err, repositories.ErrCapabilityNotFound) {
			return cqrs.EmptyResult(), ErrSourceCapabilityNotFound
		}
		return cqrs.EmptyResult(), err
	}

	_, err = h.capabilityRepository.GetByID(ctx, targetCapabilityID.Value())
	if err != nil {
		if errors.Is(err, repositories.ErrCapabilityNotFound) {
			return cqrs.EmptyResult(), ErrTargetCapabilityNotFound
		}
		return cqrs.EmptyResult(), err
	}

	dependencyType, err := valueobjects.NewDependencyType(command.DependencyType)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	description, err := valueobjects.NewDescription(command.Description)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	dependency, err := aggregates.NewCapabilityDependency(
		sourceCapabilityID,
		targetCapabilityID,
		dependencyType,
		description,
	)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.dependencyRepository.Save(ctx, dependency); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.NewResult(dependency.ID()), nil
}
