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
	ID                string
	Name              string
	Description       string
	DomainArchitectID string
}

func (m *mockBusinessDomainReadModel) Insert(ctx context.Context, dto readmodels.BusinessDomainDTO) error {
	if m.insertErr != nil {
		return m.insertErr
	}
	m.insertedDomains = append(m.insertedDomains, dto)
	return nil
}

func (m *mockBusinessDomainReadModel) Update(ctx context.Context, id string, update readmodels.BusinessDomainUpdate) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.updatedDomains = append(m.updatedDomains, updateBusinessDomainCall{
		ID:                id,
		Name:              update.Name,
		Description:       update.Description,
		DomainArchitectID: update.DomainArchitectID,
	})
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

func TestBusinessDomainProjector_HandleBusinessDomainCreated(t *testing.T) {
	mock := &mockBusinessDomainReadModel{}
	projector := NewBusinessDomainProjector(mock)

	event := events.NewBusinessDomainCreated(
		"bd-123",
		"Finance",
		"Financial operations and planning",
		"architect-1",
	)

	eventData, err := json.Marshal(event.EventData())
	require.NoError(t, err)

	err = projector.ProjectEvent(context.Background(), "BusinessDomainCreated", eventData)
	require.NoError(t, err)

	require.Len(t, mock.insertedDomains, 1)
	assert.Equal(t, "bd-123", mock.insertedDomains[0].ID)
	assert.Equal(t, "Finance", mock.insertedDomains[0].Name)
	assert.Equal(t, "Financial operations and planning", mock.insertedDomains[0].Description)
	assert.Equal(t, "architect-1", mock.insertedDomains[0].DomainArchitectID)
	assert.WithinDuration(t, time.Now().UTC(), mock.insertedDomains[0].CreatedAt, time.Second)
}

func TestBusinessDomainProjector_HandleBusinessDomainUpdated(t *testing.T) {
	mock := &mockBusinessDomainReadModel{}
	projector := NewBusinessDomainProjector(mock)

	event := events.NewBusinessDomainUpdated(
		"bd-123",
		"Finance & Accounting",
		"Updated description",
		"architect-2",
	)

	eventData, err := json.Marshal(event.EventData())
	require.NoError(t, err)

	err = projector.ProjectEvent(context.Background(), "BusinessDomainUpdated", eventData)
	require.NoError(t, err)

	require.Len(t, mock.updatedDomains, 1)
	assert.Equal(t, updateBusinessDomainCall{
		ID:                "bd-123",
		Name:              "Finance & Accounting",
		Description:       "Updated description",
		DomainArchitectID: "architect-2",
	}, mock.updatedDomains[0])
}

type bdProjectorEvent interface {
	EventType() string
	EventData() map[string]interface{}
}

func TestBusinessDomainProjector_SingleSlotEvents(t *testing.T) {
	tests := []struct {
		name           string
		event          bdProjectorEvent
		wantInSlot     []string
		slotFromMock   func(*mockBusinessDomainReadModel) []string
	}{
		{
			name:         "BusinessDomainDeleted records the deleted id",
			event:        events.NewBusinessDomainDeleted("bd-123"),
			wantInSlot:   []string{"bd-123"},
			slotFromMock: func(m *mockBusinessDomainReadModel) []string { return m.deletedIDs },
		},
		{
			name:         "CapabilityAssignedToDomain increments capability count for the domain",
			event:        events.NewCapabilityAssignedToDomain("assign-123", "bd-456", "cap-789"),
			wantInSlot:   []string{"bd-456"},
			slotFromMock: func(m *mockBusinessDomainReadModel) []string { return m.incrementCapCountCalls },
		},
		{
			name:         "CapabilityUnassignedFromDomain decrements capability count for the domain",
			event:        events.NewCapabilityUnassignedFromDomain("assign-123", "bd-456", "cap-789"),
			wantInSlot:   []string{"bd-456"},
			slotFromMock: func(m *mockBusinessDomainReadModel) []string { return m.decrementCapCountCalls },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockBusinessDomainReadModel{}
			projector := NewBusinessDomainProjector(mock)

			eventData, err := json.Marshal(tt.event.EventData())
			require.NoError(t, err)

			require.NoError(t, projector.ProjectEvent(context.Background(), tt.event.EventType(), eventData))
			assert.Equal(t, tt.wantInSlot, tt.slotFromMock(mock))
		})
	}
}

func TestBusinessDomainProjector_UnknownEventType_Ignored(t *testing.T) {
	mock := &mockBusinessDomainReadModel{}
	projector := NewBusinessDomainProjector(mock)

	err := projector.ProjectEvent(context.Background(), "SomethingElse", []byte("{}"))
	require.NoError(t, err)

	assert.Empty(t, mock.insertedDomains)
	assert.Empty(t, mock.updatedDomains)
	assert.Empty(t, mock.deletedIDs)
	assert.Empty(t, mock.incrementCapCountCalls)
	assert.Empty(t, mock.decrementCapCountCalls)
}
