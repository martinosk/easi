package handlers

import (
	"context"
	"errors"
	"strings"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/domain/aggregates"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	"easi/backend/internal/shared/cqrs"
)

type UpdateCapabilityMetadataRepository interface {
	GetByID(ctx context.Context, id string) (*aggregates.Capability, error)
	Save(ctx context.Context, capability *aggregates.Capability) error
}

type EAOwnerResolver interface {
	ResolveEAOwner(ctx context.Context, value string) (string, error)
}

type UpdateCapabilityMetadataHandler struct {
	repository      UpdateCapabilityMetadataRepository
	eaOwnerResolver EAOwnerResolver
}

func NewUpdateCapabilityMetadataHandler(repository UpdateCapabilityMetadataRepository, eaOwnerResolver EAOwnerResolver) *UpdateCapabilityMetadataHandler {
	return &UpdateCapabilityMetadataHandler{
		repository:      repository,
		eaOwnerResolver: eaOwnerResolver,
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

func (h *UpdateCapabilityMetadataHandler) resolveEAOwner(ctx context.Context, capability *aggregates.Capability, value string) (valueobjects.EAOwner, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return valueobjects.EAOwner{}, nil
	}
	resolved, err := h.eaOwnerResolver.ResolveEAOwner(ctx, value)
	if err == nil {
		return valueobjects.NewEAOwner(resolved)
	}
	if isUnresolvable(err) && value == capability.EAOwner().Value() {
		return capability.EAOwner(), nil
	}
	return valueobjects.EAOwner{}, err
}

func isUnresolvable(err error) bool {
	return errors.Is(err, valueobjects.ErrEAOwnerNotUser) || errors.Is(err, valueobjects.ErrEAOwnerAmbiguous)
}

func buildMetadata(cmd *commands.UpdateCapabilityMetadata, eaOwner valueobjects.EAOwner) (valueobjects.CapabilityMetadata, error) {
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
		eaOwner,
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

	eaOwner, err := h.resolveEAOwner(ctx, capability, command.EAOwner)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	metadata, err := buildMetadata(command, eaOwner)
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
