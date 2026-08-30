package handlers

import (
	"context"
	"errors"

	"easi/backend/internal/architectureviews/application/commands"
	"easi/backend/internal/architectureviews/domain/aggregates"
	"easi/backend/internal/architectureviews/domain/valueobjects"
	"easi/backend/internal/architectureviews/infrastructure/repositories"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/cqrs"
)

var ErrNotAuthorizedToChangeVisibility = errors.New("not authorized to change view visibility")

type ChangeViewVisibilityHandler struct {
	viewRepository *repositories.ArchitectureViewRepository
}

func NewChangeViewVisibilityHandler(
	viewRepository *repositories.ArchitectureViewRepository,
) *ChangeViewVisibilityHandler {
	return &ChangeViewVisibilityHandler{
		viewRepository: viewRepository,
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
	canEdit := canActorEditView(ownership)

	if err := applyVisibilityChange(view, command.IsPrivate, ownership, canEdit); err != nil {
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
	actorRole  sharedctx.Role
	hasNoOwner bool
	isOwner    bool
}

func newOwnershipContext(view *aggregates.ArchitectureView, actor sharedctx.Actor) ownershipContext {
	hasNoOwner := view.Owner().IsEmpty()
	return ownershipContext{
		actorID:    actor.ID,
		actorEmail: actor.Email,
		actorRole:  actor.Role,
		hasNoOwner: hasNoOwner,
		isOwner:    !hasNoOwner && view.Owner().UserID() == actor.ID,
	}
}

func (o ownershipContext) actorAsOwner() valueobjects.ViewOwner {
	owner, _ := valueobjects.NewViewOwner(o.actorID, o.actorEmail)
	return owner
}

func canActorEditView(ownership ownershipContext) bool {
	if ownership.hasNoOwner || ownership.isOwner {
		return true
	}
	return ownership.actorRole == sharedctx.RoleAdmin
}

func applyVisibilityChange(view *aggregates.ArchitectureView, makePrivate bool, ownership ownershipContext, canEdit bool) error {
	if makePrivate {
		return makeViewPrivate(view, ownership, canEdit)
	}
	return makeViewPublic(view, ownership, canEdit)
}

func makeViewPrivate(view *aggregates.ArchitectureView, ownership ownershipContext, canEdit bool) error {
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

func makeViewPublic(view *aggregates.ArchitectureView, ownership ownershipContext, canEdit bool) error {
	if !canEdit {
		return ErrNotAuthorizedToChangeVisibility
	}

	newOwner := view.Owner()
	if ownership.hasNoOwner || !ownership.isOwner {
		newOwner = ownership.actorAsOwner()
	}

	return view.MakePublic(newOwner)
}
