package api

import (
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
)

var subjectDetailPathPrefixes = map[string]string{
	"capability":      "/capabilities/",
	"application":     "/components/",
	"acquired-entity": "/acquired-entities/",
	"vendor":          "/vendors/",
	"internal-team":   "/internal-teams/",
}

func onePagerViewPath(subjectType, subjectID string) string {
	return "/one-pagers/" + subjectType + "/" + subjectID
}

func subjectDetailPath(subjectType, subjectID string) string {
	return subjectDetailPathPrefixes[subjectType] + subjectID
}

func (l *OnePagerLinks) viewLinks(subjectType, subjectID string, actor sharedctx.Actor) sharedAPI.Links {
	links := sharedAPI.Links{
		"self":      l.Get(onePagerViewPath(subjectType, subjectID)),
		"x-subject": l.Get(subjectDetailPath(subjectType, subjectID)),
	}
	factsCtx := factsLinkContext{subjectType: subjectType, subjectID: subjectID, actor: actor}
	if factsCtx.canWrite() {
		links["x-record"] = l.Put(factsPath(subjectType, subjectID))
	}
	return links
}
