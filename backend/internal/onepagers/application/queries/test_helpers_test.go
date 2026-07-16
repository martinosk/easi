package queries_test

import (
	"context"
	"encoding/json"
	"testing"

	"easi/backend/internal/onepagers/application/ports"
	"easi/backend/internal/onepagers/application/queries"
	"easi/backend/internal/onepagers/application/readmodels"
	"easi/backend/internal/onepagers/domain/valueobjects"

	"github.com/stretchr/testify/require"
)

type countingConfigSource struct {
	calls          int
	gotSubjectType string
	record         *readmodels.ConfigurationRecord
	err            error
}

func (f *countingConfigSource) GetBySubjectType(_ context.Context, subjectType string) (*readmodels.ConfigurationRecord, error) {
	f.calls++
	f.gotSubjectType = subjectType
	return f.record, f.err
}

type countingFactsSource struct {
	calls      int
	gotSubject readmodels.SubjectKey
	records    []readmodels.FactRecord
	err        error
}

func (f *countingFactsSource) GetForSubject(_ context.Context, subject readmodels.SubjectKey) ([]readmodels.FactRecord, error) {
	f.calls++
	f.gotSubject = subject
	return f.records, f.err
}

type countingSubjectSource struct {
	calls          int
	snapshot       *ports.SubjectSnapshot
	gotIncludedIDs []string
	err            error
	count          int
	filled         map[string]map[string]bool
	filledErr      error
	filledCalls    int
	gotFilledIDs   []string
	gotEntryIDs    []string
	withValueCount int
	withValueErr   error
	withValueCalls int
	gotEntryID     string
}

func (f *countingSubjectSource) FetchSubject(_ context.Context, _ string, includedEntryIDs []string) (*ports.SubjectSnapshot, error) {
	f.calls++
	f.gotIncludedIDs = includedEntryIDs
	return f.snapshot, f.err
}

func (f *countingSubjectSource) CountSubjects(_ context.Context) (int, error) {
	return f.count, nil
}

func (f *countingSubjectSource) FilledBuiltInFields(_ context.Context, subjectIDs, entryIDs []string) (map[string]map[string]bool, error) {
	f.filledCalls++
	f.gotFilledIDs = subjectIDs
	f.gotEntryIDs = entryIDs
	return f.filled, f.filledErr
}

func (f *countingSubjectSource) CountSubjectsWithBuiltInValue(_ context.Context, entryID string) (int, error) {
	f.withValueCalls++
	f.gotEntryID = entryID
	return f.withValueCount, f.withValueErr
}

type countingMaturitySource struct {
	calls    int
	sections []ports.MaturitySection
	err      error
}

func (f *countingMaturitySource) Sections(_ context.Context) ([]ports.MaturitySection, error) {
	f.calls++
	return f.sections, f.err
}

func mustSubjectType(t *testing.T, value string) valueobjects.SubjectType {
	t.Helper()
	subjectType, err := valueobjects.NewSubjectType(value)
	require.NoError(t, err)
	return subjectType
}

func snapshotNamed(name string, fields map[string]ports.BuiltInFieldValue) *ports.SubjectSnapshot {
	if fields == nil {
		fields = map[string]ports.BuiltInFieldValue{}
	}
	return &ports.SubjectSnapshot{Name: name, Fields: fields}
}

func envelopeOf(t *testing.T, valueType, rawValue string) valueobjects.ValueEnvelope {
	t.Helper()
	return valueobjects.ValueEnvelope{Type: valueType, Version: valueobjects.ValueEnvelopeVersion, Value: json.RawMessage(rawValue)}
}

type depsParams struct {
	subjectType string
	subjects    *countingSubjectSource
	configs     *countingConfigSource
	facts       *countingFactsSource
	maturity    *countingMaturitySource
}

func buildDeps(p depsParams) queries.OnePagerQueryDeps {
	return queries.OnePagerQueryDeps{
		Configurations: p.configs,
		Facts:          p.facts,
		Subjects:       map[string]ports.BuiltInFieldSource{p.subjectType: p.subjects},
		MaturityScale:  p.maturity,
	}
}
