package handlers

import (
	"context"

	"easi/backend/internal/accessdelegation/application/commands"
	"easi/backend/internal/accessdelegation/domain/aggregates"
	"easi/backend/internal/accessdelegation/domain/valueobjects"
	"easi/backend/internal/accessdelegation/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type CreateEditGrantHandler struct {
	repository *repositories.EditGrantRepository
}

func NewCreateEditGrantHandler(repository *repositories.EditGrantRepository) *CreateEditGrantHandler {
	return &CreateEditGrantHandler{repository: repository}
}

func (h *CreateEditGrantHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.CreateEditGrant)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	artifactType, err := valueobjects.NewArtifactType(command.ArtifactType)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	artifactRef, err := valueobjects.NewArtifactRef(artifactType, command.ArtifactID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	scope, err := valueobjects.NewGrantScope(command.Scope)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	grantor, err := valueobjects.NewGrantor(command.GrantorID, command.GrantorEmail)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	granteeEmail, err := valueobjects.NewGranteeEmail(command.GranteeEmail)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	reason, err := valueobjects.NewReason(command.Reason)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	grant, err := aggregates.NewEditGrant(grantor, granteeEmail, artifactRef, scope, reason)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, grant); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.NewResult(grant.ID()), nil
}
