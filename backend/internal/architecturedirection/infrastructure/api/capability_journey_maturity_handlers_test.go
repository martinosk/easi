package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"easi/backend/internal/architecturedirection/application/commands"
	"easi/backend/internal/architecturedirection/application/readmodels"
	"easi/backend/internal/architecturedirection/domain/valueobjects"
	sharedAPI "easi/backend/internal/shared/api"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func maturityJourneyDTO() *readmodels.CapabilityJourneyDTO {
	journey := plannedJourneyDTO()
	journey.Kind = valueobjects.JourneyKindMaturity
	journey.Maturity = &readmodels.JourneyMaturityDTO{TargetMaturity: 65, CurrentMaturity: 30, MaturityGap: 35}
	return journey
}

func TestGetJourneyForCapability_CaptureLinksAreTrackScoped_Spec211Rule6(t *testing.T) {
	cases := []struct {
		name        string
		active      []readmodels.CapabilityJourneyDTO
		wantLinks   []string
		wantNoLinks []string
	}{
		{
			name:      "no active journey offers both tracks",
			wantLinks: []string{"x-capture", "x-capture-maturity"},
		},
		{
			name:        "an active migration leaves only the maturity track free",
			active:      []readmodels.CapabilityJourneyDTO{*plannedJourneyDTO()},
			wantLinks:   []string{"x-capture-maturity"},
			wantNoLinks: []string{"x-capture"},
		},
		{
			name:        "an active maturity journey leaves only the application track free",
			active:      []readmodels.CapabilityJourneyDTO{*maturityJourneyDTO()},
			wantLinks:   []string{"x-capture"},
			wantNoLinks: []string{"x-capture-maturity"},
		},
		{
			name:        "both tracks occupied offers neither",
			active:      []readmodels.CapabilityJourneyDTO{*plannedJourneyDTO(), *maturityJourneyDTO()},
			wantNoLinks: []string{"x-capture", "x-capture-maturity"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := setupCapabilityJourneyHandlers(&mockCommandBus{}, &mockCapabilityJourneyQueries{active: tc.active})
			r := capabilityJourneyRouter(h)

			req := withActor(httptest.NewRequest(http.MethodGet, "/capabilities/"+uuid.New().String()+"/journey", nil), architectActor())
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			require.Equal(t, http.StatusOK, rec.Code)
			var body struct {
				Journeys []readmodels.CapabilityJourneyDTO `json:"journeys"`
				Links    sharedAPI.Links                   `json:"_links"`
			}
			require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
			assert.Len(t, body.Journeys, len(tc.active))
			for _, link := range tc.wantLinks {
				assert.Contains(t, body.Links, link)
			}
			for _, link := range tc.wantNoLinks {
				assert.NotContains(t, body.Links, link)
			}
		})
	}
}

func TestGetJourneyForCapability_MaturityJourney_ExposesGapAndNoSourceAffordance(t *testing.T) {
	journey := maturityJourneyDTO()
	h := setupCapabilityJourneyHandlers(&mockCommandBus{}, &mockCapabilityJourneyQueries{active: []readmodels.CapabilityJourneyDTO{*journey}})
	r := capabilityJourneyRouter(h)

	req := withActor(httptest.NewRequest(http.MethodGet, "/capabilities/"+journey.CapabilityID+"/journey", nil), architectActor())
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var body struct {
		Journeys []readmodels.CapabilityJourneyDTO `json:"journeys"`
	}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
	require.Len(t, body.Journeys, 1)
	require.NotNil(t, body.Journeys[0].Maturity)
	assert.Equal(t, 65, body.Journeys[0].Maturity.TargetMaturity)
	assert.Equal(t, 30, body.Journeys[0].Maturity.CurrentMaturity)
	assert.Equal(t, 35, body.Journeys[0].Maturity.MaturityGap)
	assert.NotContains(t, body.Journeys[0].Links, "x-change-sources")
	assert.Contains(t, body.Journeys[0].Links, "x-add-milestone")
}

func TestCaptureJourney_MaturityKind_PassesTargetMaturityToCommand(t *testing.T) {
	capID := uuid.New().String()
	created := maturityJourneyDTO()
	created.CapabilityID = capID
	bus := &mockCommandBus{createdID: created.ID}
	queries := &mockCapabilityJourneyQueries{byIDReturns: []*readmodels.CapabilityJourneyDTO{created}}
	r := capabilityJourneyRouter(setupCapabilityJourneyHandlers(bus, queries))

	target := 65
	body, _ := json.Marshal(CaptureJourneyRequest{Kind: valueobjects.JourneyKindMaturity, TargetMaturity: &target})
	req := withActor(httptest.NewRequest(http.MethodPost, "/capabilities/"+capID+"/journey", bytes.NewReader(body)), architectActor())
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	require.Len(t, bus.dispatched, 1)
	cmd := bus.dispatched[0].(*commands.PlanJourney)
	assert.Equal(t, valueobjects.JourneyKindMaturity, cmd.Kind)
	require.NotNil(t, cmd.TargetMaturity)
	assert.Equal(t, 65, *cmd.TargetMaturity)
	assert.Empty(t, cmd.ToComponentID)
}
