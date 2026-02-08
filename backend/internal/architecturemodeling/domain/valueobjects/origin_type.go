package valueobjects

import (
	"errors"
	"fmt"

	domain "easi/backend/internal/shared/eventsourcing"
)

var ErrInvalidOriginType = errors.New("invalid origin type: must be acquired-via, purchased-from, or built-by")

const (
	OriginTypeAcquiredVia   = "acquired-via"
	OriginTypePurchasedFrom = "purchased-from"
	OriginTypeBuiltBy       = "built-by"
)

type OriginType struct {
	value string
}

func NewOriginType(value string) (OriginType, error) {
	switch value {
	case OriginTypeAcquiredVia, OriginTypePurchasedFrom, OriginTypeBuiltBy:
		return OriginType{value: value}, nil
	default:
		return OriginType{}, fmt.Errorf("%w: %s", ErrInvalidOriginType, value)
	}
}

func (o OriginType) String() string {
	return o.value
}

func (o OriginType) Equals(other domain.ValueObject) bool {
	if otherType, ok := other.(OriginType); ok {
		return o.value == otherType.value
	}
	return false
}
