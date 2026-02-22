package api

import (
	"net/http"

	"easi/backend/internal/archassistant/application/orchestrator"
	"easi/backend/internal/archassistant/infrastructure/adapters"
	"easi/backend/internal/archassistant/infrastructure/ratelimit"
	"easi/backend/internal/archassistant/infrastructure/repositories"
	authPL "easi/backend/internal/auth/publishedlanguage"
	"easi/backend/internal/infrastructure/database"

	"github.com/go-chi/chi/v5"
)

type AuthMiddleware interface {
	RequirePermission(permission authPL.Permission) func(http.Handler) http.Handler
}

type ArchAssistantRoutesDeps struct {
	Router          chi.Router
	DB              *database.TenantAwareDB
	AuthMiddleware  AuthMiddleware
	RateLimiter     *ratelimit.Limiter
	LoopbackBaseURL string
}

func SetupArchAssistantRoutes(deps ArchAssistantRoutesDeps) error {
	aiConfigRepo := repositories.NewAIConfigurationRepository(deps.DB)
	configHandlers := NewAIConfigHandlers(aiConfigRepo)

	deps.Router.Route("/assistant-config", func(r chi.Router) {
		r.Use(deps.AuthMiddleware.RequirePermission(authPL.PermMetaModelWrite))
		r.Get("/", configHandlers.GetConfig)
		r.Put("/", configHandlers.UpdateConfig)
		r.Post("/connection-tests", configHandlers.TestConnection)
	})

	convRepo := repositories.NewConversationRepository(deps.DB)
	configProvider := adapters.NewAIConfigProviderAdapter(aiConfigRepo)
	clientFactory := adapters.NewLLMClientFactory()
	orch := orchestrator.New(convRepo, clientFactory)
	convHandlers := NewConversationHandlers(ConversationHandlersDeps{
		ConfigProvider:  configProvider,
		RateLimiter:     deps.RateLimiter,
		Orchestrator:    orch,
		ConvRepo:        convRepo,
		LoopbackBaseURL: deps.LoopbackBaseURL,
	})

	deps.Router.Route("/assistant/conversations", func(r chi.Router) {
		r.Use(deps.AuthMiddleware.RequirePermission(authPL.PermAssistantUse))
		r.Get("/", convHandlers.ListConversations)
		r.Post("/", convHandlers.CreateConversationWithLimit)
		r.Get("/{id}", convHandlers.GetConversation)
		r.Delete("/{id}", convHandlers.DeleteConversation)
		r.Post("/{id}/messages", convHandlers.SendMessage)
	})

	return nil
}
