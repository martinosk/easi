package valueobjects

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUserEmail_AcceptsValidAddress(t *testing.T) {
	email, err := NewUserEmail(" admin@example.com ")
	require.NoError(t, err)
	assert.Equal(t, "admin@example.com", email.Value())
}

func TestNewUserEmail_RejectsInvalid(t *testing.T) {
	_, err := NewUserEmail("not-an-email")
	assert.ErrorIs(t, err, ErrUserEmailInvalid)

	_, err = NewUserEmail("")
	assert.ErrorIs(t, err, ErrUserEmailEmpty)
}

func TestNewTimestamp_RejectsZero(t *testing.T) {
	_, err := NewTimestamp(time.Time{})
	assert.ErrorIs(t, err, ErrTimestampZero)
}

func TestTimestampNow_ReturnsNonZeroUTC(t *testing.T) {
	ts := TimestampNow()
	assert.False(t, ts.Value().IsZero())
	assert.Equal(t, "UTC", ts.Value().Location().String())
}
