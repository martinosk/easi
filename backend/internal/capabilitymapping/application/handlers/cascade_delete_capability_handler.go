package handlers

import (
	"context"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/domain/aggregates"
	"easi/backend/internal/capabilitymapping/domain/services"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	"easi/backend/internal/shared/cqrs"
)

type CascadeDeleteRepository interface {
	GetByID(ctx context.Context, id string) (*aggregates.Capability, error)
	Save(ctx context.Context, capability *aggregates.Capability) error
}

type CommandDispatcher interface {
	Dispatch(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error)
}

type CascadeHierarchyService interface {
	GetDescendants(ctx context.Context, capabilityID valueobjects.CapabilityID) ([]valueobjects.CapabilityID, error)
}

type CascadeRealizationReadModel interface {
	GetByCapabilityID(ctx context.Context, capabilityID string) ([]readmodels.RealizationDTO, error)
}

type CascadeDependencyReadModel interface {
	GetOutgoing(ctx context.Context, capabilityID string) ([]readmodels.DependencyDTO, error)
	GetIncoming(ctx context.Context, capabilityID string) ([]readmodels.DependencyDTO, error)
}

type CascadeDeleteDeps struct {
	Repository       CascadeDeleteRepository
	HierarchyService CascadeHierarchyService
	RealizationRM    CascadeRealizationReadModel
	DependencyRM     CascadeDependencyReadModel
	CommandBus       CommandDispatcher
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

	return cqrs.EmptyResult(), h.executeCascade(ctx, scope)
}

func (h *CascadeDeleteCapabilityHandler) executeCascade(ctx context.Context, scope valueobjects.DeletionScope) error {
	if err := h.dispatchRealizationDeletes(ctx, scope); err != nil {
		return err
	}

	if err := h.dispatchDependencyDeletes(ctx, scope); err != nil {
		return err
	}

	return h.deleteCapabilitiesBottomUp(ctx, scope)
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

func (h *CascadeDeleteCapabilityHandler) dispatchRealizationDeletes(ctx context.Context, scope valueobjects.DeletionScope) error {
	realizations, err := collectDirectRealizations(ctx, h.deps.RealizationRM, scope)
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

	if err := capability.Delete(); err != nil {
		return err
	}

	return h.deps.Repository.Save(ctx, capability)
}

func collectDirectRealizations(ctx context.Context, rm CascadeRealizationReadModel, scope valueobjects.DeletionScope) ([]readmodels.RealizationDTO, error) {
	var direct []readmodels.RealizationDTO
	for _, capID := range scope.AllIDs() {
		realizations, err := rm.GetByCapabilityID(ctx, capID.Value())
		if err != nil {
			return nil, err
		}
		direct = append(direct, directOnly(realizations)...)
	}
	return direct, nil
}

func directOnly(realizations []readmodels.RealizationDTO) []readmodels.RealizationDTO {
	var direct []readmodels.RealizationDTO
	for _, r := range realizations {
		if r.Origin == readmodels.OriginDirect {
			direct = append(direct, r)
		}
	}
	return direct
}
