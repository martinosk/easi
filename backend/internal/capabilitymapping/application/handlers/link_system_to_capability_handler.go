package handlers

import (
	"context"
	"errors"

	archReadModels "easi/backend/internal/architecturemodeling/application/readmodels"
	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/domain/aggregates"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	"easi/backend/internal/capabilitymapping/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

var (
	ErrComponentNotFound                = errors.New("component not found")
	ErrCapabilityNotFoundForRealization = errors.New("capability not found")
)

type LinkSystemToCapabilityHandler struct {
	realizationRepository *repositories.RealizationRepository
	capabilityRepository  *repositories.CapabilityRepository
	componentReadModel    *archReadModels.ApplicationComponentReadModel
}

func NewLinkSystemToCapabilityHandler(
	realizationRepository *repositories.RealizationRepository,
	capabilityRepository *repositories.CapabilityRepository,
	componentReadModel *archReadModels.ApplicationComponentReadModel,
) *LinkSystemToCapabilityHandler {
	return &LinkSystemToCapabilityHandler{
		realizationRepository: realizationRepository,
		capabilityRepository:  capabilityRepository,
		componentReadModel:    componentReadModel,
	}
}

func (h *LinkSystemToCapabilityHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(*commands.LinkSystemToCapability)
	if !ok {
		return cqrs.ErrInvalidCommand
	}

	capabilityID, err := valueobjects.NewCapabilityIDFromString(command.CapabilityID)
	if err != nil {
		return err
	}

	componentID, err := valueobjects.NewComponentIDFromString(command.ComponentID)
	if err != nil {
		return err
	}

	_, err = h.capabilityRepository.GetByID(ctx, capabilityID.Value())
	if err != nil {
		if errors.Is(err, repositories.ErrCapabilityNotFound) {
			return ErrCapabilityNotFoundForRealization
		}
		return err
	}

	component, err := h.componentReadModel.GetByID(ctx, componentID.Value())
	if err != nil {
		return err
	}
	if component == nil {
		return ErrComponentNotFound
	}

	realizationLevel, err := valueobjects.NewRealizationLevel(command.RealizationLevel)
	if err != nil {
		return err
	}

	notes := valueobjects.NewDescription(command.Notes)

	realization, err := aggregates.NewCapabilityRealization(
		capabilityID,
		componentID,
		realizationLevel,
		notes,
	)
	if err != nil {
		return err
	}

	command.ID = realization.ID()

	return h.realizationRepository.Save(ctx, realization)
}
