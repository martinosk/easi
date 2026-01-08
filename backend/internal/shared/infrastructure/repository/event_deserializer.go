package repository

import (
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
)

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

func GetInt(data map[string]interface{}, key string) int {
	if v, ok := data[key].(int); ok {
		return v
	}
	if v, ok := data[key].(float64); ok {
		return int(v)
	}
	return 0
}

func GetString(data map[string]interface{}, key string) string {
	if v, ok := data[key].(string); ok {
		return v
	}
	return ""
}

func GetTime(data map[string]interface{}, key string) time.Time {
	if str, ok := data[key].(string); ok {
		t, _ := time.Parse(time.RFC3339Nano, str)
		return t
	}
	return time.Time{}
}

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

func GetMap(data map[string]interface{}, key string) map[string]interface{} {
	if m, ok := data[key].(map[string]interface{}); ok {
		return m
	}
	return nil
}
