package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"easi/backend/internal/enterprisearchitecture/application/handlers"
	"easi/backend/internal/enterprisearchitecture/application/readmodels"
	"easi/backend/internal/enterprisearchitecture/domain/valueobjects"
	"easi/backend/internal/shared/cqrs"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockCommandBus struct {
	dispatchedCommands []cqrs.Command
	dispatchErr        error
	createdID          string
}

func (m *mockCommandBus) Dispatch(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	if m.dispatchErr != nil {
		return cqrs.EmptyResult(), m.dispatchErr
	}
	m.dispatchedCommands = append(m.dispatchedCommands, cmd)
	return cqrs.CommandResult{CreatedID: m.createdID}, nil
}

func (m *mockCommandBus) Register(commandName string, handler cqrs.CommandHandler) {}

type mockSessionProvider struct {
	email string
	err   error
}

func (m *mockSessionProvider) GetCurrentUserEmail(ctx context.Context) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.email, nil
}

type mockCapabilityReadModel struct {
	capabilities map[string]*readmodels.EnterpriseCapabilityDTO
}

func newMockCapabilityReadModel() *mockCapabilityReadModel {
	return &mockCapabilityReadModel{
		capabilities: make(map[string]*readmodels.EnterpriseCapabilityDTO),
	}
}

func (m *mockCapabilityReadModel) GetByID(ctx context.Context, id string) (*readmodels.EnterpriseCapabilityDTO, error) {
	if cap, ok := m.capabilities[id]; ok {
		return cap, nil
	}
	return nil, nil
}

func (m *mockCapabilityReadModel) GetAll(ctx context.Context) ([]readmodels.EnterpriseCapabilityDTO, error) {
	caps := make([]readmodels.EnterpriseCapabilityDTO, 0, len(m.capabilities))
	for _, c := range m.capabilities {
		caps = append(caps, *c)
	}
	return caps, nil
}

type mockLinkReadModel struct {
	links map[string]*readmodels.EnterpriseCapabilityLinkDTO
}

func newMockLinkReadModel() *mockLinkReadModel {
	return &mockLinkReadModel{
		links: make(map[string]*readmodels.EnterpriseCapabilityLinkDTO),
	}
}

func (m *mockLinkReadModel) GetByID(ctx context.Context, id string) (*readmodels.EnterpriseCapabilityLinkDTO, error) {
	if link, ok := m.links[id]; ok {
		return link, nil
	}
	return nil, nil
}

func (m *mockLinkReadModel) GetByEnterpriseCapabilityID(ctx context.Context, enterpriseCapabilityID string) ([]readmodels.EnterpriseCapabilityLinkDTO, error) {
	return nil, nil
}

func (m *mockLinkReadModel) GetByDomainCapabilityID(ctx context.Context, domainCapabilityID string) (*readmodels.EnterpriseCapabilityLinkDTO, error) {
	return nil, nil
}

func (m *mockLinkReadModel) GetLinkStatus(ctx context.Context, domainCapabilityID string) (*readmodels.CapabilityLinkStatusDTO, error) {
	return nil, nil
}

func (m *mockLinkReadModel) GetBatchLinkStatus(ctx context.Context, domainCapabilityIDs []string) ([]readmodels.CapabilityLinkStatusDTO, error) {
	return nil, nil
}

type mockImportanceReadModel struct{}

func (m *mockImportanceReadModel) GetByID(ctx context.Context, id string) (*readmodels.EnterpriseStrategicImportanceDTO, error) {
	return nil, nil
}

func (m *mockImportanceReadModel) GetByEnterpriseCapabilityID(ctx context.Context, enterpriseCapabilityID string) ([]readmodels.EnterpriseStrategicImportanceDTO, error) {
	return nil, nil
}

type mockMaturityAnalysisReadModel struct{}

func (m *mockMaturityAnalysisReadModel) GetMaturityAnalysisCandidates(ctx context.Context, sortBy string) ([]readmodels.MaturityAnalysisCandidateDTO, readmodels.MaturityAnalysisSummaryDTO, error) {
	return nil, readmodels.MaturityAnalysisSummaryDTO{}, nil
}

func (m *mockMaturityAnalysisReadModel) GetMaturityGapDetail(ctx context.Context, enterpriseCapabilityID string) (*readmodels.MaturityGapDetailDTO, error) {
	return nil, nil
}

type testHarness struct {
	commandBus      *mockCommandBus
	capabilityRM    *mockCapabilityReadModel
	linkRM          *mockLinkReadModel
	importanceRM    *mockImportanceReadModel
	sessionProvider *mockSessionProvider
	handlers        *EnterpriseCapabilityHandlers
}

func newTestHarness() *testHarness {
	commandBus := &mockCommandBus{}
	capabilityRM := newMockCapabilityReadModel()
	linkRM := newMockLinkReadModel()
	importanceRM := &mockImportanceReadModel{}
	sessionProvider := &mockSessionProvider{email: "test@example.com"}

	rm := &EnterpriseCapabilityReadModels{
		Capability:       capabilityRM,
		Link:             linkRM,
		Importance:       importanceRM,
		MaturityAnalysis: &mockMaturityAnalysisReadModel{},
	}

	return &testHarness{
		commandBus:      commandBus,
		capabilityRM:    capabilityRM,
		linkRM:          linkRM,
		importanceRM:    importanceRM,
		sessionProvider: sessionProvider,
		handlers:        NewEnterpriseCapabilityHandlers(commandBus, rm, sessionProvider),
	}
}

type requestSpec struct {
	method string
	path   string
	id     string
	body   []byte
}

func (s requestSpec) build() *http.Request {
	var bodyReader *bytes.Reader
	if s.body != nil {
		bodyReader = bytes.NewReader(s.body)
	}
	var r *http.Request
	if bodyReader != nil {
		r = httptest.NewRequest(s.method, s.path, bodyReader)
		r.Header.Set("Content-Type", "application/json")
	} else {
		r = httptest.NewRequest(s.method, s.path, nil)
	}
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", s.id)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func TestCreateEnterpriseCapability_ErrorResponses(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		reqName    string
		wantStatus int
	}{
		{"invalid name returns 400", valueobjects.ErrEnterpriseCapabilityNameEmpty, "", http.StatusBadRequest},
		{"duplicate name returns 409", handlers.ErrEnterpriseCapabilityNameExists, "Payroll", http.StatusConflict},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			th := newTestHarness()
			th.commandBus.dispatchErr = tt.err

			body, _ := json.Marshal(CreateEnterpriseCapabilityRequest{
				Name:        tt.reqName,
				Description: "Test",
			})

			r := httptest.NewRequest(http.MethodPost, "/enterprise-capabilities", bytes.NewReader(body))
			r.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			th.handlers.CreateEnterpriseCapability(w, r)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestNonExistentCapability_Returns404(t *testing.T) {
	tests := []struct {
		name      string
		method    string
		handlerFn func(*EnterpriseCapabilityHandlers) http.HandlerFunc
	}{
		{"GET by ID", http.MethodGet, func(h *EnterpriseCapabilityHandlers) http.HandlerFunc { return h.GetEnterpriseCapabilityByID }},
		{"DELETE", http.MethodDelete, func(h *EnterpriseCapabilityHandlers) http.HandlerFunc { return h.DeleteEnterpriseCapability }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			th := newTestHarness()
			r := requestSpec{method: tt.method, path: "/enterprise-capabilities/non-existent-id", id: "non-existent-id"}.build()
			w := httptest.NewRecorder()

			tt.handlerFn(th.handlers)(w, r)

			assert.Equal(t, http.StatusNotFound, w.Code)
		})
	}
}

func TestGetEnterpriseCapabilityByID_Exists_ReturnsWithHATEOASLinks(t *testing.T) {
	th := newTestHarness()

	capID := uuid.New().String()
	th.capabilityRM.capabilities[capID] = &readmodels.EnterpriseCapabilityDTO{
		ID:          capID,
		Name:        "Payroll",
		Description: "Test description",
		Active:      true,
		CreatedAt:   time.Now(),
	}

	r := requestSpec{method: http.MethodGet, path: "/enterprise-capabilities/" + capID, id: capID}.build()
	w := httptest.NewRecorder()

	th.handlers.GetEnterpriseCapabilityByID(w, r)

	require.Equal(t, http.StatusOK, w.Code)

	var response readmodels.EnterpriseCapabilityDTO
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))

	assert.Equal(t, capID, response.ID)
	assert.Equal(t, "Payroll", response.Name)
	require.NotNil(t, response.Links)
	assert.Contains(t, response.Links, "self")
}

func TestDeleteEnterpriseCapability_Success_Returns204(t *testing.T) {
	th := newTestHarness()

	capID := uuid.New().String()
	th.capabilityRM.capabilities[capID] = &readmodels.EnterpriseCapabilityDTO{
		ID:     capID,
		Name:   "To Delete",
		Active: true,
	}

	r := requestSpec{method: http.MethodDelete, path: "/enterprise-capabilities/" + capID, id: capID}.build()
	w := httptest.NewRecorder()

	th.handlers.DeleteEnterpriseCapability(w, r)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestSetStrategicImportance_InvalidValue_Returns400(t *testing.T) {
	th := newTestHarness()
	th.commandBus.dispatchErr = valueobjects.ErrImportanceOutOfRange

	capID := uuid.New().String()
	th.capabilityRM.capabilities[capID] = &readmodels.EnterpriseCapabilityDTO{
		ID:     capID,
		Name:   "Test",
		Active: true,
	}

	body, _ := json.Marshal(SetStrategicImportanceRequest{
		PillarID:   uuid.New().String(),
		PillarName: "Test Pillar",
		Importance: 0,
		Rationale:  "",
	})

	r := requestSpec{method: http.MethodPost, path: "/enterprise-capabilities/" + capID + "/strategic-importance", id: capID, body: body}.build()
	w := httptest.NewRecorder()

	th.handlers.SetStrategicImportance(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLinkCapability_ConflictErrors_Return409(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"domain capability already linked", handlers.ErrDomainCapabilityAlreadyLinked},
		{"ancestor linked to different", handlers.ErrAncestorLinkedToDifferent},
		{"descendant linked to different", handlers.ErrDescendantLinkedToDifferent},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			th := newTestHarness()
			th.commandBus.dispatchErr = tt.err

			capID := uuid.New().String()
			th.capabilityRM.capabilities[capID] = &readmodels.EnterpriseCapabilityDTO{
				ID:     capID,
				Name:   "Test Capability",
				Active: true,
			}

			body, _ := json.Marshal(LinkCapabilityRequest{
				DomainCapabilityID: uuid.New().String(),
			})

			r := requestSpec{method: http.MethodPost, path: "/enterprise-capabilities/" + capID + "/links", id: capID, body: body}.build()
			w := httptest.NewRecorder()

			th.handlers.LinkCapability(w, r)

			assert.Equal(t, http.StatusConflict, w.Code)
		})
	}
}

func TestLinkCapability_Success_Returns201WithLocation(t *testing.T) {
	th := newTestHarness()

	capID := uuid.New().String()
	createdLinkID := uuid.New().String()
	th.commandBus.createdID = createdLinkID
	th.capabilityRM.capabilities[capID] = &readmodels.EnterpriseCapabilityDTO{
		ID:     capID,
		Name:   "Test Capability",
		Active: true,
	}
	th.linkRM.links[createdLinkID] = &readmodels.EnterpriseCapabilityLinkDTO{
		ID:                     createdLinkID,
		EnterpriseCapabilityID: capID,
		DomainCapabilityID:     uuid.New().String(),
		LinkedBy:               "test@example.com",
		LinkedAt:               time.Now(),
	}

	body, _ := json.Marshal(LinkCapabilityRequest{
		DomainCapabilityID: uuid.New().String(),
	})

	r := requestSpec{method: http.MethodPost, path: "/enterprise-capabilities/" + capID + "/links", id: capID, body: body}.build()
	w := httptest.NewRecorder()

	th.handlers.LinkCapability(w, r)

	assert.Equal(t, http.StatusCreated, w.Code)
	location := w.Header().Get("Location")
	assert.Contains(t, location, "/enterprise-capabilities/"+capID+"/links/"+createdLinkID)
}
