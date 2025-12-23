package repository

import "easi/backend/internal/shared/eventsourcing"

type EventDeserializerFunc func(data map[string]interface{}) domain.DomainEvent

type EventDeserializers map[string]EventDeserializerFunc

func (d EventDeserializers) Deserialize(storedEvents []domain.DomainEvent) []domain.DomainEvent {
	domainEvents := make([]domain.DomainEvent, 0, len(storedEvents))

	for _, event := range storedEvents {
		deserializer, exists := d[event.EventType()]
		if !exists {
			continue
		}
		domainEvents = append(domainEvents, deserializer(event.EventData()))
	}

	return domainEvents
}
