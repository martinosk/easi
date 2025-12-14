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

func testTenantIDValue() sharedvo.TenantID {
	return sharedvo.MustNewTenantID("acme")
}

func withTestTenant(req *http.Request) *http.Request {
	ctx := sharedcontext.WithTenant(req.Context(), testTenantIDValue())
	return req.WithContext(ctx)
}

func testTenantID() string {
	return "acme"
}

func tenantContext() context.Context {
	return sharedcontext.WithTenant(context.Background(), testTenantIDValue())
}
