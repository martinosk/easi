package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRelationType_Valid(t *testing.T) {
	rt, err := NewRelationType("Triggers")
	assert.NoError(t, err)
	assert.Equal(t, RelationTypeTriggers, rt)
}

func TestNewRelationType_InvalidType(t *testing.T) {
	_, err := NewRelationType("Invalid")
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidRelationType, err)
}

func TestNewRelationType_Empty(t *testing.T) {
	_, err := NewRelationType("")
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidRelationType, err)
}

func TestRelationType_Equals(t *testing.T) {
	triggers1, _ := NewRelationType("Triggers")
	triggers2, _ := NewRelationType("Triggers")
	serves, _ := NewRelationType("Serves")

	assert.True(t, triggers1.Equals(triggers2))
	assert.False(t, triggers1.Equals(serves))
}
