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

func (TextValue) isBuiltInFieldValue()     {}
func (DateValue) isBuiltInFieldValue()     {}
func (MaturityValue) isBuiltInFieldValue() {}
func (ExpertsValue) isBuiltInFieldValue()  {}

type SubjectSnapshot struct {
	Name   string
	Fields map[string]BuiltInFieldValue
}

type BuiltInFieldSource interface {
	FetchSubject(ctx context.Context, subjectID string) (*SubjectSnapshot, error)
	CountSubjects(ctx context.Context) (int, error)
}

type MaturitySection struct {
	Name     string
	MinValue int
	MaxValue int
}

type MaturityScaleSource interface {
	Sections(ctx context.Context) ([]MaturitySection, error)
}
