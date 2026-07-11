package valueobjects

import (
	"errors"
	"strings"

	domain "easi/backend/internal/shared/eventsourcing"
)

var ErrSubjectIDEmpty = errors.New("subject ID cannot be empty")

type SubjectRef struct {
	subjectType SubjectType
	subjectID   string
}

func NewSubjectRef(subjectType, subjectID string) (SubjectRef, error) {
	parsedType, err := NewSubjectType(subjectType)
	if err != nil {
		return SubjectRef{}, err
	}
	trimmedID := strings.TrimSpace(subjectID)
	if trimmedID == "" {
		return SubjectRef{}, ErrSubjectIDEmpty
	}
	return SubjectRef{subjectType: parsedType, subjectID: trimmedID}, nil
}

func (r SubjectRef) SubjectType() SubjectType {
	return r.subjectType
}

func (r SubjectRef) SubjectID() string {
	return r.subjectID
}

func (r SubjectRef) Equals(other domain.ValueObject) bool {
	if o, ok := other.(SubjectRef); ok {
		return r.subjectType == o.subjectType && r.subjectID == o.subjectID
	}
	return false
}
