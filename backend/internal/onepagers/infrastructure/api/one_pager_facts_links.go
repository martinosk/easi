package api

import (
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
)

var subjectWritePermissions = map[string]string{
	"capability":      "capabilities:write",
	"application":     "components:write",
	"acquired-entity": "components:write",
	"vendor":          "components:write",
	"internal-team":   "components:write",
}

type factsLinkContext struct {
	subjectType string
	subjectID   string
	actor       sharedctx.Actor
}

func (c factsLinkContext) canWrite() bool {
	permission, found := subjectWritePermissions[c.subjectType]
	return found && c.actor.HasPermission(permission)
}

func factsPath(subjectType, subjectID string) string {
	return "/one-pagers/" + subjectType + "/" + subjectID + "/facts"
}

func fieldValuePath(subjectType, subjectID, fieldID string) string {
	return factsPath(subjectType, subjectID) + "/" + fieldID
}

func (l *OnePagerLinks) factsLinks(ctx factsLinkContext) sharedAPI.Links {
	links := sharedAPI.Links{"self": l.Get(factsPath(ctx.subjectType, ctx.subjectID))}
	if ctx.canWrite() {
		links["x-record"] = l.Put(factsPath(ctx.subjectType, ctx.subjectID))
	}
	return links
}

func (l *OnePagerLinks) fieldValueLinks(ctx factsLinkContext, fieldID string) sharedAPI.Links {
	if !ctx.canWrite() {
		return nil
	}
	path := fieldValuePath(ctx.subjectType, ctx.subjectID, fieldID)
	return sharedAPI.Links{
		"x-record": l.Put(path),
		"x-clear":  l.Del(path),
	}
}
