package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTimeGrade_ValidValues(t *testing.T) {
	for _, v := range []string{TimeGradeInvest, TimeGradeTolerate, TimeGradeMigrate, TimeGradeEliminate} {
		grade, err := NewTimeGrade(v)
		require.NoError(t, err)
		assert.Equal(t, v, grade.Value())
	}
}

func TestNewTimeGrade_InvalidValue_Fails(t *testing.T) {
	_, err := NewTimeGrade("invest")
	assert.ErrorIs(t, err, ErrInvalidTimeGrade)
}

func TestNewTimeGrade_Empty_Fails(t *testing.T) {
	_, err := NewTimeGrade("")
	assert.ErrorIs(t, err, ErrInvalidTimeGrade)
}

func TestTimeGrade_Equals(t *testing.T) {
	a, _ := NewTimeGrade(TimeGradeMigrate)
	b, _ := NewTimeGrade(TimeGradeMigrate)
	c, _ := NewTimeGrade(TimeGradeEliminate)
	assert.True(t, a.Equals(b))
	assert.False(t, a.Equals(c))
}
