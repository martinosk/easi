package handlers

import (
	"context"
	"testing"
	"time"

	"easi/backend/internal/shared/eventsourcing"

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

func (e mockEvent) AggregateID() string                  { return e.aggregateID }
func (e mockEvent) EventType() string                    { return e.eventType }
func (e mockEvent) OccurredAt() time.Time                { return time.Now() }
func (e mockEvent) EventData() map[string]interface{}   { return e.eventData }

var _ domain.DomainEvent = mockEvent{}

func TestMaturityScaleConfigUpdatedHandler_InvalidatesCache(t *testing.T) {
	cacheInvalidator := &mockCacheInvalidator{}
	handler := NewMaturityScaleConfigUpdatedHandler(cacheInvalidator)

	event := mockEvent{
		aggregateID: "config-123",
		eventType:   "MaturityScaleConfigUpdated",
		eventData: map[string]interface{}{
			"tenantId": "tenant-456",
		},
	}

	err := handler.Handle(context.Background(), event)

	assert.NoError(t, err)
	assert.Equal(t, []string{"tenant-456"}, cacheInvalidator.invalidatedTenants)
}

func TestMaturityScaleConfigUpdatedHandler_MissingTenantID_NoError(t *testing.T) {
	cacheInvalidator := &mockCacheInvalidator{}
	handler := NewMaturityScaleConfigUpdatedHandler(cacheInvalidator)

	event := mockEvent{
		aggregateID: "config-123",
		eventType:   "MaturityScaleConfigUpdated",
		eventData:   map[string]interface{}{},
	}

	err := handler.Handle(context.Background(), event)

	assert.NoError(t, err)
	assert.Empty(t, cacheInvalidator.invalidatedTenants)
}

func TestMaturityScaleConfigUpdatedHandler_InvalidTenantIDType_NoError(t *testing.T) {
	cacheInvalidator := &mockCacheInvalidator{}
	handler := NewMaturityScaleConfigUpdatedHandler(cacheInvalidator)

	event := mockEvent{
		aggregateID: "config-123",
		eventType:   "MaturityScaleConfigUpdated",
		eventData: map[string]interface{}{
			"tenantId": 12345,
		},
	}

	err := handler.Handle(context.Background(), event)

	assert.NoError(t, err)
	assert.Empty(t, cacheInvalidator.invalidatedTenants)
}
