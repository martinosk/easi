//go:build integration

package readmodels

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"easi/backend/internal/stewardship/application/home"
)

type background struct {
	engagement, finance          home.DomainID
	customerManagement, profiles home.CapabilityID
	invoicing, ledger            home.CapabilityID
	crm, portal, billing, erp    home.ComponentID
}

func seedBackground(l *landscape) background {
	b := background{engagement: l.domain("Customer Engagement"), finance: l.domain("Finance")}
	b.customerManagement = l.l1In(b.engagement, "Customer Management")
	b.profiles = l.capability("Customer Profiles", "L2", b.customerManagement)
	planned := l.capability("Loyalty", "L2", b.customerManagement)
	l.must(l.caches.SetCapabilityMetadata(l.ctx, home.CachedCapabilityMetadata{CapabilityID: planned, Status: "Planned"}))
	deprecated := l.capability("Segmentation", "L3", b.profiles)
	l.must(l.caches.SetCapabilityMetadata(l.ctx, home.CachedCapabilityMetadata{CapabilityID: deprecated, Status: "Deprecated"}))
	accounting := l.l1In(b.finance, "Accounting")
	b.invoicing = l.capability("Invoicing", "L2", accounting)
	b.ledger = l.capability("Ledger", "L2", accounting)

	b.crm, b.portal = l.application("CRM"), l.application("Portal")
	b.billing, b.erp = l.application("Billing"), l.application("ERP")
	l.realise(b.profiles, b.crm)
	l.realise(deprecated, b.portal)
	l.realise(b.invoicing, b.billing)
	l.realise(b.ledger, b.erp)
	return b
}

func TestHome_StewardCoversTheirDomain(t *testing.T) {
	l := newLandscape(t)
	b := seedBackground(l)
	mette := newPerson()
	l.steward(b.engagement, "assessment", mette)

	got := l.home(mette.caller())

	assert.Equal(t, home.ScopeKindPersonal, got.Scope.Kind)
	require.Len(t, got.Scope.Stewardships, 1)
	assert.Equal(t, "Customer Engagement", got.Scope.Stewardships[0].Domain.Name)
	assert.Equal(t, "Assessment", got.Scope.Stewardships[0].ConcernLabel)
	assert.Equal(t, &home.CapabilityTile{Total: 4, ByStatus: home.StatusBreakdown{Active: 2, Planned: 1, Deprecated: 1}}, got.Portfolio.Capabilities)
	assert.Equal(t, 2, got.Portfolio.Applications.Total)
	assert.Equal(t, &home.DomainTile{Total: 1, Names: []string{"Customer Engagement"}}, got.Portfolio.Domains)
	assert.Equal(t, &home.MyWork{Items: []home.WorkItem{}}, got.MyWork)
}

func TestHome_DomainArchitectCoversTheirDomain(t *testing.T) {
	l := newLandscape(t)
	b := seedBackground(l)
	alice := newPerson()
	l.architect(b.finance, "Finance", alice)

	got := l.home(alice.caller())

	assert.Equal(t, []home.DomainRef{{ID: string(b.finance), Name: "Finance"}}, got.Scope.ArchitectedDomains)
	assert.Equal(t, 3, got.Portfolio.Capabilities.Total)
	assert.Equal(t, 2, got.Portfolio.Applications.Total)
}

func TestHome_EAOwnerAnchorsTheCapabilityOnly(t *testing.T) {
	l := newLandscape(t)
	b := seedBackground(l)
	jonas := newPerson()
	l.eaOwner(b.invoicing, jonas.id)

	got := l.home(jonas.caller())

	assert.Equal(t, 1, got.Portfolio.Capabilities.Total)
	assert.Equal(t, 1, got.Portfolio.Applications.Total)
	assert.Equal(t, &home.DomainTile{Total: 1, Names: []string{"Finance"}}, got.Portfolio.Domains)
	assert.Equal(t, []home.WorkItem{{SubjectType: home.SubjectCapability, ID: string(b.invoicing), Name: "Invoicing", Level: "L2", Relation: home.RelationEAOwner}}, got.MyWork.Items)
}

func TestHome_LegacyFreeTextEAOwnerMatchesNobody(t *testing.T) {
	l := newLandscape(t)
	b := seedBackground(l)
	jonas := newPerson()
	l.eaOwner(b.invoicing, "Jonas Holm")

	assert.Equal(t, home.ScopeKindEmpty, l.home(jonas.caller()).Scope.Kind)
}

func TestHome_ApplicationOwnerSeesEveryDomainTheApplicationRealises(t *testing.T) {
	l := newLandscape(t)
	b := seedBackground(l)
	mette := newPerson()
	l.realise(b.invoicing, b.crm)
	l.grade(b.profiles, b.crm, "Invest")
	l.own(b.crm, ownership{state: "owned", kind: "user", owner: mette.id})

	got := l.home(mette.caller())

	assert.Equal(t, 0, got.Portfolio.Capabilities.Total)
	assert.Equal(t, 1, got.Portfolio.Applications.Total)
	assert.Equal(t, &home.DomainTile{Total: 2, Names: []string{"Customer Engagement", "Finance"}}, got.Portfolio.Domains)
	assert.Equal(t, &home.TimeTile{Total: 2, Shares: &home.TimeShares{Invest: 50, NotAssessed: 50}}, got.Portfolio.Time)
	assert.Equal(t, 1, got.Scope.OwnedApplications)
	assert.Equal(t, []home.WorkItem{{SubjectType: home.SubjectApplication, ID: string(b.crm), Name: "CRM", Relation: home.RelationOwner, DominantGrade: home.GradeInvest}}, got.MyWork.Items)
}

func TestHome_StewardTimeTileStaysInsideTheDomain(t *testing.T) {
	l := newLandscape(t)
	b := seedBackground(l)
	mette := newPerson()
	l.steward(b.engagement, "assessment", mette)
	l.realise(b.invoicing, b.crm)
	l.grade(b.profiles, b.crm, "Tolerate")
	l.grade(b.invoicing, b.crm, "Eliminate")

	got := l.home(mette.caller())

	assert.Equal(t, &home.TimeTile{Total: 2, Shares: &home.TimeShares{Tolerate: 50, NotAssessed: 50}}, got.Portfolio.Time)
	assert.Equal(t, 2, got.Portfolio.Applications.Total)
}

func TestHome_NominatedOwnerAndTeamOwnership(t *testing.T) {
	l := newLandscape(t)
	b := seedBackground(l)
	mette := newPerson()
	l.own(b.portal, ownership{state: "nominated", kind: "user", owner: mette.id})
	l.own(b.crm, ownership{state: "managed", kind: "team", owner: mette.id})
	l.own(b.billing, ownership{state: "nominated", kind: "team", owner: mette.id})

	got := l.home(mette.caller())

	assert.Equal(t, []home.WorkItem{{SubjectType: home.SubjectApplication, ID: string(b.portal), Name: "Portal", Relation: home.RelationNominated}}, got.MyWork.Items)
}

func TestHome_EditGrantOnApplicationIsAnAnchor(t *testing.T) {
	l := newLandscape(t)
	b := seedBackground(l)
	ole := newPerson()
	expiry := time.Date(2099, 10, 21, 8, 0, 0, 0, time.UTC)
	l.grant(componentArtifact(b.crm), "Ole.Berg+"+ole.id+"@Example.com", expiry)

	got := l.home(home.Caller{UserID: ole.id, Email: "ole.BERG+" + ole.id + "@example.COM"})

	assert.Equal(t, home.ScopeKindPersonal, got.Scope.Kind)
	assert.Equal(t, 1, got.Scope.EditGrants)
	assert.Equal(t, 1, got.Portfolio.Applications.Total)
	assert.Equal(t, []home.WorkItem{{SubjectType: home.SubjectApplication, ID: string(b.crm), Name: "CRM", Relation: home.RelationEditGrant, GrantExpiresOn: "2099-10-21"}}, got.MyWork.Items)
}

func TestAnchors_GrantExclusions(t *testing.T) {
	l := newLandscape(t)
	b := seedBackground(l)
	now := time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name  string
		setup func(email string)
	}{
		{"expired", func(email string) { l.grant(componentArtifact(b.crm), email, now.Add(-24*time.Hour)) }},
		{"expiring exactly now", func(email string) { l.grant(componentArtifact(b.crm), email, now) }},
		{"revoked", func(email string) {
			l.must(l.caches.DeleteEditGrant(l.ctx, l.grant(componentArtifact(b.crm), email, now.Add(time.Hour))))
		}},
		{"other artifact type", func(email string) { l.grant(artifact{kind: "view", id: string(b.crm)}, email, now.Add(time.Hour)) }},
		{"deleted capability", func(email string) {
			gone := l.capability("Gone", "L1", "")
			l.grant(capabilityArtifact(gone), email, now.Add(time.Hour))
			l.must(l.caches.DeleteCapability(l.ctx, gone))
		}},
		{"deleted application", func(email string) {
			gone := l.application("Legacy Portal")
			l.grant(componentArtifact(gone), email, now.Add(time.Hour))
			l.must(l.caches.DeleteApplication(l.ctx, gone))
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ole := newPerson()
			tt.setup(ole.email)

			assert.Empty(t, l.anchorsAt(ole.caller(), now).Subjects)
		})
	}
}

func TestAnchors_ActorWithoutEmailHasNoGrantAnchors(t *testing.T) {
	l := newLandscape(t)
	b := seedBackground(l)
	agent := newPerson()
	l.grant(componentArtifact(b.crm), agent.email, time.Now().Add(time.Hour))

	assert.Empty(t, l.anchorsAt(home.Caller{UserID: agent.id}, time.Now()).Subjects)
}

func TestAnchors_ReleasedStewardshipIsNoAnchor(t *testing.T) {
	l := newLandscape(t)
	b := seedBackground(l)
	mette := newPerson()
	l.must(l.stewards.Delete(l.ctx, l.steward(b.engagement, "assessment", mette)))

	assert.Equal(t, home.ScopeKindEmpty, l.home(mette.caller()).Scope.Kind)
}

func TestHome_CapabilityInSeveralDomains(t *testing.T) {
	l := newLandscape(t)
	b := seedBackground(l)
	l.must(l.caches.AssignToDomain(l.ctx, home.CachedAssignment{CapabilityID: b.customerManagement, DomainID: b.finance}))
	alice := newPerson()
	l.architect(b.finance, "Finance", alice)
	jonas := newPerson()
	l.eaOwner(b.profiles, jonas.id)

	assert.Equal(t, 7, l.home(alice.caller()).Portfolio.Capabilities.Total)
	assert.Equal(t, &home.DomainTile{Total: 2, Names: []string{"Customer Engagement", "Finance"}}, l.home(jonas.caller()).Portfolio.Domains)
}

func TestHome_ReparentedCapabilityFollowsItsNewL1(t *testing.T) {
	l := newLandscape(t)
	b := seedBackground(l)
	jonas := newPerson()
	l.eaOwner(b.profiles, jonas.id)
	accounting := l.l1In(b.finance, "Treasury")
	l.must(l.caches.MoveCapability(l.ctx, home.CachedPlacement{CapabilityID: b.profiles, ParentID: accounting, Level: "L2"}))
	mette := newPerson()
	l.steward(b.engagement, "structure", mette)

	assert.Equal(t, []string{"Finance"}, l.home(jonas.caller()).Portfolio.Domains.Names)
	assert.Equal(t, 2, l.home(mette.caller()).Portfolio.Capabilities.Total)
}

func TestHome_InheritedRealisationsAreNotCounted(t *testing.T) {
	l := newLandscape(t)
	b := seedBackground(l)
	alice := newPerson()
	l.architect(b.engagement, "Customer Engagement", alice)

	assert.Equal(t, 2, l.home(alice.caller()).Portfolio.Time.Total)
}

func TestHome_DomainTileNamesFiveAlphabetically(t *testing.T) {
	l := newLandscape(t)
	app := l.application("Hub")
	for _, name := range []string{"Foxtrot", "Alpha", "Echo", "Charlie", "Delta", "Bravo"} {
		l.realise(l.l1In(l.domain(name), name+" L1"), app)
	}
	owner := newPerson()
	l.own(app, ownership{state: "owned", kind: "user", owner: owner.id})

	assert.Equal(t, &home.DomainTile{Total: 6, Names: []string{"Alpha", "Bravo", "Charlie", "Delta", "Echo"}}, l.home(owner.caller()).Portfolio.Domains)
}

func TestHome_TenantAndEmptyKinds(t *testing.T) {
	l := newLandscape(t)
	seedBackground(l)
	per := newPerson()
	tenantCaller := per.caller()
	tenantCaller.MayAssignStewards = true

	tenantHome := l.home(tenantCaller)
	emptyHome := l.home(newPerson().caller())

	assert.Equal(t, home.ScopeKindTenant, tenantHome.Scope.Kind)
	assert.Equal(t, 7, tenantHome.Portfolio.Capabilities.Total)
	assert.Equal(t, 4, tenantHome.Portfolio.Applications.Total)
	assert.Equal(t, 2, tenantHome.Portfolio.Domains.Total)
	assert.Equal(t, 4, tenantHome.Portfolio.Time.Total)
	assert.Equal(t, &home.MyWork{Items: []home.WorkItem{}}, tenantHome.MyWork)
	assert.Equal(t, home.ScopeKindEmpty, emptyHome.Scope.Kind)
	assert.Nil(t, emptyHome.Portfolio)
	assert.Nil(t, emptyHome.MyWork)
}

func TestHome_ArchitectWithAnAnchorSeesOnlyTheirScope(t *testing.T) {
	l := newLandscape(t)
	b := seedBackground(l)
	per := newPerson()
	l.eaOwner(b.invoicing, per.id)
	caller := per.caller()
	caller.MayAssignStewards = true

	got := l.home(caller)

	assert.Equal(t, home.ScopeKindPersonal, got.Scope.Kind)
	assert.Equal(t, 1, got.Portfolio.Capabilities.Total)
}

func TestHome_TimeIgnoresAssessmentsOfDeletedRealisationsAndCountsStaleGrades(t *testing.T) {
	l := newLandscape(t)
	b := seedBackground(l)
	alice := newPerson()
	l.architect(b.finance, "Finance", alice)
	l.must(l.caches.SaveTimeAssessment(l.ctx, home.CachedTimeAssessment{CapabilityID: b.invoicing, ComponentID: b.billing, Grade: "Migrate", AssessedAt: time.Now().AddDate(-3, 0, 0)}))
	l.grade(b.invoicing, b.erp, "Invest")

	assert.Equal(t, &home.TimeTile{Total: 2, Shares: &home.TimeShares{Migrate: 50, NotAssessed: 50}}, l.home(alice.caller()).Portfolio.Time)
}

func TestHome_AGradeTheReadSideDoesNotKnowFailsInsteadOfBeingDropped(t *testing.T) {
	l := newLandscape(t)
	b := seedBackground(l)
	mette := newPerson()
	l.steward(b.engagement, "assessment", mette)
	l.grade(b.profiles, b.crm, "Retire")

	_, err := home.NewView(l.queries, time.Now).Compose(l.ctx, mette.caller())

	assert.ErrorIs(t, err, home.ErrUnknownGrade)
}

func TestHome_MyWorkDedupesGrantedSubjects(t *testing.T) {
	l := newLandscape(t)
	b := seedBackground(l)
	per := newPerson()
	l.eaOwner(b.invoicing, per.id)
	l.own(b.crm, ownership{state: "owned", kind: "user", owner: per.id})
	l.grant(capabilityArtifact(b.invoicing), per.email, time.Now().Add(time.Hour))
	l.grant(componentArtifact(b.crm), per.email, time.Now().Add(time.Hour))
	l.grant(componentArtifact(b.erp), per.email, time.Now().Add(time.Hour))

	got := l.home(per.caller())

	assert.Equal(t, 3, got.MyWork.Total)
	assert.Equal(t, []string{"Invoicing", "CRM", "ERP"}, itemNames(got.MyWork.Items))
	assert.Equal(t, []home.Relation{home.RelationEAOwner, home.RelationOwner, home.RelationEditGrant}, itemRelations(got.MyWork.Items))
}

func TestHome_TenantIsolation(t *testing.T) {
	l := newLandscape(t)
	shared := newPerson()
	b := seedBackground(l)
	l.eaOwner(b.invoicing, shared.id)
	tenantA := l.tenant

	l.useTenant("home-it-" + strings.ToLower(shared.id[:8]) + "b")
	other := seedBackground(l)
	l.eaOwner(other.ledger, shared.id)
	l.own(other.crm, ownership{state: "owned", kind: "user", owner: shared.id})
	l.grant(componentArtifact(other.erp), shared.email, time.Now().Add(time.Hour))
	l.steward(other.engagement, "assessment", shared)
	l.useTenant(tenantA)

	got := l.home(shared.caller())

	assert.Equal(t, []string{"Invoicing"}, itemNames(got.MyWork.Items))
	assert.Empty(t, got.Scope.Stewardships)
	assert.Equal(t, 1, got.Portfolio.Capabilities.Total)
}

func itemNames(items []home.WorkItem) []string {
	names := make([]string, len(items))
	for i, item := range items {
		names[i] = item.Name
	}
	return names
}

func itemRelations(items []home.WorkItem) []home.Relation {
	relations := make([]home.Relation, len(items))
	for i, item := range items {
		relations[i] = item.Relation
	}
	return relations
}
