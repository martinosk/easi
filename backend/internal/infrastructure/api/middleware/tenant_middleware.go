package middleware

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"

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
			var ctx context.Context
			var err error

			if config.IsAuthBypassed() {
				ctx, err = buildBypassContext(r)
			} else {
				ctx, err = buildSessionContext(r, sessionManager, userRoleLookup)
			}

			if err != nil {
				handleTenantError(w, err)
				return
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func buildBypassContext(r *http.Request) (context.Context, error) {
	tenantID, err := extractTenantFromHeader(r)
	if err != nil {
		return nil, err
	}
	ctx := sharedctx.WithTenant(r.Context(), tenantID)
	logTenantContext(r, tenantID)
	return ctx, nil
}

func buildSessionContext(r *http.Request, sessionManager *session.SessionManager, userRoleLookup UserRoleLookup) (context.Context, error) {
	info, err := extractSessionInfo(r, sessionManager)
	if err != nil {
		return nil, err
	}
	ctx := sharedctx.WithTenant(r.Context(), info.tenantID)

	role := resolveUserRole(ctx, info.userEmail, userRoleLookup)
	actor := sharedctx.NewActor(info.userID, info.userEmail, role)
	ctx = sharedctx.WithActor(ctx, actor)
	logTenantContext(r, info.tenantID)
	return ctx, nil
}

func resolveUserRole(ctx context.Context, email string, lookup UserRoleLookup) sharedctx.Role {
	if lookup == nil {
		return ""
	}
	roleStr, _ := lookup.GetRoleByEmail(ctx, email)
	return sharedctx.Role(roleStr)
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
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			ip := strings.TrimSpace(parts[0])
			if ip != "" {
				return ip
			}
		}
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
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
