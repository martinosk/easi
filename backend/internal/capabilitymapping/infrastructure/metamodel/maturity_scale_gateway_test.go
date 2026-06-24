package metamodel

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	sharedctx "easi/backend/internal/shared/context"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ctxWithTenant(tenantID string) context.Context {
	tid, _ := sharedvo.NewTenantID(tenantID)
	return sharedctx.WithTenant(context.Background(), tid)
}

func standardSectionsResponse(genesisName string) map[string]interface{} {
	return map[string]interface{}{
		"sections": []map[string]interface{}{
			{"order": 1, "name": genesisName, "minValue": 0, "maxValue": 24},
			{"order": 2, "name": "Custom Built", "minValue": 25, "maxValue": 49},
			{"order": 3, "name": "Product", "minValue": 50, "maxValue": 74},
			{"order": 4, "name": "Commodity", "minValue": 75, "maxValue": 99},
		},
	}
}

func newCountingMaturityScaleServer(callCount *int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		*callCount++
		_ = json.NewEncoder(w).Encode(standardSectionsResponse("Genesis"))
	}))
}

func TestMaturityScaleGateway_GetMaturityScaleConfig_FetchesFromAPI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/meta-model/maturity-scale", r.URL.Path)
		assert.Equal(t, "tenant-123", r.Header.Get("X-Tenant-ID"))
		_ = json.NewEncoder(w).Encode(standardSectionsResponse("Genesis"))
	}))
	defer server.Close()

	gateway := NewMaturityScaleGateway(server.URL)
	ctx := ctxWithTenant("tenant-123")

	config, err := gateway.GetMaturityScaleConfig(ctx)

	require.NoError(t, err)
	require.NotNil(t, config)
	assert.Equal(t, 4, len(config.Sections))
	assert.Equal(t, "Genesis", config.Sections[0].Name)
	assert.Equal(t, 0, config.Sections[0].MinValue)
	assert.Equal(t, 24, config.Sections[0].MaxValue)
}

func TestMaturityScaleGateway_GetMaturityScaleConfig_CacheBehavior(t *testing.T) {
	const tenantID = "tenant-123"

	cases := []struct {
		name              string
		ttl               time.Duration
		betweenCalls      func(g MaturityScaleGateway)
		expectedCallCount int
		message           string
	}{
		{
			name:              "caches result within ttl",
			ttl:               5 * time.Minute,
			betweenCalls:      func(MaturityScaleGateway) {},
			expectedCallCount: 1,
			message:           "Should use cached result",
		},
		{
			name:              "refetches after cache expires",
			ttl:               10 * time.Millisecond,
			betweenCalls:      func(MaturityScaleGateway) { time.Sleep(20 * time.Millisecond) },
			expectedCallCount: 2,
			message:           "Should fetch again after cache expires",
		},
		{
			name:              "refetches after cache invalidation",
			ttl:               5 * time.Minute,
			betweenCalls:      func(g MaturityScaleGateway) { g.InvalidateCache(tenantID) },
			expectedCallCount: 2,
			message:           "Should fetch again after cache invalidation",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			callCount := 0
			server := newCountingMaturityScaleServer(&callCount)
			defer server.Close()

			gateway := NewMaturityScaleGatewayWithClient(server.URL, server.Client(), tc.ttl)
			ctx := ctxWithTenant(tenantID)

			_, err := gateway.GetMaturityScaleConfig(ctx)
			require.NoError(t, err)
			assert.Equal(t, 1, callCount)

			tc.betweenCalls(gateway)

			_, err = gateway.GetMaturityScaleConfig(ctx)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedCallCount, callCount, tc.message)
		})
	}
}

func TestMaturityScaleGateway_GetMaturityScaleConfig_ReturnsNilOnNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	gateway := NewMaturityScaleGateway(server.URL)
	ctx := ctxWithTenant("tenant-123")

	config, err := gateway.GetMaturityScaleConfig(ctx)

	require.NoError(t, err)
	assert.Nil(t, config)
}

func TestMaturityScaleGateway_GetMaturityScaleConfig_ReturnsErrorOnServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	gateway := NewMaturityScaleGateway(server.URL)
	ctx := ctxWithTenant("tenant-123")

	_, err := gateway.GetMaturityScaleConfig(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected status code 500")
}

func TestMaturityScaleGateway_GetMaturityScaleConfig_SeparatesCachePerTenant(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		_ = json.NewEncoder(w).Encode(standardSectionsResponse("Genesis for " + tenantID))
	}))
	defer server.Close()

	gateway := NewMaturityScaleGatewayWithClient(server.URL, server.Client(), 5*time.Minute)

	ctx1 := ctxWithTenant("tenant-1")
	config1, err := gateway.GetMaturityScaleConfig(ctx1)
	require.NoError(t, err)
	assert.Equal(t, "Genesis for tenant-1", config1.Sections[0].Name)

	ctx2 := ctxWithTenant("tenant-2")
	config2, err := gateway.GetMaturityScaleConfig(ctx2)
	require.NoError(t, err)
	assert.Equal(t, "Genesis for tenant-2", config2.Sections[0].Name)
}

func TestDefaultMaturityScaleConfig_ReturnsValidDefaults(t *testing.T) {
	config := DefaultMaturityScaleConfig()

	assert.NotNil(t, config)
	assert.Equal(t, 4, len(config.Sections))

	expected := []struct {
		name     string
		minValue int
		maxValue int
		order    int
	}{
		{"Genesis", 0, 24, 1},
		{"Custom Built", 25, 49, 2},
		{"Product", 50, 74, 3},
		{"Commodity", 75, 99, 4},
	}

	for i, exp := range expected {
		assert.Equal(t, exp.name, config.Sections[i].Name)
		assert.Equal(t, exp.minValue, config.Sections[i].MinValue)
		assert.Equal(t, exp.maxValue, config.Sections[i].MaxValue)
		assert.Equal(t, exp.order, config.Sections[i].Order)
	}
}
