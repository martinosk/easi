package adapters

import (
	"encoding/json"
	"fmt"
	"time"

	"easi/backend/internal/onepagers/application/ports"
	"easi/backend/internal/onepagers/application/readmodels"
)

const nameEntryID = "name"

type attributeDecoder func(json.RawMessage) (ports.BuiltInFieldValue, error)

type builtInAttribute struct {
	entryID   string
	attribute string
	decode    attributeDecoder
}

var builtInAttributesBySubjectType = map[string][]builtInAttribute{
	"capability": {
		{entryID: "description", attribute: "description", decode: decodeText},
		{entryID: "maturity", attribute: "maturityValue", decode: decodeMaturity},
		{entryID: "experts", attribute: "experts", decode: decodeExperts},
	},
	"enterprise-capability": {
		{entryID: "description", attribute: "description", decode: decodeText},
		{entryID: "category", attribute: "category", decode: decodeText},
	},
	"application": {
		{entryID: "description", attribute: "description", decode: decodeText},
		{entryID: "experts", attribute: "experts", decode: decodeExperts},
	},
	"acquired-entity": {
		{entryID: "acquisition-date", attribute: "acquisitionDate", decode: decodeDate},
		{entryID: "integration-status", attribute: "integrationStatus", decode: decodeText},
	},
	"vendor": {
		{entryID: "implementation-partner", attribute: "implementationPartner", decode: decodeText},
		{entryID: "notes", attribute: "notes", decode: decodeText},
	},
	"internal-team": {
		{entryID: "department", attribute: "department", decode: decodeText},
		{entryID: "contact-person", attribute: "contactPerson", decode: decodeText},
	},
}

func decodeText(raw json.RawMessage) (ports.BuiltInFieldValue, error) {
	var text string
	if err := decodeAttribute(raw, &text); err != nil {
		return nil, err
	}
	if text == "" {
		return nil, nil
	}
	return ports.TextValue{Text: text}, nil
}

func decodeDate(raw json.RawMessage) (ports.BuiltInFieldValue, error) {
	var date *time.Time
	if err := decodeAttribute(raw, &date); err != nil {
		return nil, err
	}
	if date == nil {
		return nil, nil
	}
	return ports.DateValue{Date: *date}, nil
}

func decodeMaturity(raw json.RawMessage) (ports.BuiltInFieldValue, error) {
	var maturity int
	if err := decodeAttribute(raw, &maturity); err != nil {
		return nil, err
	}
	return ports.MaturityValue{Value: maturity}, nil
}

func decodeExperts(raw json.RawMessage) (ports.BuiltInFieldValue, error) {
	var cached []readmodels.SubjectExpert
	if err := decodeAttribute(raw, &cached); err != nil {
		return nil, err
	}
	if len(cached) == 0 {
		return nil, nil
	}
	experts := make([]ports.Expert, len(cached))
	for i, expert := range cached {
		experts[i] = ports.Expert{Name: expert.Name, Role: expert.Role, Contact: expert.Contact}
	}
	return ports.ExpertsValue{Experts: experts}, nil
}

func decodeAttribute(raw json.RawMessage, target any) error {
	if len(raw) == 0 {
		return nil
	}
	if err := json.Unmarshal(raw, target); err != nil {
		return fmt.Errorf("decode cached built-in attribute: %w", err)
	}
	return nil
}
