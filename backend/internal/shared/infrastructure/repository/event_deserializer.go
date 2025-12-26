package repository

import "easi/backend/internal/shared/eventsourcing"

type EventDeserializerFunc func(data map[string]interface{}) domain.DomainEvent

type EventDeserializers struct {
	deserializers map[string]EventDeserializerFunc
	upcasters     domain.UpcasterChain
}

func NewEventDeserializers(deserializers map[string]EventDeserializerFunc, upcasters ...domain.Upcaster) EventDeserializers {
	return EventDeserializers{
		deserializers: deserializers,
		upcasters:     upcasters,
	}
}

func (d EventDeserializers) Deserialize(storedEvents []domain.DomainEvent) []domain.DomainEvent {
	domainEvents := make([]domain.DomainEvent, 0, len(storedEvents))

	for _, event := range storedEvents {
		deserializer, exists := d.deserializers[event.EventType()]
		if !exists {
			continue
		}

		eventData := event.EventData()
		if len(d.upcasters) > 0 {
			eventData = d.upcasters.Upcast(event.EventType(), eventData)
		}

		domainEvents = append(domainEvents, deserializer(eventData))
	}

	return domainEvents
}
