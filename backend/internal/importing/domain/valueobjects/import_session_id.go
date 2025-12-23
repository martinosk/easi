package valueobjects

import (
	"errors"

	"easi/backend/internal/shared/eventsourcing"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

var ErrInvalidImportSessionID = errors.New("invalid import session ID: must be a valid UUID")

type ImportSessionID struct {
	sharedvo.UUIDValue
}

func NewImportSessionID() ImportSessionID {
	return ImportSessionID{UUIDValue: sharedvo.NewUUIDValue()}
}

func NewImportSessionIDFromString(value string) (ImportSessionID, error) {
	uuidValue, err := sharedvo.NewUUIDValueFromString(value)
	if err != nil {
		return ImportSessionID{}, ErrInvalidImportSessionID
	}
	return ImportSessionID{UUIDValue: uuidValue}, nil
}

func (id ImportSessionID) Equals(other domain.ValueObject) bool {
	if otherID, ok := other.(ImportSessionID); ok {
		return id.UUIDValue.EqualsValue(otherID.UUIDValue)
	}
	return false
}
