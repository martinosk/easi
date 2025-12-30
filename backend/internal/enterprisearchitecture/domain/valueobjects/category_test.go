package valueobjects

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCategory_Valid(t *testing.T) {
	category, err := NewCategory("HR Operations")
	require.NoError(t, err)
	assert.Equal(t, "HR Operations", category.Value())
}

func TestNewCategory_TrimsWhitespace(t *testing.T) {
	category, err := NewCategory("  HR Operations  ")
	require.NoError(t, err)
	assert.Equal(t, "HR Operations", category.Value())
}

func TestNewCategory_Empty_Allowed(t *testing.T) {
	category, err := NewCategory("")
	require.NoError(t, err)
	assert.Equal(t, "", category.Value())
	assert.True(t, category.IsEmpty())
}

func TestNewCategory_OnlyWhitespace_TreatedAsEmpty(t *testing.T) {
	category, err := NewCategory("   ")
	require.NoError(t, err)
	assert.Equal(t, "", category.Value())
	assert.True(t, category.IsEmpty())
}

func TestNewCategory_MaxLength(t *testing.T) {
	maxCategory := strings.Repeat("a", MaxCategoryLength)
	category, err := NewCategory(maxCategory)
	require.NoError(t, err)
	assert.Equal(t, maxCategory, category.Value())
}

func TestNewCategory_TooLong(t *testing.T) {
	tooLong := strings.Repeat("a", MaxCategoryLength+1)
	_, err := NewCategory(tooLong)
	assert.Error(t, err)
	assert.Equal(t, ErrCategoryTooLong, err)
}

func TestEmptyCategory(t *testing.T) {
	category := EmptyCategory()
	assert.Equal(t, "", category.Value())
	assert.True(t, category.IsEmpty())
}

func TestCategory_Equals(t *testing.T) {
	cat1, _ := NewCategory("HR Operations")
	cat2, _ := NewCategory("HR Operations")
	assert.True(t, cat1.Equals(cat2))
}

func TestCategory_NotEquals(t *testing.T) {
	cat1, _ := NewCategory("HR Operations")
	cat2, _ := NewCategory("Finance")
	assert.False(t, cat1.Equals(cat2))
}
