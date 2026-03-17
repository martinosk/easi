package handlers

import (
	"context"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/domain/events"
	"easi/backend/internal/capabilitymapping/domain/services"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	"easi/backend/internal/shared/cqrs"
)

type CommandDispatcher interface {
	Dispatch(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error)
}

type CascadeHierarchyService interface {
	GetDescendants(ctx context.Context, capabilityID valueobjects.CapabilityID) ([]valueobjects.CapabilityID, error)
}

type CascadeRealizationReadModel interface {
	GetByCapabilityID(ctx context.Context, capabilityID string) ([]readmodels.RealizationDTO, error)
	GetByComponentID(ctx context.Context, componentID string) ([]readmodels.RealizationDTO, error)
}

type CascadeDependencyReadModel interface {
	GetOutgoing(ctx context.Context, capabilityID string) ([]readmodels.DependencyDTO, error)
	GetIncoming(ctx context.Context, capabilityID string) ([]readmodels.DependencyDTO, error)
}

type ComponentDeleter interface {
	DeleteComponent(ctx context.Context, componentID string) error
}

type CascadeDeleteDeps struct {
	Repository       DeleteCapabilityRepository
	HierarchyService CascadeHierarchyService
	RealizationRM    CascadeRealizationReadModel
	DependencyRM     CascadeDependencyReadModel
	CommandBus       CommandDispatcher
	CapabilityLookup CapabilityParentLookup
	ComponentDeleter ComponentDeleter
}

type CascadeDeleteCapabilityHandler struct {
	deps CascadeDeleteDeps
}

func NewCascadeDeleteCapabilityHandler(deps CascadeDeleteDeps) *CascadeDeleteCapabilityHandler {
	return &CascadeDeleteCapabilityHandler{deps: deps}
}

func (h *CascadeDeleteCapabilityHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.CascadeDeleteCapability)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	scope, err := h.buildScope(ctx, command.ID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if scope.HasDescendants() && !command.Cascade {
		return cqrs.EmptyResult(), services.ErrCascadeRequiredForChildCapabilities
	}

	return cqrs.EmptyResult(), h.executeCascade(ctx, scope, command.DeleteRealisingApplications)
}

func (h *CascadeDeleteCapabilityHandler) executeCascade(ctx context.Context, scope valueobjects.DeletionScope, deleteApps bool) error {
	exclusiveComponentIDs, err := collectExclusiveComponentIDs(ctx, h.deps.RealizationRM, scope)
	if err != nil {
		return err
	}

	if err := h.dispatchRealizationDeletes(ctx, scope); err != nil {
		return err
	}

	if err := h.dispatchDependencyDeletes(ctx, scope); err != nil {
		return err
	}

	if err := h.deleteCapabilitiesBottomUp(ctx, scope); err != nil {
		return err
	}

	if deleteApps {
		h.deleteExclusiveComponents(ctx, exclusiveComponentIDs)
	}

	return nil
}

func (h *CascadeDeleteCapabilityHandler) buildScope(ctx context.Context, id string) (valueobjects.DeletionScope, error) {
	rootID, err := valueobjects.NewCapabilityIDFromString(id)
	if err != nil {
		return valueobjects.DeletionScope{}, err
	}

	if _, err := h.deps.Repository.GetByID(ctx, rootID.Value()); err != nil {
		return valueobjects.DeletionScope{}, err
	}

	descendants, err := h.deps.HierarchyService.GetDescendants(ctx, rootID)
	if err != nil {
		return valueobjects.DeletionScope{}, err
	}

	return valueobjects.NewDeletionScope(rootID, descendants), nil
}

func collectExclusiveComponentIDs(ctx context.Context, rm CascadeRealizationReadModel, scope valueobjects.DeletionScope) ([]string, error) {
	allRealizations, err := collectAllRealizations(ctx, rm, scope)
	if err != nil {
		return nil, err
	}

	classified := make(map[string]bool)
	var exclusive []string

	for _, r := range allRealizations {
		if _, done := classified[r.ComponentID]; done {
			continue
		}
		isExclusive, err := isComponentExclusiveToScope(ctx, rm, r.ComponentID, scope)
		if err != nil {
			return nil, err
		}
		classified[r.ComponentID] = isExclusive
		if isExclusive {
			exclusive = append(exclusive, r.ComponentID)
		}
	}
	return exclusive, nil
}

func collectAllRealizations(ctx context.Context, rm CascadeRealizationReadModel, scope valueobjects.DeletionScope) ([]readmodels.RealizationDTO, error) {
	var all []readmodels.RealizationDTO
	for _, capID := range scope.AllIDs() {
		realizations, err := rm.GetByCapabilityID(ctx, capID.Value())
		if err != nil {
			return nil, err
		}
		all = append(all, realizations...)
	}
	return all, nil
}

func isComponentExclusiveToScope(ctx context.Context, rm CascadeRealizationReadModel, componentID string, scope valueobjects.DeletionScope) (bool, error) {
	allForComponent, err := rm.GetByComponentID(ctx, componentID)
	if err != nil {
		return false, err
	}
	for _, cr := range allForComponent {
		if !scope.Contains(cr.CapabilityID) {
			return false, nil
		}
	}
	return true, nil
}

func (h *CascadeDeleteCapabilityHandler) dispatchRealizationDeletes(ctx context.Context, scope valueobjects.DeletionScope) error {
	realizations, err := collectAllRealizations(ctx, h.deps.RealizationRM, scope)
	if err != nil {
		return err
	}
	for _, r := range realizations {
		if _, err := h.deps.CommandBus.Dispatch(ctx, &commands.DeleteSystemRealization{ID: r.ID}); err != nil {
			return err
		}
	}
	return nil
}

func (h *CascadeDeleteCapabilityHandler) dispatchDependencyDeletes(ctx context.Context, scope valueobjects.DeletionScope) error {
	allDeps, err := h.collectAllDependencies(ctx, scope)
	if err != nil {
		return err
	}
	seen := make(map[string]bool)
	for _, dep := range allDeps {
		if seen[dep.ID] {
			continue
		}
		seen[dep.ID] = true
		if _, err := h.deps.CommandBus.Dispatch(ctx, &commands.DeleteCapabilityDependency{ID: dep.ID}); err != nil {
			return err
		}
	}
	return nil
}

func (h *CascadeDeleteCapabilityHandler) collectAllDependencies(ctx context.Context, scope valueobjects.DeletionScope) ([]readmodels.DependencyDTO, error) {
	var all []readmodels.DependencyDTO
	for _, capID := range scope.AllIDs() {
		deps, err := h.collectDependencies(ctx, capID.Value())
		if err != nil {
			return nil, err
		}
		all = append(all, deps...)
	}
	return all, nil
}

func (h *CascadeDeleteCapabilityHandler) collectDependencies(ctx context.Context, capabilityID string) ([]readmodels.DependencyDTO, error) {
	outgoing, err := h.deps.DependencyRM.GetOutgoing(ctx, capabilityID)
	if err != nil {
		return nil, err
	}
	incoming, err := h.deps.DependencyRM.GetIncoming(ctx, capabilityID)
	if err != nil {
		return nil, err
	}
	return append(outgoing, incoming...), nil
}

func (h *CascadeDeleteCapabilityHandler) deleteCapabilitiesBottomUp(ctx context.Context, scope valueobjects.DeletionScope) error {
	for _, capID := range scope.BottomUp() {
		if err := h.deleteSingleCapability(ctx, capID.Value()); err != nil {
			return err
		}
	}
	return nil
}

func (h *CascadeDeleteCapabilityHandler) deleteSingleCapability(ctx context.Context, id string) error {
	capability, err := h.deps.Repository.GetByID(ctx, id)
	if err != nil {
		return err
	}

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

	return h.deps.Repository.Save(ctx, capability)
}

func (h *CascadeDeleteCapabilityHandler) buildInheritanceRemovals(ctx context.Context, capability interface {
	ID() string
	ParentID() valueobjects.CapabilityID
}) ([]events.RealizationInheritanceRemoval, error) {
	parentID := capability.ParentID().Value()
	if parentID == "" {
		return nil, nil
	}

	realizations, err := h.deps.RealizationRM.GetByCapabilityID(ctx, capability.ID())
	if err != nil {
		return nil, err
	}
	if len(realizations) == 0 {
		return nil, nil
	}

	ancestorIDs, err := CollectAncestorIDs(ctx, h.deps.CapabilityLookup, parentID)
	if err != nil {
		return nil, err
	}
	if len(ancestorIDs) == 0 {
		return nil, nil
	}

	return BuildRealizationRemovals(realizations, ancestorIDs), nil
}

func (h *CascadeDeleteCapabilityHandler) deleteExclusiveComponents(ctx context.Context, componentIDs []string) {
	for _, id := range componentIDs {
		_ = h.deps.ComponentDeleter.DeleteComponent(ctx, id)
	}
}
