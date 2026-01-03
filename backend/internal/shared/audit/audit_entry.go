package audit

import (
	"regexp"
	"strings"
	"time"
)

type AuditEntry struct {
	EventID     int64                  `json:"eventId"`
	AggregateID string                 `json:"aggregateId"`
	EventType   string                 `json:"eventType"`
	DisplayName string                 `json:"displayName"`
	EventData   map[string]interface{} `json:"eventData"`
	OccurredAt  time.Time              `json:"occurredAt"`
	Version     int                    `json:"version"`
	ActorID     string                 `json:"actorId"`
	ActorEmail  string                 `json:"actorEmail"`
}

type AuditHistoryResponse struct {
	Entries    []AuditEntry      `json:"entries"`
	Pagination *PaginationInfo   `json:"pagination,omitempty"`
	Links      map[string]string `json:"_links"`
}

type PaginationInfo struct {
	HasMore    bool   `json:"hasMore"`
	NextCursor string `json:"nextCursor,omitempty"`
}

var camelCaseRegex = regexp.MustCompile("([a-z])([A-Z])")

func FormatEventTypeDisplayName(eventType string) string {
	parts := strings.Split(eventType, ".")
	action := parts[len(parts)-1]

	spaced := camelCaseRegex.ReplaceAllString(action, "${1} ${2}")

	return strings.Title(strings.ToLower(spaced))
}
