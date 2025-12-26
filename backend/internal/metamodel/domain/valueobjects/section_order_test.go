package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSectionOrder_ValidOrders(t *testing.T) {
	for order := 1; order <= 4; order++ {
		t.Run("order "+string(rune('0'+order)), func(t *testing.T) {
			so, err := NewSectionOrder(order)

			require.NoError(t, err)
			assert.Equal(t, order, so.Value())
		})
	}
}

func TestNewSectionOrder_InvalidOrders(t *testing.T) {
	testCases := []struct {
		name  string
		order int
	}{
		{"zero", 0},
		{"negative", -1},
		{"five", 5},
		{"large number", 100},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewSectionOrder(tc.order)

			assert.Error(t, err)
			assert.Equal(t, ErrSectionOrderOutOfRange, err)
		})
	}
}

func TestSectionOrder_Equals(t *testing.T) {
	so1, _ := NewSectionOrder(1)
	so2, _ := NewSectionOrder(1)
	so3, _ := NewSectionOrder(2)

	assert.True(t, so1.Equals(so2))
	assert.False(t, so1.Equals(so3))
}
