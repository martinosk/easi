package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTimeAssessmentID(t *testing.T) {
	id := NewTimeAssessmentID()
	assert.NotEmpty(t, id.Value())
}

func TestNewTimeAssessmentIDFromString_Valid(t *testing.T) {
	id := NewTimeAssessmentID()
	parsed, err := NewTimeAssessmentIDFromString(id.Value())
	require.NoError(t, err)
	assert.Equal(t, id.Value(), parsed.Value())
}

func TestNewTimeAssessmentIDFromString_Empty(t *testing.T) {
	_, err := NewTimeAssessmentIDFromString("")
	assert.Error(t, err)
}

func TestNewTimeAssessmentIDFromString_Invalid(t *testing.T) {
	_, err := NewTimeAssessmentIDFromString("not-a-uuid")
	assert.Error(t, err)
}

func TestTimeAssessmentID_Equals(t *testing.T) {
	id := NewTimeAssessmentID()
	parsed, _ := NewTimeAssessmentIDFromString(id.Value())
	assert.True(t, id.Equals(parsed))
}
