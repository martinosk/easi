package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"easi/backend/internal/architecturedirection/application/commands"
	"easi/backend/internal/architecturedirection/application/readmodels"
	"easi/backend/internal/architecturedirection/domain/aggregates"
	"easi/backend/internal/architecturedirection/domain/services"
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newSourceRouter(h *DirectionHandlers) *chi.Mux {
	r := chi.NewRouter()
	r.Post("/enterprise-capabilities/{id}/direction/sources", h.AddDirectionSource)
	r.Delete("/enterprise-capabilities/{id}/direction/sources/{capabilityId}", h.RemoveDirectionSource)
	return r
}

func postJSON(t *testing.T, r http.Handler, target string, body any) *httptest.ResponseRecorder {
	t.Helper()
	payload, err := json.Marshal(body)
	require.NoError(t, err)
	return performAsArchitect(r, httptest.NewRequest(http.MethodPost, target, bytes.NewReader(payload)))
}

func performWithoutBody(r http.Handler, method, target string) *httptest.ResponseRecorder {
	return performAsArchitect(r, httptest.NewRequest(method, target, nil))
}

func performAsArchitect(r http.Handler, req *http.Request) *httptest.ResponseRecorder {
	req = req.WithContext(sharedctx.WithActor(req.Context(), architectActor()))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func decodeErrorBody(t *testing.T, rec *httptest.ResponseRecorder) sharedAPI.ErrorResponse {
	t.Helper()
	var body sharedAPI.ErrorResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
	return body
}

func TestAddDirectionSource_DispatchesCommandWithActor(t *testing.T) {
	ecID := uuid.New().String()
	did := uuid.New().String()
	capID := uuid.New().String()
	bus := &mockCommandBus{}
	queries := &mockDirectionQueries{activeByEC: &readmodels.DirectionDTO{
		ID: did, EnterpriseCapabilityID: ecID, Status: "draft",
	}}
	h := setupHandlers(bus, queries)

	rec := postJSON(t, newSourceRouter(h),
		"/enterprise-capabilities/"+ecID+"/direction/sources", AddDirectionSourceRequest{CapabilityID: capID})

	require.Equal(t, http.StatusOK, rec.Code, "adding a source returns the existing direction resource, not 201")
	require.Len(t, bus.dispatched, 1)
	cmd, ok := bus.dispatched[0].(*commands.AddDirectionSource)
	require.True(t, ok)
	assert.Equal(t, did, cmd.DirectionID)
	assert.Equal(t, capID, cmd.CapabilityID)
	assert.Equal(t, "user@example.com", cmd.Actor, "the acting architect is recorded (R9)")

	var direction readmodels.DirectionDTO
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&direction))
	assert.Equal(t, did, direction.ID)
}

func TestAddDirectionSource_NoActiveDirection_404(t *testing.T) {
	h := setupHandlers(&mockCommandBus{}, &mockDirectionQueries{})

	rec := postJSON(t, newSourceRouter(h),
		"/enterprise-capabilities/"+uuid.New().String()+"/direction/sources",
		AddDirectionSourceRequest{CapabilityID: uuid.New().String()})

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestAddDirectionSource_AgreedDirection_SourceSetFrozenBody(t *testing.T) {
	ecID := uuid.New().String()
	bus := &mockCommandBus{err: aggregates.ErrDirectionSourceSetFrozen}
	queries := &mockDirectionQueries{activeByEC: &readmodels.DirectionDTO{
		ID: uuid.New().String(), EnterpriseCapabilityID: ecID, Status: "agreed",
	}}
	h := setupHandlers(bus, queries)

	rec := postJSON(t, newSourceRouter(h),
		"/enterprise-capabilities/"+ecID+"/direction/sources",
		AddDirectionSourceRequest{CapabilityID: uuid.New().String()})

	require.Equal(t, http.StatusConflict, rec.Code)
	body := decodeErrorBody(t, rec)
	assert.Equal(t, "Conflict", body.Error)
	assert.Equal(t, "This direction is agreed and its source set is frozen. To recompose, reject the direction and capture a new one.", body.Message)
	assert.Equal(t, "agreed", body.Details["directionStatus"])
	require.Contains(t, body.Links, "x-reject")
	assert.Equal(t, "/api/v1/enterprise-capabilities/"+ecID+"/direction/reject", body.Links["x-reject"].Href)
	assert.Equal(t, "POST", body.Links["x-reject"].Method)
}

func TestAddDirectionSource_ProposedDirection_SourceSetFrozenBody(t *testing.T) {
	ecID := uuid.New().String()
	bus := &mockCommandBus{err: aggregates.ErrDirectionSourceSetFrozen}
	queries := &mockDirectionQueries{activeByEC: &readmodels.DirectionDTO{
		ID: uuid.New().String(), EnterpriseCapabilityID: ecID, Status: "proposed",
	}}
	h := setupHandlers(bus, queries)

	rec := postJSON(t, newSourceRouter(h),
		"/enterprise-capabilities/"+ecID+"/direction/sources",
		AddDirectionSourceRequest{CapabilityID: uuid.New().String()})

	require.Equal(t, http.StatusConflict, rec.Code)
	body := decodeErrorBody(t, rec)
	assert.Equal(t, "proposed", body.Details["directionStatus"])
	assert.Equal(t, "This direction is proposed and its source set is frozen. To recompose, reject the direction and capture a new one.", body.Message)
	require.Contains(t, body.Links, "x-reject")
}

func TestAddDirectionSource_R1Conflict_BodyWithDetailsAndConflictLink(t *testing.T) {
	ecID := uuid.New().String()
	bus := &mockCommandBus{err: &services.SourceConflictError{Conflict: services.SourceConflict{
		CapabilityID:             "cap-001",
		CapabilityName:           "Customer Account Creation",
		EnterpriseCapabilityID:   "ec-customer-identity",
		EnterpriseCapabilityName: "Customer Identity",
	}}}
	queries := &mockDirectionQueries{activeByEC: &readmodels.DirectionDTO{
		ID: uuid.New().String(), EnterpriseCapabilityID: ecID, Status: "draft",
	}}
	h := setupHandlers(bus, queries)

	rec := postJSON(t, newSourceRouter(h),
		"/enterprise-capabilities/"+ecID+"/direction/sources",
		AddDirectionSourceRequest{CapabilityID: "cap-001"})

	require.Equal(t, http.StatusConflict, rec.Code)
	body := decodeErrorBody(t, rec)
	assert.Equal(t, "Conflict", body.Error)
	assert.Equal(t, "Capability 'Customer Account Creation' is already an explicit source of an active direction on 'Customer Identity'. A domain capability may be the explicit source of at most one active direction.", body.Message)
	assert.Equal(t, map[string]string{
		"capabilityId":                        "cap-001",
		"capabilityName":                      "Customer Account Creation",
		"conflictingEnterpriseCapabilityId":   "ec-customer-identity",
		"conflictingEnterpriseCapabilityName": "Customer Identity",
	}, body.Details)
	require.Contains(t, body.Links, "x-conflicting-ec")
	assert.Equal(t, "/api/v1/enterprise-capabilities/ec-customer-identity", body.Links["x-conflicting-ec"].Href)
}

func TestRemoveDirectionSource_Dispatches204(t *testing.T) {
	ecID := uuid.New().String()
	did := uuid.New().String()
	capID := uuid.New().String()
	bus := &mockCommandBus{}
	queries := &mockDirectionQueries{activeByEC: &readmodels.DirectionDTO{
		ID: did, EnterpriseCapabilityID: ecID, Status: "draft",
	}}
	h := setupHandlers(bus, queries)

	rec := performWithoutBody(newSourceRouter(h), http.MethodDelete,
		"/enterprise-capabilities/"+ecID+"/direction/sources/"+capID)

	require.Equal(t, http.StatusNoContent, rec.Code)
	require.Len(t, bus.dispatched, 1)
	cmd, ok := bus.dispatched[0].(*commands.RemoveDirectionSource)
	require.True(t, ok)
	assert.Equal(t, did, cmd.DirectionID)
	assert.Equal(t, capID, cmd.CapabilityID)
	assert.Equal(t, "user@example.com", cmd.Actor)
}

func TestRemoveDirectionSource_NotASource_404(t *testing.T) {
	ecID := uuid.New().String()
	bus := &mockCommandBus{err: aggregates.ErrSourceCapabilityNotInDirection}
	queries := &mockDirectionQueries{activeByEC: &readmodels.DirectionDTO{
		ID: uuid.New().String(), EnterpriseCapabilityID: ecID, Status: "draft",
	}}
	h := setupHandlers(bus, queries)

	rec := performWithoutBody(newSourceRouter(h), http.MethodDelete,
		"/enterprise-capabilities/"+ecID+"/direction/sources/"+uuid.New().String())

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestRemoveDirectionSource_NonDraftDirection_SourceSetFrozenBody(t *testing.T) {
	for _, status := range []string{"agreed", "proposed"} {
		t.Run(status, func(t *testing.T) {
			ecID := uuid.New().String()
			bus := &mockCommandBus{err: aggregates.ErrDirectionSourceSetFrozen}
			queries := &mockDirectionQueries{activeByEC: &readmodels.DirectionDTO{
				ID: uuid.New().String(), EnterpriseCapabilityID: ecID, Status: status,
			}}
			h := setupHandlers(bus, queries)

			rec := performWithoutBody(newSourceRouter(h), http.MethodDelete,
				"/enterprise-capabilities/"+ecID+"/direction/sources/"+uuid.New().String())

			require.Equal(t, http.StatusConflict, rec.Code)
			body := decodeErrorBody(t, rec)
			assert.Equal(t, status, body.Details["directionStatus"])
			assert.Contains(t, body.Links, "x-reject")
		})
	}
}

func TestCaptureDirection_InactiveEC_409WithContractMessage(t *testing.T) {
	ecID := uuid.New().String()
	bus := &mockCommandBus{err: services.ErrEnterpriseCapabilityInactive}
	h := setupHandlers(bus, &mockDirectionQueries{})

	r := chi.NewRouter()
	r.Post("/enterprise-capabilities/{id}/direction", h.CaptureDirection)
	rec := postJSON(t, r, "/enterprise-capabilities/"+ecID+"/direction", CaptureDirectionRequest{
		Type: "consolidate", SourceCapabilityIDs: []string{}, Horizon: "now",
	})

	require.Equal(t, http.StatusConflict, rec.Code)
	body := decodeErrorBody(t, rec)
	assert.Equal(t, "Directions can only be captured on active enterprise capabilities.", body.Message)
	assert.Empty(t, body.Details)
	assert.Empty(t, body.Links)
}

func TestCaptureDirection_R1Conflict_ContractBody(t *testing.T) {
	ecID := uuid.New().String()
	bus := &mockCommandBus{err: &services.SourceConflictError{Conflict: services.SourceConflict{
		CapabilityID:             "cap-001",
		CapabilityName:           "Customer Account Creation",
		EnterpriseCapabilityID:   "ec-x",
		EnterpriseCapabilityName: "Customer Identity",
	}}}
	h := setupHandlers(bus, &mockDirectionQueries{})

	r := chi.NewRouter()
	r.Post("/enterprise-capabilities/{id}/direction", h.CaptureDirection)
	rec := postJSON(t, r, "/enterprise-capabilities/"+ecID+"/direction", CaptureDirectionRequest{
		Type: "consolidate", SourceCapabilityIDs: []string{"cap-001"}, Horizon: "now",
	})

	require.Equal(t, http.StatusConflict, rec.Code)
	body := decodeErrorBody(t, rec)
	assert.Equal(t, "cap-001", body.Details["capabilityId"])
	assert.Contains(t, body.Links, "x-conflicting-ec")
}

func TestProposeDirection_CardinalityNotMet_400WithTypeSpecificMessage(t *testing.T) {
	cases := []struct {
		directionType string
		wantMessage   string
	}{
		{"consolidate", "A 'consolidate' direction requires at least 2 sources to be proposed."},
		{"decompose", "A 'decompose' direction requires exactly 1 source to be proposed."},
		{"stay", "A 'stay' direction requires exactly 1 source to be proposed."},
	}
	for _, tc := range cases {
		t.Run(tc.directionType, func(t *testing.T) {
			ecID := uuid.New().String()
			bus := &mockCommandBus{err: aggregates.ErrInvalidSourceCardinality}
			queries := &mockDirectionQueries{activeByEC: &readmodels.DirectionDTO{
				ID: uuid.New().String(), EnterpriseCapabilityID: ecID, Status: "draft", Type: tc.directionType,
			}}
			h := setupHandlers(bus, queries)

			r := chi.NewRouter()
			r.Post("/enterprise-capabilities/{id}/direction/propose", h.ProposeDirection)
			rec := performWithoutBody(r, http.MethodPost, "/enterprise-capabilities/"+ecID+"/direction/propose")

			require.Equal(t, http.StatusBadRequest, rec.Code)
			body := decodeErrorBody(t, rec)
			assert.Equal(t, tc.wantMessage, body.Message)
		})
	}
}

func TestDirectionLinks_OnlyDraftIncludesAddSource(t *testing.T) {
	ecID := uuid.New().String()
	cases := []struct {
		status        string
		wantAddSource bool
	}{
		{"draft", true},
		{"proposed", false},
		{"agreed", false},
		{"rejected", false},
	}
	for _, tc := range cases {
		t.Run(tc.status, func(t *testing.T) {
			queries := &mockDirectionQueries{activeByEC: &readmodels.DirectionDTO{
				ID: uuid.New().String(), EnterpriseCapabilityID: ecID, Status: tc.status,
			}}
			code, body := getDirection(t, queries, ecID, architectActor())
			require.Equal(t, http.StatusOK, code)
			if tc.wantAddSource {
				assert.Contains(t, body.Direction.Links, "x-add-source")
			} else {
				assert.NotContains(t, body.Direction.Links, "x-add-source")
			}
		})
	}
}
