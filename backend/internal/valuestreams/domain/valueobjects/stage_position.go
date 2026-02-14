package valueobjects

import (
	"errors"

	domain "easi/backend/internal/shared/eventsourcing"
)

var ErrStagePositionInvalid = errors.New("stage position must be a positive integer")

type StagePosition struct {
	value int
}

func NewStagePosition(value int) (StagePosition, error) {
	if value <= 0 {
		return StagePosition{}, ErrStagePositionInvalid
	}
	return StagePosition{value: value}, nil
}

func MustNewStagePosition(value int) StagePosition {
	pos, err := NewStagePosition(value)
	if err != nil {
		panic(err)
	}
	return pos
}

func (p StagePosition) Value() int {
	return p.value
}

func (p StagePosition) Equals(other domain.ValueObject) bool {
	if otherPos, ok := other.(StagePosition); ok {
		return p.value == otherPos.value
	}
	return false
}
