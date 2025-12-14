//go:build integration
// +build integration

package api

import (
	"context"
	"net/http"
	"os"

	sharedcontext "easi/backend/internal/shared/context"
	sharedvo "easi/backend/internal/shared/domain/valueobjects"
)

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func withTestTenant(req *http.Request) *http.Request {
	ctx := sharedcontext.WithTenant(req.Context(), sharedvo.DefaultTenantID())
	return req.WithContext(ctx)
}

func testTenantID() string {
	return "default"
}

func tenantContext() context.Context {
	return sharedcontext.WithTenant(context.Background(), sharedvo.DefaultTenantID())
}
