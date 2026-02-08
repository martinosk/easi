package handlers

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"easi/backend/internal/accessdelegation/application/commands"
	"easi/backend/internal/accessdelegation/domain/aggregates"
	"easi/backend/internal/accessdelegation/domain/valueobjects"
	"easi/backend/internal/accessdelegation/infrastructure/repositories"
	domain "easi/backend/internal/shared/eventsourcing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type inMemoryEventStore struct {
	mu     sync.RWMutex
	events map[string][]domain.DomainEvent
}

func newInMemoryEventStore() *inMemoryEventStore {
	return &inMemoryEventStore{
		events: make(map[string][]domain.DomainEvent),
	}
}

func (s *inMemoryEventStore) SaveEvents(_ context.Context, aggregateID string, events []domain.DomainEvent, expectedVersion int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing := s.events[aggregateID]
	if len(existing) != expectedVersion {
		return domain.ErrConcurrencyConflict
	}

	for _, evt := range events {
		jsonData, _ := json.Marshal(evt.EventData())
		stored := domain.NewGenericDomainEvent(aggregateID, evt.EventType(), jsonData, evt.OccurredAt())
		s.events[aggregateID] = append(s.events[aggregateID], stored)
	}
	return nil
}

func (s *inMemoryEventStore) GetEvents(_ context.Context, aggregateID string) ([]domain.DomainEvent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	events := s.events[aggregateID]
	if len(events) == 0 {
		return nil, nil
	}

	result := make([]domain.DomainEvent, len(events))
	copy(result, events)
	return result, nil
}

func newTestCreateHandler() *CreateEditGrantHandler {
	es := newInMemoryEventStore()
	repo := repositories.NewEditGrantRepository(es)
	return NewCreateEditGrantHandler(repo)
}

func validCreateCommand() *commands.CreateEditGrant {
	return &commands.CreateEditGrant{
		GrantorID:    "grantor-id",
		GrantorEmail: "grantor@example.com",
		GranteeEmail: "grantee@example.com",
		ArtifactType: "capability",
		ArtifactID:   "550e8400-e29b-41d4-a716-446655440000",
		Scope:        "write",
		Reason:       "collaboration needed",
	}
}

func TestCreateEditGrantHandler_ValidCommand_CreatesGrant(t *testing.T) {
	handler := newTestCreateHandler()
	result, err := handler.Handle(context.Background(), validCreateCommand())
	require.NoError(t, err)
	assert.NotEmpty(t, result.CreatedID)
}

func TestCreateEditGrantHandler_InvalidArtifactType_ReturnsError(t *testing.T) {
	handler := newTestCreateHandler()
	cmd := validCreateCommand()
	cmd.ArtifactType = "invalid"
	_, err := handler.Handle(context.Background(), cmd)
	assert.Equal(t, valueobjects.ErrInvalidArtifactType, err)
}

func TestCreateEditGrantHandler_EmptyArtifactID_ReturnsError(t *testing.T) {
	handler := newTestCreateHandler()
	cmd := validCreateCommand()
	cmd.ArtifactID = ""
	_, err := handler.Handle(context.Background(), cmd)
	assert.Equal(t, valueobjects.ErrEmptyArtifactID, err)
}

func TestCreateEditGrantHandler_InvalidArtifactID_ReturnsError(t *testing.T) {
	handler := newTestCreateHandler()
	cmd := validCreateCommand()
	cmd.ArtifactID = "not-a-uuid"
	_, err := handler.Handle(context.Background(), cmd)
	assert.Equal(t, valueobjects.ErrInvalidArtifactID, err)
}

func TestCreateEditGrantHandler_SelfGrant_ReturnsError(t *testing.T) {
	handler := newTestCreateHandler()
	cmd := validCreateCommand()
	cmd.GranteeEmail = cmd.GrantorEmail
	_, err := handler.Handle(context.Background(), cmd)
	assert.Equal(t, aggregates.ErrCannotGrantToSelf, err)
}

func TestCreateEditGrantHandler_EmptyGrantorID_ReturnsError(t *testing.T) {
	handler := newTestCreateHandler()
	cmd := validCreateCommand()
	cmd.GrantorID = ""
	_, err := handler.Handle(context.Background(), cmd)
	assert.Equal(t, valueobjects.ErrGrantorIDEmpty, err)
}

func TestCreateEditGrantHandler_EmptyGrantorEmail_ReturnsError(t *testing.T) {
	handler := newTestCreateHandler()
	cmd := validCreateCommand()
	cmd.GrantorEmail = ""
	_, err := handler.Handle(context.Background(), cmd)
	assert.Equal(t, valueobjects.ErrGrantorEmailEmpty, err)
}

func TestCreateEditGrantHandler_InvalidScope_ReturnsError(t *testing.T) {
	handler := newTestCreateHandler()
	cmd := validCreateCommand()
	cmd.Scope = "invalid"
	_, err := handler.Handle(context.Background(), cmd)
	assert.Equal(t, valueobjects.ErrInvalidGrantScope, err)
}

func TestCreateEditGrantHandler_EmptyGranteeEmail_ReturnsError(t *testing.T) {
	handler := newTestCreateHandler()
	cmd := validCreateCommand()
	cmd.GranteeEmail = ""
	_, err := handler.Handle(context.Background(), cmd)
	assert.Equal(t, valueobjects.ErrGranteeEmailEmpty, err)
}

func TestCreateEditGrantHandler_InvalidGranteeEmail_ReturnsError(t *testing.T) {
	handler := newTestCreateHandler()
	cmd := validCreateCommand()
	cmd.GranteeEmail = "not-an-email"
	_, err := handler.Handle(context.Background(), cmd)
	assert.Equal(t, valueobjects.ErrGranteeEmailInvalid, err)
}

func TestCreateEditGrantHandler_ReasonTooLong_ReturnsError(t *testing.T) {
	handler := newTestCreateHandler()
	cmd := validCreateCommand()
	longReason := make([]byte, valueobjects.MaxReasonLength+1)
	for i := range longReason {
		longReason[i] = 'a'
	}
	cmd.Reason = string(longReason)
	_, err := handler.Handle(context.Background(), cmd)
	assert.Equal(t, valueobjects.ErrReasonTooLong, err)
}

func TestCreateEditGrantHandler_MultipleCreates_ProduceUniqueIDs(t *testing.T) {
	handler := newTestCreateHandler()

	result1, err := handler.Handle(context.Background(), validCreateCommand())
	require.NoError(t, err)

	result2, err := handler.Handle(context.Background(), validCreateCommand())
	require.NoError(t, err)

	assert.NotEqual(t, result1.CreatedID, result2.CreatedID)
}

type mockDomainEvent struct {
	aggregateID string
}

func (e mockDomainEvent) AggregateID() string                { return e.aggregateID }
func (e mockDomainEvent) EventType() string                  { return "MockEvent" }
func (e mockDomainEvent) OccurredAt() time.Time              { return time.Now() }
func (e mockDomainEvent) EventData() map[string]interface{} { return nil }
