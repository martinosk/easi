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
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockTimeAssessmentQueries struct {
	byPairReturns   []*readmodels.TimeAssessmentDTO
	byPairCallIdx   int
	byCapability    []readmodels.TimeAssessmentDTO
	rollups         []readmodels.TimeAssessmentRollupDTO
	receivedCapIDs  []string
	receivedCompIDs []string
}

func (m *mockTimeAssessmentQueries) GetByPair(_ context.Context, _, _ string) (*readmodels.TimeAssessmentDTO, error) {
	var result *readmodels.TimeAssessmentDTO
	if m.byPairCallIdx < len(m.byPairReturns) {
		result = m.byPairReturns[m.byPairCallIdx]
	}
	m.byPairCallIdx++
	return result, nil
}

func (m *mockTimeAssessmentQueries) GetByCapabilityIDs(_ context.Context, ids []string) ([]readmodels.TimeAssessmentDTO, error) {
	m.receivedCapIDs = ids
	return m.byCapability, nil
}

func (m *mockTimeAssessmentQueries) GetRollupsByComponentIDs(_ context.Context, ids []string) ([]readmodels.TimeAssessmentRollupDTO, error) {
	m.receivedCompIDs = ids
	return m.rollups, nil
}

func setupTimeAssessmentHandlers(bus *mockCommandBus, queries TimeAssessmentQueries) *TimeAssessmentHandlers {
	links := NewTimeAssessmentLinks(sharedAPI.NewHATEOASLinks(""))
	return NewTimeAssessmentHandlers(bus, queries, links)
}

func timeAssessmentRouter(h *TimeAssessmentHandlers) chi.Router {
	r := chi.NewRouter()
	r.Get("/capabilities/{id}/components/{componentId}/time-assessment", h.GetTimeAssessment)
	r.Put("/capabilities/{id}/components/{componentId}/time-assessment", h.PutTimeAssessment)
	r.Delete("/capabilities/{id}/components/{componentId}/time-assessment", h.DeleteTimeAssessment)
	r.Get("/time-assessments", h.GetTimeAssessments)
	r.Get("/time-assessments/rollups", h.GetTimeAssessmentRollups)
	return r
}

func TestGetTimeAssessment_Unassessed_Returns404(t *testing.T) {
	queries := &mockTimeAssessmentQueries{}
	h := setupTimeAssessmentHandlers(&mockCommandBus{}, queries)
	r := timeAssessmentRouter(h)

	capID, compID := uuid.New().String(), uuid.New().String()
	req := httptest.NewRequest(http.MethodGet, "/capabilities/"+capID+"/components/"+compID+"/time-assessment", nil)
	req = req.WithContext(sharedctx.WithActor(req.Context(), architectActor()))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestGetTimeAssessment_Assessed_EditDeleteLinkVisibilityByActor(t *testing.T) {
	capID, compID := uuid.New().String(), uuid.New().String()
	current := &readmodels.TimeAssessmentDTO{ID: uuid.New().String(), CapabilityID: capID, ComponentID: compID, Grade: "Migrate"}
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
			queries := &mockTimeAssessmentQueries{byPairReturns: []*readmodels.TimeAssessmentDTO{current}}
			h := setupTimeAssessmentHandlers(&mockCommandBus{}, queries)
			r := timeAssessmentRouter(h)

			req := httptest.NewRequest(http.MethodGet, "/capabilities/"+capID+"/components/"+compID+"/time-assessment", nil)
			req = req.WithContext(sharedctx.WithActor(req.Context(), tc.actor))
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			require.Equal(t, http.StatusOK, rec.Code)
			var body readmodels.TimeAssessmentDTO
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

func TestPutTimeAssessment_FirstAssessment_Returns201WithLocation(t *testing.T) {
	capID, compID := uuid.New().String(), uuid.New().String()
	created := &readmodels.TimeAssessmentDTO{ID: uuid.New().String(), CapabilityID: capID, ComponentID: compID, Grade: "Migrate"}
	queries := &mockTimeAssessmentQueries{byPairReturns: []*readmodels.TimeAssessmentDTO{nil, created}}
	bus := &mockCommandBus{}
	h := setupTimeAssessmentHandlers(bus, queries)
	r := timeAssessmentRouter(h)

	body, _ := json.Marshal(AssessRealizationRequest{Grade: "Migrate", Rationale: "carve-out candidate"})
	req := httptest.NewRequest(http.MethodPut, "/capabilities/"+capID+"/components/"+compID+"/time-assessment", bytes.NewReader(body))
	req = req.WithContext(sharedctx.WithActor(req.Context(), architectActor()))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	assert.Equal(t, "/api/v1/capabilities/"+capID+"/components/"+compID+"/time-assessment", rec.Header().Get("Location"))
	require.Len(t, bus.dispatched, 1)
	cmd := bus.dispatched[0].(*commands.AssessRealization)
	assert.Equal(t, capID, cmd.CapabilityID)
	assert.Equal(t, compID, cmd.ComponentID)
	assert.Equal(t, "Migrate", cmd.Grade)
	assert.Equal(t, "carve-out candidate", cmd.Rationale)
	assert.Equal(t, "user@example.com", cmd.AssessedBy)
}

func TestPutTimeAssessment_Reassessment_Returns200(t *testing.T) {
	capID, compID := uuid.New().String(), uuid.New().String()
	existing := &readmodels.TimeAssessmentDTO{ID: uuid.New().String(), CapabilityID: capID, ComponentID: compID, Grade: "Tolerate"}
	replaced := &readmodels.TimeAssessmentDTO{ID: existing.ID, CapabilityID: capID, ComponentID: compID, Grade: "Eliminate"}
	queries := &mockTimeAssessmentQueries{byPairReturns: []*readmodels.TimeAssessmentDTO{existing, replaced}}
	h := setupTimeAssessmentHandlers(&mockCommandBus{}, queries)
	r := timeAssessmentRouter(h)

	body, _ := json.Marshal(AssessRealizationRequest{Grade: "Eliminate"})
	req := httptest.NewRequest(http.MethodPut, "/capabilities/"+capID+"/components/"+compID+"/time-assessment", bytes.NewReader(body))
	req = req.WithContext(sharedctx.WithActor(req.Context(), architectActor()))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Empty(t, rec.Header().Get("Location"))
}

func TestDeleteTimeAssessment_DispatchesCommand_Returns204(t *testing.T) {
	capID, compID := uuid.New().String(), uuid.New().String()
	bus := &mockCommandBus{}
	h := setupTimeAssessmentHandlers(bus, &mockTimeAssessmentQueries{})
	r := timeAssessmentRouter(h)

	req := httptest.NewRequest(http.MethodDelete, "/capabilities/"+capID+"/components/"+compID+"/time-assessment", nil)
	req = req.WithContext(sharedctx.WithActor(req.Context(), architectActor()))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNoContent, rec.Code)
	require.Len(t, bus.dispatched, 1)
	cmd := bus.dispatched[0].(*commands.RemoveTimeAssessment)
	assert.Equal(t, capID, cmd.CapabilityID)
	assert.Equal(t, compID, cmd.ComponentID)
	assert.Equal(t, "user@example.com", cmd.RemovedBy)
}

func TestGetTimeAssessments_ParsesCommaSeparatedCapabilityIDs(t *testing.T) {
	capA, capB := uuid.New().String(), uuid.New().String()
	queries := &mockTimeAssessmentQueries{}
	h := setupTimeAssessmentHandlers(&mockCommandBus{}, queries)
	r := timeAssessmentRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/time-assessments?capabilityIds="+capA+","+capB, nil)
	req = req.WithContext(sharedctx.WithActor(req.Context(), architectActor()))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, []string{capA, capB}, queries.receivedCapIDs)
}

func TestGetTimeAssessments_XAssessLinkVisibilityByActor(t *testing.T) {
	cases := []struct {
		name       string
		actor      sharedctx.Actor
		shouldShow bool
	}{
		{"architect sees x-assess", architectActor(), true},
		{"stakeholder does not see x-assess", stakeholderActor(), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			queries := &mockTimeAssessmentQueries{}
			h := setupTimeAssessmentHandlers(&mockCommandBus{}, queries)
			r := timeAssessmentRouter(h)

			req := httptest.NewRequest(http.MethodGet, "/time-assessments?capabilityIds="+uuid.New().String(), nil)
			req = req.WithContext(sharedctx.WithActor(req.Context(), tc.actor))
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			require.Equal(t, http.StatusOK, rec.Code)
			var body sharedAPI.CollectionResponse
			require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
			if tc.shouldShow {
				assert.Contains(t, body.Links, "x-assess")
			} else {
				assert.NotContains(t, body.Links, "x-assess")
			}
		})
	}
}

func TestGetTimeAssessmentRollups_ParsesCommaSeparatedComponentIDs(t *testing.T) {
	compA, compB := uuid.New().String(), uuid.New().String()
	queries := &mockTimeAssessmentQueries{rollups: []readmodels.TimeAssessmentRollupDTO{
		{ComponentID: compA, Counts: readmodels.TimeGradeCounts{Invest: 1, Migrate: 1}},
	}}
	h := setupTimeAssessmentHandlers(&mockCommandBus{}, queries)
	r := timeAssessmentRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/time-assessments/rollups?componentIds="+compA+","+compB, nil)
	req = req.WithContext(sharedctx.WithActor(req.Context(), architectActor()))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, []string{compA, compB}, queries.receivedCompIDs)
	var body sharedAPI.CollectionResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
	assert.Contains(t, body.Links, "self")
}
