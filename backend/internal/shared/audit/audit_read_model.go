package audit

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
)

type AuditHistoryReadModel struct {
	db *database.TenantAwareDB
}

func NewAuditHistoryReadModel(db *database.TenantAwareDB) *AuditHistoryReadModel {
	return &AuditHistoryReadModel{db: db}
}

func (rm *AuditHistoryReadModel) GetHistoryByAggregateID(ctx context.Context, aggregateID string, limit int, cursor string) ([]AuditEntry, bool, string, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, false, "", fmt.Errorf("failed to get tenant from context: %w", err)
	}

	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	var entries []AuditEntry
	var hasMore bool
	var nextCursor string

	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		query := `
			SELECT id, aggregate_id, event_type, event_data, version, occurred_at, actor_id, actor_email
			FROM events
			WHERE tenant_id = $1 AND aggregate_id = $2
		`
		args := []interface{}{tenantID.Value(), aggregateID}

		if cursor != "" {
			cursorID, err := strconv.ParseInt(cursor, 10, 64)
			if err == nil {
				query += " AND id < $3"
				args = append(args, cursorID)
			}
		}

		query += " ORDER BY id DESC LIMIT $" + strconv.Itoa(len(args)+1)
		args = append(args, limit+1)

		rows, err := tx.QueryContext(ctx, query, args...)
		if err != nil {
			return fmt.Errorf("failed to query events: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var entry AuditEntry
			var eventDataJSON string

			if err := rows.Scan(
				&entry.EventID,
				&entry.AggregateID,
				&entry.EventType,
				&eventDataJSON,
				&entry.Version,
				&entry.OccurredAt,
				&entry.ActorID,
				&entry.ActorEmail,
			); err != nil {
				return fmt.Errorf("failed to scan event: %w", err)
			}

			if err := json.Unmarshal([]byte(eventDataJSON), &entry.EventData); err != nil {
				entry.EventData = map[string]interface{}{"raw": eventDataJSON}
			}

			entry.DisplayName = FormatEventTypeDisplayName(entry.EventType)

			entries = append(entries, entry)
		}

		return rows.Err()
	})

	if err != nil {
		return nil, false, "", err
	}

	if len(entries) > limit {
		hasMore = true
		nextCursor = strconv.FormatInt(entries[limit-1].EventID, 10)
		entries = entries[:limit]
	}

	return entries, hasMore, nextCursor, nil
}
