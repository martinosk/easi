package repositories

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "easi/backend/internal/shared/eventsourcing"
	"easi/backend/internal/stewardship/domain/aggregates"
	"easi/backend/internal/stewardship/domain/valueobjects"
)

func storedCopies(t *testing.T, raised []domain.DomainEvent) []domain.DomainEvent {
	t.Helper()
	stored := make([]domain.DomainEvent, len(raised))
	for i, event := range raised {
		data, err := json.Marshal(event.EventData())
		require.NoError(t, err)
		stored[i] = domain.NewGenericDomainEvent(event.AggregateID(), event.EventType(), data, event.OccurredAt())
	}
	return stored
}

func TestStewardshipDeserializers_RoundTrip(t *testing.T) {
	concern, _ := valueobjects.NewConcern("planning")
	mette, _ := valueobjects.NewStewardRef("mette", valueobjects.UserActive)
	jonas, _ := valueobjects.NewStewardRef("jonas", valueobjects.UserActive)
	original, err := aggregates.NewStewardship(aggregates.Assignment{DomainID: "domain-1", Concern: concern, Steward: mette, AssignedBy: "alice@example.com"})
	require.NoError(t, err)
	require.NoError(t, original.Assign(jonas, "bob@example.com"))
	require.NoError(t, original.Release("bob@example.com"))

	deserialized, err := stewardshipEventDeserializers.Deserialize(storedCopies(t, original.GetUncommittedChanges()))
	require.NoError(t, err)
	require.Len(t, deserialized, 3)

	loaded, err := aggregates.LoadStewardshipFromHistory(deserialized)
	require.NoError(t, err)
	assert.Equal(t, original.ID(), loaded.ID())
	assert.Equal(t, "domain-1", loaded.DomainID())
	assert.Equal(t, "planning", loaded.Concern().Value())
	assert.Equal(t, "jonas", loaded.Steward().Value())
	assert.Equal(t, original.AssignedAt().UTC(), loaded.AssignedAt().UTC())
	assert.True(t, loaded.IsReleased())
}
