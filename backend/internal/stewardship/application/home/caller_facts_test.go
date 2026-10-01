package home

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCallerFacts_EveryKindOfFactAnchorsWithItsRelation(t *testing.T) {
	facts := CallerFacts{
		Stewardships:        []StewardshipAnchor{{Domain: Domain{ID: "d1", Name: "Finance"}, Concern: "assessment"}},
		ArchitectedDomains:  []Domain{{ID: "d2", Name: "Sales"}},
		EAOwnedCapabilities: []OwnedCapability{{ID: "c1", Name: "Invoicing", Level: "L2"}},
		Ownerships: []ApplicationOwnership{
			{ApplicationID: "a1", Name: "CRM", State: "owned"},
			{ApplicationID: "a2", Name: "Portal", State: "nominated"},
		},
		Grants: []EditGrant{
			{Subject: billing.subject, Name: "Billing", Level: "L2", ExpiresAt: grantExpiry},
			{Subject: archive.subject, Name: "Archive", ExpiresAt: grantExpiry},
		},
	}

	anchors := facts.Anchors()

	assert.Equal(t, facts.Stewardships, anchors.Stewardships)
	assert.Equal(t, facts.ArchitectedDomains, anchors.ArchitectedDomains)
	assert.Equal(t, []SubjectAnchor{
		invoicing.as(RelationEAOwner),
		crm.as(RelationOwner),
		portal.as(RelationNominated),
		billing.grantedUntil(grantExpiry),
		archive.grantedUntil(grantExpiry),
	}, anchors.Subjects)
}

func TestCallerFacts_OnlyOwnedAndNominatedApplicationsAnchor(t *testing.T) {
	for _, state := range []string{"managed", "unknown", ""} {
		t.Run(state, func(t *testing.T) {
			facts := CallerFacts{Ownerships: []ApplicationOwnership{{ApplicationID: "a1", Name: "CRM", State: state}}}

			assert.Empty(t, facts.Anchors().Subjects)
		})
	}
}

func TestCallerFacts_NothingKnownOfTheCallerIsNoAnchor(t *testing.T) {
	assert.Equal(t, ScopeKindEmpty, CallerFacts{}.Anchors().scopeKind(Caller{}))
}
