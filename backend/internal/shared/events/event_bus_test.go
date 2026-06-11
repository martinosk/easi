package events

import (
	"context"
	"errors"
	"testing"
	"time"

	domain "easi/backend/internal/shared/eventsourcing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubEvent struct{ eventType string }

func (e stubEvent) EventType() string                 { return e.eventType }
func (e stubEvent) AggregateID() string               { return "agg-1" }
func (e stubEvent) OccurredAt() time.Time             { return time.Time{} }
func (e stubEvent) EventData() map[string]interface{} { return map[string]interface{}{} }

type recordingHandler struct {
	received []string
	err      error
}

func (h *recordingHandler) Handle(_ context.Context, event domain.DomainEvent) error {
	h.received = append(h.received, event.EventType())
	return h.err
}

func TestPublish_DeliversToAllHandlers(t *testing.T) {
	bus := NewInMemoryEventBus()
	first := &recordingHandler{}
	second := &recordingHandler{}
	bus.Subscribe("ThingHappened", first)
	bus.Subscribe("ThingHappened", second)

	err := bus.Publish(context.Background(), []domain.DomainEvent{stubEvent{"ThingHappened"}})

	require.NoError(t, err)
	assert.Equal(t, []string{"ThingHappened"}, first.received)
	assert.Equal(t, []string{"ThingHappened"}, second.received)
}

func TestPublish_FailingHandlerDoesNotStarveLaterHandlers(t *testing.T) {
	bus := NewInMemoryEventBus()
	failing := &recordingHandler{err: errors.New("projection broke")}
	downstream := &recordingHandler{}
	bus.Subscribe("ThingHappened", failing)
	bus.Subscribe("ThingHappened", downstream)

	err := bus.Publish(context.Background(), []domain.DomainEvent{stubEvent{"ThingHappened"}})

	assert.Error(t, err, "the failure is still reported to the publisher")
	assert.Equal(t, []string{"ThingHappened"}, downstream.received,
		"projections are independent read models; one failing subscriber must not sever the rest of the chain")
}

func TestPublish_AggregatesAllHandlerFailures(t *testing.T) {
	bus := NewInMemoryEventBus()
	firstErr := errors.New("first projection broke")
	secondErr := errors.New("second projection broke")
	bus.Subscribe("ThingHappened", &recordingHandler{err: firstErr})
	bus.Subscribe("ThingHappened", &recordingHandler{err: secondErr})

	err := bus.Publish(context.Background(), []domain.DomainEvent{stubEvent{"ThingHappened"}})

	require.Error(t, err)
	assert.ErrorIs(t, err, firstErr)
	assert.ErrorIs(t, err, secondErr)
}

func TestPublish_GlobalHandlerFailureDoesNotStarveTypedHandlers(t *testing.T) {
	bus := NewInMemoryEventBus()
	failingGlobal := &recordingHandler{err: errors.New("global handler broke")}
	typed := &recordingHandler{}
	bus.SubscribeAll(failingGlobal)
	bus.Subscribe("ThingHappened", typed)

	err := bus.Publish(context.Background(), []domain.DomainEvent{stubEvent{"ThingHappened"}})

	assert.Error(t, err)
	assert.Equal(t, []string{"ThingHappened"}, typed.received)
}
