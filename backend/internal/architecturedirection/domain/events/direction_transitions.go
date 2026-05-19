package events

import (
	"time"

	pl "easi/backend/internal/architecturedirection/publishedlanguage"
	domain "easi/backend/internal/shared/eventsourcing"
)

type DirectionProposed struct {
	domain.BaseEvent
	ID         string    `json:"id"`
	OccurredOn time.Time `json:"occurredOn"`
}

func NewDirectionProposed(id string) DirectionProposed {
	return DirectionProposed{
		BaseEvent:  domain.NewBaseEvent(id),
		ID:         id,
		OccurredOn: time.Now().UTC(),
	}
}
func (e DirectionProposed) EventType() string { return pl.DirectionProposed }
func (e DirectionProposed) EventData() map[string]interface{} {
	return map[string]interface{}{"id": e.ID, "occurredOn": e.OccurredOn}
}

type DirectionAgreed struct {
	domain.BaseEvent
	ID         string    `json:"id"`
	OccurredOn time.Time `json:"occurredOn"`
}

func NewDirectionAgreed(id string) DirectionAgreed {
	return DirectionAgreed{
		BaseEvent:  domain.NewBaseEvent(id),
		ID:         id,
		OccurredOn: time.Now().UTC(),
	}
}
func (e DirectionAgreed) EventType() string { return pl.DirectionAgreed }
func (e DirectionAgreed) EventData() map[string]interface{} {
	return map[string]interface{}{"id": e.ID, "occurredOn": e.OccurredOn}
}

type DirectionRejected struct {
	domain.BaseEvent
	ID         string    `json:"id"`
	OccurredOn time.Time `json:"occurredOn"`
}

func NewDirectionRejected(id string) DirectionRejected {
	return DirectionRejected{
		BaseEvent:  domain.NewBaseEvent(id),
		ID:         id,
		OccurredOn: time.Now().UTC(),
	}
}
func (e DirectionRejected) EventType() string { return pl.DirectionRejected }
func (e DirectionRejected) EventData() map[string]interface{} {
	return map[string]interface{}{"id": e.ID, "occurredOn": e.OccurredOn}
}
