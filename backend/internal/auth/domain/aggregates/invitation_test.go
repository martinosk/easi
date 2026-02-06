package aggregates

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"easi/backend/internal/auth/domain/events"
	"easi/backend/internal/auth/domain/valueobjects"
	domain "easi/backend/internal/shared/eventsourcing"
)

func newPendingInvitation(t *testing.T) *Invitation {
	t.Helper()
	email, _ := valueobjects.NewEmail("user@example.com")
	invitation, err := NewInvitation(email, valueobjects.RoleArchitect, nil)
	require.NoError(t, err)
	invitation.MarkChangesAsCommitted()
	return invitation
}

func newInvitationCreatedEvent(id string) events.InvitationCreated {
	return events.NewInvitationCreated(
		id,
		"user@example.com",
		"architect",
		"",
		"",
		time.Now().UTC().Add(DefaultInvitationTTL),
	)
}

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
	invitation := newPendingInvitation(t)

	err := invitation.Accept()
	require.NoError(t, err)

	assert.True(t, invitation.Status().IsAccepted())
	assert.NotNil(t, invitation.AcceptedAt())

	uncommittedEvents := invitation.GetUncommittedChanges()
	require.Len(t, uncommittedEvents, 1)

	event, ok := uncommittedEvents[0].(events.InvitationAccepted)
	require.True(t, ok)
	assert.Equal(t, invitation.ID(), event.ID)
	assert.Equal(t, invitation.Email().Value(), event.Email)
}

func TestInvitation_Accept_WhenTTLElapsed(t *testing.T) {
	invitation := newPendingInvitation(t)
	invitation.expiresAt = time.Now().UTC().Add(-1 * time.Hour)

	err := invitation.Accept()
	assert.ErrorIs(t, err, ErrInvitationExpired)
}

func TestInvitation_Accept_RejectsTerminalStates(t *testing.T) {
	tests := []struct {
		name       string
		transition func(*Invitation)
		wantErr    error
	}{
		{"WhenAlreadyAccepted", func(inv *Invitation) { inv.Accept() }, ErrInvitationAlreadyAccepted},
		{"WhenRevoked", func(inv *Invitation) { inv.Revoke() }, ErrInvitationAlreadyRevoked},
		{"WhenExpired", func(inv *Invitation) { inv.MarkExpired() }, ErrInvitationAlreadyExpired},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			invitation := newPendingInvitation(t)
			tt.transition(invitation)
			invitation.MarkChangesAsCommitted()

			err := invitation.Accept()
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestInvitation_Revoke_FromPendingStatus(t *testing.T) {
	invitation := newPendingInvitation(t)

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

func TestInvitation_Revoke_RejectsTerminalStates(t *testing.T) {
	tests := []struct {
		name       string
		transition func(*Invitation)
		wantErr    error
	}{
		{"WhenAlreadyRevoked", func(inv *Invitation) { inv.Revoke() }, ErrInvitationAlreadyRevoked},
		{"WhenAccepted", func(inv *Invitation) { inv.Accept() }, ErrInvitationNotPending},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			invitation := newPendingInvitation(t)
			tt.transition(invitation)
			invitation.MarkChangesAsCommitted()

			err := invitation.Revoke()
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestInvitation_MarkExpired_FromPendingStatus(t *testing.T) {
	invitation := newPendingInvitation(t)

	err := invitation.MarkExpired()
	require.NoError(t, err)

	assert.True(t, invitation.Status().IsExpired())

	uncommittedEvents := invitation.GetUncommittedChanges()
	require.Len(t, uncommittedEvents, 1)

	event, ok := uncommittedEvents[0].(events.InvitationExpired)
	require.True(t, ok)
	assert.Equal(t, invitation.ID(), event.ID)
}

func TestInvitation_MarkExpired_RejectsTerminalStates(t *testing.T) {
	tests := []struct {
		name       string
		transition func(*Invitation)
		wantErr    error
	}{
		{"WhenAlreadyExpired", func(inv *Invitation) { inv.MarkExpired() }, ErrInvitationAlreadyExpired},
		{"WhenAccepted", func(inv *Invitation) { inv.Accept() }, ErrInvitationNotPending},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			invitation := newPendingInvitation(t)
			tt.transition(invitation)
			invitation.MarkChangesAsCommitted()

			err := invitation.MarkExpired()
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestInvitation_IsExpired_ChecksTTL(t *testing.T) {
	invitation := newPendingInvitation(t)

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

func TestLoadInvitationFromHistory_TerminalStates(t *testing.T) {
	tests := []struct {
		name        string
		extraEvent  func(string) domain.DomainEvent
		checkStatus func(*Invitation) bool
		hasTimeset  func(*Invitation) bool
	}{
		{
			"AcceptedInvitation",
			func(id string) domain.DomainEvent { return events.NewInvitationAccepted(id, "user@example.com") },
			func(inv *Invitation) bool { return inv.Status().IsAccepted() },
			func(inv *Invitation) bool { return inv.AcceptedAt() != nil },
		},
		{
			"RevokedInvitation",
			func(id string) domain.DomainEvent { return events.NewInvitationRevoked(id) },
			func(inv *Invitation) bool { return inv.Status().IsRevoked() },
			func(inv *Invitation) bool { return inv.RevokedAt() != nil },
		},
		{
			"ExpiredInvitation",
			func(id string) domain.DomainEvent { return events.NewInvitationExpired(id) },
			func(inv *Invitation) bool { return inv.Status().IsExpired() },
			nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			invitationID := uuid.New().String()
			createdEvent := newInvitationCreatedEvent(invitationID)

			evts := []domain.DomainEvent{createdEvent, tt.extraEvent(invitationID)}
			invitation, err := LoadInvitationFromHistory(evts)
			require.NoError(t, err)

			assert.True(t, tt.checkStatus(invitation))
			if tt.hasTimeset != nil {
				assert.True(t, tt.hasTimeset(invitation))
			}
		})
	}
}
