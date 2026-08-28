package aggregates

import (
	"testing"
	"time"

	"easi/backend/internal/architecturedirection/domain/entities"
	"easi/backend/internal/architecturedirection/domain/events"
	"easi/backend/internal/architecturedirection/domain/valueobjects"
	pl "easi/backend/internal/architecturedirection/publishedlanguage"
	domain "easi/backend/internal/shared/eventsourcing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const journeyActor = "architect@example.com"

func newJourneyKind(t *testing.T, v string) valueobjects.JourneyKind {
	t.Helper()
	k, err := valueobjects.NewJourneyKind(v)
	require.NoError(t, err)
	return k
}

func newMilestoneStatusVO(t *testing.T, v string) valueobjects.MilestoneStatus {
	t.Helper()
	s, err := valueobjects.NewMilestoneStatus(v)
	require.NoError(t, err)
	return s
}

func newTargetPeriodVO(t *testing.T, year, quarter int) valueobjects.TargetPeriod {
	t.Helper()
	p, err := valueobjects.NewTargetPeriod(year, quarter)
	require.NoError(t, err)
	return p
}

func newBusinessDomainRefVO(t *testing.T) valueobjects.BusinessDomainRef {
	t.Helper()
	ref, err := valueobjects.NewBusinessDomainRef(uuid.New().String())
	require.NoError(t, err)
	return ref
}

func newProgress(t *testing.T, v int) valueobjects.JourneyProgress {
	t.Helper()
	p, err := valueobjects.NewJourneyProgress(v)
	require.NoError(t, err)
	return p
}

func newComponentRefs(t *testing.T, n int) []valueobjects.ApplicationRef {
	t.Helper()
	refs := make([]valueobjects.ApplicationRef, n)
	for i := range refs {
		refs[i] = newComponentRef(t)
	}
	return refs
}

type journeyOpts struct {
	kind          string
	fromApps      []valueobjects.ApplicationRef
	fromAppCount  int
	toApp         *valueobjects.ApplicationRef
	note          string
	targetPeriod  *valueobjects.TargetPeriod
	targetDomain  *valueobjects.BusinessDomainRef
	targetParent  *valueobjects.PhysicalCapabilityRef
	resultingName string
}

func planWith(t *testing.T, opts journeyOpts) (*CapabilityJourney, error) {
	t.Helper()
	if opts.kind == "" {
		opts.kind = valueobjects.JourneyKindMigration
	}
	fromApps := opts.fromApps
	if fromApps == nil {
		count := opts.fromAppCount
		if count == 0 && opts.kind != valueobjects.JourneyKindMove {
			count = 1
		}
		fromApps = newComponentRefs(t, count)
	}
	toApp := newComponentRef(t)
	if opts.toApp != nil {
		toApp = *opts.toApp
	}
	return PlanCapabilityJourney(CapabilityJourneyFacts{
		ID:            valueobjects.NewCapabilityJourneyID(),
		CapabilityID:  newCapabilityRef(t),
		Kind:          newJourneyKind(t, opts.kind),
		FromApps:      fromApps,
		ToApp:         toApp,
		Note:          newRationale(t, opts.note),
		TargetPeriod:  opts.targetPeriod,
		TargetDomain:  opts.targetDomain,
		TargetParent:  opts.targetParent,
		ResultingName: opts.resultingName,
		PlannedBy:     journeyActor,
	})
}

func plannedJourney(t *testing.T) *CapabilityJourney {
	t.Helper()
	j, err := planWith(t, journeyOpts{note: "route by route"})
	require.NoError(t, err)
	j.MarkChangesAsCommitted()
	return j
}

func inFlightJourney(t *testing.T) *CapabilityJourney {
	t.Helper()
	j := plannedJourney(t)
	require.NoError(t, j.Start(journeyActor))
	j.MarkChangesAsCommitted()
	return j
}

func doneJourney(t *testing.T) *CapabilityJourney {
	t.Helper()
	j := inFlightJourney(t)
	require.NoError(t, j.Complete(journeyActor))
	j.MarkChangesAsCommitted()
	return j
}

func abandonedJourney(t *testing.T) *CapabilityJourney {
	t.Helper()
	j := plannedJourney(t)
	require.NoError(t, j.Abandon(journeyActor))
	j.MarkChangesAsCommitted()
	return j
}

func plannedMoveJourney(t *testing.T) (*CapabilityJourney, valueobjects.BusinessDomainRef, valueobjects.PhysicalCapabilityRef) {
	t.Helper()
	domainRef := newBusinessDomainRefVO(t)
	parentRef := newCapabilityRef(t)
	j, err := planWith(t, journeyOpts{
		kind:          valueobjects.JourneyKindMove,
		targetDomain:  &domainRef,
		targetParent:  &parentRef,
		resultingName: "Freight invoicing",
	})
	require.NoError(t, err)
	return j, domainRef, parentRef
}

func assertMoveDestination(t *testing.T, j *CapabilityJourney, domainRef valueobjects.BusinessDomainRef, parentRef valueobjects.PhysicalCapabilityRef) {
	t.Helper()
	require.NotNil(t, j.TargetDomain())
	assert.Equal(t, domainRef.Value(), j.TargetDomain().Value())
	require.NotNil(t, j.TargetParent())
	assert.Equal(t, parentRef.Value(), j.TargetParent().Value())
	assert.Equal(t, "Freight invoicing", j.ResultingName())
}

func plannedMilestoneFacts(t *testing.T, id, label string) MilestoneFacts {
	t.Helper()
	return MilestoneFacts{
		MilestoneID: id,
		Label:       label,
		Status:      newMilestoneStatusVO(t, valueobjects.MilestoneStatusPlanned),
		Actor:       journeyActor,
	}
}

func addPlannedMilestone(t *testing.T, j *CapabilityJourney, id, label string) {
	t.Helper()
	require.NoError(t, j.AddMilestone(plannedMilestoneFacts(t, id, label)))
}

func detailsFacts(t *testing.T, note string, targetPeriod *valueobjects.TargetPeriod, resultingName string) JourneyDetailsFacts {
	t.Helper()
	return JourneyDetailsFacts{
		Note:          newRationale(t, note),
		TargetPeriod:  targetPeriod,
		ResultingName: resultingName,
		Actor:         journeyActor,
	}
}

func TestPlanCapabilityJourney_Migration_Succeeds_Rule1(t *testing.T) {
	j, err := planWith(t, journeyOpts{kind: valueobjects.JourneyKindMigration, fromAppCount: 1, note: "route by route"})

	require.NoError(t, err)
	assert.NotEmpty(t, j.ID())
	assert.Equal(t, valueobjects.JourneyStatusPlanned, j.Status().Value())
	assert.Equal(t, valueobjects.JourneyKindMigration, j.Kind().Value())
	assert.Len(t, j.FromApps(), 1)
	assert.Equal(t, "route by route", j.Note().Value())
	assert.Empty(t, j.Milestones())
	assert.Nil(t, j.Progress())

	changes := j.GetUncommittedChanges()
	require.Len(t, changes, 1)
	evt, ok := changes[0].(events.JourneyPlanned)
	require.True(t, ok)
	assert.Equal(t, journeyActor, evt.PlannedBy)
	assert.False(t, evt.OccurredOn.IsZero(), "Rule 13: server timestamp recorded on the event")
}

func TestPlanCapabilityJourney_KindCardinality_Rule3(t *testing.T) {
	cases := []struct {
		name    string
		kind    string
		count   int
		wantErr bool
	}{
		{"migration with zero sources is rejected", valueobjects.JourneyKindMigration, 0, true},
		{"migration with one source succeeds", valueobjects.JourneyKindMigration, 1, false},
		{"consolidation with zero sources is rejected", valueobjects.JourneyKindConsolidation, 0, true},
		{"consolidation with one source succeeds", valueobjects.JourneyKindConsolidation, 1, false},
		{"consolidation with two sources succeeds", valueobjects.JourneyKindConsolidation, 2, false},
		{"carve-out with zero sources is rejected", valueobjects.JourneyKindCarveOut, 0, true},
		{"carve-out with one source succeeds", valueobjects.JourneyKindCarveOut, 1, false},
		{"carve-out with two sources is rejected", valueobjects.JourneyKindCarveOut, 2, true},
		{"move with zero sources succeeds", valueobjects.JourneyKindMove, 0, false},
		{"move with one source succeeds", valueobjects.JourneyKindMove, 1, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			opts := journeyOpts{kind: c.kind, fromApps: newComponentRefs(t, c.count)}
			if c.kind == valueobjects.JourneyKindMove {
				domainRef := newBusinessDomainRefVO(t)
				opts.targetDomain = &domainRef
				opts.resultingName = "Freight invoicing"
			}
			_, err := planWith(t, opts)
			if c.wantErr {
				assert.ErrorIs(t, err, valueobjects.ErrInvalidSourceApplicationCount)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPlanCapabilityJourney_TargetMustNotBeAmongSources_Rule4(t *testing.T) {
	fromApps := newComponentRefs(t, 1)

	_, err := planWith(t, journeyOpts{fromApps: fromApps, toApp: &fromApps[0]})

	assert.ErrorIs(t, err, ErrJourneyTargetAmongSources)
}

func TestPlanCapabilityJourney_Move_RequiresTargetDomain_Rule5(t *testing.T) {
	_, err := planWith(t, journeyOpts{kind: valueobjects.JourneyKindMove, resultingName: "Freight invoicing"})

	assert.ErrorIs(t, err, ErrJourneyMoveRequiresTargetDomain)
}

func TestPlanCapabilityJourney_Move_RequiresResultingName_Rule5(t *testing.T) {
	domainRef := newBusinessDomainRefVO(t)

	_, err := planWith(t, journeyOpts{kind: valueobjects.JourneyKindMove, targetDomain: &domainRef})

	assert.ErrorIs(t, err, valueobjects.ErrResultingCapabilityNameRequired)
}

func TestPlanCapabilityJourney_Move_TargetParentOptional_Rule5(t *testing.T) {
	j, domainRef, parentRef := plannedMoveJourney(t)

	assertMoveDestination(t, j, domainRef, parentRef)
}

func TestPlanCapabilityJourney_Move_WithoutTargetParent_Succeeds_Rule5(t *testing.T) {
	domainRef := newBusinessDomainRefVO(t)

	j, err := planWith(t, journeyOpts{kind: valueobjects.JourneyKindMove, targetDomain: &domainRef, resultingName: "Freight invoicing"})

	require.NoError(t, err)
	assert.Nil(t, j.TargetParent())
}

func TestPlanCapabilityJourney_NonMove_RejectsMoveFields_Rule5(t *testing.T) {
	domainRef := newBusinessDomainRefVO(t)
	parentRef := newCapabilityRef(t)
	cases := []struct {
		name string
		opts journeyOpts
	}{
		{"target domain", journeyOpts{kind: valueobjects.JourneyKindMigration, targetDomain: &domainRef}},
		{"target parent", journeyOpts{kind: valueobjects.JourneyKindMigration, targetParent: &parentRef}},
		{"resulting name", journeyOpts{kind: valueobjects.JourneyKindMigration, resultingName: "Freight invoicing"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := planWith(t, c.opts)
			assert.ErrorIs(t, err, ErrJourneyMoveFieldsOnNonMove)
		})
	}
}

func startJourney(j *CapabilityJourney) error    { return j.Start(journeyActor) }
func completeJourney(j *CapabilityJourney) error { return j.Complete(journeyActor) }
func abandonJourney(j *CapabilityJourney) error  { return j.Abandon(journeyActor) }

func transitionActorAndTime(t *testing.T, evt domain.DomainEvent) (string, time.Time) {
	t.Helper()
	switch e := evt.(type) {
	case events.JourneyStarted:
		return e.StartedBy, e.OccurredOn
	case events.JourneyCompleted:
		return e.CompletedBy, e.OccurredOn
	case events.JourneyAbandoned:
		return e.AbandonedBy, e.OccurredOn
	}
	t.Fatalf("unexpected transition event %T", evt)
	return "", time.Time{}
}

func TestCapabilityJourney_Transitions_Succeed_Rule6(t *testing.T) {
	cases := []struct {
		name       string
		setup      func(*testing.T) *CapabilityJourney
		transition func(*CapabilityJourney) error
		wantStatus string
		wantEvent  string
	}{
		{"start from planned", plannedJourney, startJourney, valueobjects.JourneyStatusInFlight, pl.JourneyStarted},
		{"complete from in-flight", inFlightJourney, completeJourney, valueobjects.JourneyStatusDone, pl.JourneyCompleted},
		{"abandon from planned", plannedJourney, abandonJourney, valueobjects.JourneyStatusAbandoned, pl.JourneyAbandoned},
		{"abandon from in-flight", inFlightJourney, abandonJourney, valueobjects.JourneyStatusAbandoned, pl.JourneyAbandoned},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			j := c.setup(t)

			require.NoError(t, c.transition(j))

			assert.Equal(t, c.wantStatus, j.Status().Value())
			changes := j.GetUncommittedChanges()
			require.Len(t, changes, 1, "each transition is one discrete past-tense event")
			assert.Equal(t, c.wantEvent, changes[0].EventType())
			actor, occurredOn := transitionActorAndTime(t, changes[0])
			assert.Equal(t, journeyActor, actor, "Rule 13: event carries the acting user")
			assert.False(t, occurredOn.IsZero(), "Rule 13: event carries occurred-at")
		})
	}
}

func TestCapabilityJourney_Transitions_RejectedFromWrongStatus_Rule6(t *testing.T) {
	cases := []struct {
		name       string
		setup      func(*testing.T) *CapabilityJourney
		transition func(*CapabilityJourney) error
	}{
		{"start from in-flight", inFlightJourney, startJourney},
		{"start from done", doneJourney, startJourney},
		{"start from abandoned", abandonedJourney, startJourney},
		{"complete from planned", plannedJourney, completeJourney},
		{"complete from done", doneJourney, completeJourney},
		{"complete from abandoned", abandonedJourney, completeJourney},
		{"abandon from done", doneJourney, abandonJourney},
		{"abandon from abandoned", abandonedJourney, abandonJourney},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := c.transition(c.setup(t))
			assert.ErrorIs(t, err, ErrInvalidJourneyTransition)
		})
	}
}

func TestCapabilityJourney_Complete_DoesNotTouchProgressOrOtherState_Rule10(t *testing.T) {
	j := inFlightJourney(t)
	require.NoError(t, j.UpdateProgress(newProgress(t, 35), journeyActor))
	j.MarkChangesAsCommitted()
	fromAppsBefore := j.FromApps()
	kindBefore := j.Kind().Value()

	err := j.Complete(journeyActor)

	require.NoError(t, err)
	require.NotNil(t, j.Progress())
	assert.Equal(t, 35, j.Progress().Value(), "completion must not touch previously recorded progress")
	assert.Equal(t, kindBefore, j.Kind().Value())
	assert.Len(t, j.FromApps(), len(fromAppsBefore))
	changes := j.GetUncommittedChanges()
	require.Len(t, changes, 1, "completion raises exactly one discrete event and nothing else")
}

func TestCapabilityJourney_UpdateProgress_WhileActive_Succeeds_Rule7(t *testing.T) {
	j := inFlightJourney(t)

	err := j.UpdateProgress(newProgress(t, 60), journeyActor)

	require.NoError(t, err)
	require.NotNil(t, j.Progress())
	assert.Equal(t, 60, j.Progress().Value())
	changes := j.GetUncommittedChanges()
	require.Len(t, changes, 1)
	evt, ok := changes[0].(events.JourneyProgressUpdated)
	require.True(t, ok)
	assert.Equal(t, 60, evt.Progress)
	assert.Equal(t, journeyActor, evt.UpdatedBy)
	assert.False(t, evt.OccurredOn.IsZero())
}

func TestCapabilityJourney_EditsRejectedOnceTerminal_Rules6_7_8(t *testing.T) {
	edits := []struct {
		name string
		act  func(*testing.T, *CapabilityJourney) error
	}{
		{"update progress", func(t *testing.T, j *CapabilityJourney) error {
			return j.UpdateProgress(newProgress(t, 10), journeyActor)
		}},
		{"update details", func(t *testing.T, j *CapabilityJourney) error {
			return j.UpdateDetails(detailsFacts(t, "n", nil, ""))
		}},
		{"change source applications", func(t *testing.T, j *CapabilityJourney) error {
			return j.ChangeSourceApplications(newComponentRefs(t, 1), journeyActor)
		}},
		{"add milestone", func(t *testing.T, j *CapabilityJourney) error {
			return j.AddMilestone(plannedMilestoneFacts(t, uuid.New().String(), "Late milestone"))
		}},
		{"update milestone", func(t *testing.T, j *CapabilityJourney) error {
			return j.UpdateMilestone(plannedMilestoneFacts(t, "m1", "x"))
		}},
		{"remove milestone", func(t *testing.T, j *CapabilityJourney) error {
			return j.RemoveMilestone("m1", journeyActor)
		}},
	}
	terminals := []struct {
		name  string
		setup func(*testing.T) *CapabilityJourney
	}{
		{"done", doneJourney},
		{"abandoned", abandonedJourney},
	}
	for _, terminal := range terminals {
		for _, edit := range edits {
			t.Run(edit.name+" on "+terminal.name, func(t *testing.T) {
				err := edit.act(t, terminal.setup(t))
				assert.ErrorIs(t, err, ErrJourneyFrozen)
			})
		}
	}
}

func TestCapabilityJourney_UpdateDetails_WhileActive_Succeeds_Rule9(t *testing.T) {
	j := plannedJourney(t)
	period := newTargetPeriodVO(t, 2027, 2)

	err := j.UpdateDetails(detailsFacts(t, "revised plan", &period, ""))

	require.NoError(t, err)
	assert.Equal(t, "revised plan", j.Note().Value())
	require.NotNil(t, j.TargetPeriod())
	assert.Equal(t, 2027, j.TargetPeriod().Year())
	assert.Equal(t, 2, j.TargetPeriod().Quarter())
}

func TestCapabilityJourney_UpdateDetails_ClearsTargetPeriod_FullReplace(t *testing.T) {
	j := plannedJourney(t)
	period := newTargetPeriodVO(t, 2027, 2)
	require.NoError(t, j.UpdateDetails(detailsFacts(t, "n", &period, "")))
	require.NotNil(t, j.TargetPeriod())

	err := j.UpdateDetails(detailsFacts(t, "n", nil, ""))

	require.NoError(t, err)
	assert.Nil(t, j.TargetPeriod())
}

func TestCapabilityJourney_UpdateDetails_ResultingName_OnlyForMove(t *testing.T) {
	j := plannedJourney(t)

	err := j.UpdateDetails(detailsFacts(t, "n", nil, "New name"))

	assert.ErrorIs(t, err, ErrJourneyMoveFieldsOnNonMove)
}

func TestCapabilityJourney_UpdateDetails_ResultingName_EditableForMove_Rule5(t *testing.T) {
	j, _, _ := plannedMoveJourney(t)
	j.MarkChangesAsCommitted()

	err := j.UpdateDetails(detailsFacts(t, "n", nil, "Freight invoicing v2"))

	require.NoError(t, err)
	assert.Equal(t, "Freight invoicing v2", j.ResultingName())
}

func TestCapabilityJourney_ChangeSourceApplications_RevalidatesCardinality_Rule3(t *testing.T) {
	j, err := planWith(t, journeyOpts{kind: valueobjects.JourneyKindConsolidation, fromAppCount: 2})
	require.NoError(t, err)
	j.MarkChangesAsCommitted()

	err = j.ChangeSourceApplications(newComponentRefs(t, 0), journeyActor)
	assert.ErrorIs(t, err, valueobjects.ErrInvalidSourceApplicationCount)

	err = j.ChangeSourceApplications(newComponentRefs(t, 1), journeyActor)
	assert.NoError(t, err)
}

func TestCapabilityJourney_ChangeSourceApplications_RevalidatesTargetNotInSources_Rule4(t *testing.T) {
	j := plannedJourney(t)

	err := j.ChangeSourceApplications([]valueobjects.ApplicationRef{j.ToApp()}, journeyActor)

	assert.ErrorIs(t, err, ErrJourneyTargetAmongSources)
}

func TestCapabilityJourney_ChangeSourceApplications_Succeeds(t *testing.T) {
	j := plannedJourney(t)
	newSources := newComponentRefs(t, 2)

	err := j.ChangeSourceApplications(newSources, journeyActor)

	require.NoError(t, err)
	assert.Len(t, j.FromApps(), 2)
	changes := j.GetUncommittedChanges()
	require.Len(t, changes, 1)
	evt, ok := changes[0].(events.JourneySourceApplicationsChanged)
	require.True(t, ok)
	assert.Equal(t, journeyActor, evt.ChangedBy)
	assert.False(t, evt.OccurredOn.IsZero())
}

func TestCapabilityJourney_AddMilestone_WhileActive_Succeeds_Rule8(t *testing.T) {
	j := plannedJourney(t)
	period := newTargetPeriodVO(t, 2026, 4)

	err := j.AddMilestone(MilestoneFacts{
		MilestoneID:  uuid.New().String(),
		Label:        "Route cutover",
		TargetPeriod: &period,
		Status:       newMilestoneStatusVO(t, valueobjects.MilestoneStatusPlanned),
		Actor:        journeyActor,
	})

	require.NoError(t, err)
	require.Len(t, j.Milestones(), 1)
	assert.Equal(t, "Route cutover", j.Milestones()[0].Label())
	changes := j.GetUncommittedChanges()
	require.Len(t, changes, 1)
	evt, ok := changes[0].(events.JourneyMilestoneAdded)
	require.True(t, ok)
	assert.Equal(t, journeyActor, evt.AddedBy)
	assert.False(t, evt.OccurredOn.IsZero())
}

func TestCapabilityJourney_AddMilestone_EmptyLabel_Fails_Rule8(t *testing.T) {
	j := plannedJourney(t)

	err := j.AddMilestone(plannedMilestoneFacts(t, uuid.New().String(), "   "))

	assert.ErrorIs(t, err, entities.ErrMilestoneLabelRequired)
}

func TestCapabilityJourney_UpdateMilestone_Succeeds_Rule8(t *testing.T) {
	j := plannedJourney(t)
	addPlannedMilestone(t, j, "m1", "Route cutover")
	j.MarkChangesAsCommitted()

	err := j.UpdateMilestone(MilestoneFacts{
		MilestoneID: "m1",
		Label:       "Route cutover done",
		Status:      newMilestoneStatusVO(t, valueobjects.MilestoneStatusDone),
		Actor:       journeyActor,
	})

	require.NoError(t, err)
	require.Len(t, j.Milestones(), 1)
	assert.Equal(t, "Route cutover done", j.Milestones()[0].Label())
	assert.Equal(t, valueobjects.MilestoneStatusDone, j.Milestones()[0].Status().Value())
	changes := j.GetUncommittedChanges()
	require.Len(t, changes, 1)
	evt, ok := changes[0].(events.JourneyMilestoneUpdated)
	require.True(t, ok)
	assert.Equal(t, journeyActor, evt.UpdatedBy)
	assert.False(t, evt.OccurredOn.IsZero())
}

func TestCapabilityJourney_UpdateMilestone_UnknownID_Fails(t *testing.T) {
	j := plannedJourney(t)

	err := j.UpdateMilestone(plannedMilestoneFacts(t, "missing", "x"))

	assert.ErrorIs(t, err, ErrJourneyMilestoneNotFound)
}

func TestCapabilityJourney_RemoveMilestone_Succeeds_Rule8(t *testing.T) {
	j := plannedJourney(t)
	addPlannedMilestone(t, j, "m1", "First")
	addPlannedMilestone(t, j, "m2", "Second")
	j.MarkChangesAsCommitted()

	err := j.RemoveMilestone("m1", journeyActor)

	require.NoError(t, err)
	require.Len(t, j.Milestones(), 1)
	assert.Equal(t, "m2", j.Milestones()[0].ID(), "removal compacts the ordered list without leaving gaps")
	changes := j.GetUncommittedChanges()
	require.Len(t, changes, 1)
	evt, ok := changes[0].(events.JourneyMilestoneRemoved)
	require.True(t, ok)
	assert.Equal(t, "m1", evt.MilestoneID)
	assert.Equal(t, journeyActor, evt.RemovedBy)
	assert.False(t, evt.OccurredOn.IsZero())
}

func TestCapabilityJourney_RemoveMilestone_UnknownID_Fails(t *testing.T) {
	j := plannedJourney(t)

	err := j.RemoveMilestone("missing", journeyActor)

	assert.ErrorIs(t, err, ErrJourneyMilestoneNotFound)
}

func TestCapabilityJourney_MilestonesPreserveInsertionOrder_Rule8(t *testing.T) {
	j := plannedJourney(t)
	addPlannedMilestone(t, j, "m1", "First")
	addPlannedMilestone(t, j, "m2", "Second")
	addPlannedMilestone(t, j, "m3", "Third")

	ids := make([]string, len(j.Milestones()))
	for i, m := range j.Milestones() {
		ids[i] = m.ID()
	}
	assert.Equal(t, []string{"m1", "m2", "m3"}, ids)
}

func TestLoadCapabilityJourneyFromHistory_ReconstructsFullLifecycle(t *testing.T) {
	j, err := planWith(t, journeyOpts{kind: valueobjects.JourneyKindMigration, fromAppCount: 2, note: "route by route"})
	require.NoError(t, err)
	require.NoError(t, j.Start(journeyActor))
	require.NoError(t, j.UpdateProgress(newProgress(t, 40), journeyActor))
	addPlannedMilestone(t, j, "m1", "Route 1 live")
	addPlannedMilestone(t, j, "m2", "Route 2 live")
	require.NoError(t, j.UpdateMilestone(MilestoneFacts{
		MilestoneID: "m1",
		Label:       "Route 1 live",
		Status:      newMilestoneStatusVO(t, valueobjects.MilestoneStatusDone),
		Actor:       journeyActor,
	}))
	require.NoError(t, j.RemoveMilestone("m2", journeyActor))
	require.NoError(t, j.UpdateProgress(newProgress(t, 100), journeyActor))
	require.NoError(t, j.Complete(journeyActor))
	history := j.GetUncommittedChanges()

	loaded, err := LoadCapabilityJourneyFromHistory(history)

	require.NoError(t, err)
	assert.Equal(t, j.ID(), loaded.ID())
	assert.Equal(t, j.CapabilityID().Value(), loaded.CapabilityID().Value())
	assert.Equal(t, j.Kind().Value(), loaded.Kind().Value())
	assert.Equal(t, j.Status().Value(), loaded.Status().Value())
	assert.Equal(t, j.Note().Value(), loaded.Note().Value())
	require.NotNil(t, loaded.Progress())
	assert.Equal(t, 100, loaded.Progress().Value())
	require.Len(t, loaded.Milestones(), 1)
	assert.Equal(t, "m1", loaded.Milestones()[0].ID())
	assert.Equal(t, valueobjects.MilestoneStatusDone, loaded.Milestones()[0].Status().Value())
	assert.Empty(t, loaded.GetUncommittedChanges())
}

func TestLoadCapabilityJourneyFromHistory_ReconstructsMoveJourney(t *testing.T) {
	j, domainRef, parentRef := plannedMoveJourney(t)
	history := j.GetUncommittedChanges()

	loaded, err := LoadCapabilityJourneyFromHistory(history)

	require.NoError(t, err)
	assertMoveDestination(t, loaded, domainRef, parentRef)
}

func TestApplyCapabilityJourney_UnknownEvent_Fails(t *testing.T) {
	_, err := LoadCapabilityJourneyFromHistory([]domain.DomainEvent{unknownCapabilityJourneyEventForTest{}})
	assert.ErrorIs(t, err, ErrUnknownCapabilityJourneyEvent)
}

type unknownCapabilityJourneyEventForTest struct{}

func (unknownCapabilityJourneyEventForTest) AggregateID() string               { return "" }
func (unknownCapabilityJourneyEventForTest) EventType() string                 { return "UnknownEvent" }
func (unknownCapabilityJourneyEventForTest) EventData() map[string]interface{} { return nil }
func (unknownCapabilityJourneyEventForTest) OccurredAt() time.Time             { return time.Time{} }

func journeyWithThreeMilestones(t *testing.T) *CapabilityJourney {
	t.Helper()
	j := plannedJourney(t)
	addPlannedMilestone(t, j, "m1", "Contract signed")
	addPlannedMilestone(t, j, "m2", "Rollout")
	addPlannedMilestone(t, j, "m3", "Pilot")
	j.MarkChangesAsCommitted()
	return j
}

func milestoneIDs(j *CapabilityJourney) []string {
	ids := make([]string, len(j.Milestones()))
	for i, m := range j.Milestones() {
		ids[i] = m.ID()
	}
	return ids
}

func TestCapabilityJourney_ReorderMilestones_RecordsOneEventWithResultingOrder_Rules1_3(t *testing.T) {
	j := journeyWithThreeMilestones(t)

	err := j.ReorderMilestones([]string{"m1", "m3", "m2"}, journeyActor)

	require.NoError(t, err)
	assert.Equal(t, []string{"m1", "m3", "m2"}, milestoneIDs(j))
	assert.Equal(t, "Pilot", j.Milestones()[1].Label(), "labels travel with their milestone")
	changes := j.GetUncommittedChanges()
	require.Len(t, changes, 1)
	evt, ok := changes[0].(events.JourneyMilestonesReordered)
	require.True(t, ok)
	assert.Equal(t, []string{"m1", "m3", "m2"}, evt.MilestoneIDs)
	assert.Equal(t, journeyActor, evt.ReorderedBy)
	assert.False(t, evt.OccurredOn.IsZero())
}

func TestCapabilityJourney_ReorderMilestones_RejectsIncompleteDuplicateOrUnknown_Rule1(t *testing.T) {
	cases := []struct {
		name  string
		order []string
		want  error
	}{
		{name: "omits one", order: []string{"m1", "m2"}, want: ErrJourneyMilestoneOrderIncomplete},
		{name: "repeats one", order: []string{"m1", "m2", "m2"}, want: ErrJourneyMilestoneOrderDuplicate},
		{name: "unknown id", order: []string{"m1", "m2", "m9"}, want: ErrJourneyMilestoneNotFound},
		{name: "too many", order: []string{"m1", "m2", "m3", "m4"}, want: ErrJourneyMilestoneOrderIncomplete},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			j := journeyWithThreeMilestones(t)

			err := j.ReorderMilestones(tc.order, journeyActor)

			assert.ErrorIs(t, err, tc.want)
			assert.Equal(t, []string{"m1", "m2", "m3"}, milestoneIDs(j))
			assert.Empty(t, j.GetUncommittedChanges())
		})
	}
}

func TestCapabilityJourney_ReorderMilestones_TerminalJourneyRejected_Rule2(t *testing.T) {
	j := journeyWithThreeMilestones(t)
	require.NoError(t, j.Abandon(journeyActor))
	j.MarkChangesAsCommitted()

	err := j.ReorderMilestones([]string{"m3", "m2", "m1"}, journeyActor)

	assert.ErrorIs(t, err, ErrJourneyFrozen)
	assert.Empty(t, j.GetUncommittedChanges())
}

func TestCapabilityJourney_ReorderMilestones_NoOpRejected_Rule4(t *testing.T) {
	j := journeyWithThreeMilestones(t)

	err := j.ReorderMilestones([]string{"m1", "m2", "m3"}, journeyActor)

	assert.ErrorIs(t, err, ErrJourneyMilestoneOrderUnchanged)
	assert.Empty(t, j.GetUncommittedChanges())
}

func TestCapabilityJourney_ReorderMilestones_OrderStableUnderAddAndRemove_Rule5(t *testing.T) {
	j := journeyWithThreeMilestones(t)
	require.NoError(t, j.ReorderMilestones([]string{"m3", "m1", "m2"}, journeyActor))

	addPlannedMilestone(t, j, "m4", "Go live")
	assert.Equal(t, []string{"m3", "m1", "m2", "m4"}, milestoneIDs(j), "added milestone appends")

	require.NoError(t, j.RemoveMilestone("m1", journeyActor))
	assert.Equal(t, []string{"m3", "m2", "m4"}, milestoneIDs(j), "removal compacts without disturbing the rest")
}

func TestLoadCapabilityJourneyFromHistory_ReplaysReorder_Rule3(t *testing.T) {
	j := plannedJourney(t)
	addPlannedMilestone(t, j, "m1", "Contract signed")
	addPlannedMilestone(t, j, "m2", "Rollout")
	addPlannedMilestone(t, j, "m3", "Pilot")
	require.NoError(t, j.ReorderMilestones([]string{"m1", "m3", "m2"}, journeyActor))

	loaded, err := LoadCapabilityJourneyFromHistory(j.GetUncommittedChanges())

	require.NoError(t, err)
	assert.Equal(t, []string{"m1", "m3", "m2"}, milestoneIDs(loaded))
}

func TestLoadCapabilityJourneyFromHistory_CorruptReorder_Fails(t *testing.T) {
	j := plannedJourney(t)
	addPlannedMilestone(t, j, "m1", "Contract signed")
	history := append(j.GetUncommittedChanges(), events.NewJourneyMilestonesReordered(events.JourneyMilestonesReorderedFields{
		ID: j.ID(), MilestoneIDs: []string{"ghost"}, ReorderedBy: journeyActor,
	}))

	_, err := LoadCapabilityJourneyFromHistory(history)

	assert.ErrorIs(t, err, ErrCorruptedCapabilityJourneyEvent)
}
