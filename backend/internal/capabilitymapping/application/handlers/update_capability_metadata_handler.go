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

func (h *UpdateCapabilityMetadataHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(*commands.UpdateCapabilityMetadata)
	if !ok {
		return cqrs.ErrInvalidCommand
	}

	capability, err := h.repository.GetByID(ctx, command.ID)
	if err != nil {
		return err
	}

	strategyPillar, err := valueobjects.NewStrategyPillar(command.StrategyPillar)
	if err != nil {
		return err
	}

	pillarWeight, err := valueobjects.NewPillarWeight(command.PillarWeight)
	if err != nil {
		return err
	}

	maturityLevel, err := valueobjects.NewMaturityLevel(command.MaturityLevel)
	if err != nil {
		return err
	}

	ownershipModel, err := valueobjects.NewOwnershipModel(command.OwnershipModel)
	if err != nil {
		return err
	}

	primaryOwner := valueobjects.NewOwner(command.PrimaryOwner)
	eaOwner := valueobjects.NewOwner(command.EAOwner)

	status, err := valueobjects.NewCapabilityStatus(command.Status)
	if err != nil {
		return err
	}

	metadata := valueobjects.NewCapabilityMetadata(
		strategyPillar,
		pillarWeight,
		maturityLevel,
		ownershipModel,
		primaryOwner,
		eaOwner,
		status,
	)

	if err := capability.UpdateMetadata(metadata); err != nil {
		return err
	}

	return h.repository.Save(ctx, capability)
}
