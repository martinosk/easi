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
	"easi/backend/internal/auth/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

var ErrNoValidInvitation = errors.New("no valid invitation found for this email")
var ErrUserDisabled = errors.New("user account is disabled")

type LoginService struct {
	userReadModel       *readmodels.UserReadModel
	invitationReadModel *readmodels.InvitationReadModel
	commandBus          cqrs.CommandBus
	userAggregateRepo   *repositories.UserAggregateRepository
}

func NewLoginService(
	userReadModel *readmodels.UserReadModel,
	invitationReadModel *readmodels.InvitationReadModel,
	commandBus cqrs.CommandBus,
	userAggregateRepo *repositories.UserAggregateRepository,
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

	invitation, err := s.invitationReadModel.GetAnyPendingByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if invitation == nil {
		return nil, ErrNoValidInvitation
	}

	now := time.Now().UTC()
	if invitation.ExpiresAt.Before(now) {
		expireCmd := &commands.MarkInvitationExpired{
			ID: invitation.ID,
		}
		if err := s.commandBus.Dispatch(ctx, expireCmd); err != nil {
			return nil, err
		}
		return nil, ErrNoValidInvitation
	}

	cmd := &commands.AcceptInvitation{
		Email: email,
	}
	if err := s.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, err
	}

	emailVO, err := valueobjects.NewEmail(email)
	if err != nil {
		return nil, err
	}

	role, err := valueobjects.RoleFromString(invitation.Role)
	if err != nil {
		return nil, err
	}

	user, err := aggregates.NewUser(emailVO, name, role, "", invitation.ID)
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

	return &LoginResult{
		UserID: newUserID,
		Email:  email,
		Role:   invitation.Role,
		IsNew:  true,
	}, nil
}
