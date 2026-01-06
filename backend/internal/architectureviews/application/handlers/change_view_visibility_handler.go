package handlers

import (
	"context"
	"errors"

	"easi/backend/internal/architectureviews/application/commands"
	"easi/backend/internal/architectureviews/domain/aggregates"
	"easi/backend/internal/architectureviews/domain/valueobjects"
	"easi/backend/internal/architectureviews/infrastructure/repositories"
	authReadModels "easi/backend/internal/auth/application/readmodels"
	"easi/backend/internal/shared/cqrs"
	sharedctx "easi/backend/internal/shared/context"
)

var (
	ErrNotAuthorizedToChangeVisibility = errors.New("not authorized to change view visibility")
)

type ChangeViewVisibilityHandler struct {
	viewRepository *repositories.ArchitectureViewRepository
	userReadModel  *authReadModels.UserReadModel
}

func NewChangeViewVisibilityHandler(
	viewRepository *repositories.ArchitectureViewRepository,
	userReadModel *authReadModels.UserReadModel,
) *ChangeViewVisibilityHandler {
	return &ChangeViewVisibilityHandler{
		viewRepository: viewRepository,
		userReadModel:  userReadModel,
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

	isAdmin, err := h.isActorAdmin(ctx, actor.ID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	hasNoOwner := view.Owner().IsEmpty()
	isOwner := !hasNoOwner && view.Owner().UserID() == actor.ID
	canEdit := hasNoOwner || isOwner || isAdmin

	if command.IsPrivate {
		if !canEdit {
			return cqrs.EmptyResult(), aggregates.ErrOnlyOwnerCanMakePrivate
		}
		if hasNoOwner || !isOwner {
			actorOwner, _ := valueobjects.NewViewOwner(actor.ID, actor.Email)
			if err := view.SetOwner(actorOwner); err != nil {
				return cqrs.EmptyResult(), err
			}
		}
		if err := view.MakePrivate(); err != nil {
			return cqrs.EmptyResult(), err
		}
	} else {
		if !canEdit {
			return cqrs.EmptyResult(), ErrNotAuthorizedToChangeVisibility
		}

		newOwner := view.Owner()
		if hasNoOwner || (isAdmin && !isOwner) {
			newOwner, _ = valueobjects.NewViewOwner(actor.ID, actor.Email)
		}

		if err := view.MakePublic(newOwner); err != nil {
			return cqrs.EmptyResult(), err
		}
	}

	if err := h.viewRepository.Save(ctx, view); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.EmptyResult(), nil
}

func (h *ChangeViewVisibilityHandler) isActorAdmin(ctx context.Context, actorID string) (bool, error) {
	user, err := h.userReadModel.GetByIDString(ctx, actorID)
	if err != nil {
		return false, err
	}
	if user == nil {
		return false, nil
	}
	return user.Role == "admin", nil
}
