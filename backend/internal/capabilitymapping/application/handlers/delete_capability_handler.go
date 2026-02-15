package handlers

import (
	"context"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/domain/aggregates"
	"easi/backend/internal/capabilitymapping/domain/events"
	"easi/backend/internal/capabilitymapping/domain/services"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	"easi/backend/internal/shared/cqrs"
)

type DeleteCapabilityRepository interface {
	GetByID(ctx context.Context, id string) (*aggregates.Capability, error)
	Save(ctx context.Context, capability *aggregates.Capability) error
}

type DeleteCapabilityRealizationReadModel interface {
	GetByCapabilityID(ctx context.Context, capabilityID string) ([]readmodels.RealizationDTO, error)
}

type DeleteCapabilityHandler struct {
	repository           DeleteCapabilityRepository
	deletionService      services.CapabilityDeletionService
	realizationReadModel DeleteCapabilityRealizationReadModel
	capabilityLookup     CapabilityParentLookup
}

func NewDeleteCapabilityHandler(
	repository DeleteCapabilityRepository,
	deletionService services.CapabilityDeletionService,
	realizationReadModel DeleteCapabilityRealizationReadModel,
	capabilityLookup CapabilityParentLookup,
) *DeleteCapabilityHandler {
	return &DeleteCapabilityHandler{
		repository:           repository,
		deletionService:      deletionService,
		realizationReadModel: realizationReadModel,
		capabilityLookup:     capabilityLookup,
	}
}

func (h *DeleteCapabilityHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.DeleteCapability)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	if err := h.validateDeletion(ctx, command.ID); err != nil {
		return cqrs.EmptyResult(), err
	}

	capability, err := h.repository.GetByID(ctx, command.ID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.deleteAndCleanupInheritance(ctx, capability); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.EmptyResult(), h.repository.Save(ctx, capability)
}

func (h *DeleteCapabilityHandler) validateDeletion(ctx context.Context, id string) error {
	capabilityID, err := valueobjects.NewCapabilityIDFromString(id)
	if err != nil {
		return err
	}
	return h.deletionService.CanDelete(ctx, capabilityID)
}

func (h *DeleteCapabilityHandler) deleteAndCleanupInheritance(ctx context.Context, capability *aggregates.Capability) error {
	removals, err := h.buildInheritanceRemovals(ctx, capability)
	if err != nil {
		return err
	}

	if err := capability.Delete(); err != nil {
		return err
	}

	if len(removals) > 0 {
		capability.RaiseEvent(events.NewCapabilityRealizationsUninherited(capability.ID(), removals))
	}

	return nil
}

func (h *DeleteCapabilityHandler) buildInheritanceRemovals(ctx context.Context, capability *aggregates.Capability) ([]events.RealizationInheritanceRemoval, error) {
	parentID := capability.ParentID().Value()
	if parentID == "" {
		return nil, nil
	}

	realizations, err := h.realizationReadModel.GetByCapabilityID(ctx, capability.ID())
	if err != nil {
		return nil, err
	}
	if len(realizations) == 0 {
		return nil, nil
	}

	ancestorIDs, err := CollectAncestorIDs(ctx, h.capabilityLookup, parentID)
	if err != nil {
		return nil, err
	}
	if len(ancestorIDs) == 0 {
		return nil, nil
	}

	return BuildRealizationRemovals(realizations, ancestorIDs), nil
}
