package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"easi/backend/internal/metamodel/application/commands"
	"easi/backend/internal/metamodel/application/handlers"
	"easi/backend/internal/metamodel/application/readmodels"
	"easi/backend/internal/metamodel/domain/aggregates"
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/cqrs"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeSchemaReader struct {
	bySubjectType map[string]*readmodels.SubjectAttributeSchemaRecord
}

func newFakeSchemaReader(records ...*readmodels.SubjectAttributeSchemaRecord) *fakeSchemaReader {
	reader := &fakeSchemaReader{bySubjectType: map[string]*readmodels.SubjectAttributeSchemaRecord{}}
	for _, record := range records {
		reader.bySubjectType[record.SubjectType] = record
	}
	return reader
}

func (f *fakeSchemaReader) GetBySubjectType(_ context.Context, subjectType string) (*readmodels.SubjectAttributeSchemaRecord, error) {
	return f.bySubjectType[subjectType], nil
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

type fakeSessionProvider struct{ email string }

func (f *fakeSessionProvider) GetCurrentUserEmail(_ context.Context) (string, error) {
	return f.email, nil
}

func adminActor() sharedctx.Actor {
	return sharedctx.NewActor("user-1", "admin@example.com", sharedctx.RoleAdmin)
}

func stakeholderActor() sharedctx.Actor {
	return sharedctx.NewActor("user-2", "viewer@example.com", sharedctx.RoleStakeholder)
}

func applicationSchema() *readmodels.SubjectAttributeSchemaRecord {
	min := 1.0
	return &readmodels.SubjectAttributeSchemaRecord{
		ID:          "schema-1",
		TenantID:    "tenant-123",
		SubjectType: "application",
		Attributes: []readmodels.SubjectAttributeRecord{
			{ID: "attr-selection", Name: "Hosting Region", Type: "selection", Active: true, Options: []readmodels.AttributeOptionRecord{
				{ID: "opt-eu", Label: "EU", Active: true}, {ID: "opt-mars", Label: "Mars", Active: false},
			}},
			{ID: "attr-number", Name: "Score", Type: "number", Active: true, Min: &min},
			{ID: "attr-retired", Name: "Old", Type: "text", Active: false},
		},
		Version:    4,
		CreatedAt:  time.Now(),
		ModifiedAt: time.Now(),
		ModifiedBy: "steward@example.com",
	}
}

func newAttributeHandlers(reader *fakeSchemaReader, bus *fakeCommandBus) *SubjectAttributeHandlers {
	return NewSubjectAttributeHandlers(bus, reader, NewMetaModelLinks(sharedAPI.NewHATEOASLinks("/api/v1")), &fakeSessionProvider{email: "steward@example.com"})
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
	ctx := sharedctx.WithActor(sharedctx.WithTenant(req.Context(), tenantID), spec.actor)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("subjectType", spec.subjectType)
	for key, value := range spec.params {
		rctx.URLParams.Add(key, value)
	}
	return httptest.NewRecorder(), req.WithContext(context.WithValue(ctx, chi.RouteCtxKey, rctx))
}

func decodeSchemaDTO(t *testing.T, rec *httptest.ResponseRecorder) SubjectAttributeSchemaDTO {
	t.Helper()
	var dto SubjectAttributeSchemaDTO
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &dto))
	return dto
}

func TestGetSchema_ReturnsExistingSchemaWithLinksGatedOnPermission(t *testing.T) {
	bus := &fakeCommandBus{}
	h := newAttributeHandlers(newFakeSchemaReader(applicationSchema()), bus)

	rec, req := requestFor(t, requestSpec{method: http.MethodGet, path: "/meta-model/subject-types/application/attributes", subjectType: "application", actor: stakeholderActor()})
	h.GetSchema(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	dto := decodeSchemaDTO(t, rec)
	assert.Equal(t, "schema-1", dto.ID)
	assert.Equal(t, 4, dto.Version)
	assert.Equal(t, "/api/v1/meta-model/subject-types/application/attributes", dto.Links["self"].Href)
	assert.Equal(t, "GET", dto.Links["self"].Method)
	assert.NotContains(t, dto.Links, "x-define")
	require.Len(t, dto.Attributes, 3)
	assert.Empty(t, dto.Attributes[0].Links)
	assert.Empty(t, bus.dispatched)
}

func TestGetSchema_AdminSeesSchemaAndAttributeLinks(t *testing.T) {
	h := newAttributeHandlers(newFakeSchemaReader(applicationSchema()), &fakeCommandBus{})

	rec, req := requestFor(t, requestSpec{method: http.MethodGet, path: "/meta-model/subject-types/application/attributes", subjectType: "application", actor: adminActor()})
	h.GetSchema(rec, req)

	dto := decodeSchemaDTO(t, rec)
	assert.Equal(t, "POST", dto.Links["x-define"].Method)
	selection := dto.Attributes[0]
	assert.Equal(t, "/api/v1/meta-model/subject-types/application/attributes/attr-selection", selection.Links["x-rename"].Href)
	assert.Equal(t, "PUT", selection.Links["x-rename"].Method)
	assert.Contains(t, selection.Links, "x-retire")
	assert.Contains(t, selection.Links, "x-add-option")
	assert.NotContains(t, selection.Links, "x-set-bounds")
	assert.Equal(t, "/api/v1/meta-model/subject-types/application/attributes/attr-selection/options/opt-eu/retire", selection.Options[0].Links["x-retire"].Href)
	assert.Empty(t, selection.Options[1].Links, "retired options carry no links")
	number := dto.Attributes[1]
	assert.Contains(t, number.Links, "x-set-bounds")
	assert.NotContains(t, number.Links, "x-add-option")
	assert.Equal(t, 1.0, *number.Min)
	retired := dto.Attributes[2]
	assert.Equal(t, map[string]bool{"x-reactivate": true}, linkNames(retired.Links))
}

func linkNames(links map[string]sharedAPI.Link) map[string]bool {
	names := map[string]bool{}
	for name := range links {
		names[name] = true
	}
	return names
}

func TestGetSchema_LazilyCreatesSchemaOnFirstRead(t *testing.T) {
	reader := newFakeSchemaReader()
	bus := &fakeCommandBus{}
	bus.onDispatch = func(cmd cqrs.Command) (cqrs.CommandResult, error) {
		create, ok := cmd.(*commands.CreateSubjectAttributeSchema)
		require.True(t, ok)
		assert.Equal(t, "vendor", create.SubjectType)
		assert.Equal(t, "tenant-123", create.TenantID)
		assert.Equal(t, "steward@example.com", create.CreatedBy)
		reader.bySubjectType["vendor"] = &readmodels.SubjectAttributeSchemaRecord{ID: "schema-v", SubjectType: "vendor", Version: 1}
		return cqrs.NewResult("schema-v"), nil
	}
	h := newAttributeHandlers(reader, bus)

	rec, req := requestFor(t, requestSpec{method: http.MethodGet, path: "/meta-model/subject-types/vendor/attributes", subjectType: "vendor", actor: adminActor()})
	h.GetSchema(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	dto := decodeSchemaDTO(t, rec)
	assert.Equal(t, "schema-v", dto.ID)
	assert.NotNil(t, dto.Attributes)
}

func TestGetSchema_RecoversWhenConcurrentCreateWins(t *testing.T) {
	reader := newFakeSchemaReader()
	bus := &fakeCommandBus{onDispatch: func(cmd cqrs.Command) (cqrs.CommandResult, error) {
		reader.bySubjectType["application"] = applicationSchema()
		return cqrs.EmptyResult(), handlers.ErrSchemaAlreadyExists
	}}
	h := newAttributeHandlers(reader, bus)

	rec, req := requestFor(t, requestSpec{method: http.MethodGet, path: "/meta-model/subject-types/application/attributes", subjectType: "application", actor: adminActor()})
	h.GetSchema(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestGetSchema_UnknownSubjectTypeIs404(t *testing.T) {
	h := newAttributeHandlers(newFakeSchemaReader(), &fakeCommandBus{})

	rec, req := requestFor(t, requestSpec{method: http.MethodGet, path: "/meta-model/subject-types/starship/attributes", subjectType: "starship", actor: adminActor()})
	h.GetSchema(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestDefineAttribute_DispatchesCommandAndReturns201(t *testing.T) {
	reader := newFakeSchemaReader(applicationSchema())
	bus := &fakeCommandBus{onDispatch: func(cmd cqrs.Command) (cqrs.CommandResult, error) {
		define, ok := cmd.(*commands.DefineSubjectAttribute)
		require.True(t, ok)
		assert.Equal(t, "schema-1", define.SchemaID)
		assert.Equal(t, "Owner Team", define.Name)
		assert.Equal(t, "selection", define.AttributeType)
		assert.Equal(t, []string{"Platform", "Data"}, define.OptionLabels)
		assert.Equal(t, "steward@example.com", define.ModifiedBy)
		return cqrs.NewResult("attr-new"), nil
	}}
	h := newAttributeHandlers(reader, bus)

	rec, req := requestFor(t, requestSpec{
		method: http.MethodPost, path: "/meta-model/subject-types/application/attributes", subjectType: "application", actor: adminActor(),
		body: map[string]any{"name": "Owner Team", "type": "selection", "helpText": "", "options": []string{"Platform", "Data"}, "version": 4},
	})
	h.DefineAttribute(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Equal(t, "/api/v1/meta-model/subject-types/application/attributes", rec.Header().Get("Location"))
	require.Len(t, bus.dispatched, 1)
	assert.Equal(t, "schema-1", decodeSchemaDTO(t, rec).ID)
}

func TestDefineAttribute_VersionMismatchIs409(t *testing.T) {
	bus := &fakeCommandBus{}
	h := newAttributeHandlers(newFakeSchemaReader(applicationSchema()), bus)

	rec, req := requestFor(t, requestSpec{
		method: http.MethodPost, path: "/meta-model/subject-types/application/attributes", subjectType: "application", actor: adminActor(),
		body: map[string]any{"name": "Owner Team", "type": "text", "version": 3},
	})
	h.DefineAttribute(rec, req)

	assert.Equal(t, http.StatusConflict, rec.Code)
	assert.Empty(t, bus.dispatched)
}

func TestDefineAttribute_DomainErrorMapsThroughRegistry(t *testing.T) {
	bus := &fakeCommandBus{onDispatch: func(cmd cqrs.Command) (cqrs.CommandResult, error) {
		return cqrs.EmptyResult(), aggregates.ErrDuplicateAttributeName
	}}
	h := newAttributeHandlers(newFakeSchemaReader(applicationSchema()), bus)

	rec, req := requestFor(t, requestSpec{
		method: http.MethodPost, path: "/meta-model/subject-types/application/attributes", subjectType: "application", actor: adminActor(),
		body: map[string]any{"name": "Score", "type": "text", "version": 4},
	})
	h.DefineAttribute(rec, req)

	assert.Equal(t, http.StatusConflict, rec.Code)
}

func TestAttributeWrites_DispatchExpectedCommands(t *testing.T) {
	min := 2.0
	cases := []struct {
		name    string
		method  string
		handler func(h *SubjectAttributeHandlers) http.HandlerFunc
		params  map[string]string
		body    map[string]any
		status  int
		command cqrs.Command
	}{
		{
			name: "rename", method: http.MethodPut, handler: func(h *SubjectAttributeHandlers) http.HandlerFunc { return h.RenameAttribute },
			params: map[string]string{"attributeID": "attr-number"},
			body:   map[string]any{"name": "Rating", "helpText": "h", "type": "number", "version": 4}, status: http.StatusOK,
			command: &commands.RenameSubjectAttribute{SchemaID: "schema-1", AttributeID: "attr-number", Name: "Rating", HelpText: "h", RequestedType: "number", ModifiedBy: "steward@example.com"},
		},
		{
			name: "retire", method: http.MethodPost, handler: func(h *SubjectAttributeHandlers) http.HandlerFunc { return h.RetireAttribute },
			params: map[string]string{"attributeID": "attr-number"}, body: map[string]any{"version": 4}, status: http.StatusOK,
			command: &commands.RetireSubjectAttribute{SchemaID: "schema-1", AttributeID: "attr-number", ModifiedBy: "steward@example.com"},
		},
		{
			name: "reactivate", method: http.MethodPost, handler: func(h *SubjectAttributeHandlers) http.HandlerFunc { return h.ReactivateAttribute },
			params: map[string]string{"attributeID": "attr-retired"}, body: map[string]any{"version": 4}, status: http.StatusOK,
			command: &commands.ReactivateSubjectAttribute{SchemaID: "schema-1", AttributeID: "attr-retired", ModifiedBy: "steward@example.com"},
		},
		{
			name: "add option", method: http.MethodPost, handler: func(h *SubjectAttributeHandlers) http.HandlerFunc { return h.AddOption },
			params: map[string]string{"attributeID": "attr-selection"}, body: map[string]any{"label": "US", "version": 4}, status: http.StatusCreated,
			command: &commands.AddSubjectAttributeOption{SchemaID: "schema-1", AttributeID: "attr-selection", Label: "US", ModifiedBy: "steward@example.com"},
		},
		{
			name: "retire option", method: http.MethodPost, handler: func(h *SubjectAttributeHandlers) http.HandlerFunc { return h.RetireOption },
			params: map[string]string{"attributeID": "attr-selection", "optionID": "opt-eu"}, body: map[string]any{"version": 4}, status: http.StatusOK,
			command: &commands.RetireSubjectAttributeOption{SchemaID: "schema-1", AttributeID: "attr-selection", OptionID: "opt-eu", ModifiedBy: "steward@example.com"},
		},
		{
			name: "set bounds", method: http.MethodPut, handler: func(h *SubjectAttributeHandlers) http.HandlerFunc { return h.SetBounds },
			params: map[string]string{"attributeID": "attr-number"}, body: map[string]any{"min": 2, "version": 4}, status: http.StatusOK,
			command: &commands.SetSubjectAttributeBounds{SchemaID: "schema-1", AttributeID: "attr-number", Min: &min, ModifiedBy: "steward@example.com"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			bus := &fakeCommandBus{}
			h := newAttributeHandlers(newFakeSchemaReader(applicationSchema()), bus)
			rec, req := requestFor(t, requestSpec{method: tc.method, path: "/meta-model/subject-types/application/attributes/x", subjectType: "application", params: tc.params, body: tc.body, actor: adminActor()})

			tc.handler(h)(rec, req)

			assert.Equal(t, tc.status, rec.Code, rec.Body.String())
			require.Len(t, bus.dispatched, 1)
			assert.Equal(t, tc.command, bus.dispatched[0])
		})
	}
}
