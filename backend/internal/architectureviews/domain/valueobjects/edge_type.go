package valueobjects

import (
	"errors"

	"easi/backend/internal/shared/domain"
)

var (
	ErrInvalidEdgeType = errors.New("invalid edge type: must be 'default', 'step', 'smoothstep', or 'straight'")
)

type EdgeType struct {
	value string
}

func NewEdgeType(value string) (EdgeType, error) {
	switch value {
	case "default", "step", "smoothstep", "straight":
		return EdgeType{value: value}, nil
	default:
		return EdgeType{}, ErrInvalidEdgeType
	}
}

func DefaultEdgeType() EdgeType {
	return EdgeType{value: "default"}
}

func (e EdgeType) Value() string {
	return e.value
}

func (e EdgeType) Equals(other domain.ValueObject) bool {
	if otherEdgeType, ok := other.(EdgeType); ok {
		return e.value == otherEdgeType.value
	}
	return false
}

func (e EdgeType) String() string {
	return e.value
}
