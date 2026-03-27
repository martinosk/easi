package handlers

import (
	"context"
	"errors"
	"testing"

	"easi/backend/internal/auth/application/commands"
	"easi/backend/internal/auth/domain/aggregates"
	"easi/backend/internal/auth/domain/valueobjects"
	"easi/backend/internal/shared/cqrs"

	"github.com/google/uuid"
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
	handler := NewChangeUserRoleHandler(mockRepo, &mockChangeUserRoleReadModel{})

	_, err := handler.Handle(context.Background(), &commands.ChangeUserRole{
		UserID:      user.ID(),
		NewRole:     "architect",
		ChangedByID: uuid.New().String(),
	})
	require.NoError(t, err)

	require.Len(t, mockRepo.savedUsers, 1)
	assert.Equal(t, "architect", mockRepo.savedUsers[0].Role().String())
}

func TestChangeUserRoleHandler_ReturnsError(t *testing.T) {
	tests := []struct {
		name        string
		role        string
		newRole     string
		isLastAdmin bool
		getErr      error
		readErr     error
		wantErr     error
	}{
		{
			name:    "invalid role",
			role:    "admin",
			newRole: "superadmin",
			wantErr: valueobjects.ErrInvalidRole,
		},
		{
			name:        "cannot demote last admin",
			role:        "admin",
			newRole:     "architect",
			isLastAdmin: true,
			wantErr:     aggregates.ErrCannotDemoteLastAdmin,
		},
		{
			name:    "same role",
			role:    "architect",
			newRole: "architect",
			wantErr: aggregates.ErrSameRole,
		},
		{
			name:   "user not found",
			role:   "admin",
			getErr: errors.New("not found"),
		},
		{
			name:    "read model error",
			role:    "admin",
			newRole: "architect",
			readErr: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := createUserForTest(t, tt.role)
			mockRepo := &mockChangeUserRoleRepository{userToLoad: user, getErr: tt.getErr}
			mockReadModel := &mockChangeUserRoleReadModel{isLastAdmin: tt.isLastAdmin, checkErr: tt.readErr}
			handler := NewChangeUserRoleHandler(mockRepo, mockReadModel)

			newRole := tt.newRole
			if newRole == "" {
				newRole = "architect"
			}

			_, err := handler.Handle(context.Background(), &commands.ChangeUserRole{
				UserID:      user.ID(),
				NewRole:     newRole,
				ChangedByID: uuid.New().String(),
			})

			assert.Error(t, err)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			}
			assert.Empty(t, mockRepo.savedUsers)
		})
	}
}

func TestChangeUserRoleHandler_InvalidCommand_ReturnsError(t *testing.T) {
	handler := NewChangeUserRoleHandler(&mockChangeUserRoleRepository{}, &mockChangeUserRoleReadModel{})

	_, err := handler.Handle(context.Background(), &commands.DisableUser{})
	assert.ErrorIs(t, err, cqrs.ErrInvalidCommand)
}

func createUserForTest(t *testing.T, roleName string) *aggregates.User {
	t.Helper()

	email, err := valueobjects.NewEmail("test@example.com")
	require.NoError(t, err)

	role, err := valueobjects.RoleFromString(roleName)
	require.NoError(t, err)

	profile := valueobjects.NewExternalProfile("Test User", "ext-test")
	user, err := aggregates.NewUser(email, profile, role, "inv-test")
	require.NoError(t, err)

	user.MarkChangesAsCommitted()

	return user
}
