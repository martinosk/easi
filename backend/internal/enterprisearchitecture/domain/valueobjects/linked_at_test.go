package valueobjects

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewLinkedAt(t *testing.T) {
	before := time.Now().UTC()
	la := NewLinkedAt()
	after := time.Now().UTC()

	assert.False(t, la.IsZero())
	assert.True(t, la.Value().After(before) || la.Value().Equal(before))
	assert.True(t, la.Value().Before(after) || la.Value().Equal(after))
}

func TestNewLinkedAtFromTime(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	la := NewLinkedAtFromTime(testTime)

	assert.Equal(t, testTime, la.Value())
}

func TestNewLinkedAtFromTime_ConvertsToUTC(t *testing.T) {
	loc, _ := time.LoadLocation("America/New_York")
	localTime := time.Date(2024, 1, 15, 10, 30, 0, 0, loc)
	la := NewLinkedAtFromTime(localTime)

	assert.Equal(t, time.UTC, la.Value().Location())
	assert.Equal(t, localTime.UTC(), la.Value())
}

func TestLinkedAt_Equals(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	la1 := NewLinkedAtFromTime(testTime)
	la2 := NewLinkedAtFromTime(testTime)
	la3 := NewLinkedAtFromTime(testTime.Add(time.Hour))

	assert.True(t, la1.Equals(la2))
	assert.False(t, la1.Equals(la3))
}

func TestLinkedAt_String(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	la := NewLinkedAtFromTime(testTime)

	assert.Equal(t, "2024-01-15T10:30:00Z", la.String())
}
