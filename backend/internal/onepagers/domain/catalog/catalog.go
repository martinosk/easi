package catalog

import (
	"easi/backend/internal/onepagers/domain/valueobjects"
)

type Entry struct {
	ID    string
	Label string
}

var entriesBySubjectType = map[string][]Entry{
	"capability": {
		{ID: "name", Label: "Name"},
		{ID: "description", Label: "Description"},
		{ID: "maturity", Label: "Maturity"},
		{ID: "experts", Label: "Experts"},
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
	},
	"acquired-entity": {
		{ID: "name", Label: "Name"},
		{ID: "acquisition-date", Label: "Acquisition Date"},
		{ID: "integration-status", Label: "Integration Status"},
	},
	"vendor": {
		{ID: "name", Label: "Name"},
		{ID: "implementation-partner", Label: "Implementation Partner"},
		{ID: "notes", Label: "Notes"},
	},
	"internal-team": {
		{ID: "name", Label: "Name"},
		{ID: "department", Label: "Department"},
		{ID: "contact-person", Label: "Contact Person"},
	},
}

func EntriesFor(subjectType valueobjects.SubjectType) []Entry {
	entries := entriesBySubjectType[subjectType.Value()]
	copied := make([]Entry, len(entries))
	copy(copied, entries)
	return copied
}

func LookupEntry(subjectType valueobjects.SubjectType, entryID string) (Entry, bool) {
	for _, entry := range entriesBySubjectType[subjectType.Value()] {
		if entry.ID == entryID {
			return entry, true
		}
	}
	return Entry{}, false
}
