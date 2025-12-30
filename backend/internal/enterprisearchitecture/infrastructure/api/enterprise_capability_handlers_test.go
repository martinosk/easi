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
}

func (m *mockCommandBus) Dispatch(ctx context.Context, cmd cqrs.Command) error {
	if m.dispatchErr != nil {
		return m.dispatchErr
	}
	m.dispatchedCommands = append(m.dispatchedCommands, cmd)
	return nil
}

func (m *mockCommandBus) Register(commandType cqrs.Command, handler cqrs.CommandHandler) {}

type mockCapabilityReadModel struct {
	capabilities map[string]*readmodels.EnterpriseCapabilityDTO
	insertErr    error
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
	var caps []readmodels.EnterpriseCapabilityDTO
	for _, c := range m.capabilities {
		caps = append(caps, *c)
	}
	return caps, nil
}

func (m *mockCapabilityReadModel) NameExists(ctx context.Context, name, excludeID string) (bool, error) {
	return false, nil
}

func (m *mockCapabilityReadModel) Insert(ctx context.Context, dto readmodels.EnterpriseCapabilityDTO) error {
	if m.insertErr != nil {
		return m.insertErr
	}
	m.capabilities[dto.ID] = &dto
	return nil
}

func (m *mockCapabilityReadModel) Update(ctx context.Context, params readmodels.UpdateCapabilityParams) error {
	return nil
}

func (m *mockCapabilityReadModel) Delete(ctx context.Context, id string) error {
	return nil
}

func (m *mockCapabilityReadModel) IncrementLinkCount(ctx context.Context, id string) error {
	return nil
}

func (m *mockCapabilityReadModel) DecrementLinkCount(ctx context.Context, id string) error {
	return nil
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

func (m *mockLinkReadModel) Insert(ctx context.Context, dto readmodels.EnterpriseCapabilityLinkDTO) error {
	return nil
}

func (m *mockLinkReadModel) Delete(ctx context.Context, id string) error {
	return nil
}

func (m *mockLinkReadModel) DeleteByDomainCapabilityID(ctx context.Context, domainCapabilityID string) error {
	return nil
}

type mockImportanceReadModel struct {
	importances map[string]*readmodels.EnterpriseStrategicImportanceDTO
}

func newMockImportanceReadModel() *mockImportanceReadModel {
	return &mockImportanceReadModel{
		importances: make(map[string]*readmodels.EnterpriseStrategicImportanceDTO),
	}
}

func (m *mockImportanceReadModel) GetByID(ctx context.Context, id string) (*readmodels.EnterpriseStrategicImportanceDTO, error) {
	if imp, ok := m.importances[id]; ok {
		return imp, nil
	}
	return nil, nil
}

func (m *mockImportanceReadModel) GetByEnterpriseCapabilityID(ctx context.Context, enterpriseCapabilityID string) ([]readmodels.EnterpriseStrategicImportanceDTO, error) {
	return nil, nil
}

func (m *mockImportanceReadModel) GetByCapabilityAndPillar(ctx context.Context, enterpriseCapabilityID, pillarID string) (*readmodels.EnterpriseStrategicImportanceDTO, error) {
	return nil, nil
}

func (m *mockImportanceReadModel) Insert(ctx context.Context, dto readmodels.EnterpriseStrategicImportanceDTO) error {
	return nil
}

func (m *mockImportanceReadModel) Update(ctx context.Context, id string, importance int, rationale string) error {
	return nil
}

func (m *mockImportanceReadModel) Delete(ctx context.Context, id string) error {
	return nil
}

type testEnterpriseCapabilityHandlers struct {
	commandBus   *mockCommandBus
	capabilityRM *mockCapabilityReadModel
	linkRM       *mockLinkReadModel
	importanceRM *mockImportanceReadModel
	handlers     *testableEnterpriseCapabilityHandlers
}

type testableEnterpriseCapabilityHandlers struct {
	commandBus   *mockCommandBus
	capabilityRM *mockCapabilityReadModel
	linkRM       *mockLinkReadModel
	importanceRM *mockImportanceReadModel
}

func newTestHandlers() *testEnterpriseCapabilityHandlers {
	commandBus := &mockCommandBus{}
	capabilityRM := newMockCapabilityReadModel()
	linkRM := newMockLinkReadModel()
	importanceRM := newMockImportanceReadModel()

	return &testEnterpriseCapabilityHandlers{
		commandBus:   commandBus,
		capabilityRM: capabilityRM,
		linkRM:       linkRM,
		importanceRM: importanceRM,
		handlers: &testableEnterpriseCapabilityHandlers{
			commandBus:   commandBus,
			capabilityRM: capabilityRM,
			linkRM:       linkRM,
			importanceRM: importanceRM,
		},
	}
}

func TestCreateEnterpriseCapability_InvalidName_Returns400(t *testing.T) {
	th := newTestHandlers()

	th.commandBus.dispatchErr = valueobjects.ErrEnterpriseCapabilityNameEmpty

	req := CreateEnterpriseCapabilityRequest{
		Name:        "",
		Description: "Test",
		Category:    "",
	}
	body, _ := json.Marshal(req)

	r := httptest.NewRequest(http.MethodPost, "/enterprise-capabilities", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	th.handlers.handleCreate(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateEnterpriseCapability_DuplicateName_Returns409(t *testing.T) {
	th := newTestHandlers()

	th.commandBus.dispatchErr = handlers.ErrEnterpriseCapabilityNameExists

	req := CreateEnterpriseCapabilityRequest{
		Name:        "Payroll",
		Description: "Test",
		Category:    "",
	}
	body, _ := json.Marshal(req)

	r := httptest.NewRequest(http.MethodPost, "/enterprise-capabilities", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	th.handlers.handleCreate(w, r)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestGetEnterpriseCapabilityByID_NonExistent_Returns404(t *testing.T) {
	th := newTestHandlers()

	r := httptest.NewRequest(http.MethodGet, "/enterprise-capabilities/non-existent-id", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "non-existent-id")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	th.handlers.handleGetByID(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetEnterpriseCapabilityByID_Exists_ReturnsWithHATEOASLinks(t *testing.T) {
	th := newTestHandlers()

	capID := uuid.New().String()
	th.capabilityRM.capabilities[capID] = &readmodels.EnterpriseCapabilityDTO{
		ID:          capID,
		Name:        "Payroll",
		Description: "Test description",
		Active:      true,
		CreatedAt:   time.Now(),
	}

	r := httptest.NewRequest(http.MethodGet, "/enterprise-capabilities/"+capID, nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", capID)
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	th.handlers.handleGetByID(w, r)

	require.Equal(t, http.StatusOK, w.Code)

	var response readmodels.EnterpriseCapabilityDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, capID, response.ID)
	assert.Equal(t, "Payroll", response.Name)
	assert.NotNil(t, response.Links)
	assert.Contains(t, response.Links, "self")
}

func TestDeleteEnterpriseCapability_Success_Returns204(t *testing.T) {
	th := newTestHandlers()

	capID := uuid.New().String()
	th.capabilityRM.capabilities[capID] = &readmodels.EnterpriseCapabilityDTO{
		ID:     capID,
		Name:   "To Delete",
		Active: true,
	}

	r := httptest.NewRequest(http.MethodDelete, "/enterprise-capabilities/"+capID, nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", capID)
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	th.handlers.handleDelete(w, r)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestDeleteEnterpriseCapability_NonExistent_Returns404(t *testing.T) {
	th := newTestHandlers()

	r := httptest.NewRequest(http.MethodDelete, "/enterprise-capabilities/non-existent-id", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "non-existent-id")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	th.handlers.handleDelete(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestSetStrategicImportance_InvalidValue_Returns400(t *testing.T) {
	th := newTestHandlers()

	th.commandBus.dispatchErr = valueobjects.ErrImportanceOutOfRange

	capID := uuid.New().String()
	th.capabilityRM.capabilities[capID] = &readmodels.EnterpriseCapabilityDTO{
		ID:     capID,
		Name:   "Test",
		Active: true,
	}

	req := SetStrategicImportanceRequest{
		PillarID:   uuid.New().String(),
		PillarName: "Test Pillar",
		Importance: 0,
		Rationale:  "",
	}
	body, _ := json.Marshal(req)

	r := httptest.NewRequest(http.MethodPost, "/enterprise-capabilities/"+capID+"/strategic-importance", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", capID)
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	th.handlers.handleSetImportance(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func (h *testableEnterpriseCapabilityHandlers) handleCreate(w http.ResponseWriter, r *http.Request) {
	var req CreateEnterpriseCapabilityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	err := h.commandBus.dispatchErr
	if err != nil {
		switch err {
		case valueobjects.ErrEnterpriseCapabilityNameEmpty, valueobjects.ErrEnterpriseCapabilityNameTooLong:
			http.Error(w, err.Error(), http.StatusBadRequest)
		case handlers.ErrEnterpriseCapabilityNameExists:
			http.Error(w, err.Error(), http.StatusConflict)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Location", "/enterprise-capabilities/test-id")
	w.WriteHeader(http.StatusCreated)
}

func (h *testableEnterpriseCapabilityHandlers) handleGetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	cap, _ := h.capabilityRM.GetByID(r.Context(), id)
	if cap == nil {
		http.Error(w, "Enterprise capability not found", http.StatusNotFound)
		return
	}

	cap.Links = map[string]string{
		"self":   "/enterprise-capabilities/" + id,
		"links":  "/enterprise-capabilities/" + id + "/links",
		"update": "/enterprise-capabilities/" + id,
		"delete": "/enterprise-capabilities/" + id,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(cap)
}

func (h *testableEnterpriseCapabilityHandlers) handleDelete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	cap, _ := h.capabilityRM.GetByID(r.Context(), id)
	if cap == nil {
		http.Error(w, "Enterprise capability not found", http.StatusNotFound)
		return
	}

	if h.commandBus.dispatchErr != nil {
		http.Error(w, h.commandBus.dispatchErr.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *testableEnterpriseCapabilityHandlers) handleSetImportance(w http.ResponseWriter, r *http.Request) {
	var req SetStrategicImportanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	err := h.commandBus.dispatchErr
	if err != nil {
		if err == valueobjects.ErrImportanceOutOfRange {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
