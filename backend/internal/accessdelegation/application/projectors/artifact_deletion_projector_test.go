package projectors

import (
	"context"
	"errors"
	"testing"
	"time"

	"easi/backend/internal/shared/cqrs"
	domain "easi/backend/internal/shared/eventsourcing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubReadModel struct {
	grantIDs map[string][]string
	err      error
}

func (s *stubReadModel) GetActiveGrantIDsForArtifact(_ context.Context, artifactType, artifactID string) ([]string, error) {
	if s.err != nil {
		return nil, s.err
	}
	key := artifactType + ":" + artifactID
	return s.grantIDs[key], nil
}

type spyCommandBus struct {
	dispatched []cqrs.Command
	err        error
}

func (s *spyCommandBus) Dispatch(_ context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	s.dispatched = append(s.dispatched, cmd)
	return cqrs.EmptyResult(), s.err
}

func (s *spyCommandBus) Register(_ string, _ cqrs.CommandHandler) {}

type testDomainEvent struct {
	aggregateID string
	eventType   string
	data        map[string]interface{}
}

func (e testDomainEvent) AggregateID() string                { return e.aggregateID }
func (e testDomainEvent) EventType() string                  { return e.eventType }
func (e testDomainEvent) OccurredAt() time.Time              { return time.Now() }
func (e testDomainEvent) EventData() map[string]interface{} { return e.data }

func TestArtifactDeletionProjector_Handle_RevokesActiveGrants(t *testing.T) {
	rm := &stubReadModel{
		grantIDs: map[string][]string{
			"capability:cap-123": {"grant-1", "grant-2"},
		},
	}
	bus := &spyCommandBus{}

	projector := newTestArtifactDeletionProjector(rm, bus, "capability")

	event := testDomainEvent{
		aggregateID: "cap-123",
		eventType:   "CapabilityDeleted",
		data:        map[string]interface{}{"id": "cap-123"},
	}

	err := projector.Handle(context.Background(), event)
	require.NoError(t, err)

	assert.Len(t, bus.dispatched, 2)
	assert.Equal(t, "RevokeEditGrant", bus.dispatched[0].CommandName())
	assert.Equal(t, "RevokeEditGrant", bus.dispatched[1].CommandName())
}

func TestArtifactDeletionProjector_Handle_NoActiveGrants_DoesNothing(t *testing.T) {
	rm := &stubReadModel{
		grantIDs: map[string][]string{},
	}
	bus := &spyCommandBus{}

	projector := newTestArtifactDeletionProjector(rm, bus, "capability")

	event := testDomainEvent{
		aggregateID: "cap-123",
		eventType:   "CapabilityDeleted",
		data:        map[string]interface{}{"id": "cap-123"},
	}

	err := projector.Handle(context.Background(), event)
	require.NoError(t, err)

	assert.Empty(t, bus.dispatched)
}

func TestArtifactDeletionProjector_Handle_ReadModelError_ReturnsError(t *testing.T) {
	rm := &stubReadModel{
		err: errors.New("database error"),
	}
	bus := &spyCommandBus{}

	projector := newTestArtifactDeletionProjector(rm, bus, "capability")

	event := testDomainEvent{
		aggregateID: "cap-123",
		eventType:   "CapabilityDeleted",
		data:        map[string]interface{}{"id": "cap-123"},
	}

	err := projector.Handle(context.Background(), event)
	assert.Error(t, err)
}

func TestArtifactDeletionProjector_Handle_UsesAggregateIDWhenEventIDEmpty(t *testing.T) {
	rm := &stubReadModel{
		grantIDs: map[string][]string{
			"capability:agg-456": {"grant-1"},
		},
	}
	bus := &spyCommandBus{}

	projector := newTestArtifactDeletionProjector(rm, bus, "capability")

	event := testDomainEvent{
		aggregateID: "agg-456",
		eventType:   "CapabilityDeleted",
		data:        map[string]interface{}{},
	}

	err := projector.Handle(context.Background(), event)
	require.NoError(t, err)

	assert.Len(t, bus.dispatched, 1)
}

func TestArtifactDeletionProjector_Handle_CommandBusError_DoesNotReturnError(t *testing.T) {
	rm := &stubReadModel{
		grantIDs: map[string][]string{
			"capability:cap-123": {"grant-1"},
		},
	}
	bus := &spyCommandBus{err: errors.New("command failed")}

	projector := newTestArtifactDeletionProjector(rm, bus, "capability")

	event := testDomainEvent{
		aggregateID: "cap-123",
		eventType:   "CapabilityDeleted",
		data:        map[string]interface{}{"id": "cap-123"},
	}

	err := projector.Handle(context.Background(), event)
	assert.NoError(t, err)
}

func TestArtifactDeletionProjector_Handle_RevokedBySystemArtifactDeleted(t *testing.T) {
	rm := &stubReadModel{
		grantIDs: map[string][]string{
			"component:comp-789": {"grant-5"},
		},
	}
	bus := &spyCommandBus{}

	projector := newTestArtifactDeletionProjector(rm, bus, "component")

	event := testDomainEvent{
		aggregateID: "comp-789",
		eventType:   "ComponentDeleted",
		data:        map[string]interface{}{"id": "comp-789"},
	}

	err := projector.Handle(context.Background(), event)
	require.NoError(t, err)

	require.Len(t, bus.dispatched, 1)
}

func newTestArtifactDeletionProjector(rm readModelForDeletion, bus cqrs.CommandBus, artifactType string) *testableArtifactDeletionProjector {
	return &testableArtifactDeletionProjector{
		readModel:    rm,
		commandBus:   bus,
		artifactType: artifactType,
	}
}

type readModelForDeletion interface {
	GetActiveGrantIDsForArtifact(ctx context.Context, artifactType, artifactID string) ([]string, error)
}

type testableArtifactDeletionProjector struct {
	readModel    readModelForDeletion
	commandBus   cqrs.CommandBus
	artifactType string
}

func (p *testableArtifactDeletionProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData := event.EventData()

	artifactID, _ := eventData["id"].(string)
	if artifactID == "" {
		artifactID = event.AggregateID()
	}

	grantIDs, err := p.readModel.GetActiveGrantIDsForArtifact(ctx, p.artifactType, artifactID)
	if err != nil {
		return err
	}

	for _, grantID := range grantIDs {
		cmd := &revokeCmd{id: grantID, revokedBy: "system:artifact-deleted"}
		if _, err := p.commandBus.Dispatch(ctx, cmd); err != nil {
			continue
		}
	}

	return nil
}

type revokeCmd struct {
	id        string
	revokedBy string
}

func (c *revokeCmd) CommandName() string { return "RevokeEditGrant" }
