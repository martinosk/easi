package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"easi/backend/internal/capabilitymapping/infrastructure/metamodel"

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
		Links map[string]string  `json:"_links"`
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
		Links map[string]string  `json:"_links"`
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
		Links map[string]string  `json:"_links"`
	}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.NotNil(t, response.Links)
	assert.Equal(t, "/api/v1/capabilities/metadata/maturity-levels", response.Links["self"])
	assert.Equal(t, "/api/v1/meta-model/maturity-scale", response.Links["configureAt"])
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

func TestGetMaturityLevels_FallsBackToDefaultsOnError(t *testing.T) {
	gateway := &mockMaturityScaleGateway{err: assert.AnError}
	handlers := NewMaturityLevelHandlers(gateway)

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
}

func TestGetMaturityLevels_FallsBackToDefaultsOnNilConfig(t *testing.T) {
	gateway := &mockMaturityScaleGateway{config: nil}
	handlers := NewMaturityLevelHandlers(gateway)

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
}

func TestGetStatuses_ReturnsAllStatuses(t *testing.T) {
	handlers := NewMaturityLevelHandlers(nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/capabilities/metadata/statuses", nil)
	w := httptest.NewRecorder()

	handlers.GetStatuses(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Data  []StatusDTO       `json:"data"`
		Links map[string]string `json:"_links"`
	}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, 3, len(response.Data))

	expectedStatuses := []struct {
		value       string
		displayName string
		sortOrder   int
	}{
		{"Active", "Active", 1},
		{"Planned", "Planned", 2},
		{"Deprecated", "Deprecated", 3},
	}

	for i, expected := range expectedStatuses {
		assert.Equal(t, expected.value, response.Data[i].Value)
		assert.Equal(t, expected.displayName, response.Data[i].DisplayName)
		assert.Equal(t, expected.sortOrder, response.Data[i].SortOrder)
	}

	assert.Equal(t, "/api/v1/capabilities/metadata/statuses", response.Links["self"])
}

func TestGetStatuses_IncludesCacheHeaders(t *testing.T) {
	handlers := NewMaturityLevelHandlers(nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/capabilities/metadata/statuses", nil)
	w := httptest.NewRecorder()

	handlers.GetStatuses(w, req)

	assert.Contains(t, w.Header().Get("Cache-Control"), "public")
	assert.NotEmpty(t, w.Header().Get("ETag"))
}

func TestGetOwnershipModels_ReturnsAllModels(t *testing.T) {
	handlers := NewMaturityLevelHandlers(nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/capabilities/metadata/ownership-models", nil)
	w := httptest.NewRecorder()

	handlers.GetOwnershipModels(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Data  []OwnershipModelDTO `json:"data"`
		Links map[string]string   `json:"_links"`
	}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, 4, len(response.Data))

	expectedModels := []struct {
		value       string
		displayName string
	}{
		{"TribeOwned", "Tribe Owned"},
		{"TeamOwned", "Team Owned"},
		{"Shared", "Shared"},
		{"EnterpriseService", "Enterprise Service"},
	}

	for i, expected := range expectedModels {
		assert.Equal(t, expected.value, response.Data[i].Value)
		assert.Equal(t, expected.displayName, response.Data[i].DisplayName)
	}

	assert.Equal(t, "/api/v1/capabilities/metadata/ownership-models", response.Links["self"])
}

func TestGetStrategyPillars_ReturnsAllPillars(t *testing.T) {
	handlers := NewMaturityLevelHandlers(nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/capabilities/metadata/strategy-pillars", nil)
	w := httptest.NewRecorder()

	handlers.GetStrategyPillars(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Data  []StrategyPillarDTO `json:"data"`
		Links map[string]string   `json:"_links"`
	}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, 3, len(response.Data))

	expectedPillars := []struct {
		value       string
		displayName string
	}{
		{"AlwaysOn", "Always On"},
		{"Grow", "Grow"},
		{"Transform", "Transform"},
	}

	for i, expected := range expectedPillars {
		assert.Equal(t, expected.value, response.Data[i].Value)
		assert.Equal(t, expected.displayName, response.Data[i].DisplayName)
	}

	assert.Equal(t, "/api/v1/capabilities/metadata/strategy-pillars", response.Links["self"])
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

	assert.Equal(t, "/api/v1/capabilities/metadata", response.Links["self"])
	assert.Equal(t, "/api/v1/capabilities/metadata/maturity-levels", response.Links["maturityLevels"])
	assert.Equal(t, "/api/v1/capabilities/metadata/statuses", response.Links["statuses"])
	assert.Equal(t, "/api/v1/capabilities/metadata/ownership-models", response.Links["ownershipModels"])
	assert.Equal(t, "/api/v1/capabilities/metadata/strategy-pillars", response.Links["strategyPillars"])
}
