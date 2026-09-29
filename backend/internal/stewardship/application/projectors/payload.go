package projectors

import (
	"encoding/json"
	"fmt"

	domain "easi/backend/internal/shared/eventsourcing"
)

func decodePayload[T any](event domain.DomainEvent) (T, error) {
	var payload T
	data, err := json.Marshal(event.EventData())
	if err != nil {
		return payload, fmt.Errorf("marshal %s payload of %s: %w", event.EventType(), event.AggregateID(), err)
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return payload, fmt.Errorf("unmarshal %s payload of %s: %w", event.EventType(), event.AggregateID(), err)
	}
	return payload, nil
}

type idPayload struct {
	ID string `json:"id"`
}

func subjectID(event domain.DomainEvent, payloadID string) string {
	if payloadID != "" {
		return payloadID
	}
	return event.AggregateID()
}
