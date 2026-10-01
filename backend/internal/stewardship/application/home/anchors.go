package home

import (
	"slices"
	"strings"
	"time"
)

type ScopeKind string

const (
	ScopeKindPersonal ScopeKind = "personal"
	ScopeKindTenant   ScopeKind = "tenant"
	ScopeKindEmpty    ScopeKind = "empty"
)

type Relation string

const (
	RelationEAOwner   Relation = "ea-owner"
	RelationOwner     Relation = "owner"
	RelationNominated Relation = "nominated"
	RelationEditGrant Relation = "edit-grant"
)

var relationOrder = []Relation{RelationEAOwner, RelationOwner, RelationNominated, RelationEditGrant}

type Caller struct {
	UserID            string
	Email             string
	MayAssignStewards bool
}

type Domain struct {
	ID   string
	Name string
}

type StewardshipAnchor struct {
	Domain  Domain
	Concern string
}

type SubjectAnchor struct {
	Subject
	Name           string
	Level          string
	Relation       Relation
	GrantExpiresAt time.Time
}

type Anchors struct {
	Stewardships       []StewardshipAnchor
	ArchitectedDomains []Domain
	Subjects           []SubjectAnchor
}

func (a Anchors) any() bool {
	return len(a.Stewardships) > 0 || len(a.ArchitectedDomains) > 0 || len(a.Subjects) > 0
}

func (a Anchors) scopeKind(caller Caller) ScopeKind {
	switch {
	case a.any():
		return ScopeKindPersonal
	case caller.MayAssignStewards:
		return ScopeKindTenant
	default:
		return ScopeKindEmpty
	}
}

func (a Anchors) seed() PortfolioSeed {
	domains := newIDSet()
	for _, stewardship := range a.Stewardships {
		domains.add(stewardship.Domain.ID)
	}
	for _, domain := range a.ArchitectedDomains {
		domains.add(domain.ID)
	}
	anchored := a.anchoredSubjects()
	return PortfolioSeed{
		DomainIDs:      domains.ids,
		CapabilityIDs:  anchored.IDsOf(SubjectCapability),
		ApplicationIDs: anchored.IDsOf(SubjectApplication),
	}
}

func (a Anchors) anchoredSubjects() Subjects {
	subjects := make(Subjects, len(a.Subjects))
	for i, anchor := range a.Subjects {
		subjects[i] = anchor.Subject
	}
	return subjects
}

func (a Anchors) count(relations ...Relation) int {
	counted := map[Subject]bool{}
	for _, anchor := range a.Subjects {
		if slices.Contains(relations, anchor.Relation) {
			counted[anchor.Subject] = true
		}
	}
	return len(counted)
}

func (a Anchors) workItems() []WorkItem {
	ordered := slices.Clone(a.Subjects)
	slices.SortStableFunc(ordered, compareSubjects)
	listed := map[Subject]bool{}
	items := make([]WorkItem, 0, len(ordered))
	for _, anchor := range ordered {
		if !listed[anchor.Subject] {
			listed[anchor.Subject] = true
			items = append(items, anchor.workItem())
		}
	}
	return items
}

func compareSubjects(x, y SubjectAnchor) int {
	if byRelation := slices.Index(relationOrder, x.Relation) - slices.Index(relationOrder, y.Relation); byRelation != 0 {
		return byRelation
	}
	if byName := strings.Compare(strings.ToLower(x.Name), strings.ToLower(y.Name)); byName != 0 {
		return byName
	}
	if byID := strings.Compare(x.ID, y.ID); byID != 0 {
		return byID
	}
	return y.GrantExpiresAt.Compare(x.GrantExpiresAt)
}

func (s SubjectAnchor) workItem() WorkItem {
	item := WorkItem{SubjectType: s.Type, ID: s.ID, Name: s.Name, Level: s.Level, Relation: s.Relation}
	if s.Relation == RelationEditGrant {
		item.GrantExpiresOn = s.GrantExpiresAt.UTC().Format(time.DateOnly)
	}
	return item
}

type idSet struct {
	seen map[string]bool
	ids  []string
}

func newIDSet() *idSet {
	return &idSet{seen: map[string]bool{}, ids: []string{}}
}

func (s *idSet) add(id string) {
	if s.seen[id] {
		return
	}
	s.seen[id] = true
	s.ids = append(s.ids, id)
}
