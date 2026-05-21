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
	"easi/backend/internal/shared/cqrs"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockCommandBus struct {
	dispatched    []cqrs.Command
	createdID     string
	err           error
	afterDispatch func()
}

func (m *mockCommandBus) Dispatch(_ context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	if m.err != nil {
		return cqrs.EmptyResult(), m.err
	}
	m.dispatched = append(m.dispatched, cmd)
	if m.afterDispatch != nil {
		m.afterDispatch()
	}
	return cqrs.CommandResult{CreatedID: m.createdID}, nil
}

func (m *mockCommandBus) Register(_ string, _ cqrs.CommandHandler) {}

type mockDirectionQueries struct {
	directionByID *readmodels.DirectionDTO
	activeByEC    *readmodels.DirectionDTO
}

func (m *mockDirectionQueries) GetByID(_ context.Context, _ readmodels.DirectionID) (*readmodels.DirectionDTO, error) {
	return m.directionByID, nil
}

func (m *mockDirectionQueries) GetActiveByEnterpriseCapabilityID(_ context.Context, _ string) (*readmodels.DirectionDTO, error) {
	return m.activeByEC, nil
}

func architectActor() sharedctx.Actor {
	return sharedctx.NewActor("u1", "user@example.com", sharedctx.RoleArchitect)
}

func stakeholderActor() sharedctx.Actor {
	return sharedctx.NewActor("u2", "stake@example.com", sharedctx.RoleStakeholder)
}

func setupHandlers(bus cqrs.CommandBus, queries DirectionQueries) *DirectionHandlers {
	links := NewDirectionLinks(sharedAPI.NewHATEOASLinks(""))
	return NewDirectionHandlers(bus, queries, links)
}

func getDirection(t *testing.T, queries DirectionQueries, ecID string, actor sharedctx.Actor) (int, struct {
	Direction *readmodels.DirectionDTO `json:"direction"`
	Links     sharedAPI.Links          `json:"_links"`
}) {
	t.Helper()
	h := setupHandlers(&mockCommandBus{}, queries)
	r := chi.NewRouter()
	r.Get("/enterprise-capabilities/{id}/direction", h.GetDirectionForEnterpriseCapability)
	req := httptest.NewRequest(http.MethodGet, "/enterprise-capabilities/"+ecID+"/direction", nil)
	req = req.WithContext(sharedctx.WithActor(req.Context(), actor))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	var body struct {
		Direction *readmodels.DirectionDTO `json:"direction"`
		Links     sharedAPI.Links          `json:"_links"`
	}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
	return rec.Code, body
}

func TestGetDirectionForEC_NoDirection_CaptureLinkVisibilityByActor(t *testing.T) {
	ecID := uuid.New().String()
	cases := []struct {
		name             string
		actor            sharedctx.Actor
		captureLinkShown bool
	}{
		{"architect sees capture link", architectActor(), true},
		{"stakeholder does not see capture link", stakeholderActor(), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			code, body := getDirection(t, &mockDirectionQueries{}, ecID, tc.actor)
			require.Equal(t, http.StatusOK, code)
			assert.Nil(t, body.Direction)
			if tc.captureLinkShown {
				assert.Contains(t, body.Links, "x-capture-direction")
			} else {
				assert.NotContains(t, body.Links, "x-capture-direction")
			}
		})
	}
}

func TestGetDirectionForEC_DirectionAffordancesByStatus(t *testing.T) {
	ecID := uuid.New().String()
	did := uuid.New().String()
	cases := []struct {
		name         string
		status       string
		expectLinks  []string
		forbidLinks  []string
		assertReason string
	}{
		{
			name:        "draft shows propose, reject, up; not agree",
			status:      "draft",
			expectLinks: []string{"x-propose", "x-reject", "up"},
			forbidLinks: []string{"x-agree"},
		},
		{
			name:         "agreed shows only reject; not edit/propose/agree",
			status:       "agreed",
			expectLinks:  []string{"x-reject"},
			forbidLinks:  []string{"edit", "x-propose", "x-agree"},
			assertReason: "spec allows reject-and-replace from agreed",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			queries := &mockDirectionQueries{activeByEC: &readmodels.DirectionDTO{
				ID: did, EnterpriseCapabilityID: ecID, Status: tc.status,
			}}
			code, body := getDirection(t, queries, ecID, architectActor())
			require.Equal(t, http.StatusOK, code)
			require.NotNil(t, body.Direction)
			for _, link := range tc.expectLinks {
				assert.Contains(t, body.Direction.Links, link, tc.assertReason)
			}
			for _, link := range tc.forbidLinks {
				assert.NotContains(t, body.Direction.Links, link)
			}
		})
	}
}

func TestCaptureDirection_DispatchesCommand(t *testing.T) {
	ecID := uuid.New().String()
	src1, src2 := uuid.New().String(), uuid.New().String()
	dom := uuid.New().String()
	did := uuid.New().String()
	bus := &mockCommandBus{createdID: did}
	queries := &mockDirectionQueries{activeByEC: &readmodels.DirectionDTO{
		ID: did, EnterpriseCapabilityID: ecID, Status: "draft",
	}}
	h := setupHandlers(bus, queries)

	r := chi.NewRouter()
	r.Post("/enterprise-capabilities/{id}/direction", h.CaptureDirection)

	reqBody, _ := json.Marshal(CaptureDirectionRequest{
		Type:                "consolidate",
		SourceCapabilityIDs: []string{src1, src2},
		Placements:          []PlacementRequest{{TargetBusinessDomainID: dom}},
		Horizon:             "next",
		Narrative:           "consolidating",
	})
	req := httptest.NewRequest(http.MethodPost, "/enterprise-capabilities/"+ecID+"/direction", bytes.NewReader(reqBody))
	req = req.WithContext(sharedctx.WithActor(req.Context(), architectActor()))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)
	require.Len(t, bus.dispatched, 1)
	cmd := bus.dispatched[0].(*commands.CaptureDirection)
	assert.Equal(t, ecID, cmd.EnterpriseCapabilityID)
	assert.Equal(t, "consolidate", cmd.Type)
	assert.Len(t, cmd.SourceCapabilityIDs, 2)
}

func TestAdvanceDirection_DispatchesCommandWithTargetStatus(t *testing.T) {
	cases := []struct {
		name         string
		subPath      string
		fromStatus   string
		targetStatus string
		bind         func(h *DirectionHandlers) http.HandlerFunc
	}{
		{
			name: "propose moves draft to proposed", subPath: "propose",
			fromStatus: "draft", targetStatus: "proposed",
			bind: func(h *DirectionHandlers) http.HandlerFunc { return h.ProposeDirection },
		},
		{
			name: "agree moves proposed to agreed", subPath: "agree",
			fromStatus: "proposed", targetStatus: "agreed",
			bind: func(h *DirectionHandlers) http.HandlerFunc { return h.AgreeDirection },
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ecID := uuid.New().String()
			did := uuid.New().String()
			bus := &mockCommandBus{}
			queries := &mockDirectionQueries{activeByEC: &readmodels.DirectionDTO{
				ID: did, EnterpriseCapabilityID: ecID, Status: tc.fromStatus,
			}}
			h := setupHandlers(bus, queries)

			r := chi.NewRouter()
			r.Post("/enterprise-capabilities/{id}/direction/"+tc.subPath, tc.bind(h))

			req := httptest.NewRequest(http.MethodPost, "/enterprise-capabilities/"+ecID+"/direction/"+tc.subPath, nil)
			req = req.WithContext(sharedctx.WithActor(req.Context(), architectActor()))
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			require.Equal(t, http.StatusOK, rec.Code)
			require.Len(t, bus.dispatched, 1)
			cmd := bus.dispatched[0].(*commands.AdvanceDirection)
			assert.Equal(t, did, cmd.DirectionID)
			assert.Equal(t, tc.targetStatus, cmd.TargetStatus)
		})
	}
}

func TestRejectDirection_DispatchesCommandAndReturnsRejectedDirection(t *testing.T) {
	ecID := uuid.New().String()
	did := uuid.New().String()
	bus := &mockCommandBus{}
	queries := &mockDirectionQueries{
		activeByEC: &readmodels.DirectionDTO{
			ID: did, EnterpriseCapabilityID: ecID, Status: "agreed",
		},
		directionByID: &readmodels.DirectionDTO{
			ID: did, EnterpriseCapabilityID: ecID, Status: "rejected",
		},
	}
	bus.afterDispatch = func() {
		queries.activeByEC = nil
	}
	h := setupHandlers(bus, queries)

	r := chi.NewRouter()
	r.Post("/enterprise-capabilities/{id}/direction/reject", h.RejectDirection)

	req := httptest.NewRequest(http.MethodPost, "/enterprise-capabilities/"+ecID+"/direction/reject", nil)
	req = req.WithContext(sharedctx.WithActor(req.Context(), architectActor()))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code, "rejection must succeed even after projection removes the active direction")
	require.Len(t, bus.dispatched, 1)
	cmd := bus.dispatched[0].(*commands.RejectDirection)
	assert.Equal(t, did, cmd.DirectionID)

	var body readmodels.DirectionDTO
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
	assert.Equal(t, did, body.ID)
	assert.Equal(t, "rejected", body.Status)
}

func TestUpdateDirection_DispatchesSingleAtomicCommand(t *testing.T) {
	ecID := uuid.New().String()
	did := uuid.New().String()
	bus := &mockCommandBus{}
	queries := &mockDirectionQueries{activeByEC: &readmodels.DirectionDTO{
		ID: did, EnterpriseCapabilityID: ecID, Status: "draft",
	}}
	h := setupHandlers(bus, queries)

	r := chi.NewRouter()
	r.Put("/enterprise-capabilities/{id}/direction", h.UpdateDirection)

	narrative := "Refined"
	horizon := "later"
	reqBody, _ := json.Marshal(UpdateDirectionRequest{
		Narrative: &narrative,
		Horizon:   &horizon,
	})
	req := httptest.NewRequest(http.MethodPut, "/enterprise-capabilities/"+ecID+"/direction", bytes.NewReader(reqBody))
	req = req.WithContext(sharedctx.WithActor(req.Context(), architectActor()))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Len(t, bus.dispatched, 1, "multi-field update dispatches one atomic command")
	cmd, ok := bus.dispatched[0].(*commands.UpdateDirection)
	require.True(t, ok)
	assert.Equal(t, did, cmd.DirectionID)
	require.NotNil(t, cmd.Narrative)
	assert.Equal(t, "Refined", *cmd.Narrative)
	require.NotNil(t, cmd.Horizon)
	assert.Equal(t, "later", *cmd.Horizon)
	assert.Nil(t, cmd.SourceCapabilityIDs)
	assert.Nil(t, cmd.Placements)
}

func TestRejectDirection_NoActiveDirection_404(t *testing.T) {
	ecID := uuid.New().String()
	bus := &mockCommandBus{}
	queries := &mockDirectionQueries{} // no active
	h := setupHandlers(bus, queries)

	r := chi.NewRouter()
	r.Post("/enterprise-capabilities/{id}/direction/reject", h.RejectDirection)

	req := httptest.NewRequest(http.MethodPost, "/enterprise-capabilities/"+ecID+"/direction/reject", nil)
	req = req.WithContext(sharedctx.WithActor(req.Context(), architectActor()))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Empty(t, bus.dispatched)
}
