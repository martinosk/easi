package ratelimit_test

import (
	"testing"

	"easi/backend/internal/archassistant/infrastructure/ratelimit"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestLimiter(t *testing.T) *ratelimit.Limiter {
	t.Helper()
	l := ratelimit.NewLimiter()
	t.Cleanup(l.Stop)
	return l
}

func TestLimiter_AcquireStream(t *testing.T) {
	limiter := newTestLimiter(t)

	err := limiter.AcquireStream("user-1")
	require.NoError(t, err)

	err = limiter.AcquireStream("user-1")
	assert.ErrorIs(t, err, ratelimit.ErrConcurrentStream)

	err = limiter.AcquireStream("user-2")
	assert.NoError(t, err)
}

func TestLimiter_ReleaseStream(t *testing.T) {
	limiter := newTestLimiter(t)

	require.NoError(t, limiter.AcquireStream("user-1"))
	limiter.ReleaseStream("user-1")

	err := limiter.AcquireStream("user-1")
	assert.NoError(t, err)
}

func TestLimiter_AllowMessage_WithinLimits(t *testing.T) {
	limiter := newTestLimiter(t)

	err := limiter.AllowMessage("user-1", "tenant-1")
	assert.NoError(t, err)
}

func TestLimiter_AllowMessage_RateLimits(t *testing.T) {
	tests := []struct {
		name        string
		fillCount   int
		fillUserFn  func(i int) string
		fillTenant  string
		finalUser   string
		finalTenant string
		wantErr     error
	}{
		{
			name:        "user minute limit exceeded",
			fillCount:   10,
			fillUserFn:  func(int) string { return "user-1" },
			fillTenant:  "tenant-1",
			finalUser:   "user-1",
			finalTenant: "tenant-1",
			wantErr:     ratelimit.ErrUserMinuteLimit,
		},
		{
			name:        "different users have separate limits",
			fillCount:   10,
			fillUserFn:  func(int) string { return "user-1" },
			fillTenant:  "tenant-1",
			finalUser:   "user-2",
			finalTenant: "tenant-1",
			wantErr:     nil,
		},
		{
			name:        "tenant minute limit exceeded",
			fillCount:   50,
			fillUserFn:  func(i int) string { return "user-" + string(rune('a'+i%26)) },
			fillTenant:  "tenant-1",
			finalUser:   "new-user",
			finalTenant: "tenant-1",
			wantErr:     ratelimit.ErrTenantMinuteLimit,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limiter := newTestLimiter(t)

			for i := 0; i < tt.fillCount; i++ {
				require.NoError(t, limiter.AllowMessage(tt.fillUserFn(i), tt.fillTenant))
			}

			err := limiter.AllowMessage(tt.finalUser, tt.finalTenant)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
