package api

import (
	"time"

	"easi/backend/internal/onepagers/application/readmodels"
	"easi/backend/internal/onepagers/domain/catalog"
	"easi/backend/internal/onepagers/domain/valueobjects"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/types"
)

type SelectionOptionDTO struct {
	ID     string      `json:"id"`
	Label  string      `json:"label"`
	Active bool        `json:"active"`
	Links  types.Links `json:"_links,omitempty"`
}

type CustomFieldDTO struct {
	ID       string               `json:"id"`
	Name     string               `json:"name"`
	Type     string               `json:"type"`
	Required bool                 `json:"required"`
	HelpText string               `json:"helpText"`
	Active   bool                 `json:"active"`
	Options  []SelectionOptionDTO `json:"options,omitempty"`
	Min      *float64             `json:"min,omitempty"`
	Max      *float64             `json:"max,omitempty"`
	Links    types.Links          `json:"_links,omitempty"`
}

type BuiltInFieldDTO struct {
	ID       string      `json:"id"`
	Label    string      `json:"label"`
	Included bool        `json:"included"`
	Required bool        `json:"required"`
	Links    types.Links `json:"_links,omitempty"`
}

type FieldRefDTO struct {
	Kind string `json:"kind"`
	ID   string `json:"id"`
}

type OnePagerConfigurationDTO struct {
	ID            string            `json:"id"`
	SubjectType   string            `json:"subjectType"`
	BuiltInFields []BuiltInFieldDTO `json:"builtInFields"`
	CustomFields  []CustomFieldDTO  `json:"customFields"`
	DisplayOrder  []FieldRefDTO     `json:"displayOrder"`
	Version       int               `json:"version"`
	CreatedAt     time.Time         `json:"createdAt"`
	ModifiedAt    time.Time         `json:"modifiedAt"`
	ModifiedBy    string            `json:"modifiedBy"`
	Links         types.Links       `json:"_links,omitempty"`
}

func BuildConfigurationDTO(record *readmodels.ConfigurationRecord, links *OnePagerLinks, actor sharedctx.Actor) OnePagerConfigurationDTO {
	ctx := linkContext{subjectType: record.SubjectType, actor: actor}
	return OnePagerConfigurationDTO{
		ID:            record.ID,
		SubjectType:   record.SubjectType,
		BuiltInFields: buildBuiltInFieldDTOs(record, links, ctx),
		CustomFields:  buildCustomFieldDTOs(record, links, ctx),
		DisplayOrder:  buildDisplayOrderDTOs(record),
		Version:       record.Version,
		CreatedAt:     record.CreatedAt,
		ModifiedAt:    record.ModifiedAt,
		ModifiedBy:    record.ModifiedBy,
		Links:         links.ConfigurationLinks(ctx),
	}
}

func buildBuiltInFieldDTOs(record *readmodels.ConfigurationRecord, links *OnePagerLinks, ctx linkContext) []BuiltInFieldDTO {
	included := make(map[string]bool)
	for _, ref := range record.Document.DisplayOrder {
		if ref.Kind == string(valueobjects.FieldRefKindBuiltIn) {
			included[ref.ID] = true
		}
	}

	subjectType, err := valueobjects.NewSubjectType(record.SubjectType)
	if err != nil {
		return nil
	}
	entries := catalog.EntriesFor(subjectType)
	dtos := make([]BuiltInFieldDTO, len(entries))
	for i, entry := range entries {
		dto := BuiltInFieldDTO{
			ID:       entry.ID,
			Label:    entry.Label,
			Included: included[entry.ID],
			Required: record.Document.BuiltInRequired(entry.ID),
		}
		dto.Links = links.builtInFieldLinks(ctx, dto)
		dtos[i] = dto
	}
	return dtos
}

func buildCustomFieldDTOs(record *readmodels.ConfigurationRecord, links *OnePagerLinks, ctx linkContext) []CustomFieldDTO {
	dtos := make([]CustomFieldDTO, len(record.Document.CustomFields))
	for i, field := range record.Document.CustomFields {
		dto := CustomFieldDTO{
			ID:       field.ID,
			Name:     field.Name,
			Type:     field.Type,
			Required: field.Required,
			HelpText: field.HelpText,
			Active:   field.Active,
			Options:  buildOptionDTOs(ctx, field, links),
			Min:      field.Min,
			Max:      field.Max,
		}
		dto.Links = links.customFieldLinks(ctx, dto)
		dtos[i] = dto
	}
	return dtos
}

func buildOptionDTOs(ctx linkContext, field readmodels.CustomFieldRecord, links *OnePagerLinks) []SelectionOptionDTO {
	if len(field.Options) == 0 {
		return nil
	}
	dtos := make([]SelectionOptionDTO, len(field.Options))
	for i, option := range field.Options {
		dto := SelectionOptionDTO{ID: option.ID, Label: option.Label, Active: option.Active}
		dto.Links = links.optionLinks(optionLinkParams{
			linkContext: ctx,
			fieldID:     field.ID,
			option:      dto,
			fieldActive: field.Active,
		})
		dtos[i] = dto
	}
	return dtos
}

func buildDisplayOrderDTOs(record *readmodels.ConfigurationRecord) []FieldRefDTO {
	order := make([]FieldRefDTO, len(record.Document.DisplayOrder))
	for i, ref := range record.Document.DisplayOrder {
		order[i] = FieldRefDTO{Kind: ref.Kind, ID: ref.ID}
	}
	return order
}
