package eventstore

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/events"
	domain "easi/backend/internal/shared/eventsourcing"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
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
	db       *database.TenantAwareDB
	eventBus events.EventBus
}

// NewPostgresEventStore creates a new PostgreSQL event store
func NewPostgresEventStore(db *database.TenantAwareDB) *PostgresEventStore {
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
	ID          int64
	AggregateID string
	EventType   string
	EventData   string
	Version     int
	OccurredAt  time.Time
	CreatedAt   time.Time
}

// SaveEvents saves events to the event store
func (s *PostgresEventStore) SaveEvents(ctx context.Context, aggregateID string, events []domain.DomainEvent, expectedVersion int) error {
	if len(events) == 0 {
		return nil
	}

	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return fmt.Errorf("failed to get tenant from context: %w", err)
	}

	tx, err := s.db.BeginTxWithTenant(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := s.checkVersionConflict(ctx, tx, tenantID, aggregateID, expectedVersion); err != nil {
		return err
	}

	if err := s.insertEvents(ctx, tx, tenantID, events, expectedVersion); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.publishEventsIfAvailable(ctx, events)

	return nil
}

func (s *PostgresEventStore) checkVersionConflict(ctx context.Context, tx *sql.Tx, tenantID sharedvo.TenantID, aggregateID string, expectedVersion int) error {
	var currentVersion int
	err := tx.QueryRowContext(ctx,
		"SELECT COALESCE(MAX(version), 0) FROM infrastructure.events WHERE tenant_id = $1 AND aggregate_id = $2",
		tenantID.Value(),
		aggregateID,
	).Scan(&currentVersion)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to check current version: %w", err)
	}

	if currentVersion != expectedVersion {
		return fmt.Errorf("%w: expected version %d, got %d", domain.ErrConcurrencyConflict, expectedVersion, currentVersion)
	}

	return nil
}

func (s *PostgresEventStore) insertEvents(ctx context.Context, tx *sql.Tx, tenantID sharedvo.TenantID, events []domain.DomainEvent, expectedVersion int) error {
	actor, hasActor := sharedctx.GetActor(ctx)

	stmt, err := tx.PrepareContext(ctx,
		"INSERT INTO infrastructure.events (tenant_id, aggregate_id, event_type, event_data, version, occurred_at, actor_id, actor_email) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
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

		actorID := actor.ID
		actorEmail := actor.Email
		if !hasActor {
			actorID = "system"
			actorEmail = "system@easi.app"
		}

		_, err = stmt.ExecContext(ctx,
			tenantID.Value(),
			event.AggregateID(),
			event.EventType(),
			eventData,
			version,
			event.OccurredAt(),
			actorID,
			actorEmail,
		)
		if err != nil {
			return fmt.Errorf("failed to insert event: %w", err)
		}
	}

	return nil
}

func (s *PostgresEventStore) publishEventsIfAvailable(ctx context.Context, events []domain.DomainEvent) {
	if s.eventBus != nil {
		if err := s.eventBus.Publish(ctx, events); err != nil {
			fmt.Printf("Warning: failed to publish events to event bus: %v\n", err)
		}
	}
}

// GetEvents retrieves all events for an aggregate
func (s *PostgresEventStore) GetEvents(ctx context.Context, aggregateID string) ([]domain.DomainEvent, error) {
	// Extract tenant from context - this is infrastructure concern
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant from context: %w", err)
	}

	// Use tenant-aware read-only transaction that sets app.current_tenant for RLS
	var storedEvents []StoredEvent
	err = s.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx,
			"SELECT id, aggregate_id, event_type, event_data, version, occurred_at, created_at FROM infrastructure.events WHERE tenant_id = $1 AND aggregate_id = $2 ORDER BY version ASC",
			tenantID.Value(),
			aggregateID,
		)
		if err != nil {
			return fmt.Errorf("failed to query events: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var se StoredEvent
			if err := rows.Scan(&se.ID, &se.AggregateID, &se.EventType, &se.EventData, &se.Version, &se.OccurredAt, &se.CreatedAt); err != nil {
				return fmt.Errorf("failed to scan event: %w", err)
			}
			storedEvents = append(storedEvents, se)
		}

		return rows.Err()
	})

	if err != nil {
		return nil, fmt.Errorf("error retrieving events: %w", err)
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
