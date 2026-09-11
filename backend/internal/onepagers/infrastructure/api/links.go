package api

import (
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
)

const metaModelWritePermission = "metamodel:write"

type OnePagerLinks struct {
	*sharedAPI.HATEOASLinks
}

func NewOnePagerLinks(h *sharedAPI.HATEOASLinks) *OnePagerLinks {
	return &OnePagerLinks{HATEOASLinks: h}
}

func configurationPath(subjectType string) string {
	return "/one-pagers/configurations/" + subjectType
}

func customFieldPath(subjectType, fieldID string) string {
	return configurationPath(subjectType) + "/custom-fields/" + fieldID
}

type linkContext struct {
	subjectType string
	actor       sharedctx.Actor
}

func (c linkContext) canWrite() bool {
	return c.actor.HasPermission(metaModelWritePermission)
}

func (l *OnePagerLinks) ConfigurationLinks(ctx linkContext) sharedAPI.Links {
	links := sharedAPI.Links{
		"self":               l.Get(configurationPath(ctx.subjectType)),
		"x-attribute-schema": l.Get("/meta-model/subject-types/" + ctx.subjectType + "/attributes"),
	}
	if ctx.canWrite() {
		links["x-reorder"] = l.Put(configurationPath(ctx.subjectType) + "/display-order")
		links["x-impact-preview"] = l.Get(configurationPath(ctx.subjectType) + "/impact-preview")
	}
	return links
}

func (l *OnePagerLinks) builtInFieldLinks(ctx linkContext, field BuiltInFieldDTO) sharedAPI.Links {
	if !ctx.canWrite() {
		return nil
	}
	base := configurationPath(ctx.subjectType) + "/built-in-fields/" + field.ID
	if !field.Included {
		return sharedAPI.Links{"x-include": l.Post(base + "/include")}
	}
	return sharedAPI.Links{
		"x-exclude":         l.Post(base + "/exclude"),
		"x-set-requirement": l.Put(base + "/requirement"),
	}
}

func (l *OnePagerLinks) customFieldLinks(ctx linkContext, field CustomFieldDTO) sharedAPI.Links {
	if !canSetRequirement(ctx, field) {
		return nil
	}
	return sharedAPI.Links{
		"x-set-requirement": l.Put(customFieldPath(ctx.subjectType, field.ID) + "/requirement"),
	}
}

func canSetRequirement(ctx linkContext, field CustomFieldDTO) bool {
	return ctx.canWrite() && field.Included && field.Active
}
