package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDescription_Empty(t *testing.T) {
	description := NewDescription("")

	assert.Equal(t, "", description.Value())
	assert.True(t, description.IsEmpty())
}

func TestDescription_IsEmpty(t *testing.T) {
	emptyDesc := NewDescription("")
	assert.True(t, emptyDesc.IsEmpty())

	whitespaceDesc := NewDescription("   ")
	assert.True(t, whitespaceDesc.IsEmpty())

	nonEmptyDesc := NewDescription("Some content")
	assert.False(t, nonEmptyDesc.IsEmpty())
}

func TestDescription_Equals(t *testing.T) {
	desc1 := NewDescription("User authentication module")
	desc2 := NewDescription("User authentication module")
	desc3 := NewDescription("Different description")
	emptyDesc1 := NewDescription("")
	emptyDesc2 := NewDescription("")

	assert.True(t, desc1.Equals(desc2))
	assert.False(t, desc1.Equals(desc3))
	assert.True(t, emptyDesc1.Equals(emptyDesc2))
	assert.False(t, emptyDesc1.Equals(desc1))
}
