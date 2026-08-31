package projectors_test

import (
	"os"
	"regexp"
	"strings"
	"testing"

	amPL "easi/backend/internal/architecturemodeling/publishedlanguage"
	capPL "easi/backend/internal/capabilitymapping/publishedlanguage"
	"easi/backend/internal/onepagers/domain/catalog"
	"easi/backend/internal/onepagers/domain/valueobjects"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const migration148Path = "../../../../deploy-scripts/migrations/148_backfill_onepagers_subject_caches.sql"

type relationCoverageFixture struct {
	eventType string
	payload   map[string]any
}

var relationCoverageFixtures = []relationCoverageFixture{
	{capPL.SystemLinkedToCapability, map[string]any{"id": "r-1", "capabilityId": "cap-1", "componentId": "app-9"}},
	{capPL.CapabilityDependencyCreated, map[string]any{"id": "d-1", "sourceCapabilityId": "cap-1", "targetCapabilityId": "cap-2"}},
	{capPL.CapabilityAssignedToDomain, map[string]any{"id": "a-1", "businessDomainId": "bd-1", "capabilityId": "cap-1"}},
	{capPL.CapabilityCreated, map[string]any{"id": "cap-2", "name": "Billing", "parentId": "cap-1"}},
	{amPL.ComponentRelationCreated, map[string]any{"id": "cr-1", "sourceComponentId": "app-1", "targetComponentId": "app-2"}},
	{amPL.OriginLinkSet, map[string]any{"componentId": "app-1", "originType": "built-by", "entityId": "t-1"}},
	{amPL.OriginLinkSet, map[string]any{"componentId": "app-2", "originType": "purchased-from", "entityId": "v-1"}},
	{amPL.OriginLinkSet, map[string]any{"componentId": "app-3", "originType": "acquired-via", "entityId": "e-1"}},
}

func producedRelationEntries(t *testing.T) map[string]bool {
	t.Helper()
	h := newRelationHarness(t)
	for _, fixture := range relationCoverageFixtures {
		h.project(fixture.eventType, fixture.payload)
	}

	produced := map[string]bool{}
	for _, saved := range h.relations.saved {
		produced[saved.subject.SubjectType+"/"+saved.entry.EntryID] = true
	}
	for _, replaced := range h.relations.replaced {
		if len(replaced.entries) == 0 {
			continue
		}
		produced[replaced.subject.SubjectType+"/"+replaced.entryID] = true
	}
	return produced
}

var subjectRelationCacheInsertPattern = regexp.MustCompile(`(?is)INSERT INTO onepagers.subject_relation_cache\s*\(([^)]+)\)\s*SELECT(.*?)\sFROM\s`)

var quotedLiteralPattern = regexp.MustCompile(`'([^']*)'`)

func backfillRelationEntryIDs(t *testing.T) map[string]bool {
	t.Helper()
	content, err := os.ReadFile(migration148Path)
	require.NoError(t, err)

	entryIDs := map[string]bool{}
	matches := subjectRelationCacheInsertPattern.FindAllStringSubmatch(string(content), -1)
	require.NotEmpty(t, matches, "no subject_relation_cache backfill statements found in migration 148")

	for _, match := range matches {
		columns := splitTopLevelCommas(match[1])
		index := columnIndex(columns, "entry_id")
		require.GreaterOrEqualf(t, index, 0, "backfill statement is missing an entry_id column: %s", match[1])

		values := splitTopLevelCommas(match[2])
		require.Greaterf(t, len(values), index, "backfill SELECT list is shorter than its column list: %s", match[2])

		entryIDs[quotedLiteral(t, values[index])] = true
	}
	return entryIDs
}

func splitTopLevelCommas(text string) []string {
	var parts []string
	depth := 0
	last := 0
	for i, r := range text {
		switch r {
		case '(':
			depth++
		case ')':
			depth--
		case ',':
			if depth == 0 {
				parts = append(parts, text[last:i])
				last = i + 1
			}
		}
	}
	parts = append(parts, text[last:])
	return parts
}

func columnIndex(columns []string, name string) int {
	for i, column := range columns {
		if strings.TrimSpace(column) == name {
			return i
		}
	}
	return -1
}

func quotedLiteral(t *testing.T, expr string) string {
	t.Helper()
	match := quotedLiteralPattern.FindStringSubmatch(expr)
	require.NotNilf(t, match, "expected a quoted literal, got %q", expr)
	return match[1]
}

func TestRelationCatalogEntries_HaveProjectorAndBackfillCoverage(t *testing.T) {
	produced := producedRelationEntries(t)
	backfilled := backfillRelationEntryIDs(t)

	for _, subjectType := range valueobjects.AllSubjectTypes() {
		for _, entry := range catalog.EntriesFor(subjectType) {
			if !entry.Relation {
				continue
			}
			key := subjectType.Value() + "/" + entry.ID
			assert.Containsf(t, produced, key, "%s has no projector-handled event path caching it", key)
			assert.Containsf(t, backfilled, entry.ID, "%s has no backfill statement in migration 148 for entry_id %q", key, entry.ID)
		}
	}
}
