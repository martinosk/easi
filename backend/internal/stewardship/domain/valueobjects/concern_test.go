package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConcern_AcceptsEveryFixedConcern(t *testing.T) {
	for _, value := range []string{"ownership", "assessment", "documentation", "planning", "structure"} {
		concern, err := NewConcern(value)
		require.NoError(t, err, value)
		assert.Equal(t, value, concern.Value())
		assert.NotEmpty(t, concern.Label(), value)
		assert.NotEmpty(t, concern.Description(), value)
	}
}

func TestNewConcern_RejectsAnyOtherValue(t *testing.T) {
	for _, value := range []string{"", "Ownership", "model", "security"} {
		_, err := NewConcern(value)
		assert.ErrorIs(t, err, ErrUnknownConcern, value)
	}
}

func TestAllConcerns_ListsTheFixedConcernsInOrder(t *testing.T) {
	var values []string
	for _, concern := range AllConcerns() {
		values = append(values, concern.Value())
	}
	assert.Equal(t, []string{"ownership", "assessment", "documentation", "planning", "structure"}, values)
}

func TestAllConcerns_ReturnsACopy(t *testing.T) {
	concerns := AllConcerns()
	concerns[0] = Concern{}
	assert.Equal(t, "ownership", AllConcerns()[0].Value())
}
