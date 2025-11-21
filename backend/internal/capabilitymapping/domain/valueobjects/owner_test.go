package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewOwner_ValidValue(t *testing.T) {
	owner := NewOwner("Platform Tribe - John Doe")
	assert.Equal(t, "Platform Tribe - John Doe", owner.Value())
}

func TestNewOwner_TrimSpace(t *testing.T) {
	owner := NewOwner("  Data Team - Jane Smith  ")
	assert.Equal(t, "Data Team - Jane Smith", owner.Value())
}

func TestNewOwner_Empty(t *testing.T) {
	owner := NewOwner("")
	assert.True(t, owner.IsEmpty())
}

func TestOwner_Value(t *testing.T) {
	owner := NewOwner("Engineering Tribe - Bob Wilson")
	assert.Equal(t, "Engineering Tribe - Bob Wilson", owner.Value())
}

func TestOwner_String(t *testing.T) {
	owner := NewOwner("Security Team - Alice Brown")
	assert.Equal(t, "Security Team - Alice Brown", owner.String())
}

func TestOwner_Equals(t *testing.T) {
	owner1 := NewOwner("Platform Tribe - John Doe")
	owner2 := NewOwner("Platform Tribe - John Doe")
	owner3 := NewOwner("Data Team - Jane Smith")

	assert.True(t, owner1.Equals(owner2))
	assert.False(t, owner1.Equals(owner3))
}

func TestOwner_IsEmpty(t *testing.T) {
	emptyOwner := NewOwner("")
	nonEmptyOwner := NewOwner("Team Lead")

	assert.True(t, emptyOwner.IsEmpty())
	assert.False(t, nonEmptyOwner.IsEmpty())
}
