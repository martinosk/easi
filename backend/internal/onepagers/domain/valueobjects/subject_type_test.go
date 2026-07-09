package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSubjectType_AcceptsAllSixValues(t *testing.T) {
	values := []string{"capability", "enterprise-capability", "application", "acquired-entity", "vendor", "internal-team"}
	for _, v := range values {
		st, err := NewSubjectType(v)
		require.NoError(t, err, v)
		assert.Equal(t, v, st.Value())
	}
}

func TestNewSubjectType_RejectsUnknownValue(t *testing.T) {
	_, err := NewSubjectType("business-unit")
	assert.ErrorIs(t, err, ErrInvalidSubjectType)
}

func TestNewSubjectType_RejectsEmptyValue(t *testing.T) {
	_, err := NewSubjectType("")
	assert.ErrorIs(t, err, ErrInvalidSubjectType)
}

func TestAllSubjectTypes_ContainsSixDistinctValues(t *testing.T) {
	all := AllSubjectTypes()
	require.Len(t, all, 6)
	seen := map[string]bool{}
	for _, st := range all {
		seen[st.Value()] = true
	}
	assert.Len(t, seen, 6)
}

func TestSubjectType_Equals(t *testing.T) {
	a, _ := NewSubjectType("vendor")
	b, _ := NewSubjectType("vendor")
	c, _ := NewSubjectType("application")
	assert.True(t, a.Equals(b))
	assert.False(t, a.Equals(c))
}
