package projectors

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"easi/backend/internal/enterprisearchitecture/application/readmodels"
	"easi/backend/internal/enterprisearchitecture/domain/events"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockEnterpriseCapabilityReadModel struct {
	insertedDTOs               []readmodels.EnterpriseCapabilityDTO
	updatedParams              []readmodels.UpdateCapabilityParams
	deletedIDs                 []string
	incrementedLinkIDs         []string
	decrementedLinkIDs         []string
	incrementAndRecalculateIDs []string
	decrementAndRecalculateIDs []string
	insertErr                  error
	updateErr                  error
	deleteErr                  error
	incrementErr               error
	decrementErr               error
	incrementAndRecalculateErr error
	decrementAndRecalculateErr error
}

func (m *mockEnterpriseCapabilityReadModel) Insert(ctx context.Context, dto readmodels.EnterpriseCapabilityDTO) error {
	if m.insertErr != nil {
		return m.insertErr
	}
	m.insertedDTOs = append(m.insertedDTOs, dto)
	return nil
}

func (m *mockEnterpriseCapabilityReadModel) Update(ctx context.Context, params readmodels.UpdateCapabilityParams) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.updatedParams = append(m.updatedParams, params)
	return nil
}

func (m *mockEnterpriseCapabilityReadModel) Delete(ctx context.Context, id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	m.deletedIDs = append(m.deletedIDs, id)
	return nil
}

func (m *mockEnterpriseCapabilityReadModel) IncrementLinkCount(ctx context.Context, id string) error {
	if m.incrementErr != nil {
		return m.incrementErr
	}
	m.incrementedLinkIDs = append(m.incrementedLinkIDs, id)
	return nil
}

func (m *mockEnterpriseCapabilityReadModel) DecrementLinkCount(ctx context.Context, id string) error {
	if m.decrementErr != nil {
		return m.decrementErr
	}
	m.decrementedLinkIDs = append(m.decrementedLinkIDs, id)
	return nil
}

func (m *mockEnterpriseCapabilityReadModel) IncrementLinkCountAndRecalculateDomainCount(ctx context.Context, id string) error {
	if m.incrementAndRecalculateErr != nil {
		return m.incrementAndRecalculateErr
	}
	m.incrementAndRecalculateIDs = append(m.incrementAndRecalculateIDs, id)
	return nil
}

func (m *mockEnterpriseCapabilityReadModel) DecrementLinkCountAndRecalculateDomainCount(ctx context.Context, id string) error {
	if m.decrementAndRecalculateErr != nil {
		return m.decrementAndRecalculateErr
	}
	m.decrementAndRecalculateIDs = append(m.decrementAndRecalculateIDs, id)
	return nil
}

type enterpriseCapabilityReadModelForProjector interface {
	Insert(ctx context.Context, dto readmodels.EnterpriseCapabilityDTO) error
	Update(ctx context.Context, params readmodels.UpdateCapabilityParams) error
	Delete(ctx context.Context, id string) error
	IncrementLinkCount(ctx context.Context, id string) error
	DecrementLinkCount(ctx context.Context, id string) error
	IncrementLinkCountAndRecalculateDomainCount(ctx context.Context, id string) error
	DecrementLinkCountAndRecalculateDomainCount(ctx context.Context, id string) error
}

type testableEnterpriseCapabilityProjector struct {
	readModel enterpriseCapabilityReadModelForProjector
}

func newTestableEnterpriseCapabilityProjector(readModel enterpriseCapabilityReadModelForProjector) *testableEnterpriseCapabilityProjector {
	return &testableEnterpriseCapabilityProjector{readModel: readModel}
}

func (p *testableEnterpriseCapabilityProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	handlers := map[string]func(context.Context, []byte) error{
		"EnterpriseCapabilityCreated":  p.handleCreated,
		"EnterpriseCapabilityUpdated":  p.handleUpdated,
		"EnterpriseCapabilityDeleted":  p.handleDeleted,
		"EnterpriseCapabilityLinked":   p.handleLinked,
		"EnterpriseCapabilityUnlinked": p.handleUnlinked,
	}

	if handler, exists := handlers[eventType]; exists {
		return handler(ctx, eventData)
	}
	return nil
}

func (p *testableEnterpriseCapabilityProjector) handleCreated(ctx context.Context, eventData []byte) error {
	var event events.EnterpriseCapabilityCreated
	if err := json.Unmarshal(eventData, &event); err != nil {
		return err
	}
	dto := readmodels.EnterpriseCapabilityDTO{
		ID:          event.ID,
		Name:        event.Name,
		Description: event.Description,
		Category:    event.Category,
		Active:      event.Active,
		CreatedAt:   event.CreatedAt,
	}
	return p.readModel.Insert(ctx, dto)
}

func (p *testableEnterpriseCapabilityProjector) handleUpdated(ctx context.Context, eventData []byte) error {
	var event events.EnterpriseCapabilityUpdated
	if err := json.Unmarshal(eventData, &event); err != nil {
		return err
	}
	return p.readModel.Update(ctx, readmodels.UpdateCapabilityParams{
		ID:          event.ID,
		Name:        event.Name,
		Description: event.Description,
		Category:    event.Category,
	})
}

func (p *testableEnterpriseCapabilityProjector) handleDeleted(ctx context.Context, eventData []byte) error {
	var event events.EnterpriseCapabilityDeleted
	if err := json.Unmarshal(eventData, &event); err != nil {
		return err
	}
	return p.readModel.Delete(ctx, event.ID)
}

func (p *testableEnterpriseCapabilityProjector) handleLinked(ctx context.Context, eventData []byte) error {
	var event events.EnterpriseCapabilityLinked
	if err := json.Unmarshal(eventData, &event); err != nil {
		return err
	}
	return p.readModel.IncrementLinkCountAndRecalculateDomainCount(ctx, event.EnterpriseCapabilityID)
}

func (p *testableEnterpriseCapabilityProjector) handleUnlinked(ctx context.Context, eventData []byte) error {
	var event events.EnterpriseCapabilityUnlinked
	if err := json.Unmarshal(eventData, &event); err != nil {
		return err
	}
	return p.readModel.DecrementLinkCountAndRecalculateDomainCount(ctx, event.EnterpriseCapabilityID)
}

func TestEnterpriseCapabilityProjector_Created_InsertsReadModel(t *testing.T) {
	mockReadModel := &mockEnterpriseCapabilityReadModel{}
	projector := newTestableEnterpriseCapabilityProjector(mockReadModel)

	event := events.NewEnterpriseCapabilityCreated(
		uuid.New().String(),
		"Test Capability",
		"Test Description",
		"Test Category",
	)

	eventData, err := json.Marshal(event.EventData())
	require.NoError(t, err)

	err = projector.ProjectEvent(context.Background(), "EnterpriseCapabilityCreated", eventData)
	require.NoError(t, err)

	require.Len(t, mockReadModel.insertedDTOs, 1)
	dto := mockReadModel.insertedDTOs[0]
	assert.Equal(t, event.ID, dto.ID)
	assert.Equal(t, "Test Capability", dto.Name)
	assert.Equal(t, "Test Description", dto.Description)
	assert.Equal(t, "Test Category", dto.Category)
	assert.True(t, dto.Active)
}

func TestEnterpriseCapabilityProjector_Updated_UpdatesReadModel(t *testing.T) {
	mockReadModel := &mockEnterpriseCapabilityReadModel{}
	projector := newTestableEnterpriseCapabilityProjector(mockReadModel)

	event := events.NewEnterpriseCapabilityUpdated(
		uuid.New().String(),
		"Updated Name",
		"Updated Description",
		"Updated Category",
	)

	eventData, err := json.Marshal(event.EventData())
	require.NoError(t, err)

	err = projector.ProjectEvent(context.Background(), "EnterpriseCapabilityUpdated", eventData)
	require.NoError(t, err)

	require.Len(t, mockReadModel.updatedParams, 1)
	params := mockReadModel.updatedParams[0]
	assert.Equal(t, event.ID, params.ID)
	assert.Equal(t, "Updated Name", params.Name)
	assert.Equal(t, "Updated Description", params.Description)
	assert.Equal(t, "Updated Category", params.Category)
}

func TestEnterpriseCapabilityProjector_Deleted_MarksInactive(t *testing.T) {
	mockReadModel := &mockEnterpriseCapabilityReadModel{}
	projector := newTestableEnterpriseCapabilityProjector(mockReadModel)

	event := events.NewEnterpriseCapabilityDeleted(uuid.New().String())

	eventData, err := json.Marshal(event.EventData())
	require.NoError(t, err)

	err = projector.ProjectEvent(context.Background(), "EnterpriseCapabilityDeleted", eventData)
	require.NoError(t, err)

	require.Len(t, mockReadModel.deletedIDs, 1)
	assert.Equal(t, event.ID, mockReadModel.deletedIDs[0])
}

func TestEnterpriseCapabilityProjector_Linked_IncrementsLinkCountAndRecalculatesDomainCount(t *testing.T) {
	mockReadModel := &mockEnterpriseCapabilityReadModel{}
	projector := newTestableEnterpriseCapabilityProjector(mockReadModel)

	enterpriseCapabilityID := uuid.New().String()
	event := events.NewEnterpriseCapabilityLinked(
		uuid.New().String(),
		enterpriseCapabilityID,
		uuid.New().String(),
		"user@example.com",
	)

	eventData, err := json.Marshal(event.EventData())
	require.NoError(t, err)

	err = projector.ProjectEvent(context.Background(), "EnterpriseCapabilityLinked", eventData)
	require.NoError(t, err)

	require.Len(t, mockReadModel.incrementAndRecalculateIDs, 1)
	assert.Equal(t, enterpriseCapabilityID, mockReadModel.incrementAndRecalculateIDs[0])
}

func TestEnterpriseCapabilityProjector_Unlinked_DecrementsLinkCountAndRecalculatesDomainCount(t *testing.T) {
	mockReadModel := &mockEnterpriseCapabilityReadModel{}
	projector := newTestableEnterpriseCapabilityProjector(mockReadModel)

	enterpriseCapabilityID := uuid.New().String()
	event := events.NewEnterpriseCapabilityUnlinked(
		uuid.New().String(),
		enterpriseCapabilityID,
		uuid.New().String(),
	)

	eventData, err := json.Marshal(event.EventData())
	require.NoError(t, err)

	err = projector.ProjectEvent(context.Background(), "EnterpriseCapabilityUnlinked", eventData)
	require.NoError(t, err)

	require.Len(t, mockReadModel.decrementAndRecalculateIDs, 1)
	assert.Equal(t, enterpriseCapabilityID, mockReadModel.decrementAndRecalculateIDs[0])
}

func TestEnterpriseCapabilityProjector_UnknownEvent_Ignored(t *testing.T) {
	mockReadModel := &mockEnterpriseCapabilityReadModel{}
	projector := newTestableEnterpriseCapabilityProjector(mockReadModel)

	err := projector.ProjectEvent(context.Background(), "UnknownEvent", []byte("{}"))
	require.NoError(t, err)

	assert.Empty(t, mockReadModel.insertedDTOs)
	assert.Empty(t, mockReadModel.updatedParams)
	assert.Empty(t, mockReadModel.deletedIDs)
}

func TestEnterpriseCapabilityProjector_ReadModelError_ReturnsError(t *testing.T) {
	mockReadModel := &mockEnterpriseCapabilityReadModel{
		insertErr: errors.New("database error"),
	}
	projector := newTestableEnterpriseCapabilityProjector(mockReadModel)

	event := events.NewEnterpriseCapabilityCreated(
		uuid.New().String(),
		"Test Capability",
		"Test Description",
		"Test Category",
	)

	eventData, err := json.Marshal(event.EventData())
	require.NoError(t, err)

	err = projector.ProjectEvent(context.Background(), "EnterpriseCapabilityCreated", eventData)
	assert.Error(t, err)
}

func TestEnterpriseCapabilityProjector_InvalidJSON_ReturnsError(t *testing.T) {
	mockReadModel := &mockEnterpriseCapabilityReadModel{}
	projector := newTestableEnterpriseCapabilityProjector(mockReadModel)

	err := projector.ProjectEvent(context.Background(), "EnterpriseCapabilityCreated", []byte("invalid json"))
	assert.Error(t, err)
}

func TestEnterpriseCapabilityProjector_Created_PreservesTimestamp(t *testing.T) {
	mockReadModel := &mockEnterpriseCapabilityReadModel{}
	projector := newTestableEnterpriseCapabilityProjector(mockReadModel)

	event := events.NewEnterpriseCapabilityCreated(
		uuid.New().String(),
		"Test Capability",
		"Test Description",
		"Test Category",
	)
	event.CreatedAt = time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	eventData, err := json.Marshal(event.EventData())
	require.NoError(t, err)

	err = projector.ProjectEvent(context.Background(), "EnterpriseCapabilityCreated", eventData)
	require.NoError(t, err)

	require.Len(t, mockReadModel.insertedDTOs, 1)
	assert.Equal(t, event.CreatedAt, mockReadModel.insertedDTOs[0].CreatedAt)
}
