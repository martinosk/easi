package api

import (
	"log"
	"net/http"

	"easi/backend/internal/shared/config"
	sharedctx "easi/backend/internal/shared/context"

	"github.com/go-chi/chi/v5"
)

func RequireWriteOrEditGrant(artifactType, idParam string) func(http.Handler) http.Handler {
	return RequireWriteOrEditGrantFor(artifactType, artifactType, idParam)
}

func RequireWriteOrEditGrantFor(permission, artifactType, idParam string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if config.IsAuthBypassed() {
				next.ServeHTTP(w, r)
				return
			}

			actor, ok := sharedctx.GetActor(r.Context())
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			if actor.CanWrite(permission) {
				next.ServeHTTP(w, r)
				return
			}

			artifactID := chi.URLParam(r, idParam)
			if artifactID == "" {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			if actor.HasEditGrant(artifactType, artifactID) {
				log.Printf("[AUDIT] edit-grant-used actor=%s artifact-type=%s artifact-id=%s method=%s path=%s", actor.ID, artifactType, artifactID, r.Method, r.URL.Path)
				next.ServeHTTP(w, r)
				return
			}

			http.Error(w, "Forbidden", http.StatusForbidden)
		})
	}
}
