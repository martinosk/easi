//go:build integration
// +build integration

package eventstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/events"
	domain "easi/backend/internal/shared/eventsourcing"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

type atomicityTestContext struct {
	db       *sql.DB
	tenantDB *database.TenantAwareDB
	tenantID sharedvo.TenantID
	ctx      context.Context
}

func setupAtomicityTest(t *testing.T) (*atomicityTestContext, func()) {
	t.Helper()

	dbHost := getEnv("INTEGRATION_TEST_DB_HOST", "localhost")
	dbPort := getEnv("INTEGRATION_TEST_DB_PORT", "5432")
	dbUser := getEnv("INTEGRATION_TEST_DB_USER", "easi_app")
	dbPassword := getEnv("INTEGRATION_TEST_DB_PASSWORD", "localdev")
	dbName := getEnv("INTEGRATION_TEST_DB_NAME", "easi")
	dbSSLMode := getEnv("INTEGRATION_TEST_DB_SSLMODE", "disable")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbPassword, dbName, dbSSLMode)
	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err)
	require.NoError(t, db.Ping())

	tenantIDStr := fmt.Sprintf("evtx-%d", time.Now().UnixNano())
	_, err = db.Exec(
		`INSERT INTO auth.tenants (id, name, status, created_at, updated_at) VALUES ($1, $2, 'active', NOW(), NOW())`,
		tenantIDStr, "Atomicity Test Tenant",
	)
	require.NoError(t, err)

	tenantID, err := sharedvo.NewTenantID(tenantIDStr)
	require.NoError(t, err)

	tenantDB := database.NewTenantAwareDB(db)
	ctx := sharedctx.WithTenant(context.Background(), tenantID)

	cleanup := func() {
		tx, err := db.Begin()
		if err == nil {
			_, _ = tx.Exec(fmt.Sprintf("SET LOCAL app.current_tenant = '%s'", tenantIDStr))
			_, _ = tx.Exec(`DELETE FROM auth.users WHERE tenant_id = $1`, tenantIDStr)
			_, _ = tx.Exec(`DELETE FROM infrastructure.events WHERE tenant_id = $1`, tenantIDStr)
			_, _ = tx.Exec(`DELETE FROM auth.tenants WHERE id = $1`, tenantIDStr)
			_ = tx.Commit()
		}
		_ = db.Close()
	}

	return &atomicityTestContext{db: db, tenantDB: tenantDB, tenantID: tenantID, ctx: ctx}, cleanup
}

func (tc *atomicityTestContext) queryScopedToTenant(t *testing.T, query string, args ...interface{}) int {
	t.Helper()

	tx, err := tc.db.Begin()
	require.NoError(t, err)
	defer func() { _ = tx.Rollback() }()

	_, err = tx.Exec(fmt.Sprintf("SET LOCAL app.current_tenant = '%s'", tc.tenantID.Value()))
	require.NoError(t, err)

	var count int
	require.NoError(t, tx.QueryRow(query, args...).Scan(&count))
	return count
}

func (tc *atomicityTestContext) eventCount(t *testing.T, aggregateID string) int {
	t.Helper()
	return tc.queryScopedToTenant(t,
		`SELECT COUNT(*) FROM infrastructure.events WHERE tenant_id = $1 AND aggregate_id = $2`,
		tc.tenantID.Value(), aggregateID,
	)
}

func (tc *atomicityTestContext) userCount(t *testing.T, email string) int {
	t.Helper()
	return tc.queryScopedToTenant(t,
		`SELECT COUNT(*) FROM auth.users WHERE tenant_id = $1 AND email = $2`,
		tc.tenantID.Value(), email,
	)
}

func insertTestUser(ctx context.Context, tenantDB *database.TenantAwareDB, tenantID, email string) error {
	_, err := tenantDB.ExecContext(ctx,
		`INSERT INTO auth.users (id, tenant_id, email, role, status, created_at) VALUES (gen_random_uuid(), $1, $2, 'admin', 'active', NOW())`,
		tenantID, email,
	)
	return err
}

func newAtomicityScenario(t *testing.T) (*atomicityTestContext, *events.InMemoryEventBus, *PostgresEventStore) {
	t.Helper()

	tc, cleanup := setupAtomicityTest(t)
	t.Cleanup(cleanup)

	eventBus := events.NewInMemoryEventBus()
	store := NewPostgresEventStore(tc.tenantDB)
	store.SetEventBus(eventBus)

	return tc, eventBus, store
}

func failingHandler(message string) events.EventHandlerFunc {
	return func(context.Context, domain.DomainEvent) error {
		return errors.New(message)
	}
}

func expectedRowCountFor(t *testing.T, err error, injectFault bool) int {
	t.Helper()

	if injectFault {
		require.Error(t, err)
		return 0
	}
	require.NoError(t, err)
	return 1
}

func TestSaveEvents_SubscriberOutcome(t *testing.T) {
	tests := []struct {
		name        string
		injectFault bool
	}{
		{name: "failure rolls back event and projector write", injectFault: true},
		{name: "success commits event and projector write", injectFault: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc, eventBus, store := newAtomicityScenario(t)
			aggregateID := "primary-" + tc.tenantID.Value()
			email := "atomicity-" + tc.tenantID.Value() + "@test.local"

			eventBus.Subscribe("PrimaryCreated", events.EventHandlerFunc(func(ctx context.Context, _ domain.DomainEvent) error {
				return insertTestUser(ctx, tc.tenantDB, tc.tenantID.Value(), email)
			}))
			if tt.injectFault {
				eventBus.Subscribe("PrimaryCreated", failingHandler("simulated projector failure"))
			}

			evt := NewMockEvent(aggregateID, "PrimaryCreated", map[string]interface{}{"id": aggregateID})
			err := store.SaveEvents(tc.ctx, aggregateID, []domain.DomainEvent{evt}, 0)

			wantCount := expectedRowCountFor(t, err, tt.injectFault)
			require.Equal(t, wantCount, tc.eventCount(t, aggregateID))
			require.Equal(t, wantCount, tc.userCount(t, email))
		})
	}
}

func TestSaveEvents_NestedDispatchOutcome(t *testing.T) {
	tests := []struct {
		name        string
		injectFault bool
	}{
		{name: "failure rolls back both aggregates", injectFault: true},
		{name: "success commits both aggregates", injectFault: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc, eventBus, store := newAtomicityScenario(t)
			primaryID := "nested-primary-" + tc.tenantID.Value()
			secondaryID := "nested-secondary-" + tc.tenantID.Value()

			eventBus.Subscribe("PrimaryCreated", events.EventHandlerFunc(func(ctx context.Context, _ domain.DomainEvent) error {
				evt := NewMockEvent(secondaryID, "SecondaryCreated", map[string]interface{}{"id": secondaryID})
				return store.SaveEvents(ctx, secondaryID, []domain.DomainEvent{evt}, 0)
			}))
			if tt.injectFault {
				eventBus.Subscribe("SecondaryCreated", failingHandler("simulated secondary subscriber failure"))
			}

			evt := NewMockEvent(primaryID, "PrimaryCreated", map[string]interface{}{"id": primaryID})
			err := store.SaveEvents(tc.ctx, primaryID, []domain.DomainEvent{evt}, 0)

			wantCount := expectedRowCountFor(t, err, tt.injectFault)
			require.Equal(t, wantCount, tc.eventCount(t, primaryID))
			require.Equal(t, wantCount, tc.eventCount(t, secondaryID))
		})
	}
}
