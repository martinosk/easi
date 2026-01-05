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

type LinkSystemRealizationRepository interface {
	Save(ctx context.Context, realization *aggregates.CapabilityRealization) error
}

type LinkSystemCapabilityRepository interface {
	GetByID(ctx context.Context, id string) (*aggregates.Capability, error)
}

type LinkSystemComponentReadModel interface {
	GetByID(ctx context.Context, id string) (*archReadModels.ApplicationComponentDTO, error)
}

type LinkSystemToCapabilityHandler struct {
	realizationRepository LinkSystemRealizationRepository
	capabilityRepository  LinkSystemCapabilityRepository
	componentReadModel    LinkSystemComponentReadModel
}

func NewLinkSystemToCapabilityHandler(
	realizationRepository LinkSystemRealizationRepository,
	capabilityRepository LinkSystemCapabilityRepository,
	componentReadModel LinkSystemComponentReadModel,
) *LinkSystemToCapabilityHandler {
	return &LinkSystemToCapabilityHandler{
		realizationRepository: realizationRepository,
		capabilityRepository:  capabilityRepository,
		componentReadModel:    componentReadModel,
	}
}

func (h *LinkSystemToCapabilityHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.LinkSystemToCapability)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	capabilityID, err := valueobjects.NewCapabilityIDFromString(command.CapabilityID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	componentID, err := valueobjects.NewComponentIDFromString(command.ComponentID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	_, err = h.capabilityRepository.GetByID(ctx, capabilityID.Value())
	if err != nil {
		if errors.Is(err, repositories.ErrCapabilityNotFound) {
			return cqrs.EmptyResult(), ErrCapabilityNotFoundForRealization
		}
		return cqrs.EmptyResult(), err
	}

	component, err := h.componentReadModel.GetByID(ctx, componentID.Value())
	if err != nil {
		return cqrs.EmptyResult(), err
	}
	if component == nil {
		return cqrs.EmptyResult(), ErrComponentNotFound
	}

	realizationLevel, err := valueobjects.NewRealizationLevel(command.RealizationLevel)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	notes, err := valueobjects.NewDescription(command.Notes)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	realization, err := aggregates.NewCapabilityRealization(
		capabilityID,
		componentID,
		component.Name,
		realizationLevel,
		notes,
	)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.realizationRepository.Save(ctx, realization); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.NewResult(realization.ID()), nil
}
