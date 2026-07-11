//go:build integration
// +build integration

package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"easi/backend/internal/architecturemodeling/application/readmodels"
	"easi/backend/internal/infrastructure/database"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/cqrs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupCompletenessComponentHandlers(db *sql.DB, source OnePagerCompletenessSource) *ComponentHandlers {
	tenantDB := database.NewTenantAwareDB(db)
	commandBus := cqrs.NewInMemoryCommandBus()
	links := NewArchitectureModelingLinks(sharedAPI.NewHATEOASLinks("/api/v1"))
	readModel := readmodels.NewApplicationComponentReadModel(tenantDB)
	return NewComponentHandlers(commandBus, readModel, links, source)
}

func seedCompletenessComponents(t *testing.T, testCtx *testContext) (string, string, string) {
	token := fmt.Sprintf("opcomplete%d", time.Now().UnixNano())
	idA := token + "-a"
	idB := token + "-b"
	testCtx.createTestComponent(t, idA, token+" Alpha", "")
	testCtx.createTestComponent(t, idB, token+" Beta", "")
	return token, idA, idB
}

func getComponentsPage(t *testing.T, testCtx *testContext, handlers *ComponentHandlers, token string) ([]map[string]any, string) {
	w, req := testCtx.makeRequest(t, requestSpec{Method: http.MethodGet, URL: "/api/v1/components?name=" + token})
	handlers.GetAllComponents(w, req)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())

	var response struct {
		Data []map[string]any `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
	return response.Data, w.Body.String()
}

func TestGetAllComponents_DecoratesRowsWhenIndicatorPresent_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	token, idA, idB := seedCompletenessComponents(t, testCtx)
	source := &fakeOnePagerCompletenessSource{indicators: map[string]bool{idA: true, idB: false}, present: true}
	handlers := setupCompletenessComponentHandlers(testCtx.db, source)

	rows, _ := getComponentsPage(t, testCtx, handlers, token)

	require.Len(t, rows, 2)
	assert.ElementsMatch(t, []string{idA, idB}, source.gotIDs)
	byID := map[string]map[string]any{}
	for _, row := range rows {
		byID[row["id"].(string)] = row
	}
	assert.Equal(t, true, byID[idA]["onePagerComplete"])
	assert.Equal(t, false, byID[idB]["onePagerComplete"])
}

func TestGetAllComponents_NoIndicatorWhenSourceNilOrAbsent_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	token, _, _ := seedCompletenessComponents(t, testCtx)

	cases := []struct {
		name   string
		source OnePagerCompletenessSource
	}{
		{"nil source", nil},
		{"indicator absent", &fakeOnePagerCompletenessSource{present: false}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			handlers := setupCompletenessComponentHandlers(testCtx.db, tc.source)

			rows, body := getComponentsPage(t, testCtx, handlers, token)

			require.Len(t, rows, 2)
			assert.NotContains(t, body, "onePagerComplete")
		})
	}
}
