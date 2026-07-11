//go:build integration
// +build integration

package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/infrastructure/database"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/cqrs"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupCompletenessCapabilityHandlers(db *sql.DB, source OnePagerCompletenessSource) *CapabilityHandlers {
	tenantDB := database.NewTenantAwareDB(db)
	commandBus := cqrs.NewInMemoryCommandBus()
	links := NewCapabilityMappingLinks(sharedAPI.NewHATEOASLinks("/api/v1"))
	readModel := readmodels.NewCapabilityReadModel(tenantDB)
	return NewCapabilityHandlers(CapabilityHandlersDeps{CommandBus: commandBus, ReadModel: readModel, Links: links, Completeness: source})
}

func getAllCapabilities(t *testing.T, handlers *CapabilityHandlers) ([]map[string]any, string) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/capabilities", nil)
	req = withTestTenant(req)
	w := httptest.NewRecorder()
	handlers.GetAllCapabilities(w, req)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())

	var response struct {
		Data []map[string]any `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
	return response.Data, w.Body.String()
}

func TestGetAllCapabilities_DecoratesRowsWhenIndicatorPresent_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	idA := uuid.New().String()
	idB := uuid.New().String()
	testCtx.createTestCapability(t, capabilitySpec{ID: idA, Name: "OP Complete Alpha", Level: "L1"})
	testCtx.createTestCapability(t, capabilitySpec{ID: idB, Name: "OP Complete Beta", Level: "L1"})

	source := &fakeOnePagerCompletenessSource{indicators: map[string]bool{idA: true}, present: true}
	handlers := setupCompletenessCapabilityHandlers(testCtx.db, source)

	rows, _ := getAllCapabilities(t, handlers)

	byID := map[string]map[string]any{}
	for _, row := range rows {
		byID[row["id"].(string)] = row
	}
	require.Contains(t, byID, idA)
	require.Contains(t, byID, idB)
	assert.Equal(t, true, byID[idA]["onePagerComplete"])
	assert.Equal(t, false, byID[idB]["onePagerComplete"])
	assert.Contains(t, source.gotIDs, idA)
	assert.Contains(t, source.gotIDs, idB)
}

func TestGetAllCapabilities_NoIndicatorWhenSourceNilOrAbsent_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	idA := uuid.New().String()
	testCtx.createTestCapability(t, capabilitySpec{ID: idA, Name: "OP Complete Gamma", Level: "L1"})

	cases := []struct {
		name   string
		source OnePagerCompletenessSource
	}{
		{"nil source", nil},
		{"indicator absent", &fakeOnePagerCompletenessSource{present: false}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			handlers := setupCompletenessCapabilityHandlers(testCtx.db, tc.source)

			_, body := getAllCapabilities(t, handlers)

			assert.NotContains(t, body, "onePagerComplete")
		})
	}
}
