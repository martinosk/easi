package queries

import (
	"easi/backend/internal/onepagers/application/readmodels"
	"easi/backend/internal/onepagers/domain/catalog"
	"easi/backend/internal/onepagers/domain/valueobjects"
)

func documentFor(config *readmodels.ConfigurationRecord, subjectType valueobjects.SubjectType) readmodels.ConfigurationDocument {
	if config != nil {
		return config.Document
	}
	return defaultDocument(subjectType)
}

func defaultDocument(subjectType valueobjects.SubjectType) readmodels.ConfigurationDocument {
	entries := catalog.DefaultEntriesFor(subjectType)
	order := make([]readmodels.FieldRefRecord, len(entries))
	for i, entry := range entries {
		order[i] = readmodels.FieldRefRecord{Kind: string(valueobjects.FieldRefKindBuiltIn), ID: entry.ID}
	}
	return readmodels.ConfigurationDocument{DisplayOrder: order}
}
