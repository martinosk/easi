package valueobjects

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEnterpriseCapabilityName_Valid(t *testing.T) {
	name, err := NewEnterpriseCapabilityName("Payroll")
	require.NoError(t, err)
	assert.Equal(t, "Payroll", name.Value())
}

func TestNewEnterpriseCapabilityName_TrimsWhitespace(t *testing.T) {
	name, err := NewEnterpriseCapabilityName("  Payroll  ")
	require.NoError(t, err)
	assert.Equal(t, "Payroll", name.Value())
}

func TestNewEnterpriseCapabilityName_Empty(t *testing.T) {
	_, err := NewEnterpriseCapabilityName("")
	assert.Error(t, err)
	assert.Equal(t, ErrEnterpriseCapabilityNameEmpty, err)
}

func TestNewEnterpriseCapabilityName_OnlyWhitespace(t *testing.T) {
	_, err := NewEnterpriseCapabilityName("   ")
	assert.Error(t, err)
	assert.Equal(t, ErrEnterpriseCapabilityNameEmpty, err)
}

func TestNewEnterpriseCapabilityName_MaxLength(t *testing.T) {
	maxName := strings.Repeat("a", MaxEnterpriseCapabilityNameLength)
	name, err := NewEnterpriseCapabilityName(maxName)
	require.NoError(t, err)
	assert.Equal(t, maxName, name.Value())
}

func TestNewEnterpriseCapabilityName_TooLong(t *testing.T) {
	tooLong := strings.Repeat("a", MaxEnterpriseCapabilityNameLength+1)
	_, err := NewEnterpriseCapabilityName(tooLong)
	assert.Error(t, err)
	assert.Equal(t, ErrEnterpriseCapabilityNameTooLong, err)
}

func TestEnterpriseCapabilityName_Equals(t *testing.T) {
	name1, _ := NewEnterpriseCapabilityName("Payroll")
	name2, _ := NewEnterpriseCapabilityName("Payroll")
	assert.True(t, name1.Equals(name2))
}

func TestEnterpriseCapabilityName_NotEquals(t *testing.T) {
	name1, _ := NewEnterpriseCapabilityName("Payroll")
	name2, _ := NewEnterpriseCapabilityName("HR")
	assert.False(t, name1.Equals(name2))
}

func TestEnterpriseCapabilityName_EqualsLowerCase(t *testing.T) {
	name1, _ := NewEnterpriseCapabilityName("Payroll")
	assert.True(t, name1.EqualsIgnoreCase("payroll"))
	assert.True(t, name1.EqualsIgnoreCase("PAYROLL"))
}
