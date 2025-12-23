package valueobjects

import (
	"strings"

	"easi/backend/internal/shared/eventsourcing"
)

type Owner struct {
	value string
}

func NewOwner(value string) Owner {
	return Owner{value: strings.TrimSpace(value)}
}

func (o Owner) Value() string {
	return o.value
}

func (o Owner) IsEmpty() bool {
	return o.value == ""
}

func (o Owner) Equals(other domain.ValueObject) bool {
	if otherOwner, ok := other.(Owner); ok {
		return o.value == otherOwner.value
	}
	return false
}

func (o Owner) String() string {
	return o.value
}
