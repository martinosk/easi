package repository

import (
	"encoding/json"
	"log/slog"

	domain "easi/backend/internal/shared/eventsourcing"
)

func JSONDeserializer[T domain.DomainEvent](data map[string]interface{}) (domain.DomainEvent, error) {
	var event T
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(jsonBytes, &event); err != nil {
		return nil, err
	}
	return event, nil
}

type EventDeserializerFunc func(data map[string]interface{}) (domain.DomainEvent, error)

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

func (d EventDeserializers) Deserialize(storedEvents []domain.DomainEvent) ([]domain.DomainEvent, error) {
	domainEvents := make([]domain.DomainEvent, 0, len(storedEvents))

	for i, event := range storedEvents {
		sequenceNumber := i + 1

		eventData := event.EventData()
		if len(d.upcasters) > 0 {
			eventData = d.upcasters.Upcast(event.EventType(), eventData)
		}

		deserializer, exists := d.deserializers[event.EventType()]
		if !exists {
			slog.Warn("unknown event type skipped for forward compatibility",
				"aggregateId", event.AggregateID(),
				"eventType", event.EventType(),
				"sequenceNumber", sequenceNumber,
			)
			continue
		}

		domainEvent, err := deserializer(eventData)
		if err != nil {
			return nil, NewDeserializationError(
				event.AggregateID(),
				event.EventType(),
				sequenceNumber,
				err,
			)
		}
		domainEvents = append(domainEvents, domainEvent)
	}

	return domainEvents, nil
}
