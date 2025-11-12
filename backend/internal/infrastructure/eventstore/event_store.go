package eventstore

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"easi/backend/internal/shared/domain"
	"easi/backend/internal/shared/events"
	sharedctx "easi/backend/internal/shared/context"
)

// EventStore defines the interface for event storage
type EventStore interface {
	// SaveEvents saves events for an aggregate
	SaveEvents(ctx context.Context, aggregateID string, events []domain.DomainEvent, expectedVersion int) error

	// GetEvents retrieves all events for an aggregate
	GetEvents(ctx context.Context, aggregateID string) ([]domain.DomainEvent, error)
}

// PostgresEventStore implements EventStore using PostgreSQL
type PostgresEventStore struct {
	db       *sql.DB
	eventBus events.EventBus
}

// NewPostgresEventStore creates a new PostgreSQL event store
func NewPostgresEventStore(db *sql.DB) *PostgresEventStore {
	return &PostgresEventStore{
		db:       db,
		eventBus: nil, // Will be set via SetEventBus
	}
}

// SetEventBus sets the event bus for publishing events after they're saved
func (s *PostgresEventStore) SetEventBus(eventBus events.EventBus) {
	s.eventBus = eventBus
}

// StoredEvent represents an event as stored in the database
type StoredEvent struct {
	ID            int64
	AggregateID   string
	EventType     string
	EventData     string
	Version       int
	OccurredAt    time.Time
	CreatedAt     time.Time
}

// SaveEvents saves events to the event store
func (s *PostgresEventStore) SaveEvents(ctx context.Context, aggregateID string, events []domain.DomainEvent, expectedVersion int) error {
	if len(events) == 0 {
		return nil
	}

	// Extract tenant from context - this is infrastructure concern
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return fmt.Errorf("failed to get tenant from context: %w", err)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Check current version - filtered by tenant
	var currentVersion int
	err = tx.QueryRowContext(ctx,
		"SELECT COALESCE(MAX(version), 0) FROM events WHERE tenant_id = $1 AND aggregate_id = $2",
		tenantID.Value(),
		aggregateID,
	).Scan(&currentVersion)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to check current version: %w", err)
	}

	if currentVersion != expectedVersion {
		return fmt.Errorf("concurrency conflict: expected version %d, got %d", expectedVersion, currentVersion)
	}

	// Insert events with tenant_id
	stmt, err := tx.PrepareContext(ctx,
		"INSERT INTO events (tenant_id, aggregate_id, event_type, event_data, version, occurred_at) VALUES ($1, $2, $3, $4, $5, $6)",
	)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for i, event := range events {
		eventData, err := json.Marshal(event.EventData())
		if err != nil {
			return fmt.Errorf("failed to marshal event data: %w", err)
		}

		version := expectedVersion + i + 1
		_, err = stmt.ExecContext(ctx,
			tenantID.Value(), // Infrastructure adds tenant_id
			event.AggregateID(),
			event.EventType(),
			eventData,
			version,
			event.OccurredAt(),
		)
		if err != nil {
			return fmt.Errorf("failed to insert event: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Publish events to event bus after successful commit
	if s.eventBus != nil {
		if err := s.eventBus.Publish(ctx, events); err != nil {
			// Log the error but don't fail the operation since events are already persisted
			// In a production system, you might want to implement retry logic or dead letter queue
			fmt.Printf("Warning: failed to publish events to event bus: %v\n", err)
		}
	}

	return nil
}

// GetEvents retrieves all events for an aggregate
func (s *PostgresEventStore) GetEvents(ctx context.Context, aggregateID string) ([]domain.DomainEvent, error) {
	// Extract tenant from context - this is infrastructure concern
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant from context: %w", err)
	}

	rows, err := s.db.QueryContext(ctx,
		"SELECT id, aggregate_id, event_type, event_data, version, occurred_at, created_at FROM events WHERE tenant_id = $1 AND aggregate_id = $2 ORDER BY version ASC",
		tenantID.Value(),
		aggregateID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}
	defer rows.Close()

	var storedEvents []StoredEvent
	for rows.Next() {
		var se StoredEvent
		if err := rows.Scan(&se.ID, &se.AggregateID, &se.EventType, &se.EventData, &se.Version, &se.OccurredAt, &se.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		storedEvents = append(storedEvents, se)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating events: %w", err)
	}

	// Convert stored events to domain events
	var domainEvents []domain.DomainEvent
	for _, se := range storedEvents {
		// Create a base domain event with the stored data
		// The event data is already in JSON format, we'll pass it as a generic event
		domainEvent := domain.NewGenericDomainEvent(
			se.AggregateID,
			se.EventType,
			[]byte(se.EventData),
			se.OccurredAt,
		)
		domainEvents = append(domainEvents, domainEvent)
	}

	return domainEvents, nil
}
