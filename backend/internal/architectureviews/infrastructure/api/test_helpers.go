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

const testActorID = "test-user-id"
const testActorEmail = "test@example.com"

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// withTestTenant adds default tenant and actor context to an HTTP request for testing
func withTestTenant(req *http.Request) *http.Request {
	ctx := sharedcontext.WithTenant(req.Context(), sharedvo.DefaultTenantID())
	ctx = sharedcontext.WithActor(ctx, sharedcontext.Actor{
		ID:    testActorID,
		Email: testActorEmail,
	})
	return req.WithContext(ctx)
}

// testTenantID returns the default tenant ID value for test data insertion
func testTenantID() string {
	return "default"
}

// tenantContext returns a context with default tenant and actor for test operations
func tenantContext() context.Context {
	ctx := sharedcontext.WithTenant(context.Background(), sharedvo.DefaultTenantID())
	ctx = sharedcontext.WithActor(ctx, sharedcontext.Actor{
		ID:    testActorID,
		Email: testActorEmail,
	})
	return ctx
}
