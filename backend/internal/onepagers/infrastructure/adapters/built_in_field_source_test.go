package adapters_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/onepagers/application/ports"
	"easi/backend/internal/onepagers/application/readmodels"
	"easi/backend/internal/onepagers/domain/catalog"
	"easi/backend/internal/onepagers/domain/valueobjects"
	"easi/backend/internal/onepagers/infrastructure/adapters"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeAttributeStore struct {
	rows       map[string][]readmodels.SubjectAttributeRow
	count      int
	exists     map[string]bool
	err        error
	gotIDs     []string
	gotSubject readmodels.SubjectKey
}

func (f *fakeAttributeStore) AttributeRows(_ context.Context, subjectType string, subjectIDs []string) ([]readmodels.SubjectAttributeRow, error) {
	f.gotIDs = subjectIDs
	if f.err != nil {
		return nil, f.err
	}
	return f.rows[subjectType], nil
}

func (f *fakeAttributeStore) CountSubjects(_ context.Context, _ string) (int, error) {
	return f.count, f.err
}

func (f *fakeAttributeStore) Exists(_ context.Context, subject readmodels.SubjectKey) (bool, error) {
	f.gotSubject = subject
	return f.exists[subject.SubjectID], f.err
}

type fakeRelationReader struct {
	references  map[string][]readmodels.RelationReference
	count       int
	err         error
	gotEntryIDs []string
	gotEntryID  string
}

func (f *fakeRelationReader) References(_ context.Context, query readmodels.RelationQuery) (map[string][]readmodels.RelationReference, error) {
	f.gotEntryIDs = query.EntryIDs
	if f.err != nil {
		return nil, f.err
	}
	return f.references, nil
}

func (f *fakeRelationReader) CountSubjectsWithEntry(_ context.Context, _, entryID string) (int, error) {
	f.gotEntryID = entryID
	return f.count, f.err
}

func attributeRow(t *testing.T, id, name string, values map[string]any) readmodels.SubjectAttributeRow {
	t.Helper()
	attributes := readmodels.SubjectAttributes{}
	for key, value := range values {
		require.NoError(t, attributes.Set(key, value))
	}
	return readmodels.SubjectAttributeRow{SubjectID: id, Name: name, Attributes: attributes}
}

func sourceFor(subjectType string, subjects *fakeAttributeStore, relations *fakeRelationReader) ports.BuiltInFieldSource {
	return adapters.NewBuiltInFieldSources(subjects, relations)[subjectType]
}

func catalogAttributeIDs(t *testing.T, subjectType string) []string {
	t.Helper()
	st, err := valueobjects.NewSubjectType(subjectType)
	require.NoError(t, err)
	ids := []string{}
	for _, entry := range catalog.DefaultEntriesFor(st) {
		ids = append(ids, entry.ID)
	}
	return ids
}

func TestBuiltInFieldSources_KeysMatchSubjectTypes(t *testing.T) {
	sources := adapters.NewBuiltInFieldSources(&fakeAttributeStore{}, &fakeRelationReader{})

	got := make([]string, 0, len(sources))
	for subjectType := range sources {
		got = append(got, subjectType)
	}

	want := make([]string, 0, len(valueobjects.AllSubjectTypes()))
	for _, st := range valueobjects.AllSubjectTypes() {
		want = append(want, st.Value())
	}
	assert.ElementsMatch(t, want, got)
}

func TestBuiltInFieldSource_FetchSubject_CarriesEveryCatalogAttribute(t *testing.T) {
	for _, subjectType := range []string{"capability", "application", "acquired-entity", "vendor", "internal-team"} {
		t.Run(subjectType, func(t *testing.T) {
			subjects := &fakeAttributeStore{rows: map[string][]readmodels.SubjectAttributeRow{
				subjectType: {attributeRow(t, "s-1", "Subject", nil)},
			}}
			source := sourceFor(subjectType, subjects, &fakeRelationReader{})

			snapshot, err := source.FetchSubject(context.Background(), "s-1", nil)

			require.NoError(t, err)
			require.NotNil(t, snapshot)
			assert.Equal(t, "Subject", snapshot.Name)
			for _, entryID := range catalogAttributeIDs(t, subjectType) {
				_, present := snapshot.Fields[entryID]
				assert.Truef(t, present, "catalog entry %q missing from snapshot fields", entryID)
			}
		})
	}
}

func TestBuiltInFieldSource_FetchSubject_DecodesEveryValueType(t *testing.T) {
	acquired := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
	cases := []struct {
		name        string
		subjectType string
		values      map[string]any
		entryID     string
		want        ports.BuiltInFieldValue
	}{
		{"text", "capability", map[string]any{"description": "Invoices"}, "description", ports.TextValue{Text: "Invoices"}},
		{"empty text is unfilled", "capability", map[string]any{"description": ""}, "description", nil},
		{"absent text is unfilled", "capability", map[string]any{}, "description", nil},
		{"maturity", "capability", map[string]any{"maturityValue": 62}, "maturity", ports.MaturityValue{Value: 62}},
		{"absent maturity reads zero", "capability", map[string]any{}, "maturity", ports.MaturityValue{Value: 0}},
		{"date", "acquired-entity", map[string]any{"acquisitionDate": acquired}, "acquisition-date", ports.DateValue{Date: acquired}},
		{"absent date is unfilled", "acquired-entity", map[string]any{}, "acquisition-date", nil},
		{
			"experts", "application",
			map[string]any{"experts": []readmodels.SubjectExpert{{Name: "Alice", Role: "Owner", Contact: "alice@dfds.com"}}},
			"experts",
			ports.ExpertsValue{Experts: []ports.Expert{{Name: "Alice", Role: "Owner", Contact: "alice@dfds.com"}}},
		},
		{"empty experts are unfilled", "application", map[string]any{"experts": []readmodels.SubjectExpert{}}, "experts", nil},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			subjects := &fakeAttributeStore{rows: map[string][]readmodels.SubjectAttributeRow{
				tc.subjectType: {attributeRow(t, "s-1", "Subject", tc.values)},
			}}
			source := sourceFor(tc.subjectType, subjects, &fakeRelationReader{})

			snapshot, err := source.FetchSubject(context.Background(), "s-1", nil)

			require.NoError(t, err)
			assert.Equal(t, tc.want, snapshot.Fields[tc.entryID])
		})
	}
}

func TestBuiltInFieldSource_FetchSubject_NameComesFromTheSubjectRow(t *testing.T) {
	subjects := &fakeAttributeStore{rows: map[string][]readmodels.SubjectAttributeRow{
		"vendor": {attributeRow(t, "v-1", "Contoso", nil)},
	}}
	source := sourceFor("vendor", subjects, &fakeRelationReader{})

	snapshot, err := source.FetchSubject(context.Background(), "v-1", nil)

	require.NoError(t, err)
	assert.Equal(t, ports.TextValue{Text: "Contoso"}, snapshot.Fields["name"])
}

func TestBuiltInFieldSource_FetchSubject_MissingSubjectYieldsNilSnapshot(t *testing.T) {
	source := sourceFor("capability", &fakeAttributeStore{}, &fakeRelationReader{})

	snapshot, err := source.FetchSubject(context.Background(), "missing", nil)

	require.NoError(t, err)
	assert.Nil(t, snapshot)
}

func TestBuiltInFieldSource_FetchSubject_ResolvesOnlyIncludedRelations(t *testing.T) {
	subjects := &fakeAttributeStore{rows: map[string][]readmodels.SubjectAttributeRow{
		"capability": {attributeRow(t, "cap-1", "Billing", nil)},
	}}
	relations := &fakeRelationReader{references: map[string][]readmodels.RelationReference{
		"cap-1": {{EntryID: "realizing-applications", RelatedType: "application", RelatedID: "app-9", Label: "Payments"}},
	}}
	source := sourceFor("capability", subjects, relations)

	snapshot, err := source.FetchSubject(context.Background(), "cap-1", []string{"name", "realizing-applications"})

	require.NoError(t, err)
	assert.Equal(t, []string{"realizing-applications"}, relations.gotEntryIDs, "only included relation entries are queried")
	assert.Equal(t, ports.ReferenceListValue{References: []ports.Reference{
		{ID: "app-9", Label: "Payments", SubjectType: "application"},
	}}, snapshot.Fields["realizing-applications"])
	_, hasExcluded := snapshot.Fields["depends-on"]
	assert.False(t, hasExcluded, "excluded relations never appear in the snapshot")
}

func TestBuiltInFieldSource_FetchSubject_IncludedRelationWithoutEdgesIsAnEmptyList(t *testing.T) {
	subjects := &fakeAttributeStore{rows: map[string][]readmodels.SubjectAttributeRow{
		"capability": {attributeRow(t, "cap-1", "Billing", nil)},
	}}
	source := sourceFor("capability", subjects, &fakeRelationReader{})

	snapshot, err := source.FetchSubject(context.Background(), "cap-1", []string{"depends-on"})

	require.NoError(t, err)
	value, present := snapshot.Fields["depends-on"]
	require.True(t, present)
	assert.IsType(t, ports.ReferenceListValue{}, value)
	assert.False(t, ports.ValueFilled(value))
}

func TestBuiltInFieldSource_FetchSubject_WrapsStoreErrors(t *testing.T) {
	wantErr := errors.New("boom")

	t.Run("attribute store", func(t *testing.T) {
		source := sourceFor("capability", &fakeAttributeStore{err: wantErr}, &fakeRelationReader{})

		_, err := source.FetchSubject(context.Background(), "cap-1", nil)

		assert.ErrorIs(t, err, wantErr)
	})

	t.Run("relation store", func(t *testing.T) {
		subjects := &fakeAttributeStore{rows: map[string][]readmodels.SubjectAttributeRow{
			"capability": {attributeRow(t, "cap-1", "Billing", nil)},
		}}
		source := sourceFor("capability", subjects, &fakeRelationReader{err: wantErr})

		_, err := source.FetchSubject(context.Background(), "cap-1", []string{"depends-on"})

		assert.ErrorIs(t, err, wantErr)
	})
}

func TestBuiltInFieldSource_CountSubjects(t *testing.T) {
	t.Run("counts the cached population", func(t *testing.T) {
		source := sourceFor("capability", &fakeAttributeStore{count: 42}, &fakeRelationReader{})

		count, err := source.CountSubjects(context.Background())

		require.NoError(t, err)
		assert.Equal(t, 42, count)
	})

	t.Run("wraps the store error", func(t *testing.T) {
		wantErr := errors.New("boom")
		source := sourceFor("capability", &fakeAttributeStore{err: wantErr}, &fakeRelationReader{})

		_, err := source.CountSubjects(context.Background())

		assert.ErrorIs(t, err, wantErr)
	})
}

func TestBuiltInFieldSource_FilledBuiltInFields(t *testing.T) {
	subjects := &fakeAttributeStore{rows: map[string][]readmodels.SubjectAttributeRow{
		"application": {
			attributeRow(t, "app-1", "Billing", map[string]any{
				"description": "Invoices",
				"experts":     []readmodels.SubjectExpert{},
			}),
			attributeRow(t, "app-2", "Payments", map[string]any{
				"experts": []readmodels.SubjectExpert{{Name: "Alice"}},
			}),
		},
	}}
	source := sourceFor("application", subjects, &fakeRelationReader{})

	filled, err := source.FilledBuiltInFields(context.Background(), []string{"app-1", "app-2"}, []string{"description", "experts"})

	require.NoError(t, err)
	assert.Equal(t, []string{"app-1", "app-2"}, subjects.gotIDs)
	assert.Equal(t, map[string]map[string]bool{
		"app-1": {"description": true, "experts": false},
		"app-2": {"description": false, "experts": true},
	}, filled)
}

func TestBuiltInFieldSource_FilledBuiltInFields_CountsRelationEntries(t *testing.T) {
	subjects := &fakeAttributeStore{rows: map[string][]readmodels.SubjectAttributeRow{
		"capability": {attributeRow(t, "cap-1", "Billing", nil), attributeRow(t, "cap-2", "Orders", nil)},
	}}
	relations := &fakeRelationReader{references: map[string][]readmodels.RelationReference{
		"cap-1": {{EntryID: "depends-on", RelatedType: "capability", RelatedID: "cap-9", Label: "Ledger"}},
	}}
	source := sourceFor("capability", subjects, relations)

	filled, err := source.FilledBuiltInFields(context.Background(), []string{"cap-1", "cap-2"}, []string{"depends-on"})

	require.NoError(t, err)
	assert.Equal(t, map[string]map[string]bool{
		"cap-1": {"depends-on": true},
		"cap-2": {"depends-on": false},
	}, filled)
}

func TestBuiltInFieldSource_FilledBuiltInFields_EmptyInputsSkipTheStore(t *testing.T) {
	subjects := &fakeAttributeStore{err: errors.New("must not be called")}
	source := sourceFor("capability", subjects, &fakeRelationReader{})

	filled, err := source.FilledBuiltInFields(context.Background(), nil, []string{"description"})

	require.NoError(t, err)
	assert.Empty(t, filled)
}

func TestBuiltInFieldSource_CountSubjectsWithBuiltInValue(t *testing.T) {
	t.Run("attribute entries are counted from the cached attributes", func(t *testing.T) {
		subjects := &fakeAttributeStore{rows: map[string][]readmodels.SubjectAttributeRow{
			"application": {
				attributeRow(t, "app-1", "Billing", map[string]any{"experts": []readmodels.SubjectExpert{{Name: "Alice"}}}),
				attributeRow(t, "app-2", "Payments", nil),
				attributeRow(t, "app-3", "Ledger", map[string]any{"experts": []readmodels.SubjectExpert{{Name: "Bob"}}}),
			},
		}}
		source := sourceFor("application", subjects, &fakeRelationReader{})

		count, err := source.CountSubjectsWithBuiltInValue(context.Background(), "experts")

		require.NoError(t, err)
		assert.Equal(t, 2, count)
		assert.Nil(t, subjects.gotIDs, "the whole population is scanned")
	})

	t.Run("relation entries are counted from the relation cache", func(t *testing.T) {
		relations := &fakeRelationReader{count: 3}
		source := sourceFor("capability", &fakeAttributeStore{}, relations)

		count, err := source.CountSubjectsWithBuiltInValue(context.Background(), "depends-on")

		require.NoError(t, err)
		assert.Equal(t, 3, count)
		assert.Equal(t, "depends-on", relations.gotEntryID)
	})

	t.Run("wraps the store error", func(t *testing.T) {
		wantErr := errors.New("boom")
		source := sourceFor("capability", &fakeAttributeStore{err: wantErr}, &fakeRelationReader{})

		_, err := source.CountSubjectsWithBuiltInValue(context.Background(), "description")

		assert.ErrorIs(t, err, wantErr)
	})
}

func TestNewOnePagerBuiltInFieldSources_BuildsOverTheOwnedCaches(t *testing.T) {
	sources := adapters.NewOnePagerBuiltInFieldSources(database.NewTenantAwareDB(nil))

	assert.Len(t, sources, len(valueobjects.AllSubjectTypes()))
}
