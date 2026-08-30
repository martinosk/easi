package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"easi/backend/internal/onepagers/application/readmodels"
	sharedAPI "easi/backend/internal/shared/api"
)

type fakeSubjectCompleteness struct {
	rows map[string][]readmodels.SubjectCompleteness
}

func (f *fakeSubjectCompleteness) CompletenessFor(_ context.Context, subjectType string) ([]readmodels.SubjectCompleteness, error) {
	return f.rows[subjectType], nil
}

func TestGetCompleteness_ReportsEverySubjectOfTheType(t *testing.T) {
	source := &fakeSubjectCompleteness{rows: map[string][]readmodels.SubjectCompleteness{
		"application": {
			{SubjectID: "app-1", Required: 2, Filled: 2},
			{SubjectID: "app-2", Required: 2, Filled: 1},
		},
	}}
	handlers := NewOnePagerCompletenessHandlers(source, NewOnePagerLinks(sharedAPI.NewHATEOASLinks("/api/v1")))

	rec := httptest.NewRecorder()
	handlers.GetCompleteness("application")(rec, httptest.NewRequest(http.MethodGet, "/one-pagers/application/completeness", nil))

	require.Equal(t, http.StatusOK, rec.Code)
	var body OnePagerCompletenessResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, []OnePagerCompletenessDTO{{SubjectID: "app-1", Complete: true}, {SubjectID: "app-2", Complete: false}}, body.Data)
	assert.Equal(t, "/api/v1/one-pagers/application/completeness", body.Links["self"].Href)
}

func TestGetCompleteness_EmptyWhenNoRequiredFields(t *testing.T) {
	source := &fakeSubjectCompleteness{rows: map[string][]readmodels.SubjectCompleteness{
		"vendor": {{SubjectID: "v-1", Required: 0, Filled: 0}},
	}}
	handlers := NewOnePagerCompletenessHandlers(source, NewOnePagerLinks(sharedAPI.NewHATEOASLinks("/api/v1")))

	rec := httptest.NewRecorder()
	handlers.GetCompleteness("vendor")(rec, httptest.NewRequest(http.MethodGet, "/one-pagers/vendor/completeness", nil))

	require.Equal(t, http.StatusOK, rec.Code)
	var body OnePagerCompletenessResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Empty(t, body.Data)
}
