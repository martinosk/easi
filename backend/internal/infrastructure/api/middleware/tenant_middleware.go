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

// TenantMiddleware extracts tenant context from request and injects it into context
// Supports two modes based on AUTH_MODE:
// 1. Bypass Mode (AUTH_MODE=bypass): Uses X-Tenant-ID header
// 2. Authenticated Mode (AUTH_MODE=production or local_oidc): Extracts from authenticated session
func TenantMiddleware() func(http.Handler) http.Handler {
	return TenantMiddlewareWithSession(nil)
}

// TenantMiddlewareWithSession creates tenant middleware with session support for production mode
// Also injects actor context for audit trail when authenticated
func TenantMiddlewareWithSession(sessionManager *session.SessionManager) func(http.Handler) http.Handler {
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
				ctx = sharedctx.WithActor(ctx, sharedctx.Actor{
					ID:    info.userID,
					Email: info.userEmail,
				})
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

// logTenantContext logs tenant context operations for audit trail
func logTenantContext(r *http.Request, tenantID sharedvo.TenantID) {
	log.Printf("[TENANT_CONTEXT] Method=%s Path=%s Tenant=%s IP=%s UserAgent=%s",
		r.Method,
		r.URL.Path,
		tenantID.Value(),
		getClientIP(r),
		r.UserAgent(),
	)
}

// getClientIP extracts the client IP address from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (for proxied requests)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}

// RequireTenant is a middleware that ensures tenant context exists
// Use this for routes that absolutely require tenant context
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

// ExtractTenantID is a helper function to extract tenant ID from request context
// Returns error if tenant context is missing
func ExtractTenantID(ctx context.Context) (string, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return "", fmt.Errorf("tenant context not found: %w", err)
	}
	return tenantID.Value(), nil
}
