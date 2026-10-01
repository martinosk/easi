package handlers

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	domain "easi/backend/internal/shared/eventsourcing"
	"easi/backend/internal/stewardship/domain/aggregates"
	"easi/backend/internal/stewardship/domain/events"
	"easi/backend/internal/stewardship/domain/valueobjects"
)

var errNotStored = errors.New("not stored")

type fakeRepository struct {
	byID     map[string]*aggregates.Stewardship
	live     map[string]string
	recorded []domain.DomainEvent
	saves    int
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{byID: map[string]*aggregates.Stewardship{}, live: map[string]string{}}
}

func (r *fakeRepository) Save(_ context.Context, s *aggregates.Stewardship) error {
	raised := s.GetUncommittedChanges()
	if len(raised) == 0 {
		return nil
	}
	r.saves++
	r.byID[s.ID()] = s
	for _, event := range raised {
		r.record(event)
	}
	s.MarkChangesAsCommitted()
	return nil
}

func (r *fakeRepository) record(event domain.DomainEvent) {
	r.recorded = append(r.recorded, event)
	switch recorded := event.(type) {
	case events.StewardAssigned:
		r.live[recorded.DomainID+"/"+recorded.Concern] = recorded.ID
	case events.StewardReleased:
		delete(r.live, recorded.DomainID+"/"+recorded.Concern)
	}
}

func (r *fakeRepository) GetByID(_ context.Context, id string) (*aggregates.Stewardship, error) {
	s, ok := r.byID[id]
	if !ok {
		return nil, errNotStored
	}
	return s, nil
}

func (r *fakeRepository) FindLiveStewardshipID(_ context.Context, domainID, concern string) (string, bool, error) {
	id, live := r.live[domainID+"/"+concern]
	return id, live, nil
}

func (r *fakeRepository) lastAssigned(t *testing.T) events.StewardAssigned {
	t.Helper()
	require.NotEmpty(t, r.recorded)
	assigned, ok := r.recorded[len(r.recorded)-1].(events.StewardAssigned)
	require.True(t, ok)
	return assigned
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
