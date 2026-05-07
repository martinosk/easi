package services

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"easi/backend/internal/auth/application/readmodels"
)

type mockUserReadModel struct {
	userByEmail *readmodels.UserDTO
	getErr      error
}

func (m *mockUserReadModel) GetByEmail(ctx context.Context, email string) (*readmodels.UserDTO, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.userByEmail, nil
}

func (m *mockUserReadModel) UpdateLastLogin(ctx context.Context, id uuid.UUID, lastLoginAt time.Time) error {
	return nil
}

type mockInvitationReadModel struct {
	invitation *readmodels.InvitationDTO
}

func (m *mockInvitationReadModel) GetAnyPendingByEmail(ctx context.Context, email string) (*readmodels.InvitationDTO, error) {
	return m.invitation, nil
}

func newLoginServiceForExistingUserTest(userReadModel LoginUserReadModel) *LoginService {
	return NewLoginService(userReadModel, &mockInvitationReadModel{}, nil, nil)
}

func TestLoginService_DisabledUser_ReturnsError(t *testing.T) {
	userID := uuid.New()
	disabledUser := &readmodels.UserDTO{
		ID:     userID,
		Email:  "disabled@example.com",
		Role:   "architect",
		Status: "disabled",
	}

	mockReadModel := &mockUserReadModel{userByEmail: disabledUser}
	service := newLoginServiceForExistingUserTest(mockReadModel)

	result, err := service.ProcessLogin(context.Background(), "disabled@example.com", "Disabled User")

	assert.ErrorIs(t, err, ErrUserDisabled)
	assert.Nil(t, result)
}

func TestLoginService_ActiveUser_Succeeds(t *testing.T) {
	userID := uuid.New()
	activeUser := &readmodels.UserDTO{
		ID:     userID,
		Email:  "active@example.com",
		Role:   "architect",
		Status: "active",
	}

	mockReadModel := &mockUserReadModel{userByEmail: activeUser}
	service := newLoginServiceForExistingUserTest(mockReadModel)

	result, err := service.ProcessLogin(context.Background(), "active@example.com", "Active User")

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, userID, result.UserID)
	assert.Equal(t, "active@example.com", result.Email)
	assert.Equal(t, "architect", result.Role)
	assert.False(t, result.IsNew)
}

func TestLoginService_NoExistingUserAndNoInvitation_ReturnsError(t *testing.T) {
	mockReadModel := &mockUserReadModel{userByEmail: nil}
	service := newLoginServiceForExistingUserTest(mockReadModel)

	result, err := service.ProcessLogin(context.Background(), "stranger@example.com", "Stranger")

	assert.ErrorIs(t, err, ErrNoValidInvitation)
	assert.Nil(t, result)
}
