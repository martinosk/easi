package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMilestoneStatus_AllValid(t *testing.T) {
	for _, v := range []string{"planned", "in-flight", "done"} {
		s, err := NewMilestoneStatus(v)
		require.NoError(t, err)
		assert.Equal(t, v, s.Value())
	}
}

func TestNewMilestoneStatus_Invalid(t *testing.T) {
	_, err := NewMilestoneStatus("blocked")
	assert.ErrorIs(t, err, ErrInvalidMilestoneStatus)
}

func TestMilestoneStatus_Equals(t *testing.T) {
	a, _ := NewMilestoneStatus("done")
	b, _ := NewMilestoneStatus("done")
	c, _ := NewMilestoneStatus("planned")
	assert.True(t, a.Equals(b))
	assert.False(t, a.Equals(c))
}
