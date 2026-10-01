//go:build integration

package readmodels

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
	domain "easi/backend/internal/shared/eventsourcing"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
	"easi/backend/internal/stewardship/application/home"
	stewardshipRM "easi/backend/internal/stewardship/application/readmodels"
	"easi/backend/internal/testing/testdb"
)

var stewardshipTables = []string{
	"stewardship.stewardships", "stewardship.domain_cache", "stewardship.user_cache",
	"stewardship.capability_cache", "stewardship.domain_assignment_cache", "stewardship.application_cache",
	"stewardship.realization_cache", "stewardship.time_assessment_cache", "stewardship.edit_grant_cache",
}

type landscape struct {
	t        *testing.T
	ctx      context.Context
	db       *database.TenantAwareDB
	caches   *Caches
	domains  *stewardshipRM.DomainCacheReadModel
	stewards *stewardshipRM.StewardshipReadModel
	queries  *Queries
	tenant   string
}

func newLandscape(t *testing.T) *landscape {
	t.Helper()
	if testing.Short() {
		t.Skip("Skipping integration test")
	}
	db := database.NewTenantAwareDB(testdb.Open(t))
	l := &landscape{t: t, db: db, caches: NewCaches(db), queries: NewQueries(db),
		domains: stewardshipRM.NewDomainCacheReadModel(db), stewards: stewardshipRM.NewStewardshipReadModel(db)}
	l.useTenant("home-it-" + uuid.NewString()[:8])
	return l
}

func (l *landscape) useTenant(tenant string) {
	l.tenant = tenant
	l.ctx = sharedctx.WithTenant(context.Background(), sharedvo.MustNewTenantID(tenant))
	admin := testdb.OpenAdmin(l.t)
	l.t.Cleanup(func() {
		for _, table := range stewardshipTables {
			_, _ = admin.Exec("DELETE FROM "+table+" WHERE tenant_id = $1", tenant)
		}
	})
}

func (l *landscape) must(err error) {
	l.t.Helper()
	require.NoError(l.t, err)
}

type componentID string

type ownership struct {
	state string
	kind  string
	owner string
}

type artifact struct {
	kind string
	id   string
}

func capabilityArtifact(id home.CapabilityID) artifact {
	return artifact{kind: "capability", id: string(id)}
}

func componentArtifact(id componentID) artifact { return artifact{kind: "component", id: string(id)} }

func (l *landscape) domain(name string) home.DomainID {
	id := uuid.NewString()
	l.must(l.domains.Upsert(l.ctx, stewardshipRM.CachedDomain{ID: id, Name: name}))
	return home.DomainID(id)
}

func (l *landscape) architect(domainID home.DomainID, name string, architect person) {
	l.must(l.domains.Upsert(l.ctx, stewardshipRM.CachedDomain{ID: string(domainID), Name: name, DomainArchitectID: architect.id}))
}

func (l *landscape) capability(name, level string, parentID home.CapabilityID) home.CapabilityID {
	id := home.CapabilityID(uuid.NewString())
	l.must(l.caches.SaveCapability(l.ctx, home.CachedCapability{ID: string(id), Name: name, Level: level, ParentID: string(parentID)}))
	return id
}

func (l *landscape) l1In(domainID home.DomainID, name string) home.CapabilityID {
	id := l.capability(name, "L1", "")
	l.must(l.caches.AssignToDomain(l.ctx, home.CachedAssignment{CapabilityID: id, DomainID: domainID}))
	return id
}

func (l *landscape) application(name string) componentID {
	id := uuid.NewString()
	l.must(l.caches.SaveApplication(l.ctx, id, name))
	return componentID(id)
}

func (l *landscape) realise(capabilityID home.CapabilityID, component componentID) {
	l.must(l.caches.SaveRealization(l.ctx, home.CachedRealization{ID: uuid.NewString(), CapabilityID: string(capabilityID), ComponentID: string(component)}))
}

func (l *landscape) grade(capabilityID home.CapabilityID, component componentID, grade string) {
	l.must(l.caches.SaveTimeAssessment(l.ctx, home.CachedTimeAssessment{CapabilityID: string(capabilityID), ComponentID: string(component), Grade: grade, AssessedAt: time.Now()}))
}

func (l *landscape) eaOwner(capabilityID home.CapabilityID, owner string) {
	l.must(l.caches.SetCapabilityMetadata(l.ctx, home.CachedCapabilityMetadata{CapabilityID: capabilityID, EAOwner: owner}))
}

func (l *landscape) own(component componentID, o ownership) {
	l.must(l.caches.SetOwnership(l.ctx, home.CachedOwnership{ComponentID: string(component), State: o.state, OwnerKind: o.kind, OwnerID: o.owner}))
}

func (l *landscape) grant(target artifact, email string, expiresAt time.Time) string {
	id := uuid.NewString()
	l.must(l.caches.SaveEditGrant(l.ctx, home.CachedEditGrant{ID: id, ArtifactType: target.kind, ArtifactID: target.id, GranteeEmail: email, ExpiresAt: expiresAt}))
	return id
}

func (l *landscape) steward(domainID home.DomainID, concern string, steward person) string {
	id := uuid.NewString()
	l.must(l.stewards.Upsert(l.ctx, stewardshipRM.StewardshipRow{ID: id, DomainID: string(domainID), Concern: concern, StewardID: steward.id, AssignedBy: "admin@example.com", AssignedAt: time.Now()}))
	return id
}

func (l *landscape) home(caller home.Caller) *home.Home {
	l.t.Helper()
	composed, err := home.NewView(l.queries, time.Now).Compose(l.ctx, caller)
	require.NoError(l.t, err)
	return composed
}

func (l *landscape) anchorsAt(caller home.Caller, now time.Time) home.Anchors {
	l.t.Helper()
	facts, err := l.queries.CallerFacts(l.ctx, caller, now)
	require.NoError(l.t, err)
	return facts.Anchors()
}

type person struct {
	id    string
	email string
}

func newPerson() person {
	id := uuid.NewString()
	return person{id: id, email: id + "@example.com"}
}

func (p person) caller() home.Caller {
	return home.Caller{UserID: p.id, Email: p.email}
}

type eventHandler interface {
	Handle(ctx context.Context, event domain.DomainEvent) error
}

func storedEvent(t *testing.T, eventType string, data map[string]any) domain.DomainEvent {
	t.Helper()
	payload, err := json.Marshal(data)
	require.NoError(t, err)
	return domain.NewGenericDomainEvent("aggregate", eventType, payload, time.Now().UTC())
}
