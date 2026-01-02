package handlers

import (
	"context"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	"easi/backend/internal/capabilitymapping/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type UpdateCapabilityMetadataHandler struct {
	repository *repositories.CapabilityRepository
}

func NewUpdateCapabilityMetadataHandler(repository *repositories.CapabilityRepository) *UpdateCapabilityMetadataHandler {
	return &UpdateCapabilityMetadataHandler{
		repository: repository,
	}
}

func resolveMaturityLevel(maturityValue int, maturityLevel string) (valueobjects.MaturityLevel, error) {
	if maturityValue > 0 {
		return valueobjects.NewMaturityLevelFromValue(maturityValue)
	}
	if maturityLevel != "" {
		return valueobjects.NewMaturityLevel(maturityLevel)
	}
	return valueobjects.MaturityGenesis, nil
}

func buildMetadata(cmd *commands.UpdateCapabilityMetadata) (valueobjects.CapabilityMetadata, error) {
	maturityLevel, err := resolveMaturityLevel(cmd.MaturityValue, cmd.MaturityLevel)
	if err != nil {
		return valueobjects.CapabilityMetadata{}, err
	}

	ownershipModel, err := valueobjects.NewOwnershipModel(cmd.OwnershipModel)
	if err != nil {
		return valueobjects.CapabilityMetadata{}, err
	}

	status, err := valueobjects.NewCapabilityStatus(cmd.Status)
	if err != nil {
		return valueobjects.CapabilityMetadata{}, err
	}

	return valueobjects.NewCapabilityMetadata(
		maturityLevel,
		ownershipModel,
		valueobjects.NewOwner(cmd.PrimaryOwner),
		valueobjects.NewOwner(cmd.EAOwner),
		status,
	), nil
}

func (h *UpdateCapabilityMetadataHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.UpdateCapabilityMetadata)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	capability, err := h.repository.GetByID(ctx, command.ID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	metadata, err := buildMetadata(command)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := capability.UpdateMetadata(metadata); err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, capability); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.EmptyResult(), nil
}
