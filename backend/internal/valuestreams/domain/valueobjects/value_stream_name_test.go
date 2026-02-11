package valueobjects

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewValueStreamName_Empty(t *testing.T) {
	_, err := NewValueStreamName("")
	assert.Error(t, err)
	assert.Equal(t, ErrValueStreamNameEmpty, err)
}

func TestNewValueStreamName_OnlyWhitespace(t *testing.T) {
	_, err := NewValueStreamName("   ")
	assert.Error(t, err)
	assert.Equal(t, ErrValueStreamNameEmpty, err)
}

func TestNewValueStreamName_ExceedsMaxLength(t *testing.T) {
	tooLongName := strings.Repeat("a", 101)
	_, err := NewValueStreamName(tooLongName)
	assert.Error(t, err)
	assert.Equal(t, ErrValueStreamNameTooLong, err)
}

func TestNewValueStreamName_ExactMaxLength(t *testing.T) {
	exactName := strings.Repeat("a", 100)
	name, err := NewValueStreamName(exactName)
	assert.NoError(t, err)
	assert.Equal(t, exactName, name.Value())
}

func TestNewValueStreamName_TrimsWhitespace(t *testing.T) {
	name, err := NewValueStreamName("  Customer Onboarding  ")
	assert.NoError(t, err)
	assert.Equal(t, "Customer Onboarding", name.Value())
}

func TestValueStreamName_Equals(t *testing.T) {
	name1, _ := NewValueStreamName("Customer Onboarding")
	name2, _ := NewValueStreamName("Customer Onboarding")
	name3, _ := NewValueStreamName("Order Fulfillment")

	assert.True(t, name1.Equals(name2))
	assert.False(t, name1.Equals(name3))
}

func TestValueStreamName_String(t *testing.T) {
	name, _ := NewValueStreamName("Customer Onboarding")
	assert.Equal(t, "Customer Onboarding", name.String())
}
