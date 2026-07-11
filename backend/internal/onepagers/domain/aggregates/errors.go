package aggregates

import "errors"

var (
	ErrFieldNotFound               = errors.New("custom field not found on this configuration")
	ErrFieldRetired                = errors.New("custom field is retired")
	ErrFieldAlreadyRetired         = errors.New("custom field is already retired")
	ErrFieldAlreadyActive          = errors.New("custom field is already active")
	ErrDuplicateFieldName          = errors.New("a field with this display name already exists on this configuration")
	ErrFieldTypeImmutable          = errors.New("field types are immutable; retire this field and define a new one")
	ErrUnknownBuiltInField         = errors.New("built-in field is not part of the catalog for this subject type")
	ErrBuiltInFieldAlreadyIncluded = errors.New("built-in field is already included")
	ErrBuiltInFieldNotIncluded     = errors.New("built-in field is not included")
	ErrInvalidDisplayOrder         = errors.New("display order must contain every included built-in and active custom field exactly once")
)
