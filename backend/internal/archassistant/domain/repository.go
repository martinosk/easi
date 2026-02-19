package domain

import (
	"context"
	"errors"

	"easi/backend/internal/archassistant/domain/aggregates"
)

var ErrConversationNotFound = errors.New("conversation not found")

type AIConfigurationRepository interface {
	GetByTenantID(ctx context.Context) (*aggregates.AIConfiguration, error)
	Save(ctx context.Context, config *aggregates.AIConfiguration) error
}

type ListConversationsParams struct {
	UserID string
	Limit  int
	Offset int
}

type ConversationRepository interface {
	Create(ctx context.Context, conv *aggregates.Conversation) error
	GetByIDAndUser(ctx context.Context, id, userID string) (*aggregates.Conversation, error)
	SaveMessage(ctx context.Context, msg *aggregates.Message) error
	GetMessages(ctx context.Context, conversationID string) ([]*aggregates.Message, error)
	UpdateConversation(ctx context.Context, conv *aggregates.Conversation) error
	ListByUser(ctx context.Context, params ListConversationsParams) ([]*aggregates.Conversation, int, error)
	Delete(ctx context.Context, id, userID string) error
	CountByUser(ctx context.Context, userID string) (int, error)
}
