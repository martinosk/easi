package handlers

import (
	"context"
	"errors"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/domain/aggregates"
	"easi/backend/internal/capabilitymapping/domain/events"
	"easi/backend/internal/capabilitymapping/domain/services"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	"easi/backend/internal/capabilitymapping/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

var (
	ErrParentCapabilityNotFound = errors.New("parent capability not found")
)

type ChangeCapabilityParentRepository interface {
	GetByID(ctx context.Context, id string) (*aggregates.Capability, error)
	Save(ctx context.Context, capability *aggregates.Capability) error
}

type ChangeCapabilityParentReadModel interface {
	GetChildren(ctx context.Context, parentID string) ([]readmodels.CapabilityDTO, error)
	GetByID(ctx context.Context, id string) (*readmodels.CapabilityDTO, error)
}

type ChangeCapabilityParentRealizationReadModel interface {
	GetByCapabilityID(ctx context.Context, capabilityID string) ([]readmodels.RealizationDTO, error)
}

type ChangeCapabilityParentHandler struct {
	repository           ChangeCapabilityParentRepository
	capabilityReadModel  ChangeCapabilityParentReadModel
	realizationReadModel ChangeCapabilityParentRealizationReadModel
	reparentingService   services.CapabilityReparentingService
}

func NewChangeCapabilityParentHandler(
	repository ChangeCapabilityParentRepository,
	capabilityReadModel ChangeCapabilityParentReadModel,
	realizationReadModel ChangeCapabilityParentRealizationReadModel,
	reparentingService services.CapabilityReparentingService,
) *ChangeCapabilityParentHandler {
	return &ChangeCapabilityParentHandler{
		repository:           repository,
		capabilityReadModel:  capabilityReadModel,
		realizationReadModel: realizationReadModel,
		reparentingService:   reparentingService,
	}
}

func (h *ChangeCapabilityParentHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.ChangeCapabilityParent)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	capability, err := h.repository.GetByID(ctx, command.CapabilityID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	newParentID, newLevel, err := h.determineNewParentAndLevel(ctx, command)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.applyParentChange(ctx, capability, newParentID, newLevel); err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.updateDescendantLevels(ctx, command.CapabilityID, newLevel); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.EmptyResult(), nil
}

func (h *ChangeCapabilityParentHandler) applyParentChange(ctx context.Context, capability *aggregates.Capability, newParentID valueobjects.CapabilityID, newLevel valueobjects.CapabilityLevel) error {
	oldParentID := capability.ParentID().Value()

	additions, removals, err := h.buildInheritanceChanges(ctx, capability.ID(), oldParentID, newParentID.Value())
	if err != nil {
		return err
	}

	if err := capability.ChangeParent(newParentID, newLevel); err != nil {
		return err
	}

	raiseInheritanceEvents(capability, additions, removals)

	return h.repository.Save(ctx, capability)
}

func raiseInheritanceEvents(capability *aggregates.Capability, additions []events.InheritedRealization, removals []events.RealizationInheritanceRemoval) {
	if len(additions) > 0 {
		capability.RaiseEvent(events.NewCapabilityRealizationsInherited(capability.ID(), additions))
	}
	if len(removals) > 0 {
		capability.RaiseEvent(events.NewCapabilityRealizationsUninherited(capability.ID(), removals))
	}
}

func (h *ChangeCapabilityParentHandler) determineNewParentAndLevel(ctx context.Context, command *commands.ChangeCapabilityParent) (valueobjects.CapabilityID, valueobjects.CapabilityLevel, error) {
	capabilityID, err := valueobjects.NewCapabilityIDFromString(command.CapabilityID)
	if err != nil {
		return valueobjects.CapabilityID{}, "", err
	}

	if command.NewParentID == "" {
		level, err := h.reparentingService.DetermineNewLevel(ctx, capabilityID, valueobjects.CapabilityID{}, valueobjects.LevelL1)
		return valueobjects.CapabilityID{}, level, err
	}

	newParentID, err := valueobjects.NewCapabilityIDFromString(command.NewParentID)
	if err != nil {
		return valueobjects.CapabilityID{}, "", err
	}

	parent, err := h.repository.GetByID(ctx, command.NewParentID)
	if err != nil {
		if errors.Is(err, repositories.ErrCapabilityNotFound) {
			return valueobjects.CapabilityID{}, "", ErrParentCapabilityNotFound
		}
		return valueobjects.CapabilityID{}, "", err
	}

	newLevel, err := h.reparentingService.DetermineNewLevel(ctx, capabilityID, newParentID, parent.Level())
	if err != nil {
		return valueobjects.CapabilityID{}, "", err
	}

	return newParentID, newLevel, nil
}

func (h *ChangeCapabilityParentHandler) buildInheritanceChanges(ctx context.Context, capabilityID, oldParentID, newParentID string) ([]events.InheritedRealization, []events.RealizationInheritanceRemoval, error) {
	if oldParentID == newParentID {
		return nil, nil, nil
	}

	realizations, err := h.realizationReadModel.GetByCapabilityID(ctx, capabilityID)
	if err != nil {
		return nil, nil, err
	}
	if len(realizations) == 0 {
		return nil, nil, nil
	}

	capability, err := h.capabilityReadModel.GetByID(ctx, capabilityID)
	if err != nil {
		return nil, nil, err
	}
	if capability == nil {
		return nil, nil, nil
	}

	additions, err := h.buildInheritanceAdditions(ctx, newParentID, capability, realizations)
	if err != nil {
		return nil, nil, err
	}

	removals, err := h.buildInheritanceRemovals(ctx, oldParentID, realizations)
	if err != nil {
		return nil, nil, err
	}

	return additions, removals, nil
}

func (h *ChangeCapabilityParentHandler) buildInheritanceAdditions(ctx context.Context, newParentID string, capability *readmodels.CapabilityDTO, realizations []readmodels.RealizationDTO) ([]events.InheritedRealization, error) {
	if newParentID == "" {
		return nil, nil
	}

	ancestorIDs, err := h.collectAncestorIDs(ctx, newParentID)
	if err != nil {
		return nil, err
	}
	if len(ancestorIDs) == 0 {
		return nil, nil
	}

	additions := make([]events.InheritedRealization, 0, len(ancestorIDs)*len(realizations))
	for _, realization := range realizations {
		sourceID, sourceCapabilityID, sourceCapabilityName := buildInheritanceSource(realization, capability)
		for _, ancestorID := range ancestorIDs {
			additions = append(additions, events.InheritedRealization{
				CapabilityID:         ancestorID,
				ComponentID:          realization.ComponentID,
				ComponentName:        realization.ComponentName,
				RealizationLevel:     "Full",
				Notes:                "",
				Origin:               "Inherited",
				SourceRealizationID:  sourceID,
				SourceCapabilityID:   sourceCapabilityID,
				SourceCapabilityName: sourceCapabilityName,
				LinkedAt:             realization.LinkedAt,
			})
		}
	}

	return additions, nil
}

func (h *ChangeCapabilityParentHandler) buildInheritanceRemovals(ctx context.Context, oldParentID string, realizations []readmodels.RealizationDTO) ([]events.RealizationInheritanceRemoval, error) {
	if oldParentID == "" {
		return nil, nil
	}

	ancestorIDs, err := h.collectAncestorIDs(ctx, oldParentID)
	if err != nil {
		return nil, err
	}
	if len(ancestorIDs) == 0 {
		return nil, nil
	}

	return BuildRealizationRemovals(realizations, ancestorIDs), nil
}

func (h *ChangeCapabilityParentHandler) collectAncestorIDs(ctx context.Context, startID string) ([]string, error) {
	return CollectAncestorIDs(ctx, h.capabilityReadModel, startID)
}

func resolveSourceRealizationID(realization readmodels.RealizationDTO) string {
	if realization.Origin == "Inherited" && realization.SourceRealizationID != "" {
		return realization.SourceRealizationID
	}
	return realization.ID
}

func buildInheritanceSource(realization readmodels.RealizationDTO, capability *readmodels.CapabilityDTO) (string, string, string) {
	if realization.Origin == "Inherited" && realization.SourceRealizationID != "" {
		return realization.SourceRealizationID, realization.SourceCapabilityID, realization.SourceCapabilityName
	}

	return realization.ID, capability.ID, capability.Name
}

func (h *ChangeCapabilityParentHandler) updateDescendantLevels(ctx context.Context, parentID string, parentLevel valueobjects.CapabilityLevel) error {
	children, err := h.capabilityReadModel.GetChildren(ctx, parentID)
	if err != nil {
		return err
	}

	if len(children) == 0 {
		return nil
	}

	childLevel, err := h.reparentingService.CalculateChildLevel(parentLevel)
	if err != nil {
		return err
	}

	for _, child := range children {
		if err := h.updateChildLevel(ctx, child.ID, childLevel); err != nil {
			return err
		}
	}

	return nil
}

func (h *ChangeCapabilityParentHandler) updateChildLevel(ctx context.Context, childID string, childLevel valueobjects.CapabilityLevel) error {
	childCapability, err := h.repository.GetByID(ctx, childID)
	if err != nil {
		return err
	}

	if err := childCapability.ChangeLevel(childLevel); err != nil {
		return err
	}

	if err := h.repository.Save(ctx, childCapability); err != nil {
		return err
	}

	return h.updateDescendantLevels(ctx, childID, childLevel)
}
