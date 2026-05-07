package services

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"easi/backend/internal/auth/application/commands"
	"easi/backend/internal/auth/application/readmodels"
	"easi/backend/internal/auth/domain/aggregates"
	"easi/backend/internal/auth/domain/valueobjects"
	"easi/backend/internal/shared/cqrs"
)

var ErrNoValidInvitation = errors.New("no valid invitation found for this email")
var ErrUserDisabled = errors.New("user account is disabled")

type LoginUserReadModel interface {
	GetByEmail(ctx context.Context, email string) (*readmodels.UserDTO, error)
	UpdateLastLogin(ctx context.Context, id uuid.UUID, lastLoginAt time.Time) error
}

type LoginInvitationReadModel interface {
	GetAnyPendingByEmail(ctx context.Context, email string) (*readmodels.InvitationDTO, error)
}

type LoginUserAggregateRepository interface {
	Save(ctx context.Context, user *aggregates.User) error
}

type LoginService struct {
	userReadModel       LoginUserReadModel
	invitationReadModel LoginInvitationReadModel
	commandBus          cqrs.CommandBus
	userAggregateRepo   LoginUserAggregateRepository
}

func NewLoginService(
	userReadModel LoginUserReadModel,
	invitationReadModel LoginInvitationReadModel,
	commandBus cqrs.CommandBus,
	userAggregateRepo LoginUserAggregateRepository,
) *LoginService {
	return &LoginService{
		userReadModel:       userReadModel,
		invitationReadModel: invitationReadModel,
		commandBus:          commandBus,
		userAggregateRepo:   userAggregateRepo,
	}
}

type LoginResult struct {
	UserID uuid.UUID
	Email  string
	Role   string
	IsNew  bool
}

func (s *LoginService) ProcessLogin(ctx context.Context, email, name string) (*LoginResult, error) {
	existingUser, err := s.userReadModel.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return s.loginExistingUser(ctx, existingUser)
	}

	invitation, err := s.invitationReadModel.GetAnyPendingByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if invitation == nil {
		return nil, ErrNoValidInvitation
	}
	if invitation.ExpiresAt.Before(time.Now().UTC()) {
		return s.expirePendingInvitation(ctx, invitation.ID)
	}

	return s.createUserFromInvitation(ctx, email, name, invitation)
}

func (s *LoginService) loginExistingUser(ctx context.Context, user *readmodels.UserDTO) (*LoginResult, error) {
	if user.Status == "disabled" {
		return nil, ErrUserDisabled
	}
	if err := s.userReadModel.UpdateLastLogin(ctx, user.ID, time.Now().UTC()); err != nil {
		return nil, err
	}
	return &LoginResult{UserID: user.ID, Email: user.Email, Role: user.Role, IsNew: false}, nil
}

func (s *LoginService) expirePendingInvitation(ctx context.Context, invitationID string) (*LoginResult, error) {
	if _, err := s.commandBus.Dispatch(ctx, &commands.MarkInvitationExpired{ID: invitationID}); err != nil {
		return nil, err
	}
	return nil, ErrNoValidInvitation
}

func (s *LoginService) createUserFromInvitation(ctx context.Context, email, name string, invitation *readmodels.InvitationDTO) (*LoginResult, error) {
	if _, err := s.commandBus.Dispatch(ctx, &commands.AcceptInvitation{Email: email}); err != nil {
		return nil, err
	}

	user, err := buildUserAggregate(email, name, invitation)
	if err != nil {
		return nil, err
	}
	if err := s.userAggregateRepo.Save(ctx, user); err != nil {
		return nil, err
	}

	newUserID, err := uuid.Parse(user.ID())
	if err != nil {
		return nil, err
	}
	return &LoginResult{UserID: newUserID, Email: email, Role: invitation.Role, IsNew: true}, nil
}

func buildUserAggregate(email, name string, invitation *readmodels.InvitationDTO) (*aggregates.User, error) {
	emailVO, err := valueobjects.NewEmail(email)
	if err != nil {
		return nil, err
	}
	role, err := valueobjects.RoleFromString(invitation.Role)
	if err != nil {
		return nil, err
	}
	profile := valueobjects.NewExternalProfile(name, "")
	return aggregates.NewUser(emailVO, profile, role, invitation.ID)
}
