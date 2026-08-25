package aggregates

import (
	"testing"

	"easi/backend/internal/capabilitymapping/domain/events"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	domain "easi/backend/internal/shared/eventsourcing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildMetadataWithEAOwner(t *testing.T, eaOwner valueobjects.EAOwner) valueobjects.CapabilityMetadata {
	t.Helper()
	maturityLevel, err := valueobjects.NewMaturityLevelFromValue(50)
	require.NoError(t, err)
	ownershipModel, err := valueobjects.NewOwnershipModel("TeamOwned")
	require.NoError(t, err)
	status, err := valueobjects.NewCapabilityStatus("Active")
	require.NoError(t, err)
	return valueobjects.NewCapabilityMetadata(
		maturityLevel,
		ownershipModel,
		valueobjects.NewOwner("Primary Person"),
		eaOwner,
		status,
	)
}

func TestUpdateMetadata_CarriesEAOwnerIntoEvent(t *testing.T) {
	capability := createCapability(t, "Customer Engagement", "L1")
	capability.MarkChangesAsCommitted()

	ref, err := valueobjects.NewEAOwner("2ec46b70-63b3-4d6d-92f0-1d385f9d4c4b")
	require.NoError(t, err)

	require.NoError(t, capability.UpdateMetadata(buildMetadataWithEAOwner(t, ref)))

	uncommitted := capability.GetUncommittedChanges()
	require.Len(t, uncommitted, 1)
	event, ok := uncommitted[0].(events.CapabilityMetadataUpdated)
	require.True(t, ok)
	assert.Equal(t, "2ec46b70-63b3-4d6d-92f0-1d385f9d4c4b", event.EAOwner)
	assert.Equal(t, "2ec46b70-63b3-4d6d-92f0-1d385f9d4c4b", capability.EAOwner().Value())
}

func TestLoadFromHistory_PreservesLegacyFreeTextEAOwner(t *testing.T) {
	capability := createCapability(t, "Customer Engagement", "L1")
	created := capability.GetUncommittedChanges()

	metadataUpdated := events.NewCapabilityMetadataUpdated(
		capability.ID(), "", 0, 50, "TeamOwned", "Primary Person", "Alice Smith", "Active",
	)

	loaded, err := LoadCapabilityFromHistory(append(created, domain.DomainEvent(metadataUpdated)))
	require.NoError(t, err)
	assert.Equal(t, "Alice Smith", loaded.EAOwner().Value())
}
