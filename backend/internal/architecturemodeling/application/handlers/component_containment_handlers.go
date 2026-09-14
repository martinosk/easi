package handlers

import (
	"context"
	"errors"

	"easi/backend/internal/architecturemodeling/application/commands"
	"easi/backend/internal/architecturemodeling/application/readmodels"
	"easi/backend/internal/architecturemodeling/domain/aggregates"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	"easi/backend/internal/shared/cqrs"
)

var (
	ErrPartComponentNotFound   = errors.New("component to attach does not exist")
	ErrParentComponentNotFound = errors.New("referenced parent component does not exist")
)

type ComponentContainmentsRepository interface {
	GetByID(ctx context.Context, id string) (*aggregates.ComponentContainments, error)
	Save(ctx context.Context, containments *aggregates.ComponentContainments) error
}

type ContainmentsAggregateLookup interface {
	FindContainmentsAggregateID(ctx context.Context) (string, bool, error)
}

type ComponentRecordReader interface {
	GetByID(ctx context.Context, id string) (*readmodels.ApplicationComponentDTO, error)
}

type containmentCommandBase struct {
	containments ComponentContainmentsRepository
	lookup       ContainmentsAggregateLookup
}

func (b containmentCommandBase) mutateContainments(ctx context.Context, mutate func(*aggregates.ComponentContainments) error) (cqrs.CommandResult, error) {
	containments, err := b.loadContainments(ctx)
	if err != nil {
		return cqrs.EmptyResult(), err
	}
	if err := mutate(containments); err != nil {
		return cqrs.EmptyResult(), err
	}
	if err := b.containments.Save(ctx, containments); err != nil {
		return cqrs.EmptyResult(), err
	}
	return cqrs.EmptyResult(), nil
}

func (b containmentCommandBase) loadContainments(ctx context.Context) (*aggregates.ComponentContainments, error) {
	id, found, err := b.lookup.FindContainmentsAggregateID(ctx)
	if err != nil {
		return nil, err
	}
	if !found {
		return aggregates.NewComponentContainments(), nil
	}
	return b.containments.GetByID(ctx, id)
}

type AttachComponentHandler struct {
	containmentCommandBase
	components ComponentRecordReader
}

func NewAttachComponentHandler(
	containments ComponentContainmentsRepository,
	lookup ContainmentsAggregateLookup,
	components ComponentRecordReader,
) *AttachComponentHandler {
	return &AttachComponentHandler{
		containmentCommandBase: containmentCommandBase{containments: containments, lookup: lookup},
		components:             components,
	}
}

func (h *AttachComponentHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.AttachComponent)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}
	part, err := valueobjects.NewComponentIDFromString(command.ComponentID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}
	parent, err := valueobjects.NewComponentIDFromString(command.ParentID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}
	kind, err := valueobjects.NewContainmentKind(command.Kind)
	if err != nil {
		return cqrs.EmptyResult(), err
	}
	if err := h.requireComponent(ctx, part.Value(), ErrPartComponentNotFound); err != nil {
		return cqrs.EmptyResult(), err
	}
	if err := h.requireComponent(ctx, parent.Value(), ErrParentComponentNotFound); err != nil {
		return cqrs.EmptyResult(), err
	}
	return h.mutateContainments(ctx, func(containments *aggregates.ComponentContainments) error {
		return containments.Attach(part, parent, kind)
	})
}

func (h *AttachComponentHandler) requireComponent(ctx context.Context, id string, missing error) error {
	component, err := h.components.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if component == nil {
		return missing
	}
	return nil
}

type DetachComponentHandler struct {
	containmentCommandBase
}

func NewDetachComponentHandler(containments ComponentContainmentsRepository, lookup ContainmentsAggregateLookup) *DetachComponentHandler {
	return &DetachComponentHandler{containmentCommandBase: containmentCommandBase{containments: containments, lookup: lookup}}
}

func (h *DetachComponentHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.DetachComponent)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}
	part, err := valueobjects.NewComponentIDFromString(command.ComponentID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}
	return h.mutateContainments(ctx, func(containments *aggregates.ComponentContainments) error {
		return containments.Detach(part)
	})
}
