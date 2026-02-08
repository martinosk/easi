package aggregates

import (
	"errors"
	"time"

	"easi/backend/internal/accessdelegation/domain/events"
	"easi/backend/internal/accessdelegation/domain/valueobjects"
	domain "easi/backend/internal/shared/eventsourcing"
)

const DefaultEditGrantTTL = 30 * 24 * time.Hour

var (
	ErrCannotGrantToSelf  = errors.New("cannot grant edit access to yourself")
	ErrGrantAlreadyRevoked = errors.New("edit grant has already been revoked")
	ErrGrantAlreadyExpired = errors.New("edit grant has already expired")
	ErrGrantNotActive      = errors.New("edit grant is not active")
)

type EditGrant struct {
	domain.AggregateRoot
	artifactRef  valueobjects.ArtifactRef
	grantor      valueobjects.Grantor
	granteeEmail string
	scope        valueobjects.GrantScope
	status       valueobjects.GrantStatus
	reason       string
	createdAt    time.Time
	expiresAt    time.Time
	revokedAt    *time.Time
}

func NewEditGrant(
	grantor valueobjects.Grantor,
	granteeEmail string,
	artifactRef valueobjects.ArtifactRef,
	scope valueobjects.GrantScope,
	reason string,
) (*EditGrant, error) {
	if grantor.Email() == granteeEmail {
		return nil, ErrCannotGrantToSelf
	}

	grant := &EditGrant{
		AggregateRoot: domain.NewAggregateRoot(),
	}

	now := time.Now().UTC()
	event := events.EditGrantActivated{
		BaseEvent:    domain.NewBaseEvent(grant.ID()),
		ID:           grant.ID(),
		ArtifactType: artifactRef.Type().String(),
		ArtifactID:   artifactRef.ID(),
		GrantorID:    grantor.ID(),
		GrantorEmail: grantor.Email(),
		GranteeEmail: granteeEmail,
		Scope:        scope.String(),
		Reason:       reason,
		CreatedAt:    now,
		ExpiresAt:    now.Add(DefaultEditGrantTTL),
	}

	grant.apply(event)
	grant.RaiseEvent(event)

	return grant, nil
}

func LoadEditGrantFromHistory(evts []domain.DomainEvent) (*EditGrant, error) {
	grant := &EditGrant{
		AggregateRoot: domain.NewAggregateRoot(),
	}

	grant.LoadFromHistory(evts, func(event domain.DomainEvent) {
		grant.apply(event)
	})

	return grant, nil
}

func (g *EditGrant) Revoke(revokedBy string) error {
	if err := g.ensureActive(); err != nil {
		return err
	}

	event := events.NewEditGrantRevoked(g.ID(), revokedBy)
	g.apply(event)
	g.RaiseEvent(event)

	return nil
}

func (g *EditGrant) MarkExpired() error {
	if err := g.ensureActive(); err != nil {
		return err
	}

	event := events.NewEditGrantExpired(g.ID())
	g.apply(event)
	g.RaiseEvent(event)

	return nil
}

func (g *EditGrant) ensureActive() error {
	if g.status.IsActive() {
		return nil
	}
	if g.status.IsRevoked() {
		return ErrGrantAlreadyRevoked
	}
	if g.status.IsExpired() {
		return ErrGrantAlreadyExpired
	}
	return ErrGrantNotActive
}

func (g *EditGrant) IsExpired() bool {
	return time.Now().UTC().After(g.expiresAt)
}

func (g *EditGrant) apply(event domain.DomainEvent) {
	switch e := event.(type) {
	case events.EditGrantActivated:
		g.AggregateRoot = domain.NewAggregateRootWithID(e.ID)
		artifactType, _ := valueobjects.NewArtifactType(e.ArtifactType)
		g.artifactRef, _ = valueobjects.NewArtifactRef(artifactType, e.ArtifactID)
		g.grantor, _ = valueobjects.NewGrantor(e.GrantorID, e.GrantorEmail)
		g.granteeEmail = e.GranteeEmail
		g.scope, _ = valueobjects.NewGrantScope(e.Scope)
		g.status = valueobjects.GrantStatusActive
		g.reason = e.Reason
		g.createdAt = e.CreatedAt
		g.expiresAt = e.ExpiresAt
	case events.EditGrantRevoked:
		g.status = valueobjects.GrantStatusRevoked
		g.revokedAt = &e.RevokedAt
	case events.EditGrantExpired:
		g.status = valueobjects.GrantStatusExpired
	}
}

func (g *EditGrant) ArtifactRef() valueobjects.ArtifactRef { return g.artifactRef }
func (g *EditGrant) Grantor() valueobjects.Grantor         { return g.grantor }
func (g *EditGrant) GrantorID() string                     { return g.grantor.ID() }
func (g *EditGrant) GrantorEmail() string                  { return g.grantor.Email() }
func (g *EditGrant) GranteeEmail() string                  { return g.granteeEmail }
func (g *EditGrant) Scope() valueobjects.GrantScope        { return g.scope }
func (g *EditGrant) Status() valueobjects.GrantStatus      { return g.status }
func (g *EditGrant) Reason() string                        { return g.reason }
func (g *EditGrant) CreatedAt() time.Time                  { return g.createdAt }
func (g *EditGrant) ExpiresAt() time.Time                  { return g.expiresAt }
func (g *EditGrant) RevokedAt() *time.Time                 { return g.revokedAt }
