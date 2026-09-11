package api

import (
	"easi/backend/internal/metamodel/application/readmodels"
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
)

const metaModelWritePermission = "metamodel:write"

type schemaLinkContext struct {
	subjectType string
	actor       sharedctx.Actor
}

func (c schemaLinkContext) canWrite() bool {
	return c.actor.HasPermission(metaModelWritePermission)
}

func subjectAttributesPath(subjectType string) string {
	return "/meta-model/subject-types/" + subjectType + "/attributes"
}

func subjectAttributePath(subjectType, attributeID string) string {
	return subjectAttributesPath(subjectType) + "/" + attributeID
}

func (h *MetaModelLinks) SubjectAttributeSchemaLinks(ctx schemaLinkContext) sharedAPI.Links {
	links := sharedAPI.Links{"self": h.Get(subjectAttributesPath(ctx.subjectType))}
	if ctx.canWrite() {
		links["x-define"] = h.Post(subjectAttributesPath(ctx.subjectType))
	}
	return links
}

func (h *MetaModelLinks) subjectAttributeLinks(ctx schemaLinkContext, attribute SubjectAttributeDTO) sharedAPI.Links {
	if !ctx.canWrite() {
		return nil
	}
	base := subjectAttributePath(ctx.subjectType, attribute.ID)
	if !attribute.Active {
		return sharedAPI.Links{"x-reactivate": h.Post(base + "/reactivate")}
	}
	links := sharedAPI.Links{
		"x-rename": h.Put(base),
		"x-retire": h.Post(base + "/retire"),
	}
	if attribute.Type == "selection" {
		links["x-add-option"] = h.Post(base + "/options")
	}
	if attribute.Type == "number" {
		links["x-set-bounds"] = h.Put(base + "/bounds")
	}
	return links
}

func (h *MetaModelLinks) attributeOptionLinks(ctx schemaLinkContext, attribute readmodels.SubjectAttributeRecord, option AttributeOptionDTO) sharedAPI.Links {
	if !canRetireOption(ctx, attribute, option) {
		return nil
	}
	return sharedAPI.Links{
		"x-retire": h.Post(subjectAttributePath(ctx.subjectType, attribute.ID) + "/options/" + option.ID + "/retire"),
	}
}

func canRetireOption(ctx schemaLinkContext, attribute readmodels.SubjectAttributeRecord, option AttributeOptionDTO) bool {
	return ctx.canWrite() && attribute.Active && option.Active
}
