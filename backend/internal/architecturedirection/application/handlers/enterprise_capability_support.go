package handlers

import (
	"context"

	"easi/backend/internal/architecturedirection/application/commands"
	"easi/backend/internal/architecturedirection/domain/aggregates"
	"easi/backend/internal/architecturedirection/domain/valueobjects"
)

type capabilityNameUniqueness interface {
	NameExists(ctx context.Context, name, excludeID string) (bool, error)
}

func rejectDuplicateCapabilityName(ctx context.Context, names capabilityNameUniqueness, name, excludeID string) error {
	taken, err := names.NameExists(ctx, name, excludeID)
	if err != nil {
		return err
	}
	if taken {
		return ErrEnterpriseCapabilityNameExists
	}
	return nil
}

type capabilityDetails struct {
	name        valueobjects.EnterpriseCapabilityName
	description valueobjects.Description
	category    valueobjects.Category
}

func newCapabilityDetails(name, description, category string) (capabilityDetails, error) {
	capabilityName, err := valueobjects.NewEnterpriseCapabilityName(name)
	if err != nil {
		return capabilityDetails{}, err
	}

	capabilityDescription, err := valueobjects.NewDescription(description)
	if err != nil {
		return capabilityDetails{}, err
	}

	capabilityCategory, err := valueobjects.NewCategory(category)
	if err != nil {
		return capabilityDetails{}, err
	}

	return capabilityDetails{name: capabilityName, description: capabilityDescription, category: capabilityCategory}, nil
}

func newEnterpriseImportanceParams(command *commands.SetEnterpriseStrategicImportance) (aggregates.NewEnterpriseImportanceParams, error) {
	var params aggregates.NewEnterpriseImportanceParams

	enterpriseCapabilityID, err := valueobjects.NewEnterpriseCapabilityIDFromString(command.EnterpriseCapabilityID)
	if err != nil {
		return params, err
	}

	pillarID, err := valueobjects.NewPillarIDFromString(command.PillarID)
	if err != nil {
		return params, err
	}

	importance, err := valueobjects.NewImportance(command.Importance)
	if err != nil {
		return params, err
	}

	rationale, err := valueobjects.NewRationale(command.Rationale)
	if err != nil {
		return params, err
	}

	return aggregates.NewEnterpriseImportanceParams{
		EnterpriseCapabilityID: enterpriseCapabilityID,
		PillarID:               pillarID,
		PillarName:             command.PillarName,
		Importance:             importance,
		Rationale:              rationale,
	}, nil
}
