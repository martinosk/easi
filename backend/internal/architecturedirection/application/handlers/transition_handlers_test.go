package handlers

import (
	"context"
	"errors"
	"testing"

	"easi/backend/internal/architecturedirection/application/commands"
	"easi/backend/internal/architecturedirection/domain/aggregates"
	"easi/backend/internal/architecturedirection/domain/valueobjects"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeRepoWithLoad struct {
	direction *aggregates.Direction
	loadErr   error
	saveErr   error
	saved     []*aggregates.Direction
}

func (f *fakeRepoWithLoad) Save(_ context.Context, d *aggregates.Direction) error {
	if f.saveErr != nil {
		return f.saveErr
	}
	f.saved = append(f.saved, d)
	return nil
}

func (f *fakeRepoWithLoad) GetByID(_ context.Context, _ string) (*aggregates.Direction, error) {
	if f.loadErr != nil {
		return nil, f.loadErr
	}
	return f.direction, nil
}

func draftFixture(t *testing.T) *aggregates.Direction {
	t.Helper()
	ec, err := valueobjects.NewEnterpriseCapabilityRef(uuid.New().String())
	require.NoError(t, err)
	dt, err := valueobjects.NewDirectionType("consolidate")
	require.NoError(t, err)
	r1, err := valueobjects.NewPhysicalCapabilityRef(uuid.New().String())
	require.NoError(t, err)
	r2, err := valueobjects.NewPhysicalCapabilityRef(uuid.New().String())
	require.NoError(t, err)
	p, err := valueobjects.NewPlacement(uuid.New().String(), "")
	require.NoError(t, err)
	h, err := valueobjects.NewHorizon("next")
	require.NoError(t, err)
	n, err := sharedvo.NewDescription("Some narrative.")
	require.NoError(t, err)
	d, err := aggregates.DraftDirection(aggregates.DraftParams{
		EnterpriseCapabilityID: ec,
		Type:                   dt,
		SourceCapabilityIDs:    []valueobjects.PhysicalCapabilityRef{r1, r2},
		Placements:             []valueobjects.Placement{p},
		Horizon:                h,
		Narrative:              n,
	})
	require.NoError(t, err)
	d.MarkChangesAsCommitted()
	return d
}

type statusCase struct {
	name      string
	target    string
	expectErr error
	expectFn  func(*aggregates.Direction) bool
}

func TestAdvanceDirectionHandler(t *testing.T) {
	cases := []statusCase{
		{name: "draft to proposed", target: "proposed", expectFn: func(d *aggregates.Direction) bool { return d.Status().IsProposed() }},
		{name: "draft to agreed rejected", target: "agreed", expectErr: aggregates.ErrInvalidStatusTransition},
		{name: "unknown target rejected", target: "rejected", expectErr: ErrUnknownAdvanceTarget},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			d := draftFixture(t)
			repo := &fakeRepoWithLoad{direction: d}
			_, err := NewAdvanceDirectionHandler(repo).Handle(context.Background(),
				&commands.AdvanceDirection{DirectionID: d.ID(), TargetStatus: c.target})
			if c.expectErr != nil {
				assert.ErrorIs(t, err, c.expectErr)
				return
			}
			require.NoError(t, err)
			require.Len(t, repo.saved, 1)
			assert.True(t, c.expectFn(repo.saved[0]))
		})
	}
}

func TestAdvanceDirectionHandler_NotFound(t *testing.T) {
	repo := &fakeRepoWithLoad{loadErr: errors.New("not found")}
	_, err := NewAdvanceDirectionHandler(repo).Handle(context.Background(), &commands.AdvanceDirection{
		DirectionID:  uuid.New().String(),
		TargetStatus: "proposed",
	})
	assert.Error(t, err)
}

func TestRejectDirectionHandler_FromDraft(t *testing.T) {
	d := draftFixture(t)
	repo := &fakeRepoWithLoad{direction: d}
	_, err := NewRejectDirectionHandler(repo).Handle(context.Background(), &commands.RejectDirection{DirectionID: d.ID()})
	require.NoError(t, err)
	require.Len(t, repo.saved, 1)
	assert.True(t, repo.saved[0].Status().IsRejected())
}

func TestUpdateDirectionHandler_PerField(t *testing.T) {
	newPhysIDs := []string{uuid.New().String(), uuid.New().String(), uuid.New().String()}
	newDomID := uuid.New().String()
	narrative := "Refined narrative."
	horizon := "later"

	type fieldCase struct {
		name      string
		command   func(directionID string) *commands.UpdateDirection
		assertion func(*testing.T, *aggregates.Direction)
	}

	cases := []fieldCase{
		{
			name: "narrative only",
			command: func(id string) *commands.UpdateDirection {
				return &commands.UpdateDirection{DirectionID: id, Narrative: &narrative}
			},
			assertion: func(t *testing.T, d *aggregates.Direction) {
				assert.Equal(t, narrative, d.Narrative().Value())
			},
		},
		{
			name: "horizon only",
			command: func(id string) *commands.UpdateDirection {
				return &commands.UpdateDirection{DirectionID: id, Horizon: &horizon}
			},
			assertion: func(t *testing.T, d *aggregates.Direction) {
				assert.Equal(t, horizon, d.Horizon().Value())
			},
		},
		{
			name: "source capabilities only",
			command: func(id string) *commands.UpdateDirection {
				return &commands.UpdateDirection{DirectionID: id, SourceCapabilityIDs: &newPhysIDs}
			},
			assertion: func(t *testing.T, d *aggregates.Direction) {
				assert.Len(t, d.SourceCapabilityIDs(), 3)
			},
		},
		{
			name: "placements only",
			command: func(id string) *commands.UpdateDirection {
				placements := []commands.PlacementInput{{TargetBusinessDomainID: newDomID, ResultingName: "Unified Payroll"}}
				return &commands.UpdateDirection{DirectionID: id, Placements: &placements}
			},
			assertion: func(t *testing.T, d *aggregates.Direction) {
				assert.Len(t, d.Placements(), 1)
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			d := draftFixture(t)
			repo := &fakeRepoWithLoad{direction: d}
			_, err := NewUpdateDirectionHandler(repo).Handle(context.Background(), c.command(d.ID()))
			require.NoError(t, err)
			require.Len(t, repo.saved, 1)
			c.assertion(t, repo.saved[0])
		})
	}
}

func TestUpdateDirectionHandler_MultipleFieldsAtomic(t *testing.T) {
	d := draftFixture(t)
	repo := &fakeRepoWithLoad{direction: d}
	narrative := "Updated narrative."
	horizon := "later"
	placements := []commands.PlacementInput{
		{TargetBusinessDomainID: uuid.New().String(), ResultingName: "Combined Payroll"},
	}
	cmd := &commands.UpdateDirection{
		DirectionID: d.ID(),
		Narrative:   &narrative,
		Horizon:     &horizon,
		Placements:  &placements,
	}

	_, err := NewUpdateDirectionHandler(repo).Handle(context.Background(), cmd)

	require.NoError(t, err)
	require.Len(t, repo.saved, 1, "atomic update saves the aggregate once")
	saved := repo.saved[0]
	assert.Equal(t, narrative, saved.Narrative().Value())
	assert.Equal(t, horizon, saved.Horizon().Value())
	assert.Len(t, saved.Placements(), 1)
}

func TestUpdateDirectionHandler_InvalidValueDoesNotPersist(t *testing.T) {
	d := draftFixture(t)
	repo := &fakeRepoWithLoad{direction: d}
	narrative := "Updated narrative."
	badHorizon := "yesterday"
	cmd := &commands.UpdateDirection{
		DirectionID: d.ID(),
		Narrative:   &narrative,
		Horizon:     &badHorizon,
	}

	_, err := NewUpdateDirectionHandler(repo).Handle(context.Background(), cmd)

	assert.Error(t, err)
	assert.Empty(t, repo.saved, "failure on any field means nothing persists")
}
