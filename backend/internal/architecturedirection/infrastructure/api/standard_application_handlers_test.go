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

type mockStandardApplicationQueries struct {
	currentReturns []*readmodels.StandardApplicationDTO
	history        *readmodels.StandardApplicationHistoryDTO
	aggregateID    string
	hasAggregate   bool
	callIdx        int
}

func (m *mockStandardApplicationQueries) GetCurrentByEnterpriseCapability(_ context.Context, _ string) (*readmodels.StandardApplicationDTO, error) {
	var result *readmodels.StandardApplicationDTO
	if m.callIdx < len(m.currentReturns) {
		result = m.currentReturns[m.callIdx]
	}
	m.callIdx++
	return result, nil
}

func (m *mockStandardApplicationQueries) GetHistoryByAggregateID(_ context.Context, _ string) (*readmodels.StandardApplicationHistoryDTO, error) {
	return m.history, nil
}

func (m *mockStandardApplicationQueries) FindAggregateIDForEnterpriseCapability(_ context.Context, _ string) (string, bool, error) {
	return m.aggregateID, m.hasAggregate, nil
}

func setupStandardApplicationHandlers(bus *mockCommandBus, queries StandardApplicationQueries) *StandardApplicationHandlers {
	links := NewStandardApplicationLinks(sharedAPI.NewHATEOASLinks(""))
	return NewStandardApplicationHandlers(bus, queries, links)
}

func getStandard(t *testing.T, queries StandardApplicationQueries, ecID string, actor sharedctx.Actor) (int, ECStandardApplicationResponse) {
	t.Helper()
	h := setupStandardApplicationHandlers(&mockCommandBus{}, queries)
	r := chi.NewRouter()
	r.Get("/enterprise-capabilities/{id}/standard-application", h.GetStandardApplicationForEnterpriseCapability)
	req := httptest.NewRequest(http.MethodGet, "/enterprise-capabilities/"+ecID+"/standard-application", nil)
	req = req.WithContext(sharedctx.WithActor(req.Context(), actor))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	var body ECStandardApplicationResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
	return rec.Code, body
}

func putStandard(t *testing.T, queries StandardApplicationQueries, ecID string, payload SetStandardApplicationRequest) (int, *readmodels.StandardApplicationDTO, string) {
	t.Helper()
	bus := &mockCommandBus{}
	h := setupStandardApplicationHandlers(bus, queries)
	r := chi.NewRouter()
	r.Put("/enterprise-capabilities/{id}/standard-application", h.SetStandardApplication)
	reqBody, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPut, "/enterprise-capabilities/"+ecID+"/standard-application", bytes.NewReader(reqBody))
	req = req.WithContext(sharedctx.WithActor(req.Context(), architectActor()))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	body := &readmodels.StandardApplicationDTO{}
	if rec.Body.Len() > 0 {
		require.NoError(t, json.NewDecoder(rec.Body).Decode(body))
	}
	return rec.Code, body, rec.Header().Get("Location")
}

func TestGetStandardApplication_NoStandard_SetLinkVisibilityByActor(t *testing.T) {
	ecID := uuid.New().String()
	cases := []struct {
		name       string
		actor      sharedctx.Actor
		shouldShow bool
	}{
		{"architect sees x-set-standard", architectActor(), true},
		{"stakeholder does not see x-set-standard", stakeholderActor(), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			code, body := getStandard(t, &mockStandardApplicationQueries{}, ecID, tc.actor)
			require.Equal(t, http.StatusOK, code)
			assert.Nil(t, body.Standard)
			if tc.shouldShow {
				assert.Contains(t, body.Links, "x-set-standard")
				assert.NotContains(t, body.Links, "edit")
				assert.Equal(t, http.MethodPut, body.Links["x-set-standard"].Method)
			} else {
				assert.NotContains(t, body.Links, "x-set-standard")
				assert.NotContains(t, body.Links, "edit")
			}
			assert.Contains(t, body.Links, "x-history")
			assert.Contains(t, body.Links, "self")
		})
	}
}

func TestGetStandardApplication_StandardExists_EditLinkVisibilityByActor(t *testing.T) {
	ecID := uuid.New().String()
	current := &readmodels.StandardApplicationDTO{
		ID:                     ecID,
		EnterpriseCapabilityID: ecID,
		ApplicationID:          uuid.New().String(),
		Narrative:              "the standard",
	}
	cases := []struct {
		name       string
		actor      sharedctx.Actor
		shouldShow bool
	}{
		{"architect sees edit", architectActor(), true},
		{"stakeholder does not see edit", stakeholderActor(), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			code, body := getStandard(t, &mockStandardApplicationQueries{currentReturns: []*readmodels.StandardApplicationDTO{current}}, ecID, tc.actor)
			require.Equal(t, http.StatusOK, code)
			require.NotNil(t, body.Standard)
			if tc.shouldShow {
				assert.Contains(t, body.Standard.Links, "edit")
				assert.Equal(t, http.MethodPut, body.Standard.Links["edit"].Method)
				assert.Contains(t, body.Links, "edit")
			} else {
				assert.NotContains(t, body.Standard.Links, "edit")
				assert.NotContains(t, body.Links, "edit")
			}
		})
	}
}

func TestSetStandardApplication_FirstSet_Returns201WithLocation(t *testing.T) {
	ecID := uuid.New().String()
	appID := uuid.New().String()
	queries := &mockStandardApplicationQueries{
		currentReturns: []*readmodels.StandardApplicationDTO{
			nil,
			{ID: ecID, EnterpriseCapabilityID: ecID, ApplicationID: appID, Narrative: "first"},
		},
	}

	code, body, location := putStandard(t, queries, ecID, SetStandardApplicationRequest{
		ApplicationID: appID,
		Narrative:     "first",
	})

	require.Equal(t, http.StatusCreated, code)
	assert.Equal(t, "/api/v1/enterprise-capabilities/"+ecID+"/standard-application", location)
	require.NotNil(t, body.Links["self"], "PUT response must carry HATEOAS self link")
	assert.Equal(t, http.MethodGet, body.Links["self"].Method)
	assert.Contains(t, body.Links, "edit")
}

func TestSetStandardApplication_Replacement_Returns200(t *testing.T) {
	ecID := uuid.New().String()
	appID := uuid.New().String()
	previous := &readmodels.StandardApplicationDTO{
		ID: ecID, EnterpriseCapabilityID: ecID, ApplicationID: uuid.New().String(), Narrative: "previous",
	}
	queries := &mockStandardApplicationQueries{
		currentReturns: []*readmodels.StandardApplicationDTO{
			previous,
			{ID: ecID, EnterpriseCapabilityID: ecID, ApplicationID: appID, Narrative: "replacement"},
		},
	}

	code, body, location := putStandard(t, queries, ecID, SetStandardApplicationRequest{
		ApplicationID: appID,
		Narrative:     "replacement",
	})

	require.Equal(t, http.StatusOK, code)
	assert.Empty(t, location, "replacement must not emit a Location header")
	require.NotNil(t, body.Links["self"])
}

func TestSetStandardApplication_DispatchesCommandWithRequestBody(t *testing.T) {
	ecID := uuid.New().String()
	appID := uuid.New().String()
	bus := &mockCommandBus{}
	queries := &mockStandardApplicationQueries{
		currentReturns: []*readmodels.StandardApplicationDTO{
			nil,
			{ID: ecID, EnterpriseCapabilityID: ecID, ApplicationID: appID, Narrative: "n"},
		},
	}
	h := setupStandardApplicationHandlers(bus, queries)

	r := chi.NewRouter()
	r.Put("/enterprise-capabilities/{id}/standard-application", h.SetStandardApplication)

	reqBody, _ := json.Marshal(SetStandardApplicationRequest{
		ApplicationID: appID,
		Narrative:     "covers operational and reporting layers",
	})
	req := httptest.NewRequest(http.MethodPut, "/enterprise-capabilities/"+ecID+"/standard-application", bytes.NewReader(reqBody))
	req = req.WithContext(sharedctx.WithActor(req.Context(), architectActor()))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Len(t, bus.dispatched, 1)
	cmd := bus.dispatched[0].(*commands.SetStandardApplication)
	assert.Equal(t, ecID, cmd.EnterpriseCapabilityID)
	assert.Equal(t, appID, cmd.ApplicationID)
	assert.Equal(t, "covers operational and reporting layers", cmd.Narrative)
}

func TestGetStandardApplicationHistory_NoStandardEverSet_Returns200WithEmptyEntries(t *testing.T) {
	ecID := uuid.New().String()
	queries := &mockStandardApplicationQueries{hasAggregate: false}
	h := setupStandardApplicationHandlers(&mockCommandBus{}, queries)

	r := chi.NewRouter()
	r.Get("/enterprise-capabilities/{id}/standard-application/history", h.GetStandardApplicationHistory)

	req := httptest.NewRequest(http.MethodGet, "/enterprise-capabilities/"+ecID+"/standard-application/history", nil)
	req = req.WithContext(sharedctx.WithActor(req.Context(), architectActor()))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var body readmodels.StandardApplicationHistoryDTO
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
	assert.Empty(t, body.Entries, "history must be present-but-empty when no standard has been set")
	assert.Contains(t, body.Links, "self")
	assert.Contains(t, body.Links, "up")
}

func TestGetStandardApplicationHistory_Returns200WithEntries(t *testing.T) {
	ecID := uuid.New().String()
	queries := &mockStandardApplicationQueries{
		aggregateID:  ecID,
		hasAggregate: true,
		history: &readmodels.StandardApplicationHistoryDTO{
			StandardApplicationID:  ecID,
			EnterpriseCapabilityID: ecID,
			Entries: []readmodels.StandardApplicationHistoryEntryDTO{
				{ApplicationID: uuid.New().String(), Narrative: "second"},
				{ApplicationID: uuid.New().String(), Narrative: "first"},
			},
		},
	}
	h := setupStandardApplicationHandlers(&mockCommandBus{}, queries)

	r := chi.NewRouter()
	r.Get("/enterprise-capabilities/{id}/standard-application/history", h.GetStandardApplicationHistory)

	req := httptest.NewRequest(http.MethodGet, "/enterprise-capabilities/"+ecID+"/standard-application/history", nil)
	req = req.WithContext(sharedctx.WithActor(req.Context(), architectActor()))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var body readmodels.StandardApplicationHistoryDTO
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
	assert.Len(t, body.Entries, 2)
	assert.Contains(t, body.Links, "self")
	assert.Contains(t, body.Links, "up")
}
