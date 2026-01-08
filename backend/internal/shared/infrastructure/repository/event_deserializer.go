package repository

import (
	"log/slog"
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
)

// Deprecated: Use EventDeserializerFuncWithError instead
type EventDeserializerFunc func(data map[string]interface{}) domain.DomainEvent

type EventDeserializerFuncWithError func(data map[string]interface{}) (domain.DomainEvent, error)

type EventDeserializers struct {
	deserializers          map[string]EventDeserializerFunc
	deserializersWithError map[string]EventDeserializerFuncWithError
	upcasters              domain.UpcasterChain
}

// Deprecated: Use NewEventDeserializersWithError instead
func NewEventDeserializers(deserializers map[string]EventDeserializerFunc, upcasters ...domain.Upcaster) EventDeserializers {
	return EventDeserializers{
		deserializers:          deserializers,
		deserializersWithError: make(map[string]EventDeserializerFuncWithError),
		upcasters:              upcasters,
	}
}

func NewEventDeserializersWithError(deserializers map[string]EventDeserializerFuncWithError, upcasters ...domain.Upcaster) EventDeserializers {
	return EventDeserializers{
		deserializers:          make(map[string]EventDeserializerFunc),
		deserializersWithError: deserializers,
		upcasters:              upcasters,
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

		if deserializer, exists := d.deserializersWithError[event.EventType()]; exists {
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
			continue
		}

		if deserializer, exists := d.deserializers[event.EventType()]; exists {
			domainEvents = append(domainEvents, deserializer(eventData))
			continue
		}

		slog.Warn("unknown event type skipped for forward compatibility",
			"aggregateId", event.AggregateID(),
			"eventType", event.EventType(),
			"sequenceNumber", sequenceNumber,
		)
	}

	return domainEvents, nil
}

// Deprecated: Use GetRequiredInt or GetOptionalInt instead
func GetInt(data map[string]interface{}, key string) int {
	if v, ok := data[key].(int); ok {
		return v
	}
	if v, ok := data[key].(float64); ok {
		return int(v)
	}
	return 0
}

// Deprecated: Use GetRequiredString or GetOptionalString instead
func GetString(data map[string]interface{}, key string) string {
	if v, ok := data[key].(string); ok {
		return v
	}
	return ""
}

// Deprecated: Use GetRequiredTime or GetOptionalTime instead
func GetTime(data map[string]interface{}, key string) time.Time {
	if str, ok := data[key].(string); ok {
		t, _ := time.Parse(time.RFC3339Nano, str)
		return t
	}
	return time.Time{}
}

// Deprecated: Use GetRequiredMapSlice or GetOptionalMapSlice instead
func GetMapSlice(data map[string]interface{}, key string) []map[string]interface{} {
	raw, ok := data[key].([]interface{})
	if !ok {
		return nil
	}
	result := make([]map[string]interface{}, 0, len(raw))
	for _, item := range raw {
		if m, ok := item.(map[string]interface{}); ok {
			result = append(result, m)
		}
	}
	return result
}

// Deprecated: Use GetRequiredMap or GetOptionalMap instead
func GetMap(data map[string]interface{}, key string) map[string]interface{} {
	if m, ok := data[key].(map[string]interface{}); ok {
		return m
	}
	return nil
}
