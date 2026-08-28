package entities

import (
	"testing"

	"easi/backend/internal/architecturedirection/domain/valueobjects"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func plannedMilestone(t *testing.T, id, label string) Milestone {
	t.Helper()
	m, err := NewMilestone(id, label, nil, newMilestoneStatus(t, valueobjects.MilestoneStatusPlanned))
	require.NoError(t, err)
	return m
}

func threeMilestones(t *testing.T) Milestones {
	t.Helper()
	return NoMilestones().
		Record(plannedMilestone(t, "m1", "Contract signed")).
		Record(plannedMilestone(t, "m2", "Rollout")).
		Record(plannedMilestone(t, "m3", "Pilot"))
}

func ids(ms Milestones) []string {
	out := make([]string, 0, ms.Count())
	for _, m := range ms.List() {
		out = append(out, m.ID())
	}
	return out
}

func TestMilestones_Record_AppendsNewAndReplacesKnown_Rule8(t *testing.T) {
	ms := threeMilestones(t)

	renamed := ms.Record(plannedMilestone(t, "m2", "Rollout to all routes"))

	assert.Equal(t, []string{"m1", "m2", "m3"}, ids(renamed), "a known milestone keeps its place")
	assert.Equal(t, "Rollout to all routes", renamed.List()[1].Label())
	assert.Equal(t, []string{"m1", "m2", "m3", "m4"}, ids(renamed.Record(plannedMilestone(t, "m4", "Go live"))))
}

func TestMilestones_Remove_CompactsWithoutDisturbingOthers_Rule5(t *testing.T) {
	ms := threeMilestones(t)

	assert.Equal(t, []string{"m1", "m3"}, ids(ms.Remove("m2")))
	assert.Equal(t, []string{"m1", "m2", "m3"}, ids(ms.Remove("missing")))
}

func TestMilestones_Has(t *testing.T) {
	ms := threeMilestones(t)

	assert.True(t, ms.Has("m2"))
	assert.False(t, ms.Has("m9"))
}

func TestMilestones_ValidateSequence_RejectsIncompleteDuplicateUnknown_Rule1(t *testing.T) {
	ms := threeMilestones(t)
	cases := []struct {
		name string
		seq  []string
		want error
	}{
		{name: "omits one", seq: []string{"m1", "m2"}, want: ErrMilestoneOrderIncomplete},
		{name: "one too many", seq: []string{"m1", "m2", "m3", "m4"}, want: ErrMilestoneOrderIncomplete},
		{name: "repeats one", seq: []string{"m1", "m2", "m2"}, want: ErrMilestoneOrderDuplicate},
		{name: "unknown id", seq: []string{"m1", "m2", "m9"}, want: ErrMilestoneNotFound},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.ErrorIs(t, ms.ValidateSequence(tc.seq), tc.want)
		})
	}
	assert.NoError(t, ms.ValidateSequence([]string{"m3", "m1", "m2"}))
}

func TestMilestones_InSequence(t *testing.T) {
	ms := threeMilestones(t)

	assert.True(t, ms.InSequence([]string{"m1", "m2", "m3"}))
	assert.False(t, ms.InSequence([]string{"m1", "m3", "m2"}))
}

func TestMilestones_Reorder_ReturnsRequestedSequenceKeepingEachMilestone(t *testing.T) {
	ms := threeMilestones(t)

	reordered, err := ms.Reorder([]string{"m3", "m1", "m2"})

	require.NoError(t, err)
	assert.Equal(t, []string{"m3", "m1", "m2"}, ids(reordered))
	assert.Equal(t, "Pilot", reordered.List()[0].Label())
	assert.Equal(t, []string{"m1", "m2", "m3"}, ids(ms), "the original sequence is untouched")
}

func TestMilestones_Reorder_InvalidSequence_Fails_Rule1(t *testing.T) {
	_, err := threeMilestones(t).Reorder([]string{"m1"})

	assert.ErrorIs(t, err, ErrMilestoneOrderIncomplete)
}

func TestMilestones_List_IsACopy(t *testing.T) {
	ms := threeMilestones(t)

	listed := ms.List()
	listed[0] = plannedMilestone(t, "hacked", "Hacked")

	assert.Equal(t, "m1", ms.List()[0].ID())
	assert.Empty(t, NoMilestones().List())
	assert.Equal(t, 0, NoMilestones().Count())
}
