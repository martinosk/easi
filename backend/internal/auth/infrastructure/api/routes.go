package api

import (
	"database/sql"
	"net/http"
	"os"
	"strings"
	"time"

	"easi/backend/internal/auth/application/handlers"
	"easi/backend/internal/auth/application/projectors"
	"easi/backend/internal/auth/application/readmodels"
	"easi/backend/internal/auth/application/services"
	"easi/backend/internal/auth/domain/valueobjects"
	"easi/backend/internal/auth/infrastructure/repositories"
	"easi/backend/internal/auth/infrastructure/session"
	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/infrastructure/eventstore"
	"easi/backend/internal/shared/config"
	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/shared/events"

	"github.com/alexedwards/scs/postgresstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
)

type AuthDependencies struct {
	SCSManager     *scs.SessionManager
	SessionManager *session.SessionManager
	AuthMiddleware *AuthMiddleware
	AuthHandlers   *AuthHandlers
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
	authHandlers := NewAuthHandlers(deps.SessionManager, tenantOIDCRepo, AuthHandlersConfig{
		ClientSecret:   clientSecret,
		RedirectURL:    redirectURL,
		AllowedOrigins: allowedOrigins,
	})

	deps.AuthHandlers = authHandlers

	userRepo := NewUserRepositoryAdapter(repositories.NewUserRepository(db))
	tenantRepo := NewTenantRepositoryAdapter(repositories.NewTenantRepository(db))
	sessionHandlers := NewSessionHandlers(deps.SessionManager, userRepo, tenantRepo)

	r.Route("/auth", func(r chi.Router) {
		r.Post("/sessions", authHandlers.PostSessions)
		r.Get("/callback", authHandlers.GetCallback)
		r.Get("/sessions/current", sessionHandlers.GetCurrentSession)
		r.Delete("/sessions/current", sessionHandlers.DeleteCurrentSession)
	})

	return nil
}

func WireLoginService(deps *AuthDependencies, invDeps *InvitationDependencies) {
	if deps.AuthHandlers == nil || invDeps == nil {
		return
	}

	loginService := services.NewLoginService(
		invDeps.UserReadModel,
		invDeps.InvitationReadModel,
		invDeps.CommandBus,
	)
	deps.AuthHandlers.WithLoginService(loginService)
}

type InvitationDependencies struct {
	UserReadModel       *readmodels.UserReadModel
	InvitationReadModel *readmodels.InvitationReadModel
	CommandBus          cqrs.CommandBus
}

type InvitationRoutesDeps struct {
	Router     chi.Router
	CommandBus cqrs.CommandBus
	EventStore eventstore.EventStore
	EventBus   events.EventBus
	DB         *database.TenantAwareDB
	AuthDeps   *AuthDependencies
}

func SetupInvitationRoutes(deps InvitationRoutesDeps) (*InvitationDependencies, error) {
	invitationRepo := repositories.NewInvitationRepository(deps.EventStore)
	invitationReadModel := readmodels.NewInvitationReadModel(deps.DB)
	userReadModel := readmodels.NewUserReadModel(deps.DB)

	registerCommandHandlers(deps.CommandBus, invitationRepo, invitationReadModel)
	registerEventSubscriptions(deps.EventBus, invitationReadModel)

	deps.AuthDeps.AuthMiddleware.WithUserReadModel(userReadModel)

	invitationHandlers := NewInvitationHandlers(deps.CommandBus, invitationReadModel)
	registerInvitationRoutes(deps.Router, deps.AuthDeps.AuthMiddleware, invitationHandlers)

	return &InvitationDependencies{
		UserReadModel:       userReadModel,
		InvitationReadModel: invitationReadModel,
		CommandBus:          deps.CommandBus,
	}, nil
}

func registerCommandHandlers(commandBus cqrs.CommandBus, repo *repositories.InvitationRepository, readModel *readmodels.InvitationReadModel) {
	commandBus.Register("CreateInvitation", handlers.NewCreateInvitationHandler(repo))
	commandBus.Register("RevokeInvitation", handlers.NewRevokeInvitationHandler(repo))
	commandBus.Register("AcceptInvitation", handlers.NewAcceptInvitationHandler(repo, readModel))
	commandBus.Register("MarkInvitationExpired", handlers.NewMarkInvitationExpiredHandler(repo))
}

func registerEventSubscriptions(eventBus events.EventBus, readModel *readmodels.InvitationReadModel) {
	projector := projectors.NewInvitationProjector(readModel)
	eventBus.Subscribe("InvitationCreated", projector)
	eventBus.Subscribe("InvitationAccepted", projector)
	eventBus.Subscribe("InvitationRevoked", projector)
	eventBus.Subscribe("InvitationExpired", projector)
}

func registerInvitationRoutes(r chi.Router, authMiddleware *AuthMiddleware, h *InvitationHandlers) {
	r.Route("/invitations", func(r chi.Router) {
		r.Use(authMiddleware.RequirePermission(valueobjects.PermInvitationsManage))
		r.Post("/", h.CreateInvitation)
		r.Get("/", h.GetAllInvitations)
		r.Get("/{id}", h.GetInvitationByID)
		r.Post("/{id}/revoke", h.RevokeInvitation)
	})
}
