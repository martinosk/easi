package valueobjects

import (
	"easi/backend/internal/shared/domain"
	"errors"

	"github.com/google/uuid"
)

var ErrInvalidImportSessionID = errors.New("invalid import session ID: must be a valid UUID")

type ImportSessionID struct {
	value string
}

func NewImportSessionID() ImportSessionID {
	return ImportSessionID{value: uuid.New().String()}
}

func NewImportSessionIDFromString(value string) (ImportSessionID, error) {
	if value == "" {
		return ImportSessionID{}, ErrInvalidImportSessionID
	}
	if _, err := uuid.Parse(value); err != nil {
		return ImportSessionID{}, ErrInvalidImportSessionID
	}
	return ImportSessionID{value: value}, nil
}

func (id ImportSessionID) Value() string {
	return id.value
}

func (id ImportSessionID) Equals(other domain.ValueObject) bool {
	if otherID, ok := other.(ImportSessionID); ok {
		return id.value == otherID.value
	}
	return false
}

func (id ImportSessionID) String() string {
	return id.value
}
