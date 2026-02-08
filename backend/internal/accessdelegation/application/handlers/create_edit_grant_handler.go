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

	grant, err := buildEditGrant(command)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, grant); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.NewResult(grant.ID()), nil
}

func buildEditGrant(cmd *commands.CreateEditGrant) (*aggregates.EditGrant, error) {
	artifactType, err := valueobjects.NewArtifactType(cmd.ArtifactType)
	if err != nil {
		return nil, err
	}

	artifactRef, err := valueobjects.NewArtifactRef(artifactType, cmd.ArtifactID)
	if err != nil {
		return nil, err
	}

	scope, err := valueobjects.NewGrantScope(cmd.Scope)
	if err != nil {
		return nil, err
	}

	grantor, err := valueobjects.NewGrantor(cmd.GrantorID, cmd.GrantorEmail)
	if err != nil {
		return nil, err
	}

	granteeEmail, err := valueobjects.NewGranteeEmail(cmd.GranteeEmail)
	if err != nil {
		return nil, err
	}

	reason, err := valueobjects.NewReason(cmd.Reason)
	if err != nil {
		return nil, err
	}

	return aggregates.NewEditGrant(aggregates.GrantRequest{
		Grantor:      grantor,
		GranteeEmail: granteeEmail,
		ArtifactRef:  artifactRef,
		Scope:        scope,
		Reason:       reason,
	})
}
