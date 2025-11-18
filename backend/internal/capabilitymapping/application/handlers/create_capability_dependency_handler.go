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

type CreateCapabilityDependencyHandler struct {
	dependencyRepository *repositories.DependencyRepository
	capabilityRepository *repositories.CapabilityRepository
}

func NewCreateCapabilityDependencyHandler(
	dependencyRepository *repositories.DependencyRepository,
	capabilityRepository *repositories.CapabilityRepository,
) *CreateCapabilityDependencyHandler {
	return &CreateCapabilityDependencyHandler{
		dependencyRepository: dependencyRepository,
		capabilityRepository: capabilityRepository,
	}
}

func (h *CreateCapabilityDependencyHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(*commands.CreateCapabilityDependency)
	if !ok {
		return cqrs.ErrInvalidCommand
	}

	sourceCapabilityID, err := valueobjects.NewCapabilityIDFromString(command.SourceCapabilityID)
	if err != nil {
		return err
	}

	targetCapabilityID, err := valueobjects.NewCapabilityIDFromString(command.TargetCapabilityID)
	if err != nil {
		return err
	}

	_, err = h.capabilityRepository.GetByID(ctx, sourceCapabilityID.Value())
	if err != nil {
		if errors.Is(err, repositories.ErrCapabilityNotFound) {
			return ErrSourceCapabilityNotFound
		}
		return err
	}

	_, err = h.capabilityRepository.GetByID(ctx, targetCapabilityID.Value())
	if err != nil {
		if errors.Is(err, repositories.ErrCapabilityNotFound) {
			return ErrTargetCapabilityNotFound
		}
		return err
	}

	dependencyType, err := valueobjects.NewDependencyType(command.DependencyType)
	if err != nil {
		return err
	}

	description := valueobjects.NewDescription(command.Description)

	dependency, err := aggregates.NewCapabilityDependency(
		sourceCapabilityID,
		targetCapabilityID,
		dependencyType,
		description,
	)
	if err != nil {
		return err
	}

	command.ID = dependency.ID()

	return h.dependencyRepository.Save(ctx, dependency)
}
