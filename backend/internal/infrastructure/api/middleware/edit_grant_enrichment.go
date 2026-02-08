package middleware

import (
	"context"
	"log"
	"net/http"

	"easi/backend/internal/shared/config"
	sharedctx "easi/backend/internal/shared/context"
)

type EditGrantResolver interface {
	ResolveEditGrants(ctx context.Context, email string) (map[string]map[string]bool, error)
}

func hasFullWriteAccess(actor sharedctx.Actor) bool {
	return actor.CanWrite("components") && actor.CanWrite("views") && actor.CanWrite("capabilities")
}

func EditGrantEnrichment(resolver EditGrantResolver) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if config.IsAuthBypassed() {
				next.ServeHTTP(w, r)
				return
			}

			actor, ok := sharedctx.GetActor(r.Context())
			if !ok {
				next.ServeHTTP(w, r)
				return
			}

			if hasFullWriteAccess(actor) {
				next.ServeHTTP(w, r)
				return
			}

			grants, err := resolver.ResolveEditGrants(r.Context(), actor.Email)
			if err != nil {
				log.Printf("Failed to resolve edit grants for %s: %v", actor.ID, err)
				next.ServeHTTP(w, r)
				return
			}

			if len(grants) > 0 {
				enriched := actor.WithEditGrants(grants)
				ctx := sharedctx.WithActor(r.Context(), enriched)
				r = r.WithContext(ctx)
			}

			next.ServeHTTP(w, r)
		})
	}
}
