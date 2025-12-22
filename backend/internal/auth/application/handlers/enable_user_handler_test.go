package handlers

import (
	"context"
	"errors"
	"testing"

	"easi/backend/internal/auth/application/commands"
	"easi/backend/internal/auth/domain/aggregates"
	"easi/backend/internal/shared/cqrs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockUserRepositoryForEnable struct {
	savedUsers []*aggregates.User
	userToLoad *aggregates.User
	getErr     error
	saveErr    error
}

func (m *mockUserRepositoryForEnable) Save(ctx context.Context, user *aggregates.User) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.savedUsers = append(m.savedUsers, user)
	return nil
}

func (m *mockUserRepositoryForEnable) GetByID(ctx context.Context, id string) (*aggregates.User, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.userToLoad, nil
}

type userRepositoryForEnable interface {
	Save(ctx context.Context, user *aggregates.User) error
	GetByID(ctx context.Context, id string) (*aggregates.User, error)
}

type testableEnableUserHandler struct {
	repository userRepositoryForEnable
}

func newTestableEnableUserHandler(repository userRepositoryForEnable) *testableEnableUserHandler {
	return &testableEnableUserHandler{
		repository: repository,
	}
}

func (h *testableEnableUserHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(*commands.EnableUser)
	if !ok {
		return cqrs.ErrInvalidCommand
	}

	user, err := h.repository.GetByID(ctx, command.UserID)
	if err != nil {
		return err
	}

	if err := user.Enable(command.EnabledByID); err != nil {
		return err
	}

	return h.repository.Save(ctx, user)
}

func TestEnableUserHandler_EnablesUserSuccessfully(t *testing.T) {
	user := createUserForTestEnable(t, "architect")
	user.Disable("disabler-111", false, false)
	user.MarkChangesAsCommitted()

	mockRepo := &mockUserRepositoryForEnable{userToLoad: user}

	handler := newTestableEnableUserHandler(mockRepo)

	cmd := &commands.EnableUser{
		UserID:      user.ID(),
		EnabledByID: "enabler-123",
	}

	err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	require.Len(t, mockRepo.savedUsers, 1)
	savedUser := mockRepo.savedUsers[0]
	assert.True(t, savedUser.Status().IsActive())
}

func TestEnableUserHandler_UserAlreadyActive_ReturnsError(t *testing.T) {
	user := createUserForTestEnable(t, "architect")
	mockRepo := &mockUserRepositoryForEnable{userToLoad: user}

	handler := newTestableEnableUserHandler(mockRepo)

	cmd := &commands.EnableUser{
		UserID:      user.ID(),
		EnabledByID: "enabler-456",
	}

	err := handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, aggregates.ErrUserAlreadyActive)
	assert.Empty(t, mockRepo.savedUsers)
}

func TestEnableUserHandler_UserNotFound_ReturnsError(t *testing.T) {
	mockRepo := &mockUserRepositoryForEnable{getErr: errors.New("not found")}

	handler := newTestableEnableUserHandler(mockRepo)

	cmd := &commands.EnableUser{
		UserID:      "nonexistent-id",
		EnabledByID: "enabler-789",
	}

	err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
}

func TestEnableUserHandler_InvalidCommand_ReturnsError(t *testing.T) {
	mockRepo := &mockUserRepositoryForEnable{}

	handler := newTestableEnableUserHandler(mockRepo)

	invalidCmd := &commands.DisableUser{}

	err := handler.Handle(context.Background(), invalidCmd)
	assert.ErrorIs(t, err, cqrs.ErrInvalidCommand)
}

func TestEnableUserHandler_SaveError_ReturnsError(t *testing.T) {
	user := createUserForTestEnable(t, "architect")
	user.Disable("disabler-222", false, false)
	user.MarkChangesAsCommitted()

	mockRepo := &mockUserRepositoryForEnable{
		userToLoad: user,
		saveErr:    errors.New("save failed"),
	}

	handler := newTestableEnableUserHandler(mockRepo)

	cmd := &commands.EnableUser{
		UserID:      user.ID(),
		EnabledByID: "enabler-err",
	}

	err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
}

func TestEnableUserHandler_EnableDisabledAdmin(t *testing.T) {
	user := createUserForTestEnable(t, "admin")
	user.Disable("disabler-333", false, false)
	user.MarkChangesAsCommitted()

	mockRepo := &mockUserRepositoryForEnable{userToLoad: user}

	handler := newTestableEnableUserHandler(mockRepo)

	cmd := &commands.EnableUser{
		UserID:      user.ID(),
		EnabledByID: "enabler-999",
	}

	err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	require.Len(t, mockRepo.savedUsers, 1)
	savedUser := mockRepo.savedUsers[0]
	assert.True(t, savedUser.Status().IsActive())
	assert.Equal(t, "admin", savedUser.Role().String())
}

func createUserForTestEnable(t *testing.T, roleName string) *aggregates.User {
	t.Helper()
	return createUserForTest(t, roleName)
}
