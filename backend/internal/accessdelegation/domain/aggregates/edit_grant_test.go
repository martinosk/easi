package aggregates

import (
	"testing"
	"time"

	"easi/backend/internal/accessdelegation/domain/valueobjects"
	domain "easi/backend/internal/shared/eventsourcing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEditGrant_CreatesActiveGrantWith30DayTTL(t *testing.T) {
	before := time.Now().UTC()

	grant := createEditGrant(t)

	after := time.Now().UTC()

	assert.NotEmpty(t, grant.ID())
	assert.Equal(t, valueobjects.ArtifactTypeCapability, grant.ArtifactRef().Type())
	assert.Equal(t, "cap-123", grant.ArtifactRef().ID())
	assert.Equal(t, "grantor-id", grant.GrantorID())
	assert.Equal(t, "grantor@example.com", grant.GrantorEmail())
	assert.Equal(t, "grantee@example.com", grant.GranteeEmail())
	assert.Equal(t, valueobjects.GrantScopeWrite, grant.Scope())
	assert.True(t, grant.Status().IsActive())
	assert.Equal(t, "collaboration needed", grant.Reason())

	expectedExpiry := before.Add(DefaultEditGrantTTL)
	assert.True(t, grant.ExpiresAt().After(expectedExpiry) || grant.ExpiresAt().Equal(expectedExpiry))
	assert.True(t, grant.ExpiresAt().Before(after.Add(DefaultEditGrantTTL)) || grant.ExpiresAt().Equal(after.Add(DefaultEditGrantTTL)))

	assert.Nil(t, grant.RevokedAt())
}

func TestNewEditGrant_RaisesEditGrantActivatedEvent(t *testing.T) {
	grant := createEditGrant(t)

	uncommittedEvents := grant.GetUncommittedChanges()
	require.Len(t, uncommittedEvents, 1)
	assert.Equal(t, "EditGrantActivated", uncommittedEvents[0].EventType())

	eventData := uncommittedEvents[0].EventData()
	assert.Equal(t, grant.ID(), eventData["id"])
	assert.Equal(t, "capability", eventData["artifactType"])
	assert.Equal(t, "cap-123", eventData["artifactId"])
	assert.Equal(t, "grantor-id", eventData["grantorId"])
	assert.Equal(t, "grantor@example.com", eventData["grantorEmail"])
	assert.Equal(t, "grantee@example.com", eventData["granteeEmail"])
	assert.Equal(t, "write", eventData["scope"])
	assert.Equal(t, "collaboration needed", eventData["reason"])
}

func TestNewEditGrant_RejectsSelfGrant(t *testing.T) {
	artifactRef := mustNewArtifactRef(t, "capability", "cap-123")
	scope, _ := valueobjects.NewGrantScope("write")
	grantor := mustNewGrantor(t, "user-id", "same@example.com")

	grant, err := NewEditGrant(grantor, "same@example.com", artifactRef, scope, "reason")

	assert.Nil(t, grant)
	assert.Error(t, err)
	assert.Equal(t, ErrCannotGrantToSelf, err)
}

func TestRevoke_ActiveGrant_TransitionsToRevoked(t *testing.T) {
	grant := createEditGrant(t)
	grant.MarkChangesAsCommitted()

	err := grant.Revoke("admin@example.com")
	require.NoError(t, err)

	assert.True(t, grant.Status().IsRevoked())
	assert.NotNil(t, grant.RevokedAt())
}

func TestRevoke_ActiveGrant_RaisesEditGrantRevokedEvent(t *testing.T) {
	grant := createEditGrant(t)
	grant.MarkChangesAsCommitted()

	err := grant.Revoke("admin@example.com")
	require.NoError(t, err)

	uncommittedEvents := grant.GetUncommittedChanges()
	require.Len(t, uncommittedEvents, 1)
	assert.Equal(t, "EditGrantRevoked", uncommittedEvents[0].EventType())

	eventData := uncommittedEvents[0].EventData()
	assert.Equal(t, grant.ID(), eventData["id"])
	assert.Equal(t, "admin@example.com", eventData["revokedBy"])
}

func TestRevoke_AlreadyRevokedGrant_Fails(t *testing.T) {
	grant := createEditGrant(t)
	_ = grant.Revoke("admin@example.com")
	grant.MarkChangesAsCommitted()

	err := grant.Revoke("admin@example.com")
	assert.Error(t, err)
	assert.Equal(t, ErrGrantAlreadyRevoked, err)
}

func TestRevoke_ExpiredGrant_Fails(t *testing.T) {
	grant := createEditGrant(t)
	_ = grant.MarkExpired()
	grant.MarkChangesAsCommitted()

	err := grant.Revoke("admin@example.com")
	assert.Error(t, err)
	assert.Equal(t, ErrGrantAlreadyExpired, err)
}

func TestMarkExpired_ActiveGrant_TransitionsToExpired(t *testing.T) {
	grant := createEditGrant(t)
	grant.MarkChangesAsCommitted()

	err := grant.MarkExpired()
	require.NoError(t, err)

	assert.True(t, grant.Status().IsExpired())
}

func TestMarkExpired_ActiveGrant_RaisesEditGrantExpiredEvent(t *testing.T) {
	grant := createEditGrant(t)
	grant.MarkChangesAsCommitted()

	err := grant.MarkExpired()
	require.NoError(t, err)

	uncommittedEvents := grant.GetUncommittedChanges()
	require.Len(t, uncommittedEvents, 1)
	assert.Equal(t, "EditGrantExpired", uncommittedEvents[0].EventType())
}

func TestMarkExpired_AlreadyExpiredGrant_Fails(t *testing.T) {
	grant := createEditGrant(t)
	_ = grant.MarkExpired()
	grant.MarkChangesAsCommitted()

	err := grant.MarkExpired()
	assert.Error(t, err)
	assert.Equal(t, ErrGrantAlreadyExpired, err)
}

func TestMarkExpired_RevokedGrant_Fails(t *testing.T) {
	grant := createEditGrant(t)
	_ = grant.Revoke("admin@example.com")
	grant.MarkChangesAsCommitted()

	err := grant.MarkExpired()
	assert.Error(t, err)
	assert.Equal(t, ErrGrantAlreadyRevoked, err)
}

func TestLoadEditGrantFromHistory_ReconstructsActiveGrant(t *testing.T) {
	original := createEditGrant(t)

	allEvents := original.GetUncommittedChanges()

	loaded, err := LoadEditGrantFromHistory(allEvents)
	require.NoError(t, err)

	assert.Equal(t, original.ID(), loaded.ID())
	assert.Equal(t, original.ArtifactRef().Type(), loaded.ArtifactRef().Type())
	assert.Equal(t, original.ArtifactRef().ID(), loaded.ArtifactRef().ID())
	assert.Equal(t, original.GrantorID(), loaded.GrantorID())
	assert.Equal(t, original.GrantorEmail(), loaded.GrantorEmail())
	assert.Equal(t, original.GranteeEmail(), loaded.GranteeEmail())
	assert.Equal(t, original.Scope(), loaded.Scope())
	assert.True(t, loaded.Status().IsActive())
	assert.Equal(t, original.Reason(), loaded.Reason())
}

func TestLoadEditGrantFromHistory_ReconstructsRevokedGrant(t *testing.T) {
	original := createEditGrant(t)
	_ = original.Revoke("admin@example.com")

	allEvents := original.GetUncommittedChanges()

	loaded, err := LoadEditGrantFromHistory(allEvents)
	require.NoError(t, err)

	assert.Equal(t, original.ID(), loaded.ID())
	assert.True(t, loaded.Status().IsRevoked())
	assert.NotNil(t, loaded.RevokedAt())
}

func TestLoadEditGrantFromHistory_ReconstructsExpiredGrant(t *testing.T) {
	original := createEditGrant(t)
	_ = original.MarkExpired()

	allEvents := original.GetUncommittedChanges()

	loaded, err := LoadEditGrantFromHistory(allEvents)
	require.NoError(t, err)

	assert.Equal(t, original.ID(), loaded.ID())
	assert.True(t, loaded.Status().IsExpired())
}

func TestLoadEditGrantFromHistory_PreservesAggregateState(t *testing.T) {
	original := createEditGrant(t)

	allEvents := original.GetUncommittedChanges()

	loaded, err := LoadEditGrantFromHistory(allEvents)
	require.NoError(t, err)

	assert.Equal(t, original.GrantorEmail(), loaded.GrantorEmail())
	assert.Equal(t, original.GranteeEmail(), loaded.GranteeEmail())
	assert.Equal(t, original.Reason(), loaded.Reason())
	assert.Equal(t, original.ExpiresAt().Unix(), loaded.ExpiresAt().Unix())
}

func TestNewEditGrant_DifferentArtifactTypes(t *testing.T) {
	tests := []struct {
		name         string
		artifactType string
		artifactID   string
	}{
		{"capability", "capability", "cap-1"},
		{"component", "component", "comp-1"},
		{"view", "view", "view-1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			artifactRef := mustNewArtifactRef(t, tt.artifactType, tt.artifactID)
			scope, _ := valueobjects.NewGrantScope("write")

			grantor := mustNewGrantor(t, "grantor-id", "grantor@example.com")
			grant, err := NewEditGrant(grantor, "grantee@example.com", artifactRef, scope, "")
			require.NoError(t, err)
			assert.Equal(t, tt.artifactID, grant.ArtifactRef().ID())
		})
	}
}

func TestNewEditGrant_EmptyReasonIsAllowed(t *testing.T) {
	artifactRef := mustNewArtifactRef(t, "capability", "cap-123")
	scope, _ := valueobjects.NewGrantScope("write")

	grantor := mustNewGrantor(t, "grantor-id", "grantor@example.com")
	grant, err := NewEditGrant(grantor, "grantee@example.com", artifactRef, scope, "")
	require.NoError(t, err)
	assert.Equal(t, "", grant.Reason())
}

func TestDefaultEditGrantTTL_Is30Days(t *testing.T) {
	assert.Equal(t, 30*24*time.Hour, DefaultEditGrantTTL)
}

func TestLoadEditGrantFromHistory_EmptyEvents(t *testing.T) {
	loaded, err := LoadEditGrantFromHistory([]domain.DomainEvent{})

	require.NoError(t, err)
	assert.NotNil(t, loaded)
}

func createEditGrant(t *testing.T) *EditGrant {
	t.Helper()

	artifactRef := mustNewArtifactRef(t, "capability", "cap-123")
	scope, err := valueobjects.NewGrantScope("write")
	require.NoError(t, err)

	grantor := mustNewGrantor(t, "grantor-id", "grantor@example.com")
	grant, err := NewEditGrant(grantor, "grantee@example.com", artifactRef, scope, "collaboration needed")
	require.NoError(t, err)

	return grant
}

func mustNewArtifactRef(t *testing.T, artifactType string, artifactID string) valueobjects.ArtifactRef {
	t.Helper()

	at, err := valueobjects.NewArtifactType(artifactType)
	require.NoError(t, err)

	ref, err := valueobjects.NewArtifactRef(at, artifactID)
	require.NoError(t, err)

	return ref
}

func mustNewGrantor(t *testing.T, id string, email string) valueobjects.Grantor {
	t.Helper()

	grantor, err := valueobjects.NewGrantor(id, email)
	require.NoError(t, err)

	return grantor
}
