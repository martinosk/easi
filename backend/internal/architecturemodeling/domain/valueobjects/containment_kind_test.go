package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewContainmentKind_AcceptsDeclaredKinds(t *testing.T) {
	cases := []struct {
		value         string
		isComposition bool
	}{
		{ContainmentComposition, true},
		{ContainmentAggregation, false},
	}
	for _, tc := range cases {
		t.Run(tc.value, func(t *testing.T) {
			kind, err := NewContainmentKind(tc.value)

			require.NoError(t, err)
			assert.Equal(t, tc.value, kind.String())
			assert.Equal(t, tc.isComposition, kind.IsComposition())
		})
	}
}

func TestNewContainmentKind_RejectsUnknownKind(t *testing.T) {
	for _, value := range []string{"", "containment", "Composition"} {
		t.Run(value, func(t *testing.T) {
			_, err := NewContainmentKind(value)

			assert.ErrorIs(t, err, ErrInvalidContainmentKind)
		})
	}
}

func TestContainmentKind_Equals(t *testing.T) {
	composition, _ := NewContainmentKind(ContainmentComposition)
	sameComposition, _ := NewContainmentKind(ContainmentComposition)
	aggregation, _ := NewContainmentKind(ContainmentAggregation)

	assert.True(t, composition.Equals(sameComposition))
	assert.False(t, composition.Equals(aggregation))
	assert.False(t, composition.Equals(UnknownHostingClassification()))
}
