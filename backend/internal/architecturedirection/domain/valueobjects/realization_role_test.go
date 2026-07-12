package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRealizationRole_ValidValues(t *testing.T) {
	for _, v := range []string{RealizationRoleStandard, RealizationRoleLegacy} {
		role, err := NewRealizationRole(v)
		require.NoError(t, err)
		assert.Equal(t, v, role.Value())
	}
}

func TestNewRealizationRole_InvalidValue_Fails(t *testing.T) {
	_, err := NewRealizationRole("Standard")
	assert.ErrorIs(t, err, ErrInvalidRealizationRole)
}

func TestNewRealizationRole_Empty_Fails(t *testing.T) {
	_, err := NewRealizationRole("")
	assert.ErrorIs(t, err, ErrInvalidRealizationRole)
}

func TestRealizationRole_Equals(t *testing.T) {
	a, _ := NewRealizationRole(RealizationRoleStandard)
	b, _ := NewRealizationRole(RealizationRoleStandard)
	c, _ := NewRealizationRole(RealizationRoleLegacy)
	assert.True(t, a.Equals(b))
	assert.False(t, a.Equals(c))
}
