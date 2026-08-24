package services

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"easi/backend/internal/auth/application/readmodels"
	"easi/backend/internal/auth/domain/aggregates"
	"easi/backend/internal/shared/cqrs"
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

type exactMatchUserReadModel struct {
	storedEmail string
	user        *readmodels.UserDTO
}

func (m *exactMatchUserReadModel) GetByEmail(ctx context.Context, email string) (*readmodels.UserDTO, error) {
	if email == m.storedEmail {
		return m.user, nil
	}
	return nil, nil
}

func (m *exactMatchUserReadModel) UpdateLastLogin(ctx context.Context, id uuid.UUID, lastLoginAt time.Time) error {
	return nil
}

type exactMatchInvitationReadModel struct {
	invitation *readmodels.InvitationDTO
}

func (m *exactMatchInvitationReadModel) GetAnyPendingByEmail(ctx context.Context, email string) (*readmodels.InvitationDTO, error) {
	if m.invitation != nil && email == m.invitation.Email {
		return m.invitation, nil
	}
	return nil, nil
}

type recordingCommandBus struct {
	dispatched []cqrs.Command
}

func (b *recordingCommandBus) Dispatch(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	b.dispatched = append(b.dispatched, cmd)
	return cqrs.EmptyResult(), nil
}

func (b *recordingCommandBus) Register(commandType string, handler cqrs.CommandHandler) {}

type capturingUserRepo struct {
	saved *aggregates.User
}

func (r *capturingUserRepo) Save(ctx context.Context, user *aggregates.User) error {
	r.saved = user
	return nil
}

func TestLoginService_MixedCaseEmailClaim_MatchesLowercaseInvitation(t *testing.T) {
	invitation := &readmodels.InvitationDTO{
		ID:        uuid.NewString(),
		Email:     "udicr@dfds.com",
		Role:      "architect",
		Status:    "pending",
		ExpiresAt: time.Now().UTC().Add(24 * time.Hour),
	}
	userRepo := &capturingUserRepo{}
	service := NewLoginService(
		&exactMatchUserReadModel{},
		&exactMatchInvitationReadModel{invitation: invitation},
		&recordingCommandBus{},
		userRepo,
	)

	result, err := service.ProcessLogin(context.Background(), "UDICR@dfds.com", "Udi Cr")

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "udicr@dfds.com", result.Email)
	assert.True(t, result.IsNew)
	require.NotNil(t, userRepo.saved)
}

func TestLoginService_MixedCaseEmailClaim_MatchesExistingLowercaseUser(t *testing.T) {
	userID := uuid.New()
	mockReadModel := &exactMatchUserReadModel{
		storedEmail: "active@example.com",
		user: &readmodels.UserDTO{
			ID:     userID,
			Email:  "active@example.com",
			Role:   "architect",
			Status: "active",
		},
	}
	service := newLoginServiceForExistingUserTest(mockReadModel)

	result, err := service.ProcessLogin(context.Background(), "ACTIVE@Example.COM", "Active User")

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, userID, result.UserID)
	assert.Equal(t, "active@example.com", result.Email)
	assert.False(t, result.IsNew)
}
