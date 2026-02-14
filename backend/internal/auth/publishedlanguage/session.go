package publishedlanguage

import (
	"context"
	"net/http"
)

type SessionProvider interface {
	GetCurrentUserEmail(ctx context.Context) (string, error)
}

type AuthMiddleware interface {
	RequirePermission(permission Permission) func(http.Handler) http.Handler
}
