package projectors

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/domain/events"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockBusinessDomainReadModel struct {
	insertedDomains        []readmodels.BusinessDomainDTO
	updatedDomains         []updateBusinessDomainCall
	deletedIDs             []string
	incrementCapCountCalls []string
	decrementCapCountCalls []string
	insertErr              error
	updateErr              error
	deleteErr              error
	incrementErr           error
	decrementErr           error
}

type updateBusinessDomainCall struct {
	ID          string
	Name        string
	Description string
}

func (m *mockBusinessDomainReadModel) Insert(ctx context.Context, dto readmodels.BusinessDomainDTO) error {
	if m.insertErr != nil {
		return m.insertErr
	}
	m.insertedDomains = append(m.insertedDomains, dto)
	return nil
}

func (m *mockBusinessDomainReadModel) Update(ctx context.Context, id, name, description string) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.updatedDomains = append(m.updatedDomains, updateBusinessDomainCall{ID: id, Name: name, Description: description})
	return nil
}

func (m *mockBusinessDomainReadModel) Delete(ctx context.Context, id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	m.deletedIDs = append(m.deletedIDs, id)
	return nil
}

func (m *mockBusinessDomainReadModel) IncrementCapabilityCount(ctx context.Context, id string) error {
	if m.incrementErr != nil {
		return m.incrementErr
	}
	m.incrementCapCountCalls = append(m.incrementCapCountCalls, id)
	return nil
}

func (m *mockBusinessDomainReadModel) DecrementCapabilityCount(ctx context.Context, id string) error {
	if m.decrementErr != nil {
		return m.decrementErr
	}
	m.decrementCapCountCalls = append(m.decrementCapCountCalls, id)
	return nil
}

type businessDomainReadModelInterface interface {
	Insert(ctx context.Context, dto readmodels.BusinessDomainDTO) error
	Update(ctx context.Context, id, name, description string) error
	Delete(ctx context.Context, id string) error
	IncrementCapabilityCount(ctx context.Context, id string) error
	DecrementCapabilityCount(ctx context.Context, id string) error
}

type testableBusinessDomainProjector struct {
	readModel businessDomainReadModelInterface
}

func newTestableBusinessDomainProjector(readModel businessDomainReadModelInterface) *testableBusinessDomainProjector {
	return &testableBusinessDomainProjector{
		readModel: readModel,
	}
}

func (p *testableBusinessDomainProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	switch eventType {
	case "BusinessDomainCreated":
		return p.handleBusinessDomainCreated(ctx, eventData)
	case "BusinessDomainUpdated":
		return p.handleBusinessDomainUpdated(ctx, eventData)
	case "BusinessDomainDeleted":
		return p.handleBusinessDomainDeleted(ctx, eventData)
	case "CapabilityAssignedToDomain":
		return p.handleCapabilityAssignedToDomain(ctx, eventData)
	case "CapabilityUnassignedFromDomain":
		return p.handleCapabilityUnassignedFromDomain(ctx, eventData)
	}
	return nil
}

func (p *testableBusinessDomainProjector) handleBusinessDomainCreated(ctx context.Context, eventData []byte) error {
	var event events.BusinessDomainCreated
	if err := json.Unmarshal(eventData, &event); err != nil {
		return err
	}

	dto := readmodels.BusinessDomainDTO{
		ID:          event.ID,
		Name:        event.Name,
		Description: event.Description,
		CreatedAt:   event.CreatedAt,
	}
	return p.readModel.Insert(ctx, dto)
}

func (p *testableBusinessDomainProjector) handleBusinessDomainUpdated(ctx context.Context, eventData []byte) error {
	var event events.BusinessDomainUpdated
	if err := json.Unmarshal(eventData, &event); err != nil {
		return err
	}
	return p.readModel.Update(ctx, event.ID, event.Name, event.Description)
}

func (p *testableBusinessDomainProjector) handleBusinessDomainDeleted(ctx context.Context, eventData []byte) error {
	var event events.BusinessDomainDeleted
	if err := json.Unmarshal(eventData, &event); err != nil {
		return err
	}
	return p.readModel.Delete(ctx, event.ID)
}

func (p *testableBusinessDomainProjector) handleCapabilityAssignedToDomain(ctx context.Context, eventData []byte) error {
	var event events.CapabilityAssignedToDomain
	if err := json.Unmarshal(eventData, &event); err != nil {
		return err
	}
	return p.readModel.IncrementCapabilityCount(ctx, event.BusinessDomainID)
}

func (p *testableBusinessDomainProjector) handleCapabilityUnassignedFromDomain(ctx context.Context, eventData []byte) error {
	var event events.CapabilityUnassignedFromDomain
	if err := json.Unmarshal(eventData, &event); err != nil {
		return err
	}
	return p.readModel.DecrementCapabilityCount(ctx, event.BusinessDomainID)
}

func TestBusinessDomainProjector_HandleBusinessDomainCreated(t *testing.T) {
	mock := &mockBusinessDomainReadModel{}
	projector := newTestableBusinessDomainProjector(mock)

	event := events.NewBusinessDomainCreated(
		"bd-123",
		"Finance",
		"Financial operations and planning",
	)

	eventData, err := json.Marshal(event)
	require.NoError(t, err)

	ctx := context.Background()
	err = projector.ProjectEvent(ctx, "BusinessDomainCreated", eventData)
	require.NoError(t, err)

	assert.Len(t, mock.insertedDomains, 1)
	assert.Equal(t, "bd-123", mock.insertedDomains[0].ID)
	assert.Equal(t, "Finance", mock.insertedDomains[0].Name)
	assert.Equal(t, "Financial operations and planning", mock.insertedDomains[0].Description)
	assert.WithinDuration(t, time.Now().UTC(), mock.insertedDomains[0].CreatedAt, time.Second)
}

func TestBusinessDomainProjector_HandleBusinessDomainUpdated(t *testing.T) {
	mock := &mockBusinessDomainReadModel{}
	projector := newTestableBusinessDomainProjector(mock)

	event := events.NewBusinessDomainUpdated(
		"bd-123",
		"Finance & Accounting",
		"Updated description",
	)

	eventData, err := json.Marshal(event)
	require.NoError(t, err)

	ctx := context.Background()
	err = projector.ProjectEvent(ctx, "BusinessDomainUpdated", eventData)
	require.NoError(t, err)

	assert.Len(t, mock.updatedDomains, 1)
	assert.Equal(t, "bd-123", mock.updatedDomains[0].ID)
	assert.Equal(t, "Finance & Accounting", mock.updatedDomains[0].Name)
	assert.Equal(t, "Updated description", mock.updatedDomains[0].Description)
}

func TestBusinessDomainProjector_HandleBusinessDomainDeleted(t *testing.T) {
	mock := &mockBusinessDomainReadModel{}
	projector := newTestableBusinessDomainProjector(mock)

	event := events.NewBusinessDomainDeleted("bd-123")

	eventData, err := json.Marshal(event)
	require.NoError(t, err)

	ctx := context.Background()
	err = projector.ProjectEvent(ctx, "BusinessDomainDeleted", eventData)
	require.NoError(t, err)

	assert.Len(t, mock.deletedIDs, 1)
	assert.Equal(t, "bd-123", mock.deletedIDs[0])
}

func TestBusinessDomainProjector_HandleCapabilityAssignedToDomain(t *testing.T) {
	mock := &mockBusinessDomainReadModel{}
	projector := newTestableBusinessDomainProjector(mock)

	event := events.NewCapabilityAssignedToDomain(
		"assign-123",
		"bd-456",
		"cap-789",
	)

	eventData, err := json.Marshal(event)
	require.NoError(t, err)

	ctx := context.Background()
	err = projector.ProjectEvent(ctx, "CapabilityAssignedToDomain", eventData)
	require.NoError(t, err)

	assert.Len(t, mock.incrementCapCountCalls, 1)
	assert.Equal(t, "bd-456", mock.incrementCapCountCalls[0])
}

func TestBusinessDomainProjector_HandleCapabilityUnassignedFromDomain(t *testing.T) {
	mock := &mockBusinessDomainReadModel{}
	projector := newTestableBusinessDomainProjector(mock)

	event := events.NewCapabilityUnassignedFromDomain(
		"assign-123",
		"bd-456",
		"cap-789",
	)

	eventData, err := json.Marshal(event)
	require.NoError(t, err)

	ctx := context.Background()
	err = projector.ProjectEvent(ctx, "CapabilityUnassignedFromDomain", eventData)
	require.NoError(t, err)

	assert.Len(t, mock.decrementCapCountCalls, 1)
	assert.Equal(t, "bd-456", mock.decrementCapCountCalls[0])
}
