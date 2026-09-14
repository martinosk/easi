package repositories

import (
	"testing"

	"easi/backend/internal/architecturemodeling/domain/aggregates"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComponentContainmentsDeserializers_RoundTrip(t *testing.T) {
	quoting, billing, crm := valueobjects.NewComponentID(), valueobjects.NewComponentID(), valueobjects.NewComponentID()
	composition, _ := valueobjects.NewContainmentKind(valueobjects.ContainmentComposition)
	aggregation, _ := valueobjects.NewContainmentKind(valueobjects.ContainmentAggregation)
	original := aggregates.NewComponentContainments()
	require.NoError(t, original.Attach(quoting, crm, composition))
	require.NoError(t, original.Attach(billing, crm, aggregation))
	require.NoError(t, original.Detach(quoting))

	storedEvents := simulateComponentEventStoreRoundTrip(t, original.GetUncommittedChanges())
	deserialized, err := componentContainmentsEventDeserializers.Deserialize(storedEvents)
	require.NoError(t, err)
	require.Len(t, deserialized, 3)

	loaded, err := aggregates.LoadComponentContainmentsFromHistory(deserialized)
	require.NoError(t, err)

	assert.Equal(t, original.ID(), loaded.ID())
	assert.False(t, loaded.IsPart(quoting))
	containment, ok := loaded.ContainmentOf(billing)
	require.True(t, ok)
	assert.True(t, containment.Parent.Equals(crm))
	assert.Equal(t, valueobjects.ContainmentAggregation, containment.Kind.String())
}
