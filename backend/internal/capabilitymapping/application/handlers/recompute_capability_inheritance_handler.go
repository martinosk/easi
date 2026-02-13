package handlers

import (
	"context"
	"sort"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/domain/aggregates"
	"easi/backend/internal/capabilitymapping/domain/events"
	"easi/backend/internal/shared/cqrs"
)

type RecomputeCapabilityInheritanceRepository interface {
	GetByID(ctx context.Context, id string) (*aggregates.Capability, error)
	Save(ctx context.Context, capability *aggregates.Capability) error
}

type RecomputeCapabilityInheritanceCapabilityReadModel interface {
	GetByID(ctx context.Context, id string) (*readmodels.CapabilityDTO, error)
}

type RecomputeCapabilityInheritanceRealizationReadModel interface {
	GetByCapabilityID(ctx context.Context, capabilityID string) ([]readmodels.RealizationDTO, error)
	GetInheritedCapabilityIDsBySourceRealizationID(ctx context.Context, sourceRealizationID string) ([]string, error)
}

type RecomputeCapabilityInheritanceHandler struct {
	repository           RecomputeCapabilityInheritanceRepository
	capabilityReadModel  RecomputeCapabilityInheritanceCapabilityReadModel
	realizationReadModel RecomputeCapabilityInheritanceRealizationReadModel
}

func NewRecomputeCapabilityInheritanceHandler(
	repository RecomputeCapabilityInheritanceRepository,
	capabilityReadModel RecomputeCapabilityInheritanceCapabilityReadModel,
	realizationReadModel RecomputeCapabilityInheritanceRealizationReadModel,
) *RecomputeCapabilityInheritanceHandler {
	return &RecomputeCapabilityInheritanceHandler{
		repository:           repository,
		capabilityReadModel:  capabilityReadModel,
		realizationReadModel: realizationReadModel,
	}
}

func (h *RecomputeCapabilityInheritanceHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.RecomputeCapabilityInheritance)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	capability, err := h.repository.GetByID(ctx, command.CapabilityID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	realizations, err := h.realizationReadModel.GetByCapabilityID(ctx, command.CapabilityID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	ancestorIDs, err := h.collectAncestorIDs(ctx, capability.ParentID().Value())
	if err != nil {
		return cqrs.EmptyResult(), err
	}
	desiredAncestorSet := make(map[string]struct{}, len(ancestorIDs))
	for _, ancestorID := range ancestorIDs {
		desiredAncestorSet[ancestorID] = struct{}{}
	}

	additions := []events.InheritedRealization{}
	removals := []events.RealizationInheritanceRemoval{}

	for _, realization := range realizations {
		if realization.Origin != "Direct" {
			continue
		}

		sourceID := realization.ID
		currentIDs, err := h.realizationReadModel.GetInheritedCapabilityIDsBySourceRealizationID(ctx, sourceID)
		if err != nil {
			return cqrs.EmptyResult(), err
		}
		currentSet := make(map[string]struct{}, len(currentIDs))
		for _, id := range currentIDs {
			currentSet[id] = struct{}{}
		}

		for _, ancestorID := range ancestorIDs {
			if _, exists := currentSet[ancestorID]; exists {
				continue
			}
			additions = append(additions, events.InheritedRealization{
				CapabilityID:         ancestorID,
				ComponentID:          realization.ComponentID,
				ComponentName:        realization.ComponentName,
				RealizationLevel:     "Full",
				Notes:                "",
				Origin:               "Inherited",
				SourceRealizationID:  sourceID,
				SourceCapabilityID:   capability.ID(),
				SourceCapabilityName: capability.Name().Value(),
				LinkedAt:             realization.LinkedAt,
			})
		}

		toRemove := make([]string, 0)
		for _, currentID := range currentIDs {
			if _, keep := desiredAncestorSet[currentID]; !keep {
				toRemove = append(toRemove, currentID)
			}
		}
		if len(toRemove) > 0 {
			sort.Strings(toRemove)
			removals = append(removals, events.RealizationInheritanceRemoval{
				SourceRealizationID: sourceID,
				CapabilityIDs:       toRemove,
			})
		}
	}

	if len(additions) == 0 && len(removals) == 0 {
		return cqrs.EmptyResult(), nil
	}

	if len(additions) > 0 {
		capability.RaiseEvent(events.NewCapabilityRealizationsInherited(capability.ID(), additions))
	}
	if len(removals) > 0 {
		capability.RaiseEvent(events.NewCapabilityRealizationsUninherited(capability.ID(), removals))
	}

	if err := h.repository.Save(ctx, capability); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.EmptyResult(), nil
}

func (h *RecomputeCapabilityInheritanceHandler) collectAncestorIDs(ctx context.Context, startID string) ([]string, error) {
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
