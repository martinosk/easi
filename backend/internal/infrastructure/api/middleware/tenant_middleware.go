package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"easi/backend/internal/auth/infrastructure/session"
	"easi/backend/internal/shared/config"
	sharedctx "easi/backend/internal/shared/context"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

type UserRoleLookup interface {
	GetRoleByEmail(ctx context.Context, email string) (string, error)
}

func TenantMiddleware() func(http.Handler) http.Handler {
	return TenantMiddlewareWithSession(nil, nil)
}

func TenantMiddlewareWithSession(sessionManager *session.SessionManager, userRoleLookup UserRoleLookup) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			if config.IsAuthBypassed() {
				tenantID, err := extractTenantFromHeader(r)
				if err != nil {
					handleTenantError(w, err)
					return
				}
				ctx = sharedctx.WithTenant(ctx, tenantID)
				logTenantContext(r, tenantID)
			} else {
				info, err := extractSessionInfo(r, sessionManager)
				if err != nil {
					handleTenantError(w, err)
					return
				}
				ctx = sharedctx.WithTenant(ctx, info.tenantID)

				role := ""
				if userRoleLookup != nil {
					role, _ = userRoleLookup.GetRoleByEmail(ctx, info.userEmail)
				}

				actor := sharedctx.NewActor(info.userID, info.userEmail, role)
				ctx = sharedctx.WithActor(ctx, actor)
				logTenantContext(r, info.tenantID)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

type tenantError struct {
	message    string
	statusCode int
}

func (e tenantError) Error() string { return e.message }

func extractTenantFromHeader(r *http.Request) (sharedvo.TenantID, error) {
	tenantIDStr := r.Header.Get("X-Tenant-ID")
	if tenantIDStr == "" {
		log.Println("No X-Tenant-ID header found, using default tenant")
		return sharedvo.DefaultTenantID(), nil
	}

	tenantID, err := sharedvo.NewTenantID(tenantIDStr)
	if err != nil {
		log.Printf("Invalid tenant ID in header: %v", err)
		return sharedvo.TenantID{}, tenantError{"Invalid tenant ID", http.StatusBadRequest}
	}

	log.Printf("AUTH_MODE=bypass: Using tenant '%s' from X-Tenant-ID header", tenantID.Value())
	return tenantID, nil
}

type sessionInfo struct {
	tenantID  sharedvo.TenantID
	userID    string
	userEmail string
}

func extractSessionInfo(r *http.Request, sessionManager *session.SessionManager) (sessionInfo, error) {
	if sessionManager == nil {
		return sessionInfo{}, tenantError{"Authentication required", http.StatusUnauthorized}
	}

	authSession, err := sessionManager.LoadAuthenticatedSession(r.Context())
	if err != nil {
		return sessionInfo{}, tenantError{"Authentication required", http.StatusUnauthorized}
	}

	tenantID, err := sharedvo.NewTenantID(authSession.TenantID())
	if err != nil {
		log.Printf("Invalid tenant ID in session: %v", err)
		return sessionInfo{}, tenantError{"Invalid session", http.StatusUnauthorized}
	}

	log.Printf("AUTH_MODE=%s: Using tenant '%s' from session", config.GetAuthMode(), tenantID.Value())
	return sessionInfo{
		tenantID:  tenantID,
		userID:    authSession.UserID().String(),
		userEmail: authSession.UserEmail(),
	}, nil
}

func handleTenantError(w http.ResponseWriter, err error) {
	if te, ok := err.(tenantError); ok {
		http.Error(w, te.message, te.statusCode)
		return
	}
	http.Error(w, "Internal server error", http.StatusInternalServerError)
}

func logTenantContext(r *http.Request, tenantID sharedvo.TenantID) {
	log.Printf("[TENANT_CONTEXT] Method=%s Path=%s Tenant=%s IP=%s UserAgent=%s",
		r.Method,
		r.URL.Path,
		tenantID.Value(),
		getClientIP(r),
		r.UserAgent(),
	)
}

func getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}

	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	return r.RemoteAddr
}

func RequireTenant() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := sharedctx.GetTenant(r.Context())
			if err != nil {
				log.Printf("Request missing tenant context: %v", err)
				http.Error(w, "Tenant context required", http.StatusBadRequest)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func ExtractTenantID(ctx context.Context) (string, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return "", fmt.Errorf("tenant context not found: %w", err)
	}
	return tenantID.Value(), nil
}
