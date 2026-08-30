package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sharedAPI "easi/backend/internal/shared/api"
)

type fakeSubjectIDs struct {
	ids map[string][]string
}

func (f *fakeSubjectIDs) SubjectIDs(_ context.Context, subjectType string) ([]string, error) {
	return f.ids[subjectType], nil
}

type fakeCompletenessIndicators struct {
	indicators map[string]bool
	applies    bool
	gotIDs     []string
}

func (f *fakeCompletenessIndicators) ForSubjects(_ context.Context, _ string, subjectIDs []string) (map[string]bool, bool, error) {
	f.gotIDs = subjectIDs
	return f.indicators, f.applies, nil
}

func TestGetCompleteness_ReportsEverySubjectOfTheType(t *testing.T) {
	subjects := &fakeSubjectIDs{ids: map[string][]string{"application": {"app-1", "app-2"}}}
	indicators := &fakeCompletenessIndicators{indicators: map[string]bool{"app-1": true, "app-2": false}, applies: true}
	handlers := NewOnePagerCompletenessHandlers(subjects, indicators, NewOnePagerLinks(sharedAPI.NewHATEOASLinks("/api/v1")))

	rec := httptest.NewRecorder()
	handlers.GetCompleteness("application")(rec, httptest.NewRequest(http.MethodGet, "/one-pagers/application/completeness", nil))

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, []string{"app-1", "app-2"}, indicators.gotIDs)
	var body OnePagerCompletenessResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, []OnePagerCompletenessDTO{{SubjectID: "app-1", Complete: true}, {SubjectID: "app-2", Complete: false}}, body.Data)
	assert.Equal(t, "/api/v1/one-pagers/application/completeness", body.Links["self"].Href)
}

func TestGetCompleteness_EmptyWhenNoRequiredFields(t *testing.T) {
	subjects := &fakeSubjectIDs{ids: map[string][]string{"vendor": {"v-1"}}}
	indicators := &fakeCompletenessIndicators{applies: false}
	handlers := NewOnePagerCompletenessHandlers(subjects, indicators, NewOnePagerLinks(sharedAPI.NewHATEOASLinks("/api/v1")))

	rec := httptest.NewRecorder()
	handlers.GetCompleteness("vendor")(rec, httptest.NewRequest(http.MethodGet, "/one-pagers/vendor/completeness", nil))

	require.Equal(t, http.StatusOK, rec.Code)
	var body OnePagerCompletenessResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Empty(t, body.Data)
}
