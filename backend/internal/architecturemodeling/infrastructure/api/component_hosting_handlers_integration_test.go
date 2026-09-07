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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (s *componentStack) classifyHosting(t *testing.T, ctx *testContext, componentID, hosting string) *httptest.ResponseRecorder {
	payload, err := json.Marshal(ClassifyHostingRequest{Hosting: hosting})
	require.NoError(t, err)
	w, req := ctx.makeRequest(t, requestSpec{
		Method:    http.MethodPut,
		URL:       "/api/v1/components/" + componentID + "/hosting",
		Body:      payload,
		URLParams: map[string]string{"id": componentID},
	})
	req = withArchitectActor(req)
	s.hosting.ClassifyHosting(w, req)
	return w
}

func TestHostingClassification_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()
	stack := setupComponentStack(testCtx.db)

	componentID := testCtx.createComponentViaAPI(t, stack.component, "Hosting Service", "")

	created, err := stack.readModel.GetByID(tenantContext(), componentID)
	require.NoError(t, err)
	assert.Equal(t, "unknown", created.Hosting)

	w := stack.classifyHosting(t, testCtx, componentID, "saas")
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	classified := decodeComponent(t, w.Body)
	assert.Equal(t, "saas", classified["hosting"])
	assert.Contains(t, classified["_links"].(map[string]any), "x-classify-hosting")

	w = stack.classifyHosting(t, testCtx, componentID, "on-premises")
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	assert.Equal(t, "on-premises", decodeComponent(t, w.Body)["hosting"])
}

func TestHostingClassificationInvalidValue_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()
	stack := setupComponentStack(testCtx.db)

	componentID := testCtx.createComponentViaAPI(t, stack.component, "Hosting Invalid Service", "")

	w := stack.classifyHosting(t, testCtx, componentID, "mainframe")
	assert.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
}

func TestHostingFilter_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()
	stack := setupComponentStack(testCtx.db)

	prefix := fmt.Sprintf("HostFilter-%d", time.Now().UnixNano())
	saasID := testCtx.createComponentViaAPI(t, stack.component, prefix+" A", "")
	testCtx.createComponentViaAPI(t, stack.component, prefix+" B", "")

	w := stack.classifyHosting(t, testCtx, saasID, "saas")
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())

	rw, req := testCtx.makeRequest(t, requestSpec{
		Method: http.MethodGet,
		URL:    "/api/v1/components?name=" + prefix + "&hosting=saas",
	})
	req = withArchitectActor(req)
	stack.component.GetAllComponents(rw, req)
	require.Equal(t, http.StatusOK, rw.Code, rw.Body.String())

	var page struct {
		Data []map[string]any `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rw.Body.Bytes(), &page))
	require.Len(t, page.Data, 1)
	assert.Equal(t, saasID, page.Data[0]["id"])
	assert.Equal(t, "saas", page.Data[0]["hosting"])
}

func TestHostingFilterInvalidValue_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()
	stack := setupComponentStack(testCtx.db)

	rw, req := testCtx.makeRequest(t, requestSpec{
		Method: http.MethodGet,
		URL:    "/api/v1/components?hosting=mainframe",
	})
	req = withArchitectActor(req)
	stack.component.GetAllComponents(rw, req)
	assert.Equal(t, http.StatusBadRequest, rw.Code, rw.Body.String())
}

func TestHostingStatistics_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()
	stack := setupComponentStack(testCtx.db)

	before, err := stack.readModel.Statistics(tenantContext())
	require.NoError(t, err)

	cloudID := testCtx.createComponentViaAPI(t, stack.component, "Hosting Stats A", "")
	testCtx.createComponentViaAPI(t, stack.component, "Hosting Stats B", "")

	w := stack.classifyHosting(t, testCtx, cloudID, "cloud")
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())

	after, err := stack.readModel.Statistics(tenantContext())
	require.NoError(t, err)
	assert.Equal(t, before.Total+2, after.Total)
	assert.Equal(t, before.Hosting.Cloud+1, after.Hosting.Cloud)
	assert.Equal(t, before.Hosting.Unknown+1, after.Hosting.Unknown)
}
