package api

import (
	"easi/backend/internal/onepagers/application/handlers"
	"easi/backend/internal/onepagers/domain/aggregates"
	"easi/backend/internal/onepagers/domain/valueobjects"
	"easi/backend/internal/onepagers/infrastructure/repositories"
	sharedAPI "easi/backend/internal/shared/api"
)

func init() {
	registry := sharedAPI.GetErrorRegistry()

	registry.RegisterNotFound(repositories.ErrOnePagerConfigurationNotFound, "One-pager configuration not found")
	registry.RegisterNotFound(aggregates.ErrFieldNotFound, "Custom field not found")
	registry.RegisterNotFound(valueobjects.ErrOptionNotFound, "Selection option not found")

	registry.RegisterConflict(handlers.ErrConfigurationAlreadyExists, "A configuration already exists for this subject type")
	registry.RegisterConflict(aggregates.ErrDuplicateFieldName, "A field with this display name already exists")
	registry.RegisterConflict(aggregates.ErrFieldTypeImmutable, "Field types are immutable; retire this field and define a new one")
	registry.RegisterConflict(aggregates.ErrFieldRetired, "Custom field is retired")
	registry.RegisterConflict(aggregates.ErrFieldAlreadyRetired, "Custom field is already retired")
	registry.RegisterConflict(aggregates.ErrFieldAlreadyActive, "Custom field is already active")
	registry.RegisterConflict(aggregates.ErrBuiltInFieldAlreadyIncluded, "Built-in field is already included")
	registry.RegisterConflict(aggregates.ErrBuiltInFieldNotIncluded, "Built-in field is not included")
	registry.RegisterConflict(valueobjects.ErrDuplicateOptionLabel, "Option label already exists on this field")
	registry.RegisterConflict(valueobjects.ErrOptionAlreadyRetired, "Option is already retired")
	registry.RegisterConflict(valueobjects.ErrLastActiveOption, "Cannot retire the last active option")
	registry.RegisterConflict(valueobjects.ErrNotSelectionField, "Field is not a selection field")

	registry.RegisterValidation(valueobjects.ErrInvalidSubjectType, "Invalid subject type")
	registry.RegisterValidation(valueobjects.ErrInvalidFieldType, "Invalid field type")
	registry.RegisterValidation(valueobjects.ErrInvalidFieldID, "Invalid field ID")
	registry.RegisterValidation(valueobjects.ErrInvalidOptionID, "Invalid option ID")
	registry.RegisterValidation(valueobjects.ErrFieldNameEmpty, "Field name cannot be empty")
	registry.RegisterValidation(valueobjects.ErrFieldNameTooLong, "Field name is too long")
	registry.RegisterValidation(valueobjects.ErrHelpTextTooLong, "Help text is too long")
	registry.RegisterValidation(valueobjects.ErrOptionLabelEmpty, "Option label cannot be empty")
	registry.RegisterValidation(valueobjects.ErrOptionLabelTooLong, "Option label is too long")
	registry.RegisterValidation(valueobjects.ErrSelectionOptionRequired, "A selection field must define at least one option")
	registry.RegisterValidation(valueobjects.ErrOptionsNotAllowed, "Only selection fields can define options")
	registry.RegisterValidation(valueobjects.ErrInvalidFieldRefKind, "Invalid field reference kind")
	registry.RegisterValidation(valueobjects.ErrFieldRefIDEmpty, "Field reference ID cannot be empty")
	registry.RegisterValidation(aggregates.ErrInvalidDisplayOrder, "Display order must contain every included built-in and active custom field exactly once")
	registry.RegisterValidation(aggregates.ErrUnknownBuiltInField, "Built-in field is not part of the catalog for this subject type")
}
