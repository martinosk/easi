package home

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var readAll = Readable{Capabilities: true, Applications: true, Direction: true, Domains: true}

func fullHome() *Home {
	return &Home{
		Scope: Scope{
			Kind:                   ScopeKindPersonal,
			Stewardships:           []ScopeStewardship{{Domain: &DomainRef{ID: "d1", Name: "Finance"}, Concern: "assessment", ConcernLabel: "Assessment"}},
			ArchitectedDomains:     []DomainRef{{ID: "d2", Name: "Sales"}},
			ArchitectedDomainCount: 1,
		},
		Portfolio: &Portfolio{
			Capabilities: &CapabilityTile{Total: 1},
			Applications: &CountTile{Total: 1},
			Domains:      &DomainTile{Total: 2, Names: []string{"Finance", "Sales"}},
			Time:         &TimeTile{Total: 1, Shares: &TimeShares{Invest: 100}},
		},
		MyWork: &MyWork{Total: 2, Items: []WorkItem{
			{SubjectType: SubjectCapability, ID: "c1", Relation: RelationEAOwner, DominantGrade: GradeInvest},
			{SubjectType: SubjectApplication, ID: "a1", Relation: RelationOwner, DominantGrade: GradeMigrate},
		}},
	}
}

func without(change func(*Readable)) Readable {
	readable := readAll
	change(&readable)
	return readable
}

func TestRedactFor_ReaderOfEverythingKeepsEverything(t *testing.T) {
	home := fullHome()

	home.RedactFor(readAll)

	assert.Equal(t, fullHome(), home)
}

func TestRedactFor_WithoutCapabilitiesRead(t *testing.T) {
	home := fullHome()

	home.RedactFor(without(func(r *Readable) { r.Capabilities = false }))

	assert.Nil(t, home.Portfolio.Capabilities)
	assert.NotNil(t, home.Portfolio.Applications)
	assert.Equal(t, 1, home.MyWork.Total)
	assert.Equal(t, "a1", home.MyWork.Items[0].ID)
}

func TestRedactFor_WithoutComponentsRead(t *testing.T) {
	home := fullHome()

	home.RedactFor(without(func(r *Readable) { r.Applications = false }))

	assert.Nil(t, home.Portfolio.Applications)
	assert.NotNil(t, home.Portfolio.Capabilities)
	assert.Equal(t, 1, home.MyWork.Total)
	assert.Equal(t, "c1", home.MyWork.Items[0].ID)
}

func TestRedactFor_WithoutDirectionRead(t *testing.T) {
	home := fullHome()

	home.RedactFor(without(func(r *Readable) { r.Direction = false }))

	assert.Nil(t, home.Portfolio.Time)
	assert.Equal(t, 2, home.MyWork.Total)
	for _, item := range home.MyWork.Items {
		assert.Empty(t, item.DominantGrade)
	}
}

func TestRedactFor_WithoutDomainsReadKeepsCountsOnly(t *testing.T) {
	home := fullHome()

	home.RedactFor(without(func(r *Readable) { r.Domains = false }))

	assert.Nil(t, home.Scope.ArchitectedDomains)
	assert.Equal(t, 1, home.Scope.ArchitectedDomainCount)
	assert.Nil(t, home.Scope.Stewardships[0].Domain)
	assert.Equal(t, "assessment", home.Scope.Stewardships[0].Concern)
	assert.Equal(t, &DomainTile{Total: 2}, home.Portfolio.Domains)
}

func TestRedactFor_EmptyHomeHasNothingToRedact(t *testing.T) {
	home := &Home{Scope: Scope{Kind: ScopeKindEmpty, Stewardships: []ScopeStewardship{}}}

	home.RedactFor(Readable{})

	assert.Nil(t, home.Portfolio)
	assert.Nil(t, home.MyWork)
}
