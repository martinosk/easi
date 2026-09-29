package events

import (
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
	pl "easi/backend/internal/stewardship/publishedlanguage"
)

type StewardReleased struct {
	domain.BaseEvent
	ID         string    `json:"id"`
	DomainID   string    `json:"domainId"`
	Concern    string    `json:"concern"`
	ReleasedBy string    `json:"releasedBy"`
	ReleasedAt time.Time `json:"releasedAt"`
}

func NewStewardReleased(id, domainID, concern, releasedBy string) StewardReleased {
	return StewardReleased{
		BaseEvent:  domain.NewBaseEvent(id),
		ID:         id,
		DomainID:   domainID,
		Concern:    concern,
		ReleasedBy: releasedBy,
		ReleasedAt: time.Now().UTC(),
	}
}

func (e StewardReleased) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}

func (e StewardReleased) EventType() string { return pl.StewardReleased }

func (e StewardReleased) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":         e.ID,
		"domainId":   e.DomainID,
		"concern":    e.Concern,
		"releasedBy": e.ReleasedBy,
		"releasedAt": e.ReleasedAt.Format(time.RFC3339Nano),
	}
}
