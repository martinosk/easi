package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"easi/backend/internal/onepagers/application/commands"
	"easi/backend/internal/onepagers/application/handlers"
	"easi/backend/internal/onepagers/application/readmodels"
	"easi/backend/internal/onepagers/domain/valueobjects"
	"easi/backend/internal/shared/cqrs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testSubjectID     = "app-1"
	testTextFieldID   = "9f0d5e69-0000-0000-0000-000000000003"
	testSelectFieldID = "9f0d5e69-0000-0000-0000-000000000001"
	retiredOptionID   = "9f0d5e69-0000-0000-0000-00000000000a"
)

type fakeFactsReader struct {
	records []readmodels.FactRecord
	err     error
}

func (f *fakeFactsReader) GetForSubject(_ context.Context, _ readmodels.SubjectKey) ([]readmodels.FactRecord, error) {
	return f.records, f.err
}

func textFactRecord(t *testing.T) readmodels.FactRecord {
	t.Helper()
	value, err := valueobjects.NewTextValue("Runs on shared Kubernetes cluster")
	require.NoError(t, err)
	envelope, err := valueobjects.NewValueEnvelope(value)
	require.NoError(t, err)
	return readmodels.FactRecord{
		FactsID:     "facts-1",
		TenantID:    "tenant-123",
		SubjectType: "application",
		SubjectID:   testSubjectID,
		FieldID:     testTextFieldID,
		Value:       &envelope,
		ValueType:   "text",
		DisplayText: "Runs on shared Kubernetes cluster",
		ModifiedAt:  time.Now().UTC(),
		ModifiedBy:  "architect@example.com",
	}
}

func retiredOptionFactRecord(t *testing.T) readmodels.FactRecord {
	t.Helper()
	value, err := valueobjects.NewSelectionValue(retiredOptionID)
	require.NoError(t, err)
	envelope, err := valueobjects.NewValueEnvelope(value)
	require.NoError(t, err)
	record := textFactRecord(t)
	record.FieldID = testSelectFieldID
	record.Value = &envelope
	record.ValueType = "selection"
	record.DisplayText = retiredOptionID
	return record
}

func newFactsHandlers(reader *fakeFactsReader, configs *fakeReader, bus *fakeCommandBus) *OnePagerFactsHandlers {
	return NewOnePagerFactsHandlers(OnePagerFactsHandlersDeps{
		CommandBus:      bus,
		Facts:           reader,
		Configs:         configs,
		Links:           testLinks(),
		SessionProvider: &fakeSessionProvider{email: "architect@example.com"},
	})
}

func decodeFactsDTO(t *testing.T, body []byte) OnePagerFactsDTO {
	t.Helper()
	var dto OnePagerFactsDTO
	require.NoError(t, json.Unmarshal(body, &dto))
	return dto
}

func applicationSubjectType(t *testing.T) valueobjects.SubjectType {
	t.Helper()
	subjectType, err := valueobjects.NewSubjectType("application")
	require.NoError(t, err)
	return subjectType
}

func TestGetFacts_ReturnsRecordedValuesWithSelfLink(t *testing.T) {
	h := newFactsHandlers(&fakeFactsReader{records: []readmodels.FactRecord{textFactRecord(t)}}, newFakeReader(applicationRecord()), &fakeCommandBus{})

	rec, req := requestFor(t, requestSpec{
		method: http.MethodGet,
		path:   "/one-pagers/application/" + testSubjectID + "/facts",
		params: map[string]string{"subjectID": testSubjectID},
		actor:  stakeholderActor(),
	})
	h.GetFacts(applicationSubjectType(t))(rec, req)

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	dto := decodeFactsDTO(t, rec.Body.Bytes())
	assert.Equal(t, "application", dto.SubjectType)
	assert.Equal(t, testSubjectID, dto.SubjectID)
	require.Len(t, dto.Values, 1)
	assert.Equal(t, testTextFieldID, dto.Values[0].FieldID)
	assert.Equal(t, "text", dto.Values[0].Value.Type)
	assert.Equal(t, 1, dto.Values[0].Value.Version)
	assert.False(t, dto.Values[0].RetiredOption)

	self, ok := dto.Links["self"]
	require.True(t, ok)
	assert.Equal(t, "/api/v1/one-pagers/application/"+testSubjectID+"/facts", self.Href)
	assert.Equal(t, "GET", self.Method)
	assert.NotContains(t, dto.Links, "x-record")
	assert.Empty(t, dto.Values[0].Links)
}

func TestGetFacts_FlagsRetiredSelectionOption(t *testing.T) {
	h := newFactsHandlers(&fakeFactsReader{records: []readmodels.FactRecord{retiredOptionFactRecord(t)}}, newFakeReader(applicationRecord()), &fakeCommandBus{})

	rec, req := requestFor(t, requestSpec{
		method: http.MethodGet,
		path:   "/one-pagers/application/" + testSubjectID + "/facts",
		params: map[string]string{"subjectID": testSubjectID},
		actor:  stakeholderActor(),
	})
	h.GetFacts(applicationSubjectType(t))(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	dto := decodeFactsDTO(t, rec.Body.Bytes())
	require.Len(t, dto.Values, 1)
	assert.True(t, dto.Values[0].RetiredOption)
}

func TestGetFacts_WriteAffordancesForAdmin(t *testing.T) {
	h := newFactsHandlers(&fakeFactsReader{records: []readmodels.FactRecord{textFactRecord(t)}}, newFakeReader(applicationRecord()), &fakeCommandBus{})

	rec, req := requestFor(t, requestSpec{
		method: http.MethodGet,
		path:   "/one-pagers/application/" + testSubjectID + "/facts",
		params: map[string]string{"subjectID": testSubjectID},
		actor:  adminActor(),
	})
	h.GetFacts(applicationSubjectType(t))(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	dto := decodeFactsDTO(t, rec.Body.Bytes())

	record, ok := dto.Links["x-record"]
	require.True(t, ok)
	assert.Equal(t, "PUT", record.Method)

	fieldBase := "/api/v1/one-pagers/application/" + testSubjectID + "/facts/" + testTextFieldID
	value := dto.Values[0]
	assert.Equal(t, fieldBase, value.Links["x-record"].Href)
	assert.Equal(t, "PUT", value.Links["x-record"].Method)
	assert.Equal(t, fieldBase, value.Links["x-clear"].Href)
	assert.Equal(t, "DELETE", value.Links["x-clear"].Method)
}

func TestRecordValue_DispatchesCommandAndReturnsFacts(t *testing.T) {
	bus := &fakeCommandBus{}
	h := newFactsHandlers(&fakeFactsReader{records: []readmodels.FactRecord{textFactRecord(t)}}, newFakeReader(applicationRecord()), bus)

	rec, req := requestFor(t, requestSpec{
		method: http.MethodPut,
		path:   "/one-pagers/application/" + testSubjectID + "/facts/" + testTextFieldID,
		body: map[string]any{
			"value": map[string]any{"type": "text", "version": 1, "value": "Runs on shared Kubernetes cluster"},
		},
		params: map[string]string{"subjectID": testSubjectID, "fieldID": testTextFieldID},
		actor:  adminActor(),
	})
	h.RecordValue(applicationSubjectType(t))(rec, req)

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.Len(t, bus.dispatched, 1)
	cmd, ok := bus.dispatched[0].(*commands.RecordFieldValue)
	require.True(t, ok)
	assert.Equal(t, "tenant-123", cmd.TenantID)
	assert.Equal(t, "application", cmd.SubjectType)
	assert.Equal(t, testSubjectID, cmd.SubjectID)
	assert.Equal(t, testTextFieldID, cmd.FieldID)
	assert.Equal(t, "text", cmd.Value.Type)
	assert.Equal(t, "architect@example.com", cmd.ModifiedBy)

	dto := decodeFactsDTO(t, rec.Body.Bytes())
	assert.Len(t, dto.Values, 1)
}

func TestRecordValue_MapsDomainErrorsToStatusCodes(t *testing.T) {
	cases := []struct {
		name   string
		err    error
		status int
	}{
		{"unknown field", handlers.ErrFieldNotDefined, http.StatusNotFound},
		{"missing subject", handlers.ErrSubjectNotFound, http.StatusNotFound},
		{"type mismatch", handlers.ErrValueTypeMismatch, http.StatusBadRequest},
		{"retired field", handlers.ErrFieldDefinitionRetired, http.StatusConflict},
		{"retired option", handlers.ErrOptionRetired, http.StatusConflict},
		{"invalid value", valueobjects.ErrTextValueEmpty, http.StatusBadRequest},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			bus := &fakeCommandBus{onDispatch: func(cqrs.Command) (cqrs.CommandResult, error) {
				return cqrs.EmptyResult(), tc.err
			}}
			h := newFactsHandlers(&fakeFactsReader{}, newFakeReader(applicationRecord()), bus)

			rec, req := requestFor(t, requestSpec{
				method: http.MethodPut,
				path:   "/one-pagers/application/" + testSubjectID + "/facts/" + testTextFieldID,
				body: map[string]any{
					"value": map[string]any{"type": "text", "version": 1, "value": "v"},
				},
				params: map[string]string{"subjectID": testSubjectID, "fieldID": testTextFieldID},
				actor:  adminActor(),
			})
			h.RecordValue(applicationSubjectType(t))(rec, req)

			assert.Equal(t, tc.status, rec.Code, rec.Body.String())
		})
	}
}

func TestClearValue_DispatchesCommand(t *testing.T) {
	bus := &fakeCommandBus{}
	h := newFactsHandlers(&fakeFactsReader{}, newFakeReader(applicationRecord()), bus)

	rec, req := requestFor(t, requestSpec{
		method: http.MethodDelete,
		path:   "/one-pagers/application/" + testSubjectID + "/facts/" + testTextFieldID,
		params: map[string]string{"subjectID": testSubjectID, "fieldID": testTextFieldID},
		actor:  adminActor(),
	})
	h.ClearValue(applicationSubjectType(t))(rec, req)

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.Len(t, bus.dispatched, 1)
	cmd, ok := bus.dispatched[0].(*commands.ClearFieldValue)
	require.True(t, ok)
	assert.Equal(t, testTextFieldID, cmd.FieldID)
	assert.Equal(t, testSubjectID, cmd.SubjectID)
}

func TestFactsWriteEndpoints_UnauthenticatedIs401(t *testing.T) {
	h := NewOnePagerFactsHandlers(OnePagerFactsHandlersDeps{
		CommandBus:      &fakeCommandBus{},
		Facts:           &fakeFactsReader{},
		Configs:         newFakeReader(applicationRecord()),
		Links:           testLinks(),
		SessionProvider: &fakeSessionProvider{err: errors.New("no session")},
	})

	rec, req := requestFor(t, requestSpec{
		method: http.MethodPut,
		path:   "/one-pagers/application/" + testSubjectID + "/facts/" + testTextFieldID,
		body: map[string]any{
			"value": map[string]any{"type": "text", "version": 1, "value": "v"},
		},
		params: map[string]string{"subjectID": testSubjectID, "fieldID": testTextFieldID},
		actor:  adminActor(),
	})
	h.RecordValue(applicationSubjectType(t))(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestGetFacts_EmptySubjectIDIs404(t *testing.T) {
	h := newFactsHandlers(&fakeFactsReader{}, newFakeReader(applicationRecord()), &fakeCommandBus{})

	rec, req := requestFor(t, requestSpec{
		method: http.MethodGet,
		path:   "/one-pagers/application//facts",
		params: map[string]string{"subjectID": ""},
		actor:  adminActor(),
	})
	h.GetFacts(applicationSubjectType(t))(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}
