package valueobjects

import (
	"errors"

	"easi/backend/internal/shared/eventsourcing"
)

var (
	ErrVersionInvalid = errors.New("version must be a positive integer (>= 1)")
)

type Version struct {
	value int
}

func NewVersion(value int) (Version, error) {
	if value < 1 {
		return Version{}, ErrVersionInvalid
	}
	return Version{value: value}, nil
}

func InitialVersion() Version {
	return Version{value: 1}
}

func (v Version) Value() int {
	return v.value
}

func (v Version) Increment() Version {
	return Version{value: v.value + 1}
}

func (v Version) Equals(other domain.ValueObject) bool {
	if otherVer, ok := other.(Version); ok {
		return v.value == otherVer.value
	}
	return false
}
