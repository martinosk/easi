package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSubjectType_AcceptsAllFiveValues(t *testing.T) {
	values := []string{"capability", "application", "acquired-entity", "vendor", "internal-team"}
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

func TestNewSubjectType_RejectsRetiredEnterpriseCapability(t *testing.T) {
	_, err := NewSubjectType("enterprise-capability")
	assert.ErrorIs(t, err, ErrInvalidSubjectType)
}

func TestNewSubjectType_RejectsEmptyValue(t *testing.T) {
	_, err := NewSubjectType("")
	assert.ErrorIs(t, err, ErrInvalidSubjectType)
}

func TestAllSubjectTypes_ContainsFiveDistinctValues(t *testing.T) {
	all := AllSubjectTypes()
	require.Len(t, all, 5)
	seen := map[string]bool{}
	for _, st := range all {
		seen[st.Value()] = true
	}
	assert.Len(t, seen, 5)
}

func TestSubjectType_Equals(t *testing.T) {
	a, _ := NewSubjectType("vendor")
	b, _ := NewSubjectType("vendor")
	c, _ := NewSubjectType("application")
	assert.True(t, a.Equals(b))
	assert.False(t, a.Equals(c))
}
