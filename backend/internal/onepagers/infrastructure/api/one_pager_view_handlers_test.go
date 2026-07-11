package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"easi/backend/internal/onepagers/application/ports"
	"easi/backend/internal/onepagers/application/queries"
	"easi/backend/internal/onepagers/domain/valueobjects"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeOnePagerReader struct {
	onePager *queries.OnePager
	err      error
}

func (f *fakeOnePagerReader) Get(_ context.Context, _ valueobjects.SubjectType, _ string) (*queries.OnePager, error) {
	return f.onePager, f.err
}

func newViewHandlers(reader *fakeOnePagerReader) *OnePagerViewHandlers {
	return NewOnePagerViewHandlers(reader, testLinks())
}

func subjectTypeFor(t *testing.T, value string) valueobjects.SubjectType {
	t.Helper()
	subjectType, err := valueobjects.NewSubjectType(value)
	require.NoError(t, err)
	return subjectType
}

func textEnvelope(t *testing.T, text string) *valueobjects.ValueEnvelope {
	t.Helper()
	value, err := valueobjects.NewTextValue(text)
	require.NoError(t, err)
	envelope, err := valueobjects.NewValueEnvelope(value)
	require.NoError(t, err)
	return &envelope
}

func performGetOnePager(t *testing.T, h *OnePagerViewHandlers) *httptest.ResponseRecorder {
	t.Helper()
	rec, req := requestFor(t, requestSpec{
		method: http.MethodGet,
		path:   "/one-pagers/application/" + testSubjectID,
		params: map[string]string{"subjectID": testSubjectID},
		actor:  stakeholderActor(),
	})
	h.GetOnePager(applicationSubjectType(t))(rec, req)
	return rec
}

func getOnePager(t *testing.T, h *OnePagerViewHandlers, subjectType valueobjects.SubjectType) OnePagerDTO {
	t.Helper()
	rec, req := requestFor(t, requestSpec{
		method: http.MethodGet,
		path:   "/one-pagers/" + subjectType.Value() + "/" + testSubjectID,
		params: map[string]string{"subjectID": testSubjectID},
		actor:  stakeholderActor(),
	})
	h.GetOnePager(subjectType)(rec, req)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	var dto OnePagerDTO
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &dto))
	return dto
}

func getOnePagerFieldsRaw(t *testing.T, h *OnePagerViewHandlers) []any {
	t.Helper()
	rec := performGetOnePager(t, h)
	require.Equal(t, http.StatusOK, rec.Code)
	var raw map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &raw))
	return raw["fields"].([]any)
}

func assertFieldValueRendersJSONNull(t *testing.T, fields []any, kind string) {
	t.Helper()
	field := fields[0].(map[string]any)[kind].(map[string]any)
	value, hasValue := field["value"]
	require.True(t, hasValue, "value key must be present as JSON null, not omitted")
	assert.Nil(t, value)
}

func TestGetOnePager_ReturnsSubjectHeaderAndInterleavedFields(t *testing.T) {
	onePager := &queries.OnePager{
		SubjectType: "application",
		SubjectID:   testSubjectID,
		SubjectName: "Checkout Service",
		Fields: []queries.Field{
			{BuiltIn: &queries.BuiltInField{ID: "description", Label: "Description", Value: ports.TextValue{Text: "Handles checkout"}}},
			{Custom: &queries.CustomField{FieldID: testTextFieldID, Name: "Contract link", FieldType: "link", HelpText: "URL", DisplayText: "Vendor contract", Value: textEnvelope(t, "https://example.com")}},
			{BuiltIn: &queries.BuiltInField{ID: "experts", Label: "Experts", Value: ports.ExpertsValue{Experts: []ports.Expert{{Name: "Ann", Role: "Owner", Contact: "ann@example.com"}}}}},
		},
	}
	h := newViewHandlers(&fakeOnePagerReader{onePager: onePager})

	dto := getOnePager(t, h, applicationSubjectType(t))

	assert.Equal(t, "application", dto.SubjectType)
	assert.Equal(t, testSubjectID, dto.SubjectID)
	assert.Equal(t, "Checkout Service", dto.SubjectName)
	require.Len(t, dto.Fields, 3)

	assert.Equal(t, "builtIn", dto.Fields[0].Kind)
	assert.Equal(t, "description", dto.Fields[0].BuiltIn.ID)
	assert.Equal(t, "custom", dto.Fields[1].Kind)
	assert.Equal(t, testTextFieldID, dto.Fields[1].Custom.FieldID)
	assert.Equal(t, "builtIn", dto.Fields[2].Kind)
	assert.Equal(t, "experts", dto.Fields[2].BuiltIn.ID)
}

func TestGetOnePager_KindDiscriminatorSetsExactlyOnePointer(t *testing.T) {
	onePager := &queries.OnePager{
		SubjectType: "application",
		SubjectID:   testSubjectID,
		Fields: []queries.Field{
			{BuiltIn: &queries.BuiltInField{ID: "description", Label: "Description", Value: ports.TextValue{Text: "x"}}},
			{Custom: &queries.CustomField{FieldID: testTextFieldID, Name: "Contract link", FieldType: "link"}},
		},
	}
	h := newViewHandlers(&fakeOnePagerReader{onePager: onePager})

	dto := getOnePager(t, h, applicationSubjectType(t))

	require.Len(t, dto.Fields, 2)
	assert.NotNil(t, dto.Fields[0].BuiltIn)
	assert.Nil(t, dto.Fields[0].Custom)
	assert.Nil(t, dto.Fields[1].BuiltIn)
	assert.NotNil(t, dto.Fields[1].Custom)
}

type builtInValueCase struct {
	name    string
	value   ports.BuiltInFieldValue
	section string
	assert  func(t *testing.T, dto *BuiltInValueDTO)
}

func runBuiltInValueCases(t *testing.T, cases []builtInValueCase) {
	t.Helper()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			onePager := &queries.OnePager{
				SubjectType: "application",
				SubjectID:   testSubjectID,
				Fields: []queries.Field{
					{BuiltIn: &queries.BuiltInField{ID: "field", Label: "Field", Value: tc.value, MaturitySection: tc.section}},
				},
			}
			h := newViewHandlers(&fakeOnePagerReader{onePager: onePager})

			dto := getOnePager(t, h, applicationSubjectType(t))

			require.Len(t, dto.Fields, 1)
			tc.assert(t, dto.Fields[0].BuiltIn.Value)
		})
	}
}

func TestGetOnePager_BuiltInValueVariants(t *testing.T) {
	acquisitionDate := time.Date(2024, time.March, 15, 0, 0, 0, 0, time.UTC)

	runBuiltInValueCases(t, []builtInValueCase{
		{
			name:  "text",
			value: ports.TextValue{Text: "Runs on shared Kubernetes cluster"},
			assert: func(t *testing.T, dto *BuiltInValueDTO) {
				require.NotNil(t, dto)
				assert.Equal(t, "text", dto.Type)
				require.NotNil(t, dto.Text)
				assert.Equal(t, "Runs on shared Kubernetes cluster", *dto.Text)
			},
		},
		{
			name:  "date",
			value: ports.DateValue{Date: acquisitionDate},
			assert: func(t *testing.T, dto *BuiltInValueDTO) {
				require.NotNil(t, dto)
				assert.Equal(t, "date", dto.Type)
				require.NotNil(t, dto.Date)
				assert.Equal(t, "2024-03-15", *dto.Date)
			},
		},
		{
			name:    "maturity with section",
			value:   ports.MaturityValue{Value: 85},
			section: "Optimizing",
			assert: func(t *testing.T, dto *BuiltInValueDTO) {
				require.NotNil(t, dto)
				assert.Equal(t, "maturity", dto.Type)
				require.NotNil(t, dto.Maturity)
				assert.Equal(t, 85, dto.Maturity.Value)
				assert.Equal(t, "Optimizing", dto.Maturity.Section)
			},
		},
		{
			name:  "maturity without section",
			value: ports.MaturityValue{Value: 10},
			assert: func(t *testing.T, dto *BuiltInValueDTO) {
				require.NotNil(t, dto)
				assert.Equal(t, "maturity", dto.Type)
				require.NotNil(t, dto.Maturity)
				assert.Equal(t, 10, dto.Maturity.Value)
				assert.Empty(t, dto.Maturity.Section)
			},
		},
		{
			name:  "experts",
			value: ports.ExpertsValue{Experts: []ports.Expert{{Name: "Ann", Role: "Owner", Contact: "ann@example.com"}}},
			assert: func(t *testing.T, dto *BuiltInValueDTO) {
				require.NotNil(t, dto)
				assert.Equal(t, "experts", dto.Type)
				require.Len(t, dto.Experts, 1)
				assert.Equal(t, "Ann", dto.Experts[0].Name)
				assert.Equal(t, "Owner", dto.Experts[0].Role)
				assert.Equal(t, "ann@example.com", dto.Experts[0].Contact)
			},
		},
		{
			name:  "nil renders as JSON null",
			value: nil,
			assert: func(t *testing.T, dto *BuiltInValueDTO) {
				assert.Nil(t, dto)
			},
		},
	})
}

func TestGetOnePager_CustomFieldWithValue(t *testing.T) {
	onePager := &queries.OnePager{
		SubjectType: "application",
		SubjectID:   testSubjectID,
		Fields: []queries.Field{
			{Custom: &queries.CustomField{
				FieldID:     testTextFieldID,
				Name:        "Contract link",
				FieldType:   "link",
				HelpText:    "URL",
				DisplayText: "Vendor contract",
				Value:       textEnvelope(t, "https://example.com"),
			}},
		},
	}
	h := newViewHandlers(&fakeOnePagerReader{onePager: onePager})

	dto := getOnePager(t, h, applicationSubjectType(t))

	require.Len(t, dto.Fields, 1)
	custom := dto.Fields[0].Custom
	assert.Equal(t, testTextFieldID, custom.FieldID)
	assert.Equal(t, "Contract link", custom.Name)
	assert.Equal(t, "link", custom.Type)
	assert.Equal(t, "URL", custom.HelpText)
	assert.Equal(t, "Vendor contract", custom.DisplayText)
	require.NotNil(t, custom.Value)
	assert.Equal(t, "text", custom.Value.Type)
	assert.False(t, custom.RetiredOption)
}

func TestGetOnePager_NullValueRendersAsJSONNull(t *testing.T) {
	cases := []struct {
		name   string
		fields []queries.Field
		kind   string
	}{
		{
			name:   "built-in nil value",
			fields: []queries.Field{{BuiltIn: &queries.BuiltInField{ID: "description", Label: "Description", Value: nil}}},
			kind:   "builtIn",
		},
		{
			name:   "custom field with null value",
			fields: []queries.Field{{Custom: &queries.CustomField{FieldID: testTextFieldID, Name: "Contract link", FieldType: "link"}}},
			kind:   "custom",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			onePager := &queries.OnePager{SubjectType: "application", SubjectID: testSubjectID, Fields: tc.fields}
			h := newViewHandlers(&fakeOnePagerReader{onePager: onePager})

			fields := getOnePagerFieldsRaw(t, h)

			require.Len(t, fields, 1)
			assertFieldValueRendersJSONNull(t, fields, tc.kind)
		})
	}
}

func TestGetOnePager_CustomFieldRetiredOptionPassthrough(t *testing.T) {
	onePager := &queries.OnePager{
		SubjectType: "application",
		SubjectID:   testSubjectID,
		Fields: []queries.Field{
			{Custom: &queries.CustomField{
				FieldID:       testSelectFieldID,
				Name:          "Hosting model",
				FieldType:     "selection",
				DisplayText:   "On-prem",
				Value:         textEnvelope(t, "On-prem"),
				RetiredOption: true,
			}},
		},
	}
	h := newViewHandlers(&fakeOnePagerReader{onePager: onePager})

	dto := getOnePager(t, h, applicationSubjectType(t))

	require.Len(t, dto.Fields, 1)
	assert.True(t, dto.Fields[0].Custom.RetiredOption)
}

func TestGetOnePager_LinksForAllSubjectTypes(t *testing.T) {
	cases := []struct {
		subjectType  string
		expectedPath string
	}{
		{"capability", "/capabilities/" + testSubjectID},
		{"enterprise-capability", "/enterprise-capabilities/" + testSubjectID},
		{"application", "/components/" + testSubjectID},
		{"acquired-entity", "/acquired-entities/" + testSubjectID},
		{"vendor", "/vendors/" + testSubjectID},
		{"internal-team", "/internal-teams/" + testSubjectID},
	}

	for _, tc := range cases {
		t.Run(tc.subjectType, func(t *testing.T) {
			onePager := &queries.OnePager{SubjectType: tc.subjectType, SubjectID: testSubjectID, SubjectName: "Subject"}
			h := newViewHandlers(&fakeOnePagerReader{onePager: onePager})

			dto := getOnePager(t, h, subjectTypeFor(t, tc.subjectType))

			self, ok := dto.Links["self"]
			require.True(t, ok)
			assert.Equal(t, "/api/v1/one-pagers/"+tc.subjectType+"/"+testSubjectID, self.Href)
			assert.Equal(t, "GET", self.Method)

			subject, ok := dto.Links["x-subject"]
			require.True(t, ok)
			assert.Equal(t, "/api/v1"+tc.expectedPath, subject.Href)
			assert.Equal(t, "GET", subject.Method)
		})
	}
}

func TestGetOnePager_ErrorMapsToStatusCode(t *testing.T) {
	cases := []struct {
		name           string
		err            error
		expectedStatus int
	}{
		{"subject not found is 404", queries.ErrSubjectNotFound, http.StatusNotFound},
		{"other error is 500", errors.New("boom"), http.StatusInternalServerError},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := newViewHandlers(&fakeOnePagerReader{err: tc.err})

			rec := performGetOnePager(t, h)

			assert.Equal(t, tc.expectedStatus, rec.Code)
		})
	}
}
