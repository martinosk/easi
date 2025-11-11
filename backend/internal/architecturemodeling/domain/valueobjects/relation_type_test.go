package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRelationType_ValidTriggers(t *testing.T) {
	rt, err := NewRelationType("Triggers")
	assert.NoError(t, err)
	assert.Equal(t, RelationTypeTriggers, rt)
	assert.Equal(t, "Triggers", rt.Value())
}

func TestNewRelationType_ValidServes(t *testing.T) {
	rt, err := NewRelationType("Serves")
	assert.NoError(t, err)
	assert.Equal(t, RelationTypeServes, rt)
	assert.Equal(t, "Serves", rt.Value())
}

func TestNewRelationType_InvalidType(t *testing.T) {
	testCases := []struct {
		name  string
		value string
	}{
		{"empty string", ""},
		{"invalid type", "Invalid"},
		{"lowercase triggers", "triggers"},
		{"lowercase serves", "serves"},
		{"random string", "RandomRelation"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewRelationType(tc.value)
			assert.Error(t, err)
			assert.Equal(t, ErrInvalidRelationType, err)
		})
	}
}

func TestRelationType_Equals(t *testing.T) {
	triggers1, _ := NewRelationType("Triggers")
	triggers2, _ := NewRelationType("Triggers")
	serves, _ := NewRelationType("Serves")

	assert.True(t, triggers1.Equals(triggers2))
	assert.False(t, triggers1.Equals(serves))
	assert.False(t, serves.Equals(triggers1))
}

func TestRelationType_String(t *testing.T) {
	triggers, _ := NewRelationType("Triggers")
	serves, _ := NewRelationType("Serves")

	assert.Equal(t, "Triggers", triggers.String())
	assert.Equal(t, "Serves", serves.String())
}
