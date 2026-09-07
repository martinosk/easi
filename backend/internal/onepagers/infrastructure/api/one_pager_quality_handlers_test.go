package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"easi/backend/internal/onepagers/application/readmodels"
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakePageSource struct {
	gotQuery readmodels.SubjectIndexQuery
	records  []readmodels.SubjectIndexRecord
	hasMore  bool
	err      error
}

func (f *fakePageSource) Page(_ context.Context, query readmodels.SubjectIndexQuery) ([]readmodels.SubjectIndexRecord, bool, error) {
	f.gotQuery = query
	return f.records, f.hasMore, f.err
}

func qualityHandler(source *fakePageSource) http.HandlerFunc {
	links := NewOnePagerLinks(sharedAPI.NewHATEOASLinks("/api/v1"))
	return NewOnePagerQualityHandlers(source, links).GetQualityList
}

func requestWithActor(target string, permissions map[string]bool) *http.Request {
	req := httptest.NewRequest(http.MethodGet, target, nil)
	actor := sharedctx.Actor{ID: "u1", Permissions: permissions}
	return req.WithContext(sharedctx.WithActor(req.Context(), actor))
}

func allReadPermissions() map[string]bool {
	return map[string]bool{"capabilities:read": true, "components:read": true}
}

func TestQualityList_403WhenNoReadPermission(t *testing.T) {
	rec := httptest.NewRecorder()
	qualityHandler(&fakePageSource{}).ServeHTTP(rec, requestWithActor("/api/v1/one-pager-quality", map[string]bool{"views:read": true}))
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestQualityList_FiltersToReadableSubjectTypes(t *testing.T) {
	source := &fakePageSource{}
	rec := httptest.NewRecorder()
	qualityHandler(source).ServeHTTP(rec, requestWithActor("/api/v1/one-pager-quality", map[string]bool{"components:read": true}))

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, []string{"application", "acquired-entity", "vendor", "internal-team"}, source.gotQuery.SubjectTypes)
	assert.Equal(t, readmodels.SortCompleteness, source.gotQuery.Sort)
	assert.Equal(t, readmodels.OrderAsc, source.gotQuery.Order)
	assert.Equal(t, 50, source.gotQuery.Limit)
}

func TestQualityList_ParsesSortOrderLimit(t *testing.T) {
	source := &fakePageSource{}
	rec := httptest.NewRecorder()
	qualityHandler(source).ServeHTTP(rec, requestWithActor("/api/v1/one-pager-quality?sort=name&order=desc&limit=10", allReadPermissions()))

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, readmodels.SortName, source.gotQuery.Sort)
	assert.Equal(t, readmodels.OrderDesc, source.gotQuery.Order)
	assert.Equal(t, 10, source.gotQuery.Limit)
}

func TestQualityList_400OnInvalidSort(t *testing.T) {
	rec := httptest.NewRecorder()
	qualityHandler(&fakePageSource{}).ServeHTTP(rec, requestWithActor("/api/v1/one-pager-quality?sort=bogus", allReadPermissions()))
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestQualityList_400OnInvalidCursor(t *testing.T) {
	rec := httptest.NewRecorder()
	qualityHandler(&fakePageSource{}).ServeHTTP(rec, requestWithActor("/api/v1/one-pager-quality?after=%21%21bad", allReadPermissions()))
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestQualityList_RowsAndNextLink(t *testing.T) {
	source := &fakePageSource{
		hasMore: true,
		records: []readmodels.SubjectIndexRecord{
			{SubjectType: "application", SubjectID: "app-1", Name: "Billing", CreatorEmail: "a@x.com", CreatedAt: time.Now(), LastUpdatedAt: time.Now(), RequiredCount: 2, FilledCount: 1},
			{SubjectType: "vendor", SubjectID: "ven-1", Name: "Acme", CreatorEmail: "b@x.com", CreatedAt: time.Now(), LastUpdatedAt: time.Now(), RequiredCount: 0, FilledCount: 0},
		},
	}
	rec := httptest.NewRecorder()
	qualityHandler(source).ServeHTTP(rec, requestWithActor("/api/v1/one-pager-quality?sort=name&order=asc", allReadPermissions()))

	require.Equal(t, http.StatusOK, rec.Code)
	var body sharedAPI.PaginatedResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))

	rows, ok := body.Data.([]any)
	require.True(t, ok)
	assert.Len(t, rows, 2)
	assert.True(t, body.Pagination.HasMore)
	assert.NotEmpty(t, body.Pagination.Cursor)

	self, ok := body.Links["self"]
	require.True(t, ok)
	assert.Contains(t, self.Href, "sort=name")
	next, ok := body.Links["next"]
	require.True(t, ok)
	assert.Contains(t, next.Href, "sort=name")
	assert.Contains(t, next.Href, "order=asc")
	assert.Contains(t, next.Href, "after=")
}

func TestQualityList_NoNextLinkOnLastPage(t *testing.T) {
	source := &fakePageSource{hasMore: false, records: []readmodels.SubjectIndexRecord{
		{SubjectType: "application", SubjectID: "app-1", Name: "Billing", CreatedAt: time.Now(), LastUpdatedAt: time.Now()},
	}}
	rec := httptest.NewRecorder()
	qualityHandler(source).ServeHTTP(rec, requestWithActor("/api/v1/one-pager-quality", allReadPermissions()))

	require.Equal(t, http.StatusOK, rec.Code)
	var body sharedAPI.PaginatedResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	_, hasNext := body.Links["next"]
	assert.False(t, hasNext)
}

type qualityListBody struct {
	Data []QualityRowDTO `json:"data"`
}

func TestQualityList_RowsCarryEditGrantsLinkGatedOnPermission(t *testing.T) {
	source := &fakePageSource{records: []readmodels.SubjectIndexRecord{
		{SubjectType: "application", SubjectID: "app-1", Name: "Billing", CreatedAt: time.Now(), LastUpdatedAt: time.Now()},
		{SubjectType: "business-unit", SubjectID: "bu-1", Name: "Payments BU", CreatedAt: time.Now(), LastUpdatedAt: time.Now()},
	}}
	rec := httptest.NewRecorder()
	permissions := allReadPermissions()
	permissions["components:write"] = true
	permissions["edit-grants:manage"] = true
	qualityHandler(source).ServeHTTP(rec, requestWithActor("/api/v1/one-pager-quality", permissions))

	require.Equal(t, http.StatusOK, rec.Code)
	var body qualityListBody
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Len(t, body.Data, 2)

	appLink, ok := body.Data[0].Links["x-edit-grants"]
	require.True(t, ok, "application row should carry x-edit-grants")
	assert.Equal(t, "/api/v1/edit-grants", appLink.Href)
	assert.Equal(t, "POST", appLink.Method)

	_, unknownHasLink := body.Data[1].Links["x-edit-grants"]
	assert.False(t, unknownHasLink, "a row of an unknown subject type must never carry x-edit-grants")
}

func TestQualityList_NoEditGrantsLinkWithoutGrantorPermission(t *testing.T) {
	source := &fakePageSource{records: []readmodels.SubjectIndexRecord{
		{SubjectType: "application", SubjectID: "app-1", Name: "Billing", CreatedAt: time.Now(), LastUpdatedAt: time.Now()},
	}}
	rec := httptest.NewRecorder()
	qualityHandler(source).ServeHTTP(rec, requestWithActor("/api/v1/one-pager-quality", allReadPermissions()))

	require.Equal(t, http.StatusOK, rec.Code)
	var body qualityListBody
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Len(t, body.Data, 1)
	_, hasLink := body.Data[0].Links["x-edit-grants"]
	assert.False(t, hasLink, "row must carry no x-edit-grants link without the grantor permission")
}
