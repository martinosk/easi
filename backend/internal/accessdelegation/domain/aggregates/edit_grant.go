package aggregates

import (
	"errors"
	"fmt"
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
	granteeEmail valueobjects.GranteeEmail
	scope        valueobjects.GrantScope
	status       valueobjects.GrantStatus
	reason       valueobjects.Reason
	createdAt    time.Time
	expiresAt    time.Time
	revokedAt    *time.Time
}

type GrantRequest struct {
	Grantor      valueobjects.Grantor
	GranteeEmail valueobjects.GranteeEmail
	ArtifactRef  valueobjects.ArtifactRef
	Scope        valueobjects.GrantScope
	Reason       valueobjects.Reason
}

func NewEditGrant(req GrantRequest) (*EditGrant, error) {
	if req.Grantor.Email() == req.GranteeEmail.Value() {
		return nil, ErrCannotGrantToSelf
	}

	grant := &EditGrant{
		AggregateRoot: domain.NewAggregateRoot(),
	}

	now := time.Now().UTC()
	event := events.EditGrantActivated{
		BaseEvent:    domain.NewBaseEvent(grant.ID()),
		ID:           grant.ID(),
		ArtifactType: req.ArtifactRef.Type().String(),
		ArtifactID:   req.ArtifactRef.ID(),
		GrantorID:    req.Grantor.ID(),
		GrantorEmail: req.Grantor.Email(),
		GranteeEmail: req.GranteeEmail.Value(),
		Scope:        req.Scope.String(),
		Reason:       req.Reason.Value(),
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

func (g *EditGrant) apply(event domain.DomainEvent) {
	switch e := event.(type) {
	case events.EditGrantActivated:
		g.applyActivated(e)
	case events.EditGrantRevoked:
		g.status = valueobjects.GrantStatusRevoked
		g.revokedAt = &e.RevokedAt
	case events.EditGrantExpired:
		g.status = valueobjects.GrantStatusExpired
	}
}

func (g *EditGrant) applyActivated(e events.EditGrantActivated) {
	g.AggregateRoot = domain.NewAggregateRootWithID(e.ID)
	g.artifactRef = mustReconstructArtifactRef(e.ArtifactType, e.ArtifactID)
	g.grantor = mustReconstruct("Grantor", func() (valueobjects.Grantor, error) {
		return valueobjects.NewGrantor(e.GrantorID, e.GrantorEmail)
	})
	g.granteeEmail = mustReconstruct("GranteeEmail", func() (valueobjects.GranteeEmail, error) {
		return valueobjects.NewGranteeEmail(e.GranteeEmail)
	})
	g.scope = mustReconstruct("GrantScope", func() (valueobjects.GrantScope, error) {
		return valueobjects.NewGrantScope(e.Scope)
	})
	g.reason = mustReconstruct("Reason", func() (valueobjects.Reason, error) {
		return valueobjects.NewReason(e.Reason)
	})
	g.status = valueobjects.GrantStatusActive
	g.createdAt = e.CreatedAt
	g.expiresAt = e.ExpiresAt
}

func mustReconstructArtifactRef(artifactType, artifactID string) valueobjects.ArtifactRef {
	at := mustReconstruct("ArtifactType", func() (valueobjects.ArtifactType, error) {
		return valueobjects.NewArtifactType(artifactType)
	})
	return mustReconstruct("ArtifactRef", func() (valueobjects.ArtifactRef, error) {
		return valueobjects.NewArtifactRef(at, artifactID)
	})
}

func mustReconstruct[T any](name string, fn func() (T, error)) T {
	val, err := fn()
	if err != nil {
		panic(fmt.Sprintf("corrupt event data for %s: %v", name, err))
	}
	return val
}

func (g *EditGrant) ArtifactRef() valueobjects.ArtifactRef { return g.artifactRef }
func (g *EditGrant) Grantor() valueobjects.Grantor         { return g.grantor }
func (g *EditGrant) GrantorID() string                     { return g.grantor.ID() }
func (g *EditGrant) GrantorEmail() string                  { return g.grantor.Email() }
func (g *EditGrant) GranteeEmail() string                  { return g.granteeEmail.Value() }
func (g *EditGrant) Scope() valueobjects.GrantScope        { return g.scope }
func (g *EditGrant) Status() valueobjects.GrantStatus      { return g.status }
func (g *EditGrant) Reason() string                        { return g.reason.Value() }
func (g *EditGrant) CreatedAt() time.Time                  { return g.createdAt }
func (g *EditGrant) ExpiresAt() time.Time                  { return g.expiresAt }
func (g *EditGrant) RevokedAt() *time.Time                 { return g.revokedAt }
