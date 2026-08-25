package handlers

import (
	"context"
	"testing"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/domain/aggregates"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockUpdateMetadataRepository struct {
	capability *aggregates.Capability
	saved      *aggregates.Capability
}

func (m *mockUpdateMetadataRepository) GetByID(ctx context.Context, id string) (*aggregates.Capability, error) {
	return m.capability, nil
}

func (m *mockUpdateMetadataRepository) Save(ctx context.Context, capability *aggregates.Capability) error {
	m.saved = capability
	return nil
}

type mockEAOwnerResolver struct {
	resolved      string
	err           error
	receivedValue string
	called        bool
}

func (m *mockEAOwnerResolver) ResolveEAOwner(ctx context.Context, value string) (string, error) {
	m.called = true
	m.receivedValue = value
	if m.err != nil {
		return "", m.err
	}
	return m.resolved, nil
}

func newMetadataTestCapability(t *testing.T) *aggregates.Capability {
	t.Helper()
	name, err := valueobjects.NewCapabilityName("Payments")
	require.NoError(t, err)
	capability, err := aggregates.NewCapability(name, valueobjects.MustNewDescription(""), valueobjects.CapabilityID{}, valueobjects.LevelL1)
	require.NoError(t, err)
	capability.MarkChangesAsCommitted()
	return capability
}

func newMetadataCommand(capabilityID, eaOwner string) *commands.UpdateCapabilityMetadata {
	return &commands.UpdateCapabilityMetadata{
		ID:            capabilityID,
		MaturityValue: 50,
		EAOwner:       eaOwner,
		Status:        "Active",
	}
}

func TestUpdateCapabilityMetadataHandler_ResolvesEAOwnerToUserID(t *testing.T) {
	capability := newMetadataTestCapability(t)
	repo := &mockUpdateMetadataRepository{capability: capability}
	resolver := &mockEAOwnerResolver{resolved: "2ec46b70-63b3-4d6d-92f0-1d385f9d4c4b"}
	handler := NewUpdateCapabilityMetadataHandler(repo, resolver)

	_, err := handler.Handle(context.Background(), newMetadataCommand(capability.ID(), "Alice Smith"))
	require.NoError(t, err)

	assert.Equal(t, "Alice Smith", resolver.receivedValue)
	require.NotNil(t, repo.saved)
	assert.Equal(t, "2ec46b70-63b3-4d6d-92f0-1d385f9d4c4b", repo.saved.EAOwner().Value())
}

func TestUpdateCapabilityMetadataHandler_RejectsUnresolvableEAOwner(t *testing.T) {
	capability := newMetadataTestCapability(t)
	repo := &mockUpdateMetadataRepository{capability: capability}
	resolver := &mockEAOwnerResolver{err: valueobjects.ErrEAOwnerNotUser}
	handler := NewUpdateCapabilityMetadataHandler(repo, resolver)

	_, err := handler.Handle(context.Background(), newMetadataCommand(capability.ID(), "Zaphod"))

	assert.ErrorIs(t, err, valueobjects.ErrEAOwnerNotUser)
	assert.Nil(t, repo.saved)
}

func TestUpdateCapabilityMetadataHandler_KeepsUnchangedLegacyEAOwner(t *testing.T) {
	capability := newMetadataTestCapability(t)
	ownershipModel, err := valueobjects.NewOwnershipModel("")
	require.NoError(t, err)
	legacyMetadata := valueobjects.NewCapabilityMetadata(
		valueobjects.MaturityGenesis,
		ownershipModel,
		valueobjects.NewOwner(""),
		valueobjects.EAOwnerFromHistory("Old Legacy Owner"),
		valueobjects.StatusActive,
	)
	require.NoError(t, capability.UpdateMetadata(legacyMetadata))
	capability.MarkChangesAsCommitted()

	repo := &mockUpdateMetadataRepository{capability: capability}
	resolver := &mockEAOwnerResolver{err: valueobjects.ErrEAOwnerNotUser}
	handler := NewUpdateCapabilityMetadataHandler(repo, resolver)

	_, err = handler.Handle(context.Background(), newMetadataCommand(capability.ID(), "Old Legacy Owner"))
	require.NoError(t, err)

	require.NotNil(t, repo.saved)
	assert.Equal(t, "Old Legacy Owner", repo.saved.EAOwner().Value())
}

func TestUpdateCapabilityMetadataHandler_SkipsResolutionForEmptyEAOwner(t *testing.T) {
	capability := newMetadataTestCapability(t)
	repo := &mockUpdateMetadataRepository{capability: capability}
	resolver := &mockEAOwnerResolver{}
	handler := NewUpdateCapabilityMetadataHandler(repo, resolver)

	_, err := handler.Handle(context.Background(), newMetadataCommand(capability.ID(), ""))
	require.NoError(t, err)

	assert.False(t, resolver.called)
	require.NotNil(t, repo.saved)
	assert.True(t, repo.saved.EAOwner().IsEmpty())
}
