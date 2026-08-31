package aggregates

import (
	"testing"

	"easi/backend/internal/architecturedirection/domain/events"
	"easi/backend/internal/architecturedirection/domain/valueobjects"
	domain "easi/backend/internal/shared/eventsourcing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTargetMaturityVO(t *testing.T, value int) *valueobjects.TargetMaturity {
	t.Helper()
	m, err := valueobjects.NewTargetMaturity(value)
	require.NoError(t, err)
	return &m
}

func maturityOpts(t *testing.T, target, current int) journeyOpts {
	t.Helper()
	return journeyOpts{
		kind:            valueobjects.JourneyKindMaturity,
		targetMaturity:  newTargetMaturityVO(t, target),
		currentMaturity: current,
	}
}

func TestPlanCapabilityJourney_Maturity_Succeeds_Spec211Rules1And2(t *testing.T) {
	j, err := planWith(t, maturityOpts(t, 65, 30))

	require.NoError(t, err)
	assert.Equal(t, valueobjects.JourneyKindMaturity, j.Kind().Value())
	assert.Equal(t, valueobjects.JourneyStatusPlanned, j.Status().Value())
	assert.Empty(t, j.FromApps())
	assert.Empty(t, j.ToApp().Value())
	require.NotNil(t, j.TargetMaturity())
	assert.Equal(t, 65, j.TargetMaturity().Value())

	changes := j.GetUncommittedChanges()
	require.Len(t, changes, 1)
	evt, ok := changes[0].(events.JourneyPlanned)
	require.True(t, ok)
	require.NotNil(t, evt.TargetMaturity)
	assert.Equal(t, 65, *evt.TargetMaturity)
}

func TestPlanCapabilityJourney_Maturity_RejectsApplicationFields_Spec211Rule2(t *testing.T) {
	toApp := newComponentRef(t)
	opts := maturityOpts(t, 65, 30)
	opts.toApp = &toApp

	_, err := planWith(t, opts)

	assert.ErrorIs(t, err, ErrJourneyMaturityRefusesApplications)
}

func TestPlanCapabilityJourney_Maturity_RejectsFromApplications_Spec211Rule2(t *testing.T) {
	opts := maturityOpts(t, 65, 30)
	opts.fromApps = newComponentRefs(t, 1)

	_, err := planWith(t, opts)

	assert.ErrorIs(t, err, valueobjects.ErrInvalidSourceApplicationCount)
}

func TestPlanCapabilityJourney_Maturity_RequiresTargetMaturity_Spec211Rule2(t *testing.T) {
	opts := maturityOpts(t, 65, 30)
	opts.targetMaturity = nil

	_, err := planWith(t, opts)

	assert.ErrorIs(t, err, ErrJourneyMaturityRequiresTarget)
}

func TestPlanCapabilityJourney_Maturity_TargetMustExceedCurrent_Spec211Rule3(t *testing.T) {
	cases := []struct {
		name    string
		target  int
		current int
		wantErr bool
	}{
		{"below current is rejected", 25, 30, true},
		{"equal to current is rejected", 30, 30, true},
		{"above current succeeds", 31, 30, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := planWith(t, maturityOpts(t, c.target, c.current))
			if c.wantErr {
				assert.ErrorIs(t, err, ErrJourneyMaturityTargetNotAboveCurrent)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPlanCapabilityJourney_NonMaturity_RejectsTargetMaturity_Spec211Rule4(t *testing.T) {
	for _, kind := range []string{valueobjects.JourneyKindMigration, valueobjects.JourneyKindConsolidation, valueobjects.JourneyKindCarveOut} {
		t.Run(kind, func(t *testing.T) {
			_, err := planWith(t, journeyOpts{kind: kind, targetMaturity: newTargetMaturityVO(t, 65)})
			assert.ErrorIs(t, err, ErrJourneyTargetMaturityOnNonMaturity)
		})
	}
}

func TestCapabilityJourney_Maturity_RunsTheNormalLifecycle_Spec211Rule5(t *testing.T) {
	j, err := planWith(t, maturityOpts(t, 65, 30))
	require.NoError(t, err)

	require.NoError(t, j.Start(journeyActor))
	require.NoError(t, j.UpdateProgress(newProgress(t, 40), journeyActor))
	require.NoError(t, j.AddMilestone(plannedMilestoneFacts(t, "m1", "Name a single owner")))
	require.NoError(t, j.Complete(journeyActor))

	assert.Equal(t, valueobjects.JourneyStatusDone, j.Status().Value())
	assert.Len(t, j.Milestones(), 1)
}

func TestLoadCapabilityJourneyFromHistory_ReconstructsMaturityJourney_Spec211Rule5(t *testing.T) {
	original, err := planWith(t, maturityOpts(t, 65, 30))
	require.NoError(t, err)

	replayed, err := LoadCapabilityJourneyFromHistory(original.GetUncommittedChanges())

	require.NoError(t, err)
	assert.Equal(t, valueobjects.JourneyKindMaturity, replayed.Kind().Value())
	assert.Empty(t, replayed.ToApp().Value())
	require.NotNil(t, replayed.TargetMaturity())
	assert.Equal(t, 65, replayed.TargetMaturity().Value())
}

func TestLoadCapabilityJourneyFromHistory_CorruptTargetMaturity_Fails(t *testing.T) {
	outOfRange := 250
	evt := events.NewJourneyPlanned(events.JourneyPlannedFields{
		ID:             valueobjects.NewCapabilityJourneyID().Value(),
		CapabilityID:   newCapabilityRef(t).Value(),
		Kind:           valueobjects.JourneyKindMaturity,
		TargetMaturity: &outOfRange,
		PlannedBy:      journeyActor,
	})

	_, err := LoadCapabilityJourneyFromHistory([]domain.DomainEvent{evt})

	assert.ErrorIs(t, err, ErrCorruptedCapabilityJourneyEvent)
}
