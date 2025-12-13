package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"easi/backend/internal/auth/infrastructure/session"
	"easi/backend/internal/shared/config"
	sharedctx "easi/backend/internal/shared/context"
	sharedvo "easi/backend/internal/shared/domain/valueobjects"
)

// TenantMiddleware extracts tenant context from request and injects it into context
// Supports two modes based on AUTH_MODE:
// 1. Bypass Mode (AUTH_MODE=bypass): Uses X-Tenant-ID header
// 2. Authenticated Mode (AUTH_MODE=production or local_oidc): Extracts from authenticated session
func TenantMiddleware() func(http.Handler) http.Handler {
	return TenantMiddlewareWithSession(nil)
}

// TenantMiddlewareWithSession creates tenant middleware with session support for production mode
func TenantMiddlewareWithSession(sessionManager *session.SessionManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var tenantID sharedvo.TenantID
			var err error

			if config.IsAuthBypassed() {
				// Bypass Mode: Extract from X-Tenant-ID header
				tenantIDStr := r.Header.Get("X-Tenant-ID")

				if tenantIDStr == "" {
					log.Println("No X-Tenant-ID header found, using default tenant")
					tenantID = sharedvo.DefaultTenantID()
				} else {
					tenantID, err = sharedvo.NewTenantID(tenantIDStr)
					if err != nil {
						log.Printf("Invalid tenant ID in header: %v", err)
						http.Error(w, "Invalid tenant ID", http.StatusBadRequest)
						return
					}
				}

				log.Printf("AUTH_MODE=bypass: Using tenant '%s' from X-Tenant-ID header", tenantID.Value())
			} else {
				// Authenticated Mode: Extract from authenticated session
				if sessionManager == nil {
					http.Error(w, "Authentication required", http.StatusUnauthorized)
					return
				}

				authSession, err := sessionManager.LoadAuthenticatedSession(r.Context())
				if err != nil {
					http.Error(w, "Authentication required", http.StatusUnauthorized)
					return
				}

				tenantID, err = sharedvo.NewTenantID(authSession.TenantID())
				if err != nil {
					log.Printf("Invalid tenant ID in session: %v", err)
					http.Error(w, "Invalid session", http.StatusUnauthorized)
					return
				}

				log.Printf("AUTH_MODE=%s: Using tenant '%s' from session", config.GetAuthMode(), tenantID.Value())
			}

			ctx := sharedctx.WithTenant(r.Context(), tenantID)
			logTenantContext(r, tenantID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
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
