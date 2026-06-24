//go:build integration
// +build integration

package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	sharedctx "easi/backend/internal/shared/context"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testContext struct {
	db       *sql.DB
	tenantDB *TenantAwareDB
	testID   string
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func setupTestDB(t *testing.T) (*testContext, func()) {
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

	err = db.Ping()
	require.NoError(t, err)

	testID := fmt.Sprintf("test-%d", time.Now().UnixNano())
	tenantDB := NewTenantAwareDB(db)

	ctx := &testContext{
		db:       db,
		tenantDB: tenantDB,
		testID:   testID,
	}

	cleanup := func() {
		db.Exec(fmt.Sprintf("DELETE FROM infrastructure.events WHERE aggregate_id LIKE '%s%%'", testID))
		db.Exec(fmt.Sprintf("DELETE FROM architecturemodeling.application_components WHERE id LIKE '%s%%'", testID))
		db.Exec(fmt.Sprintf("DELETE FROM capabilitymapping.capabilities WHERE id LIKE '%s%%'", testID))
		db.Close()
	}

	return ctx, cleanup
}

func (tc *testContext) uniqueID(suffix string) string {
	return fmt.Sprintf("%s-%s", tc.testID, suffix)
}

func mustTenants(t *testing.T) (sharedvo.TenantID, sharedvo.TenantID) {
	tenantA, err := sharedvo.NewTenantID("tenant-a")
	require.NoError(t, err)
	tenantB, err := sharedvo.NewTenantID("tenant-b")
	require.NoError(t, err)
	return tenantA, tenantB
}

type eventRow struct {
	tenant      sharedvo.TenantID
	aggregateID string
	eventType   string
	eventJSON   []byte
	version     int
}

func (tc *testContext) insertEvent(t *testing.T, ctx context.Context, e eventRow) {
	tx, err := tc.tenantDB.BeginTxWithTenant(ctx, nil)
	require.NoError(t, err)
	_, err = tx.ExecContext(ctx,
		"INSERT INTO infrastructure.events (tenant_id, aggregate_id, event_type, event_data, version, occurred_at, actor_id, actor_email) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
		e.tenant.Value(), e.aggregateID, e.eventType, e.eventJSON, e.version, time.Now(), "test-user-id", "test@example.com",
	)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
}

func (tc *testContext) scanInt(t *testing.T, ctx context.Context, query string, args ...interface{}) int {
	var result int
	err := tc.tenantDB.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx, query, args...).Scan(&result)
	})
	require.NoError(t, err)
	return result
}

func (tc *testContext) scanString(t *testing.T, ctx context.Context, query string, args ...interface{}) string {
	var result string
	err := tc.tenantDB.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx, query, args...).Scan(&result)
	})
	require.NoError(t, err)
	return result
}

func (tc *testContext) queryNames(t *testing.T, ctx context.Context, query string, args ...interface{}) []string {
	var names []string
	err := tc.tenantDB.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, query, args...)
		if err != nil {
			return err
		}
		defer func() { _ = rows.Close() }()
		names, err = scanStringRows(rows)
		if err != nil {
			return err
		}
		return rows.Err()
	})
	require.NoError(t, err)
	return names
}

func scanStringRows(rows *sql.Rows) ([]string, error) {
	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		names = append(names, name)
	}
	return names, nil
}

func TestTenantIsolation_ReadModelData(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cleanup := setupTestDB(t)
	defer cleanup()

	tenantA, tenantB := mustTenants(t)

	componentIDTenantA := ctx.uniqueID("comp-a")
	componentIDTenantB := ctx.uniqueID("comp-b")

	ctxA := sharedctx.WithTenant(context.Background(), tenantA)
	ctxB := sharedctx.WithTenant(context.Background(), tenantB)

	_, err := ctx.tenantDB.ExecContext(ctxA,
		"INSERT INTO architecturemodeling.application_components (id, tenant_id, name, description, created_at) VALUES ($1, $2, $3, $4, $5)",
		componentIDTenantA, tenantA.Value(), "Component A", "Tenant A Component", time.Now(),
	)
	require.NoError(t, err)

	_, err = ctx.tenantDB.ExecContext(ctxB,
		"INSERT INTO architecturemodeling.application_components (id, tenant_id, name, description, created_at) VALUES ($1, $2, $3, $4, $5)",
		componentIDTenantB, tenantB.Value(), "Component B", "Tenant B Component", time.Now(),
	)
	require.NoError(t, err)

	const countByID = "SELECT COUNT(*) FROM architecturemodeling.application_components WHERE tenant_id = $1 AND id = $2"
	const nameByID = "SELECT name FROM architecturemodeling.application_components WHERE tenant_id = $1 AND id = $2"

	assert.Equal(t, 0, ctx.scanInt(t, ctxB, countByID, tenantB.Value(), componentIDTenantA))
	assert.Equal(t, 1, ctx.scanInt(t, ctxA, countByID, tenantA.Value(), componentIDTenantA))
	assert.Equal(t, "Component A", ctx.scanString(t, ctxA, nameByID, tenantA.Value(), componentIDTenantA))
	assert.Equal(t, "Component B", ctx.scanString(t, ctxB, nameByID, tenantB.Value(), componentIDTenantB))
}

func TestTenantIsolation_EventStoreData(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cleanup := setupTestDB(t)
	defer cleanup()

	tenantA, tenantB := mustTenants(t)

	aggregateID := ctx.uniqueID("aggregate-1")

	ctxA := sharedctx.WithTenant(context.Background(), tenantA)
	ctxB := sharedctx.WithTenant(context.Background(), tenantB)

	eventData := map[string]interface{}{
		"name":        "Test Component",
		"description": "Test Description",
	}
	eventJSON, err := json.Marshal(eventData)
	require.NoError(t, err)

	ctx.insertEvent(t, ctxA, eventRow{tenantA, aggregateID, "ComponentCreated", eventJSON, 1})

	const countByAggregate = "SELECT COUNT(*) FROM infrastructure.events WHERE tenant_id = $1 AND aggregate_id = $2"
	const eventTypeByAggregate = "SELECT event_type FROM infrastructure.events WHERE tenant_id = $1 AND aggregate_id = $2"

	assert.Equal(t, 0, ctx.scanInt(t, ctxB, countByAggregate, tenantB.Value(), aggregateID))
	assert.Equal(t, 1, ctx.scanInt(t, ctxA, countByAggregate, tenantA.Value(), aggregateID))
	assert.Equal(t, "ComponentCreated", ctx.scanString(t, ctxA, eventTypeByAggregate, tenantA.Value(), aggregateID))
}

func TestTenantIsolation_MultipleTablesConsistency(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cleanup := setupTestDB(t)
	defer cleanup()

	tenantA, tenantB := mustTenants(t)

	capabilityIDTenantA := ctx.uniqueID("cap-a")
	capabilityIDTenantB := ctx.uniqueID("cap-b")

	ctxA := sharedctx.WithTenant(context.Background(), tenantA)
	ctxB := sharedctx.WithTenant(context.Background(), tenantB)

	const insertCapability = "INSERT INTO capabilitymapping.capabilities (id, tenant_id, name, description, level, created_at) VALUES ($1, $2, $3, $4, $5, $6)"
	_, err := ctx.tenantDB.ExecContext(ctxA, insertCapability,
		capabilityIDTenantA, tenantA.Value(), "Capability A", "Tenant A Capability", "L1", time.Now(),
	)
	require.NoError(t, err)

	_, err = ctx.tenantDB.ExecContext(ctxB, insertCapability,
		capabilityIDTenantB, tenantB.Value(), "Capability B", "Tenant B Capability", "L1", time.Now(),
	)
	require.NoError(t, err)

	const namesByIDs = "SELECT name FROM capabilitymapping.capabilities WHERE tenant_id = $1 AND id IN ($2, $3)"

	resultsA := ctx.queryNames(t, ctxA, namesByIDs, tenantA.Value(), capabilityIDTenantA, capabilityIDTenantB)
	assert.Len(t, resultsA, 1)
	assert.Equal(t, "Capability A", resultsA[0])

	resultsB := ctx.queryNames(t, ctxB, namesByIDs, tenantB.Value(), capabilityIDTenantA, capabilityIDTenantB)
	assert.Len(t, resultsB, 1)
	assert.Equal(t, "Capability B", resultsB[0])
}

func TestMissingTenantContext_FailsSafely(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cleanup := setupTestDB(t)
	defer cleanup()

	emptyContext := context.Background()

	err := ctx.tenantDB.WithTenantContext(emptyContext, func(conn *sql.Conn) error {
		return nil
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get tenant from context")

	_, err = ctx.tenantDB.ExecContext(emptyContext, "SELECT 1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get tenant from context")

	err = ctx.tenantDB.WithReadOnlyTx(emptyContext, func(tx *sql.Tx) error {
		return nil
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get tenant from context")

	_, err = ctx.tenantDB.BeginTxWithTenant(emptyContext, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get tenant from context")
}

func TestDefaultTenantFallback(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cleanup := setupTestDB(t)
	defer cleanup()

	emptyContext := context.Background()

	tenantID := sharedctx.GetTenantOrDefault(emptyContext)
	assert.Equal(t, "default", tenantID.Value())
	assert.True(t, tenantID.IsDefault())

	defaultTenant := sharedvo.DefaultTenantID()
	ctxWithDefault := sharedctx.WithTenant(context.Background(), defaultTenant)

	componentID := ctx.uniqueID("default-comp")
	_, err := ctx.tenantDB.ExecContext(ctxWithDefault,
		"INSERT INTO architecturemodeling.application_components (id, tenant_id, name, description, created_at) VALUES ($1, $2, $3, $4, $5)",
		componentID, defaultTenant.Value(), "Default Component", "Default Tenant Component", time.Now(),
	)
	require.NoError(t, err)

	name := ctx.scanString(t, ctxWithDefault,
		"SELECT name FROM architecturemodeling.application_components WHERE tenant_id = $1 AND id = $2",
		defaultTenant.Value(), componentID,
	)
	assert.Equal(t, "Default Component", name)
}

func TestTenantContext_TransactionIsolation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cleanup := setupTestDB(t)
	defer cleanup()

	tenantA, tenantB := mustTenants(t)

	componentID := ctx.uniqueID("tx-comp")

	ctxA := sharedctx.WithTenant(context.Background(), tenantA)
	ctxB := sharedctx.WithTenant(context.Background(), tenantB)

	txA, err := ctx.tenantDB.BeginTxWithTenant(ctxA, nil)
	require.NoError(t, err)

	_, err = txA.ExecContext(ctxA,
		"INSERT INTO architecturemodeling.application_components (id, tenant_id, name, description, created_at) VALUES ($1, $2, $3, $4, $5)",
		componentID, tenantA.Value(), "Transaction Component", "Test", time.Now(),
	)
	require.NoError(t, err)

	const countByID = "SELECT COUNT(*) FROM architecturemodeling.application_components WHERE tenant_id = $1 AND id = $2"

	assert.Equal(t, 0, ctx.scanInt(t, ctxA, countByID, tenantA.Value(), componentID))

	require.NoError(t, txA.Commit())

	assert.Equal(t, 1, ctx.scanInt(t, ctxA, countByID, tenantA.Value(), componentID))
	assert.Equal(t, 0, ctx.scanInt(t, ctxB, countByID, tenantB.Value(), componentID))
}

func TestTenantContext_ReadOnlyTransaction(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cleanup := setupTestDB(t)
	defer cleanup()

	tenantA, err := sharedvo.NewTenantID("tenant-a")
	require.NoError(t, err)

	componentID := ctx.uniqueID("readonly-comp")

	ctxA := sharedctx.WithTenant(context.Background(), tenantA)

	_, err = ctx.tenantDB.ExecContext(ctxA,
		"INSERT INTO architecturemodeling.application_components (id, tenant_id, name, description, created_at) VALUES ($1, $2, $3, $4, $5)",
		componentID, tenantA.Value(), "Read Only Test", "Test", time.Now(),
	)
	require.NoError(t, err)

	name := ctx.scanString(t, ctxA,
		"SELECT name FROM architecturemodeling.application_components WHERE tenant_id = $1 AND id = $2",
		tenantA.Value(), componentID,
	)
	assert.Equal(t, "Read Only Test", name)

	err = ctx.tenantDB.WithReadOnlyTx(ctxA, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctxA,
			"UPDATE architecturemodeling.application_components SET name = $1 WHERE tenant_id = $2 AND id = $3",
			"Modified Name", tenantA.Value(), componentID,
		)
		return err
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "read-only")
}

func TestTenantContext_EventVersioning(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cleanup := setupTestDB(t)
	defer cleanup()

	tenantA, tenantB := mustTenants(t)

	aggregateID := ctx.uniqueID("versioned-agg")

	ctxA := sharedctx.WithTenant(context.Background(), tenantA)
	ctxB := sharedctx.WithTenant(context.Background(), tenantB)

	eventData := map[string]interface{}{"test": "data"}
	eventJSON, err := json.Marshal(eventData)
	require.NoError(t, err)

	ctx.insertEvent(t, ctxA, eventRow{tenantA, aggregateID, "Event1", eventJSON, 1})
	ctx.insertEvent(t, ctxB, eventRow{tenantB, aggregateID, "Event1", eventJSON, 1})

	const versionByAggregate = "SELECT version FROM infrastructure.events WHERE tenant_id = $1 AND aggregate_id = $2"

	assert.Equal(t, 1, ctx.scanInt(t, ctxA, versionByAggregate, tenantA.Value(), aggregateID))
	assert.Equal(t, 1, ctx.scanInt(t, ctxB, versionByAggregate, tenantB.Value(), aggregateID))
}
