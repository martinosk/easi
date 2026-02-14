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

	additions, removals, err := h.computeInheritanceDiff(ctx, capability, realizations, ancestorIDs)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	raiseInheritanceEvents(capability, additions, removals)

	if len(capability.GetUncommittedChanges()) > 0 {
		if err := h.repository.Save(ctx, capability); err != nil {
			return cqrs.EmptyResult(), err
		}
	}

	return cqrs.EmptyResult(), nil
}

func (h *RecomputeCapabilityInheritanceHandler) computeInheritanceDiff(ctx context.Context, capability *aggregates.Capability, realizations []readmodels.RealizationDTO, ancestorIDs []string) ([]events.InheritedRealization, []events.RealizationInheritanceRemoval, error) {
	desiredAncestorSet := toSet(ancestorIDs)
	var additions []events.InheritedRealization
	var removals []events.RealizationInheritanceRemoval

	for _, realization := range realizations {
		if realization.Origin != "Direct" {
			continue
		}

		currentIDs, err := h.realizationReadModel.GetInheritedCapabilityIDsBySourceRealizationID(ctx, realization.ID)
		if err != nil {
			return nil, nil, err
		}
		currentSet := toSet(currentIDs)

		additions = append(additions, findMissingInheritances(ancestorIDs, currentSet, capability, realization)...)
		if removal := findStaleInheritances(currentIDs, desiredAncestorSet, realization.ID); removal != nil {
			removals = append(removals, *removal)
		}
	}

	return additions, removals, nil
}

func findMissingInheritances(ancestorIDs []string, currentSet map[string]struct{}, capability *aggregates.Capability, realization readmodels.RealizationDTO) []events.InheritedRealization {
	var additions []events.InheritedRealization
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
			SourceRealizationID:  realization.ID,
			SourceCapabilityID:   capability.ID(),
			SourceCapabilityName: capability.Name().Value(),
			LinkedAt:             realization.LinkedAt,
		})
	}
	return additions
}

func findStaleInheritances(currentIDs []string, desiredAncestorSet map[string]struct{}, sourceRealizationID string) *events.RealizationInheritanceRemoval {
	var toRemove []string
	for _, currentID := range currentIDs {
		if _, keep := desiredAncestorSet[currentID]; !keep {
			toRemove = append(toRemove, currentID)
		}
	}
	if len(toRemove) == 0 {
		return nil
	}
	sort.Strings(toRemove)
	return &events.RealizationInheritanceRemoval{
		SourceRealizationID: sourceRealizationID,
		CapabilityIDs:       toRemove,
	}
}

func toSet(ids []string) map[string]struct{} {
	set := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		set[id] = struct{}{}
	}
	return set
}

func (h *RecomputeCapabilityInheritanceHandler) collectAncestorIDs(ctx context.Context, startID string) ([]string, error) {
	return CollectAncestorIDs(ctx, h.capabilityReadModel, startID)
}
