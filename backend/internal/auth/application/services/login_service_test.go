package services

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

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

type userReadModelForLogin interface {
	GetByEmail(ctx context.Context, email string) (*readmodels.UserDTO, error)
	UpdateLastLogin(ctx context.Context, id uuid.UUID, lastLoginAt time.Time) error
}

type testableLoginService struct {
	userReadModel userReadModelForLogin
}

func newTestableLoginService(userReadModel userReadModelForLogin) *testableLoginService {
	return &testableLoginService{
		userReadModel: userReadModel,
	}
}

func (s *testableLoginService) ProcessLogin(ctx context.Context, email, name string) (*LoginResult, error) {
	existingUser, err := s.userReadModel.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if existingUser != nil {
		if existingUser.Status == "disabled" {
			return nil, ErrUserDisabled
		}
		if err := s.userReadModel.UpdateLastLogin(ctx, existingUser.ID, time.Now().UTC()); err != nil {
			return nil, err
		}
		return &LoginResult{
			UserID: existingUser.ID,
			Email:  existingUser.Email,
			Role:   existingUser.Role,
			IsNew:  false,
		}, nil
	}

	return nil, ErrNoValidInvitation
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
	service := newTestableLoginService(mockReadModel)

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
	service := newTestableLoginService(mockReadModel)

	result, err := service.ProcessLogin(context.Background(), "active@example.com", "Active User")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, userID, result.UserID)
	assert.Equal(t, "active@example.com", result.Email)
	assert.Equal(t, "architect", result.Role)
	assert.False(t, result.IsNew)
}
