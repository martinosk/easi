package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"easi/backend/internal/enterprisearchitecture/application/readmodels"
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockTimeSuggestionReadModel struct {
	suggestions []readmodels.TimeSuggestionDTO
	err         error
}

func (m *mockTimeSuggestionReadModel) GetAllSuggestions(ctx context.Context) ([]readmodels.TimeSuggestionDTO, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.suggestions, nil
}

func (m *mockTimeSuggestionReadModel) filterBy(matchFn func(readmodels.TimeSuggestionDTO) bool) ([]readmodels.TimeSuggestionDTO, error) {
	if m.err != nil {
		return nil, m.err
	}
	var result []readmodels.TimeSuggestionDTO
	for _, s := range m.suggestions {
		if matchFn(s) {
			result = append(result, s)
		}
	}
	return result, nil
}

func (m *mockTimeSuggestionReadModel) GetByCapability(ctx context.Context, capabilityID string) ([]readmodels.TimeSuggestionDTO, error) {
	return m.filterBy(func(s readmodels.TimeSuggestionDTO) bool { return s.CapabilityID == capabilityID })
}

func (m *mockTimeSuggestionReadModel) GetByComponent(ctx context.Context, componentID string) ([]readmodels.TimeSuggestionDTO, error) {
	return m.filterBy(func(s readmodels.TimeSuggestionDTO) bool { return s.ComponentID == componentID })
}

func createTestTimeSuggestionsHandler(mock *mockTimeSuggestionReadModel) *TimeSuggestionsHandlers {
	return NewTimeSuggestionsHandlers(mock, sharedAPI.NewHATEOASLinks(""))
}

func TestGetTimeSuggestions_ReturnsAllSuggestions(t *testing.T) {
	suggestedTime := "INVEST"
	techGap := 1.2
	funcGap := 0.8

	mock := &mockTimeSuggestionReadModel{
		suggestions: []readmodels.TimeSuggestionDTO{
			{
				CapabilityID:   "cap-1",
				CapabilityName: "Customer Management",
				ComponentID:    "comp-1",
				ComponentName:  "CRM System",
				SuggestedTime:  &suggestedTime,
				TechnicalGap:   &techGap,
				FunctionalGap:  &funcGap,
				Confidence:     "HIGH",
			},
		},
	}

	handler := createTestTimeSuggestionsHandler(mock)

	req := httptest.NewRequest(http.MethodGet, "/time-suggestions", nil)
	req = addTestTenantContext(req)
	rr := httptest.NewRecorder()

	handler.GetTimeSuggestions(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response sharedAPI.CollectionResponse
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Len(t, response.Data, 1)
}

func createTwoSuggestionMock(suggestedTime string, techGap, funcGap float64) *mockTimeSuggestionReadModel {
	return &mockTimeSuggestionReadModel{
		suggestions: []readmodels.TimeSuggestionDTO{
			{
				CapabilityID:   "cap-1",
				CapabilityName: "Customer Management",
				ComponentID:    "comp-1",
				ComponentName:  "CRM System",
				SuggestedTime:  &suggestedTime,
				TechnicalGap:   &techGap,
				FunctionalGap:  &funcGap,
				Confidence:     "HIGH",
			},
			{
				CapabilityID:   "cap-2",
				CapabilityName: "Order Management",
				ComponentID:    "comp-2",
				ComponentName:  "Order System",
				SuggestedTime:  &suggestedTime,
				TechnicalGap:   &techGap,
				FunctionalGap:  &funcGap,
				Confidence:     "MEDIUM",
			},
		},
	}
}

func TestGetTimeSuggestions_FilterByDimension(t *testing.T) {
	tests := []struct {
		name  string
		query string
	}{
		{"by capability", "/time-suggestions?capabilityId=cap-1"},
		{"by component", "/time-suggestions?componentId=comp-2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := createTwoSuggestionMock("MIGRATE", 2.0, 0.5)
			handler := createTestTimeSuggestionsHandler(mock)

			req := httptest.NewRequest(http.MethodGet, tt.query, nil)
			req = addTestTenantContext(req)
			rr := httptest.NewRecorder()

			handler.GetTimeSuggestions(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code)

			var response sharedAPI.CollectionResponse
			err := json.Unmarshal(rr.Body.Bytes(), &response)
			require.NoError(t, err)
			assert.Len(t, response.Data, 1)
		})
	}
}

func TestGetTimeSuggestions_ReturnsEmptyArrayWhenNoSuggestions(t *testing.T) {
	mock := &mockTimeSuggestionReadModel{
		suggestions: []readmodels.TimeSuggestionDTO{},
	}

	handler := createTestTimeSuggestionsHandler(mock)

	req := httptest.NewRequest(http.MethodGet, "/time-suggestions", nil)
	req = addTestTenantContext(req)
	rr := httptest.NewRecorder()

	handler.GetTimeSuggestions(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response sharedAPI.CollectionResponse
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Empty(t, response.Data)
}

func addTestTenantContext(req *http.Request) *http.Request {
	ctx := req.Context()
	tenantID, _ := sharedvo.NewTenantID("test-tenant")
	ctx = sharedctx.WithTenant(ctx, tenantID)
	rctx := chi.NewRouteContext()
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
	return req.WithContext(ctx)
}
