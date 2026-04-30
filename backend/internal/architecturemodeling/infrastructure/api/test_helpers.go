//go:build integration
// +build integration

package api

import (
	"context"
	"net/http"
	"os"

	sharedcontext "easi/backend/internal/shared/context"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// withTestTenant adds default tenant context to an HTTP request for testing
func withTestTenant(req *http.Request) *http.Request {
	ctx := sharedcontext.WithTenant(req.Context(), sharedvo.DefaultTenantID())
	return req.WithContext(ctx)
}

func withArchitectActor(req *http.Request) *http.Request {
	architect := sharedcontext.NewActor("test-architect", "architect@example.com", sharedcontext.RoleArchitect)
	ctx := sharedcontext.WithActor(req.Context(), architect)
	return req.WithContext(ctx)
}

// testTenantID returns the default tenant ID value for test data insertion
func testTenantID() string {
	return "default"
}

// tenantContext returns a context with default tenant for test operations
func tenantContext() context.Context {
	return sharedcontext.WithTenant(context.Background(), sharedvo.DefaultTenantID())
}
