package api

import (
	"net/http"

	"easi/backend/internal/auth/application/readmodels"
	"easi/backend/internal/auth/infrastructure/session"
	"easi/backend/internal/infrastructure/database"
)

func SessionTenantMiddleware(sessionManager *session.SessionManager, db *database.TenantAwareDB) func(http.Handler) http.Handler {
	return TenantMiddlewareWithSession(sessionManager, readmodels.NewUserReadModel(db))
}
