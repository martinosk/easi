package valueobjects

import (
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

type Description = sharedvo.Description

var (
	NewDescription        = sharedvo.NewDescription
	MustNewDescription    = sharedvo.MustNewDescription
	ErrDescriptionTooLong = sharedvo.ErrDescriptionTooLong
)
