package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"easi/backend/internal/architecturedirection/application/commands"
	"easi/backend/internal/architecturedirection/application/readmodels"
	"easi/backend/internal/architecturedirection/domain/valueobjects"
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRealizationRoleQueries struct {
	byPairReturns  []*readmodels.RealizationRoleDTO
	byPairCallIdx  int
	byCapability   []readmodels.RealizationRoleDTO
	receivedCapIDs []string
}

func (m *mockRealizationRoleQueries) GetByPair(_ context.Context, _, _ string) (*readmodels.RealizationRoleDTO, error) {
	var result *readmodels.RealizationRoleDTO
	if m.byPairCallIdx < len(m.byPairReturns) {
		result = m.byPairReturns[m.byPairCallIdx]
	}
	m.byPairCallIdx++
	return result, nil
}

func (m *mockRealizationRoleQueries) GetByCapabilityIDs(_ context.Context, ids []string) ([]readmodels.RealizationRoleDTO, error) {
	m.receivedCapIDs = ids
	return m.byCapability, nil
}

func setupRealizationRoleHandlers(bus *mockCommandBus, queries RealizationRoleQueries) *RealizationRoleHandlers {
	links := NewRealizationRoleLinks(sharedAPI.NewHATEOASLinks(""))
	return NewRealizationRoleHandlers(bus, queries, links)
}

func realizationRoleRouter(h *RealizationRoleHandlers) chi.Router {
	r := chi.NewRouter()
	r.Get("/capabilities/{id}/components/{componentId}/realization-role", h.GetRealizationRole)
	r.Put("/capabilities/{id}/components/{componentId}/realization-role", h.PutRealizationRole)
	r.Delete("/capabilities/{id}/components/{componentId}/realization-role", h.DeleteRealizationRole)
	r.Get("/realization-roles", h.GetRealizationRoles)
	return r
}

func TestGetRealizationRole_Unclassified_Returns404(t *testing.T) {
	queries := &mockRealizationRoleQueries{}
	h := setupRealizationRoleHandlers(&mockCommandBus{}, queries)
	r := realizationRoleRouter(h)

	capID, compID := uuid.New().String(), uuid.New().String()
	req := httptest.NewRequest(http.MethodGet, "/capabilities/"+capID+"/components/"+compID+"/realization-role", nil)
	req = req.WithContext(sharedctx.WithActor(req.Context(), architectActor()))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestGetRealizationRole_Assigned_EditDeleteLinkVisibilityByActor(t *testing.T) {
	capID, compID := uuid.New().String(), uuid.New().String()
	current := &readmodels.RealizationRoleDTO{CapabilityID: capID, ComponentID: compID, Role: valueobjects.RealizationRoleStandard}
	cases := []struct {
		name       string
		actor      sharedctx.Actor
		shouldShow bool
	}{
		{"architect sees edit/delete", architectActor(), true},
		{"stakeholder does not see edit/delete", stakeholderActor(), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			queries := &mockRealizationRoleQueries{byPairReturns: []*readmodels.RealizationRoleDTO{current}}
			h := setupRealizationRoleHandlers(&mockCommandBus{}, queries)
			r := realizationRoleRouter(h)

			req := httptest.NewRequest(http.MethodGet, "/capabilities/"+capID+"/components/"+compID+"/realization-role", nil)
			req = req.WithContext(sharedctx.WithActor(req.Context(), tc.actor))
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			require.Equal(t, http.StatusOK, rec.Code)
			var body readmodels.RealizationRoleDTO
			require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
			if tc.shouldShow {
				assert.Contains(t, body.Links, "edit")
				assert.Contains(t, body.Links, "delete")
			} else {
				assert.NotContains(t, body.Links, "edit")
				assert.NotContains(t, body.Links, "delete")
			}
		})
	}
}

func TestPutRealizationRole_FirstAssignment_Returns201WithLocation(t *testing.T) {
	capID, compID := uuid.New().String(), uuid.New().String()
	created := &readmodels.RealizationRoleDTO{CapabilityID: capID, ComponentID: compID, Role: valueobjects.RealizationRoleStandard}
	queries := &mockRealizationRoleQueries{byPairReturns: []*readmodels.RealizationRoleDTO{nil, created}}
	bus := &mockCommandBus{}
	h := setupRealizationRoleHandlers(bus, queries)
	r := realizationRoleRouter(h)

	body, _ := json.Marshal(AssignRealizationRoleRequest{Role: valueobjects.RealizationRoleStandard})
	req := httptest.NewRequest(http.MethodPut, "/capabilities/"+capID+"/components/"+compID+"/realization-role", bytes.NewReader(body))
	req = req.WithContext(sharedctx.WithActor(req.Context(), architectActor()))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	assert.Equal(t, "/api/v1/capabilities/"+capID+"/components/"+compID+"/realization-role", rec.Header().Get("Location"))
	require.Len(t, bus.dispatched, 1)
	cmd := bus.dispatched[0].(*commands.AssignRealizationRole)
	assert.Equal(t, capID, cmd.CapabilityID)
	assert.Equal(t, compID, cmd.ComponentID)
	assert.Equal(t, valueobjects.RealizationRoleStandard, cmd.Role)
	assert.Equal(t, "user@example.com", cmd.AssignedBy)
}

func TestPutRealizationRole_ReAssignment_Returns200(t *testing.T) {
	capID, compID := uuid.New().String(), uuid.New().String()
	existing := &readmodels.RealizationRoleDTO{CapabilityID: capID, ComponentID: compID, Role: valueobjects.RealizationRoleLegacy}
	replaced := &readmodels.RealizationRoleDTO{CapabilityID: capID, ComponentID: compID, Role: valueobjects.RealizationRoleStandard}
	queries := &mockRealizationRoleQueries{byPairReturns: []*readmodels.RealizationRoleDTO{existing, replaced}}
	h := setupRealizationRoleHandlers(&mockCommandBus{}, queries)
	r := realizationRoleRouter(h)

	body, _ := json.Marshal(AssignRealizationRoleRequest{Role: valueobjects.RealizationRoleStandard})
	req := httptest.NewRequest(http.MethodPut, "/capabilities/"+capID+"/components/"+compID+"/realization-role", bytes.NewReader(body))
	req = req.WithContext(sharedctx.WithActor(req.Context(), architectActor()))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Empty(t, rec.Header().Get("Location"))
}

func TestDeleteRealizationRole_DispatchesCommand_Returns204(t *testing.T) {
	capID, compID := uuid.New().String(), uuid.New().String()
	bus := &mockCommandBus{}
	h := setupRealizationRoleHandlers(bus, &mockRealizationRoleQueries{})
	r := realizationRoleRouter(h)

	req := httptest.NewRequest(http.MethodDelete, "/capabilities/"+capID+"/components/"+compID+"/realization-role", nil)
	req = req.WithContext(sharedctx.WithActor(req.Context(), architectActor()))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNoContent, rec.Code)
	require.Len(t, bus.dispatched, 1)
	cmd := bus.dispatched[0].(*commands.ClearRealizationRole)
	assert.Equal(t, capID, cmd.CapabilityID)
	assert.Equal(t, compID, cmd.ComponentID)
	assert.Equal(t, "user@example.com", cmd.ClearedBy)
}

func TestGetRealizationRoles_ParsesCommaSeparatedCapabilityIDs(t *testing.T) {
	capA, capB := uuid.New().String(), uuid.New().String()
	queries := &mockRealizationRoleQueries{}
	h := setupRealizationRoleHandlers(&mockCommandBus{}, queries)
	r := realizationRoleRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/realization-roles?capabilityIds="+capA+","+capB, nil)
	req = req.WithContext(sharedctx.WithActor(req.Context(), architectActor()))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, []string{capA, capB}, queries.receivedCapIDs)
}

func TestGetRealizationRoles_XAssignLinkVisibilityByActor(t *testing.T) {
	cases := []struct {
		name       string
		actor      sharedctx.Actor
		shouldShow bool
	}{
		{"architect sees x-assign", architectActor(), true},
		{"stakeholder does not see x-assign", stakeholderActor(), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			queries := &mockRealizationRoleQueries{}
			h := setupRealizationRoleHandlers(&mockCommandBus{}, queries)
			r := realizationRoleRouter(h)

			req := httptest.NewRequest(http.MethodGet, "/realization-roles?capabilityIds="+uuid.New().String(), nil)
			req = req.WithContext(sharedctx.WithActor(req.Context(), tc.actor))
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			require.Equal(t, http.StatusOK, rec.Code)
			var body sharedAPI.CollectionResponse
			require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
			if tc.shouldShow {
				assert.Contains(t, body.Links, "x-assign")
			} else {
				assert.NotContains(t, body.Links, "x-assign")
			}
		})
	}
}
