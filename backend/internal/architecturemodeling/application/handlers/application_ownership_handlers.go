package handlers

import (
	"context"
	"errors"
	"fmt"

	"easi/backend/internal/architecturemodeling/application/commands"
	"easi/backend/internal/architecturemodeling/domain/aggregates"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	"easi/backend/internal/shared/cqrs"
)

var ErrOwnerNotFound = errors.New("referenced owner does not exist")

type OwnershipComponentRepository interface {
	GetByID(ctx context.Context, id string) (*aggregates.ApplicationComponent, error)
	Save(ctx context.Context, component *aggregates.ApplicationComponent) error
}

type OwnerExistence interface {
	Exists(ctx context.Context, id string) (bool, error)
}

type ownerReference struct {
	ComponentID string
	OwnerKind   string
	OwnerID     string
}

type ownerTransition struct {
	repository OwnershipComponentRepository
	users      OwnerExistence
	teams      OwnerExistence
}

func (t ownerTransition) run(
	ctx context.Context,
	reference ownerReference,
	establish func(*aggregates.ApplicationComponent, valueobjects.OwnerReference) error,
) (cqrs.CommandResult, error) {
	owner, err := t.resolve(ctx, reference.OwnerKind, reference.OwnerID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}
	return mutateComponentOwnership(ctx, t.repository, reference.ComponentID, func(component *aggregates.ApplicationComponent) error {
		return establish(component, owner)
	})
}

func (t ownerTransition) resolve(ctx context.Context, kind, id string) (valueobjects.OwnerReference, error) {
	ref, err := valueobjects.NewOwnerReference(kind, id)
	if err != nil {
		return valueobjects.OwnerReference{}, err
	}

	directory := t.users
	if ref.IsTeam() {
		directory = t.teams
	}
	exists, err := directory.Exists(ctx, ref.ID())
	if err != nil {
		return valueobjects.OwnerReference{}, fmt.Errorf("check owner %s %s: %w", ref.Kind(), ref.ID(), err)
	}
	if !exists {
		return valueobjects.OwnerReference{}, fmt.Errorf("%w: %s %s", ErrOwnerNotFound, ref.Kind(), ref.ID())
	}
	return ref, nil
}

func mutateComponentOwnership(
	ctx context.Context,
	repository OwnershipComponentRepository,
	componentID string,
	mutate func(*aggregates.ApplicationComponent) error,
) (cqrs.CommandResult, error) {
	component, err := repository.GetByID(ctx, componentID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := mutate(component); err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := repository.Save(ctx, component); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.EmptyResult(), nil
}

type NominateApplicationComponentOwnerHandler struct {
	transition ownerTransition
}

func NewNominateApplicationComponentOwnerHandler(repository OwnershipComponentRepository, users, teams OwnerExistence) *NominateApplicationComponentOwnerHandler {
	return &NominateApplicationComponentOwnerHandler{transition: ownerTransition{repository: repository, users: users, teams: teams}}
}

func (h *NominateApplicationComponentOwnerHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.NominateApplicationComponentOwner)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}
	reference := ownerReference{ComponentID: command.ComponentID, OwnerKind: command.OwnerKind, OwnerID: command.OwnerID}
	return h.transition.run(ctx, reference, (*aggregates.ApplicationComponent).NominateOwner)
}

type AssignApplicationComponentOwnerHandler struct {
	transition ownerTransition
}

func NewAssignApplicationComponentOwnerHandler(repository OwnershipComponentRepository, users, teams OwnerExistence) *AssignApplicationComponentOwnerHandler {
	return &AssignApplicationComponentOwnerHandler{transition: ownerTransition{repository: repository, users: users, teams: teams}}
}

func (h *AssignApplicationComponentOwnerHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.AssignApplicationComponentOwner)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}
	reference := ownerReference{ComponentID: command.ComponentID, OwnerKind: command.OwnerKind, OwnerID: command.OwnerID}
	return h.transition.run(ctx, reference, (*aggregates.ApplicationComponent).AssignOwner)
}

type ConfirmApplicationComponentOwnershipHandler struct {
	repository OwnershipComponentRepository
}

func NewConfirmApplicationComponentOwnershipHandler(repository OwnershipComponentRepository) *ConfirmApplicationComponentOwnershipHandler {
	return &ConfirmApplicationComponentOwnershipHandler{repository: repository}
}

func (h *ConfirmApplicationComponentOwnershipHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.ConfirmApplicationComponentOwnership)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}
	return mutateComponentOwnership(ctx, h.repository, command.ComponentID, (*aggregates.ApplicationComponent).ConfirmOwnership)
}

type ClearApplicationComponentOwnershipHandler struct {
	repository OwnershipComponentRepository
}

func NewClearApplicationComponentOwnershipHandler(repository OwnershipComponentRepository) *ClearApplicationComponentOwnershipHandler {
	return &ClearApplicationComponentOwnershipHandler{repository: repository}
}

func (h *ClearApplicationComponentOwnershipHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.ClearApplicationComponentOwnership)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}
	return mutateComponentOwnership(ctx, h.repository, command.ComponentID, (*aggregates.ApplicationComponent).ClearOwnership)
}
