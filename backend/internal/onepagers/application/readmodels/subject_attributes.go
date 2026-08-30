package readmodels

import (
	"encoding/json"
	"fmt"
)

const ExpertsAttribute = "experts"

type SubjectExpert struct {
	Name    string `json:"expertName"`
	Role    string `json:"expertRole"`
	Contact string `json:"contactInfo"`
}

type SubjectAttributes map[string]json.RawMessage

type SubjectAttributeRow struct {
	SubjectID  string
	Name       string
	Attributes SubjectAttributes
}

func (a SubjectAttributes) Raw(key string) (json.RawMessage, bool) {
	raw, found := a[key]
	return raw, found
}

func (a SubjectAttributes) Experts() ([]SubjectExpert, error) {
	raw, found := a[ExpertsAttribute]
	if !found {
		return nil, nil
	}
	var experts []SubjectExpert
	if err := json.Unmarshal(raw, &experts); err != nil {
		return nil, fmt.Errorf("decode cached %q attribute: %w", ExpertsAttribute, err)
	}
	return experts, nil
}

func (a SubjectAttributes) Set(key string, value any) error {
	encoded, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("encode %q attribute: %w", key, err)
	}
	a[key] = encoded
	return nil
}

func (a SubjectAttributes) encode() ([]byte, error) {
	encoded, err := json.Marshal(map[string]json.RawMessage(a))
	if err != nil {
		return nil, fmt.Errorf("encode cached subject attributes: %w", err)
	}
	return encoded, nil
}

func decodeSubjectAttributes(raw []byte) (SubjectAttributes, error) {
	attributes := SubjectAttributes{}
	if len(raw) == 0 {
		return attributes, nil
	}
	if err := json.Unmarshal(raw, &attributes); err != nil {
		return nil, fmt.Errorf("decode cached subject attributes: %w", err)
	}
	return attributes, nil
}
