package api

import (
	"net/http"

	"easi/backend/internal/auth/infrastructure/session"
	"easi/backend/internal/shared/config"
	sharedctx "easi/backend/internal/shared/context"
	sharedvo "easi/backend/internal/shared/domain/valueobjects"

	"github.com/alexedwards/scs/v2"
)

type AuthMiddleware struct {
	sessionManager *session.SessionManager
}

func NewAuthMiddleware(sessionManager *session.SessionManager) *AuthMiddleware {
	return &AuthMiddleware{
		sessionManager: sessionManager,
	}
}

func (m *AuthMiddleware) RequireAuth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if config.IsAuthBypassed() {
				next.ServeHTTP(w, r)
				return
			}

			authSession, err := m.sessionManager.LoadAuthenticatedSession(r.Context())
			if err != nil || !authSession.IsAuthenticated() {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			tenantID, err := sharedvo.NewTenantID(authSession.TenantID())
			if err != nil {
				http.Error(w, "Invalid tenant in session", http.StatusUnauthorized)
				return
			}

			ctx := sharedctx.WithTenant(r.Context(), tenantID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func SessionLoadMiddleware(scsManager *scs.SessionManager) func(http.Handler) http.Handler {
	return scsManager.LoadAndSave
}
