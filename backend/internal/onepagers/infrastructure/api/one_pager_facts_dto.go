package api

import (
	"encoding/json"
	"time"

	"easi/backend/internal/onepagers/application/readmodels"
	"easi/backend/internal/onepagers/domain/valueobjects"
	"easi/backend/internal/shared/types"
)

type ValueEnvelopeDTO struct {
	Type    string          `json:"type"`
	Version int             `json:"version"`
	Value   json.RawMessage `json:"value" swaggertype:"object"`
}

func (d ValueEnvelopeDTO) toDomain() valueobjects.ValueEnvelope {
	return valueobjects.ValueEnvelope{Type: d.Type, Version: d.Version, Value: d.Value}
}

func envelopeDTOFrom(envelope valueobjects.ValueEnvelope) ValueEnvelopeDTO {
	return ValueEnvelopeDTO{Type: envelope.Type, Version: envelope.Version, Value: envelope.Value}
}

type FieldValueDTO struct {
	FieldID       string           `json:"fieldId"`
	Value         ValueEnvelopeDTO `json:"value"`
	DisplayText   string           `json:"displayText"`
	RetiredOption bool             `json:"retiredOption,omitempty"`
	ModifiedAt    time.Time        `json:"modifiedAt"`
	ModifiedBy    string           `json:"modifiedBy"`
	Links         types.Links      `json:"_links,omitempty"`
}

type OnePagerFactsDTO struct {
	SubjectType string          `json:"subjectType"`
	SubjectID   string          `json:"subjectId"`
	Values      []FieldValueDTO `json:"values"`
	Links       types.Links     `json:"_links,omitempty"`
}

type factsDTOParams struct {
	subjectType string
	subjectID   string
	records     []readmodels.FactRecord
	config      *readmodels.ConfigurationRecord
	links       *OnePagerLinks
	ctx         factsLinkContext
}

func BuildFactsDTO(params factsDTOParams) OnePagerFactsDTO {
	values := make([]FieldValueDTO, len(params.records))
	for i, record := range params.records {
		dto := FieldValueDTO{
			FieldID:       record.FieldID,
			DisplayText:   record.DisplayText,
			RetiredOption: isRetiredOptionValue(record, params.config),
			ModifiedAt:    record.ModifiedAt,
			ModifiedBy:    record.ModifiedBy,
		}
		if record.Value != nil {
			dto.Value = envelopeDTOFrom(*record.Value)
		}
		dto.Links = params.links.fieldValueLinks(params.ctx, record.FieldID)
		values[i] = dto
	}

	return OnePagerFactsDTO{
		SubjectType: params.subjectType,
		SubjectID:   params.subjectID,
		Values:      values,
		Links:       params.links.factsLinks(params.ctx),
	}
}

func isRetiredOptionValue(record readmodels.FactRecord, config *readmodels.ConfigurationRecord) bool {
	if config == nil {
		return false
	}
	selection, ok := decodeSelectionValue(record)
	if !ok {
		return false
	}
	field, found := findFieldRecord(config, record.FieldID)
	if !found {
		return false
	}
	return hasRetiredOption(field, selection.OptionID().Value())
}

func decodeSelectionValue(record readmodels.FactRecord) (valueobjects.SelectionValue, bool) {
	if record.Value == nil {
		return valueobjects.SelectionValue{}, false
	}
	value, err := valueobjects.FieldValueFromEnvelope(*record.Value)
	if err != nil {
		return valueobjects.SelectionValue{}, false
	}
	selection, ok := value.(valueobjects.SelectionValue)
	return selection, ok
}

func findFieldRecord(config *readmodels.ConfigurationRecord, fieldID string) (readmodels.CustomFieldRecord, bool) {
	for _, field := range config.Document.CustomFields {
		if field.ID == fieldID {
			return field, true
		}
	}
	return readmodels.CustomFieldRecord{}, false
}

func hasRetiredOption(field readmodels.CustomFieldRecord, optionID string) bool {
	for _, option := range field.Options {
		if option.ID == optionID {
			return !option.Active
		}
	}
	return false
}
