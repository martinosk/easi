package home

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var grantExpiry = time.Date(2026, 10, 21, 23, 30, 0, 0, time.UTC)

type subjectRef struct {
	subject Subject
	name    string
}

var (
	accounting = subjectRef{Subject{Type: SubjectCapability, ID: "c0"}, "Accounting"}
	invoicing  = subjectRef{Subject{Type: SubjectCapability, ID: "c1"}, "Invoicing"}
	billing    = subjectRef{Subject{Type: SubjectCapability, ID: "c2"}, "Billing"}
	crm        = subjectRef{Subject{Type: SubjectApplication, ID: "a1"}, "CRM"}
	portal     = subjectRef{Subject{Type: SubjectApplication, ID: "a2"}, "Portal"}
	archive    = subjectRef{Subject{Type: SubjectApplication, ID: "a3"}, "Archive"}
)

func (r subjectRef) as(relation Relation) SubjectAnchor {
	anchor := SubjectAnchor{Subject: r.subject, Name: r.name, Relation: relation}
	if r.subject.Type == SubjectCapability {
		anchor.Level = "L2"
	}
	return anchor
}

func (r subjectRef) grantedUntil(expiresAt time.Time) SubjectAnchor {
	anchor := r.as(RelationEditGrant)
	anchor.GrantExpiresAt = expiresAt
	return anchor
}

func TestScopeKind(t *testing.T) {
	anchored := Anchors{Subjects: []SubjectAnchor{invoicing.as(RelationEAOwner)}}
	tests := []struct {
		name    string
		anchors Anchors
		caller  Caller
		want    ScopeKind
	}{
		{"anchor makes it personal", anchored, Caller{}, ScopeKindPersonal},
		{"anchor wins over domains:write", anchored, Caller{MayAssignStewards: true}, ScopeKindPersonal},
		{"stewardship is an anchor", Anchors{Stewardships: []StewardshipAnchor{{Domain: Domain{ID: "d1"}, Concern: "assessment"}}}, Caller{}, ScopeKindPersonal},
		{"architected domain is an anchor", Anchors{ArchitectedDomains: []Domain{{ID: "d1"}}}, Caller{}, ScopeKindPersonal},
		{"no anchor with domains:write", Anchors{}, Caller{MayAssignStewards: true}, ScopeKindTenant},
		{"no anchor without domains:write", Anchors{}, Caller{}, ScopeKindEmpty},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.anchors.scopeKind(tt.caller))
		})
	}
}

func TestAnchors_SeedCollectsScopeDomainsAndAnchoredSubjects(t *testing.T) {
	anchors := Anchors{
		Stewardships: []StewardshipAnchor{
			{Domain: Domain{ID: "d1"}, Concern: "assessment"},
			{Domain: Domain{ID: "d1"}, Concern: "ownership"},
		},
		ArchitectedDomains: []Domain{{ID: "d2"}},
		Subjects: []SubjectAnchor{
			invoicing.as(RelationEAOwner),
			billing.grantedUntil(grantExpiry),
			crm.as(RelationOwner),
			portal.as(RelationNominated),
			crm.grantedUntil(grantExpiry),
		},
	}

	seed := anchors.seed()

	assert.ElementsMatch(t, []string{"d1", "d2"}, seed.DomainIDs)
	assert.ElementsMatch(t, []string{"c1", "c2"}, seed.CapabilityIDs)
	assert.ElementsMatch(t, []string{"a1", "a2"}, seed.ApplicationIDs)
}

func TestWorkItems_DedupeOrderAndRelations(t *testing.T) {
	anchors := Anchors{Subjects: []SubjectAnchor{
		invoicing.grantedUntil(grantExpiry),
		crm.grantedUntil(grantExpiry),
		archive.grantedUntil(grantExpiry),
		portal.as(RelationNominated),
		crm.as(RelationOwner),
		invoicing.as(RelationEAOwner),
		accounting.as(RelationEAOwner),
	}}

	items := anchors.workItems()

	assert.Equal(t, []string{"Accounting", "Invoicing", "CRM", "Portal", "Archive"}, itemNames(items))
	assert.Equal(t, []Relation{RelationEAOwner, RelationEAOwner, RelationOwner, RelationNominated, RelationEditGrant}, itemRelations(items))
	assert.Equal(t, "2026-10-21", items[4].GrantExpiresOn)
	assert.Empty(t, items[2].GrantExpiresOn)
	assert.Equal(t, "L2", items[0].Level)
}

func TestWorkItems_RepeatedGrantShowsTheLatestExpiry(t *testing.T) {
	anchors := Anchors{Subjects: []SubjectAnchor{
		crm.grantedUntil(grantExpiry),
		crm.grantedUntil(grantExpiry.AddDate(0, 0, 3)),
	}}

	items := anchors.workItems()

	assert.Len(t, items, 1)
	assert.Equal(t, "2026-10-24", items[0].GrantExpiresOn)
}

func TestAnchors_CountsForTheScope(t *testing.T) {
	anchors := Anchors{Subjects: []SubjectAnchor{
		invoicing.as(RelationEAOwner),
		crm.as(RelationOwner),
		portal.as(RelationNominated),
		crm.grantedUntil(grantExpiry),
		crm.grantedUntil(grantExpiry.AddDate(0, 0, 1)),
	}}

	assert.Equal(t, 1, anchors.count(RelationEAOwner))
	assert.Equal(t, 2, anchors.count(RelationOwner, RelationNominated))
	assert.Equal(t, 1, anchors.count(RelationEditGrant))
}

func itemNames(items []WorkItem) []string {
	names := make([]string, len(items))
	for i, item := range items {
		names[i] = item.Name
	}
	return names
}

func itemRelations(items []WorkItem) []Relation {
	relations := make([]Relation, len(items))
	for i, item := range items {
		relations[i] = item.Relation
	}
	return relations
}
