package handlers

import (
	"easi/backend/internal/onepagers/application/commands"
	"easi/backend/internal/onepagers/domain/aggregates"
	"easi/backend/internal/onepagers/domain/valueobjects"
	"easi/backend/internal/onepagers/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

func NewDefineCustomFieldHandler(repository *repositories.OnePagerConfigurationRepository) cqrs.CommandHandler {
	return newModifyHandler(repository, defineCustomField)
}

func NewRenameCustomFieldHandler(repository *repositories.OnePagerConfigurationRepository) cqrs.CommandHandler {
	return newModifyHandler(repository, renameCustomField)
}

func NewChangeCustomFieldRequirementHandler(repository *repositories.OnePagerConfigurationRepository) cqrs.CommandHandler {
	return newModifyHandler(repository, changeCustomFieldRequirement)
}

func NewRetireCustomFieldHandler(repository *repositories.OnePagerConfigurationRepository) cqrs.CommandHandler {
	return newModifyHandler(repository, fieldAction(
		func(c *commands.RetireCustomField) string { return c.FieldID },
		(*aggregates.OnePagerConfiguration).RetireCustomField,
	))
}

func NewReactivateCustomFieldHandler(repository *repositories.OnePagerConfigurationRepository) cqrs.CommandHandler {
	return newModifyHandler(repository, fieldAction(
		func(c *commands.ReactivateCustomField) string { return c.FieldID },
		(*aggregates.OnePagerConfiguration).ReactivateCustomField,
	))
}

func NewIncludeBuiltInFieldHandler(repository *repositories.OnePagerConfigurationRepository) cqrs.CommandHandler {
	return newModifyHandler(repository, includeBuiltInField)
}

func NewExcludeBuiltInFieldHandler(repository *repositories.OnePagerConfigurationRepository) cqrs.CommandHandler {
	return newModifyHandler(repository, excludeBuiltInField)
}

func NewReorderOnePagerFieldsHandler(repository *repositories.OnePagerConfigurationRepository) cqrs.CommandHandler {
	return newModifyHandler(repository, reorderOnePagerFields)
}

func NewAddSelectionOptionHandler(repository *repositories.OnePagerConfigurationRepository) cqrs.CommandHandler {
	return newModifyHandler(repository, addSelectionOption)
}

func NewRetireSelectionOptionHandler(repository *repositories.OnePagerConfigurationRepository) cqrs.CommandHandler {
	return newModifyHandler(repository, retireSelectionOption)
}

func defineCustomField(config *aggregates.OnePagerConfiguration, c *commands.DefineCustomField, modifiedBy valueobjects.UserEmail) (string, error) {
	params, err := buildDefineParams(c)
	if err != nil {
		return "", err
	}
	fieldID, err := config.DefineCustomField(params, modifiedBy)
	if err != nil {
		return "", err
	}
	return fieldID.Value(), nil
}

func buildDefineParams(c *commands.DefineCustomField) (aggregates.DefineCustomFieldParams, error) {
	name, err := valueobjects.NewFieldName(c.Name)
	if err != nil {
		return aggregates.DefineCustomFieldParams{}, err
	}
	fieldType, err := valueobjects.NewFieldType(c.FieldType)
	if err != nil {
		return aggregates.DefineCustomFieldParams{}, err
	}
	helpText, err := valueobjects.NewHelpText(c.HelpText)
	if err != nil {
		return aggregates.DefineCustomFieldParams{}, err
	}
	labels := make([]valueobjects.OptionLabel, len(c.OptionLabels))
	for i, raw := range c.OptionLabels {
		label, err := valueobjects.NewOptionLabel(raw)
		if err != nil {
			return aggregates.DefineCustomFieldParams{}, err
		}
		labels[i] = label
	}
	return aggregates.DefineCustomFieldParams{
		Name:         name,
		Type:         fieldType,
		Required:     c.Required,
		HelpText:     helpText,
		OptionLabels: labels,
	}, nil
}

func renameCustomField(config *aggregates.OnePagerConfiguration, c *commands.RenameCustomField, modifiedBy valueobjects.UserEmail) (string, error) {
	fieldID, err := valueobjects.NewFieldIDFromString(c.FieldID)
	if err != nil {
		return "", err
	}
	name, err := valueobjects.NewFieldName(c.Name)
	if err != nil {
		return "", err
	}
	helpText, err := valueobjects.NewHelpText(c.HelpText)
	if err != nil {
		return "", err
	}
	return "", config.RenameCustomField(aggregates.RenameCustomFieldParams{
		FieldID:       fieldID,
		Name:          name,
		HelpText:      helpText,
		RequestedType: c.RequestedType,
	}, modifiedBy)
}

func changeCustomFieldRequirement(config *aggregates.OnePagerConfiguration, c *commands.ChangeCustomFieldRequirement, modifiedBy valueobjects.UserEmail) (string, error) {
	fieldID, err := valueobjects.NewFieldIDFromString(c.FieldID)
	if err != nil {
		return "", err
	}
	return "", config.ChangeCustomFieldRequirement(fieldID, c.Required, modifiedBy)
}

func includeBuiltInField(config *aggregates.OnePagerConfiguration, c *commands.IncludeBuiltInField, modifiedBy valueobjects.UserEmail) (string, error) {
	return "", config.IncludeBuiltInField(c.EntryID, modifiedBy)
}

func excludeBuiltInField(config *aggregates.OnePagerConfiguration, c *commands.ExcludeBuiltInField, modifiedBy valueobjects.UserEmail) (string, error) {
	return "", config.ExcludeBuiltInField(c.EntryID, modifiedBy)
}

func reorderOnePagerFields(config *aggregates.OnePagerConfiguration, c *commands.ReorderOnePagerFields, modifiedBy valueobjects.UserEmail) (string, error) {
	order := make([]valueobjects.FieldRef, len(c.Order))
	for i, ref := range c.Order {
		fieldRef, err := valueobjects.NewFieldRef(ref.Kind, ref.ID)
		if err != nil {
			return "", err
		}
		order[i] = fieldRef
	}
	return "", config.ReorderFields(order, modifiedBy)
}

func addSelectionOption(config *aggregates.OnePagerConfiguration, c *commands.AddSelectionOption, modifiedBy valueobjects.UserEmail) (string, error) {
	fieldID, err := valueobjects.NewFieldIDFromString(c.FieldID)
	if err != nil {
		return "", err
	}
	label, err := valueobjects.NewOptionLabel(c.Label)
	if err != nil {
		return "", err
	}
	optionID, err := config.AddSelectionOption(fieldID, label, modifiedBy)
	if err != nil {
		return "", err
	}
	return optionID.Value(), nil
}

func retireSelectionOption(config *aggregates.OnePagerConfiguration, c *commands.RetireSelectionOption, modifiedBy valueobjects.UserEmail) (string, error) {
	fieldID, err := valueobjects.NewFieldIDFromString(c.FieldID)
	if err != nil {
		return "", err
	}
	optionID, err := valueobjects.NewOptionIDFromString(c.OptionID)
	if err != nil {
		return "", err
	}
	return "", config.RetireSelectionOption(fieldID, optionID, modifiedBy)
}
