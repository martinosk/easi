package queries

import "easi/backend/internal/onepagers/application/readmodels"

type Completeness struct {
	RequiredCount int
	FilledCount   int
	MissingFields []MissingField
}

type MissingField struct {
	FieldID string
	Name    string
}

func computeCompleteness(document readmodels.ConfigurationDocument, factsByFieldID map[string]readmodels.FactRecord) Completeness {
	completeness := Completeness{MissingFields: []MissingField{}}
	for _, field := range document.CustomFields {
		if !field.Active || !field.Required {
			continue
		}
		completeness.RequiredCount++
		if factFilled(factsByFieldID, field.ID) {
			completeness.FilledCount++
			continue
		}
		completeness.MissingFields = append(completeness.MissingFields, MissingField{FieldID: field.ID, Name: field.Name})
	}
	return completeness
}

func factFilled(factsByFieldID map[string]readmodels.FactRecord, fieldID string) bool {
	fact, ok := factsByFieldID[fieldID]
	return ok && fact.Value != nil
}
