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

type mockEnableUserRepository struct {
	savedUsers []*aggregates.User
	userToLoad *aggregates.User
	getErr     error
	saveErr    error
}

func (m *mockEnableUserRepository) Save(ctx context.Context, user *aggregates.User) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.savedUsers = append(m.savedUsers, user)
	return nil
}

func (m *mockEnableUserRepository) GetByID(ctx context.Context, id string) (*aggregates.User, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.userToLoad, nil
}

func TestEnableUserHandler_EnablesSuccessfully(t *testing.T) {
	tests := []struct {
		name string
		role string
	}{
		{name: "architect", role: "architect"},
		{name: "admin", role: "admin"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := createUserForTestEnable(t, tt.role)
			user.Disable(valueobjects.NewUserID(), false, false)
			user.MarkChangesAsCommitted()

			mockRepo := &mockEnableUserRepository{userToLoad: user}
			handler := NewEnableUserHandler(mockRepo)

			_, err := handler.Handle(context.Background(), &commands.EnableUser{
				UserID:      user.ID(),
				EnabledByID: uuid.New().String(),
			})
			require.NoError(t, err)

			require.Len(t, mockRepo.savedUsers, 1)
			assert.True(t, mockRepo.savedUsers[0].Status().IsActive())
		})
	}
}

func TestEnableUserHandler_ReturnsError(t *testing.T) {
	tests := []struct {
		name    string
		role    string
		disable bool
		getErr  error
		saveErr error
		wantErr error
	}{
		{
			name:    "already active",
			role:    "architect",
			wantErr: aggregates.ErrUserAlreadyActive,
		},
		{
			name:   "user not found",
			role:   "architect",
			getErr: errors.New("not found"),
		},
		{
			name:    "save error",
			role:    "architect",
			disable: true,
			saveErr: errors.New("save failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := createUserForTestEnable(t, tt.role)
			if tt.disable {
				user.Disable(valueobjects.NewUserID(), false, false)
			}
			user.MarkChangesAsCommitted()

			mockRepo := &mockEnableUserRepository{
				userToLoad: user,
				getErr:     tt.getErr,
				saveErr:    tt.saveErr,
			}
			handler := NewEnableUserHandler(mockRepo)

			_, err := handler.Handle(context.Background(), &commands.EnableUser{
				UserID:      user.ID(),
				EnabledByID: uuid.New().String(),
			})

			assert.Error(t, err)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			}
		})
	}
}

func TestEnableUserHandler_InvalidCommand_ReturnsError(t *testing.T) {
	handler := NewEnableUserHandler(&mockEnableUserRepository{})

	_, err := handler.Handle(context.Background(), &commands.DisableUser{})
	assert.ErrorIs(t, err, cqrs.ErrInvalidCommand)
}

func createUserForTestEnable(t *testing.T, roleName string) *aggregates.User {
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
