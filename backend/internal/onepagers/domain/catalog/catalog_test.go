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

func TestEntriesFor_MatchesFullCatalogPerSubjectType(t *testing.T) {
	cases := map[string][]string{
		"capability":            {"name", "description", "maturity", "experts", "realizing-applications", "business-domains", "parent-capability", "child-capabilities", "depends-on"},
		"enterprise-capability": {"name", "description", "category", "included-capabilities"},
		"application":           {"name", "description", "experts", "realized-capabilities", "built-by", "purchased-from", "acquired-via", "component-relations"},
		"acquired-entity":       {"name", "acquisition-date", "integration-status", "acquired-applications"},
		"vendor":                {"name", "implementation-partner", "notes", "purchased-applications"},
		"internal-team":         {"name", "department", "contact-person", "built-applications"},
	}
	for subject, expected := range cases {
		entries := EntriesFor(subjectType(t, subject))
		assert.Equal(t, expected, entryIDs(entries), subject)
	}
}

func TestDefaultEntriesFor_ExcludesRelations(t *testing.T) {
	cases := map[string][]string{
		"capability":            {"name", "description", "maturity", "experts"},
		"enterprise-capability": {"name", "description", "category"},
		"application":           {"name", "description", "experts"},
		"acquired-entity":       {"name", "acquisition-date", "integration-status"},
		"vendor":                {"name", "implementation-partner", "notes"},
		"internal-team":         {"name", "department", "contact-person"},
	}
	for subject, expected := range cases {
		entries := DefaultEntriesFor(subjectType(t, subject))
		assert.Equal(t, expected, entryIDs(entries), subject)
	}
}

func TestEntriesFor_ExposesRelationLabelsPerSubjectType(t *testing.T) {
	cases := map[string]map[string]string{
		"capability": {
			"realizing-applications": "Realizing Applications",
			"business-domains":       "Business Domains",
			"parent-capability":      "Parent Capability",
			"child-capabilities":     "Child Capabilities",
			"depends-on":             "Depends On",
		},
		"enterprise-capability": {"included-capabilities": "Included Capabilities"},
		"application": {
			"realized-capabilities": "Realized Capabilities",
			"built-by":              "Built By",
			"purchased-from":        "Purchased From",
			"acquired-via":          "Acquired Via",
			"component-relations":   "Triggers / Serves",
		},
		"acquired-entity": {"acquired-applications": "Applications"},
		"vendor":          {"purchased-applications": "Applications"},
		"internal-team":   {"built-applications": "Applications"},
	}
	for subject, relations := range cases {
		for entryID, label := range relations {
			entry, found := LookupEntry(subjectType(t, subject), entryID)
			require.Truef(t, found, "%s/%s", subject, entryID)
			assert.Equalf(t, label, entry.Label, "%s/%s", subject, entryID)
			assert.Truef(t, entry.Relation, "%s/%s must be marked as a relation", subject, entryID)
		}
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
