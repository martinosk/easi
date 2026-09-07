package valueobjects

import (
	"errors"
	"fmt"

	domain "easi/backend/internal/shared/eventsourcing"
)

var (
	ErrInvalidOwnerKind = errors.New("invalid owner kind: must be user or team")
	ErrEmptyOwnerID     = errors.New("owner id must not be empty")
)

const (
	OwnerKindUser = "user"
	OwnerKindTeam = "team"
)

type OwnerReference struct {
	kind string
	id   string
}

func NewOwnerReference(kind, id string) (OwnerReference, error) {
	if kind != OwnerKindUser && kind != OwnerKindTeam {
		return OwnerReference{}, fmt.Errorf("%w: %s", ErrInvalidOwnerKind, kind)
	}
	if id == "" {
		return OwnerReference{}, ErrEmptyOwnerID
	}
	return OwnerReference{kind: kind, id: id}, nil
}

func (r OwnerReference) Kind() string {
	return r.kind
}

func (r OwnerReference) ID() string {
	return r.id
}

func (r OwnerReference) IsUser() bool {
	return r.kind == OwnerKindUser
}

func (r OwnerReference) IsTeam() bool {
	return r.kind == OwnerKindTeam
}

func (r OwnerReference) ResolvedOwnershipState() OwnershipState {
	if r.IsTeam() {
		return OwnershipState{value: OwnershipStateManaged}
	}
	return OwnershipState{value: OwnershipStateOwned}
}

func (r OwnerReference) Equals(other domain.ValueObject) bool {
	if otherRef, ok := other.(OwnerReference); ok {
		return r.kind == otherRef.kind && r.id == otherRef.id
	}
	return false
}
