package api

import (
	"context"
	"net/http"

	"easi/backend/internal/auth/application/readmodels"
	"easi/backend/internal/auth/domain/valueobjects"
	"easi/backend/internal/auth/infrastructure/session"
	"easi/backend/internal/shared/config"
	sharedctx "easi/backend/internal/shared/context"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"

	"github.com/alexedwards/scs/v2"
)

type AuthMiddleware struct {
	sessionManager *session.SessionManager
	userReadModel  *readmodels.UserReadModel
}

func NewAuthMiddleware(sessionManager *session.SessionManager) *AuthMiddleware {
	return &AuthMiddleware{
		sessionManager: sessionManager,
	}
}

func (m *AuthMiddleware) WithUserReadModel(userReadModel *readmodels.UserReadModel) *AuthMiddleware {
	m.userReadModel = userReadModel
	return m
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

func (m *AuthMiddleware) RequirePermission(permission valueobjects.Permission) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if config.IsAuthBypassed() {
				next.ServeHTTP(w, r)
				return
			}

			origCtx := r.Context()
			ctx, err := m.authenticateAndAuthorize(origCtx, permission)
			if err != nil {
				err.Write(w)
				return
			}

			if actor, ok := sharedctx.GetActor(origCtx); ok {
				ctx = sharedctx.WithActor(ctx, actor)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

type authError struct {
	message string
	status  int
}

func (e *authError) Write(w http.ResponseWriter) {
	http.Error(w, e.message, e.status)
}

func (m *AuthMiddleware) authenticateAndAuthorize(ctx context.Context, permission valueobjects.Permission) (context.Context, *authError) {
	authSession, err := m.sessionManager.LoadAuthenticatedSession(ctx)
	if err != nil || !authSession.IsAuthenticated() {
		return nil, &authError{"Unauthorized", http.StatusUnauthorized}
	}

	tenantCtx, err := m.createTenantContext(ctx, authSession.TenantID())
	if err != nil {
		return nil, &authError{"Invalid tenant in session", http.StatusUnauthorized}
	}

	if err := m.checkPermission(tenantCtx, authSession.UserEmail(), permission); err != nil {
		return nil, err
	}

	return tenantCtx, nil
}

func (m *AuthMiddleware) createTenantContext(ctx context.Context, tenantIDStr string) (context.Context, error) {
	tenantID, err := sharedvo.NewTenantID(tenantIDStr)
	if err != nil {
		return nil, err
	}
	return sharedctx.WithTenant(ctx, tenantID), nil
}

func (m *AuthMiddleware) checkPermission(ctx context.Context, email string, permission valueobjects.Permission) *authError {
	user, err := m.userReadModel.GetByEmail(ctx, email)
	if err != nil {
		return &authError{"Failed to load user", http.StatusInternalServerError}
	}
	if user == nil {
		return &authError{"User not found", http.StatusForbidden}
	}

	role, err := valueobjects.RoleFromString(user.Role)
	if err != nil {
		return &authError{"Invalid user role", http.StatusInternalServerError}
	}

	if !role.HasPermission(permission) {
		return &authError{"Forbidden", http.StatusForbidden}
	}

	return nil
}
