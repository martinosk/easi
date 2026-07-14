package handlers

import (
	"context"
	"testing"

	"easi/backend/internal/architecturedirection/application/commands"
	"easi/backend/internal/architecturedirection/domain/aggregates"
	"easi/backend/internal/architecturedirection/domain/valueobjects"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddJourneyMilestoneHandler_Active_AddsMilestoneWithGeneratedID(t *testing.T) {
	j := plannedJourneyFixture(t)
	repo := &mockCapabilityJourneyRepository{loaded: j}
	year, quarter := 2026, 4

	_, err := NewAddJourneyMilestoneHandler(repo).Handle(context.Background(), &commands.AddJourneyMilestone{
		JourneyID:     j.ID(),
		Label:         "Cut over region A",
		TargetYear:    &year,
		TargetQuarter: &quarter,
		Status:        valueobjects.MilestoneStatusPlanned,
		Actor:         "a@example.com",
	})

	require.NoError(t, err)
	require.Len(t, repo.saved, 1)
	milestones := repo.saved[0].Milestones()
	require.Len(t, milestones, 1)
	assert.NotEmpty(t, milestones[0].ID())
	assert.Equal(t, "Cut over region A", milestones[0].Label())
}

func TestAddJourneyMilestoneHandler_LabelMissing_Fails(t *testing.T) {
	j := plannedJourneyFixture(t)
	repo := &mockCapabilityJourneyRepository{loaded: j}

	_, err := NewAddJourneyMilestoneHandler(repo).Handle(context.Background(), &commands.AddJourneyMilestone{
		JourneyID: j.ID(),
		Status:    valueobjects.MilestoneStatusPlanned,
		Actor:     "a@example.com",
	})

	assert.Error(t, err)
	assert.Empty(t, repo.saved)
}

func TestAddJourneyMilestoneHandler_Terminal_Fails(t *testing.T) {
	j := plannedJourneyFixture(t)
	require.NoError(t, j.Abandon("a@example.com"))
	j.MarkChangesAsCommitted()
	repo := &mockCapabilityJourneyRepository{loaded: j}

	_, err := NewAddJourneyMilestoneHandler(repo).Handle(context.Background(), &commands.AddJourneyMilestone{
		JourneyID: j.ID(),
		Label:     "too late",
		Status:    valueobjects.MilestoneStatusPlanned,
		Actor:     "a@example.com",
	})

	assert.ErrorIs(t, err, aggregates.ErrJourneyFrozen)
	assert.Empty(t, repo.saved)
}

func journeyWithMilestoneFixture(t *testing.T) (*aggregates.CapabilityJourney, string) {
	t.Helper()
	j := plannedJourneyFixture(t)
	require.NoError(t, j.AddMilestone(aggregates.MilestoneFacts{
		MilestoneID: uuid.New().String(),
		Label:       "First milestone",
		Status:      mustMilestoneStatus(t, valueobjects.MilestoneStatusPlanned),
		Actor:       "a@example.com",
	}))
	j.MarkChangesAsCommitted()
	return j, j.Milestones()[0].ID()
}

func mustMilestoneStatus(t *testing.T, status string) valueobjects.MilestoneStatus {
	t.Helper()
	s, err := valueobjects.NewMilestoneStatus(status)
	require.NoError(t, err)
	return s
}

func TestUpdateJourneyMilestoneHandler_ExistingMilestone_Updates(t *testing.T) {
	j, milestoneID := journeyWithMilestoneFixture(t)
	repo := &mockCapabilityJourneyRepository{loaded: j}

	_, err := NewUpdateJourneyMilestoneHandler(repo).Handle(context.Background(), &commands.UpdateJourneyMilestone{
		JourneyID:   j.ID(),
		MilestoneID: milestoneID,
		Label:       "First milestone done",
		Status:      valueobjects.MilestoneStatusDone,
		Actor:       "a@example.com",
	})

	require.NoError(t, err)
	require.Len(t, repo.saved, 1)
	milestones := repo.saved[0].Milestones()
	require.Len(t, milestones, 1)
	assert.Equal(t, "First milestone done", milestones[0].Label())
	assert.Equal(t, valueobjects.MilestoneStatusDone, milestones[0].Status().Value())
}

func TestJourneyMilestoneHandlers_UnknownMilestone_Fails(t *testing.T) {
	cases := []struct {
		name   string
		handle func(repo *mockCapabilityJourneyRepository, journeyID, milestoneID string) error
	}{
		{
			name: "update unknown milestone",
			handle: func(repo *mockCapabilityJourneyRepository, journeyID, milestoneID string) error {
				_, err := NewUpdateJourneyMilestoneHandler(repo).Handle(context.Background(), &commands.UpdateJourneyMilestone{
					JourneyID:   journeyID,
					MilestoneID: milestoneID,
					Label:       "does not exist",
					Status:      valueobjects.MilestoneStatusPlanned,
					Actor:       "a@example.com",
				})
				return err
			},
		},
		{
			name: "remove unknown milestone",
			handle: func(repo *mockCapabilityJourneyRepository, journeyID, milestoneID string) error {
				_, err := NewRemoveJourneyMilestoneHandler(repo).Handle(context.Background(), &commands.RemoveJourneyMilestone{
					JourneyID:   journeyID,
					MilestoneID: milestoneID,
					Actor:       "a@example.com",
				})
				return err
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			j := plannedJourneyFixture(t)
			repo := &mockCapabilityJourneyRepository{loaded: j}

			err := tc.handle(repo, j.ID(), uuid.New().String())

			assert.ErrorIs(t, err, aggregates.ErrJourneyMilestoneNotFound)
			assert.Empty(t, repo.saved)
		})
	}
}

func TestRemoveJourneyMilestoneHandler_ExistingMilestone_Removes(t *testing.T) {
	j, milestoneID := journeyWithMilestoneFixture(t)
	repo := &mockCapabilityJourneyRepository{loaded: j}

	_, err := NewRemoveJourneyMilestoneHandler(repo).Handle(context.Background(), &commands.RemoveJourneyMilestone{
		JourneyID:   j.ID(),
		MilestoneID: milestoneID,
		Actor:       "a@example.com",
	})

	require.NoError(t, err)
	require.Len(t, repo.saved, 1)
	assert.Empty(t, repo.saved[0].Milestones())
}
