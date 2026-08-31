package valueobjects

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewSetAt(t *testing.T) {
	before := time.Now().UTC()
	sa := NewSetAt()
	after := time.Now().UTC()

	assert.False(t, sa.IsZero())
	assert.True(t, sa.Value().After(before) || sa.Value().Equal(before))
	assert.True(t, sa.Value().Before(after) || sa.Value().Equal(after))
}

func TestNewSetAtFromTime(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	sa := NewSetAtFromTime(testTime)

	assert.Equal(t, testTime, sa.Value())
}

func TestNewSetAtFromTime_ConvertsToUTC(t *testing.T) {
	loc, _ := time.LoadLocation("America/New_York")
	localTime := time.Date(2024, 1, 15, 10, 30, 0, 0, loc)
	sa := NewSetAtFromTime(localTime)

	assert.Equal(t, time.UTC, sa.Value().Location())
	assert.Equal(t, localTime.UTC(), sa.Value())
}

func TestSetAt_Equals(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	sa1 := NewSetAtFromTime(testTime)
	sa2 := NewSetAtFromTime(testTime)
	sa3 := NewSetAtFromTime(testTime.Add(time.Hour))

	assert.True(t, sa1.Equals(sa2))
	assert.False(t, sa1.Equals(sa3))
}

func TestSetAt_String(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	sa := NewSetAtFromTime(testTime)

	assert.Equal(t, "2024-01-15T10:30:00Z", sa.String())
}
