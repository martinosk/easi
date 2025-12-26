package handlers

import (
	"context"

	"easi/backend/internal/metamodel/application/commands"
	"easi/backend/internal/metamodel/domain/valueobjects"
	"easi/backend/internal/metamodel/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type UpdateMaturityScaleHandler struct {
	repository *repositories.MetaModelConfigurationRepository
}

func NewUpdateMaturityScaleHandler(repository *repositories.MetaModelConfigurationRepository) *UpdateMaturityScaleHandler {
	return &UpdateMaturityScaleHandler{
		repository: repository,
	}
}

func (h *UpdateMaturityScaleHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(*commands.UpdateMaturityScale)
	if !ok {
		return cqrs.ErrInvalidCommand
	}

	config, err := h.repository.GetByID(ctx, command.ID)
	if err != nil {
		return err
	}

	modifiedBy, err := valueobjects.NewUserEmail(command.ModifiedBy)
	if err != nil {
		return err
	}

	scaleConfig, err := buildMaturityScaleConfig(command.Sections)
	if err != nil {
		return err
	}

	if err := config.UpdateMaturityScale(scaleConfig, modifiedBy); err != nil {
		return err
	}

	return h.repository.Save(ctx, config)
}

func buildMaturityScaleConfig(sectionData [4]commands.MaturitySectionInput) (valueobjects.MaturityScaleConfig, error) {
	var sections [4]valueobjects.MaturitySection
	for i, s := range sectionData {
		section, err := buildMaturitySection(s)
		if err != nil {
			return valueobjects.MaturityScaleConfig{}, err
		}
		sections[i] = section
	}
	return valueobjects.NewMaturityScaleConfig(sections)
}

func buildMaturitySection(s commands.MaturitySectionInput) (valueobjects.MaturitySection, error) {
	order, err := valueobjects.NewSectionOrder(s.Order)
	if err != nil {
		return valueobjects.MaturitySection{}, err
	}
	name, err := valueobjects.NewSectionName(s.Name)
	if err != nil {
		return valueobjects.MaturitySection{}, err
	}
	minValue, err := valueobjects.NewMaturityValue(s.MinValue)
	if err != nil {
		return valueobjects.MaturitySection{}, err
	}
	maxValue, err := valueobjects.NewMaturityValue(s.MaxValue)
	if err != nil {
		return valueobjects.MaturitySection{}, err
	}
	return valueobjects.NewMaturitySection(order, name, minValue, maxValue)
}
