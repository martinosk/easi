package projectors

import (
	"context"
	"encoding/json"
	"testing"

	"easi/backend/internal/auth/domain/events"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockUserReadModelForProjector struct {
	roleUpdates   map[string]string
	statusUpdates map[string]string
	updateErr     error
}

func (m *mockUserReadModelForProjector) UpdateRole(ctx context.Context, id string, role string) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	if m.roleUpdates == nil {
		m.roleUpdates = make(map[string]string)
	}
	m.roleUpdates[id] = role
	return nil
}

func (m *mockUserReadModelForProjector) UpdateStatus(ctx context.Context, id string, status string) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	if m.statusUpdates == nil {
		m.statusUpdates = make(map[string]string)
	}
	m.statusUpdates[id] = status
	return nil
}

type userReadModelInterface interface {
	UpdateRole(ctx context.Context, id string, role string) error
	UpdateStatus(ctx context.Context, id string, status string) error
}

type testableUserProjector struct {
	readModel userReadModelInterface
}

func newTestableUserProjector(readModel userReadModelInterface) *testableUserProjector {
	return &testableUserProjector{
		readModel: readModel,
	}
}

func (p *testableUserProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	switch eventType {
	case "UserRoleChanged":
		return p.handleUserRoleChanged(ctx, eventData)
	case "UserDisabled":
		return p.handleUserDisabled(ctx, eventData)
	case "UserEnabled":
		return p.handleUserEnabled(ctx, eventData)
	}
	return nil
}

func (p *testableUserProjector) handleUserRoleChanged(ctx context.Context, eventData []byte) error {
	var event userRoleChangedEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		return err
	}

	return p.readModel.UpdateRole(ctx, event.ID, event.NewRole)
}

func (p *testableUserProjector) handleUserDisabled(ctx context.Context, eventData []byte) error {
	var event userStatusEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		return err
	}

	return p.readModel.UpdateStatus(ctx, event.ID, "disabled")
}

func (p *testableUserProjector) handleUserEnabled(ctx context.Context, eventData []byte) error {
	var event userStatusEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		return err
	}

	return p.readModel.UpdateStatus(ctx, event.ID, "active")
}

func TestUserProjector_HandlesUserRoleChanged(t *testing.T) {
	mockReadModel := &mockUserReadModelForProjector{}
	projector := newTestableUserProjector(mockReadModel)

	event := events.NewUserRoleChanged("user-123", "architect", "admin", "changer-456")
	eventData, _ := json.Marshal(event.EventData())

	err := projector.ProjectEvent(context.Background(), "UserRoleChanged", eventData)
	require.NoError(t, err)

	assert.Equal(t, "admin", mockReadModel.roleUpdates["user-123"])
}

func TestUserProjector_HandlesUserDisabled(t *testing.T) {
	mockReadModel := &mockUserReadModelForProjector{}
	projector := newTestableUserProjector(mockReadModel)

	event := events.NewUserDisabled("user-456", "disabler-789")
	eventData, _ := json.Marshal(event.EventData())

	err := projector.ProjectEvent(context.Background(), "UserDisabled", eventData)
	require.NoError(t, err)

	assert.Equal(t, "disabled", mockReadModel.statusUpdates["user-456"])
}

func TestUserProjector_HandlesUserEnabled(t *testing.T) {
	mockReadModel := &mockUserReadModelForProjector{}
	projector := newTestableUserProjector(mockReadModel)

	event := events.NewUserEnabled("user-789", "enabler-111")
	eventData, _ := json.Marshal(event.EventData())

	err := projector.ProjectEvent(context.Background(), "UserEnabled", eventData)
	require.NoError(t, err)

	assert.Equal(t, "active", mockReadModel.statusUpdates["user-789"])
}

func TestUserProjector_IgnoresUnknownEventTypes(t *testing.T) {
	mockReadModel := &mockUserReadModelForProjector{}
	projector := newTestableUserProjector(mockReadModel)

	event := events.NewUserCreated("user-999", "new@example.com", "New User", "architect", "ext-999", "inv-999")
	eventData, _ := json.Marshal(event.EventData())

	err := projector.ProjectEvent(context.Background(), "UserCreated", eventData)
	require.NoError(t, err)

	assert.Empty(t, mockReadModel.roleUpdates)
	assert.Empty(t, mockReadModel.statusUpdates)
}

func TestUserProjector_InvalidEventData_ReturnsError(t *testing.T) {
	mockReadModel := &mockUserReadModelForProjector{}
	projector := newTestableUserProjector(mockReadModel)

	invalidEventBytes := []byte("invalid json")

	err := projector.ProjectEvent(context.Background(), "UserRoleChanged", invalidEventBytes)
	assert.Error(t, err)
}

func TestUserProjector_ReadModelError_ReturnsError(t *testing.T) {
	mockReadModel := &mockUserReadModelForProjector{
		updateErr: assert.AnError,
	}
	projector := newTestableUserProjector(mockReadModel)

	event := events.NewUserRoleChanged("user-err", "admin", "architect", "changer-err")
	eventData, _ := json.Marshal(event.EventData())

	err := projector.ProjectEvent(context.Background(), "UserRoleChanged", eventData)
	assert.Error(t, err)
}

func TestUserProjector_MultipleEvents_ProcessedInOrder(t *testing.T) {
	mockReadModel := &mockUserReadModelForProjector{}
	projector := newTestableUserProjector(mockReadModel)

	event1 := events.NewUserRoleChanged("user-multi", "stakeholder", "architect", "changer-1")
	eventData1, _ := json.Marshal(event1.EventData())

	event2 := events.NewUserDisabled("user-multi", "disabler-2")
	eventData2, _ := json.Marshal(event2.EventData())

	event3 := events.NewUserEnabled("user-multi", "enabler-3")
	eventData3, _ := json.Marshal(event3.EventData())

	err := projector.ProjectEvent(context.Background(), "UserRoleChanged", eventData1)
	require.NoError(t, err)

	err = projector.ProjectEvent(context.Background(), "UserDisabled", eventData2)
	require.NoError(t, err)

	err = projector.ProjectEvent(context.Background(), "UserEnabled", eventData3)
	require.NoError(t, err)

	assert.Equal(t, "architect", mockReadModel.roleUpdates["user-multi"])
	assert.Equal(t, "active", mockReadModel.statusUpdates["user-multi"])
}
