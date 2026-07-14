package handlers

import (
	"context"
	"testing"

	"easi/backend/internal/architecturedirection/application/commands"
	"easi/backend/internal/architecturedirection/domain/aggregates"
	"easi/backend/internal/architecturedirection/domain/services"
	"easi/backend/internal/architecturedirection/domain/valueobjects"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateJourneyProgressHandler_VariousJourneyStates(t *testing.T) {
	cases := []struct {
		name     string
		setup    func(t *testing.T, j *aggregates.CapabilityJourney)
		progress int
		wantErr  error
	}{
		{
			name:     "active journey updates progress",
			setup:    func(*testing.T, *aggregates.CapabilityJourney) {},
			progress: 60,
		},
		{
			name:     "out of range progress rejected",
			setup:    func(*testing.T, *aggregates.CapabilityJourney) {},
			progress: 150,
			wantErr:  valueobjects.ErrInvalidJourneyProgress,
		},
		{
			name: "terminal journey rejects update",
			setup: func(t *testing.T, j *aggregates.CapabilityJourney) {
				require.NoError(t, j.Abandon("a@example.com"))
				j.MarkChangesAsCommitted()
			},
			progress: 60,
			wantErr:  aggregates.ErrJourneyFrozen,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			j := plannedJourneyFixture(t)
			tc.setup(t, j)
			repo := &mockCapabilityJourneyRepository{loaded: j}

			_, err := NewUpdateJourneyProgressHandler(repo).Handle(context.Background(),
				&commands.UpdateJourneyProgress{JourneyID: j.ID(), Progress: tc.progress, Actor: "a@example.com"})

			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
				assert.Empty(t, repo.saved)
				return
			}
			require.NoError(t, err)
			require.Len(t, repo.saved, 1)
			require.NotNil(t, repo.saved[0].Progress())
			assert.Equal(t, tc.progress, repo.saved[0].Progress().Value())
		})
	}
}

func TestUpdateJourneyDetailsHandler_Active_UpdatesNoteAndPeriod(t *testing.T) {
	j := plannedJourneyFixture(t)
	repo := &mockCapabilityJourneyRepository{loaded: j}
	year, quarter := 2027, 2

	_, err := NewUpdateJourneyDetailsHandler(repo).Handle(context.Background(), &commands.UpdateJourneyDetails{
		JourneyID:     j.ID(),
		Note:          "updated plan",
		TargetYear:    &year,
		TargetQuarter: &quarter,
		Actor:         "a@example.com",
	})

	require.NoError(t, err)
	require.Len(t, repo.saved, 1)
	assert.Equal(t, "updated plan", repo.saved[0].Note().Value())
	require.NotNil(t, repo.saved[0].TargetPeriod())
	assert.Equal(t, 2027, repo.saved[0].TargetPeriod().Year())
}

func moveJourneyFixture(t *testing.T) *aggregates.CapabilityJourney {
	t.Helper()
	j, err := aggregates.PlanCapabilityJourney(aggregates.CapabilityJourneyFacts{
		ID:            valueobjects.NewCapabilityJourneyID(),
		CapabilityID:  mustPhysicalCapabilityRef(t, uuid.New().String()),
		Kind:          mustJourneyKind(t, valueobjects.JourneyKindMove),
		ToApp:         mustNewApplicationRef(t, uuid.New().String()),
		Note:          mustNewNarrative(t, "moving domain"),
		TargetDomain:  ptrBusinessDomainRef(t, uuid.New().String()),
		ResultingName: "Freight invoicing",
		PlannedBy:     "architect@example.com",
	})
	require.NoError(t, err)
	j.MarkChangesAsCommitted()
	return j
}

func ptrBusinessDomainRef(t *testing.T, id string) *valueobjects.BusinessDomainRef {
	t.Helper()
	ref, err := valueobjects.NewBusinessDomainRef(id)
	require.NoError(t, err)
	return &ref
}

func TestUpdateJourneyDetailsHandler_MoveFieldValidation(t *testing.T) {
	cases := []struct {
		name        string
		journey     func(t *testing.T) *aggregates.CapabilityJourney
		cmd         func(journeyID string) *commands.UpdateJourneyDetails
		wantErr     error
		assertSaved func(t *testing.T, saved *aggregates.CapabilityJourney)
	}{
		{
			name:    "target period missing quarter",
			journey: plannedJourneyFixture,
			cmd: func(journeyID string) *commands.UpdateJourneyDetails {
				year := 2027
				return &commands.UpdateJourneyDetails{JourneyID: journeyID, Note: "updated plan", TargetYear: &year, Actor: "a@example.com"}
			},
			wantErr: ErrTargetPeriodRequiresBoth,
		},
		{
			name:    "resulting name on non-move journey rejected",
			journey: plannedJourneyFixture,
			cmd: func(journeyID string) *commands.UpdateJourneyDetails {
				return &commands.UpdateJourneyDetails{JourneyID: journeyID, Note: "updated plan", ResultingName: "New name", Actor: "a@example.com"}
			},
			wantErr: aggregates.ErrJourneyMoveFieldsOnNonMove,
		},
		{
			name:    "resulting name on move journey updates",
			journey: moveJourneyFixture,
			cmd: func(journeyID string) *commands.UpdateJourneyDetails {
				return &commands.UpdateJourneyDetails{JourneyID: journeyID, Note: "updated plan", ResultingName: "Freight invoicing v2", Actor: "a@example.com"}
			},
			assertSaved: func(t *testing.T, saved *aggregates.CapabilityJourney) {
				assert.Equal(t, "Freight invoicing v2", saved.ResultingName())
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			j := tc.journey(t)
			repo := &mockCapabilityJourneyRepository{loaded: j}

			_, err := NewUpdateJourneyDetailsHandler(repo).Handle(context.Background(), tc.cmd(j.ID()))

			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
				assert.Empty(t, repo.saved)
				return
			}
			require.NoError(t, err)
			require.Len(t, repo.saved, 1)
			tc.assertSaved(t, repo.saved[0])
		})
	}
}

func TestChangeJourneySourceApplicationsHandler_ValidChange_Succeeds(t *testing.T) {
	j := plannedJourneyFixture(t)
	repo := &mockCapabilityJourneyRepository{loaded: j}
	newFrom := uuid.New().String()

	_, err := NewChangeJourneySourceApplicationsHandler(repo, services.ComponentExists(alwaysExists)).Handle(context.Background(),
		&commands.ChangeJourneySourceApplications{JourneyID: j.ID(), FromComponentIDs: []string{newFrom}, Actor: "a@example.com"})

	require.NoError(t, err)
	require.Len(t, repo.saved, 1)
	require.Len(t, repo.saved[0].FromApps(), 1)
	assert.Equal(t, newFrom, repo.saved[0].FromApps()[0].Value())
}

func TestChangeJourneySourceApplicationsHandler_CardinalityViolation_Fails(t *testing.T) {
	j := plannedJourneyFixture(t)
	repo := &mockCapabilityJourneyRepository{loaded: j}

	_, err := NewChangeJourneySourceApplicationsHandler(repo, services.ComponentExists(alwaysExists)).Handle(context.Background(),
		&commands.ChangeJourneySourceApplications{JourneyID: j.ID(), FromComponentIDs: []string{}, Actor: "a@example.com"})

	assert.ErrorIs(t, err, valueobjects.ErrInvalidSourceApplicationCount)
	assert.Empty(t, repo.saved)
}

func TestChangeJourneySourceApplicationsHandler_TargetAmongSources_Fails(t *testing.T) {
	j := plannedJourneyFixture(t)
	repo := &mockCapabilityJourneyRepository{loaded: j}

	_, err := NewChangeJourneySourceApplicationsHandler(repo, services.ComponentExists(alwaysExists)).Handle(context.Background(),
		&commands.ChangeJourneySourceApplications{JourneyID: j.ID(), FromComponentIDs: []string{j.ToApp().Value()}, Actor: "a@example.com"})

	assert.ErrorIs(t, err, aggregates.ErrJourneyTargetAmongSources)
	assert.Empty(t, repo.saved)
}

func TestChangeJourneySourceApplicationsHandler_ComponentDoesNotExist_Fails(t *testing.T) {
	j := plannedJourneyFixture(t)
	repo := &mockCapabilityJourneyRepository{loaded: j}

	_, err := NewChangeJourneySourceApplicationsHandler(repo, services.ComponentExists(neverExists)).Handle(context.Background(),
		&commands.ChangeJourneySourceApplications{JourneyID: j.ID(), FromComponentIDs: []string{uuid.New().String()}, Actor: "a@example.com"})

	assert.ErrorIs(t, err, services.ErrReferencedEntityNotFound)
	assert.Empty(t, repo.saved)
}
