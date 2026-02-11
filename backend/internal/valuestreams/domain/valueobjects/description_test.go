package valueobjects

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDescription_Empty(t *testing.T) {
	desc, err := NewDescription("")
	assert.NoError(t, err)
	assert.Equal(t, "", desc.Value())
	assert.True(t, desc.IsEmpty())
}

func TestNewDescription_Valid(t *testing.T) {
	desc, err := NewDescription("End-to-end customer onboarding process")
	assert.NoError(t, err)
	assert.Equal(t, "End-to-end customer onboarding process", desc.Value())
	assert.False(t, desc.IsEmpty())
}

func TestNewDescription_ExceedsMaxLength(t *testing.T) {
	tooLong := strings.Repeat("a", 501)
	_, err := NewDescription(tooLong)
	assert.Error(t, err)
	assert.Equal(t, ErrDescriptionTooLong, err)
}

func TestNewDescription_ExactMaxLength(t *testing.T) {
	exact := strings.Repeat("a", 500)
	desc, err := NewDescription(exact)
	assert.NoError(t, err)
	assert.Equal(t, exact, desc.Value())
}

func TestNewDescription_TrimsWhitespace(t *testing.T) {
	desc, err := NewDescription("  some description  ")
	assert.NoError(t, err)
	assert.Equal(t, "some description", desc.Value())
}

func TestDescription_Equals(t *testing.T) {
	desc1, _ := NewDescription("Description A")
	desc2, _ := NewDescription("Description A")
	desc3, _ := NewDescription("Description B")

	assert.True(t, desc1.Equals(desc2))
	assert.False(t, desc1.Equals(desc3))
}

func TestMustNewDescription_Panics(t *testing.T) {
	tooLong := strings.Repeat("a", 501)
	assert.Panics(t, func() {
		MustNewDescription(tooLong)
	})
}

func TestMustNewDescription_Valid(t *testing.T) {
	desc := MustNewDescription("Valid description")
	assert.Equal(t, "Valid description", desc.Value())
}
