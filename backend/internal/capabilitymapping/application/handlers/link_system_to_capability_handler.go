package handlers

import (
	"context"
	"errors"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/domain/events"
	"easi/backend/internal/capabilitymapping/domain/aggregates"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	"easi/backend/internal/capabilitymapping/infrastructure/architecturemodeling"
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
	Save(ctx context.Context, capability *aggregates.Capability) error
}

type LinkSystemComponentReadModel interface {
	GetByID(ctx context.Context, id string) (*architecturemodeling.ComponentDTO, error)
}

type LinkSystemCapabilityReadModel interface {
	GetByID(ctx context.Context, id string) (*readmodels.CapabilityDTO, error)
}

type LinkSystemToCapabilityHandler struct {
	realizationRepository LinkSystemRealizationRepository
	capabilityRepository  LinkSystemCapabilityRepository
	capabilityReadModel   LinkSystemCapabilityReadModel
	componentReadModel    LinkSystemComponentReadModel
}

func NewLinkSystemToCapabilityHandler(
	realizationRepository LinkSystemRealizationRepository,
	capabilityRepository LinkSystemCapabilityRepository,
	capabilityReadModel LinkSystemCapabilityReadModel,
	componentReadModel LinkSystemComponentReadModel,
) *LinkSystemToCapabilityHandler {
	return &LinkSystemToCapabilityHandler{
		realizationRepository: realizationRepository,
		capabilityRepository:  capabilityRepository,
		capabilityReadModel:   capabilityReadModel,
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

	capability, err := h.capabilityRepository.GetByID(ctx, capabilityID.Value())
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

	additions, err := h.buildInheritanceAdditions(ctx, capability, realization, component.Name)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if len(additions) > 0 {
		capability.RaiseEvent(events.NewCapabilityRealizationsInherited(capability.ID(), additions))
		if err := h.capabilityRepository.Save(ctx, capability); err != nil {
			return cqrs.EmptyResult(), err
		}
	}

	return cqrs.NewResult(realization.ID()), nil
}

func (h *LinkSystemToCapabilityHandler) buildInheritanceAdditions(ctx context.Context, capability *aggregates.Capability, realization *aggregates.CapabilityRealization, componentName string) ([]events.InheritedRealization, error) {
	ancestorIDs, err := h.collectAncestorIDs(ctx, capability.ParentID().Value())
	if err != nil {
		return nil, err
	}
	if len(ancestorIDs) == 0 {
		return nil, nil
	}

	additions := make([]events.InheritedRealization, 0, len(ancestorIDs))
	for _, ancestorID := range ancestorIDs {
		additions = append(additions, events.InheritedRealization{
			CapabilityID:         ancestorID,
			ComponentID:          realization.ComponentID().Value(),
			ComponentName:        componentName,
			RealizationLevel:     "Full",
			Notes:                "",
			Origin:               "Inherited",
			SourceRealizationID:  realization.ID(),
			SourceCapabilityID:   capability.ID(),
			SourceCapabilityName: capability.Name().Value(),
			LinkedAt:             realization.LinkedAt(),
		})
	}

	return additions, nil
}

func (h *LinkSystemToCapabilityHandler) collectAncestorIDs(ctx context.Context, startID string) ([]string, error) {
	if startID == "" {
		return nil, nil
	}

	ids := []string{}
	visited := map[string]struct{}{}
	currentID := startID

	for currentID != "" {
		if _, seen := visited[currentID]; seen {
			break
		}
		visited[currentID] = struct{}{}
		ids = append(ids, currentID)

		capability, err := h.capabilityReadModel.GetByID(ctx, currentID)
		if err != nil {
			return nil, err
		}
		if capability == nil {
			break
		}
		currentID = capability.ParentID
	}

	return ids, nil
}
