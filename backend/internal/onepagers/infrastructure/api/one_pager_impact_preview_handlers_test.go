package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"easi/backend/internal/onepagers/application/queries"
	"easi/backend/internal/onepagers/domain/valueobjects"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeImpactPreviewReader struct {
	preview        *queries.ImpactPreview
	err            error
	gotSubjectType valueobjects.SubjectType
	gotField       queries.PreviewField
}

func (f *fakeImpactPreviewReader) Preview(_ context.Context, subjectType valueobjects.SubjectType, field queries.PreviewField) (*queries.ImpactPreview, error) {
	f.gotSubjectType = subjectType
	f.gotField = field
	return f.preview, f.err
}

func performImpactPreview(t *testing.T, reader *fakeImpactPreviewReader, subjectType, query string) *httptest.ResponseRecorder {
	t.Helper()
	h := NewImpactPreviewHandlers(reader, testLinks())
	rec, req := requestFor(t, requestSpec{
		method:      http.MethodGet,
		path:        "/one-pagers/configurations/" + subjectType + "/impact-preview" + query,
		subjectType: subjectType,
		actor:       adminActor(),
	})
	h.GetImpactPreview(rec, req)
	return rec
}

func decodeImpactPreviewDTO(t *testing.T, rec *httptest.ResponseRecorder) ImpactPreviewDTO {
	t.Helper()
	var dto ImpactPreviewDTO
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &dto))
	return dto
}

func TestGetImpactPreview_ExistingFieldReturns200(t *testing.T) {
	reader := &fakeImpactPreviewReader{preview: &queries.ImpactPreview{SubjectType: "application", FieldID: "contract-link", AffectedSubjectCount: 37}}

	rec := performImpactPreview(t, reader, "application", "?fieldId=contract-link")

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	assert.Equal(t, "application", reader.gotSubjectType.Value())
	assert.Equal(t, queries.PreviewField{Kind: "custom", ID: "contract-link"}, reader.gotField)

	dto := decodeImpactPreviewDTO(t, rec)
	assert.Equal(t, "application", dto.SubjectType)
	assert.Equal(t, "contract-link", dto.FieldID)
	assert.Equal(t, 37, dto.AffectedSubjectCount)
}

func TestGetImpactPreview_BuiltInFieldRoutesThroughKind(t *testing.T) {
	reader := &fakeImpactPreviewReader{preview: &queries.ImpactPreview{SubjectType: "application", FieldID: "experts", AffectedSubjectCount: 40}}

	rec := performImpactPreview(t, reader, "application", "?fieldId=experts&fieldKind=builtIn")

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	assert.Equal(t, queries.PreviewField{Kind: "builtIn", ID: "experts"}, reader.gotField)
	assert.Equal(t, 40, decodeImpactPreviewDTO(t, rec).AffectedSubjectCount)
}

func TestGetImpactPreview_NewFieldOmitsFieldIDParam(t *testing.T) {
	reader := &fakeImpactPreviewReader{preview: &queries.ImpactPreview{SubjectType: "vendor", FieldID: "", AffectedSubjectCount: 120}}

	rec := performImpactPreview(t, reader, "vendor", "")

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	assert.Equal(t, queries.PreviewField{Kind: "custom", ID: ""}, reader.gotField)
	assert.Equal(t, 120, decodeImpactPreviewDTO(t, rec).AffectedSubjectCount)
}

func TestGetImpactPreview_ErrorMapsToStatusCode(t *testing.T) {
	cases := []struct {
		name           string
		subjectType    string
		query          string
		readerErr      error
		expectedStatus int
	}{
		{"unknown field is 404", "application", "?fieldId=unknown", queries.ErrFieldNotConfigured, http.StatusNotFound},
		{"unknown subject type is 404", "starship", "", nil, http.StatusNotFound},
		{"unexpected error is 500", "application", "", errors.New("boom"), http.StatusInternalServerError},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := performImpactPreview(t, &fakeImpactPreviewReader{err: tc.readerErr}, tc.subjectType, tc.query)

			assert.Equal(t, tc.expectedStatus, rec.Code)
		})
	}
}

func TestGetImpactPreview_SelfLinkPresent(t *testing.T) {
	reader := &fakeImpactPreviewReader{preview: &queries.ImpactPreview{SubjectType: "application", FieldID: "contract-link", AffectedSubjectCount: 37}}

	rec := performImpactPreview(t, reader, "application", "?fieldId=contract-link")

	dto := decodeImpactPreviewDTO(t, rec)
	self, ok := dto.Links["self"]
	require.True(t, ok)
	assert.Equal(t, "/api/v1/one-pagers/configurations/application/impact-preview?fieldId=contract-link", self.Href)
	assert.Equal(t, "GET", self.Method)
}
