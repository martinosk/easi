package repositories

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "easi/backend/internal/shared/eventsourcing"
	"easi/backend/internal/stewardship/domain/aggregates"
	"easi/backend/internal/stewardship/domain/events"
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
	assert.Equal(t, 3, loaded.Version())
	assert.ErrorIs(t, loaded.Assign(mette, "bob@example.com"), aggregates.ErrStewardshipReleased)
}

func TestStewardshipDeserializers_RestoreEveryRecordedField(t *testing.T) {
	concern, _ := valueobjects.NewConcern("planning")
	mette, _ := valueobjects.NewStewardRef("mette", valueobjects.UserActive)
	original, err := aggregates.NewStewardship(aggregates.Assignment{DomainID: "domain-1", Concern: concern, Steward: mette, AssignedBy: "alice@example.com"})
	require.NoError(t, err)
	require.NoError(t, original.Release("bob@example.com"))
	raised := original.GetUncommittedChanges()

	deserialized, err := stewardshipEventDeserializers.Deserialize(storedCopies(t, raised))

	require.NoError(t, err)
	require.Len(t, deserialized, 2)
	assigned, released := deserialized[0].(events.StewardAssigned), deserialized[1].(events.StewardReleased)
	assert.Equal(t, raised[0].EventData(), assigned.EventData())
	assert.Equal(t, raised[1].EventData(), released.EventData())
}
