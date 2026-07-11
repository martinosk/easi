package queries

import (
	"context"
	"errors"
	"fmt"

	"easi/backend/internal/onepagers/application/ports"
	"easi/backend/internal/onepagers/application/readmodels"
	"easi/backend/internal/onepagers/domain/valueobjects"
)

var ErrSubjectNotFound = errors.New("one-pager subject not found")

type ConfigurationSource interface {
	GetBySubjectType(ctx context.Context, subjectType string) (*readmodels.ConfigurationRecord, error)
}

type FactsSource interface {
	GetForSubject(ctx context.Context, subject readmodels.SubjectKey) ([]readmodels.FactRecord, error)
}

type OnePagerQueryDeps struct {
	Configurations ConfigurationSource
	Facts          FactsSource
	Subjects       map[string]ports.BuiltInFieldSource
	MaturityScale  ports.MaturityScaleSource
}

type OnePager struct {
	SubjectType string
	SubjectID   string
	SubjectName string
	Fields      []Field
}

type Field struct {
	BuiltIn *BuiltInField
	Custom  *CustomField
}

type BuiltInField struct {
	ID              string
	Label           string
	Value           ports.BuiltInFieldValue
	MaturitySection string
}

type CustomField struct {
	FieldID       string
	Name          string
	FieldType     string
	HelpText      string
	Value         *valueobjects.ValueEnvelope
	DisplayText   string
	RetiredOption bool
}

type OnePagerQuery struct {
	deps OnePagerQueryDeps
}

func NewOnePagerQuery(deps OnePagerQueryDeps) *OnePagerQuery {
	return &OnePagerQuery{deps: deps}
}

func (q *OnePagerQuery) Get(ctx context.Context, subjectType valueobjects.SubjectType, subjectID string) (*OnePager, error) {
	snapshot, err := q.fetchSubjectSnapshot(ctx, subjectType, subjectID)
	if err != nil {
		return nil, err
	}
	if snapshot == nil {
		return nil, ErrSubjectNotFound
	}

	document, err := q.resolveDocument(ctx, subjectType)
	if err != nil {
		return nil, err
	}

	facts, err := q.deps.Facts.GetForSubject(ctx, readmodels.SubjectKey{SubjectType: subjectType.Value(), SubjectID: subjectID})
	if err != nil {
		return nil, fmt.Errorf("get one-pager facts for %s %s: %w", subjectType.Value(), subjectID, err)
	}

	fields := assembleFields(document, subjectType, snapshot, facts)

	if err := q.applyMaturitySections(ctx, fields); err != nil {
		return nil, err
	}

	return &OnePager{
		SubjectType: subjectType.Value(),
		SubjectID:   subjectID,
		SubjectName: snapshot.Name,
		Fields:      fields,
	}, nil
}

func (q *OnePagerQuery) fetchSubjectSnapshot(ctx context.Context, subjectType valueobjects.SubjectType, subjectID string) (*ports.SubjectSnapshot, error) {
	source, found := q.deps.Subjects[subjectType.Value()]
	if !found {
		return nil, ErrSubjectNotFound
	}
	snapshot, err := source.FetchSubject(ctx, subjectID)
	if err != nil {
		return nil, fmt.Errorf("fetch subject %s %s: %w", subjectType.Value(), subjectID, err)
	}
	return snapshot, nil
}

func (q *OnePagerQuery) resolveDocument(ctx context.Context, subjectType valueobjects.SubjectType) (readmodels.ConfigurationDocument, error) {
	config, err := q.deps.Configurations.GetBySubjectType(ctx, subjectType.Value())
	if err != nil {
		return readmodels.ConfigurationDocument{}, fmt.Errorf("get one-pager configuration for subject type %s: %w", subjectType.Value(), err)
	}
	return documentFor(config, subjectType), nil
}
