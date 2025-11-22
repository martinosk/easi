package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetMaturityLevels_ReturnsAllLevels(t *testing.T) {
	handlers := NewMaturityLevelHandlers()

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
	handlers := NewMaturityLevelHandlers()

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
		value        string
		numericValue int
	}{
		{"Genesis", 1},
		{"Custom Build", 2},
		{"Product", 3},
		{"Commodity", 4},
	}

	for i, expected := range expectedLevels {
		assert.Equal(t, expected.value, response.Data[i].Value, "Value mismatch at index %d", i)
		assert.Equal(t, expected.numericValue, response.Data[i].NumericValue, "NumericValue mismatch at index %d", i)
	}
}

func TestGetMaturityLevels_IncludesLinks(t *testing.T) {
	handlers := NewMaturityLevelHandlers()

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
}

func TestGetMaturityLevels_ReturnsCorrectContentType(t *testing.T) {
	handlers := NewMaturityLevelHandlers()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/capabilities/metadata/maturity-levels", nil)
	w := httptest.NewRecorder()

	handlers.GetMaturityLevels(w, req)

	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

func TestGetMaturityLevels_LevelsInEvolutionOrder(t *testing.T) {
	handlers := NewMaturityLevelHandlers()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/capabilities/metadata/maturity-levels", nil)
	w := httptest.NewRecorder()

	handlers.GetMaturityLevels(w, req)

	var response struct {
		Data []MaturityLevelDTO `json:"data"`
	}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	for i := 1; i < len(response.Data); i++ {
		assert.Greater(t, response.Data[i].NumericValue, response.Data[i-1].NumericValue,
			"Maturity levels should be ordered by numeric value (evolution order)")
	}
}
