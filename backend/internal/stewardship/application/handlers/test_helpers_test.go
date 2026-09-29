package handlers

import (
	"context"
	"errors"

	"easi/backend/internal/stewardship/domain/aggregates"
	"easi/backend/internal/stewardship/domain/valueobjects"
)

var errNotStored = errors.New("not stored")

type fakeRepository struct {
	byID  map[string]*aggregates.Stewardship
	saves int
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{byID: map[string]*aggregates.Stewardship{}}
}

func (r *fakeRepository) Save(_ context.Context, s *aggregates.Stewardship) error {
	if len(s.GetUncommittedChanges()) == 0 {
		return nil
	}
	r.saves++
	r.byID[s.ID()] = s
	return nil
}

func (r *fakeRepository) GetByID(_ context.Context, id string) (*aggregates.Stewardship, error) {
	s, ok := r.byID[id]
	if !ok {
		return nil, errNotStored
	}
	s.MarkChangesAsCommitted()
	return s, nil
}

func (r *fakeRepository) FindLiveStewardshipID(_ context.Context, domainID, concern string) (string, bool, error) {
	for id, s := range r.byID {
		if s.IsReleased() {
			continue
		}
		if s.DomainID() == domainID && s.Concern().Value() == concern {
			return id, true, nil
		}
	}
	return "", false, nil
}

func (r *fakeRepository) only() *aggregates.Stewardship {
	for _, s := range r.byID {
		return s
	}
	return nil
}

type fakeDomains map[string]bool

func (d fakeDomains) DomainExists(_ context.Context, domainID string) (bool, error) {
	return d[domainID], nil
}

type fakeUsers map[string]valueobjects.UserStanding

func (u fakeUsers) UserStanding(_ context.Context, userID string) (valueobjects.UserStanding, error) {
	return u[userID], nil
}

type handlerFixture struct {
	repo    *fakeRepository
	domains fakeDomains
	users   fakeUsers
}

func newHandlerFixture() handlerFixture {
	return handlerFixture{
		repo:    newFakeRepository(),
		domains: fakeDomains{"domain-1": true},
		users: fakeUsers{
			"mette": valueobjects.UserActive,
			"jonas": valueobjects.UserActive,
			"gone":  valueobjects.UserDisabled,
		},
	}
}

func (f handlerFixture) assignHandler() *AssignStewardHandler {
	return NewAssignStewardHandler(AssignStewardDeps{Repository: f.repo, Lookup: f.repo, Domains: f.domains, Users: f.users})
}

func (f handlerFixture) releaseHandler() *ReleaseStewardHandler {
	return NewReleaseStewardHandler(ReleaseStewardDeps{Repository: f.repo, Lookup: f.repo, Domains: f.domains})
}
