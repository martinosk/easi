//go:build integration
// +build integration

package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
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

type colorTestEnv struct {
	testCtx *viewTestContext
	h       *viewTestHarness
	viewID  string
}

func setupColorTest(t *testing.T) (colorTestEnv, func()) {
	testCtx, cleanup := setupViewTestDB(t)
	h := setupViewHandlers(testCtx.db)
	viewID := testCtx.createViewViaAPI(t, h, "Test View", "Test Description")
	return colorTestEnv{testCtx: testCtx, h: h, viewID: viewID}, cleanup
}

func (env colorTestEnv) patchColorScheme(t *testing.T, scheme string) *httptest.ResponseRecorder {
	body, _ := json.Marshal(UpdateColorSchemeRequest{ColorScheme: scheme})
	w, req := env.testCtx.makeRequest(t, http.MethodPatch, "/api/v1/views/"+env.viewID+"/color-scheme", body, map[string]string{"id": env.viewID})
	env.h.colorHandlers.UpdateColorScheme(w, req)
	return w
}

func (env colorTestEnv) addElement(t *testing.T, et elementColorConfig, elementID string, pos position) {
	et.addFn(env.testCtx, t, env.h, env.viewID, elementID, pos)
}

func (env colorTestEnv) setColor(t *testing.T, et elementColorConfig, elementID, color string) {
	env.testCtx.setElementColorViaAPI(t, elementColorRequest{h: env.h, viewID: env.viewID, elementID: elementID, elementType: et.name, color: color})
}

func (env colorTestEnv) currentColors(t *testing.T, et elementColorConfig) map[string]*string {
	return et.colorsFn(env.testCtx.getViewViaAPI(t, env.h, env.viewID))
}

func uniqueElementID(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}

func TestUpdateColorScheme_AllValidValues_Integration(t *testing.T) {
	env, cleanup := setupColorTest(t)
	defer cleanup()

	for _, scheme := range []string{"maturity", "classic", "custom"} {
		w := env.patchColorScheme(t, scheme)
		assert.Equal(t, http.StatusOK, w.Code)

		var response struct {
			ColorScheme string      `json:"colorScheme"`
			Links       types.Links `json:"_links"`
		}
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)
		assert.Equal(t, scheme, response.ColorScheme)
		assert.Contains(t, response.Links, "self")
		assert.Equal(t, "/api/v1/views/"+env.viewID+"/color-scheme", response.Links["self"].Href)
		assert.Contains(t, response.Links, "view")
		assert.Equal(t, "/api/v1/views/"+env.viewID, response.Links["view"].Href)

		view := env.testCtx.getViewViaAPI(t, env.h, env.viewID)
		assert.Equal(t, scheme, view.ColorScheme)
	}
}

func TestUpdateColorScheme_InvalidValue_Integration(t *testing.T) {
	env, cleanup := setupColorTest(t)
	defer cleanup()

	w := env.patchColorScheme(t, "invalid-scheme")

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetViewByID_ReturnsColorScheme_Integration(t *testing.T) {
	env, cleanup := setupColorTest(t)
	defer cleanup()

	w := env.patchColorScheme(t, "classic")
	require.Equal(t, http.StatusOK, w.Code)

	view := env.testCtx.getViewViaAPI(t, env.h, env.viewID)
	assert.Equal(t, "classic", view.ColorScheme)
}

func TestUpdateElementColor_Integration(t *testing.T) {
	for _, et := range elementTypes {
		t.Run(et.name, func(t *testing.T) {
			env, cleanup := setupColorTest(t)
			defer cleanup()

			elementID := uniqueElementID(et.prefix)
			env.addElement(t, et, elementID, position{100.0, 200.0})
			env.setColor(t, et, elementID, "#FF5733")

			colors := env.currentColors(t, et)
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
			env, cleanup := setupColorTest(t)
			defer cleanup()

			componentID := uniqueElementID("comp")
			env.testCtx.addComponentViaAPI(t, env.h, env.viewID, componentID, position{100.0, 200.0})

			body, _ := json.Marshal(UpdateElementColorRequest{Color: tc.color})
			w, req := env.testCtx.makeRequest(t, http.MethodPatch, "/api/v1/views/"+env.viewID+"/components/"+componentID+"/color", body, map[string]string{
				"id":          env.viewID,
				"componentId": componentID,
			})
			env.h.colorHandlers.UpdateComponentColor(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestClearElementColor_Integration(t *testing.T) {
	for _, et := range elementTypes {
		t.Run(et.name, func(t *testing.T) {
			env, cleanup := setupColorTest(t)
			defer cleanup()

			elementID := uniqueElementID(et.prefix)
			env.addElement(t, et, elementID, position{100.0, 200.0})
			env.setColor(t, et, elementID, "#FF5733")

			w, req := env.testCtx.makeRequest(t, http.MethodDelete, "/api/v1/views/"+env.viewID+"/"+et.urlPath+"/"+elementID+"/color", nil, map[string]string{
				"id":        env.viewID,
				et.urlParam: elementID,
			})
			et.clearFn(env.h)(w, req)
			require.Equal(t, http.StatusNoContent, w.Code)

			colors := env.currentColors(t, et)
			assert.Nil(t, colors[elementID])
		})
	}
}

func TestGetViewByID_ReturnsCustomColors_Integration(t *testing.T) {
	for _, et := range elementTypes {
		t.Run(et.name, func(t *testing.T) {
			env, cleanup := setupColorTest(t)
			defer cleanup()

			elem1 := uniqueElementID(et.prefix + "-1")
			elem2 := uniqueElementID(et.prefix + "-2")
			env.addElement(t, et, elem1, position{100.0, 200.0})
			env.addElement(t, et, elem2, position{300.0, 400.0})
			env.setColor(t, et, elem1, "#FF5733")

			colors := env.currentColors(t, et)
			require.NotNil(t, colors[elem1])
			assert.Equal(t, "#FF5733", *colors[elem1])
			assert.Nil(t, colors[elem2])
		})
	}
}

func TestGetViewByID_ReturnsHATEOASLinksForColors_Integration(t *testing.T) {
	env, cleanup := setupColorTest(t)
	defer cleanup()

	componentID := uniqueElementID("comp")
	capabilityID := uniqueElementID("cap")
	env.testCtx.addComponentViaAPI(t, env.h, env.viewID, componentID, position{100.0, 200.0})
	env.testCtx.addCapabilityViaAPI(t, env.h, env.viewID, capabilityID, position{150.0, 250.0})

	view := env.testCtx.getViewViaAPI(t, env.h, env.viewID)

	require.Len(t, view.Components, 1)
	compLinks := view.Components[0].Links
	assert.NotNil(t, compLinks)
	assert.Contains(t, compLinks, "x-update-color")
	assert.Contains(t, compLinks, "x-clear-color")
	assert.Equal(t, "/api/v1/views/"+env.viewID+"/components/"+componentID+"/color", compLinks["x-update-color"].Href)
	assert.Equal(t, "/api/v1/views/"+env.viewID+"/components/"+componentID+"/color", compLinks["x-clear-color"].Href)

	require.Len(t, view.Capabilities, 1)
	capLinks := view.Capabilities[0].Links
	assert.NotNil(t, capLinks)
	assert.Contains(t, capLinks, "x-update-color")
	assert.Contains(t, capLinks, "x-clear-color")
	assert.Equal(t, "/api/v1/views/"+env.viewID+"/capabilities/"+capabilityID+"/color", capLinks["x-update-color"].Href)
	assert.Equal(t, "/api/v1/views/"+env.viewID+"/capabilities/"+capabilityID+"/color", capLinks["x-clear-color"].Href)
}
