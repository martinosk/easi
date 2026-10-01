package home

import (
	"context"
	"time"

	"easi/backend/internal/stewardship/domain/valueobjects"
)

type Store interface {
	CallerFacts(ctx context.Context, caller Caller, now time.Time) (CallerFacts, error)
	Portfolio(ctx context.Context, seed PortfolioSeed) (PortfolioFacts, error)
	SubjectGrades(ctx context.Context, subjects Subjects) (map[Subject]GradeCounts, error)
}

type View struct {
	store Store
	now   func() time.Time
}

func NewView(store Store, now func() time.Time) *View {
	return &View{store: store, now: now}
}

func (v *View) Compose(ctx context.Context, caller Caller) (*Home, error) {
	callerFacts, err := v.store.CallerFacts(ctx, caller, v.now())
	if err != nil {
		return nil, err
	}
	anchors := callerFacts.Anchors()
	kind := anchors.scopeKind(caller)
	home := &Home{Scope: scopeOf(kind, anchors)}
	if kind == ScopeKindEmpty {
		return home, nil
	}
	seed := anchors.seed()
	seed.All = kind == ScopeKindTenant
	portfolioFacts, err := v.store.Portfolio(ctx, seed)
	if err != nil {
		return nil, err
	}
	home.Portfolio = portfolioFacts.portfolio()
	home.MyWork, err = v.myWork(ctx, anchors)
	return home, err
}

func (v *View) myWork(ctx context.Context, anchors Anchors) (*MyWork, error) {
	items := anchors.workItems()
	if len(items) == 0 {
		return &MyWork{Items: items}, nil
	}
	subjects := make(Subjects, len(items))
	for i, item := range items {
		subjects[i] = item.subject()
	}
	grades, err := v.store.SubjectGrades(ctx, subjects)
	if err != nil {
		return nil, err
	}
	for i, subject := range subjects {
		items[i].DominantGrade = grades[subject].dominant()
	}
	return &MyWork{Total: len(items), Items: items}, nil
}

func scopeOf(kind ScopeKind, anchors Anchors) Scope {
	stewardships := make([]ScopeStewardship, len(anchors.Stewardships))
	for i, anchor := range anchors.Stewardships {
		domain := DomainRef(anchor.Domain)
		stewardships[i] = ScopeStewardship{Domain: &domain, Concern: anchor.Concern, ConcernLabel: concernLabel(anchor.Concern)}
	}
	return Scope{
		Kind:                   kind,
		Stewardships:           stewardships,
		ArchitectedDomains:     domainRefs(anchors.ArchitectedDomains),
		ArchitectedDomainCount: len(anchors.ArchitectedDomains),
		EAOwnedCapabilities:    anchors.count(RelationEAOwner),
		OwnedApplications:      anchors.count(RelationOwner, RelationNominated),
		EditGrants:             anchors.count(RelationEditGrant),
	}
}

func concernLabel(value string) string {
	concern, err := valueobjects.NewConcern(value)
	if err != nil {
		return value
	}
	return concern.Label()
}

func domainRefs(domains []Domain) []DomainRef {
	refs := make([]DomainRef, len(domains))
	for i, domain := range domains {
		refs[i] = DomainRef(domain)
	}
	return refs
}
