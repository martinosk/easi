package api

import (
	"time"

	"easi/backend/internal/metamodel/application/readmodels"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/types"
)

type AttributeOptionDTO struct {
	ID     string      `json:"id"`
	Label  string      `json:"label"`
	Active bool        `json:"active"`
	Links  types.Links `json:"_links,omitempty"`
}

type SubjectAttributeDTO struct {
	ID       string               `json:"id"`
	Name     string               `json:"name"`
	Type     string               `json:"type"`
	HelpText string               `json:"helpText"`
	Active   bool                 `json:"active"`
	Options  []AttributeOptionDTO `json:"options,omitempty"`
	Min      *float64             `json:"min,omitempty"`
	Max      *float64             `json:"max,omitempty"`
	Links    types.Links          `json:"_links,omitempty"`
}

type SubjectAttributeSchemaDTO struct {
	ID          string                `json:"id"`
	SubjectType string                `json:"subjectType"`
	Attributes  []SubjectAttributeDTO `json:"attributes"`
	Version     int                   `json:"version"`
	CreatedAt   time.Time             `json:"createdAt"`
	ModifiedAt  time.Time             `json:"modifiedAt"`
	ModifiedBy  string                `json:"modifiedBy"`
	Links       types.Links           `json:"_links,omitempty"`
}

func BuildSubjectAttributeSchemaDTO(record *readmodels.SubjectAttributeSchemaRecord, links *MetaModelLinks, actor sharedctx.Actor) SubjectAttributeSchemaDTO {
	ctx := schemaLinkContext{subjectType: record.SubjectType, actor: actor}
	attributes := make([]SubjectAttributeDTO, len(record.Attributes))
	for i, attribute := range record.Attributes {
		attributes[i] = buildSubjectAttributeDTO(attribute, links, ctx)
	}
	return SubjectAttributeSchemaDTO{
		ID:          record.ID,
		SubjectType: record.SubjectType,
		Attributes:  attributes,
		Version:     record.Version,
		CreatedAt:   record.CreatedAt,
		ModifiedAt:  record.ModifiedAt,
		ModifiedBy:  record.ModifiedBy,
		Links:       links.SubjectAttributeSchemaLinks(ctx),
	}
}

func buildSubjectAttributeDTO(attribute readmodels.SubjectAttributeRecord, links *MetaModelLinks, ctx schemaLinkContext) SubjectAttributeDTO {
	dto := SubjectAttributeDTO{
		ID:       attribute.ID,
		Name:     attribute.Name,
		Type:     attribute.Type,
		HelpText: attribute.HelpText,
		Active:   attribute.Active,
		Options:  buildAttributeOptionDTOs(attribute, links, ctx),
		Min:      attribute.Min,
		Max:      attribute.Max,
	}
	dto.Links = links.subjectAttributeLinks(ctx, dto)
	return dto
}

func buildAttributeOptionDTOs(attribute readmodels.SubjectAttributeRecord, links *MetaModelLinks, ctx schemaLinkContext) []AttributeOptionDTO {
	if len(attribute.Options) == 0 {
		return nil
	}
	dtos := make([]AttributeOptionDTO, len(attribute.Options))
	for i, option := range attribute.Options {
		dto := AttributeOptionDTO{ID: option.ID, Label: option.Label, Active: option.Active}
		dto.Links = links.attributeOptionLinks(ctx, attribute, dto)
		dtos[i] = dto
	}
	return dtos
}
