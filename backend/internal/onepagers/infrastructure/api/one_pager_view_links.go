package api

import (
	sharedAPI "easi/backend/internal/shared/api"
)

var subjectDetailPathPrefixes = map[string]string{
	"capability":            "/capabilities/",
	"enterprise-capability": "/enterprise-capabilities/",
	"application":           "/components/",
	"acquired-entity":       "/acquired-entities/",
	"vendor":                "/vendors/",
	"internal-team":         "/internal-teams/",
}

func onePagerViewPath(subjectType, subjectID string) string {
	return "/one-pagers/" + subjectType + "/" + subjectID
}

func subjectDetailPath(subjectType, subjectID string) string {
	return subjectDetailPathPrefixes[subjectType] + subjectID
}

func (l *OnePagerLinks) viewLinks(subjectType, subjectID string) sharedAPI.Links {
	return sharedAPI.Links{
		"self":      l.Get(onePagerViewPath(subjectType, subjectID)),
		"x-subject": l.Get(subjectDetailPath(subjectType, subjectID)),
	}
}
