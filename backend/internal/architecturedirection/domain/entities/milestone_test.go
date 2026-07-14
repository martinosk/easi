package entities

import (
	"strings"
	"testing"

	"easi/backend/internal/architecturedirection/domain/valueobjects"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMilestoneStatus(t *testing.T, v string) valueobjects.MilestoneStatus {
	t.Helper()
	s, err := valueobjects.NewMilestoneStatus(v)
	require.NoError(t, err)
	return s
}

func TestNewMilestone_Valid_Rule8(t *testing.T) {
	id := uuid.New().String()
	period, err := valueobjects.NewTargetPeriod(2026, 4)
	require.NoError(t, err)
	status := newMilestoneStatus(t, valueobjects.MilestoneStatusPlanned)

	m, err := NewMilestone(id, "  Route cutover  ", &period, status)

	require.NoError(t, err)
	assert.Equal(t, id, m.ID())
	assert.Equal(t, "Route cutover", m.Label())
	require.NotNil(t, m.TargetPeriod())
	assert.Equal(t, 2026, m.TargetPeriod().Year())
	assert.Equal(t, valueobjects.MilestoneStatusPlanned, m.Status().Value())
}

func TestNewMilestone_NilTargetPeriod_Allowed(t *testing.T) {
	m, err := NewMilestone(uuid.New().String(), "Kickoff", nil, newMilestoneStatus(t, valueobjects.MilestoneStatusPlanned))
	require.NoError(t, err)
	assert.Nil(t, m.TargetPeriod())
}

func TestNewMilestone_EmptyLabel_Rejected_Rule8(t *testing.T) {
	_, err := NewMilestone(uuid.New().String(), "   ", nil, newMilestoneStatus(t, valueobjects.MilestoneStatusPlanned))
	assert.ErrorIs(t, err, ErrMilestoneLabelRequired)
}

func TestNewMilestone_LabelTooLong_Rejected(t *testing.T) {
	_, err := NewMilestone(uuid.New().String(), strings.Repeat("a", 201), nil, newMilestoneStatus(t, valueobjects.MilestoneStatusPlanned))
	assert.ErrorIs(t, err, ErrMilestoneLabelTooLong)
}

func TestNewMilestone_MaxLengthLabel_Accepted(t *testing.T) {
	_, err := NewMilestone(uuid.New().String(), strings.Repeat("a", 200), nil, newMilestoneStatus(t, valueobjects.MilestoneStatusPlanned))
	assert.NoError(t, err)
}

func TestNewMilestone_PreservesGivenStatus(t *testing.T) {
	m, err := NewMilestone(uuid.New().String(), "Route cutover", nil, newMilestoneStatus(t, valueobjects.MilestoneStatusDone))
	require.NoError(t, err)
	assert.Equal(t, valueobjects.MilestoneStatusDone, m.Status().Value())
}
