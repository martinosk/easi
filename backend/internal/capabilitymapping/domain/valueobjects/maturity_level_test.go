package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMaturityLevel_Valid(t *testing.T) {
	level, err := NewMaturityLevel("Product")
	assert.NoError(t, err)
	assert.Equal(t, MaturityProduct, level)
}

func TestNewMaturityLevel_Empty(t *testing.T) {
	level, err := NewMaturityLevel("")
	assert.NoError(t, err)
	assert.Equal(t, MaturityGenesis, level)
}

func TestNewMaturityLevel_InvalidValue(t *testing.T) {
	_, err := NewMaturityLevel("InvalidLevel")
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidMaturityLevel, err)
}

func TestMaturityLevel_NumericValue(t *testing.T) {
	assert.Equal(t, 1, MaturityGenesis.NumericValue())
	assert.Equal(t, 2, MaturityCustomBuild.NumericValue())
	assert.Equal(t, 3, MaturityProduct.NumericValue())
	assert.Equal(t, 4, MaturityCommodity.NumericValue())
}

func TestMaturityLevel_Equals(t *testing.T) {
	level1 := MaturityProduct
	level2 := MaturityProduct
	level3 := MaturityGenesis

	assert.True(t, level1.Equals(level2))
	assert.False(t, level1.Equals(level3))
}
