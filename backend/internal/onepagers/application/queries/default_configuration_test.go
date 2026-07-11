package queries_test

import (
	"context"
	"testing"

	"easi/backend/internal/onepagers/application/queries"
	"easi/backend/internal/onepagers/domain/catalog"
	"easi/backend/internal/onepagers/domain/valueobjects"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGet_SynthesizesDefaultConfigurationForEverySubjectType(t *testing.T) {
	for _, subjectType := range valueobjects.AllSubjectTypes() {
		t.Run(subjectType.Value(), func(t *testing.T) {
			subjects := &countingSubjectSource{snapshot: snapshotNamed("Subject", nil)}
			configs := &countingConfigSource{record: nil}
			facts := &countingFactsSource{}

			query := queries.NewOnePagerQuery(buildDeps(depsParams{subjectType: subjectType.Value(), subjects: subjects, configs: configs, facts: facts}))

			result, err := query.Get(context.Background(), subjectType, "subject-1")

			require.NoError(t, err)
			wantEntries := catalog.EntriesFor(subjectType)
			require.Len(t, result.Fields, len(wantEntries))
			for i, entry := range wantEntries {
				require.NotNil(t, result.Fields[i].BuiltIn, "entry %d (%s) should be built-in", i, entry.ID)
				assert.Nil(t, result.Fields[i].Custom)
				assert.Equal(t, entry.ID, result.Fields[i].BuiltIn.ID)
				assert.Equal(t, entry.Label, result.Fields[i].BuiltIn.Label)
			}
		})
	}
}

func TestGet_NilConfigurationNeverPersistsAnything(t *testing.T) {
	subjects := &countingSubjectSource{snapshot: snapshotNamed("Subject", nil)}
	configs := &countingConfigSource{record: nil}
	facts := &countingFactsSource{}

	query := queries.NewOnePagerQuery(buildDeps(depsParams{subjectType: "capability", subjects: subjects, configs: configs, facts: facts}))

	_, err := query.Get(context.Background(), mustSubjectType(t, "capability"), "subject-1")

	require.NoError(t, err)
	assert.Equal(t, 1, configs.calls)
}
