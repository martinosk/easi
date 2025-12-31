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

type mockChangeUserRoleRepository struct {
	savedUsers []*aggregates.User
	userToLoad *aggregates.User
	getErr     error
	saveErr    error
}

func (m *mockChangeUserRoleRepository) Save(ctx context.Context, user *aggregates.User) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.savedUsers = append(m.savedUsers, user)
	return nil
}

func (m *mockChangeUserRoleRepository) GetByID(ctx context.Context, id string) (*aggregates.User, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.userToLoad, nil
}

type mockChangeUserRoleReadModel struct {
	isLastAdmin bool
	checkErr    error
}

func (m *mockChangeUserRoleReadModel) IsLastActiveAdmin(ctx context.Context, userID string) (bool, error) {
	if m.checkErr != nil {
		return false, m.checkErr
	}
	return m.isLastAdmin, nil
}

func TestChangeUserRoleHandler_ChangesRoleSuccessfully(t *testing.T) {
	user := createUserForTest(t, "admin")
	mockRepo := &mockChangeUserRoleRepository{userToLoad: user}
	mockReadModel := &mockChangeUserRoleReadModel{isLastAdmin: false}

	handler := NewChangeUserRoleHandler(mockRepo, mockReadModel)

	cmd := &commands.ChangeUserRole{
		UserID:      user.ID(),
		NewRole:     "architect",
		ChangedByID: "changer-123",
	}

	_, err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	require.Len(t, mockRepo.savedUsers, 1)
	savedUser := mockRepo.savedUsers[0]
	assert.Equal(t, "architect", savedUser.Role().String())
}

func TestChangeUserRoleHandler_InvalidRole_ReturnsError(t *testing.T) {
	user := createUserForTest(t, "admin")
	mockRepo := &mockChangeUserRoleRepository{userToLoad: user}
	mockReadModel := &mockChangeUserRoleReadModel{isLastAdmin: false}

	handler := NewChangeUserRoleHandler(mockRepo, mockReadModel)

	cmd := &commands.ChangeUserRole{
		UserID:      user.ID(),
		NewRole:     "superadmin",
		ChangedByID: "changer-456",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
	assert.Empty(t, mockRepo.savedUsers)
}

func TestChangeUserRoleHandler_CannotDemoteLastAdmin(t *testing.T) {
	user := createUserForTest(t, "admin")
	mockRepo := &mockChangeUserRoleRepository{userToLoad: user}
	mockReadModel := &mockChangeUserRoleReadModel{isLastAdmin: true}

	handler := NewChangeUserRoleHandler(mockRepo, mockReadModel)

	cmd := &commands.ChangeUserRole{
		UserID:      user.ID(),
		NewRole:     "architect",
		ChangedByID: "changer-789",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, aggregates.ErrCannotDemoteLastAdmin)
	assert.Empty(t, mockRepo.savedUsers)
}

func TestChangeUserRoleHandler_UserNotFound_ReturnsError(t *testing.T) {
	mockRepo := &mockChangeUserRoleRepository{getErr: errors.New("not found")}
	mockReadModel := &mockChangeUserRoleReadModel{isLastAdmin: false}

	handler := NewChangeUserRoleHandler(mockRepo, mockReadModel)

	cmd := &commands.ChangeUserRole{
		UserID:      "nonexistent-id",
		NewRole:     "architect",
		ChangedByID: "changer-999",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
}

func TestChangeUserRoleHandler_InvalidCommand_ReturnsError(t *testing.T) {
	mockRepo := &mockChangeUserRoleRepository{}
	mockReadModel := &mockChangeUserRoleReadModel{}

	handler := NewChangeUserRoleHandler(mockRepo, mockReadModel)

	invalidCmd := &commands.DisableUser{}

	_, err := handler.Handle(context.Background(), invalidCmd)
	assert.ErrorIs(t, err, cqrs.ErrInvalidCommand)
}

func TestChangeUserRoleHandler_ReadModelError_ReturnsError(t *testing.T) {
	user := createUserForTest(t, "admin")
	mockRepo := &mockChangeUserRoleRepository{userToLoad: user}
	mockReadModel := &mockChangeUserRoleReadModel{checkErr: errors.New("database error")}

	handler := NewChangeUserRoleHandler(mockRepo, mockReadModel)

	cmd := &commands.ChangeUserRole{
		UserID:      user.ID(),
		NewRole:     "architect",
		ChangedByID: "changer-err",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
	assert.Empty(t, mockRepo.savedUsers)
}

func TestChangeUserRoleHandler_SameRole_ReturnsError(t *testing.T) {
	user := createUserForTest(t, "architect")
	mockRepo := &mockChangeUserRoleRepository{userToLoad: user}
	mockReadModel := &mockChangeUserRoleReadModel{isLastAdmin: false}

	handler := NewChangeUserRoleHandler(mockRepo, mockReadModel)

	cmd := &commands.ChangeUserRole{
		UserID:      user.ID(),
		NewRole:     "architect",
		ChangedByID: "changer-same",
	}

	_, err := handler.Handle(context.Background(), cmd)
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
