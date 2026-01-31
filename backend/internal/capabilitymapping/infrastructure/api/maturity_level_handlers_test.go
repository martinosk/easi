package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"easi/backend/internal/capabilitymapping/infrastructure/metamodel"
	"easi/backend/internal/shared/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockMaturityScaleGateway struct {
	config *metamodel.MaturityScaleConfigDTO
	err    error
}

func (m *mockMaturityScaleGateway) GetMaturityScaleConfig(ctx context.Context) (*metamodel.MaturityScaleConfigDTO, error) {
	return m.config, m.err
}

func (m *mockMaturityScaleGateway) InvalidateCache(tenantID string) {}

func TestGetMaturityLevels_ReturnsAllLevels(t *testing.T) {
	handlers := NewMaturityLevelHandlers(nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/capabilities/metadata/maturity-levels", nil)
	w := httptest.NewRecorder()

	handlers.GetMaturityLevels(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Data  []MaturityLevelDTO `json:"data"`
		Links types.Links  `json:"_links"`
	}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, 4, len(response.Data))
}

func TestGetMaturityLevels_HasCorrectValues(t *testing.T) {
	handlers := NewMaturityLevelHandlers(nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/capabilities/metadata/maturity-levels", nil)
	w := httptest.NewRecorder()

	handlers.GetMaturityLevels(w, req)

	var response struct {
		Data  []MaturityLevelDTO `json:"data"`
		Links types.Links  `json:"_links"`
	}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	expectedLevels := []struct {
		value    string
		minValue int
		maxValue int
		order    int
	}{
		{"Genesis", 0, 24, 1},
		{"Custom Built", 25, 49, 2},
		{"Product", 50, 74, 3},
		{"Commodity", 75, 99, 4},
	}

	for i, expected := range expectedLevels {
		assert.Equal(t, expected.value, response.Data[i].Value, "Value mismatch at index %d", i)
		assert.Equal(t, expected.minValue, response.Data[i].MinValue, "MinValue mismatch at index %d", i)
		assert.Equal(t, expected.maxValue, response.Data[i].MaxValue, "MaxValue mismatch at index %d", i)
		assert.Equal(t, expected.order, response.Data[i].Order, "Order mismatch at index %d", i)
	}
}

func TestGetMaturityLevels_IncludesLinks(t *testing.T) {
	handlers := NewMaturityLevelHandlers(nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/capabilities/metadata/maturity-levels", nil)
	w := httptest.NewRecorder()

	handlers.GetMaturityLevels(w, req)

	var response struct {
		Data  []MaturityLevelDTO `json:"data"`
		Links types.Links  `json:"_links"`
	}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.NotNil(t, response.Links)
	assert.Equal(t, "/api/v1/capabilities/metadata/maturity-levels", response.Links["self"].Href)
	assert.Equal(t, "/api/v1/meta-model/maturity-scale", response.Links["x-configure-at"].Href)
}

func TestGetMaturityLevels_ReturnsCorrectContentType(t *testing.T) {
	handlers := NewMaturityLevelHandlers(nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/capabilities/metadata/maturity-levels", nil)
	w := httptest.NewRecorder()

	handlers.GetMaturityLevels(w, req)

	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

func TestGetMaturityLevels_LevelsInEvolutionOrder(t *testing.T) {
	handlers := NewMaturityLevelHandlers(nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/capabilities/metadata/maturity-levels", nil)
	w := httptest.NewRecorder()

	handlers.GetMaturityLevels(w, req)

	var response struct {
		Data []MaturityLevelDTO `json:"data"`
	}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	for i := 1; i < len(response.Data); i++ {
		assert.Greater(t, response.Data[i].Order, response.Data[i-1].Order,
			"Maturity levels should be ordered by evolution order")
	}
}

func TestGetMaturityLevels_UsesGatewayConfig(t *testing.T) {
	customConfig := &metamodel.MaturityScaleConfigDTO{
		Sections: []metamodel.MaturitySectionDTO{
			{Order: 1, Name: "Early Stage", MinValue: 0, MaxValue: 30},
			{Order: 2, Name: "Growth", MinValue: 31, MaxValue: 60},
			{Order: 3, Name: "Mature", MinValue: 61, MaxValue: 80},
			{Order: 4, Name: "Legacy", MinValue: 81, MaxValue: 99},
		},
	}

	gateway := &mockMaturityScaleGateway{config: customConfig}
	handlers := NewMaturityLevelHandlers(gateway)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/capabilities/metadata/maturity-levels", nil)
	w := httptest.NewRecorder()

	handlers.GetMaturityLevels(w, req)

	var response struct {
		Data []MaturityLevelDTO `json:"data"`
	}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, 4, len(response.Data))
	assert.Equal(t, "Early Stage", response.Data[0].Value)
	assert.Equal(t, 0, response.Data[0].MinValue)
	assert.Equal(t, 30, response.Data[0].MaxValue)
	assert.Equal(t, "Legacy", response.Data[3].Value)
}

func TestGetMaturityLevels_FallsBackToDefaults(t *testing.T) {
	tests := []struct {
		name    string
		gateway *mockMaturityScaleGateway
	}{
		{"on gateway error", &mockMaturityScaleGateway{err: assert.AnError}},
		{"on nil config", &mockMaturityScaleGateway{config: nil}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlers := NewMaturityLevelHandlers(tt.gateway)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/capabilities/metadata/maturity-levels", nil)
			w := httptest.NewRecorder()

			handlers.GetMaturityLevels(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response struct {
				Data []MaturityLevelDTO `json:"data"`
			}
			err := json.NewDecoder(w.Body).Decode(&response)
			require.NoError(t, err)

			assert.Equal(t, 4, len(response.Data))
			assert.Equal(t, "Genesis", response.Data[0].Value)
		})
	}
}

func executeMetadataRequest(t *testing.T, handlerFn http.HandlerFunc, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	w := httptest.NewRecorder()
	handlerFn(w, req)
	return w
}

type metadataCollectionItem struct {
	Value       string `json:"value"`
	DisplayName string `json:"displayName"`
}

type metadataCollectionResponse struct {
	Data  []metadataCollectionItem `json:"data"`
	Links types.Links              `json:"_links"`
}

func assertMetadataCollection(t *testing.T, body []byte, expectedSelfLink string, expectedValues []metadataCollectionItem) {
	t.Helper()

	var response metadataCollectionResponse
	require.NoError(t, json.Unmarshal(body, &response))

	assert.Equal(t, len(expectedValues), len(response.Data))
	for i, expected := range expectedValues {
		assert.Equal(t, expected.Value, response.Data[i].Value, "Value mismatch at index %d", i)
		assert.Equal(t, expected.DisplayName, response.Data[i].DisplayName, "DisplayName mismatch at index %d", i)
	}
	assert.Equal(t, expectedSelfLink, response.Links["self"].Href)
}

func TestGetStatuses_ReturnsAllStatuses(t *testing.T) {
	handlers := NewMaturityLevelHandlers(nil)
	w := executeMetadataRequest(t, handlers.GetStatuses, "/api/v1/capabilities/metadata/statuses")
	assert.Equal(t, http.StatusOK, w.Code)

	body := w.Body.Bytes()
	assertMetadataCollection(t, body, "/api/v1/capabilities/metadata/statuses", []metadataCollectionItem{
		{"Active", "Active"},
		{"Planned", "Planned"},
		{"Deprecated", "Deprecated"},
	})

	var response struct {
		Data []StatusDTO `json:"data"`
	}
	require.NoError(t, json.Unmarshal(body, &response))
	for i, expected := range []int{1, 2, 3} {
		assert.Equal(t, expected, response.Data[i].SortOrder, "SortOrder mismatch at index %d", i)
	}
}

func TestGetStatuses_IncludesCacheHeaders(t *testing.T) {
	handlers := NewMaturityLevelHandlers(nil)
	w := executeMetadataRequest(t, handlers.GetStatuses, "/api/v1/capabilities/metadata/statuses")

	assert.Contains(t, w.Header().Get("Cache-Control"), "public")
	assert.NotEmpty(t, w.Header().Get("ETag"))
}

func TestGetOwnershipModels_ReturnsAllModels(t *testing.T) {
	handlers := NewMaturityLevelHandlers(nil)
	w := executeMetadataRequest(t, handlers.GetOwnershipModels, "/api/v1/capabilities/metadata/ownership-models")
	assert.Equal(t, http.StatusOK, w.Code)

	assertMetadataCollection(t, w.Body.Bytes(), "/api/v1/capabilities/metadata/ownership-models", []metadataCollectionItem{
		{"TribeOwned", "Tribe Owned"},
		{"TeamOwned", "Team Owned"},
		{"Shared", "Shared"},
		{"EnterpriseService", "Enterprise Service"},
	})
}

func TestGetMetadataIndex_ReturnsAllLinks(t *testing.T) {
	handlers := NewMaturityLevelHandlers(nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/capabilities/metadata", nil)
	w := httptest.NewRecorder()

	handlers.GetMetadataIndex(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response MetadataIndexDTO
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "/api/v1/capabilities/metadata", response.Links["self"].Href)
	assert.Equal(t, "/api/v1/capabilities/metadata/maturity-levels", response.Links["x-maturity-levels"].Href)
	assert.Equal(t, "/api/v1/capabilities/metadata/statuses", response.Links["x-statuses"].Href)
	assert.Equal(t, "/api/v1/capabilities/metadata/ownership-models", response.Links["x-ownership-models"].Href)
}
