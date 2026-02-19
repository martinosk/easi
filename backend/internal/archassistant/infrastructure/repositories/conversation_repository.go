package repositories

import (
	"context"
	"database/sql"
	"time"

	"easi/backend/internal/archassistant/domain"
	"easi/backend/internal/archassistant/domain/aggregates"
	vo "easi/backend/internal/archassistant/domain/valueobjects"
	"easi/backend/internal/infrastructure/database"
)

type ConversationRepository struct {
	db *database.TenantAwareDB
}

func NewConversationRepository(db *database.TenantAwareDB) domain.ConversationRepository {
	return &ConversationRepository{db: db}
}

func (r *ConversationRepository) Create(ctx context.Context, conv *aggregates.Conversation) error {
	return r.db.WithTenantContext(ctx, func(conn *sql.Conn) error {
		_, err := conn.ExecContext(ctx, `
			INSERT INTO archassistant.conversations (id, tenant_id, user_id, title, created_at, last_message_at)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, conv.ID(), conv.TenantID(), conv.UserID(), conv.Title(), conv.CreatedAt(), conv.LastMessageAt())
		return err
	})
}

func (r *ConversationRepository) GetByIDAndUser(ctx context.Context, id, userID string) (*aggregates.Conversation, error) {
	var conv *aggregates.Conversation
	err := r.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		row := tx.QueryRowContext(ctx, `
			SELECT id, tenant_id, user_id, title, created_at, last_message_at
			FROM archassistant.conversations
			WHERE id = $1 AND user_id = $2
		`, id, userID)

		c, err := scanConversation(row)
		if err == sql.ErrNoRows {
			return nil
		}
		if err != nil {
			return err
		}
		conv = c
		return nil
	})
	return conv, err
}

func (r *ConversationRepository) SaveMessage(ctx context.Context, msg *aggregates.Message) error {
	return r.db.WithTenantContext(ctx, func(conn *sql.Conn) error {
		_, err := conn.ExecContext(ctx, `
			INSERT INTO archassistant.messages (id, conversation_id, tenant_id, role, content, tokens_used, created_at)
			VALUES ($1, $2, current_setting('app.current_tenant', true), $3, $4, $5, $6)
		`, msg.ID(), msg.ConversationID(), msg.Role().String(), msg.Content(), msg.TokensUsed(), msg.CreatedAt())
		return err
	})
}

func (r *ConversationRepository) GetMessages(ctx context.Context, conversationID string) ([]*aggregates.Message, error) {
	var messages []*aggregates.Message
	err := r.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, `
			SELECT id, conversation_id, role, content, tokens_used, created_at
			FROM archassistant.messages
			WHERE conversation_id = $1
			ORDER BY created_at ASC
		`, conversationID)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			msg, err := scanMessage(rows)
			if err != nil {
				return err
			}
			messages = append(messages, msg)
		}
		return rows.Err()
	})
	return messages, err
}

func (r *ConversationRepository) UpdateConversation(ctx context.Context, conv *aggregates.Conversation) error {
	return r.db.WithTenantContext(ctx, func(conn *sql.Conn) error {
		_, err := conn.ExecContext(ctx, `
			UPDATE archassistant.conversations
			SET title = $1, last_message_at = $2
			WHERE id = $3
		`, conv.Title(), conv.LastMessageAt(), conv.ID())
		return err
	})
}

func scanConversation(s scanner) (*aggregates.Conversation, error) {
	var (
		id, tenantID, userID, title string
		createdAt, lastMessageAt    time.Time
	)

	err := s.Scan(&id, &tenantID, &userID, &title, &createdAt, &lastMessageAt)
	if err != nil {
		return nil, err
	}

	return aggregates.ReconstructConversation(aggregates.ReconstructConversationParams{
		ID:            id,
		TenantID:      tenantID,
		UserID:        userID,
		Title:         title,
		CreatedAt:     createdAt,
		LastMessageAt: lastMessageAt,
	}), nil
}

func scanMessage(s scanner) (*aggregates.Message, error) {
	var (
		id, conversationID, roleStr, content string
		tokensUsed                           *int
		createdAt                            time.Time
	)

	err := s.Scan(&id, &conversationID, &roleStr, &content, &tokensUsed, &createdAt)
	if err != nil {
		return nil, err
	}

	return aggregates.ReconstructMessage(aggregates.ReconstructMessageParams{
		ID:             id,
		ConversationID: conversationID,
		Role:           vo.MessageRole(roleStr),
		Content:        content,
		TokensUsed:     tokensUsed,
		CreatedAt:      createdAt,
	}), nil
}
