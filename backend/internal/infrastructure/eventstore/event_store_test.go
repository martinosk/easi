package eventstore

import (
	"testing"

	"github.com/easi/backend/internal/shared/domain"
)

// MockEvent implements DomainEvent for testing
type MockEvent struct {
	domain.BaseEvent
	eventType string
	data      map[string]interface{}
}

func (e MockEvent) EventType() string {
	return e.eventType
}

func (e MockEvent) EventData() map[string]interface{} {
	return e.data
}

func NewMockEvent(aggregateID, eventType string, data map[string]interface{}) MockEvent {
	return MockEvent{
		BaseEvent: domain.NewBaseEvent(aggregateID),
		eventType: eventType,
		data:      data,
	}
}

func TestEventStore_InitializeSchema(t *testing.T) {
	// This test requires a running PostgreSQL instance
	// It's marked as a integration test
	t.Skip("Integration test - requires database")
}

func TestEventStore_SaveAndGetEvents(t *testing.T) {
	// This test requires a running PostgreSQL instance
	// It's marked as a integration test
	t.Skip("Integration test - requires database")
}

func TestEventStore_ConcurrencyConflict(t *testing.T) {
	// This test requires a running PostgreSQL instance
	// It's marked as a integration test
	t.Skip("Integration test - requires database")
}
