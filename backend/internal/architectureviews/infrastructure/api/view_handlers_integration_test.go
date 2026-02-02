//go:build integration
// +build integration

package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"easi/backend/internal/architectureviews/application/readmodels"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateView_Integration(t *testing.T) {
	testCtx, cleanup := setupViewTestDB(t)
	defer cleanup()

	h := setupViewHandlers(testCtx.db)

	body, _ := json.Marshal(CreateViewRequest{
		Name:        "System Architecture",
		Description: "Overall system architecture view",
	})
	w, req := testCtx.makeRequest(t, http.MethodPost, "/api/v1/views", body, nil)
	h.viewHandlers.CreateView(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response readmodels.ArchitectureViewDTO
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.NotEmpty(t, response.ID)
	assert.Equal(t, "System Architecture", response.Name)
	assert.Equal(t, "Overall system architecture view", response.Description)
	assert.NotNil(t, response.Links)

	testCtx.trackID(response.ID)
}

func TestCreateView_ValidationErrors_Integration(t *testing.T) {
	testCases := []struct {
		name        string
		viewName    string
		description string
	}{
		{"EmptyName", "", "Some description"},
		{"NameTooLong", "This is a view name that is one hundred and one characters long and should fail the validation tests!", "Some description"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testCtx, cleanup := setupViewTestDB(t)
			defer cleanup()

			h := setupViewHandlers(testCtx.db)

			body, _ := json.Marshal(CreateViewRequest{Name: tc.viewName, Description: tc.description})
			w, req := testCtx.makeRequest(t, http.MethodPost, "/api/v1/views", body, nil)
			h.viewHandlers.CreateView(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestGetAllViews_Integration(t *testing.T) {
	testCtx, cleanup := setupViewTestDB(t)
	defer cleanup()

	h := setupViewHandlers(testCtx.db)

	id1 := testCtx.createViewViaAPI(t, h, "View A", "Description A")
	id2 := testCtx.createViewViaAPI(t, h, "View B", "Description B")

	w, req := testCtx.makeRequest(t, http.MethodGet, "/api/v1/views", nil, nil)
	h.viewHandlers.GetAllViews(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Data []readmodels.ArchitectureViewDTO `json:"data"`
	}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	foundViews := 0
	for _, view := range response.Data {
		if view.ID == id1 || view.ID == id2 {
			foundViews++
			assert.NotNil(t, view.Links)
			assert.Contains(t, view.Links, "self")
		}
	}
	assert.Equal(t, 2, foundViews, "Should find both test views")
}

func TestGetViewByID_Integration(t *testing.T) {
	testCtx, cleanup := setupViewTestDB(t)
	defer cleanup()

	h := setupViewHandlers(testCtx.db)

	viewID := testCtx.createViewViaAPI(t, h, "Test View", "Test Description")
	comp1 := fmt.Sprintf("comp-1-%d", time.Now().UnixNano())
	comp2 := fmt.Sprintf("comp-2-%d", time.Now().UnixNano())
	testCtx.addComponentViaAPI(t, h, viewID, comp1, position{100.0, 200.0})
	testCtx.addComponentViaAPI(t, h, viewID, comp2, position{300.0, 400.0})

	view := testCtx.getViewViaAPI(t, h, viewID)

	assert.Equal(t, viewID, view.ID)
	assert.Equal(t, "Test View", view.Name)
	assert.Equal(t, "Test Description", view.Description)
	assert.Len(t, view.Components, 2)
	assert.NotNil(t, view.Links)
}

func TestGetViewByID_NotFound_Integration(t *testing.T) {
	testCtx, cleanup := setupViewTestDB(t)
	defer cleanup()

	h := setupViewHandlers(testCtx.db)

	nonExistentID := fmt.Sprintf("non-existent-%d", time.Now().UnixNano())
	w, req := testCtx.makeRequest(t, http.MethodGet, "/api/v1/views/"+nonExistentID, nil, map[string]string{"id": nonExistentID})
	h.viewHandlers.GetViewByID(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestSetDefaultView_Integration(t *testing.T) {
	testCtx, cleanup := setupViewTestDB(t)
	defer cleanup()

	h := setupViewHandlers(testCtx.db)

	view1ID := testCtx.createViewViaAPI(t, h, "View 1", "First view")
	view2ID := testCtx.createViewViaAPI(t, h, "View 2", "Second view")

	w1, req1 := testCtx.makeRequest(t, http.MethodPut, "/api/v1/views/"+view1ID+"/default", nil, map[string]string{"id": view1ID})
	h.viewHandlers.SetDefaultView(w1, req1)
	require.Equal(t, http.StatusNoContent, w1.Code)

	view1 := testCtx.getViewViaAPI(t, h, view1ID)
	assert.True(t, view1.IsDefault)

	w2, req2 := testCtx.makeRequest(t, http.MethodPut, "/api/v1/views/"+view2ID+"/default", nil, map[string]string{"id": view2ID})
	h.viewHandlers.SetDefaultView(w2, req2)
	require.Equal(t, http.StatusNoContent, w2.Code)

	view2 := testCtx.getViewViaAPI(t, h, view2ID)
	assert.True(t, view2.IsDefault)

	view1 = testCtx.getViewViaAPI(t, h, view1ID)
	assert.False(t, view1.IsDefault)
}

func TestViewSettings_Integration(t *testing.T) {
	testCases := []struct {
		name        string
		endpoint    string
		fieldName   string
		validValues []string
		readField   func(readmodels.ArchitectureViewDTO) string
		handler     func(*viewTestHarness) func(http.ResponseWriter, *http.Request)
	}{
		{
			name:        "EdgeType",
			endpoint:    "/edge-type",
			fieldName:   "edgeType",
			validValues: []string{"default", "step", "smoothstep", "straight"},
			readField:   func(v readmodels.ArchitectureViewDTO) string { return v.EdgeType },
			handler:     func(h *viewTestHarness) func(http.ResponseWriter, *http.Request) { return h.viewHandlers.UpdateEdgeType },
		},
		{
			name:        "LayoutDirection",
			endpoint:    "/layout-direction",
			fieldName:   "layoutDirection",
			validValues: []string{"TB", "LR", "BT", "RL"},
			readField:   func(v readmodels.ArchitectureViewDTO) string { return v.LayoutDirection },
			handler:     func(h *viewTestHarness) func(http.ResponseWriter, *http.Request) { return h.viewHandlers.UpdateLayoutDirection },
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testCtx, cleanup := setupViewTestDB(t)
			defer cleanup()

			h := setupViewHandlers(testCtx.db)
			viewID := testCtx.createViewViaAPI(t, h, "Test View", "Test Description")

			for _, value := range tc.validValues {
				body, _ := json.Marshal(map[string]string{tc.fieldName: value})
				w, req := testCtx.makeRequest(t, http.MethodPatch, "/api/v1/views/"+viewID+tc.endpoint, body, map[string]string{"id": viewID})
				tc.handler(h)(w, req)
				assert.Equal(t, http.StatusNoContent, w.Code)

				view := testCtx.getViewViaAPI(t, h, viewID)
				assert.Equal(t, value, tc.readField(view))
			}
		})
	}
}

func TestViewSettings_InvalidValue_Integration(t *testing.T) {
	testCases := []struct {
		name      string
		endpoint  string
		fieldName string
		value     string
		handler   func(*viewTestHarness) func(http.ResponseWriter, *http.Request)
	}{
		{"InvalidEdgeType", "/edge-type", "edgeType", "invalid", func(h *viewTestHarness) func(http.ResponseWriter, *http.Request) { return h.viewHandlers.UpdateEdgeType }},
		{"InvalidLayoutDirection", "/layout-direction", "layoutDirection", "INVALID", func(h *viewTestHarness) func(http.ResponseWriter, *http.Request) { return h.viewHandlers.UpdateLayoutDirection }},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testCtx, cleanup := setupViewTestDB(t)
			defer cleanup()

			h := setupViewHandlers(testCtx.db)
			viewID := testCtx.createViewViaAPI(t, h, "Test View", "Test Description")

			body, _ := json.Marshal(map[string]string{tc.fieldName: tc.value})
			w, req := testCtx.makeRequest(t, http.MethodPatch, "/api/v1/views/"+viewID+tc.endpoint, body, map[string]string{"id": viewID})
			tc.handler(h)(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}
