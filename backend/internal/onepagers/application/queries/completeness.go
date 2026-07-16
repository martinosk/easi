package queries

import (
	"easi/backend/internal/onepagers/application/ports"
	"easi/backend/internal/onepagers/application/readmodels"
	"easi/backend/internal/onepagers/domain/catalog"
	"easi/backend/internal/onepagers/domain/valueobjects"
)

type Completeness struct {
	RequiredCount int
	FilledCount   int
	MissingFields []MissingField
}

type MissingField struct {
	FieldID string
	Name    string
}

func computeCompleteness(document readmodels.ConfigurationDocument, subjectType valueobjects.SubjectType, snapshot *ports.SubjectSnapshot, factsByFieldID map[string]readmodels.FactRecord) Completeness {
	completeness := Completeness{MissingFields: []MissingField{}}
	completeness.addRequiredCustomFields(document, factsByFieldID)
	completeness.addRequiredBuiltInFields(document, subjectType, snapshot)
	return completeness
}

func (c *Completeness) addRequiredCustomFields(document readmodels.ConfigurationDocument, factsByFieldID map[string]readmodels.FactRecord) {
	for _, field := range document.CustomFields {
		if !field.Active || !field.Required {
			continue
		}
		c.RequiredCount++
		if factFilled(factsByFieldID, field.ID) {
			c.FilledCount++
			continue
		}
		c.MissingFields = append(c.MissingFields, MissingField{FieldID: field.ID, Name: field.Name})
	}
}

func (c *Completeness) addRequiredBuiltInFields(document readmodels.ConfigurationDocument, subjectType valueobjects.SubjectType, snapshot *ports.SubjectSnapshot) {
	for _, ref := range document.DisplayOrder {
		if ref.Kind != string(valueobjects.FieldRefKindBuiltIn) || !document.BuiltInRequired(ref.ID) {
			continue
		}
		entry, found := catalog.LookupEntry(subjectType, ref.ID)
		if !found {
			continue
		}
		c.RequiredCount++
		if builtInFilled(snapshot, ref.ID) {
			c.FilledCount++
			continue
		}
		c.MissingFields = append(c.MissingFields, MissingField{FieldID: entry.ID, Name: entry.Label})
	}
}

func factFilled(factsByFieldID map[string]readmodels.FactRecord, fieldID string) bool {
	fact, ok := factsByFieldID[fieldID]
	return ok && fact.Value != nil
}

func builtInFilled(snapshot *ports.SubjectSnapshot, entryID string) bool {
	if snapshot == nil {
		return false
	}
	return ports.ValueFilled(snapshot.Fields[entryID])
}
