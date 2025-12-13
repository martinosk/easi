package api

import (
	"database/sql"
	"net/http"
	"os"
	"strings"
	"time"

	"easi/backend/internal/auth/infrastructure/repositories"
	"easi/backend/internal/auth/infrastructure/session"
	"easi/backend/internal/shared/config"

	"github.com/alexedwards/scs/postgresstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
)

type AuthDependencies struct {
	SCSManager     *scs.SessionManager
	SessionManager *session.SessionManager
	AuthMiddleware *AuthMiddleware
}

const (
	SessionLifetime = 7 * 24 * time.Hour // 7 days - matches refresh token lifetime
)

func SetupAuthDependencies(db *sql.DB) (*AuthDependencies, error) {
	scsManager := scs.New()
	scsManager.Store = postgresstore.New(db)
	scsManager.Lifetime = SessionLifetime
	scsManager.Cookie.Name = "easi_session"
	scsManager.Cookie.HttpOnly = true
	scsManager.Cookie.Secure = config.IsProduction()
	scsManager.Cookie.SameSite = http.SameSiteStrictMode

	sessionManager := session.NewSessionManager(scsManager)
	authMiddleware := NewAuthMiddleware(sessionManager)

	return &AuthDependencies{
		SCSManager:     scsManager,
		SessionManager: sessionManager,
		AuthMiddleware: authMiddleware,
	}, nil
}

func SetupAuthRoutes(r chi.Router, db *sql.DB, deps *AuthDependencies) error {
	if config.IsAuthBypassed() {
		return nil
	}

	clientSecret := os.Getenv("OIDC_CLIENT_SECRET")
	if clientSecret == "" {
		panic("OIDC_CLIENT_SECRET environment variable is required")
	}

	redirectURL := os.Getenv("OIDC_REDIRECT_URL")
	if redirectURL == "" {
		redirectURL = "http://localhost:8080/auth/callback"
	}

	allowedOriginsStr := os.Getenv("ALLOWED_ORIGINS")
	var allowedOrigins []string
	if allowedOriginsStr != "" {
		allowedOrigins = strings.Split(allowedOriginsStr, ",")
		for i, origin := range allowedOrigins {
			allowedOrigins[i] = strings.TrimSpace(origin)
		}
	}

	tenantOIDCRepo := repositories.NewTenantOIDCRepository(db)
	handlers := NewAuthHandlers(deps.SessionManager, tenantOIDCRepo, AuthHandlersConfig{
		ClientSecret:   clientSecret,
		RedirectURL:    redirectURL,
		AllowedOrigins: allowedOrigins,
	})

	userRepo := NewUserRepositoryAdapter(repositories.NewUserRepository(db))
	tenantRepo := NewTenantRepositoryAdapter(repositories.NewTenantRepository(db))
	sessionHandlers := NewSessionHandlers(deps.SessionManager, userRepo, tenantRepo)

	r.Route("/auth", func(r chi.Router) {
		r.Post("/sessions", handlers.PostSessions)
		r.Get("/callback", handlers.GetCallback)
		r.Get("/sessions/current", sessionHandlers.GetCurrentSession)
		r.Delete("/sessions/current", sessionHandlers.DeleteCurrentSession)
	})

	return nil
}
