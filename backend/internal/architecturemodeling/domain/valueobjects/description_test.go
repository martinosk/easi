package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDescription_Empty(t *testing.T) {
	description := MustNewDescription("")

	assert.Equal(t, "", description.Value())
	assert.True(t, description.IsEmpty())
}

func TestDescription_IsEmpty(t *testing.T) {
	emptyDesc := MustNewDescription("")
	assert.True(t, emptyDesc.IsEmpty())

	whitespaceDesc := MustNewDescription("   ")
	assert.True(t, whitespaceDesc.IsEmpty())

	nonEmptyDesc := MustNewDescription("Some content")
	assert.False(t, nonEmptyDesc.IsEmpty())
}

func TestDescription_Equals(t *testing.T) {
	desc1 := MustNewDescription("User authentication module")
	desc2 := MustNewDescription("User authentication module")
	desc3 := MustNewDescription("Different description")
	emptyDesc1 := MustNewDescription("")
	emptyDesc2 := MustNewDescription("")

	assert.True(t, desc1.Equals(desc2))
	assert.False(t, desc1.Equals(desc3))
	assert.True(t, emptyDesc1.Equals(emptyDesc2))
	assert.False(t, emptyDesc1.Equals(desc1))
}
