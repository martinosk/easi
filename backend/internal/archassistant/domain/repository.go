package domain

import (
	"context"

	"easi/backend/internal/archassistant/domain/aggregates"
)

type AIConfigurationRepository interface {
	GetByTenantID(ctx context.Context) (*aggregates.AIConfiguration, error)
	Save(ctx context.Context, config *aggregates.AIConfiguration) error
}

type ConversationRepository interface {
	Create(ctx context.Context, conv *aggregates.Conversation) error
	GetByIDAndUser(ctx context.Context, id, userID string) (*aggregates.Conversation, error)
	SaveMessage(ctx context.Context, msg *aggregates.Message) error
	GetMessages(ctx context.Context, conversationID string) ([]*aggregates.Message, error)
	UpdateConversation(ctx context.Context, conv *aggregates.Conversation) error
}
