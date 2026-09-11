package handlers

import (
	"easi/backend/internal/onepagers/application/commands"
	"easi/backend/internal/onepagers/domain/aggregates"
	"easi/backend/internal/onepagers/domain/valueobjects"
	"easi/backend/internal/onepagers/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

func NewIncludeCustomFieldHandler(repository *repositories.OnePagerConfigurationRepository) cqrs.CommandHandler {
	return newModifyHandler(repository, fieldAction(
		func(c *commands.IncludeCustomField) string { return c.FieldID },
		(*aggregates.OnePagerConfiguration).IncludeCustomField,
	))
}

func NewExcludeCustomFieldHandler(repository *repositories.OnePagerConfigurationRepository) cqrs.CommandHandler {
	return newModifyHandler(repository, fieldAction(
		func(c *commands.ExcludeCustomField) string { return c.FieldID },
		(*aggregates.OnePagerConfiguration).ExcludeCustomField,
	))
}

func NewChangeCustomFieldRequirementHandler(repository *repositories.OnePagerConfigurationRepository) cqrs.CommandHandler {
	return newModifyHandler(repository, changeCustomFieldRequirement)
}

func NewIncludeBuiltInFieldHandler(repository *repositories.OnePagerConfigurationRepository) cqrs.CommandHandler {
	return newModifyHandler(repository, includeBuiltInField)
}

func NewExcludeBuiltInFieldHandler(repository *repositories.OnePagerConfigurationRepository) cqrs.CommandHandler {
	return newModifyHandler(repository, excludeBuiltInField)
}

func NewChangeBuiltInFieldRequirementHandler(repository *repositories.OnePagerConfigurationRepository) cqrs.CommandHandler {
	return newModifyHandler(repository, changeBuiltInFieldRequirement)
}

func NewReorderOnePagerFieldsHandler(repository *repositories.OnePagerConfigurationRepository) cqrs.CommandHandler {
	return newModifyHandler(repository, reorderOnePagerFields)
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

func changeBuiltInFieldRequirement(config *aggregates.OnePagerConfiguration, c *commands.ChangeBuiltInFieldRequirement, modifiedBy valueobjects.UserEmail) (string, error) {
	return "", config.ChangeBuiltInFieldRequirement(c.EntryID, c.Required, modifiedBy)
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
