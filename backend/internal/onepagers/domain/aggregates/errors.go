package aggregates

import "errors"

var (
	ErrCustomFieldNotIncluded      = errors.New("custom field is not included on this configuration")
	ErrCustomFieldAlreadyIncluded  = errors.New("custom field is already included on this configuration")
	ErrUnknownBuiltInField         = errors.New("built-in field is not part of the catalog for this subject type")
	ErrBuiltInFieldAlreadyIncluded = errors.New("built-in field is already included")
	ErrBuiltInFieldNotIncluded     = errors.New("built-in field is not included")
	ErrInvalidDisplayOrder         = errors.New("display order must contain every included built-in and custom field exactly once")
)
