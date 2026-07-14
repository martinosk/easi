package queries_test

import (
	"context"
	"testing"

	"easi/backend/internal/onepagers/application/ports"
	"easi/backend/internal/onepagers/application/queries"
	"easi/backend/internal/onepagers/application/readmodels"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type countingValueCounter struct {
	calls          int
	gotSubjectType string
	gotFieldID     string
	count          int
	err            error
}

func (f *countingValueCounter) CountSubjectsWithValue(_ context.Context, subjectType, fieldID string) (int, error) {
	f.calls++
	f.gotSubjectType = subjectType
	f.gotFieldID = fieldID
	return f.count, f.err
}

func buildImpactPreviewDeps(subjectType string, subjects *countingSubjectSource, configs *countingConfigSource, facts *countingValueCounter) queries.ImpactPreviewDeps {
	return queries.ImpactPreviewDeps{
		Configurations: configs,
		Facts:          facts,
		Subjects:       map[string]ports.BuiltInFieldSource{subjectType: subjects},
	}
}

func customField() queries.PreviewField {
	return queries.PreviewField{Kind: "custom", ID: "contract-link"}
}

func TestPreview_ExistingFieldAffectedCountIsPopulationMinusFilled(t *testing.T) {
	subjects := &countingSubjectSource{count: 100}
	configs := &countingConfigSource{record: singleCustomFieldConfig(readmodels.CustomFieldRecord{ID: "contract-link", Name: "Contract link", Active: true})}
	facts := &countingValueCounter{count: 63}

	query := queries.NewImpactPreviewQuery(buildImpactPreviewDeps("application", subjects, configs, facts))

	result, err := query.Preview(context.Background(), mustSubjectType(t, "application"), customField())

	require.NoError(t, err)
	assert.Equal(t, 37, result.AffectedSubjectCount)
	assert.Equal(t, "application", result.SubjectType)
	assert.Equal(t, "contract-link", result.FieldID)
	assert.Equal(t, "application", facts.gotSubjectType)
	assert.Equal(t, "contract-link", facts.gotFieldID)
}

func TestPreview_NewFieldAffectedCountIsFullPopulation(t *testing.T) {
	subjects := &countingSubjectSource{count: 120}
	configs := &countingConfigSource{}
	facts := &countingValueCounter{}

	query := queries.NewImpactPreviewQuery(buildImpactPreviewDeps("vendor", subjects, configs, facts))

	result, err := query.Preview(context.Background(), mustSubjectType(t, "vendor"), queries.PreviewField{})

	require.NoError(t, err)
	assert.Equal(t, 120, result.AffectedSubjectCount)
	assert.Equal(t, "", result.FieldID)
	assert.Equal(t, 0, configs.calls, "new field preview must not require a configuration lookup")
	assert.Equal(t, 0, facts.calls, "new field preview must not query facts")
}

func TestPreview_AffectedCountClampsToZero(t *testing.T) {
	subjects := &countingSubjectSource{count: 5}
	configs := &countingConfigSource{record: singleCustomFieldConfig(readmodels.CustomFieldRecord{ID: "contract-link", Active: true})}
	facts := &countingValueCounter{count: 9}

	query := queries.NewImpactPreviewQuery(buildImpactPreviewDeps("application", subjects, configs, facts))

	result, err := query.Preview(context.Background(), mustSubjectType(t, "application"), customField())

	require.NoError(t, err)
	assert.Equal(t, 0, result.AffectedSubjectCount)
}

func TestPreview_UnknownFieldIDReturnsFieldNotConfigured(t *testing.T) {
	subjects := &countingSubjectSource{count: 100}
	configs := &countingConfigSource{record: singleCustomFieldConfig(readmodels.CustomFieldRecord{ID: "other-field", Active: true})}
	facts := &countingValueCounter{}

	query := queries.NewImpactPreviewQuery(buildImpactPreviewDeps("application", subjects, configs, facts))

	_, err := query.Preview(context.Background(), mustSubjectType(t, "application"), customField())

	assert.ErrorIs(t, err, queries.ErrFieldNotConfigured)
	assert.Equal(t, 0, facts.calls)
}

func TestPreview_NilConfigurationWithFieldIDReturnsFieldNotConfigured(t *testing.T) {
	subjects := &countingSubjectSource{count: 100}
	configs := &countingConfigSource{record: nil}
	facts := &countingValueCounter{}

	query := queries.NewImpactPreviewQuery(buildImpactPreviewDeps("application", subjects, configs, facts))

	_, err := query.Preview(context.Background(), mustSubjectType(t, "application"), customField())

	assert.ErrorIs(t, err, queries.ErrFieldNotConfigured)
	assert.Equal(t, 0, facts.calls)
}

func TestPreview_RetiredFieldIsStillConfigured(t *testing.T) {
	subjects := &countingSubjectSource{count: 10}
	configs := &countingConfigSource{record: singleCustomFieldConfig(readmodels.CustomFieldRecord{ID: "contract-link", Active: false})}
	facts := &countingValueCounter{count: 4}

	query := queries.NewImpactPreviewQuery(buildImpactPreviewDeps("application", subjects, configs, facts))

	result, err := query.Preview(context.Background(), mustSubjectType(t, "application"), customField())

	require.NoError(t, err)
	assert.Equal(t, 6, result.AffectedSubjectCount)
}

func TestPreview_MissingSubjectSourceReturnsError(t *testing.T) {
	configs := &countingConfigSource{}
	facts := &countingValueCounter{}

	query := queries.NewImpactPreviewQuery(queries.ImpactPreviewDeps{
		Configurations: configs,
		Facts:          facts,
		Subjects:       map[string]ports.BuiltInFieldSource{},
	})

	_, err := query.Preview(context.Background(), mustSubjectType(t, "vendor"), queries.PreviewField{})

	require.Error(t, err)
	assert.NotErrorIs(t, err, queries.ErrFieldNotConfigured)
}

func builtInField(id string) queries.PreviewField {
	return queries.PreviewField{Kind: "builtIn", ID: id}
}

func TestPreview_BuiltInFieldCountIsPopulationMinusWithValue(t *testing.T) {
	subjects := &countingSubjectSource{count: 120, withValueCount: 80}
	configs := &countingConfigSource{}
	facts := &countingValueCounter{}

	query := queries.NewImpactPreviewQuery(buildImpactPreviewDeps("application", subjects, configs, facts))

	result, err := query.Preview(context.Background(), mustSubjectType(t, "application"), builtInField("experts"))

	require.NoError(t, err)
	assert.Equal(t, 40, result.AffectedSubjectCount)
	assert.Equal(t, "experts", result.FieldID)
	assert.Equal(t, "experts", subjects.gotEntryID)
	assert.Equal(t, 1, subjects.withValueCalls)
	assert.Equal(t, 0, facts.calls, "built-in preview must not query facts")
	assert.Equal(t, 0, configs.calls, "built-in preview routes through the catalog, not a facts configuration lookup")
}

func TestPreview_BuiltInFieldCountClampsToZero(t *testing.T) {
	subjects := &countingSubjectSource{count: 5, withValueCount: 9}
	query := queries.NewImpactPreviewQuery(buildImpactPreviewDeps("application", subjects, &countingConfigSource{}, &countingValueCounter{}))

	result, err := query.Preview(context.Background(), mustSubjectType(t, "application"), builtInField("experts"))

	require.NoError(t, err)
	assert.Equal(t, 0, result.AffectedSubjectCount)
}

func TestPreview_UnknownBuiltInEntryReturnsFieldNotConfigured(t *testing.T) {
	subjects := &countingSubjectSource{count: 100, withValueCount: 10}
	query := queries.NewImpactPreviewQuery(buildImpactPreviewDeps("application", subjects, &countingConfigSource{}, &countingValueCounter{}))

	_, err := query.Preview(context.Background(), mustSubjectType(t, "application"), builtInField("maturity"))

	assert.ErrorIs(t, err, queries.ErrFieldNotConfigured)
	assert.Equal(t, 0, subjects.withValueCalls)
}
