package handlers

import (
	"context"
	"fmt"

	"easi/backend/internal/architecturemodeling/application/commands"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	"easi/backend/internal/architecturemodeling/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type SetOriginLinkHandler struct {
	repository *repositories.ComponentOriginLinkRepository
}

func NewSetOriginLinkHandler(repository *repositories.ComponentOriginLinkRepository) *SetOriginLinkHandler {
	return &SetOriginLinkHandler{repository: repository}
}

func (h *SetOriginLinkHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.SetOriginLink)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	params, err := h.validateCommand(command)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	link, err := getOrCreateComponentOriginLink(ctx, h.repository, params.componentID, params.originType)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := link.Set(params.entityID, params.notes); err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, link); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.NewResult(params.componentID.String()), nil
}

type setOriginLinkParams struct {
	componentID valueobjects.ComponentID
	originType  valueobjects.OriginType
	entityID    string
	notes       valueobjects.Notes
}

func (h *SetOriginLinkHandler) validateCommand(command *commands.SetOriginLink) (setOriginLinkParams, error) {
	componentID, err := valueobjects.NewComponentIDFromString(command.ComponentID)
	if err != nil {
		return setOriginLinkParams{}, err
	}

	originType, err := valueobjects.NewOriginType(command.OriginType)
	if err != nil {
		return setOriginLinkParams{}, err
	}

	entityID, err := validateEntityID(command.OriginType, command.EntityID)
	if err != nil {
		return setOriginLinkParams{}, err
	}

	notes, err := valueobjects.NewNotes(command.Notes)
	if err != nil {
		return setOriginLinkParams{}, err
	}

	return setOriginLinkParams{componentID, originType, entityID, notes}, nil
}

func validateEntityID(originType, entityID string) (string, error) {
	switch originType {
	case valueobjects.OriginTypeAcquiredVia:
		id, err := valueobjects.NewAcquiredEntityIDFromString(entityID)
		if err != nil {
			return "", err
		}
		return id.String(), nil
	case valueobjects.OriginTypePurchasedFrom:
		id, err := valueobjects.NewVendorIDFromString(entityID)
		if err != nil {
			return "", err
		}
		return id.String(), nil
	case valueobjects.OriginTypeBuiltBy:
		id, err := valueobjects.NewInternalTeamIDFromString(entityID)
		if err != nil {
			return "", err
		}
		return id.String(), nil
	default:
		return "", fmt.Errorf("unknown origin type: %s", originType)
	}
}
