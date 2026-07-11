package api

import (
	"easi/backend/internal/onepagers/application/handlers"
	"easi/backend/internal/onepagers/application/queries"
	"easi/backend/internal/onepagers/domain/aggregates"
	"easi/backend/internal/onepagers/domain/valueobjects"
	"easi/backend/internal/onepagers/infrastructure/repositories"
	sharedAPI "easi/backend/internal/shared/api"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

func init() {
	registry := sharedAPI.GetErrorRegistry()

	registry.RegisterNotFound(repositories.ErrOnePagerConfigurationNotFound, "One-pager configuration not found")
	registry.RegisterNotFound(repositories.ErrOnePagerFactsNotFound, "One-pager facts not found")
	registry.RegisterNotFound(aggregates.ErrFieldNotFound, "Custom field not found")
	registry.RegisterNotFound(valueobjects.ErrOptionNotFound, "Selection option not found")
	registry.RegisterNotFound(handlers.ErrFieldNotDefined, "Field is not defined on the subject type's one-pager configuration")
	registry.RegisterNotFound(handlers.ErrSubjectNotFound, "Subject does not exist")
	registry.RegisterNotFound(queries.ErrSubjectNotFound, "Subject does not exist")
	registry.RegisterNotFound(queries.ErrFieldNotConfigured, "Field is not defined on the subject type's one-pager configuration")

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
	registry.RegisterConflict(handlers.ErrFieldDefinitionRetired, "Field is retired on the subject type's one-pager configuration")
	registry.RegisterConflict(handlers.ErrOptionRetired, "Selection option is retired")
	registry.RegisterConflict(aggregates.ErrFactsArchived, "One-pager facts are archived because the subject was deleted")

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

	registerFactsValidationErrors(registry)
}

func registerFactsValidationErrors(registry *sharedAPI.ErrorRegistry) {
	registry.RegisterValidation(handlers.ErrValueTypeMismatch, "Value type does not match the field's type")
	registry.RegisterValidation(handlers.ErrOptionNotDefined, "Selection option is not defined on this field")
	registry.RegisterValidation(aggregates.ErrFieldIDRequired, "A field value requires a field ID")
	registry.RegisterValidation(aggregates.ErrFieldValueRequired, "A field value is required")
	registry.RegisterValidation(valueobjects.ErrUnknownValueType, "Unknown field value type")
	registry.RegisterValidation(valueobjects.ErrUnsupportedValueVersion, "Unsupported field value version")
	registry.RegisterValidation(valueobjects.ErrSubjectIDEmpty, "Subject ID cannot be empty")
	registry.RegisterValidation(valueobjects.ErrTextValueEmpty, "Text value cannot be empty")
	registry.RegisterValidation(valueobjects.ErrTextValueTooLong, "Text value is too long")
	registry.RegisterValidation(valueobjects.ErrNumberValueNotFinite, "Number value must be a finite number")
	registry.RegisterValidation(valueobjects.ErrDateValueInvalid, "Date value must be an ISO date (YYYY-MM-DD)")
	registry.RegisterValidation(valueobjects.ErrLinkLabelEmpty, "Link label cannot be empty")
	registry.RegisterValidation(valueobjects.ErrLinkLabelTooLong, "Link label is too long")
	registry.RegisterValidation(valueobjects.ErrContactNameEmpty, "Contact person name cannot be empty")
	registry.RegisterValidation(valueobjects.ErrContactNameTooLong, "Contact person name is too long")
	registry.RegisterValidation(valueobjects.ErrContactCompanyTooLong, "Contact person company is too long")
	registry.RegisterValidation(valueobjects.ErrUserEmailEmpty, "Email cannot be empty")
	registry.RegisterValidation(valueobjects.ErrUserEmailInvalid, "Email must be a valid email address")
	registry.RegisterValidation(valueobjects.ErrUserEmailTooLong, "Email is too long")
	registry.RegisterValidation(sharedvo.ErrURLEmpty, "URL cannot be empty")
	registry.RegisterValidation(sharedvo.ErrURLInvalid, "URL must be an absolute http or https URL")
	registry.RegisterValidation(sharedvo.ErrURLTooLong, "URL is too long")
}
