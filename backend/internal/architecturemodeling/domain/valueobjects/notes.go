package valueobjects

import (
	"errors"
	"strings"

	domain "easi/backend/internal/shared/eventsourcing"
)

const MaxNotesLength = 500

var ErrNotesTooLong = errors.New("notes exceeds maximum length of 500 characters")

type Notes struct {
	value string
}

func NewNotes(value string) (Notes, error) {
	trimmed := strings.TrimSpace(value)
	if len(trimmed) > MaxNotesLength {
		return Notes{}, ErrNotesTooLong
	}
	return Notes{value: trimmed}, nil
}

func MustNewNotes(value string) Notes {
	notes, err := NewNotes(value)
	if err != nil {
		panic(err)
	}
	return notes
}

func (n Notes) Value() string {
	return n.value
}

func (n Notes) IsEmpty() bool {
	return n.value == ""
}

func (n Notes) Equals(other domain.ValueObject) bool {
	if otherNotes, ok := other.(Notes); ok {
		return n.value == otherNotes.value
	}
	return false
}

func (n Notes) String() string {
	return n.value
}
