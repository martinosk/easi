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
	dispatched []cqrs.Command
	createdID  string
	err        error
}

func (m *mockCommandBus) Dispatch(_ context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	if m.err != nil {
		return cqrs.EmptyResult(), m.err
	}
	m.dispatched = append(m.dispatched, cmd)
	return cqrs.CommandResult{CreatedID: m.createdID}, nil
}

func (m *mockCommandBus) Register(_ string, _ cqrs.CommandHandler) {}

type mockDirectionQueries struct {
	directionByID *readmodels.DirectionDTO
	activeByEC    *readmodels.DirectionDTO
}

func (m *mockDirectionQueries) GetByID(_ context.Context, _ string) (*readmodels.DirectionDTO, error) {
	return m.directionByID, nil
}

func (m *mockDirectionQueries) GetActiveByEnterpriseCapabilityID(_ context.Context, _ string) (*readmodels.DirectionDTO, error) {
	return m.activeByEC, nil
}

type mockSessionProvider struct{}

func (m mockSessionProvider) GetCurrentUserEmail(_ context.Context) (string, error) {
	return "user@example.com", nil
}

func architectActor() sharedctx.Actor {
	return sharedctx.NewActor("u1", "user@example.com", sharedctx.RoleArchitect)
}

func stakeholderActor() sharedctx.Actor {
	return sharedctx.NewActor("u2", "stake@example.com", sharedctx.RoleStakeholder)
}

func setupHandlers(bus cqrs.CommandBus, queries DirectionQueries) *DirectionHandlers {
	links := NewDirectionLinks(sharedAPI.NewHATEOASLinks(""))
	return NewDirectionHandlers(bus, queries, mockSessionProvider{}, links)
}

func TestGetDirectionForEC_NoDirection_ReturnsNullWithCaptureLink(t *testing.T) {
	ecID := uuid.New().String()
	queries := &mockDirectionQueries{}
	h := setupHandlers(&mockCommandBus{}, queries)

	r := chi.NewRouter()
	r.Get("/enterprise-capabilities/{id}/direction", h.GetDirectionForEnterpriseCapability)

	req := httptest.NewRequest(http.MethodGet, "/enterprise-capabilities/"+ecID+"/direction", nil)
	req = req.WithContext(sharedctx.WithActor(req.Context(), architectActor()))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var body struct {
		Direction *readmodels.DirectionDTO `json:"direction"`
		Links     sharedAPI.Links          `json:"_links"`
	}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
	assert.Nil(t, body.Direction)
	assert.Contains(t, body.Links, "x-capture-direction")
}

func TestGetDirectionForEC_NoDirection_ReadOnly_NoCaptureLink(t *testing.T) {
	ecID := uuid.New().String()
	queries := &mockDirectionQueries{}
	h := setupHandlers(&mockCommandBus{}, queries)

	r := chi.NewRouter()
	r.Get("/enterprise-capabilities/{id}/direction", h.GetDirectionForEnterpriseCapability)

	req := httptest.NewRequest(http.MethodGet, "/enterprise-capabilities/"+ecID+"/direction", nil)
	req = req.WithContext(sharedctx.WithActor(req.Context(), stakeholderActor()))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var body struct {
		Direction *readmodels.DirectionDTO `json:"direction"`
		Links     sharedAPI.Links          `json:"_links"`
	}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
	assert.NotContains(t, body.Links, "x-capture-direction")
}

func TestGetDirectionForEC_WithDraftDirection_ShowsAdvanceAffordances(t *testing.T) {
	ecID := uuid.New().String()
	did := uuid.New().String()
	queries := &mockDirectionQueries{
		activeByEC: &readmodels.DirectionDTO{
			ID:                     did,
			EnterpriseCapabilityID: ecID,
			Status:                 "draft",
		},
	}
	h := setupHandlers(&mockCommandBus{}, queries)

	r := chi.NewRouter()
	r.Get("/enterprise-capabilities/{id}/direction", h.GetDirectionForEnterpriseCapability)

	req := httptest.NewRequest(http.MethodGet, "/enterprise-capabilities/"+ecID+"/direction", nil)
	req = req.WithContext(sharedctx.WithActor(req.Context(), architectActor()))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var body struct {
		Direction *readmodels.DirectionDTO `json:"direction"`
	}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
	require.NotNil(t, body.Direction)
	assert.Contains(t, body.Direction.Links, "x-advance-proposed")
	assert.Contains(t, body.Direction.Links, "x-reject")
	assert.NotContains(t, body.Direction.Links, "x-advance-agreed")
}

func TestGetDirectionForEC_AgreedDirection_NoEditAffordances(t *testing.T) {
	ecID := uuid.New().String()
	did := uuid.New().String()
	queries := &mockDirectionQueries{
		activeByEC: &readmodels.DirectionDTO{
			ID:                     did,
			EnterpriseCapabilityID: ecID,
			Status:                 "agreed",
		},
	}
	h := setupHandlers(&mockCommandBus{}, queries)

	r := chi.NewRouter()
	r.Get("/enterprise-capabilities/{id}/direction", h.GetDirectionForEnterpriseCapability)

	req := httptest.NewRequest(http.MethodGet, "/enterprise-capabilities/"+ecID+"/direction", nil)
	req = req.WithContext(sharedctx.WithActor(req.Context(), architectActor()))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var body struct {
		Direction *readmodels.DirectionDTO `json:"direction"`
	}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
	require.NotNil(t, body.Direction)
	assert.NotContains(t, body.Direction.Links, "edit")
	assert.NotContains(t, body.Direction.Links, "x-advance-agreed")
	assert.NotContains(t, body.Direction.Links, "x-reject")
}

func TestCaptureDirection_DispatchesCommand(t *testing.T) {
	ecID := uuid.New().String()
	src1, src2 := uuid.New().String(), uuid.New().String()
	dom := uuid.New().String()
	did := uuid.New().String()
	bus := &mockCommandBus{createdID: did}
	queries := &mockDirectionQueries{directionByID: &readmodels.DirectionDTO{
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

func TestAdvanceDirection_DispatchesCommand(t *testing.T) {
	did := uuid.New().String()
	bus := &mockCommandBus{}
	queries := &mockDirectionQueries{directionByID: &readmodels.DirectionDTO{
		ID: did, EnterpriseCapabilityID: uuid.New().String(), Status: "proposed",
	}}
	h := setupHandlers(bus, queries)

	r := chi.NewRouter()
	r.Post("/directions/{id}/advance/{target}", h.AdvanceDirection)

	req := httptest.NewRequest(http.MethodPost, "/directions/"+did+"/advance/proposed", nil)
	req = req.WithContext(sharedctx.WithActor(req.Context(), architectActor()))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Len(t, bus.dispatched, 1)
	cmd := bus.dispatched[0].(*commands.AdvanceDirection)
	assert.Equal(t, did, cmd.DirectionID)
	assert.Equal(t, "proposed", cmd.TargetStatus)
}

func TestRejectDirection_DispatchesCommand(t *testing.T) {
	did := uuid.New().String()
	bus := &mockCommandBus{}
	queries := &mockDirectionQueries{directionByID: &readmodels.DirectionDTO{
		ID: did, EnterpriseCapabilityID: uuid.New().String(), Status: "rejected",
	}}
	h := setupHandlers(bus, queries)

	r := chi.NewRouter()
	r.Post("/directions/{id}/reject", h.RejectDirection)

	req := httptest.NewRequest(http.MethodPost, "/directions/"+did+"/reject", nil)
	req = req.WithContext(sharedctx.WithActor(req.Context(), architectActor()))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Len(t, bus.dispatched, 1)
	_, ok := bus.dispatched[0].(*commands.RejectDirection)
	assert.True(t, ok)
}

func TestUpdateDirection_DispatchesNarrativeAndHorizon(t *testing.T) {
	did := uuid.New().String()
	bus := &mockCommandBus{}
	queries := &mockDirectionQueries{directionByID: &readmodels.DirectionDTO{
		ID: did, EnterpriseCapabilityID: uuid.New().String(), Status: "draft",
	}}
	h := setupHandlers(bus, queries)

	r := chi.NewRouter()
	r.Put("/directions/{id}", h.UpdateDirection)

	narrative := "Refined"
	horizon := "later"
	reqBody, _ := json.Marshal(UpdateDirectionRequest{
		Narrative: &narrative,
		Horizon:   &horizon,
	})
	req := httptest.NewRequest(http.MethodPut, "/directions/"+did, bytes.NewReader(reqBody))
	req = req.WithContext(sharedctx.WithActor(req.Context(), architectActor()))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Len(t, bus.dispatched, 2)
}
