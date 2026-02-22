package handlers

import (
	"context"
	"errors"

	"easi/backend/internal/architectureviews/application/commands"
	"easi/backend/internal/architectureviews/application/ports"
	"easi/backend/internal/architectureviews/domain/aggregates"
	"easi/backend/internal/architectureviews/domain/valueobjects"
	"easi/backend/internal/architectureviews/infrastructure/repositories"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/cqrs"
)

var ErrNotAuthorizedToChangeVisibility = errors.New("not authorized to change view visibility")

type ChangeViewVisibilityHandler struct {
	viewRepository  *repositories.ArchitectureViewRepository
	userRoleChecker ports.UserRoleChecker
}

func NewChangeViewVisibilityHandler(
	viewRepository *repositories.ArchitectureViewRepository,
	userRoleChecker ports.UserRoleChecker,
) *ChangeViewVisibilityHandler {
	return &ChangeViewVisibilityHandler{
		viewRepository:  viewRepository,
		userRoleChecker: userRoleChecker,
	}
}

func (h *ChangeViewVisibilityHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.ChangeViewVisibility)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	actor, ok := sharedctx.GetActor(ctx)
	if !ok {
		return cqrs.EmptyResult(), ErrActorNotFound
	}

	view, err := h.viewRepository.GetByID(ctx, command.ViewID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	ownership := newOwnershipContext(view, actor)
	canEdit, err := h.canActorEditView(ctx, ownership)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.applyVisibilityChange(view, command.IsPrivate, ownership, canEdit); err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.viewRepository.Save(ctx, view); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.EmptyResult(), nil
}

type ownershipContext struct {
	actorID    string
	actorEmail string
	hasNoOwner bool
	isOwner    bool
}

func newOwnershipContext(view *aggregates.ArchitectureView, actor sharedctx.Actor) ownershipContext {
	hasNoOwner := view.Owner().IsEmpty()
	return ownershipContext{
		actorID:    actor.ID,
		actorEmail: actor.Email,
		hasNoOwner: hasNoOwner,
		isOwner:    !hasNoOwner && view.Owner().UserID() == actor.ID,
	}
}

func (o ownershipContext) actorAsOwner() valueobjects.ViewOwner {
	owner, _ := valueobjects.NewViewOwner(o.actorID, o.actorEmail)
	return owner
}

func (h *ChangeViewVisibilityHandler) canActorEditView(ctx context.Context, ownership ownershipContext) (bool, error) {
	if ownership.hasNoOwner || ownership.isOwner {
		return true, nil
	}

	isAdmin, err := h.isActorAdmin(ctx, ownership.actorID)
	if err != nil {
		return false, err
	}
	return isAdmin, nil
}

func (h *ChangeViewVisibilityHandler) applyVisibilityChange(view *aggregates.ArchitectureView, makePrivate bool, ownership ownershipContext, canEdit bool) error {
	if makePrivate {
		return h.makeViewPrivate(view, ownership, canEdit)
	}
	return h.makeViewPublic(view, ownership, canEdit)
}

func (h *ChangeViewVisibilityHandler) makeViewPrivate(view *aggregates.ArchitectureView, ownership ownershipContext, canEdit bool) error {
	if !canEdit {
		return aggregates.ErrOnlyOwnerCanMakePrivate
	}

	if ownership.hasNoOwner || !ownership.isOwner {
		if err := view.SetOwner(ownership.actorAsOwner()); err != nil {
			return err
		}
	}

	return view.MakePrivate()
}

func (h *ChangeViewVisibilityHandler) makeViewPublic(view *aggregates.ArchitectureView, ownership ownershipContext, canEdit bool) error {
	if !canEdit {
		return ErrNotAuthorizedToChangeVisibility
	}

	newOwner := view.Owner()
	if ownership.hasNoOwner || !ownership.isOwner {
		newOwner = ownership.actorAsOwner()
	}

	return view.MakePublic(newOwner)
}

func (h *ChangeViewVisibilityHandler) isActorAdmin(ctx context.Context, actorID string) (bool, error) {
	return h.userRoleChecker.IsAdmin(ctx, actorID)
}
