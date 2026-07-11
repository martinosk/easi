package queries_test

import (
	"context"
	"errors"
	"testing"

	"easi/backend/internal/onepagers/application/ports"
	"easi/backend/internal/onepagers/application/queries"
	"easi/backend/internal/onepagers/application/readmodels"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func maturityConfig() *readmodels.ConfigurationRecord {
	return &readmodels.ConfigurationRecord{
		Document: readmodels.ConfigurationDocument{
			DisplayOrder: []readmodels.FieldRefRecord{{Kind: "builtIn", ID: "maturity"}},
		},
	}
}

func TestGet_ResolvesMaturitySectionForMaturityValue(t *testing.T) {
	subjects := &countingSubjectSource{snapshot: snapshotNamed("Cap", map[string]ports.BuiltInFieldValue{
		"maturity": ports.MaturityValue{Value: 85},
	})}
	maturity := &countingMaturitySource{sections: []ports.MaturitySection{
		{Name: "Managed", MinValue: 50, MaxValue: 79},
		{Name: "Optimizing", MinValue: 80, MaxValue: 100},
	}}

	query := queries.NewOnePagerQuery(buildDeps(depsParams{subjectType: "capability", subjects: subjects, configs: &countingConfigSource{record: maturityConfig()}, facts: &countingFactsSource{}, maturity: maturity}))

	result, err := query.Get(context.Background(), mustSubjectType(t, "capability"), "cap-1")

	require.NoError(t, err)
	assert.Equal(t, 1, maturity.calls)
	require.NotNil(t, result.Fields[0].BuiltIn)
	assert.Equal(t, "Optimizing", result.Fields[0].BuiltIn.MaturitySection)
}

func TestGet_DoesNotCallMaturityScaleWhenNoMaturityValuePresent(t *testing.T) {
	config := &readmodels.ConfigurationRecord{
		Document: readmodels.ConfigurationDocument{
			DisplayOrder: []readmodels.FieldRefRecord{{Kind: "builtIn", ID: "description"}},
		},
	}
	subjects := &countingSubjectSource{snapshot: snapshotNamed("Cap", map[string]ports.BuiltInFieldValue{
		"description": ports.TextValue{Text: "text"},
	})}
	maturity := &countingMaturitySource{}

	query := queries.NewOnePagerQuery(buildDeps(depsParams{subjectType: "capability", subjects: subjects, configs: &countingConfigSource{record: config}, facts: &countingFactsSource{}, maturity: maturity}))

	_, err := query.Get(context.Background(), mustSubjectType(t, "capability"), "cap-1")

	require.NoError(t, err)
	assert.Equal(t, 0, maturity.calls)
}

func TestGet_MaturitySectionEmptyWhenNoSectionsConfigured(t *testing.T) {
	subjects := &countingSubjectSource{snapshot: snapshotNamed("Cap", map[string]ports.BuiltInFieldValue{
		"maturity": ports.MaturityValue{Value: 85},
	})}
	maturity := &countingMaturitySource{sections: nil}

	query := queries.NewOnePagerQuery(buildDeps(depsParams{subjectType: "capability", subjects: subjects, configs: &countingConfigSource{record: maturityConfig()}, facts: &countingFactsSource{}, maturity: maturity}))

	result, err := query.Get(context.Background(), mustSubjectType(t, "capability"), "cap-1")

	require.NoError(t, err)
	assert.Equal(t, "", result.Fields[0].BuiltIn.MaturitySection)
}

func TestGet_MaturitySectionEmptyWhenValueMatchesNoSection(t *testing.T) {
	subjects := &countingSubjectSource{snapshot: snapshotNamed("Cap", map[string]ports.BuiltInFieldValue{
		"maturity": ports.MaturityValue{Value: 5},
	})}
	maturity := &countingMaturitySource{sections: []ports.MaturitySection{
		{Name: "Optimizing", MinValue: 80, MaxValue: 100},
	}}

	query := queries.NewOnePagerQuery(buildDeps(depsParams{subjectType: "capability", subjects: subjects, configs: &countingConfigSource{record: maturityConfig()}, facts: &countingFactsSource{}, maturity: maturity}))

	result, err := query.Get(context.Background(), mustSubjectType(t, "capability"), "cap-1")

	require.NoError(t, err)
	assert.Equal(t, "", result.Fields[0].BuiltIn.MaturitySection)
}

func TestGet_PropagatesMaturityScaleSectionsError(t *testing.T) {
	wantErr := errors.New("boom")
	subjects := &countingSubjectSource{snapshot: snapshotNamed("Cap", map[string]ports.BuiltInFieldValue{
		"maturity": ports.MaturityValue{Value: 85},
	})}
	maturity := &countingMaturitySource{err: wantErr}

	query := queries.NewOnePagerQuery(buildDeps(depsParams{subjectType: "capability", subjects: subjects, configs: &countingConfigSource{record: maturityConfig()}, facts: &countingFactsSource{}, maturity: maturity}))

	_, err := query.Get(context.Background(), mustSubjectType(t, "capability"), "cap-1")

	assert.ErrorIs(t, err, wantErr)
}
