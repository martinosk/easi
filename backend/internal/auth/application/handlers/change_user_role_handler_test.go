package handlers

import (
	"context"
	"errors"
	"testing"

	"easi/backend/internal/auth/application/commands"
	"easi/backend/internal/auth/domain/aggregates"
	"easi/backend/internal/auth/domain/valueobjects"
	"easi/backend/internal/shared/cqrs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockUserRepositoryForChangeRole struct {
	savedUsers []*aggregates.User
	userToLoad *aggregates.User
	getErr     error
	saveErr    error
}

func (m *mockUserRepositoryForChangeRole) Save(ctx context.Context, user *aggregates.User) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.savedUsers = append(m.savedUsers, user)
	return nil
}

func (m *mockUserRepositoryForChangeRole) GetByID(ctx context.Context, id string) (*aggregates.User, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.userToLoad, nil
}

type mockUserReadModelForChangeRole struct {
	isLastAdmin bool
	checkErr    error
}

func (m *mockUserReadModelForChangeRole) IsLastActiveAdmin(ctx context.Context, userID string) (bool, error) {
	if m.checkErr != nil {
		return false, m.checkErr
	}
	return m.isLastAdmin, nil
}

type userRepositoryForChangeRole interface {
	Save(ctx context.Context, user *aggregates.User) error
	GetByID(ctx context.Context, id string) (*aggregates.User, error)
}

type userReadModelForChangeRole interface {
	IsLastActiveAdmin(ctx context.Context, userID string) (bool, error)
}

type testableChangeUserRoleHandler struct {
	repository userRepositoryForChangeRole
	readModel  userReadModelForChangeRole
}

func newTestableChangeUserRoleHandler(
	repository userRepositoryForChangeRole,
	readModel userReadModelForChangeRole,
) *testableChangeUserRoleHandler {
	return &testableChangeUserRoleHandler{
		repository: repository,
		readModel:  readModel,
	}
}

func (h *testableChangeUserRoleHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(*commands.ChangeUserRole)
	if !ok {
		return cqrs.ErrInvalidCommand
	}

	newRole, err := valueobjects.RoleFromString(command.NewRole)
	if err != nil {
		return err
	}

	isLastAdmin, err := h.readModel.IsLastActiveAdmin(ctx, command.UserID)
	if err != nil {
		return err
	}

	user, err := h.repository.GetByID(ctx, command.UserID)
	if err != nil {
		return err
	}

	if err := user.ChangeRole(newRole, command.ChangedByID, isLastAdmin); err != nil {
		return err
	}

	return h.repository.Save(ctx, user)
}

func TestChangeUserRoleHandler_ChangesRoleSuccessfully(t *testing.T) {
	user := createUserForTest(t, "admin")
	mockRepo := &mockUserRepositoryForChangeRole{userToLoad: user}
	mockReadModel := &mockUserReadModelForChangeRole{isLastAdmin: false}

	handler := newTestableChangeUserRoleHandler(mockRepo, mockReadModel)

	cmd := &commands.ChangeUserRole{
		UserID:      user.ID(),
		NewRole:     "architect",
		ChangedByID: "changer-123",
	}

	err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	require.Len(t, mockRepo.savedUsers, 1)
	savedUser := mockRepo.savedUsers[0]
	assert.Equal(t, "architect", savedUser.Role().String())
}

func TestChangeUserRoleHandler_InvalidRole_ReturnsError(t *testing.T) {
	user := createUserForTest(t, "admin")
	mockRepo := &mockUserRepositoryForChangeRole{userToLoad: user}
	mockReadModel := &mockUserReadModelForChangeRole{isLastAdmin: false}

	handler := newTestableChangeUserRoleHandler(mockRepo, mockReadModel)

	cmd := &commands.ChangeUserRole{
		UserID:      user.ID(),
		NewRole:     "superadmin",
		ChangedByID: "changer-456",
	}

	err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
	assert.Empty(t, mockRepo.savedUsers)
}

func TestChangeUserRoleHandler_CannotDemoteLastAdmin(t *testing.T) {
	user := createUserForTest(t, "admin")
	mockRepo := &mockUserRepositoryForChangeRole{userToLoad: user}
	mockReadModel := &mockUserReadModelForChangeRole{isLastAdmin: true}

	handler := newTestableChangeUserRoleHandler(mockRepo, mockReadModel)

	cmd := &commands.ChangeUserRole{
		UserID:      user.ID(),
		NewRole:     "architect",
		ChangedByID: "changer-789",
	}

	err := handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, aggregates.ErrCannotDemoteLastAdmin)
	assert.Empty(t, mockRepo.savedUsers)
}

func TestChangeUserRoleHandler_UserNotFound_ReturnsError(t *testing.T) {
	mockRepo := &mockUserRepositoryForChangeRole{getErr: errors.New("not found")}
	mockReadModel := &mockUserReadModelForChangeRole{isLastAdmin: false}

	handler := newTestableChangeUserRoleHandler(mockRepo, mockReadModel)

	cmd := &commands.ChangeUserRole{
		UserID:      "nonexistent-id",
		NewRole:     "architect",
		ChangedByID: "changer-999",
	}

	err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
}

func TestChangeUserRoleHandler_InvalidCommand_ReturnsError(t *testing.T) {
	mockRepo := &mockUserRepositoryForChangeRole{}
	mockReadModel := &mockUserReadModelForChangeRole{}

	handler := newTestableChangeUserRoleHandler(mockRepo, mockReadModel)

	invalidCmd := &commands.DisableUser{}

	err := handler.Handle(context.Background(), invalidCmd)
	assert.ErrorIs(t, err, cqrs.ErrInvalidCommand)
}

func TestChangeUserRoleHandler_ReadModelError_ReturnsError(t *testing.T) {
	user := createUserForTest(t, "admin")
	mockRepo := &mockUserRepositoryForChangeRole{userToLoad: user}
	mockReadModel := &mockUserReadModelForChangeRole{checkErr: errors.New("database error")}

	handler := newTestableChangeUserRoleHandler(mockRepo, mockReadModel)

	cmd := &commands.ChangeUserRole{
		UserID:      user.ID(),
		NewRole:     "architect",
		ChangedByID: "changer-err",
	}

	err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
	assert.Empty(t, mockRepo.savedUsers)
}

func TestChangeUserRoleHandler_SameRole_ReturnsError(t *testing.T) {
	user := createUserForTest(t, "architect")
	mockRepo := &mockUserRepositoryForChangeRole{userToLoad: user}
	mockReadModel := &mockUserReadModelForChangeRole{isLastAdmin: false}

	handler := newTestableChangeUserRoleHandler(mockRepo, mockReadModel)

	cmd := &commands.ChangeUserRole{
		UserID:      user.ID(),
		NewRole:     "architect",
		ChangedByID: "changer-same",
	}

	err := handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, aggregates.ErrSameRole)
	assert.Empty(t, mockRepo.savedUsers)
}

func createUserForTest(t *testing.T, roleName string) *aggregates.User {
	t.Helper()

	email, err := valueobjects.NewEmail("test@example.com")
	require.NoError(t, err)

	role, err := valueobjects.RoleFromString(roleName)
	require.NoError(t, err)

	user, err := aggregates.NewUser(email, "Test User", role, "ext-test", "inv-test")
	require.NoError(t, err)

	user.MarkChangesAsCommitted()

	return user
}
