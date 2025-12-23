package aggregates

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"easi/backend/internal/auth/domain/events"
	"easi/backend/internal/auth/domain/valueobjects"
	"easi/backend/internal/shared/eventsourcing"
)

const DefaultInvitationTTL = 7 * 24 * time.Hour

var (
	ErrInvitationAlreadyAccepted = errors.New("invitation has already been accepted")
	ErrInvitationAlreadyRevoked  = errors.New("invitation has already been revoked")
	ErrInvitationAlreadyExpired  = errors.New("invitation has already expired")
	ErrInvitationNotPending      = errors.New("invitation is not in pending status")
	ErrInvitationExpired         = errors.New("invitation has expired")
)

type Invitation struct {
	domain.AggregateRoot
	email       valueobjects.Email
	role        valueobjects.Role
	status      valueobjects.InvitationStatus
	inviterInfo *valueobjects.InviterInfo
	createdAt   time.Time
	expiresAt   time.Time
	acceptedAt  *time.Time
	revokedAt   *time.Time
}

func NewInvitation(
	email valueobjects.Email,
	role valueobjects.Role,
	inviterInfo *valueobjects.InviterInfo,
) (*Invitation, error) {
	invitation := &Invitation{
		AggregateRoot: domain.NewAggregateRoot(),
	}

	expiresAt := time.Now().UTC().Add(DefaultInvitationTTL)

	var inviterID, inviterEmail string
	if inviterInfo != nil {
		inviterID = inviterInfo.UserID().String()
		inviterEmail = inviterInfo.Email().Value()
	}

	event := events.NewInvitationCreated(
		invitation.ID(),
		email.Value(),
		role.String(),
		inviterID,
		inviterEmail,
		expiresAt,
	)

	invitation.apply(event)
	invitation.RaiseEvent(event)

	return invitation, nil
}

func LoadInvitationFromHistory(evts []domain.DomainEvent) (*Invitation, error) {
	invitation := &Invitation{
		AggregateRoot: domain.NewAggregateRoot(),
	}

	invitation.LoadFromHistory(evts, func(event domain.DomainEvent) {
		invitation.apply(event)
	})

	return invitation, nil
}

func (i *Invitation) Accept() error {
	if !i.status.IsPending() {
		if i.status.IsAccepted() {
			return ErrInvitationAlreadyAccepted
		}
		if i.status.IsRevoked() {
			return ErrInvitationAlreadyRevoked
		}
		if i.status.IsExpired() {
			return ErrInvitationAlreadyExpired
		}
		return ErrInvitationNotPending
	}

	if i.IsExpired() {
		return ErrInvitationExpired
	}

	event := events.NewInvitationAccepted(i.ID(), i.email.Value())
	i.apply(event)
	i.RaiseEvent(event)

	return nil
}

func (i *Invitation) Revoke() error {
	if err := i.validatePendingForTransition(ErrInvitationAlreadyRevoked); err != nil {
		return err
	}

	event := events.NewInvitationRevoked(i.ID())
	i.apply(event)
	i.RaiseEvent(event)

	return nil
}

func (i *Invitation) MarkExpired() error {
	if err := i.validatePendingForTransition(ErrInvitationAlreadyExpired); err != nil {
		return err
	}

	event := events.NewInvitationExpired(i.ID())
	i.apply(event)
	i.RaiseEvent(event)

	return nil
}

func (i *Invitation) validatePendingForTransition(alreadyInStateErr error) error {
	if i.status.IsPending() {
		return nil
	}
	if i.status.IsRevoked() && alreadyInStateErr == ErrInvitationAlreadyRevoked {
		return ErrInvitationAlreadyRevoked
	}
	if i.status.IsExpired() && alreadyInStateErr == ErrInvitationAlreadyExpired {
		return ErrInvitationAlreadyExpired
	}
	return ErrInvitationNotPending
}

func (i *Invitation) IsExpired() bool {
	return time.Now().UTC().After(i.expiresAt)
}

func (i *Invitation) apply(event domain.DomainEvent) {
	switch e := event.(type) {
	case events.InvitationCreated:
		i.AggregateRoot = domain.NewAggregateRootWithID(e.ID)
		i.email, _ = valueobjects.NewEmail(e.Email)
		i.role, _ = valueobjects.RoleFromString(e.Role)
		i.status = valueobjects.InvitationStatusPending
		if e.InviterID != "" && e.InviterEmail != "" {
			inviterEmail, _ := valueobjects.NewEmail(e.InviterEmail)
			inviterID := parseUUID(e.InviterID)
			info, _ := valueobjects.NewInviterInfo(inviterID, inviterEmail)
			i.inviterInfo = &info
		}
		i.createdAt = e.CreatedAt
		i.expiresAt = e.ExpiresAt
	case events.InvitationAccepted:
		i.status = valueobjects.InvitationStatusAccepted
		i.acceptedAt = &e.AcceptedAt
	case events.InvitationRevoked:
		i.status = valueobjects.InvitationStatusRevoked
		i.revokedAt = &e.RevokedAt
	case events.InvitationExpired:
		i.status = valueobjects.InvitationStatusExpired
	}
}

func parseUUID(s string) uuid.UUID {
	id, _ := uuid.Parse(s)
	return id
}

func (i *Invitation) Email() valueobjects.Email {
	return i.email
}

func (i *Invitation) Role() valueobjects.Role {
	return i.role
}

func (i *Invitation) Status() valueobjects.InvitationStatus {
	return i.status
}

func (i *Invitation) InviterInfo() *valueobjects.InviterInfo {
	return i.inviterInfo
}

func (i *Invitation) CreatedAt() time.Time {
	return i.createdAt
}

func (i *Invitation) ExpiresAt() time.Time {
	return i.expiresAt
}

func (i *Invitation) AcceptedAt() *time.Time {
	return i.acceptedAt
}

func (i *Invitation) RevokedAt() *time.Time {
	return i.revokedAt
}
