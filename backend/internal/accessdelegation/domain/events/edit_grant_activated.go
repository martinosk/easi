package events

import (
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
)

type EditGrantActivated struct {
	domain.BaseEvent
	ID           string    `json:"id"`
	ArtifactType string    `json:"artifactType"`
	ArtifactID   string    `json:"artifactId"`
	GrantorID    string    `json:"grantorId"`
	GrantorEmail string    `json:"grantorEmail"`
	GranteeEmail string    `json:"granteeEmail"`
	Scope        string    `json:"scope"`
	Reason       string    `json:"reason"`
	CreatedAt    time.Time `json:"createdAt"`
	ExpiresAt    time.Time `json:"expiresAt"`
}

func (e EditGrantActivated) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}

func (e EditGrantActivated) EventType() string {
	return "EditGrantActivated"
}

func (e EditGrantActivated) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":           e.ID,
		"artifactType": e.ArtifactType,
		"artifactId":   e.ArtifactID,
		"grantorId":    e.GrantorID,
		"grantorEmail": e.GrantorEmail,
		"granteeEmail": e.GranteeEmail,
		"scope":        e.Scope,
		"reason":       e.Reason,
		"createdAt":    e.CreatedAt.Format(time.RFC3339),
		"expiresAt":    e.ExpiresAt.Format(time.RFC3339),
	}
}
