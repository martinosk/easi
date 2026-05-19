package valueobjects

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPlacement_Valid(t *testing.T) {
	domainID := uuid.New().String()
	p, err := NewPlacement(domainID, "Unified Payroll")
	require.NoError(t, err)
	assert.Equal(t, domainID, p.TargetBusinessDomainID())
	assert.Equal(t, "Unified Payroll", p.ResultingName())
}

func TestNewPlacement_NoName(t *testing.T) {
	domainID := uuid.New().String()
	p, err := NewPlacement(domainID, "")
	require.NoError(t, err)
	assert.Empty(t, p.ResultingName())
}

func TestNewPlacement_InvalidDomainID(t *testing.T) {
	_, err := NewPlacement("not-a-uuid", "")
	assert.Error(t, err)
}

func TestNewPlacement_EmptyDomainID(t *testing.T) {
	_, err := NewPlacement("", "")
	assert.Error(t, err)
}

func TestNewPlacement_NameTooLong(t *testing.T) {
	domainID := uuid.New().String()
	tooLong := strings.Repeat("a", MaxResultingNameLength+1)
	_, err := NewPlacement(domainID, tooLong)
	assert.ErrorIs(t, err, ErrResultingNameTooLong)
}

func TestPlacement_TrimsName(t *testing.T) {
	domainID := uuid.New().String()
	p, err := NewPlacement(domainID, "   spaced   ")
	require.NoError(t, err)
	assert.Equal(t, "spaced", p.ResultingName())
}

func TestPlacement_Equals(t *testing.T) {
	domainID := uuid.New().String()
	p1, _ := NewPlacement(domainID, "Name")
	p2, _ := NewPlacement(domainID, "Name")
	p3, _ := NewPlacement(domainID, "Other")
	p4, _ := NewPlacement(uuid.New().String(), "Name")

	assert.True(t, p1.Equals(p2))
	assert.False(t, p1.Equals(p3))
	assert.False(t, p1.Equals(p4))
}
