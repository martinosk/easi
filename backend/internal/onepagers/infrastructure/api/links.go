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
	links := sharedAPI.Links{"self": l.Get(configurationPath(ctx.subjectType))}
	if ctx.canWrite() {
		links["x-define-custom-field"] = l.Post(configurationPath(ctx.subjectType) + "/custom-fields")
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
	if field.Included {
		return sharedAPI.Links{"x-exclude": l.Post(base + "/exclude")}
	}
	return sharedAPI.Links{"x-include": l.Post(base + "/include")}
}

func (l *OnePagerLinks) customFieldLinks(ctx linkContext, field CustomFieldDTO) sharedAPI.Links {
	if !ctx.canWrite() {
		return nil
	}
	base := customFieldPath(ctx.subjectType, field.ID)
	if !field.Active {
		return sharedAPI.Links{"x-reactivate": l.Post(base + "/reactivate")}
	}
	links := sharedAPI.Links{
		"x-rename":          l.Put(base),
		"x-set-requirement": l.Put(base + "/requirement"),
		"x-retire":          l.Post(base + "/retire"),
	}
	if field.Type == "selection" {
		links["x-add-option"] = l.Post(base + "/options")
	}
	return links
}

type optionLinkParams struct {
	linkContext
	fieldID     string
	option      SelectionOptionDTO
	fieldActive bool
}

func (p optionLinkParams) canRetireOption() bool {
	return p.fieldActive && p.option.Active && p.canWrite()
}

func (l *OnePagerLinks) optionLinks(params optionLinkParams) sharedAPI.Links {
	if !params.canRetireOption() {
		return nil
	}
	return sharedAPI.Links{
		"x-retire": l.Post(customFieldPath(params.subjectType, params.fieldID) + "/options/" + params.option.ID + "/retire"),
	}
}
