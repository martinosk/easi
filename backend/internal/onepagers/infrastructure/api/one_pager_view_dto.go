package api

import (
	"easi/backend/internal/onepagers/application/ports"
	"easi/backend/internal/onepagers/application/queries"
	"easi/backend/internal/onepagers/domain/valueobjects"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/types"
)

type OnePagerDTO struct {
	SubjectType  string             `json:"subjectType"`
	SubjectID    string             `json:"subjectId"`
	SubjectName  string             `json:"subjectName"`
	Fields       []OnePagerFieldDTO `json:"fields"`
	Completeness CompletenessDTO    `json:"completeness"`
	Links        types.Links        `json:"_links,omitempty"`
}

type CompletenessDTO struct {
	RequiredCount int               `json:"requiredCount"`
	FilledCount   int               `json:"filledCount"`
	MissingFields []MissingFieldDTO `json:"missingFields"`
}

type MissingFieldDTO struct {
	FieldID string `json:"fieldId"`
	Name    string `json:"name"`
}

type OnePagerFieldDTO struct {
	Kind    string               `json:"kind"`
	BuiltIn *BuiltInFieldViewDTO `json:"builtIn,omitempty"`
	Custom  *CustomFieldViewDTO  `json:"custom,omitempty"`
}

type BuiltInFieldViewDTO struct {
	ID    string           `json:"id"`
	Label string           `json:"label"`
	Value *BuiltInValueDTO `json:"value"`
}

type BuiltInValueDTO struct {
	Type       string            `json:"type"`
	Text       *string           `json:"text,omitempty"`
	Date       *string           `json:"date,omitempty"`
	Maturity   *MaturityValueDTO `json:"maturity,omitempty"`
	Experts    []ExpertViewDTO   `json:"experts,omitempty"`
	References []ReferenceDTO    `json:"references,omitempty"`
}

type ReferenceDTO struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	SubjectType string `json:"subjectType,omitempty"`
}

type MaturityValueDTO struct {
	Value   int    `json:"value"`
	Section string `json:"section,omitempty"`
}

type ExpertViewDTO struct {
	Name    string `json:"name"`
	Role    string `json:"role"`
	Contact string `json:"contact"`
}

type CustomFieldViewDTO struct {
	FieldID       string            `json:"fieldId"`
	Name          string            `json:"name"`
	Type          string            `json:"type"`
	HelpText      string            `json:"helpText,omitempty"`
	Value         *ValueEnvelopeDTO `json:"value"`
	DisplayText   string            `json:"displayText,omitempty"`
	RetiredOption bool              `json:"retiredOption,omitempty"`
	OutOfBounds   bool              `json:"outOfBounds,omitempty"`
}

func BuildOnePagerDTO(onePager *queries.OnePager, links *OnePagerLinks, actor sharedctx.Actor) OnePagerDTO {
	fields := make([]OnePagerFieldDTO, 0, len(onePager.Fields))
	for _, field := range onePager.Fields {
		if dto, ok := onePagerFieldDTOFrom(field); ok {
			fields = append(fields, dto)
		}
	}
	return OnePagerDTO{
		SubjectType:  onePager.SubjectType,
		SubjectID:    onePager.SubjectID,
		SubjectName:  onePager.SubjectName,
		Fields:       fields,
		Completeness: completenessDTOFrom(onePager.Completeness),
		Links:        links.viewLinks(onePager.SubjectType, onePager.SubjectID, actor),
	}
}

func completenessDTOFrom(completeness queries.Completeness) CompletenessDTO {
	missingFields := make([]MissingFieldDTO, len(completeness.MissingFields))
	for i, field := range completeness.MissingFields {
		missingFields[i] = MissingFieldDTO{FieldID: field.FieldID, Name: field.Name}
	}
	return CompletenessDTO{
		RequiredCount: completeness.RequiredCount,
		FilledCount:   completeness.FilledCount,
		MissingFields: missingFields,
	}
}

func onePagerFieldDTOFrom(field queries.Field) (OnePagerFieldDTO, bool) {
	switch {
	case field.BuiltIn != nil:
		return OnePagerFieldDTO{Kind: "builtIn", BuiltIn: builtInFieldViewDTOFrom(field.BuiltIn)}, true
	case field.Custom != nil:
		return OnePagerFieldDTO{Kind: "custom", Custom: customFieldViewDTOFrom(field.Custom)}, true
	default:
		return OnePagerFieldDTO{}, false
	}
}

func builtInFieldViewDTOFrom(field *queries.BuiltInField) *BuiltInFieldViewDTO {
	return &BuiltInFieldViewDTO{
		ID:    field.ID,
		Label: field.Label,
		Value: builtInValueDTOFrom(field),
	}
}

func builtInValueDTOFrom(field *queries.BuiltInField) *BuiltInValueDTO {
	switch value := field.Value.(type) {
	case ports.TextValue:
		text := value.Text
		return &BuiltInValueDTO{Type: "text", Text: &text}
	case ports.DateValue:
		date := value.Date.Format(valueobjects.ISODateLayout)
		return &BuiltInValueDTO{Type: "date", Date: &date}
	case ports.MaturityValue:
		return &BuiltInValueDTO{Type: "maturity", Maturity: &MaturityValueDTO{Value: value.Value, Section: field.MaturitySection}}
	case ports.ExpertsValue:
		return &BuiltInValueDTO{Type: "experts", Experts: expertViewDTOsFrom(value.Experts)}
	case ports.ReferenceListValue:
		if len(value.References) == 0 {
			return nil
		}
		return &BuiltInValueDTO{Type: "references", References: referenceDTOsFrom(value.References)}
	default:
		return nil
	}
}

func referenceDTOsFrom(references []ports.Reference) []ReferenceDTO {
	dtos := make([]ReferenceDTO, len(references))
	for i, reference := range references {
		dtos[i] = ReferenceDTO{ID: reference.ID, Label: reference.Label, SubjectType: reference.SubjectType}
	}
	return dtos
}

func expertViewDTOsFrom(experts []ports.Expert) []ExpertViewDTO {
	dtos := make([]ExpertViewDTO, len(experts))
	for i, expert := range experts {
		dtos[i] = ExpertViewDTO{Name: expert.Name, Role: expert.Role, Contact: expert.Contact}
	}
	return dtos
}

func customFieldViewDTOFrom(field *queries.CustomField) *CustomFieldViewDTO {
	dto := &CustomFieldViewDTO{
		FieldID:       field.FieldID,
		Name:          field.Name,
		Type:          field.FieldType,
		HelpText:      field.HelpText,
		DisplayText:   field.DisplayText,
		RetiredOption: field.RetiredOption,
		OutOfBounds:   field.OutOfBounds,
	}
	if field.Value != nil {
		envelope := envelopeDTOFrom(*field.Value)
		dto.Value = &envelope
	}
	return dto
}
