//go:build integration
// +build integration

package api

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"

	sharedcontext "easi/backend/internal/shared/context"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"

	"github.com/go-chi/chi/v5"
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

func makeRequest(method, url string, body []byte, urlParams map[string]string) (*httptest.ResponseRecorder, *http.Request) {
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	req := httptest.NewRequest(method, url, bodyReader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req = withTestTenant(req)

	if len(urlParams) > 0 {
		rctx := chi.NewRouteContext()
		for key, value := range urlParams {
			rctx.URLParams.Add(key, value)
		}
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	}

	return httptest.NewRecorder(), req
}
