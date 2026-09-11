package queries

import (
	"easi/backend/internal/onepagers/application/ports"
	"easi/backend/internal/onepagers/application/readmodels"
	"easi/backend/internal/onepagers/domain/catalog"
	"easi/backend/internal/onepagers/domain/valueobjects"
)

type fieldAssemblyInput struct {
	document       readmodels.ConfigurationDocument
	definitions    readmodels.CustomFieldDefinitions
	subjectType    valueobjects.SubjectType
	snapshot       *ports.SubjectSnapshot
	factsByFieldID map[string]readmodels.FactRecord
}

func assembleFields(input fieldAssemblyInput) []Field {
	fields := make([]Field, 0, len(input.document.DisplayOrder))
	for _, ref := range input.document.DisplayOrder {
		if field, ok := buildField(ref, input); ok {
			fields = append(fields, field)
		}
	}
	return fields
}

func indexFactsByFieldID(facts []readmodels.FactRecord) map[string]readmodels.FactRecord {
	byFieldID := make(map[string]readmodels.FactRecord, len(facts))
	for _, fact := range facts {
		byFieldID[fact.FieldID] = fact
	}
	return byFieldID
}

func buildField(ref readmodels.FieldRefRecord, input fieldAssemblyInput) (Field, bool) {
	switch valueobjects.FieldRefKind(ref.Kind) {
	case valueobjects.FieldRefKindBuiltIn:
		return buildBuiltInField(ref.ID, input.subjectType, input.snapshot)
	case valueobjects.FieldRefKindCustom:
		return buildCustomField(ref.ID, input.definitions, input.factsByFieldID)
	default:
		return Field{}, false
	}
}

func buildBuiltInField(entryID string, subjectType valueobjects.SubjectType, snapshot *ports.SubjectSnapshot) (Field, bool) {
	entry, found := catalog.LookupEntry(subjectType, entryID)
	if !found {
		return Field{}, false
	}
	return Field{BuiltIn: &BuiltInField{
		ID:    entry.ID,
		Label: entry.Label,
		Value: snapshot.Fields[entry.ID],
	}}, true
}

func buildCustomField(fieldID string, definitions readmodels.CustomFieldDefinitions, factsByFieldID map[string]readmodels.FactRecord) (Field, bool) {
	record, found := definitions.ByID(fieldID)
	if !found || !record.Active {
		return Field{}, false
	}
	custom := &CustomField{
		FieldID:   record.ID,
		Name:      record.Name,
		FieldType: record.Type,
		HelpText:  record.HelpText,
	}
	if fact, ok := factsByFieldID[fieldID]; ok {
		custom.Value = fact.Value
		custom.DisplayText = fact.DisplayText
		if label, ok := record.SelectionOptionLabel(fact.Value); ok {
			custom.DisplayText = label
		}
		custom.RetiredOption = record.RetiredOptionReferenced(fact.Value)
		custom.OutOfBounds = record.NumberValueOutOfBounds(fact.Value)
	}
	return Field{Custom: custom}, true
}
