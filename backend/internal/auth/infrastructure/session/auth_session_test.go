package session

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sharedvo "easi/backend/internal/shared/domain/valueobjects"
)

func TestNewPreAuthSession(t *testing.T) {
	tenantID, _ := sharedvo.NewTenantID("acme")

	session := NewPreAuthSession(tenantID, "acme.com", "")

	assert.Equal(t, tenantID.Value(), session.TenantID())
	assert.Equal(t, "acme.com", session.ExpectedEmailDomain())
	assert.NotEmpty(t, session.State())
	assert.NotEmpty(t, session.Nonce())
	assert.NotEmpty(t, session.CodeVerifier())
	assert.False(t, session.IsAuthenticated())
}

func TestAuthSession_UpgradeToAuthenticated(t *testing.T) {
	tenantID, _ := sharedvo.NewTenantID("acme")
	preAuth := NewPreAuthSession(tenantID, "acme.com", "")

	userID := uuid.New()
	userEmail := "user@acme.com"
	accessToken := "access-token-123"
	refreshToken := "refresh-token-456"
	expiry := time.Now().Add(time.Hour)

	authenticated := preAuth.UpgradeToAuthenticated(userID, userEmail, accessToken, refreshToken, expiry)

	assert.True(t, authenticated.IsAuthenticated())
	assert.Equal(t, userID, authenticated.UserID())
	assert.Equal(t, userEmail, authenticated.UserEmail())
	assert.Equal(t, accessToken, authenticated.AccessToken())
	assert.Equal(t, refreshToken, authenticated.RefreshToken())
	assert.Equal(t, expiry.Unix(), authenticated.TokenExpiry().Unix())
	assert.Equal(t, tenantID.Value(), authenticated.TenantID())
	assert.Empty(t, authenticated.State(), "state should be cleared after authentication")
	assert.Empty(t, authenticated.Nonce(), "nonce should be cleared after authentication")
	assert.Empty(t, authenticated.CodeVerifier(), "code verifier should be cleared after authentication")
}

func TestAuthSession_IsTokenExpired(t *testing.T) {
	tenantID, _ := sharedvo.NewTenantID("acme")
	preAuth := NewPreAuthSession(tenantID, "acme.com", "")
	userID := uuid.New()

	t.Run("not expired with buffer", func(t *testing.T) {
		session := preAuth.UpgradeToAuthenticated(userID, "user@acme.com", "access", "refresh", time.Now().Add(time.Hour))
		assert.False(t, session.IsTokenExpired())
	})

	t.Run("expired", func(t *testing.T) {
		session := preAuth.UpgradeToAuthenticated(userID, "user@acme.com", "access", "refresh", time.Now().Add(-time.Hour))
		assert.True(t, session.IsTokenExpired())
	})

	t.Run("expires within buffer window", func(t *testing.T) {
		session := preAuth.UpgradeToAuthenticated(userID, "user@acme.com", "access", "refresh", time.Now().Add(3*time.Minute))
		assert.True(t, session.IsTokenExpired(), "token expiring in 3 minutes should be considered expired due to 5 minute buffer")
	})
}

func TestAuthSession_UpdateTokens(t *testing.T) {
	tenantID, _ := sharedvo.NewTenantID("acme")
	preAuth := NewPreAuthSession(tenantID, "acme.com", "")
	userID := uuid.New()
	session := preAuth.UpgradeToAuthenticated(userID, "user@acme.com", "old-access", "old-refresh", time.Now().Add(-time.Hour))

	newExpiry := time.Now().Add(time.Hour)
	updated := session.UpdateTokens("new-access", "new-refresh", newExpiry)

	assert.Equal(t, "new-access", updated.AccessToken())
	assert.Equal(t, "new-refresh", updated.RefreshToken())
	assert.Equal(t, newExpiry.Unix(), updated.TokenExpiry().Unix())
	assert.Equal(t, userID, updated.UserID())
	assert.Equal(t, "user@acme.com", updated.UserEmail())
}

func TestAuthSession_Serialization(t *testing.T) {
	tenantID, _ := sharedvo.NewTenantID("acme")
	preAuth := NewPreAuthSession(tenantID, "acme.com", "")
	userID := uuid.New()
	expiry := time.Now().Add(time.Hour).Truncate(time.Second)
	session := preAuth.UpgradeToAuthenticated(userID, "user@acme.com", "access", "refresh", expiry)

	data, err := session.Marshal()
	require.NoError(t, err)

	restored, err := UnmarshalAuthSession(data)
	require.NoError(t, err)

	assert.Equal(t, session.TenantID(), restored.TenantID())
	assert.Equal(t, session.UserID(), restored.UserID())
	assert.Equal(t, session.UserEmail(), restored.UserEmail())
	assert.Equal(t, session.AccessToken(), restored.AccessToken())
	assert.Equal(t, session.RefreshToken(), restored.RefreshToken())
	assert.Equal(t, session.TokenExpiry().Unix(), restored.TokenExpiry().Unix())
	assert.Equal(t, session.IsAuthenticated(), restored.IsAuthenticated())
}

func TestNewPreAuthSession_GeneratesUniqueValues(t *testing.T) {
	tenantID, _ := sharedvo.NewTenantID("acme")

	session1 := NewPreAuthSession(tenantID, "acme.com", "")
	session2 := NewPreAuthSession(tenantID, "acme.com", "")

	assert.NotEqual(t, session1.State(), session2.State(), "each session should have unique state")
	assert.NotEqual(t, session1.Nonce(), session2.Nonce(), "each session should have unique nonce")
	assert.NotEqual(t, session1.CodeVerifier(), session2.CodeVerifier(), "each session should have unique code verifier")
}
