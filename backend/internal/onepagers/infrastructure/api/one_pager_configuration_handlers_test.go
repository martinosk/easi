package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"easi/backend/internal/onepagers/application/commands"
	"easi/backend/internal/onepagers/application/handlers"
	"easi/backend/internal/onepagers/application/readmodels"
	"easi/backend/internal/onepagers/domain/aggregates"
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/cqrs"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeReader struct {
	bySubjectType map[string]*readmodels.ConfigurationRecord
	definitions   readmodels.CustomFieldDefinitions
	err           error
}

func newFakeReader(records ...*readmodels.ConfigurationRecord) *fakeReader {
	reader := &fakeReader{bySubjectType: map[string]*readmodels.ConfigurationRecord{}, definitions: applicationDefinitions()}
	for _, record := range records {
		reader.bySubjectType[record.SubjectType] = record
	}
	return reader
}

func (f *fakeReader) GetBySubjectType(_ context.Context, subjectType string) (*readmodels.ConfigurationRecord, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.bySubjectType[subjectType], nil
}

func (f *fakeReader) ForSubjectType(_ context.Context, _ string) (readmodels.CustomFieldDefinitions, error) {
	return f.definitions, nil
}

type fakeCommandBus struct {
	dispatched []cqrs.Command
	onDispatch func(cmd cqrs.Command) (cqrs.CommandResult, error)
}

func (f *fakeCommandBus) Dispatch(_ context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	f.dispatched = append(f.dispatched, cmd)
	if f.onDispatch != nil {
		return f.onDispatch(cmd)
	}
	return cqrs.EmptyResult(), nil
}

func (f *fakeCommandBus) Register(_ string, _ cqrs.CommandHandler) {}

type fakeSessionProvider struct {
	email string
	err   error
}

func (f *fakeSessionProvider) GetCurrentUserEmail(_ context.Context) (string, error) {
	return f.email, f.err
}

func newHandlers(reader *fakeReader, bus *fakeCommandBus) *OnePagerConfigurationHandlers {
	return NewOnePagerConfigurationHandlers(OnePagerConfigurationHandlersDeps{
		CommandBus:      bus,
		Reader:          reader,
		Definitions:     reader,
		Links:           testLinks(),
		SessionProvider: &fakeSessionProvider{email: "admin@example.com"},
	})
}

type requestSpec struct {
	method      string
	path        string
	body        any
	subjectType string
	params      map[string]string
	actor       sharedctx.Actor
}

func requestFor(t *testing.T, spec requestSpec) (*httptest.ResponseRecorder, *http.Request) {
	t.Helper()
	var bodyReader io.Reader
	if spec.body != nil {
		payload, err := json.Marshal(spec.body)
		require.NoError(t, err)
		bodyReader = bytes.NewReader(payload)
	}
	req := httptest.NewRequest(spec.method, spec.path, bodyReader)

	tenantID, err := sharedvo.NewTenantID("tenant-123")
	require.NoError(t, err)
	ctx := sharedctx.WithTenant(req.Context(), tenantID)
	ctx = sharedctx.WithActor(ctx, spec.actor)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("subjectType", spec.subjectType)
	for key, value := range spec.params {
		rctx.URLParams.Add(key, value)
	}
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)

	return httptest.NewRecorder(), req.WithContext(ctx)
}

func decodeDTO(t *testing.T, rec *httptest.ResponseRecorder) OnePagerConfigurationDTO {
	t.Helper()
	var dto OnePagerConfigurationDTO
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &dto))
	return dto
}

func TestGetConfiguration_ReturnsExistingConfiguration(t *testing.T) {
	reader := newFakeReader(applicationRecord())
	bus := &fakeCommandBus{}
	h := newHandlers(reader, bus)

	rec, req := requestFor(t, requestSpec{
		method:      http.MethodGet,
		path:        "/one-pagers/configurations/application",
		subjectType: "application",
		actor:       stakeholderActor(),
	})
	h.GetConfiguration(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	dto := decodeDTO(t, rec)
	assert.Equal(t, "config-1", dto.ID)
	assert.Empty(t, bus.dispatched)
	assert.NotContains(t, dto.Links, "x-reorder")
	assert.Contains(t, dto.Links, "x-attribute-schema")
	require.Len(t, dto.CustomFields, 3, "custom fields come from the definition cache")
}

func TestGetConfiguration_LazilyCreatesDefaultOnFirstRead(t *testing.T) {
	reader := newFakeReader()
	bus := &fakeCommandBus{}
	bus.onDispatch = func(cmd cqrs.Command) (cqrs.CommandResult, error) {
		create, ok := cmd.(*commands.CreateOnePagerConfiguration)
		require.True(t, ok)
		assert.Equal(t, "vendor", create.SubjectType)
		assert.Equal(t, "tenant-123", create.TenantID)
		assert.Equal(t, "admin@example.com", create.CreatedBy)
		record := applicationRecord()
		record.SubjectType = "vendor"
		reader.bySubjectType["vendor"] = record
		return cqrs.NewResult(record.ID), nil
	}
	h := newHandlers(reader, bus)

	rec, req := requestFor(t, requestSpec{
		method:      http.MethodGet,
		path:        "/one-pagers/configurations/vendor",
		subjectType: "vendor",
		actor:       adminActor(),
	})
	h.GetConfiguration(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	require.Len(t, bus.dispatched, 1)
	dto := decodeDTO(t, rec)
	assert.Equal(t, "vendor", dto.SubjectType)
	assert.Contains(t, dto.Links, "x-reorder")
}

func TestGetConfiguration_RecoversWhenConcurrentCreateWins(t *testing.T) {
	reader := newFakeReader()
	bus := &fakeCommandBus{}
	bus.onDispatch = func(cmd cqrs.Command) (cqrs.CommandResult, error) {
		reader.bySubjectType["application"] = applicationRecord()
		return cqrs.EmptyResult(), handlers.ErrConfigurationAlreadyExists
	}
	h := newHandlers(reader, bus)

	rec, req := requestFor(t, requestSpec{
		method:      http.MethodGet,
		path:        "/one-pagers/configurations/application",
		subjectType: "application",
		actor:       adminActor(),
	})
	h.GetConfiguration(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "config-1", decodeDTO(t, rec).ID)
}

func TestGetConfiguration_UnknownSubjectTypeIs404(t *testing.T) {
	h := newHandlers(newFakeReader(), &fakeCommandBus{})

	rec, req := requestFor(t, requestSpec{
		method:      http.MethodGet,
		path:        "/one-pagers/configurations/starship",
		subjectType: "starship",
		actor:       adminActor(),
	})
	h.GetConfiguration(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func changeRequirementRequest(t *testing.T, body map[string]any) (*httptest.ResponseRecorder, *http.Request) {
	t.Helper()
	return requestFor(t, requestSpec{
		method:      http.MethodPut,
		path:        "/one-pagers/configurations/application/custom-fields/" + testFieldID + "/requirement",
		body:        body,
		subjectType: "application",
		params:      map[string]string{"fieldID": testFieldID},
		actor:       adminActor(),
	})
}

func TestChangeCustomFieldRequirement_StaleVersionIs409WithoutDispatch(t *testing.T) {
	reader := newFakeReader(applicationRecord())
	bus := &fakeCommandBus{}
	h := newHandlers(reader, bus)

	rec, req := changeRequirementRequest(t, map[string]any{"required": true, "version": 3})
	h.ChangeCustomFieldRequirement(rec, req)

	assert.Equal(t, http.StatusConflict, rec.Code)
	assert.Empty(t, bus.dispatched)
}

func TestChangeCustomFieldRequirement_DomainConflictMapsTo409(t *testing.T) {
	reader := newFakeReader(applicationRecord())
	bus := &fakeCommandBus{onDispatch: func(cqrs.Command) (cqrs.CommandResult, error) {
		return cqrs.EmptyResult(), aggregates.ErrCustomFieldNotIncluded
	}}
	h := newHandlers(reader, bus)

	rec, req := changeRequirementRequest(t, map[string]any{"required": true, "version": 4})
	h.ChangeCustomFieldRequirement(rec, req)

	assert.Equal(t, http.StatusConflict, rec.Code)
	var errResponse sharedAPI.ErrorResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &errResponse))
	assert.Contains(t, errResponse.Message, "not included")
}

func TestReorderFields_DispatchesOrderRefs(t *testing.T) {
	reader := newFakeReader(applicationRecord())
	bus := &fakeCommandBus{}
	h := newHandlers(reader, bus)

	body := map[string]any{
		"order": []map[string]string{
			{"kind": "custom", "id": "9f0d5e69-0000-0000-0000-000000000001"},
			{"kind": "builtIn", "id": "name"},
			{"kind": "builtIn", "id": "description"},
		},
		"version": 4,
	}
	rec, req := requestFor(t, requestSpec{
		method:      http.MethodPut,
		path:        "/one-pagers/configurations/application/display-order",
		body:        body,
		subjectType: "application",
		actor:       adminActor(),
	})
	h.ReorderFields(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	require.Len(t, bus.dispatched, 1)
	cmd, ok := bus.dispatched[0].(*commands.ReorderOnePagerFields)
	require.True(t, ok)
	require.Len(t, cmd.Order, 3)
	assert.Equal(t, commands.FieldRefInput{Kind: "custom", ID: "9f0d5e69-0000-0000-0000-000000000001"}, cmd.Order[0])
}

const testFieldID = selectionFieldID

type writeEndpointCase struct {
	name     string
	invoke   func(h *OnePagerConfigurationHandlers, w http.ResponseWriter, r *http.Request)
	method   string
	path     string
	params   map[string]string
	body     map[string]any
	expected func(t *testing.T, cmd cqrs.Command)
	status   int
}

func runWriteEndpointCases(t *testing.T, cases []writeEndpointCase) {
	t.Helper()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reader := newFakeReader(applicationRecord())
			bus := &fakeCommandBus{}
			h := newHandlers(reader, bus)

			rec, req := requestFor(t, requestSpec{
				method:      tc.method,
				path:        "/one-pagers/configurations/application" + tc.path,
				body:        tc.body,
				subjectType: "application",
				params:      tc.params,
				actor:       adminActor(),
			})
			tc.invoke(h, rec, req)

			assert.Equal(t, tc.status, rec.Code, rec.Body.String())
			require.Len(t, bus.dispatched, 1)
			tc.expected(t, bus.dispatched[0])
		})
	}
}

func TestWriteEndpoints_DispatchFieldLifecycleCommands(t *testing.T) {
	runWriteEndpointCases(t, []writeEndpointCase{
		{
			name:   "requirement",
			invoke: (*OnePagerConfigurationHandlers).ChangeCustomFieldRequirement,
			method: http.MethodPut, path: "/custom-fields/" + testFieldID + "/requirement",
			params: map[string]string{"fieldID": testFieldID},
			body:   map[string]any{"required": true, "version": 4},
			expected: func(t *testing.T, cmd cqrs.Command) {
				c, ok := cmd.(*commands.ChangeCustomFieldRequirement)
				require.True(t, ok)
				assert.True(t, c.Required)
				assert.Equal(t, testFieldID, c.FieldID)
			},
			status: http.StatusOK,
		},
	})
}

func TestWriteEndpoints_DispatchBuiltInFieldCommands(t *testing.T) {
	runWriteEndpointCases(t, []writeEndpointCase{
		{
			name:   "include built-in",
			invoke: (*OnePagerConfigurationHandlers).IncludeBuiltInField,
			method: http.MethodPost, path: "/built-in-fields/experts/include",
			params: map[string]string{"entryID": "experts"},
			body:   map[string]any{"version": 4},
			expected: func(t *testing.T, cmd cqrs.Command) {
				c, ok := cmd.(*commands.IncludeBuiltInField)
				require.True(t, ok)
				assert.Equal(t, "experts", c.EntryID)
			},
			status: http.StatusOK,
		},
		{
			name:   "exclude built-in",
			invoke: (*OnePagerConfigurationHandlers).ExcludeBuiltInField,
			method: http.MethodPost, path: "/built-in-fields/experts/exclude",
			params: map[string]string{"entryID": "experts"},
			body:   map[string]any{"version": 4},
			expected: func(t *testing.T, cmd cqrs.Command) {
				c, ok := cmd.(*commands.ExcludeBuiltInField)
				require.True(t, ok)
				assert.Equal(t, "experts", c.EntryID)
			},
			status: http.StatusOK,
		},
		{
			name:   "change built-in requirement",
			invoke: (*OnePagerConfigurationHandlers).ChangeBuiltInFieldRequirement,
			method: http.MethodPut, path: "/built-in-fields/experts/requirement",
			params: map[string]string{"entryID": "experts"},
			body:   map[string]any{"required": true, "version": 4},
			expected: func(t *testing.T, cmd cqrs.Command) {
				c, ok := cmd.(*commands.ChangeBuiltInFieldRequirement)
				require.True(t, ok)
				assert.Equal(t, "experts", c.EntryID)
				assert.True(t, c.Required)
			},
			status: http.StatusOK,
		},
	})
}

func TestWriteEndpoint_UnauthenticatedIs401(t *testing.T) {
	reader := newFakeReader(applicationRecord())
	h := NewOnePagerConfigurationHandlers(OnePagerConfigurationHandlersDeps{
		CommandBus:      &fakeCommandBus{},
		Reader:          reader,
		Definitions:     reader,
		Links:           testLinks(),
		SessionProvider: &fakeSessionProvider{err: errors.New("no session")},
	})

	rec, req := changeRequirementRequest(t, map[string]any{"required": true, "version": 4})
	h.ChangeCustomFieldRequirement(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
