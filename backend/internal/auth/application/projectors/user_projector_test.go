package projectors

import (
	"context"
	"encoding/json"
	"testing"

	"easi/backend/internal/auth/application/readmodels"
	"easi/backend/internal/auth/domain/events"
	authPL "easi/backend/internal/auth/publishedlanguage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockUserReadModelForProjector struct {
	insertedUsers map[string]readmodels.UserEventData
	roleUpdates   map[string]string
	statusUpdates map[string]string
	insertErr     error
	updateErr     error
}

func (m *mockUserReadModelForProjector) InsertFromEvent(ctx context.Context, data readmodels.UserEventData) error {
	if m.insertErr != nil {
		return m.insertErr
	}
	if m.insertedUsers == nil {
		m.insertedUsers = make(map[string]readmodels.UserEventData)
	}
	m.insertedUsers[data.ID] = data
	return nil
}

func (m *mockUserReadModelForProjector) recordUpdate(target *map[string]string, id, value string) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	if *target == nil {
		*target = make(map[string]string)
	}
	(*target)[id] = value
	return nil
}

func (m *mockUserReadModelForProjector) UpdateRole(ctx context.Context, id string, role string) error {
	return m.recordUpdate(&m.roleUpdates, id, role)
}

func (m *mockUserReadModelForProjector) UpdateStatus(ctx context.Context, id string, status string) error {
	return m.recordUpdate(&m.statusUpdates, id, status)
}

func TestUserProjector_HandlesUserCreated(t *testing.T) {
	mockReadModel := &mockUserReadModelForProjector{}
	projector := NewUserProjector(mockReadModel)

	event := events.NewUserCreated("user-123", "user@example.com", "Test User", "architect", "ext-123", "inv-123")
	eventData, _ := json.Marshal(event.EventData())

	err := projector.ProjectEvent(context.Background(), authPL.UserCreated, eventData)
	require.NoError(t, err)

	require.Contains(t, mockReadModel.insertedUsers, "user-123")
	inserted := mockReadModel.insertedUsers["user-123"]
	assert.Equal(t, "user@example.com", inserted.Email)
	assert.Equal(t, "architect", inserted.Role)
}

func TestUserProjector_HandlesUserRoleChanged(t *testing.T) {
	mockReadModel := &mockUserReadModelForProjector{}
	projector := NewUserProjector(mockReadModel)

	event := events.NewUserRoleChanged("user-123", "architect", "admin", "changer-456")
	eventData, _ := json.Marshal(event.EventData())

	err := projector.ProjectEvent(context.Background(), authPL.UserRoleChanged, eventData)
	require.NoError(t, err)

	assert.Equal(t, "admin", mockReadModel.roleUpdates["user-123"])
}

func TestUserProjector_HandlesUserDisabled(t *testing.T) {
	mockReadModel := &mockUserReadModelForProjector{}
	projector := NewUserProjector(mockReadModel)

	event := events.NewUserDisabled("user-456", "disabler-789")
	eventData, _ := json.Marshal(event.EventData())

	err := projector.ProjectEvent(context.Background(), authPL.UserDisabled, eventData)
	require.NoError(t, err)

	assert.Equal(t, "disabled", mockReadModel.statusUpdates["user-456"])
}

func TestUserProjector_HandlesUserEnabled(t *testing.T) {
	mockReadModel := &mockUserReadModelForProjector{}
	projector := NewUserProjector(mockReadModel)

	event := events.NewUserEnabled("user-789", "enabler-111")
	eventData, _ := json.Marshal(event.EventData())

	err := projector.ProjectEvent(context.Background(), authPL.UserEnabled, eventData)
	require.NoError(t, err)

	assert.Equal(t, "active", mockReadModel.statusUpdates["user-789"])
}

func TestUserProjector_IgnoresUnknownEventTypes(t *testing.T) {
	mockReadModel := &mockUserReadModelForProjector{}
	projector := NewUserProjector(mockReadModel)

	err := projector.ProjectEvent(context.Background(), "SomeUnknownEventType", []byte("{}"))
	require.NoError(t, err)

	assert.Empty(t, mockReadModel.insertedUsers)
	assert.Empty(t, mockReadModel.roleUpdates)
	assert.Empty(t, mockReadModel.statusUpdates)
}

func TestUserProjector_InvalidEventData_ReturnsError(t *testing.T) {
	mockReadModel := &mockUserReadModelForProjector{}
	projector := NewUserProjector(mockReadModel)

	err := projector.ProjectEvent(context.Background(), authPL.UserRoleChanged, []byte("invalid json"))
	assert.Error(t, err)
}

func TestUserProjector_ReadModelError_ReturnsError(t *testing.T) {
	mockReadModel := &mockUserReadModelForProjector{
		updateErr: assert.AnError,
	}
	projector := NewUserProjector(mockReadModel)

	event := events.NewUserRoleChanged("user-err", "admin", "architect", "changer-err")
	eventData, _ := json.Marshal(event.EventData())

	err := projector.ProjectEvent(context.Background(), authPL.UserRoleChanged, eventData)
	assert.Error(t, err)
}

func TestUserProjector_MultipleEvents_ProcessedInOrder(t *testing.T) {
	mockReadModel := &mockUserReadModelForProjector{}
	projector := NewUserProjector(mockReadModel)

	event1 := events.NewUserRoleChanged("user-multi", "stakeholder", "architect", "changer-1")
	eventData1, _ := json.Marshal(event1.EventData())

	event2 := events.NewUserDisabled("user-multi", "disabler-2")
	eventData2, _ := json.Marshal(event2.EventData())

	event3 := events.NewUserEnabled("user-multi", "enabler-3")
	eventData3, _ := json.Marshal(event3.EventData())

	require.NoError(t, projector.ProjectEvent(context.Background(), authPL.UserRoleChanged, eventData1))
	require.NoError(t, projector.ProjectEvent(context.Background(), authPL.UserDisabled, eventData2))
	require.NoError(t, projector.ProjectEvent(context.Background(), authPL.UserEnabled, eventData3))

	assert.Equal(t, "architect", mockReadModel.roleUpdates["user-multi"])
	assert.Equal(t, "active", mockReadModel.statusUpdates["user-multi"])
}
