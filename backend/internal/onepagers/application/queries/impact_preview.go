package queries

import (
	"context"
	"errors"
	"fmt"

	"easi/backend/internal/onepagers/application/ports"
	"easi/backend/internal/onepagers/domain/valueobjects"
)

var ErrFieldNotConfigured = errors.New("one-pager field not configured")

type SubjectsWithValueCounter interface {
	CountSubjectsWithValue(ctx context.Context, subjectType, fieldID string) (int, error)
}

type ImpactPreviewDeps struct {
	Configurations ConfigurationSource
	Facts          SubjectsWithValueCounter
	Subjects       map[string]ports.BuiltInFieldSource
}

type ImpactPreview struct {
	SubjectType          string
	FieldID              string
	AffectedSubjectCount int
}

type ImpactPreviewQuery struct {
	deps ImpactPreviewDeps
}

func NewImpactPreviewQuery(deps ImpactPreviewDeps) *ImpactPreviewQuery {
	return &ImpactPreviewQuery{deps: deps}
}

func (q *ImpactPreviewQuery) Preview(ctx context.Context, subjectType valueobjects.SubjectType, fieldID string) (*ImpactPreview, error) {
	population, err := q.countPopulation(ctx, subjectType)
	if err != nil {
		return nil, err
	}

	if fieldID == "" {
		return &ImpactPreview{SubjectType: subjectType.Value(), FieldID: fieldID, AffectedSubjectCount: population}, nil
	}

	if err := q.ensureFieldConfigured(ctx, subjectType, fieldID); err != nil {
		return nil, err
	}

	withValue, err := q.deps.Facts.CountSubjectsWithValue(ctx, subjectType.Value(), fieldID)
	if err != nil {
		return nil, fmt.Errorf("count subjects with value for field %s: %w", fieldID, err)
	}

	return &ImpactPreview{SubjectType: subjectType.Value(), FieldID: fieldID, AffectedSubjectCount: clampToZero(population - withValue)}, nil
}

func (q *ImpactPreviewQuery) countPopulation(ctx context.Context, subjectType valueobjects.SubjectType) (int, error) {
	source, found := q.deps.Subjects[subjectType.Value()]
	if !found {
		return 0, fmt.Errorf("no subject population source configured for subject type %s", subjectType.Value())
	}
	population, err := source.CountSubjects(ctx)
	if err != nil {
		return 0, fmt.Errorf("count subjects for subject type %s: %w", subjectType.Value(), err)
	}
	return population, nil
}

func (q *ImpactPreviewQuery) ensureFieldConfigured(ctx context.Context, subjectType valueobjects.SubjectType, fieldID string) error {
	config, err := q.deps.Configurations.GetBySubjectType(ctx, subjectType.Value())
	if err != nil {
		return fmt.Errorf("get one-pager configuration for subject type %s: %w", subjectType.Value(), err)
	}
	if config == nil {
		return ErrFieldNotConfigured
	}
	if _, found := config.Document.CustomField(fieldID); !found {
		return ErrFieldNotConfigured
	}
	return nil
}

func clampToZero(value int) int {
	if value < 0 {
		return 0
	}
	return value
}
