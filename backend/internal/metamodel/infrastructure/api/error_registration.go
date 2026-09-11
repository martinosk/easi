package api

import (
	"easi/backend/internal/metamodel/application/handlers"
	"easi/backend/internal/metamodel/domain/aggregates"
	"easi/backend/internal/metamodel/domain/valueobjects"
	"easi/backend/internal/metamodel/infrastructure/repositories"
	sharedAPI "easi/backend/internal/shared/api"
)

func init() {
	registry := sharedAPI.GetErrorRegistry()

	registry.RegisterNotFound(repositories.ErrMetaModelConfigurationNotFound, "Meta model configuration not found")

	registry.RegisterConflict(valueobjects.ErrCannotRemoveLastActivePillar, "Cannot remove the last active pillar")
	registry.RegisterConflict(valueobjects.ErrPillarAlreadyInactive, "Pillar is already inactive")

	registry.RegisterValidation(valueobjects.ErrTooManyPillars, "Cannot have more than 20 pillars")
	registry.RegisterValidation(valueobjects.ErrPillarNameDuplicate, "Pillar name already exists")
	registry.RegisterValidation(valueobjects.ErrPillarNotFound, "Pillar not found")

	registerSubjectAttributeErrors(registry)
}

func registerSubjectAttributeErrors(registry *sharedAPI.ErrorRegistry) {
	registry.RegisterNotFound(repositories.ErrSubjectAttributeSchemaNotFound, "Subject attribute schema not found")
	registry.RegisterNotFound(aggregates.ErrAttributeNotFound, "Attribute not found")
	registry.RegisterNotFound(valueobjects.ErrOptionNotFound, "Attribute option not found")

	registry.RegisterConflict(handlers.ErrSchemaAlreadyExists, "An attribute schema already exists for this subject type")
	registry.RegisterConflict(aggregates.ErrDuplicateAttributeName, "An attribute with this name already exists")
	registry.RegisterConflict(aggregates.ErrAttributeTypeImmutable, "Attribute types are immutable; retire this attribute and define a new one")
	registry.RegisterConflict(aggregates.ErrAttributeRetired, "Attribute is retired")
	registry.RegisterConflict(aggregates.ErrAttributeAlreadyRetired, "Attribute is already retired")
	registry.RegisterConflict(aggregates.ErrAttributeAlreadyActive, "Attribute is already active")
	registry.RegisterConflict(aggregates.ErrAttributeAlreadyDefined, "Attribute is already defined")
	registry.RegisterConflict(valueobjects.ErrDuplicateOptionLabel, "Option label already exists on this attribute")
	registry.RegisterConflict(valueobjects.ErrOptionAlreadyRetired, "Option is already retired")
	registry.RegisterConflict(valueobjects.ErrLastActiveOption, "Cannot retire the last active option")
	registry.RegisterConflict(valueobjects.ErrNotSelectionAttribute, "Attribute is not a selection attribute")

	registry.RegisterValidation(valueobjects.ErrInvalidSubjectType, "Invalid subject type")
	registry.RegisterValidation(valueobjects.ErrInvalidAttributeType, "Invalid attribute type")
	registry.RegisterValidation(valueobjects.ErrInvalidAttributeID, "Invalid attribute ID")
	registry.RegisterValidation(valueobjects.ErrInvalidOptionID, "Invalid option ID")
	registry.RegisterValidation(valueobjects.ErrAttributeNameEmpty, "Attribute name cannot be empty")
	registry.RegisterValidation(valueobjects.ErrAttributeNameTooLong, "Attribute name is too long")
	registry.RegisterValidation(valueobjects.ErrHelpTextTooLong, "Help text is too long")
	registry.RegisterValidation(valueobjects.ErrOptionLabelEmpty, "Option label cannot be empty")
	registry.RegisterValidation(valueobjects.ErrOptionLabelTooLong, "Option label is too long")
	registry.RegisterValidation(valueobjects.ErrSelectionOptionRequired, "A selection attribute must define at least one option")
	registry.RegisterValidation(valueobjects.ErrOptionsNotAllowed, "Only selection attributes can define options")
	registry.RegisterValidation(valueobjects.ErrBoundsNotAllowed, "Only number attributes can define bounds")
	registry.RegisterValidation(valueobjects.ErrMinExceedsMax, "Minimum bound must not exceed maximum bound")
}
