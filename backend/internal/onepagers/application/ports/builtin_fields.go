package ports

import (
	"context"
	"time"
)

type BuiltInFieldValue interface {
	isBuiltInFieldValue()
}

type TextValue struct {
	Text string
}

type DateValue struct {
	Date time.Time
}

type MaturityValue struct {
	Value int
}

type ExpertsValue struct {
	Experts []Expert
}

type Expert struct {
	Name    string
	Role    string
	Contact string
}

type ReferenceListValue struct {
	References []Reference
}

type Reference struct {
	ID          string
	Label       string
	SubjectType string
}

func (TextValue) isBuiltInFieldValue()          {}
func (DateValue) isBuiltInFieldValue()          {}
func (MaturityValue) isBuiltInFieldValue()      {}
func (ExpertsValue) isBuiltInFieldValue()       {}
func (ReferenceListValue) isBuiltInFieldValue() {}

type SubjectSnapshot struct {
	Name   string
	Fields map[string]BuiltInFieldValue
}

func ValueFilled(value BuiltInFieldValue) bool {
	switch v := value.(type) {
	case nil:
		return false
	case ExpertsValue:
		return len(v.Experts) > 0
	case ReferenceListValue:
		return len(v.References) > 0
	case TextValue, DateValue, MaturityValue:
		return true
	default:
		return value != nil
	}
}

type BuiltInFieldSource interface {
	FetchSubject(ctx context.Context, subjectID string, includedEntryIDs []string) (*SubjectSnapshot, error)
	CountSubjects(ctx context.Context) (int, error)
	FilledBuiltInFields(ctx context.Context, subjectIDs, entryIDs []string) (map[string]map[string]bool, error)
	CountSubjectsWithBuiltInValue(ctx context.Context, entryID string) (int, error)
}

type MaturitySection struct {
	Name     string
	MinValue int
	MaxValue int
}

type MaturityScaleSource interface {
	Sections(ctx context.Context) ([]MaturitySection, error)
}
