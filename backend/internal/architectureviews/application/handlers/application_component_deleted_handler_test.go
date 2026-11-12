package handlers

import (
	"context"
	"errors"
	"testing"
	"time"

	"easi/backend/internal/architectureviews/application/commands"
	"easi/backend/internal/shared/cqrs"
	domain "easi/backend/internal/shared/eventsourcing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubViewsContainingComponent struct {
	viewIDs     []string
	err         error
	componentID string
}

func (s *stubViewsContainingComponent) GetViewsContainingComponent(_ context.Context, componentID string) ([]string, error) {
	s.componentID = componentID
	return s.viewIDs, s.err
}

type recordingCommandBus struct {
	dispatched []cqrs.Command
	failOn     string
	err        error
}

func (b *recordingCommandBus) Dispatch(_ context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	b.dispatched = append(b.dispatched, cmd)
	if removal, ok := cmd.(*commands.RemoveComponentFromView); ok && removal.ViewID == b.failOn {
		return cqrs.EmptyResult(), b.err
	}
	return cqrs.EmptyResult(), nil
}

func (b *recordingCommandBus) Register(string, cqrs.CommandHandler) {}

func deletionEvent(eventType, aggregateID string) domain.DomainEvent {
	return domain.NewGenericDomainEvent(aggregateID, eventType, []byte(`{}`), time.Now().UTC())
}

const deletedApplicationID = "billing"

func billingDeleted() domain.DomainEvent {
	return deletionEvent("ApplicationComponentDeleted", deletedApplicationID)
}

func TestApplicationComponentDeleted_RemovesApplicationFromEveryViewContainingIt(t *testing.T) {
	query := &stubViewsContainingComponent{viewIDs: []string{"view-public", "view-private"}}
	bus := &recordingCommandBus{}
	handler := NewApplicationComponentDeletedHandler(bus, query)

	err := handler.Handle(context.Background(), billingDeleted())

	require.NoError(t, err)
	assert.Equal(t, "billing", query.componentID)
	assert.Equal(t, []cqrs.Command{
		&commands.RemoveComponentFromView{ViewID: "view-public", ComponentID: "billing"},
		&commands.RemoveComponentFromView{ViewID: "view-private", ComponentID: "billing"},
	}, bus.dispatched)
}

func TestApplicationComponentDeleted_ApplicationInNoViewDispatchesNothing(t *testing.T) {
	bus := &recordingCommandBus{}
	handler := NewApplicationComponentDeletedHandler(bus, &stubViewsContainingComponent{})

	err := handler.Handle(context.Background(), billingDeleted())

	require.NoError(t, err)
	assert.Empty(t, bus.dispatched)
}

func TestApplicationComponentDeleted_FailingViewLookupFailsTheDeletion(t *testing.T) {
	lookupErr := errors.New("database unavailable")
	bus := &recordingCommandBus{}
	handler := NewApplicationComponentDeletedHandler(bus, &stubViewsContainingComponent{err: lookupErr})

	err := handler.Handle(context.Background(), billingDeleted())

	assert.ErrorIs(t, err, lookupErr)
	assert.Empty(t, bus.dispatched)
}

func TestApplicationComponentDeleted_FailingRemovalFailsTheDeletion(t *testing.T) {
	removalErr := errors.New("view has been deleted")
	query := &stubViewsContainingComponent{viewIDs: []string{"view-a", "view-b", "view-c"}}
	bus := &recordingCommandBus{failOn: "view-b", err: removalErr}
	handler := NewApplicationComponentDeletedHandler(bus, query)

	err := handler.Handle(context.Background(), billingDeleted())

	assert.ErrorIs(t, err, removalErr)
	assert.Len(t, bus.dispatched, 2, "cleanup stops at the first failure instead of logging and skipping it")
}
