package handlers

import (
	"context"
	"errors"
	"testing"

	"easi/backend/internal/architecturedirection/application/commands"
	"easi/backend/internal/architecturedirection/domain/aggregates"
	"easi/backend/internal/architecturedirection/domain/valueobjects"
	"easi/backend/internal/shared/cqrs"
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
	ec, _ := valueobjects.NewEnterpriseCapabilityRef(uuid.New().String())
	dt, _ := valueobjects.NewDirectionType("consolidate")
	r1, _ := valueobjects.NewPhysicalCapabilityRef(uuid.New().String())
	r2, _ := valueobjects.NewPhysicalCapabilityRef(uuid.New().String())
	p, _ := valueobjects.NewPlacement(uuid.New().String(), "")
	h, _ := valueobjects.NewHorizon("next")
	n, _ := sharedvo.NewDescription("Some narrative.")
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
	name       string
	target     string
	expectErr  error
	expectFn   func(*aggregates.Direction) bool
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

type fieldUpdateCase struct {
	name      string
	handler   func(DirectionLoaderRepository) cqrs.CommandHandler
	command   func(directionID string) cqrs.Command
	assertion func(*testing.T, *aggregates.Direction)
}

func TestFieldUpdateHandlers(t *testing.T) {
	newPhysIDs := []string{uuid.New().String(), uuid.New().String(), uuid.New().String()}
	newDomID := uuid.New().String()

	cases := []fieldUpdateCase{
		{
			name:    "narrative",
			handler: NewUpdateDirectionNarrativeHandler,
			command: func(id string) cqrs.Command {
				return &commands.UpdateDirectionNarrative{DirectionID: id, Narrative: "Refined narrative."}
			},
			assertion: func(t *testing.T, d *aggregates.Direction) {
				assert.Equal(t, "Refined narrative.", d.Narrative().Value())
			},
		},
		{
			name:    "horizon",
			handler: NewUpdateDirectionHorizonHandler,
			command: func(id string) cqrs.Command {
				return &commands.UpdateDirectionHorizon{DirectionID: id, Horizon: "later"}
			},
			assertion: func(t *testing.T, d *aggregates.Direction) {
				assert.Equal(t, "later", d.Horizon().Value())
			},
		},
		{
			name:    "source capabilities",
			handler: NewUpdateDirectionSourceCapabilitiesHandler,
			command: func(id string) cqrs.Command {
				return &commands.UpdateDirectionSourceCapabilities{DirectionID: id, SourceCapabilityIDs: newPhysIDs}
			},
			assertion: func(t *testing.T, d *aggregates.Direction) {
				assert.Len(t, d.SourceCapabilityIDs(), 3)
			},
		},
		{
			name:    "placements",
			handler: NewUpdateDirectionPlacementsHandler,
			command: func(id string) cqrs.Command {
				return &commands.UpdateDirectionPlacements{
					DirectionID: id,
					Placements:  []commands.PlacementInput{{TargetBusinessDomainID: newDomID, ResultingName: "Unified Payroll"}},
				}
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
			_, err := c.handler(repo).Handle(context.Background(), c.command(d.ID()))
			require.NoError(t, err)
			require.Len(t, repo.saved, 1)
			c.assertion(t, repo.saved[0])
		})
	}
}
