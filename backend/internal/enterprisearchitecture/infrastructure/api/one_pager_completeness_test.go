package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"easi/backend/internal/enterprisearchitecture/application/readmodels"
	domainservices "easi/backend/internal/enterprisearchitecture/domain/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeOnePagerCompletenessSource struct {
	calls      int
	gotIDs     []string
	indicators map[string]bool
	present    bool
	err        error
}

func (f *fakeOnePagerCompletenessSource) CompletenessFor(_ context.Context, subjectIDs []string) (map[string]bool, bool, error) {
	f.calls++
	f.gotIDs = subjectIDs
	return f.indicators, f.present, f.err
}

func TestDecorateEnterpriseCapabilitiesOnePagerCompleteness_SetsIndicatorPerRow(t *testing.T) {
	source := &fakeOnePagerCompletenessSource{indicators: map[string]bool{"ec-1": true}, present: true}
	rows := []readmodels.EnterpriseCapabilityDTO{{ID: "ec-1"}, {ID: "ec-2"}}

	err := decorateEnterpriseCapabilitiesOnePagerCompleteness(context.Background(), source, rows)

	require.NoError(t, err)
	assert.Equal(t, []string{"ec-1", "ec-2"}, source.gotIDs)
	require.NotNil(t, rows[0].OnePagerComplete)
	assert.True(t, *rows[0].OnePagerComplete)
	require.NotNil(t, rows[1].OnePagerComplete)
	assert.False(t, *rows[1].OnePagerComplete)
}

func TestDecorateEnterpriseCapabilitiesOnePagerCompleteness_NilSourceOrAbsentIndicator_LeavesRowsUntouched(t *testing.T) {
	cases := []struct {
		name   string
		source OnePagerCompletenessSource
	}{
		{"nil source", nil},
		{"indicator absent", &fakeOnePagerCompletenessSource{present: false}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rows := []readmodels.EnterpriseCapabilityDTO{{ID: "ec-1"}}

			err := decorateEnterpriseCapabilitiesOnePagerCompleteness(context.Background(), tc.source, rows)

			require.NoError(t, err)
			assert.Nil(t, rows[0].OnePagerComplete)
		})
	}
}

func TestDecorateEnterpriseCapabilitiesOnePagerCompleteness_SourceError_IsPropagated(t *testing.T) {
	source := &fakeOnePagerCompletenessSource{err: assert.AnError}
	rows := []readmodels.EnterpriseCapabilityDTO{{ID: "ec-1"}}

	err := decorateEnterpriseCapabilitiesOnePagerCompleteness(context.Background(), source, rows)

	require.Error(t, err)
	assert.Nil(t, rows[0].OnePagerComplete)
}

func newCompletenessTestHandlers(source OnePagerCompletenessSource, capabilityRM *mockCapabilityReadModel) *EnterpriseCapabilityHandlers {
	rm := &EnterpriseCapabilityReadModels{
		Capability:           capabilityRM,
		Composition:          &mockCompositionCounts{countsByEC: map[string]domainservices.CompositionCounts{}},
		Importance:           &mockImportanceReadModel{},
		MaturityAnalysis:     &mockMaturityAnalysisReadModel{},
		OnePagerCompleteness: source,
	}
	return NewEnterpriseCapabilityHandlers(&mockCommandBus{}, rm, &mockSessionProvider{email: "test@example.com"})
}

func TestGetAllEnterpriseCapabilities_DecoratesRowsWhenIndicatorPresent(t *testing.T) {
	capabilityRM := newMockCapabilityReadModel()
	capabilityRM.capabilities["ec-1"] = &readmodels.EnterpriseCapabilityDTO{ID: "ec-1", Name: "Payments", Active: true}
	source := &fakeOnePagerCompletenessSource{indicators: map[string]bool{"ec-1": true}, present: true}
	handlers := newCompletenessTestHandlers(source, capabilityRM)

	r := httptest.NewRequest(http.MethodGet, "/enterprise-capabilities", nil)
	w := httptest.NewRecorder()
	handlers.GetAllEnterpriseCapabilities(w, r)

	require.Equal(t, http.StatusOK, w.Code)
	var response struct {
		Data []map[string]any `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
	require.Len(t, response.Data, 1)
	assert.Equal(t, true, response.Data[0]["onePagerComplete"])
}

func TestGetAllEnterpriseCapabilities_NoIndicatorWhenSourceNilOrAbsent(t *testing.T) {
	cases := []struct {
		name   string
		source OnePagerCompletenessSource
	}{
		{"nil source", nil},
		{"indicator absent", &fakeOnePagerCompletenessSource{present: false}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			capabilityRM := newMockCapabilityReadModel()
			capabilityRM.capabilities["ec-1"] = &readmodels.EnterpriseCapabilityDTO{ID: "ec-1", Name: "Payments", Active: true}
			handlers := newCompletenessTestHandlers(tc.source, capabilityRM)

			r := httptest.NewRequest(http.MethodGet, "/enterprise-capabilities", nil)
			w := httptest.NewRecorder()
			handlers.GetAllEnterpriseCapabilities(w, r)

			require.Equal(t, http.StatusOK, w.Code)
			assert.NotContains(t, w.Body.String(), "onePagerComplete")
		})
	}
}
