package readmodels

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type AuditEntry struct {
	EventID     int64                  `json:"eventId"`
	AggregateID string                 `json:"aggregateId"`
	EventType   string                 `json:"eventType"`
	DisplayName string                 `json:"displayName"`
	EventData   map[string]interface{} `json:"eventData"`
	OccurredAt  time.Time              `json:"occurredAt"`
	Version     int                    `json:"version"`
	ActorID     string                 `json:"actorId"`
	ActorEmail  string                 `json:"actorEmail"`
}

var camelCaseRegex = regexp.MustCompile("([a-z])([A-Z])")

func FormatEventTypeDisplayName(eventType string) string {
	parts := strings.Split(eventType, ".")
	action := parts[len(parts)-1]

	spaced := camelCaseRegex.ReplaceAllString(action, "${1} ${2}")

	return cases.Title(language.Und).String(strings.ToLower(spaced))
}

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

	limit = normalizeHistoryLimit(limit)

	params := historyQueryParams{
		tenantID:    tenantID.Value(),
		aggregateID: aggregateID,
		cursor:      cursor,
		limit:       limit,
	}

	var entries []AuditEntry
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		history, queryErr := queryAuditHistory(ctx, tx, params)
		entries = history
		return queryErr
	})
	if err != nil {
		return nil, false, "", err
	}

	page, hasMore, nextCursor := paginateAuditEntries(entries, limit)
	return page, hasMore, nextCursor, nil
}

func normalizeHistoryLimit(limit int) int {
	if limit <= 0 {
		return 50
	}
	if limit > 100 {
		return 100
	}
	return limit
}

type historyQueryParams struct {
	tenantID    string
	aggregateID string
	cursor      string
	limit       int
}

func buildHistoryQuery(p historyQueryParams) (string, []any) {
	query := `
		SELECT id, aggregate_id, event_type, event_data, version, occurred_at, actor_id, actor_email
		FROM infrastructure.events
		WHERE tenant_id = $1 AND (aggregate_id = $2 OR event_data->>'componentId' = $2)
	`
	args := []any{p.tenantID, p.aggregateID}

	if p.cursor != "" {
		if cursorID, err := strconv.ParseInt(p.cursor, 10, 64); err == nil {
			query += " AND id < $3"
			args = append(args, cursorID)
		}
	}

	query += " ORDER BY id DESC LIMIT $" + strconv.Itoa(len(args)+1)
	args = append(args, p.limit+1)

	return query, args
}

func scanAuditEntry(rows *sql.Rows) (AuditEntry, error) {
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
		return AuditEntry{}, fmt.Errorf("failed to scan event: %w", err)
	}

	if err := json.Unmarshal([]byte(eventDataJSON), &entry.EventData); err != nil {
		entry.EventData = map[string]any{"raw": eventDataJSON}
	}

	entry.DisplayName = FormatEventTypeDisplayName(entry.EventType)

	return entry, nil
}

func queryAuditHistory(ctx context.Context, tx *sql.Tx, p historyQueryParams) ([]AuditEntry, error) {
	query, args := buildHistoryQuery(p)

	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var entries []AuditEntry
	for rows.Next() {
		entry, err := scanAuditEntry(rows)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	return entries, rows.Err()
}

func paginateAuditEntries(entries []AuditEntry, limit int) ([]AuditEntry, bool, string) {
	if len(entries) <= limit {
		return entries, false, ""
	}
	nextCursor := strconv.FormatInt(entries[limit-1].EventID, 10)
	return entries[:limit], true, nextCursor
}
