package valueobjects

import (
	"errors"

	domain "easi/backend/internal/shared/eventsourcing"
)

var (
	ErrInvalidFieldRefKind = errors.New("field reference kind must be builtIn or custom")
	ErrFieldRefIDEmpty     = errors.New("field reference ID cannot be empty")
)

type FieldRefKind string

const (
	FieldRefKindBuiltIn FieldRefKind = "builtIn"
	FieldRefKindCustom  FieldRefKind = "custom"
)

type FieldRef struct {
	kind FieldRefKind
	id   string
}

func NewFieldRef(kind string, id string) (FieldRef, error) {
	refKind := FieldRefKind(kind)
	if refKind != FieldRefKindBuiltIn && refKind != FieldRefKindCustom {
		return FieldRef{}, ErrInvalidFieldRefKind
	}
	if id == "" {
		return FieldRef{}, ErrFieldRefIDEmpty
	}
	return FieldRef{kind: refKind, id: id}, nil
}

func NewBuiltInFieldRef(entryID string) (FieldRef, error) {
	return NewFieldRef(string(FieldRefKindBuiltIn), entryID)
}

func NewCustomFieldRef(fieldID FieldID) FieldRef {
	return FieldRef{kind: FieldRefKindCustom, id: fieldID.Value()}
}

func (f FieldRef) Kind() FieldRefKind {
	return f.kind
}

func (f FieldRef) RefID() string {
	return f.id
}

func (f FieldRef) Equals(other domain.ValueObject) bool {
	if o, ok := other.(FieldRef); ok {
		return f.kind == o.kind && f.id == o.id
	}
	return false
}
