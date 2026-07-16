package queries

import (
	"context"
	"errors"
	"fmt"

	"easi/backend/internal/onepagers/application/ports"
	"easi/backend/internal/onepagers/domain/catalog"
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

type PreviewField struct {
	Kind string
	ID   string
}

func (f PreviewField) isBuiltIn() bool {
	return f.Kind == string(valueobjects.FieldRefKindBuiltIn)
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

type previewScope struct {
	subjectType valueobjects.SubjectType
	source      ports.BuiltInFieldSource
	population  int
}

func (q *ImpactPreviewQuery) Preview(ctx context.Context, subjectType valueobjects.SubjectType, field PreviewField) (*ImpactPreview, error) {
	source, err := q.subjectSource(subjectType)
	if err != nil {
		return nil, err
	}
	population, err := q.countPopulation(ctx, subjectType, source)
	if err != nil {
		return nil, err
	}
	scope := previewScope{subjectType: subjectType, source: source, population: population}

	if field.isBuiltIn() {
		return q.builtInPreview(ctx, scope, field.ID)
	}
	return q.customPreview(ctx, scope, field.ID)
}

func (q *ImpactPreviewQuery) builtInPreview(ctx context.Context, scope previewScope, entryID string) (*ImpactPreview, error) {
	if _, found := catalog.LookupEntry(scope.subjectType, entryID); !found {
		return nil, ErrFieldNotConfigured
	}
	withValue, err := scope.source.CountSubjectsWithBuiltInValue(ctx, entryID)
	if err != nil {
		return nil, fmt.Errorf("count subjects with value for built-in field %s: %w", entryID, err)
	}
	return q.impactPreview(scope.subjectType, entryID, scope.population, withValue), nil
}

func (q *ImpactPreviewQuery) customPreview(ctx context.Context, scope previewScope, fieldID string) (*ImpactPreview, error) {
	if fieldID == "" {
		return q.impactPreview(scope.subjectType, fieldID, scope.population, 0), nil
	}
	if err := q.ensureFieldConfigured(ctx, scope.subjectType, fieldID); err != nil {
		return nil, err
	}
	withValue, err := q.deps.Facts.CountSubjectsWithValue(ctx, scope.subjectType.Value(), fieldID)
	if err != nil {
		return nil, fmt.Errorf("count subjects with value for field %s: %w", fieldID, err)
	}
	return q.impactPreview(scope.subjectType, fieldID, scope.population, withValue), nil
}

func (q *ImpactPreviewQuery) impactPreview(subjectType valueobjects.SubjectType, fieldID string, population, withValue int) *ImpactPreview {
	return &ImpactPreview{
		SubjectType:          subjectType.Value(),
		FieldID:              fieldID,
		AffectedSubjectCount: clampToZero(population - withValue),
	}
}

func (q *ImpactPreviewQuery) subjectSource(subjectType valueobjects.SubjectType) (ports.BuiltInFieldSource, error) {
	source, found := q.deps.Subjects[subjectType.Value()]
	if !found {
		return nil, fmt.Errorf("no subject population source configured for subject type %s", subjectType.Value())
	}
	return source, nil
}

func (q *ImpactPreviewQuery) countPopulation(ctx context.Context, subjectType valueobjects.SubjectType, source ports.BuiltInFieldSource) (int, error) {
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
