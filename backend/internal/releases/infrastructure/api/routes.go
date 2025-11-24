package api

import (
	"database/sql"

	"easi/backend/internal/releases/infrastructure/repository"
	"github.com/go-chi/chi/v5"
)

func SetupReleasesRoutes(r chi.Router, db *sql.DB) error {
	repo := repository.NewPostgresReleaseRepository(db)
	handler := NewReleaseHandler(repo)

	r.Route("/releases", func(r chi.Router) {
		r.Get("/", handler.GetAll)
		r.Get("/latest", handler.GetLatest)
		r.Get("/{version}", handler.GetByVersion)
	})

	return nil
}
