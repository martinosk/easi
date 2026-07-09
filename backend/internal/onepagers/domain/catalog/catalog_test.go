package catalog

import (
	"testing"

	"easi/backend/internal/onepagers/domain/valueobjects"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func subjectType(t *testing.T, value string) valueobjects.SubjectType {
	t.Helper()
	st, err := valueobjects.NewSubjectType(value)
	require.NoError(t, err)
	return st
}

func entryIDs(entries []Entry) []string {
	ids := make([]string, len(entries))
	for i, e := range entries {
		ids[i] = e.ID
	}
	return ids
}

func TestEntriesFor_MatchesInitialCatalogPerSubjectType(t *testing.T) {
	cases := map[string][]string{
		"capability":            {"name", "description", "maturity", "experts"},
		"enterprise-capability": {"name", "description", "category"},
		"application":           {"name", "description", "experts"},
		"acquired-entity":       {"name", "acquisition-date", "integration-status"},
		"vendor":                {"name", "implementation-partner", "notes"},
		"internal-team":         {"name", "department", "contact-person"},
	}
	for subject, expected := range cases {
		entries := EntriesFor(subjectType(t, subject))
		assert.Equal(t, expected, entryIDs(entries), subject)
	}
}

func TestEntriesFor_EveryEntryHasALabel(t *testing.T) {
	for _, st := range valueobjects.AllSubjectTypes() {
		for _, entry := range EntriesFor(st) {
			assert.NotEmpty(t, entry.Label, "%s/%s", st.Value(), entry.ID)
		}
	}
}

func TestLookupEntry_FindsCatalogEntry(t *testing.T) {
	entry, found := LookupEntry(subjectType(t, "capability"), "maturity")
	require.True(t, found)
	assert.Equal(t, "Maturity", entry.Label)
}

func TestLookupEntry_RejectsEntryFromOtherSubjectType(t *testing.T) {
	_, found := LookupEntry(subjectType(t, "vendor"), "maturity")
	assert.False(t, found)
}

func TestEntriesFor_ReturnsACopy(t *testing.T) {
	st := subjectType(t, "vendor")
	entries := EntriesFor(st)
	entries[0].ID = "mutated"
	assert.Equal(t, "name", EntriesFor(st)[0].ID)
}
