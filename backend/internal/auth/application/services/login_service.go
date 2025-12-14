package services

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"easi/backend/internal/auth/application/commands"
	"easi/backend/internal/auth/application/readmodels"
	"easi/backend/internal/shared/cqrs"
)

var ErrNoValidInvitation = errors.New("no valid invitation found for this email")

type LoginService struct {
	userReadModel       *readmodels.UserReadModel
	invitationReadModel *readmodels.InvitationReadModel
	commandBus          cqrs.CommandBus
}

func NewLoginService(
	userReadModel *readmodels.UserReadModel,
	invitationReadModel *readmodels.InvitationReadModel,
	commandBus cqrs.CommandBus,
) *LoginService {
	return &LoginService{
		userReadModel:       userReadModel,
		invitationReadModel: invitationReadModel,
		commandBus:          commandBus,
	}
}

type LoginResult struct {
	UserID uuid.UUID
	Email  string
	Role   string
	IsNew  bool
}

func (s *LoginService) ProcessLogin(ctx context.Context, email string) (*LoginResult, error) {
	existingUser, err := s.userReadModel.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if existingUser != nil {
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

	invitation, err := s.invitationReadModel.GetPendingByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if invitation == nil {
		return nil, ErrNoValidInvitation
	}

	cmd := &commands.AcceptInvitation{
		Email: email,
	}
	if err := s.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, err
	}

	newUserID := uuid.New()
	now := time.Now().UTC()

	userDTO := readmodels.UserDTO{
		ID:           newUserID,
		Email:        email,
		Role:         invitation.Role,
		Status:       "active",
		InvitationID: parseUUID(invitation.ID),
		CreatedAt:    now,
		LastLoginAt:  &now,
	}

	if err := s.userReadModel.Insert(ctx, userDTO); err != nil {
		return nil, err
	}

	return &LoginResult{
		UserID: newUserID,
		Email:  email,
		Role:   invitation.Role,
		IsNew:  true,
	}, nil
}

func parseUUID(s string) *uuid.UUID {
	id, err := uuid.Parse(s)
	if err != nil {
		return nil
	}
	return &id
}
