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
	"easi/backend/internal/architecturedirection/domain/services"
	"easi/backend/internal/architecturedirection/domain/valueobjects"
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/cqrs"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockCapabilityJourneyQueries struct {
	active         *readmodels.CapabilityJourneyDTO
	activeErr      error
	byIDReturns    []*readmodels.CapabilityJourneyDTO
	byIDCallIdx    int
	history        []readmodels.CapabilityJourneyDTO
	bulk           []readmodels.CapabilityJourneyDTO
	receivedCapIDs []string
	all            []readmodels.CapabilityJourneyDTO
	allCalled      bool
}

func (m *mockCapabilityJourneyQueries) GetAllCurrent(_ context.Context) ([]readmodels.CapabilityJourneyDTO, error) {
	m.allCalled = true
	return m.all, nil
}

func (m *mockCapabilityJourneyQueries) GetActiveByCapabilityID(_ context.Context, _ string) (*readmodels.CapabilityJourneyDTO, error) {
	return m.active, m.activeErr
}

func (m *mockCapabilityJourneyQueries) GetHistoryByCapabilityID(_ context.Context, _ string) ([]readmodels.CapabilityJourneyDTO, error) {
	return m.history, nil
}

func (m *mockCapabilityJourneyQueries) GetCurrentByCapabilityIDs(_ context.Context, ids []string) ([]readmodels.CapabilityJourneyDTO, error) {
	m.receivedCapIDs = ids
	return m.bulk, nil
}

func (m *mockCapabilityJourneyQueries) GetByID(_ context.Context, _ string) (*readmodels.CapabilityJourneyDTO, error) {
	var result *readmodels.CapabilityJourneyDTO
	if m.byIDCallIdx < len(m.byIDReturns) {
		result = m.byIDReturns[m.byIDCallIdx]
	}
	m.byIDCallIdx++
	return result, nil
}

func setupCapabilityJourneyHandlers(bus *mockCommandBus, queries CapabilityJourneyQueries) *CapabilityJourneyHandlers {
	links := NewCapabilityJourneyLinks(sharedAPI.NewHATEOASLinks(""))
	return NewCapabilityJourneyHandlers(bus, queries, links)
}

func newMutableJourneyFixture(t *testing.T) (string, *mockCommandBus, chi.Router) {
	t.Helper()
	journeyID := uuid.New().String()
	updated := plannedJourneyDTO()
	updated.ID = journeyID
	bus := &mockCommandBus{}
	queries := &mockCapabilityJourneyQueries{byIDReturns: []*readmodels.CapabilityJourneyDTO{updated}}
	h := setupCapabilityJourneyHandlers(bus, queries)
	return journeyID, bus, capabilityJourneyRouter(h)
}

func capabilityJourneyRouter(h *CapabilityJourneyHandlers) chi.Router {
	r := chi.NewRouter()
	r.Get("/capabilities/{id}/journey", h.GetJourneyForCapability)
	r.Post("/capabilities/{id}/journey", h.CaptureJourney)
	r.Get("/capabilities/{id}/journey/history", h.GetJourneyHistory)
	r.Get("/capability-journeys", h.GetCapabilityJourneys)
	r.Post("/capability-journeys/{journeyId}/start", h.StartJourney)
	r.Post("/capability-journeys/{journeyId}/complete", h.CompleteJourney)
	r.Post("/capability-journeys/{journeyId}/abandon", h.AbandonJourney)
	r.Put("/capability-journeys/{journeyId}/details", h.PutJourneyDetails)
	r.Put("/capability-journeys/{journeyId}/progress", h.PutJourneyProgress)
	r.Put("/capability-journeys/{journeyId}/source-applications", h.PutJourneySourceApplications)
	r.Post("/capability-journeys/{journeyId}/milestones", h.PostJourneyMilestone)
	r.Put("/capability-journeys/{journeyId}/milestones/{milestoneId}", h.PutJourneyMilestone)
	r.Delete("/capability-journeys/{journeyId}/milestones/{milestoneId}", h.DeleteJourneyMilestone)
	r.Put("/capability-journeys/{journeyId}/milestone-order", h.PutJourneyMilestoneOrder)
	return r
}

func withActor(req *http.Request, actor sharedctx.Actor) *http.Request {
	return req.WithContext(sharedctx.WithActor(req.Context(), actor))
}

func TestGetJourneyForCapability_NoActiveJourney_WriterSeesCapture(t *testing.T) {
	queries := &mockCapabilityJourneyQueries{}
	h := setupCapabilityJourneyHandlers(&mockCommandBus{}, queries)
	r := capabilityJourneyRouter(h)

	capID := uuid.New().String()
	req := withActor(httptest.NewRequest(http.MethodGet, "/capabilities/"+capID+"/journey", nil), architectActor())
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var body struct {
		Journey *readmodels.CapabilityJourneyDTO `json:"journey"`
		Links   sharedAPI.Links                  `json:"_links"`
	}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
	assert.Nil(t, body.Journey)
	assert.Contains(t, body.Links, "x-capture")
	assert.Contains(t, body.Links, "x-history")
}

func TestGetJourneyForCapability_NoActiveJourney_ReadOnlyNoCapture(t *testing.T) {
	queries := &mockCapabilityJourneyQueries{}
	h := setupCapabilityJourneyHandlers(&mockCommandBus{}, queries)
	r := capabilityJourneyRouter(h)

	req := withActor(httptest.NewRequest(http.MethodGet, "/capabilities/"+uuid.New().String()+"/journey", nil), stakeholderActor())
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var body struct {
		Links sharedAPI.Links `json:"_links"`
	}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
	assert.NotContains(t, body.Links, "x-capture")
}

func plannedJourneyDTO() *readmodels.CapabilityJourneyDTO {
	return &readmodels.CapabilityJourneyDTO{
		ID:           uuid.New().String(),
		CapabilityID: uuid.New().String(),
		Kind:         valueobjects.JourneyKindMigration,
		Status:       valueobjects.JourneyStatusPlanned,
	}
}

func TestGetJourneyForCapability_ActiveJourney_WriterSeesPlannedLinks(t *testing.T) {
	journey := plannedJourneyDTO()
	queries := &mockCapabilityJourneyQueries{active: journey}
	h := setupCapabilityJourneyHandlers(&mockCommandBus{}, queries)
	r := capabilityJourneyRouter(h)

	req := withActor(httptest.NewRequest(http.MethodGet, "/capabilities/"+journey.CapabilityID+"/journey", nil), architectActor())
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var body struct {
		Journey *readmodels.CapabilityJourneyDTO `json:"journey"`
	}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
	require.NotNil(t, body.Journey)
	assert.Contains(t, body.Journey.Links, "x-start")
	assert.Contains(t, body.Journey.Links, "x-abandon")
	assert.Contains(t, body.Journey.Links, "edit")
	assert.Contains(t, body.Journey.Links, "x-progress")
	assert.Contains(t, body.Journey.Links, "x-change-sources")
	assert.Contains(t, body.Journey.Links, "x-add-milestone")
	assert.NotContains(t, body.Journey.Links, "x-complete")
}

func TestGetJourneyForCapability_LinksVaryByStatusAndActor(t *testing.T) {
	cases := []struct {
		name        string
		status      string
		actor       sharedctx.Actor
		wantLinks   []string
		wantNoLinks []string
	}{
		{
			name:        "read-only actor sees no write links",
			status:      valueobjects.JourneyStatusPlanned,
			actor:       stakeholderActor(),
			wantNoLinks: []string{"x-start", "edit"},
		},
		{
			name:        "in-flight journey shows complete not start",
			status:      valueobjects.JourneyStatusInFlight,
			actor:       architectActor(),
			wantLinks:   []string{"x-complete"},
			wantNoLinks: []string{"x-start"},
		},
		{
			name:        "terminal journey shows only self and history",
			status:      valueobjects.JourneyStatusDone,
			actor:       architectActor(),
			wantLinks:   []string{"self", "x-history"},
			wantNoLinks: []string{"edit", "x-abandon"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			journey := plannedJourneyDTO()
			journey.Status = tc.status
			queries := &mockCapabilityJourneyQueries{active: journey}
			h := setupCapabilityJourneyHandlers(&mockCommandBus{}, queries)
			r := capabilityJourneyRouter(h)

			req := withActor(httptest.NewRequest(http.MethodGet, "/capabilities/"+journey.CapabilityID+"/journey", nil), tc.actor)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			require.Equal(t, http.StatusOK, rec.Code)
			var body struct {
				Journey *readmodels.CapabilityJourneyDTO `json:"journey"`
			}
			require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
			require.NotNil(t, body.Journey)
			for _, link := range tc.wantLinks {
				assert.Contains(t, body.Journey.Links, link)
			}
			for _, link := range tc.wantNoLinks {
				assert.NotContains(t, body.Journey.Links, link)
			}
		})
	}
}

func TestCaptureJourney_Success_Returns201WithLocation(t *testing.T) {
	capID := uuid.New().String()
	created := plannedJourneyDTO()
	created.CapabilityID = capID
	bus := &mockCommandBus{createdID: created.ID}
	queries := &mockCapabilityJourneyQueries{byIDReturns: []*readmodels.CapabilityJourneyDTO{created}}
	h := setupCapabilityJourneyHandlers(bus, queries)
	r := capabilityJourneyRouter(h)

	body, _ := json.Marshal(CaptureJourneyRequest{
		Kind:             valueobjects.JourneyKindMigration,
		FromComponentIDs: []string{uuid.New().String()},
		ToComponentID:    uuid.New().String(),
		Note:             "moving on",
	})
	req := withActor(httptest.NewRequest(http.MethodPost, "/capabilities/"+capID+"/journey", bytes.NewReader(body)), architectActor())
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	assert.Equal(t, "/api/v1/capabilities/"+capID+"/journey", rec.Header().Get("Location"))
	require.Len(t, bus.dispatched, 1)
	cmd := bus.dispatched[0].(*commands.PlanJourney)
	assert.Equal(t, capID, cmd.CapabilityID)
	assert.Equal(t, "user@example.com", cmd.PlannedBy)
}

func TestCaptureJourney_ActiveJourneyExists_Returns409WithExistingID(t *testing.T) {
	existingID := uuid.New().String()
	bus := &mockCommandBus{err: &services.ActiveJourneyError{ExistingJourneyID: existingID}}
	h := setupCapabilityJourneyHandlers(bus, &mockCapabilityJourneyQueries{})
	r := capabilityJourneyRouter(h)

	body, _ := json.Marshal(CaptureJourneyRequest{Kind: valueobjects.JourneyKindMigration, FromComponentIDs: []string{uuid.New().String()}, ToComponentID: uuid.New().String()})
	req := withActor(httptest.NewRequest(http.MethodPost, "/capabilities/"+uuid.New().String()+"/journey", bytes.NewReader(body)), architectActor())
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusConflict, rec.Code)
	assert.Contains(t, rec.Body.String(), existingID)
}

func TestCaptureJourney_ParentNotInTargetDomain_Returns400WithSpecificMessage(t *testing.T) {
	bus := &mockCommandBus{err: services.ErrTargetParentNotInTargetDomain}
	h := setupCapabilityJourneyHandlers(bus, &mockCapabilityJourneyQueries{})
	r := capabilityJourneyRouter(h)

	body, _ := json.Marshal(CaptureJourneyRequest{
		Kind:           valueobjects.JourneyKindMove,
		ToComponentID:  uuid.New().String(),
		TargetDomainID: uuid.New().String(),
		TargetParentID: uuid.New().String(),
		ResultingName:  "Freight invoicing",
	})
	req := withActor(httptest.NewRequest(http.MethodPost, "/capabilities/"+uuid.New().String()+"/journey", bytes.NewReader(body)), architectActor())
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
	assert.Contains(t, rec.Body.String(), "The target parent capability must belong to the target business domain")
}

func TestGetJourneyHistory_ReturnsCollection(t *testing.T) {
	queries := &mockCapabilityJourneyQueries{history: []readmodels.CapabilityJourneyDTO{*plannedJourneyDTO()}}
	h := setupCapabilityJourneyHandlers(&mockCommandBus{}, queries)
	r := capabilityJourneyRouter(h)

	capID := uuid.New().String()
	req := withActor(httptest.NewRequest(http.MethodGet, "/capabilities/"+capID+"/journey/history", nil), architectActor())
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var body sharedAPI.CollectionResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
	assert.Contains(t, body.Links, "self")
}

func TestGetCapabilityJourneys_CapabilityIDFiltering(t *testing.T) {
	capA, capB := uuid.New().String(), uuid.New().String()
	cases := []struct {
		name          string
		query         string
		wantAll       bool
		wantForwarded []string
	}{
		{"no capabilityIds param fetches the whole collection", "", true, nil},
		{"empty capabilityIds param forwards an empty filter", "?capabilityIds=", false, []string{}},
		{"populated capabilityIds param forwards the parsed ids", "?capabilityIds=" + capA + "," + capB, false, []string{capA, capB}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			queries := &mockCapabilityJourneyQueries{all: []readmodels.CapabilityJourneyDTO{*plannedJourneyDTO()}}
			h := setupCapabilityJourneyHandlers(&mockCommandBus{}, queries)
			r := capabilityJourneyRouter(h)

			req := withActor(httptest.NewRequest(http.MethodGet, "/capability-journeys"+tc.query, nil), architectActor())
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			require.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, tc.wantAll, queries.allCalled)
			if !tc.wantAll {
				assert.Equal(t, tc.wantForwarded, queries.receivedCapIDs)
			}
		})
	}
}

func TestGetCapabilityJourneys_XCaptureLinkVisibilityByActor(t *testing.T) {
	cases := []struct {
		name       string
		actor      sharedctx.Actor
		shouldShow bool
	}{
		{"architect sees x-capture", architectActor(), true},
		{"stakeholder does not see x-capture", stakeholderActor(), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			queries := &mockCapabilityJourneyQueries{}
			h := setupCapabilityJourneyHandlers(&mockCommandBus{}, queries)
			r := capabilityJourneyRouter(h)

			req := withActor(httptest.NewRequest(http.MethodGet, "/capability-journeys?capabilityIds="+uuid.New().String(), nil), tc.actor)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			require.Equal(t, http.StatusOK, rec.Code)
			var body sharedAPI.CollectionResponse
			require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
			if tc.shouldShow {
				assert.Contains(t, body.Links, "x-capture")
			} else {
				assert.NotContains(t, body.Links, "x-capture")
			}
		})
	}
}

func TestJourneyTransitions_DispatchesCommand_Returns200(t *testing.T) {
	cases := []struct {
		name    string
		path    string
		extract func(cmd cqrs.Command) (journeyID, actor string)
	}{
		{"start", "start", func(cmd cqrs.Command) (string, string) {
			c := cmd.(*commands.StartJourney)
			return c.JourneyID, c.Actor
		}},
		{"complete", "complete", func(cmd cqrs.Command) (string, string) {
			c := cmd.(*commands.CompleteJourney)
			return c.JourneyID, c.Actor
		}},
		{"abandon", "abandon", func(cmd cqrs.Command) (string, string) {
			c := cmd.(*commands.AbandonJourney)
			return c.JourneyID, c.Actor
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			journeyID, bus, r := newMutableJourneyFixture(t)

			req := withActor(httptest.NewRequest(http.MethodPost, "/capability-journeys/"+journeyID+"/"+tc.path, nil), architectActor())
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			require.Len(t, bus.dispatched, 1)
			gotJourneyID, gotActor := tc.extract(bus.dispatched[0])
			assert.Equal(t, journeyID, gotJourneyID)
			assert.Equal(t, "user@example.com", gotActor)
		})
	}
}

type journeyMutationDispatchCase struct {
	name       string
	method     string
	pathSuffix string
	body       any
	wantStatus int
	assertCmd  func(t *testing.T, journeyID string, cmd cqrs.Command)
}

func journeyMutationDispatchCases(milestoneID, newFrom string) []journeyMutationDispatchCase {
	return []journeyMutationDispatchCase{
		{
			name: "progress", method: http.MethodPut, pathSuffix: "progress",
			body: UpdateJourneyProgressRequest{Progress: 60}, wantStatus: http.StatusOK,
			assertCmd: func(t *testing.T, journeyID string, cmd cqrs.Command) {
				c := cmd.(*commands.UpdateJourneyProgress)
				assert.Equal(t, journeyID, c.JourneyID)
				assert.Equal(t, 60, c.Progress)
			},
		},
		{
			name: "details", method: http.MethodPut, pathSuffix: "details",
			body: UpdateJourneyDetailsRequest{Note: "updated", TargetPeriod: &TargetPeriodRequest{Year: 2027, Quarter: 2}}, wantStatus: http.StatusOK,
			assertCmd: func(t *testing.T, journeyID string, cmd cqrs.Command) {
				c := cmd.(*commands.UpdateJourneyDetails)
				assert.Equal(t, journeyID, c.JourneyID)
				assert.Equal(t, "updated", c.Note)
				require.NotNil(t, c.TargetYear)
				assert.Equal(t, 2027, *c.TargetYear)
			},
		},
		{
			name: "source applications", method: http.MethodPut, pathSuffix: "source-applications",
			body: ChangeJourneySourceApplicationsRequest{ComponentIDs: []string{newFrom}}, wantStatus: http.StatusOK,
			assertCmd: func(t *testing.T, journeyID string, cmd cqrs.Command) {
				c := cmd.(*commands.ChangeJourneySourceApplications)
				assert.Equal(t, journeyID, c.JourneyID)
				assert.Equal(t, []string{newFrom}, c.FromComponentIDs)
			},
		},
		{
			name: "add milestone defaults status to planned", method: http.MethodPost, pathSuffix: "milestones",
			body: AddJourneyMilestoneRequest{Label: "Cut over region A"}, wantStatus: http.StatusCreated,
			assertCmd: func(t *testing.T, journeyID string, cmd cqrs.Command) {
				c := cmd.(*commands.AddJourneyMilestone)
				assert.Equal(t, journeyID, c.JourneyID)
				assert.Equal(t, "Cut over region A", c.Label)
				assert.Equal(t, valueobjects.MilestoneStatusPlanned, c.Status)
			},
		},
		{
			name: "update milestone", method: http.MethodPut, pathSuffix: "milestones/" + milestoneID,
			body: UpdateJourneyMilestoneRequest{Label: "done now", Status: valueobjects.MilestoneStatusDone}, wantStatus: http.StatusOK,
			assertCmd: func(t *testing.T, journeyID string, cmd cqrs.Command) {
				c := cmd.(*commands.UpdateJourneyMilestone)
				assert.Equal(t, journeyID, c.JourneyID)
				assert.Equal(t, milestoneID, c.MilestoneID)
				assert.Equal(t, valueobjects.MilestoneStatusDone, c.Status)
			},
		},
		{
			name: "reorder milestones", method: http.MethodPut, pathSuffix: "milestone-order",
			body: ReorderJourneyMilestonesRequest{MilestoneIDs: []string{milestoneID, newFrom}}, wantStatus: http.StatusOK,
			assertCmd: func(t *testing.T, journeyID string, cmd cqrs.Command) {
				c := cmd.(*commands.ReorderJourneyMilestones)
				assert.Equal(t, journeyID, c.JourneyID)
				assert.Equal(t, []string{milestoneID, newFrom}, c.MilestoneIDs)
				assert.Equal(t, "user@example.com", c.Actor)
			},
		},
	}
}

func TestJourneyMutationHandlers_DispatchCommand(t *testing.T) {
	for _, tc := range journeyMutationDispatchCases(uuid.New().String(), uuid.New().String()) {
		t.Run(tc.name, func(t *testing.T) {
			journeyID, bus, r := newMutableJourneyFixture(t)

			body, _ := json.Marshal(tc.body)
			req := withActor(httptest.NewRequest(tc.method, "/capability-journeys/"+journeyID+"/"+tc.pathSuffix, bytes.NewReader(body)), architectActor())
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			require.Equal(t, tc.wantStatus, rec.Code, rec.Body.String())
			require.Len(t, bus.dispatched, 1)
			tc.assertCmd(t, journeyID, bus.dispatched[0])
		})
	}
}

func TestDeleteJourneyMilestone_DispatchesCommand_Returns204(t *testing.T) {
	journeyID, milestoneID := uuid.New().String(), uuid.New().String()
	bus := &mockCommandBus{}
	h := setupCapabilityJourneyHandlers(bus, &mockCapabilityJourneyQueries{})
	r := capabilityJourneyRouter(h)

	req := withActor(httptest.NewRequest(http.MethodDelete, "/capability-journeys/"+journeyID+"/milestones/"+milestoneID, nil), architectActor())
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNoContent, rec.Code)
	cmd := bus.dispatched[0].(*commands.RemoveJourneyMilestone)
	assert.Equal(t, journeyID, cmd.JourneyID)
	assert.Equal(t, milestoneID, cmd.MilestoneID)
}

func TestGetJourneyForCapability_ReorderLinkGatedByActorStatusAndMilestoneCount_Rule6(t *testing.T) {
	twoMilestones := []readmodels.CapabilityJourneyMilestoneDTO{{ID: "m1", Label: "Pilot"}, {ID: "m2", Label: "Rollout"}}
	cases := []struct {
		name       string
		status     string
		milestones []readmodels.CapabilityJourneyMilestoneDTO
		actor      sharedctx.Actor
		wantLink   bool
	}{
		{name: "architect on active journey with two milestones", status: valueobjects.JourneyStatusInFlight, milestones: twoMilestones, actor: architectActor(), wantLink: true},
		{name: "read-only actor gets no affordance", status: valueobjects.JourneyStatusInFlight, milestones: twoMilestones, actor: stakeholderActor(), wantLink: false},
		{name: "terminal journey is frozen", status: valueobjects.JourneyStatusDone, milestones: twoMilestones, actor: architectActor(), wantLink: false},
		{name: "single milestone has nothing to reorder", status: valueobjects.JourneyStatusPlanned, milestones: twoMilestones[:1], actor: architectActor(), wantLink: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			journey := plannedJourneyDTO()
			journey.Status = tc.status
			journey.Milestones = tc.milestones
			h := setupCapabilityJourneyHandlers(&mockCommandBus{}, &mockCapabilityJourneyQueries{active: journey})
			r := capabilityJourneyRouter(h)

			req := withActor(httptest.NewRequest(http.MethodGet, "/capabilities/"+journey.CapabilityID+"/journey", nil), tc.actor)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			require.Equal(t, http.StatusOK, rec.Code)
			var body struct {
				Journey *readmodels.CapabilityJourneyDTO `json:"journey"`
			}
			require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
			link, present := body.Journey.Links["x-reorder-milestones"]
			assert.Equal(t, tc.wantLink, present)
			if tc.wantLink {
				assert.Equal(t, "/api/v1/capability-journeys/"+journey.ID+"/milestone-order", link.Href)
				assert.Equal(t, http.MethodPut, link.Method)
			}
		})
	}
}
