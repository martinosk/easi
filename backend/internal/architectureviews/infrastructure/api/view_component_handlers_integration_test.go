//go:build integration
// +build integration

package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddComponentToView_Integration(t *testing.T) {
	testCtx, cleanup := setupViewTestDB(t)
	defer cleanup()

	h := setupViewHandlers(testCtx.db)

	viewID := testCtx.createViewViaAPI(t, h, "Test View", "Test Description")

	componentID := fmt.Sprintf("comp-%d", time.Now().UnixNano())
	testCtx.addComponentViaAPI(t, h, viewID, componentID, position{150.5, 250.5})

	view := testCtx.getViewViaAPI(t, h, viewID)
	assert.Len(t, view.Components, 1)
	assert.Equal(t, componentID, view.Components[0].ComponentID)
	assert.Equal(t, 150.5, view.Components[0].X)
	assert.Equal(t, 250.5, view.Components[0].Y)
}

func TestUpdateComponentPosition_Integration(t *testing.T) {
	testCtx, cleanup := setupViewTestDB(t)
	defer cleanup()

	h := setupViewHandlers(testCtx.db)

	viewID := testCtx.createViewViaAPI(t, h, "Test View", "Test Description")
	componentID := fmt.Sprintf("comp-%d", time.Now().UnixNano())
	testCtx.addComponentViaAPI(t, h, viewID, componentID, position{100.0, 200.0})

	updateBody, _ := json.Marshal(UpdatePositionRequest{X: 300.0, Y: 400.0})
	w, req := testCtx.makeRequest(t, http.MethodPatch, "/api/v1/views/"+viewID+"/components/"+componentID+"/position", updateBody, map[string]string{
		"id":          viewID,
		"componentId": componentID,
	})
	h.componentHandlers.UpdateComponentPosition(w, req)
	require.Equal(t, http.StatusNoContent, w.Code)

	view := testCtx.getViewViaAPI(t, h, viewID)
	assert.Len(t, view.Components, 1)
	assert.Equal(t, 300.0, view.Components[0].X)
	assert.Equal(t, 400.0, view.Components[0].Y)
}

func TestAddComponentToView_ViewNotFound_Integration(t *testing.T) {
	testCtx, cleanup := setupViewTestDB(t)
	defer cleanup()

	h := setupViewHandlers(testCtx.db)

	nonExistentViewID := fmt.Sprintf("non-existent-view-%d", time.Now().UnixNano())
	componentID := fmt.Sprintf("comp-%d", time.Now().UnixNano())

	body, _ := json.Marshal(AddComponentRequest{ComponentID: componentID, X: 150.5, Y: 250.5})
	w, req := testCtx.makeRequest(t, http.MethodPost, "/api/v1/views/"+nonExistentViewID+"/components", body, map[string]string{"id": nonExistentViewID})
	h.componentHandlers.AddComponentToView(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
