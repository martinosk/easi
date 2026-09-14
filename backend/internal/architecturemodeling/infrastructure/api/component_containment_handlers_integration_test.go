//go:build integration
// +build integration

package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"easi/backend/internal/architecturemodeling/domain/valueobjects"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type attachment struct {
	part, parent, kind string
}

func (s *componentStack) attach(t *testing.T, ctx *testContext, a attachment) *httptest.ResponseRecorder {
	payload, err := json.Marshal(AttachComponentRequest{ParentID: a.parent, Kind: a.kind})
	require.NoError(t, err)
	w, req := ctx.makeRequest(t, requestSpec{
		Method:    http.MethodPut,
		URL:       "/api/v1/components/" + a.part + "/containment",
		Body:      payload,
		URLParams: map[string]string{"id": a.part},
	})
	s.containment.AttachComponent(w, withArchitectActor(req))
	return w
}

func (s *componentStack) detach(t *testing.T, ctx *testContext, partID string) *httptest.ResponseRecorder {
	w, req := ctx.makeRequest(t, requestSpec{
		Method:    http.MethodDelete,
		URL:       "/api/v1/components/" + partID + "/containment",
		URLParams: map[string]string{"id": partID},
	})
	s.containment.DetachComponent(w, withArchitectActor(req))
	return w
}

func (s *componentStack) deleteComponent(t *testing.T, ctx *testContext, id string) *httptest.ResponseRecorder {
	w, req := ctx.makeRequest(t, requestSpec{
		Method:    http.MethodDelete,
		URL:       "/api/v1/components/" + id,
		URLParams: map[string]string{"id": id},
	})
	s.component.DeleteApplicationComponent(w, withArchitectActor(req))
	return w
}

func (s *componentStack) getComponent(t *testing.T, ctx *testContext, id string) *httptest.ResponseRecorder {
	w, req := ctx.makeRequest(t, requestSpec{
		Method:    http.MethodGet,
		URL:       "/api/v1/components/" + id,
		URLParams: map[string]string{"id": id},
	})
	s.component.GetComponentByID(w, withArchitectActor(req))
	return w
}

func (s *componentStack) attachOK(t *testing.T, ctx *testContext, a attachment) map[string]any {
	t.Helper()
	w := s.attach(t, ctx, a)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	return decodeComponent(t, w.Body)
}

func (s *componentStack) fetchComponent(t *testing.T, ctx *testContext, id string) map[string]any {
	t.Helper()
	w := s.getComponent(t, ctx, id)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	return decodeComponent(t, w.Body)
}

func links(component map[string]any) map[string]any {
	return component["_links"].(map[string]any)
}

func partIDs(component map[string]any) []string {
	parts, _ := component["parts"].([]any)
	ids := make([]string, 0, len(parts))
	for _, part := range parts {
		ids = append(ids, part.(map[string]any)["id"].(string))
	}
	return ids
}

func TestContainmentAttach_Integration(t *testing.T) {
	for _, kind := range []string{valueobjects.ContainmentComposition, valueobjects.ContainmentAggregation} {
		t.Run(kind, func(t *testing.T) {
			testCtx, cleanup := setupTestDB(t)
			defer cleanup()
			stack := setupComponentStack(testCtx.db)
			parentID := testCtx.createComponentViaAPI(t, stack.component, "CRM Suite "+kind, "")
			partID := testCtx.createComponentViaAPI(t, stack.component, "Quoting "+kind, "")

			part := stack.attachOK(t, testCtx, attachment{partID, parentID, kind})

			partOf := part["partOf"].(map[string]any)
			assert.Equal(t, parentID, partOf["id"])
			assert.Equal(t, "CRM Suite "+kind, partOf["name"])
			assert.Equal(t, kind, partOf["kind"])
			assert.Contains(t, links(part), "x-detach")
			assert.NotContains(t, links(part), "x-attach-to")

			parent := stack.fetchComponent(t, testCtx, parentID)
			assert.Equal(t, []string{partID}, partIDs(parent))
			assert.Equal(t, kind, parent["parts"].([]any)[0].(map[string]any)["kind"])
			assert.NotContains(t, links(parent), "x-attach-to")
			assert.NotContains(t, links(parent), "x-detach")
		})
	}
}

func TestContainmentRules_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()
	stack := setupComponentStack(testCtx.db)
	crmID := testCtx.createComponentViaAPI(t, stack.component, "CRM Suite", "")
	erpID := testCtx.createComponentViaAPI(t, stack.component, "ERP Suite", "")
	quotingID := testCtx.createComponentViaAPI(t, stack.component, "Quoting", "")
	billingID := testCtx.createComponentViaAPI(t, stack.component, "Billing", "")
	stack.attachOK(t, testCtx, attachment{quotingID, crmID, valueobjects.ContainmentComposition})

	cases := []struct {
		name           string
		part, parent   string
		kind           string
		expectedStatus int
	}{
		{"a part has exactly one parent", quotingID, erpID, valueobjects.ContainmentAggregation, http.StatusConflict},
		{"a part cannot accept parts", billingID, quotingID, valueobjects.ContainmentAggregation, http.StatusConflict},
		{"a parent cannot become a part", crmID, erpID, valueobjects.ContainmentAggregation, http.StatusConflict},
		{"a component cannot contain itself", erpID, erpID, valueobjects.ContainmentAggregation, http.StatusConflict},
		{"unknown kind is rejected", billingID, erpID, "nesting", http.StatusBadRequest},
		{"unknown parent is rejected", billingID, valueobjects.NewComponentID().Value(), valueobjects.ContainmentAggregation, http.StatusBadRequest},
		{"unknown part is rejected", valueobjects.NewComponentID().Value(), erpID, valueobjects.ContainmentAggregation, http.StatusNotFound},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := stack.attach(t, testCtx, attachment{tc.part, tc.parent, tc.kind})

			assert.Equal(t, tc.expectedStatus, w.Code, w.Body.String())
		})
	}

	billing := stack.fetchComponent(t, testCtx, billingID)
	assert.Nil(t, billing["partOf"])
}

func TestContainmentDetach_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()
	stack := setupComponentStack(testCtx.db)
	crmID := testCtx.createComponentViaAPI(t, stack.component, "CRM Suite", "")
	quotingID := testCtx.createComponentViaAPI(t, stack.component, "Quoting", "")
	stack.attachOK(t, testCtx, attachment{quotingID, crmID, valueobjects.ContainmentComposition})

	w := stack.detach(t, testCtx, quotingID)

	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	quoting := decodeComponent(t, w.Body)
	assert.Nil(t, quoting["partOf"])
	assert.Contains(t, links(quoting), "x-attach-to")
	assert.NotContains(t, links(quoting), "x-detach")
	parent := stack.fetchComponent(t, testCtx, crmID)
	assert.Empty(t, partIDs(parent))
	assert.Contains(t, links(parent), "x-attach-to")

	again := stack.detach(t, testCtx, quotingID)
	assert.Equal(t, http.StatusConflict, again.Code, again.Body.String())
}

func TestContainmentDeleteParent_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()
	stack := setupComponentStack(testCtx.db)
	crmID := testCtx.createComponentViaAPI(t, stack.component, "CRM Suite", "")
	quotingID := testCtx.createComponentViaAPI(t, stack.component, "Quoting", "")
	billingID := testCtx.createComponentViaAPI(t, stack.component, "Billing", "")
	stack.attachOK(t, testCtx, attachment{quotingID, crmID, valueobjects.ContainmentComposition})
	stack.attachOK(t, testCtx, attachment{billingID, crmID, valueobjects.ContainmentAggregation})

	w := stack.deleteComponent(t, testCtx, crmID)

	require.Equal(t, http.StatusNoContent, w.Code, w.Body.String())
	assert.Equal(t, http.StatusNotFound, stack.getComponent(t, testCtx, crmID).Code)
	assert.Equal(t, http.StatusNotFound, stack.getComponent(t, testCtx, quotingID).Code)
	billing := stack.fetchComponent(t, testCtx, billingID)
	assert.Nil(t, billing["partOf"])
	assert.Contains(t, links(billing), "x-attach-to")
}

func TestContainmentDeletePart_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()
	stack := setupComponentStack(testCtx.db)
	crmID := testCtx.createComponentViaAPI(t, stack.component, "CRM Suite", "")
	quotingID := testCtx.createComponentViaAPI(t, stack.component, "Quoting", "")
	stack.attachOK(t, testCtx, attachment{quotingID, crmID, valueobjects.ContainmentComposition})

	w := stack.deleteComponent(t, testCtx, quotingID)

	require.Equal(t, http.StatusNoContent, w.Code, w.Body.String())
	parent := stack.fetchComponent(t, testCtx, crmID)
	assert.Empty(t, partIDs(parent))
	assert.Contains(t, links(parent), "x-attach-to")
}
