package api

import (
	"database/sql"
	"net/http"
	"os"
	"time"

	"easi/backend/internal/auth/infrastructure/repositories"
	"easi/backend/internal/auth/infrastructure/session"

	"github.com/alexedwards/scs/postgresstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
)

type AuthDependencies struct {
	SCSManager     *scs.SessionManager
	SessionManager *session.SessionManager
	AuthMiddleware *AuthMiddleware
}

func SetupAuthDependencies(db *sql.DB) (*AuthDependencies, error) {
	scsManager := scs.New()
	scsManager.Store = postgresstore.New(db)
	scsManager.Lifetime = 24 * time.Hour
	scsManager.Cookie.Name = "easi_session"
	scsManager.Cookie.HttpOnly = true
	scsManager.Cookie.Secure = os.Getenv("LOCAL_DEV_MODE") != "true"
	scsManager.Cookie.SameSite = http.SameSiteLaxMode

	sessionManager := session.NewSessionManager(scsManager)
	authMiddleware := NewAuthMiddleware(sessionManager)

	return &AuthDependencies{
		SCSManager:     scsManager,
		SessionManager: sessionManager,
		AuthMiddleware: authMiddleware,
	}, nil
}

func SetupAuthRoutes(r chi.Router, db *sql.DB, deps *AuthDependencies) error {
	clientSecret := os.Getenv("OIDC_CLIENT_SECRET")
	if clientSecret == "" && os.Getenv("LOCAL_DEV_MODE") == "true" {
		clientSecret = "easi-test-secret"
	}

	redirectURL := os.Getenv("OIDC_REDIRECT_URL")
	if redirectURL == "" {
		redirectURL = "http://localhost:8080/auth/callback"
	}

	tenantRepo := repositories.NewTenantOIDCRepository(db)
	handlers := NewAuthHandlers(deps.SessionManager, tenantRepo, clientSecret, redirectURL)

	r.Route("/auth", func(r chi.Router) {
		r.Post("/sessions", handlers.PostSessions)
		r.Get("/callback", handlers.GetCallback)
	})

	return nil
}
