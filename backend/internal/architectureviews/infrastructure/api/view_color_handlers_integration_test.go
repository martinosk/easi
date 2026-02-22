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
	"easi/backend/internal/shared/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type elementColorConfig struct {
	name     string
	prefix   string
	addFn    func(*viewTestContext, *testing.T, *viewTestHarness, string, string, position)
	colorsFn func(readmodels.ArchitectureViewDTO) map[string]*string
	clearFn  func(*viewTestHarness) func(http.ResponseWriter, *http.Request)
	urlPath  string
	urlParam string
}

var elementTypes = []elementColorConfig{
	{
		name:   "component",
		prefix: "comp",
		addFn: func(ctx *viewTestContext, t *testing.T, h *viewTestHarness, viewID, id string, pos position) {
			ctx.addComponentViaAPI(t, h, viewID, id, pos)
		},
		colorsFn: func(v readmodels.ArchitectureViewDTO) map[string]*string {
			m := map[string]*string{}
			for _, c := range v.Components {
				m[c.ComponentID] = c.CustomColor
			}
			return m
		},
		clearFn: func(h *viewTestHarness) func(http.ResponseWriter, *http.Request) {
			return h.colorHandlers.ClearComponentColor
		},
		urlPath:  "components",
		urlParam: "componentId",
	},
	{
		name:   "capability",
		prefix: "cap",
		addFn: func(ctx *viewTestContext, t *testing.T, h *viewTestHarness, viewID, id string, pos position) {
			ctx.addCapabilityViaAPI(t, h, viewID, id, pos)
		},
		colorsFn: func(v readmodels.ArchitectureViewDTO) map[string]*string {
			m := map[string]*string{}
			for _, c := range v.Capabilities {
				m[c.CapabilityID] = c.CustomColor
			}
			return m
		},
		clearFn: func(h *viewTestHarness) func(http.ResponseWriter, *http.Request) {
			return h.colorHandlers.ClearCapabilityColor
		},
		urlPath:  "capabilities",
		urlParam: "capabilityId",
	},
}

func TestUpdateColorScheme_AllValidValues_Integration(t *testing.T) {
	testCtx, cleanup := setupViewTestDB(t)
	defer cleanup()

	h := setupViewHandlers(testCtx.db)
	viewID := testCtx.createViewViaAPI(t, h, "Test View", "Test Description")

	for _, scheme := range []string{"maturity", "classic", "custom"} {
		body, _ := json.Marshal(UpdateColorSchemeRequest{ColorScheme: scheme})
		w, req := testCtx.makeRequest(t, http.MethodPatch, "/api/v1/views/"+viewID+"/color-scheme", body, map[string]string{"id": viewID})
		h.colorHandlers.UpdateColorScheme(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		var response struct {
			ColorScheme string      `json:"colorScheme"`
			Links       types.Links `json:"_links"`
		}
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)
		assert.Equal(t, scheme, response.ColorScheme)
		assert.Contains(t, response.Links, "self")
		assert.Equal(t, "/api/v1/views/"+viewID+"/color-scheme", response.Links["self"].Href)
		assert.Contains(t, response.Links, "view")
		assert.Equal(t, "/api/v1/views/"+viewID, response.Links["view"].Href)

		view := testCtx.getViewViaAPI(t, h, viewID)
		assert.Equal(t, scheme, view.ColorScheme)
	}
}

func TestUpdateColorScheme_InvalidValue_Integration(t *testing.T) {
	testCtx, cleanup := setupViewTestDB(t)
	defer cleanup()

	h := setupViewHandlers(testCtx.db)
	viewID := testCtx.createViewViaAPI(t, h, "Test View", "Test Description")

	body, _ := json.Marshal(UpdateColorSchemeRequest{ColorScheme: "invalid-scheme"})
	w, req := testCtx.makeRequest(t, http.MethodPatch, "/api/v1/views/"+viewID+"/color-scheme", body, map[string]string{"id": viewID})
	h.colorHandlers.UpdateColorScheme(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetViewByID_ReturnsColorScheme_Integration(t *testing.T) {
	testCtx, cleanup := setupViewTestDB(t)
	defer cleanup()

	h := setupViewHandlers(testCtx.db)
	viewID := testCtx.createViewViaAPI(t, h, "Test View", "Test Description")

	body, _ := json.Marshal(UpdateColorSchemeRequest{ColorScheme: "classic"})
	w, req := testCtx.makeRequest(t, http.MethodPatch, "/api/v1/views/"+viewID+"/color-scheme", body, map[string]string{"id": viewID})
	h.colorHandlers.UpdateColorScheme(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	view := testCtx.getViewViaAPI(t, h, viewID)
	assert.Equal(t, "classic", view.ColorScheme)
}

func TestUpdateElementColor_Integration(t *testing.T) {
	for _, et := range elementTypes {
		t.Run(et.name, func(t *testing.T) {
			testCtx, cleanup := setupViewTestDB(t)
			defer cleanup()

			h := setupViewHandlers(testCtx.db)
			viewID := testCtx.createViewViaAPI(t, h, "Test View", "Test Description")
			elementID := fmt.Sprintf("%s-%d", et.prefix, time.Now().UnixNano())
			et.addFn(testCtx, t, h, viewID, elementID, position{100.0, 200.0})

			testCtx.setElementColorViaAPI(t, h, viewID, elementID, et.name, "#FF5733")

			view := testCtx.getViewViaAPI(t, h, viewID)
			colors := et.colorsFn(view)
			require.NotNil(t, colors[elementID])
			assert.Equal(t, "#FF5733", *colors[elementID])
		})
	}
}

func TestUpdateComponentColor_InvalidValues_Integration(t *testing.T) {
	testCases := []struct {
		name  string
		color string
	}{
		{"InvalidHexColor", "invalid-color"},
		{"MissingHash", "FF5733"},
		{"TooShort", "#FFF"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testCtx, cleanup := setupViewTestDB(t)
			defer cleanup()

			h := setupViewHandlers(testCtx.db)
			viewID := testCtx.createViewViaAPI(t, h, "Test View", "Test Description")
			componentID := fmt.Sprintf("comp-%d", time.Now().UnixNano())
			testCtx.addComponentViaAPI(t, h, viewID, componentID, position{100.0, 200.0})

			body, _ := json.Marshal(UpdateElementColorRequest{Color: tc.color})
			w, req := testCtx.makeRequest(t, http.MethodPatch, "/api/v1/views/"+viewID+"/components/"+componentID+"/color", body, map[string]string{
				"id":          viewID,
				"componentId": componentID,
			})
			h.colorHandlers.UpdateComponentColor(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestClearElementColor_Integration(t *testing.T) {
	for _, et := range elementTypes {
		t.Run(et.name, func(t *testing.T) {
			testCtx, cleanup := setupViewTestDB(t)
			defer cleanup()

			h := setupViewHandlers(testCtx.db)
			viewID := testCtx.createViewViaAPI(t, h, "Test View", "Test Description")
			elementID := fmt.Sprintf("%s-%d", et.prefix, time.Now().UnixNano())
			et.addFn(testCtx, t, h, viewID, elementID, position{100.0, 200.0})

			testCtx.setElementColorViaAPI(t, h, viewID, elementID, et.name, "#FF5733")

			w, req := testCtx.makeRequest(t, http.MethodDelete, "/api/v1/views/"+viewID+"/"+et.urlPath+"/"+elementID+"/color", nil, map[string]string{
				"id":        viewID,
				et.urlParam: elementID,
			})
			et.clearFn(h)(w, req)
			require.Equal(t, http.StatusNoContent, w.Code)

			view := testCtx.getViewViaAPI(t, h, viewID)
			colors := et.colorsFn(view)
			assert.Nil(t, colors[elementID])
		})
	}
}

func TestGetViewByID_ReturnsCustomColors_Integration(t *testing.T) {
	for _, et := range elementTypes {
		t.Run(et.name, func(t *testing.T) {
			testCtx, cleanup := setupViewTestDB(t)
			defer cleanup()

			h := setupViewHandlers(testCtx.db)
			viewID := testCtx.createViewViaAPI(t, h, "Test View", "Test Description")

			elem1 := fmt.Sprintf("%s-1-%d", et.prefix, time.Now().UnixNano())
			elem2 := fmt.Sprintf("%s-2-%d", et.prefix, time.Now().UnixNano())
			et.addFn(testCtx, t, h, viewID, elem1, position{100.0, 200.0})
			et.addFn(testCtx, t, h, viewID, elem2, position{300.0, 400.0})

			testCtx.setElementColorViaAPI(t, h, viewID, elem1, et.name, "#FF5733")

			view := testCtx.getViewViaAPI(t, h, viewID)
			colors := et.colorsFn(view)
			require.NotNil(t, colors[elem1])
			assert.Equal(t, "#FF5733", *colors[elem1])
			assert.Nil(t, colors[elem2])
		})
	}
}

func TestGetViewByID_ReturnsHATEOASLinksForColors_Integration(t *testing.T) {
	testCtx, cleanup := setupViewTestDB(t)
	defer cleanup()

	h := setupViewHandlers(testCtx.db)
	viewID := testCtx.createViewViaAPI(t, h, "Test View", "Test Description")

	componentID := fmt.Sprintf("comp-%d", time.Now().UnixNano())
	capabilityID := fmt.Sprintf("cap-%d", time.Now().UnixNano())
	testCtx.addComponentViaAPI(t, h, viewID, componentID, position{100.0, 200.0})
	testCtx.addCapabilityViaAPI(t, h, viewID, capabilityID, position{150.0, 250.0})

	view := testCtx.getViewViaAPI(t, h, viewID)

	require.Len(t, view.Components, 1)
	compLinks := view.Components[0].Links
	assert.NotNil(t, compLinks)
	assert.Contains(t, compLinks, "x-update-color")
	assert.Contains(t, compLinks, "x-clear-color")
	assert.Equal(t, "/api/v1/views/"+viewID+"/components/"+componentID+"/color", compLinks["x-update-color"].Href)
	assert.Equal(t, "/api/v1/views/"+viewID+"/components/"+componentID+"/color", compLinks["x-clear-color"].Href)

	require.Len(t, view.Capabilities, 1)
	capLinks := view.Capabilities[0].Links
	assert.NotNil(t, capLinks)
	assert.Contains(t, capLinks, "x-update-color")
	assert.Contains(t, capLinks, "x-clear-color")
	assert.Equal(t, "/api/v1/views/"+viewID+"/capabilities/"+capabilityID+"/color", capLinks["x-update-color"].Href)
	assert.Equal(t, "/api/v1/views/"+viewID+"/capabilities/"+capabilityID+"/color", capLinks["x-clear-color"].Href)
}
