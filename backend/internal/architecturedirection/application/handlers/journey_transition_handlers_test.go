package handlers

import (
	"context"
	"errors"
	"testing"

	"easi/backend/internal/architecturedirection/application/commands"
	"easi/backend/internal/architecturedirection/domain/aggregates"
	"easi/backend/internal/architecturedirection/domain/valueobjects"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func plannedJourneyFixture(t *testing.T) *aggregates.CapabilityJourney {
	t.Helper()
	j, err := aggregates.PlanCapabilityJourney(aggregates.CapabilityJourneyFacts{
		ID:           valueobjects.NewCapabilityJourneyID(),
		CapabilityID: mustPhysicalCapabilityRef(t, uuid.New().String()),
		Kind:         mustJourneyKind(t, valueobjects.JourneyKindMigration),
		FromApps:     []valueobjects.ApplicationRef{mustNewApplicationRef(t, uuid.New().String())},
		ToApp:        mustNewApplicationRef(t, uuid.New().String()),
		Note:         mustNewNarrative(t, "moving on"),
		PlannedBy:    "architect@example.com",
	})
	require.NoError(t, err)
	j.MarkChangesAsCommitted()
	return j
}

func mustPhysicalCapabilityRef(t *testing.T, id string) valueobjects.PhysicalCapabilityRef {
	t.Helper()
	ref, err := valueobjects.NewPhysicalCapabilityRef(id)
	require.NoError(t, err)
	return ref
}

func mustJourneyKind(t *testing.T, kind string) valueobjects.JourneyKind {
	t.Helper()
	k, err := valueobjects.NewJourneyKind(kind)
	require.NoError(t, err)
	return k
}

func TestJourneyTransitionHandlers_ValidTransition_Succeeds(t *testing.T) {
	cases := []struct {
		name       string
		setup      func(t *testing.T, j *aggregates.CapabilityJourney)
		handle     func(repo *mockCapabilityJourneyRepository, journeyID string) error
		wantStatus string
	}{
		{
			name:  "start from planned",
			setup: func(*testing.T, *aggregates.CapabilityJourney) {},
			handle: func(repo *mockCapabilityJourneyRepository, journeyID string) error {
				_, err := NewStartJourneyHandler(repo).Handle(context.Background(), &commands.StartJourney{JourneyID: journeyID, Actor: "a@example.com"})
				return err
			},
			wantStatus: valueobjects.JourneyStatusInFlight,
		},
		{
			name: "complete from in-flight",
			setup: func(t *testing.T, j *aggregates.CapabilityJourney) {
				require.NoError(t, j.Start("a@example.com"))
				j.MarkChangesAsCommitted()
			},
			handle: func(repo *mockCapabilityJourneyRepository, journeyID string) error {
				_, err := NewCompleteJourneyHandler(repo).Handle(context.Background(), &commands.CompleteJourney{JourneyID: journeyID, Actor: "a@example.com"})
				return err
			},
			wantStatus: valueobjects.JourneyStatusDone,
		},
		{
			name:  "abandon from planned",
			setup: func(*testing.T, *aggregates.CapabilityJourney) {},
			handle: func(repo *mockCapabilityJourneyRepository, journeyID string) error {
				_, err := NewAbandonJourneyHandler(repo).Handle(context.Background(), &commands.AbandonJourney{JourneyID: journeyID, Actor: "a@example.com"})
				return err
			},
			wantStatus: valueobjects.JourneyStatusAbandoned,
		},
		{
			name: "abandon from in-flight",
			setup: func(t *testing.T, j *aggregates.CapabilityJourney) {
				require.NoError(t, j.Start("a@example.com"))
				j.MarkChangesAsCommitted()
			},
			handle: func(repo *mockCapabilityJourneyRepository, journeyID string) error {
				_, err := NewAbandonJourneyHandler(repo).Handle(context.Background(), &commands.AbandonJourney{JourneyID: journeyID, Actor: "a@example.com"})
				return err
			},
			wantStatus: valueobjects.JourneyStatusAbandoned,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			j := plannedJourneyFixture(t)
			tc.setup(t, j)
			repo := &mockCapabilityJourneyRepository{loaded: j}

			err := tc.handle(repo, j.ID())

			require.NoError(t, err)
			require.Len(t, repo.saved, 1)
			assert.Equal(t, tc.wantStatus, repo.saved[0].Status().Value())
		})
	}
}

func TestStartJourneyHandler_NotFound_Fails(t *testing.T) {
	repo := &mockCapabilityJourneyRepository{getErr: errors.New("not found")}

	_, err := NewStartJourneyHandler(repo).Handle(context.Background(), &commands.StartJourney{JourneyID: uuid.New().String()})

	assert.Error(t, err)
	assert.Empty(t, repo.saved)
}

func TestCompleteJourneyHandler_Planned_RejectsInvalidTransition(t *testing.T) {
	j := plannedJourneyFixture(t)
	repo := &mockCapabilityJourneyRepository{loaded: j}

	_, err := NewCompleteJourneyHandler(repo).Handle(context.Background(), &commands.CompleteJourney{JourneyID: j.ID(), Actor: "a@example.com"})

	assert.ErrorIs(t, err, aggregates.ErrInvalidJourneyTransition)
	assert.Empty(t, repo.saved)
}

func TestAbandonJourneyHandler_AlreadyTerminal_Fails(t *testing.T) {
	j := plannedJourneyFixture(t)
	require.NoError(t, j.Abandon("a@example.com"))
	j.MarkChangesAsCommitted()
	repo := &mockCapabilityJourneyRepository{loaded: j}

	_, err := NewAbandonJourneyHandler(repo).Handle(context.Background(), &commands.AbandonJourney{JourneyID: j.ID(), Actor: "a@example.com"})

	assert.ErrorIs(t, err, aggregates.ErrInvalidJourneyTransition)
	assert.Empty(t, repo.saved)
}

func TestJourneyTransitionHandlers_InvalidCommandType_Fails(t *testing.T) {
	repo := &mockCapabilityJourneyRepository{}
	_, err := NewStartJourneyHandler(repo).Handle(context.Background(), &commands.CompleteJourney{})
	assert.Error(t, err)
}
