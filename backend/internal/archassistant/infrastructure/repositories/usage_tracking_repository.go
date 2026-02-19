package repositories

import (
	"context"
	"database/sql"
	"time"

	"easi/backend/internal/infrastructure/database"

	"github.com/google/uuid"
)

type UsageRecord struct {
	TenantID         string
	UserID           string
	ConversationID   string
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
	ToolCallsCount   int
	ModelUsed        string
	LatencyMs        int
}

type UsageTrackingRepository struct {
	db *database.TenantAwareDB
}

func NewUsageTrackingRepository(db *database.TenantAwareDB) *UsageTrackingRepository {
	return &UsageTrackingRepository{db: db}
}

func (r *UsageTrackingRepository) Save(ctx context.Context, record UsageRecord) error {
	return r.db.WithTenantContext(ctx, func(conn *sql.Conn) error {
		_, err := conn.ExecContext(ctx, `
			INSERT INTO archassistant.usage_tracking
				(id, tenant_id, user_id, conversation_id, prompt_tokens, completion_tokens, total_tokens, tool_calls_count, model_used, latency_ms, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		`, uuid.New().String(), record.TenantID, record.UserID, record.ConversationID,
			record.PromptTokens, record.CompletionTokens, record.TotalTokens,
			record.ToolCallsCount, record.ModelUsed, record.LatencyMs, time.Now())
		return err
	})
}
