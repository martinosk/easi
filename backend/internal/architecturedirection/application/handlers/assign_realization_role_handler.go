package handlers

import (
	"context"

	"easi/backend/internal/architecturedirection/application/commands"
	"easi/backend/internal/architecturedirection/domain/aggregates"
	"easi/backend/internal/architecturedirection/domain/services"
	"easi/backend/internal/architecturedirection/domain/valueobjects"
	"easi/backend/internal/shared/cqrs"
)

type RealizationRolesRepository interface {
	Save(ctx context.Context, rr *aggregates.RealizationRoles) error
	GetByID(ctx context.Context, id string) (*aggregates.RealizationRoles, error)
}

type ExistingRealizationRolesLookup interface {
	FindAggregateIDForCapability(ctx context.Context, capabilityID string) (string, bool, error)
}

type AssignRealizationRoleHandler struct {
	repo              RealizationRolesRepository
	lookup            ExistingRealizationRolesLookup
	directRealization services.DirectRealizationLookup
}

func NewAssignRealizationRoleHandler(
	repo RealizationRolesRepository,
	lookup ExistingRealizationRolesLookup,
	directRealization services.DirectRealizationLookup,
) *AssignRealizationRoleHandler {
	return &AssignRealizationRoleHandler{
		repo:              repo,
		lookup:            lookup,
		directRealization: directRealization,
	}
}

type assignRealizationRoleInputs struct {
	capability valueobjects.PhysicalCapabilityRef
	component  valueobjects.ApplicationRef
	role       valueobjects.RealizationRole
	assignedBy string
}

func (h *AssignRealizationRoleHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.AssignRealizationRole)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}
	inputs, err := parseAssignRealizationRoleInputs(command)
	if err != nil {
		return cqrs.EmptyResult(), err
	}
	realizationID, err := h.verifyDirectRealizationExists(ctx, inputs)
	if err != nil {
		return cqrs.EmptyResult(), err
	}
	existingID, exists, err := h.lookup.FindAggregateIDForCapability(ctx, inputs.capability.Value())
	if err != nil {
		return cqrs.EmptyResult(), err
	}
	if exists {
		return h.assignToExisting(ctx, existingID, realizationID, inputs)
	}
	return h.createNew(ctx, realizationID, inputs)
}

func (h *AssignRealizationRoleHandler) verifyDirectRealizationExists(ctx context.Context, inputs assignRealizationRoleInputs) (string, error) {
	realizationID, exists, err := h.directRealization(ctx, inputs.capability.Value(), inputs.component.Value())
	if err != nil {
		return "", err
	}
	if !exists {
		return "", services.ErrReferencedEntityNotFound
	}
	return realizationID, nil
}

func (h *AssignRealizationRoleHandler) assignToExisting(
	ctx context.Context,
	existingID string,
	realizationID string,
	inputs assignRealizationRoleInputs,
) (cqrs.CommandResult, error) {
	existing, err := h.repo.GetByID(ctx, existingID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}
	if err := existing.Assign(inputs.component, realizationID, inputs.role, inputs.assignedBy); err != nil {
		return cqrs.EmptyResult(), err
	}
	if err := h.repo.Save(ctx, existing); err != nil {
		return cqrs.EmptyResult(), err
	}
	return cqrs.NewResult(existing.ID()), nil
}

func (h *AssignRealizationRoleHandler) createNew(
	ctx context.Context,
	realizationID string,
	inputs assignRealizationRoleInputs,
) (cqrs.CommandResult, error) {
	rr, err := aggregates.NewRealizationRoles(aggregates.RealizationRolesFacts{
		CapabilityID:  inputs.capability,
		ComponentID:   inputs.component,
		RealizationID: realizationID,
		Role:          inputs.role,
		AssignedBy:    inputs.assignedBy,
	})
	if err != nil {
		return cqrs.EmptyResult(), err
	}
	if err := h.repo.Save(ctx, rr); err != nil {
		return cqrs.EmptyResult(), err
	}
	return cqrs.NewResult(rr.ID()), nil
}

func parseAssignRealizationRoleInputs(command *commands.AssignRealizationRole) (assignRealizationRoleInputs, error) {
	capability, err := valueobjects.NewPhysicalCapabilityRef(command.CapabilityID)
	if err != nil {
		return assignRealizationRoleInputs{}, err
	}
	component, err := valueobjects.NewApplicationRef(command.ComponentID)
	if err != nil {
		return assignRealizationRoleInputs{}, err
	}
	role, err := valueobjects.NewRealizationRole(command.Role)
	if err != nil {
		return assignRealizationRoleInputs{}, err
	}
	return assignRealizationRoleInputs{
		capability: capability,
		component:  component,
		role:       role,
		assignedBy: command.AssignedBy,
	}, nil
}
