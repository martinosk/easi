package projectors

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"easi/backend/internal/accessdelegation/application/commands"
	capPL "easi/backend/internal/capabilitymapping/publishedlanguage"
	"easi/backend/internal/shared/cqrs"
)

type stubActiveGrantReader struct {
	grantIDs map[string][]string
	err      error
}

func (s *stubActiveGrantReader) GetActiveGrantIDsForArtifact(_ context.Context, artifactType, artifactID string) ([]string, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.grantIDs[artifactType+":"+artifactID], nil
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

func deletionEvent(aggregateID string, data map[string]interface{}) supplierEvent {
	return supplierEvent{aggregateID: aggregateID, eventType: capPL.CapabilityDeleted, data: data}
}

func TestArtifactDeletionProjector_RevokesEveryActiveGrantOnTheDeletedArtifact(t *testing.T) {
	reader := &stubActiveGrantReader{grantIDs: map[string][]string{"capability:cap-123": {"grant-1", "grant-2"}}}
	bus := &spyCommandBus{}
	projector := NewArtifactDeletionProjector(reader, bus, CapabilityArtifact)

	err := projector.Handle(context.Background(), deletionEvent("cap-123", map[string]interface{}{"id": "cap-123"}))

	require.NoError(t, err)
	require.Len(t, bus.dispatched, 2)
	revoked := make([]string, 0, 2)
	for _, cmd := range bus.dispatched {
		revoke, ok := cmd.(*commands.RevokeEditGrant)
		require.True(t, ok)
		assert.Equal(t, "system:artifact-deleted", revoke.RevokedBy)
		revoked = append(revoked, revoke.ID)
	}
	assert.Equal(t, []string{"grant-1", "grant-2"}, revoked)
}

func TestArtifactDeletionProjector_NoActiveGrants_DispatchesNothing(t *testing.T) {
	bus := &spyCommandBus{}
	projector := NewArtifactDeletionProjector(&stubActiveGrantReader{}, bus, CapabilityArtifact)

	err := projector.Handle(context.Background(), deletionEvent("cap-123", map[string]interface{}{"id": "cap-123"}))

	require.NoError(t, err)
	assert.Empty(t, bus.dispatched)
}

func TestArtifactDeletionProjector_ReadModelFailure_ReturnsError(t *testing.T) {
	projector := NewArtifactDeletionProjector(&stubActiveGrantReader{err: errors.New("database error")}, &spyCommandBus{}, CapabilityArtifact)

	err := projector.Handle(context.Background(), deletionEvent("cap-123", map[string]interface{}{"id": "cap-123"}))

	assert.ErrorContains(t, err, "cap-123")
}

func TestArtifactDeletionProjector_FallsBackToAggregateID(t *testing.T) {
	reader := &stubActiveGrantReader{grantIDs: map[string][]string{"capability:agg-456": {"grant-1"}}}
	bus := &spyCommandBus{}
	projector := NewArtifactDeletionProjector(reader, bus, CapabilityArtifact)

	err := projector.Handle(context.Background(), deletionEvent("agg-456", map[string]interface{}{}))

	require.NoError(t, err)
	assert.Len(t, bus.dispatched, 1)
}

func TestArtifactDeletionProjector_RevocationFailure_ReturnsError(t *testing.T) {
	reader := &stubActiveGrantReader{grantIDs: map[string][]string{"capability:cap-123": {"grant-1"}}}
	bus := &spyCommandBus{err: errors.New("command failed")}
	projector := NewArtifactDeletionProjector(reader, bus, CapabilityArtifact)

	err := projector.Handle(context.Background(), deletionEvent("cap-123", map[string]interface{}{"id": "cap-123"}))

	assert.ErrorContains(t, err, "grant-1")
}

func TestArtifactDeletionProjector_ScopesRevocationToItsArtifactType(t *testing.T) {
	reader := &stubActiveGrantReader{grantIDs: map[string][]string{"capability:comp-789": {"grant-5"}}}
	bus := &spyCommandBus{}
	projector := NewArtifactDeletionProjector(reader, bus, ComponentArtifact)

	err := projector.Handle(context.Background(), deletionEvent("comp-789", map[string]interface{}{"id": "comp-789"}))

	require.NoError(t, err)
	assert.Empty(t, bus.dispatched)
}
