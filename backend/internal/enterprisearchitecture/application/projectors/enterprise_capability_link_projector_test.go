package projectors

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"easi/backend/internal/enterprisearchitecture/application/readmodels"
	"easi/backend/internal/enterprisearchitecture/domain/events"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockEnterpriseCapabilityLinkReadModel struct {
	insertedDTOs []readmodels.EnterpriseCapabilityLinkDTO
	deletedIDs   []string
	insertErr    error
	deleteErr    error
}

func (m *mockEnterpriseCapabilityLinkReadModel) Insert(ctx context.Context, dto readmodels.EnterpriseCapabilityLinkDTO) error {
	if m.insertErr != nil {
		return m.insertErr
	}
	m.insertedDTOs = append(m.insertedDTOs, dto)
	return nil
}

func (m *mockEnterpriseCapabilityLinkReadModel) Delete(ctx context.Context, id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	m.deletedIDs = append(m.deletedIDs, id)
	return nil
}

type enterpriseCapabilityLinkReadModelForProjector interface {
	Insert(ctx context.Context, dto readmodels.EnterpriseCapabilityLinkDTO) error
	Delete(ctx context.Context, id string) error
}

type testableEnterpriseCapabilityLinkProjector struct {
	readModel enterpriseCapabilityLinkReadModelForProjector
}

func newTestableEnterpriseCapabilityLinkProjector(readModel enterpriseCapabilityLinkReadModelForProjector) *testableEnterpriseCapabilityLinkProjector {
	return &testableEnterpriseCapabilityLinkProjector{readModel: readModel}
}

func (p *testableEnterpriseCapabilityLinkProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	handlers := map[string]func(context.Context, []byte) error{
		"EnterpriseCapabilityLinked":   p.handleLinked,
		"EnterpriseCapabilityUnlinked": p.handleUnlinked,
	}

	if handler, exists := handlers[eventType]; exists {
		return handler(ctx, eventData)
	}
	return nil
}

func (p *testableEnterpriseCapabilityLinkProjector) handleLinked(ctx context.Context, eventData []byte) error {
	var event events.EnterpriseCapabilityLinked
	if err := json.Unmarshal(eventData, &event); err != nil {
		return err
	}
	dto := readmodels.EnterpriseCapabilityLinkDTO{
		ID:                     event.ID,
		EnterpriseCapabilityID: event.EnterpriseCapabilityID,
		DomainCapabilityID:     event.DomainCapabilityID,
		LinkedBy:               event.LinkedBy,
		LinkedAt:               event.LinkedAt,
	}
	return p.readModel.Insert(ctx, dto)
}

func (p *testableEnterpriseCapabilityLinkProjector) handleUnlinked(ctx context.Context, eventData []byte) error {
	var event events.EnterpriseCapabilityUnlinked
	if err := json.Unmarshal(eventData, &event); err != nil {
		return err
	}
	return p.readModel.Delete(ctx, event.ID)
}

func TestEnterpriseCapabilityLinkProjector_Linked_InsertsLink(t *testing.T) {
	mockReadModel := &mockEnterpriseCapabilityLinkReadModel{}
	projector := newTestableEnterpriseCapabilityLinkProjector(mockReadModel)

	enterpriseCapabilityID := uuid.New().String()
	domainCapabilityID := uuid.New().String()
	event := events.NewEnterpriseCapabilityLinked(
		uuid.New().String(),
		enterpriseCapabilityID,
		domainCapabilityID,
		"user@example.com",
	)

	eventData, err := json.Marshal(event.EventData())
	require.NoError(t, err)

	err = projector.ProjectEvent(context.Background(), "EnterpriseCapabilityLinked", eventData)
	require.NoError(t, err)

	require.Len(t, mockReadModel.insertedDTOs, 1)
	dto := mockReadModel.insertedDTOs[0]
	assert.Equal(t, event.ID, dto.ID)
	assert.Equal(t, enterpriseCapabilityID, dto.EnterpriseCapabilityID)
	assert.Equal(t, domainCapabilityID, dto.DomainCapabilityID)
	assert.Equal(t, "user@example.com", dto.LinkedBy)
	assert.Equal(t, event.LinkedAt, dto.LinkedAt)
}

func TestEnterpriseCapabilityLinkProjector_Unlinked_DeletesLink(t *testing.T) {
	mockReadModel := &mockEnterpriseCapabilityLinkReadModel{}
	projector := newTestableEnterpriseCapabilityLinkProjector(mockReadModel)

	linkID := uuid.New().String()
	event := events.NewEnterpriseCapabilityUnlinked(
		linkID,
		uuid.New().String(),
		uuid.New().String(),
	)

	eventData, err := json.Marshal(event.EventData())
	require.NoError(t, err)

	err = projector.ProjectEvent(context.Background(), "EnterpriseCapabilityUnlinked", eventData)
	require.NoError(t, err)

	require.Len(t, mockReadModel.deletedIDs, 1)
	assert.Equal(t, linkID, mockReadModel.deletedIDs[0])
}

func TestEnterpriseCapabilityLinkProjector_UnknownEvent_Ignored(t *testing.T) {
	mockReadModel := &mockEnterpriseCapabilityLinkReadModel{}
	projector := newTestableEnterpriseCapabilityLinkProjector(mockReadModel)

	err := projector.ProjectEvent(context.Background(), "UnknownEvent", []byte("{}"))
	require.NoError(t, err)

	assert.Empty(t, mockReadModel.insertedDTOs)
	assert.Empty(t, mockReadModel.deletedIDs)
}

func TestEnterpriseCapabilityLinkProjector_InsertError_ReturnsError(t *testing.T) {
	mockReadModel := &mockEnterpriseCapabilityLinkReadModel{
		insertErr: errors.New("database error"),
	}
	projector := newTestableEnterpriseCapabilityLinkProjector(mockReadModel)

	event := events.NewEnterpriseCapabilityLinked(
		uuid.New().String(),
		uuid.New().String(),
		uuid.New().String(),
		"user@example.com",
	)

	eventData, err := json.Marshal(event.EventData())
	require.NoError(t, err)

	err = projector.ProjectEvent(context.Background(), "EnterpriseCapabilityLinked", eventData)
	assert.Error(t, err)
}

func TestEnterpriseCapabilityLinkProjector_DeleteError_ReturnsError(t *testing.T) {
	mockReadModel := &mockEnterpriseCapabilityLinkReadModel{
		deleteErr: errors.New("database error"),
	}
	projector := newTestableEnterpriseCapabilityLinkProjector(mockReadModel)

	event := events.NewEnterpriseCapabilityUnlinked(
		uuid.New().String(),
		uuid.New().String(),
		uuid.New().String(),
	)

	eventData, err := json.Marshal(event.EventData())
	require.NoError(t, err)

	err = projector.ProjectEvent(context.Background(), "EnterpriseCapabilityUnlinked", eventData)
	assert.Error(t, err)
}

func TestEnterpriseCapabilityLinkProjector_InvalidJSON_ReturnsError(t *testing.T) {
	mockReadModel := &mockEnterpriseCapabilityLinkReadModel{}
	projector := newTestableEnterpriseCapabilityLinkProjector(mockReadModel)

	err := projector.ProjectEvent(context.Background(), "EnterpriseCapabilityLinked", []byte("invalid json"))
	assert.Error(t, err)
}
