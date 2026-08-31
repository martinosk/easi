package catalog

import (
	"easi/backend/internal/onepagers/domain/valueobjects"
)

type Entry struct {
	ID       string
	Label    string
	Relation bool
}

var entriesBySubjectType = map[string][]Entry{
	"capability": {
		{ID: "name", Label: "Name"},
		{ID: "description", Label: "Description"},
		{ID: "maturity", Label: "Maturity"},
		{ID: "experts", Label: "Experts"},
		{ID: "realizing-applications", Label: "Realizing Applications", Relation: true},
		{ID: "business-domains", Label: "Business Domains", Relation: true},
		{ID: "parent-capability", Label: "Parent Capability", Relation: true},
		{ID: "child-capabilities", Label: "Child Capabilities", Relation: true},
		{ID: "depends-on", Label: "Depends On", Relation: true},
	},
	"enterprise-capability": {
		{ID: "name", Label: "Name"},
		{ID: "description", Label: "Description"},
		{ID: "category", Label: "Category"},
	},
	"application": {
		{ID: "name", Label: "Name"},
		{ID: "description", Label: "Description"},
		{ID: "experts", Label: "Experts"},
		{ID: "realized-capabilities", Label: "Realized Capabilities", Relation: true},
		{ID: "built-by", Label: "Built By", Relation: true},
		{ID: "purchased-from", Label: "Purchased From", Relation: true},
		{ID: "acquired-via", Label: "Acquired Via", Relation: true},
		{ID: "component-relations", Label: "Triggers / Serves", Relation: true},
	},
	"acquired-entity": {
		{ID: "name", Label: "Name"},
		{ID: "acquisition-date", Label: "Acquisition Date"},
		{ID: "integration-status", Label: "Integration Status"},
		{ID: "acquired-applications", Label: "Applications", Relation: true},
	},
	"vendor": {
		{ID: "name", Label: "Name"},
		{ID: "implementation-partner", Label: "Implementation Partner"},
		{ID: "notes", Label: "Notes"},
		{ID: "purchased-applications", Label: "Applications", Relation: true},
	},
	"internal-team": {
		{ID: "name", Label: "Name"},
		{ID: "department", Label: "Department"},
		{ID: "contact-person", Label: "Contact Person"},
		{ID: "built-applications", Label: "Applications", Relation: true},
	},
}

func EntriesFor(subjectType valueobjects.SubjectType) []Entry {
	entries := entriesBySubjectType[subjectType.Value()]
	copied := make([]Entry, len(entries))
	copy(copied, entries)
	return copied
}

func DefaultEntriesFor(subjectType valueobjects.SubjectType) []Entry {
	entries := entriesBySubjectType[subjectType.Value()]
	defaults := make([]Entry, 0, len(entries))
	for _, entry := range entries {
		if !entry.Relation {
			defaults = append(defaults, entry)
		}
	}
	return defaults
}

func LookupEntry(subjectType valueobjects.SubjectType, entryID string) (Entry, bool) {
	for _, entry := range entriesBySubjectType[subjectType.Value()] {
		if entry.ID == entryID {
			return entry, true
		}
	}
	return Entry{}, false
}
