package aggregates

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"easi/backend/internal/auth/domain/events"
	"easi/backend/internal/auth/domain/valueobjects"
	"easi/backend/internal/shared/domain"
)

func TestNewInvitation_CreatesPendingInvitation(t *testing.T) {
	email, _ := valueobjects.NewEmail("user@example.com")
	role := valueobjects.RoleArchitect

	invitation, err := NewInvitation(email, role, nil)
	require.NoError(t, err)

	assert.NotEmpty(t, invitation.ID())
	assert.Equal(t, email, invitation.Email())
	assert.Equal(t, role, invitation.Role())
	assert.True(t, invitation.Status().IsPending())
	assert.Nil(t, invitation.InviterInfo())
}

func TestNewInvitation_WithInviterInfo(t *testing.T) {
	email, _ := valueobjects.NewEmail("user@example.com")
	role := valueobjects.RoleStakeholder
	inviterEmail, _ := valueobjects.NewEmail("admin@example.com")
	inviterInfo, _ := valueobjects.NewInviterInfo(uuid.New(), inviterEmail)

	invitation, err := NewInvitation(email, role, &inviterInfo)
	require.NoError(t, err)

	assert.NotNil(t, invitation.InviterInfo())
	assert.Equal(t, inviterInfo.UserID(), invitation.InviterInfo().UserID())
	assert.Equal(t, inviterEmail, invitation.InviterInfo().Email())
}

func TestNewInvitation_SetsSevenDayTTL(t *testing.T) {
	email, _ := valueobjects.NewEmail("user@example.com")
	role := valueobjects.RoleArchitect
	beforeCreation := time.Now().UTC()

	invitation, err := NewInvitation(email, role, nil)
	require.NoError(t, err)

	afterCreation := time.Now().UTC()

	expectedMinExpiry := beforeCreation.Add(DefaultInvitationTTL)
	expectedMaxExpiry := afterCreation.Add(DefaultInvitationTTL)

	assert.True(t, invitation.ExpiresAt().After(expectedMinExpiry) || invitation.ExpiresAt().Equal(expectedMinExpiry))
	assert.True(t, invitation.ExpiresAt().Before(expectedMaxExpiry) || invitation.ExpiresAt().Equal(expectedMaxExpiry))
}

func TestNewInvitation_RaisesInvitationCreatedEvent(t *testing.T) {
	email, _ := valueobjects.NewEmail("user@example.com")
	role := valueobjects.RoleArchitect

	invitation, err := NewInvitation(email, role, nil)
	require.NoError(t, err)

	uncommittedEvents := invitation.GetUncommittedChanges()
	require.Len(t, uncommittedEvents, 1)

	event, ok := uncommittedEvents[0].(events.InvitationCreated)
	require.True(t, ok)
	assert.Equal(t, invitation.ID(), event.ID)
	assert.Equal(t, email.Value(), event.Email)
	assert.Equal(t, role.String(), event.Role)
}

func TestInvitation_Accept_FromPendingStatus(t *testing.T) {
	email, _ := valueobjects.NewEmail("user@example.com")
	role := valueobjects.RoleArchitect

	invitation, _ := NewInvitation(email, role, nil)
	invitation.MarkChangesAsCommitted()

	err := invitation.Accept()
	require.NoError(t, err)

	assert.True(t, invitation.Status().IsAccepted())
	assert.NotNil(t, invitation.AcceptedAt())

	uncommittedEvents := invitation.GetUncommittedChanges()
	require.Len(t, uncommittedEvents, 1)

	event, ok := uncommittedEvents[0].(events.InvitationAccepted)
	require.True(t, ok)
	assert.Equal(t, invitation.ID(), event.ID)
	assert.Equal(t, email.Value(), event.Email)
}

func TestInvitation_Accept_WhenAlreadyAccepted(t *testing.T) {
	email, _ := valueobjects.NewEmail("user@example.com")
	role := valueobjects.RoleArchitect

	invitation, _ := NewInvitation(email, role, nil)
	invitation.MarkChangesAsCommitted()

	invitation.Accept()
	invitation.MarkChangesAsCommitted()

	err := invitation.Accept()
	assert.ErrorIs(t, err, ErrInvitationAlreadyAccepted)
}

func TestInvitation_Accept_WhenRevoked(t *testing.T) {
	email, _ := valueobjects.NewEmail("user@example.com")
	role := valueobjects.RoleArchitect

	invitation, _ := NewInvitation(email, role, nil)
	invitation.MarkChangesAsCommitted()

	invitation.Revoke()
	invitation.MarkChangesAsCommitted()

	err := invitation.Accept()
	assert.ErrorIs(t, err, ErrInvitationAlreadyRevoked)
}

func TestInvitation_Accept_WhenExpired(t *testing.T) {
	email, _ := valueobjects.NewEmail("user@example.com")
	role := valueobjects.RoleArchitect

	invitation, _ := NewInvitation(email, role, nil)
	invitation.MarkChangesAsCommitted()

	invitation.MarkExpired()
	invitation.MarkChangesAsCommitted()

	err := invitation.Accept()
	assert.ErrorIs(t, err, ErrInvitationAlreadyExpired)
}

func TestInvitation_Accept_WhenTTLElapsed(t *testing.T) {
	email, _ := valueobjects.NewEmail("user@example.com")
	role := valueobjects.RoleArchitect

	invitation, _ := NewInvitation(email, role, nil)
	invitation.MarkChangesAsCommitted()

	invitation.expiresAt = time.Now().UTC().Add(-1 * time.Hour)

	err := invitation.Accept()
	assert.ErrorIs(t, err, ErrInvitationExpired)
}

func TestInvitation_Revoke_FromPendingStatus(t *testing.T) {
	email, _ := valueobjects.NewEmail("user@example.com")
	role := valueobjects.RoleArchitect

	invitation, _ := NewInvitation(email, role, nil)
	invitation.MarkChangesAsCommitted()

	err := invitation.Revoke()
	require.NoError(t, err)

	assert.True(t, invitation.Status().IsRevoked())
	assert.NotNil(t, invitation.RevokedAt())

	uncommittedEvents := invitation.GetUncommittedChanges()
	require.Len(t, uncommittedEvents, 1)

	event, ok := uncommittedEvents[0].(events.InvitationRevoked)
	require.True(t, ok)
	assert.Equal(t, invitation.ID(), event.ID)
}

func TestInvitation_Revoke_WhenAlreadyRevoked(t *testing.T) {
	email, _ := valueobjects.NewEmail("user@example.com")
	role := valueobjects.RoleArchitect

	invitation, _ := NewInvitation(email, role, nil)
	invitation.MarkChangesAsCommitted()

	invitation.Revoke()
	invitation.MarkChangesAsCommitted()

	err := invitation.Revoke()
	assert.ErrorIs(t, err, ErrInvitationAlreadyRevoked)
}

func TestInvitation_Revoke_WhenAccepted(t *testing.T) {
	email, _ := valueobjects.NewEmail("user@example.com")
	role := valueobjects.RoleArchitect

	invitation, _ := NewInvitation(email, role, nil)
	invitation.MarkChangesAsCommitted()

	invitation.Accept()
	invitation.MarkChangesAsCommitted()

	err := invitation.Revoke()
	assert.ErrorIs(t, err, ErrInvitationNotPending)
}

func TestInvitation_MarkExpired_FromPendingStatus(t *testing.T) {
	email, _ := valueobjects.NewEmail("user@example.com")
	role := valueobjects.RoleArchitect

	invitation, _ := NewInvitation(email, role, nil)
	invitation.MarkChangesAsCommitted()

	err := invitation.MarkExpired()
	require.NoError(t, err)

	assert.True(t, invitation.Status().IsExpired())

	uncommittedEvents := invitation.GetUncommittedChanges()
	require.Len(t, uncommittedEvents, 1)

	event, ok := uncommittedEvents[0].(events.InvitationExpired)
	require.True(t, ok)
	assert.Equal(t, invitation.ID(), event.ID)
}

func TestInvitation_MarkExpired_WhenAlreadyExpired(t *testing.T) {
	email, _ := valueobjects.NewEmail("user@example.com")
	role := valueobjects.RoleArchitect

	invitation, _ := NewInvitation(email, role, nil)
	invitation.MarkChangesAsCommitted()

	invitation.MarkExpired()
	invitation.MarkChangesAsCommitted()

	err := invitation.MarkExpired()
	assert.ErrorIs(t, err, ErrInvitationAlreadyExpired)
}

func TestInvitation_MarkExpired_WhenAccepted(t *testing.T) {
	email, _ := valueobjects.NewEmail("user@example.com")
	role := valueobjects.RoleArchitect

	invitation, _ := NewInvitation(email, role, nil)
	invitation.MarkChangesAsCommitted()

	invitation.Accept()
	invitation.MarkChangesAsCommitted()

	err := invitation.MarkExpired()
	assert.ErrorIs(t, err, ErrInvitationNotPending)
}

func TestInvitation_IsExpired_ChecksTTL(t *testing.T) {
	email, _ := valueobjects.NewEmail("user@example.com")
	role := valueobjects.RoleArchitect

	invitation, _ := NewInvitation(email, role, nil)

	t.Run("not expired within TTL", func(t *testing.T) {
		invitation.expiresAt = time.Now().UTC().Add(1 * time.Hour)
		assert.False(t, invitation.IsExpired())
	})

	t.Run("expired after TTL", func(t *testing.T) {
		invitation.expiresAt = time.Now().UTC().Add(-1 * time.Hour)
		assert.True(t, invitation.IsExpired())
	})
}

func TestLoadInvitationFromHistory_ReconstructsState(t *testing.T) {
	invitationID := uuid.New().String()
	email := "user@example.com"
	role := "architect"
	inviterID := uuid.New().String()
	inviterEmail := "admin@example.com"
	expiresAt := time.Now().UTC().Add(DefaultInvitationTTL)

	createdEvent := events.NewInvitationCreated(
		invitationID,
		email,
		role,
		inviterID,
		inviterEmail,
		expiresAt,
	)

	evts := []domain.DomainEvent{createdEvent}
	invitation, err := LoadInvitationFromHistory(evts)
	require.NoError(t, err)

	assert.Equal(t, invitationID, invitation.ID())
	assert.Equal(t, email, invitation.Email().Value())
	assert.Equal(t, role, invitation.Role().String())
	assert.True(t, invitation.Status().IsPending())
	assert.NotNil(t, invitation.InviterInfo())
	assert.Equal(t, inviterEmail, invitation.InviterInfo().Email().Value())
}

func TestLoadInvitationFromHistory_AcceptedInvitation(t *testing.T) {
	invitationID := uuid.New().String()
	email := "user@example.com"
	expiresAt := time.Now().UTC().Add(DefaultInvitationTTL)

	createdEvent := events.NewInvitationCreated(
		invitationID,
		email,
		"stakeholder",
		"",
		"",
		expiresAt,
	)
	acceptedEvent := events.NewInvitationAccepted(invitationID, email)

	evts := []domain.DomainEvent{createdEvent, acceptedEvent}
	invitation, err := LoadInvitationFromHistory(evts)
	require.NoError(t, err)

	assert.True(t, invitation.Status().IsAccepted())
	assert.NotNil(t, invitation.AcceptedAt())
}

func TestLoadInvitationFromHistory_RevokedInvitation(t *testing.T) {
	invitationID := uuid.New().String()
	expiresAt := time.Now().UTC().Add(DefaultInvitationTTL)

	createdEvent := events.NewInvitationCreated(
		invitationID,
		"user@example.com",
		"architect",
		"",
		"",
		expiresAt,
	)
	revokedEvent := events.NewInvitationRevoked(invitationID)

	evts := []domain.DomainEvent{createdEvent, revokedEvent}
	invitation, err := LoadInvitationFromHistory(evts)
	require.NoError(t, err)

	assert.True(t, invitation.Status().IsRevoked())
	assert.NotNil(t, invitation.RevokedAt())
}

func TestLoadInvitationFromHistory_ExpiredInvitation(t *testing.T) {
	invitationID := uuid.New().String()
	expiresAt := time.Now().UTC().Add(DefaultInvitationTTL)

	createdEvent := events.NewInvitationCreated(
		invitationID,
		"user@example.com",
		"architect",
		"",
		"",
		expiresAt,
	)
	expiredEvent := events.NewInvitationExpired(invitationID)

	evts := []domain.DomainEvent{createdEvent, expiredEvent}
	invitation, err := LoadInvitationFromHistory(evts)
	require.NoError(t, err)

	assert.True(t, invitation.Status().IsExpired())
}
