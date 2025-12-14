package api

import (
	"context"
	"net/http"
	"sync"

	"easi/backend/internal/auth/infrastructure/oidc"
	"easi/backend/internal/auth/infrastructure/repositories"
	"easi/backend/internal/auth/infrastructure/session"
	"easi/backend/internal/shared/config"
)

type TokenRefreshMiddleware struct {
	sessionManager *session.SessionManager
	tenantRepo     TenantOIDCRepository
	clientSecret   string
	redirectURL    string
	mu             sync.Mutex
}

func NewTokenRefreshMiddleware(
	sessionManager *session.SessionManager,
	tenantRepo TenantOIDCRepository,
	clientSecret string,
	redirectURL string,
) *TokenRefreshMiddleware {
	return &TokenRefreshMiddleware{
		sessionManager: sessionManager,
		tenantRepo:     tenantRepo,
		clientSecret:   clientSecret,
		redirectURL:    redirectURL,
	}
}

func (m *TokenRefreshMiddleware) RefreshIfNeeded() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if config.IsAuthBypassed() {
				next.ServeHTTP(w, r)
				return
			}

			authSession, err := m.sessionManager.LoadAuthenticatedSession(r.Context())
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			if !authSession.IsTokenExpired() {
				next.ServeHTTP(w, r)
				return
			}

			if authSession.RefreshToken() == "" {
				_ = m.sessionManager.ClearSession(r.Context())
				http.Error(w, "Session expired", http.StatusUnauthorized)
				return
			}

			newSession, err := m.refreshTokens(r.Context(), authSession)
			if err != nil {
				_ = m.sessionManager.ClearSession(r.Context())
				http.Error(w, "Session expired", http.StatusUnauthorized)
				return
			}

			if err := m.sessionManager.StoreAuthenticatedSession(r.Context(), newSession); err != nil {
				http.Error(w, "Failed to update session", http.StatusInternalServerError)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (m *TokenRefreshMiddleware) refreshTokens(ctx context.Context, authSession session.AuthSession) (session.AuthSession, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	tenantConfig, err := m.lookupTenantConfig(ctx, authSession.TenantID())
	if err != nil {
		return session.AuthSession{}, err
	}

	provider, err := oidc.NewOIDCProviderFromConfig(ctx, oidc.ProviderConfig{
		DiscoveryURL: tenantConfig.DiscoveryURL,
		ClientID:     tenantConfig.ClientID,
		ClientSecret: m.clientSecret,
		RedirectURL:  m.redirectURL,
	})
	if err != nil {
		return session.AuthSession{}, err
	}

	result, err := provider.RefreshToken(ctx, authSession.RefreshToken())
	if err != nil {
		return session.AuthSession{}, err
	}

	return authSession.UpdateTokens(session.TokenInfo{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		Expiry:       result.TokenExpiry,
	}), nil
}

func (m *TokenRefreshMiddleware) lookupTenantConfig(ctx context.Context, tenantID string) (*repositories.TenantOIDCConfig, error) {
	return m.tenantRepo.GetByTenantID(ctx, tenantID)
}
