package handlers

import (
	"context"
	domain "easi/backend/internal/shared/eventsourcing"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type mockCacheInvalidator struct {
	invalidatedTenants []string
}

func (m *mockCacheInvalidator) InvalidateCache(tenantID string) {
	m.invalidatedTenants = append(m.invalidatedTenants, tenantID)
}

type mockEvent struct {
	aggregateID string
	eventType   string
	eventData   map[string]interface{}
}

func (e mockEvent) AggregateID() string               { return e.aggregateID }
func (e mockEvent) EventType() string                 { return e.eventType }
func (e mockEvent) OccurredAt() time.Time             { return time.Now() }
func (e mockEvent) EventData() map[string]interface{} { return e.eventData }

var _ domain.DomainEvent = mockEvent{}

func TestMaturityScaleConfigUpdatedHandler(t *testing.T) {
	tests := []struct {
		name               string
		eventData          map[string]interface{}
		invalidatedTenants []string
	}{
		{
			name:               "invalidates cache",
			eventData:          map[string]interface{}{"tenantId": "tenant-456"},
			invalidatedTenants: []string{"tenant-456"},
		},
		{
			name:      "missing tenant id",
			eventData: map[string]interface{}{},
		},
		{
			name:      "invalid tenant id type",
			eventData: map[string]interface{}{"tenantId": 12345},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cacheInvalidator := &mockCacheInvalidator{}
			handler := NewMaturityScaleConfigUpdatedHandler(cacheInvalidator)

			event := mockEvent{
				aggregateID: "config-123",
				eventType:   "MaturityScaleConfigUpdated",
				eventData:   tt.eventData,
			}

			err := handler.Handle(context.Background(), event)

			assert.NoError(t, err)
			if tt.invalidatedTenants == nil {
				assert.Empty(t, cacheInvalidator.invalidatedTenants)
			} else {
				assert.Equal(t, tt.invalidatedTenants, cacheInvalidator.invalidatedTenants)
			}
		})
	}
}
