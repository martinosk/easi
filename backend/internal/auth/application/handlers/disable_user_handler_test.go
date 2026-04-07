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

type mockDisableUserRepository struct {
	savedUsers []*aggregates.User
	userToLoad *aggregates.User
	getErr     error
	saveErr    error
}

func (m *mockDisableUserRepository) Save(ctx context.Context, user *aggregates.User) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.savedUsers = append(m.savedUsers, user)
	return nil
}

func (m *mockDisableUserRepository) GetByID(ctx context.Context, id string) (*aggregates.User, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.userToLoad, nil
}

type mockDisableUserReadModel struct {
	isLastAdmin bool
	checkErr    error
}

func (m *mockDisableUserReadModel) IsLastActiveAdmin(ctx context.Context, userID string) (bool, error) {
	if m.checkErr != nil {
		return false, m.checkErr
	}
	return m.isLastAdmin, nil
}

func TestDisableUserHandler_DisablesUserSuccessfully(t *testing.T) {
	tests := []struct {
		name string
		role string
	}{
		{name: "architect", role: "architect"},
		{name: "stakeholder", role: "stakeholder"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := createUserForTestDisable(t, tt.role)
			mockRepo := &mockDisableUserRepository{userToLoad: user}
			handler := NewDisableUserHandler(mockRepo, &mockDisableUserReadModel{})

			_, err := handler.Handle(context.Background(), &commands.DisableUser{
				UserID:       user.ID(),
				DisabledByID: uuid.New().String(),
			})
			require.NoError(t, err)

			require.Len(t, mockRepo.savedUsers, 1)
			assert.False(t, mockRepo.savedUsers[0].Status().IsActive())
		})
	}
}

func TestDisableUserHandler_ReturnsError(t *testing.T) {
	tests := []struct {
		name        string
		role        string
		selfDisable bool
		preDisable  bool
		isLastAdmin bool
		getErr      error
		readErr     error
		wantErr     error
	}{
		{
			name:        "cannot disable self",
			role:        "admin",
			selfDisable: true,
			wantErr:     aggregates.ErrCannotDisableSelf,
		},
		{
			name:        "cannot disable last admin",
			role:        "admin",
			isLastAdmin: true,
			wantErr:     aggregates.ErrCannotDisableLastAdmin,
		},
		{
			name:       "already disabled",
			role:       "architect",
			preDisable: true,
			wantErr:    aggregates.ErrUserAlreadyDisabled,
		},
		{
			name:   "user not found",
			role:   "admin",
			getErr: errors.New("not found"),
		},
		{
			name:    "read model error",
			role:    "admin",
			readErr: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := createUserForTestDisable(t, tt.role)
			if tt.preDisable {
				_ = user.Disable(valueobjects.NewUserID(), false, false)
			}
			user.MarkChangesAsCommitted()

			mockRepo := &mockDisableUserRepository{userToLoad: user, getErr: tt.getErr}
			mockReadModel := &mockDisableUserReadModel{isLastAdmin: tt.isLastAdmin, checkErr: tt.readErr}
			handler := NewDisableUserHandler(mockRepo, mockReadModel)

			disabledByID := uuid.New().String()
			if tt.selfDisable {
				disabledByID = user.ID()
			}

			_, err := handler.Handle(context.Background(), &commands.DisableUser{
				UserID:       user.ID(),
				DisabledByID: disabledByID,
			})

			assert.Error(t, err)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			}
			assert.Empty(t, mockRepo.savedUsers)
		})
	}
}

func TestDisableUserHandler_InvalidCommand_ReturnsError(t *testing.T) {
	handler := NewDisableUserHandler(&mockDisableUserRepository{}, &mockDisableUserReadModel{})

	_, err := handler.Handle(context.Background(), &commands.EnableUser{})
	assert.ErrorIs(t, err, cqrs.ErrInvalidCommand)
}

func createUserForTestDisable(t *testing.T, roleName string) *aggregates.User {
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
