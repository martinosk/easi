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
		db.Exec(fmt.Sprintf("DELETE FROM events WHERE aggregate_id LIKE '%s%%'", testID))
		db.Exec(fmt.Sprintf("DELETE FROM application_components WHERE id LIKE '%s%%'", testID))
		db.Exec(fmt.Sprintf("DELETE FROM capabilities WHERE id LIKE '%s%%'", testID))
		db.Close()
	}

	return ctx, cleanup
}

func (tc *testContext) uniqueID(suffix string) string {
	return fmt.Sprintf("%s-%s", tc.testID, suffix)
}

func TestTenantIsolation_ReadModelData(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cleanup := setupTestDB(t)
	defer cleanup()

	tenantA, err := sharedvo.NewTenantID("tenant-a")
	require.NoError(t, err)

	tenantB, err := sharedvo.NewTenantID("tenant-b")
	require.NoError(t, err)

	componentIDTenantA := ctx.uniqueID("comp-a")
	componentIDTenantB := ctx.uniqueID("comp-b")

	ctxA := sharedctx.WithTenant(context.Background(), tenantA)
	ctxB := sharedctx.WithTenant(context.Background(), tenantB)

	_, err = ctx.tenantDB.ExecContext(ctxA,
		"INSERT INTO application_components (id, tenant_id, name, description, created_at) VALUES ($1, $2, $3, $4, $5)",
		componentIDTenantA, tenantA.Value(), "Component A", "Tenant A Component", time.Now(),
	)
	require.NoError(t, err)

	_, err = ctx.tenantDB.ExecContext(ctxB,
		"INSERT INTO application_components (id, tenant_id, name, description, created_at) VALUES ($1, $2, $3, $4, $5)",
		componentIDTenantB, tenantB.Value(), "Component B", "Tenant B Component", time.Now(),
	)
	require.NoError(t, err)

	var countTenantB int
	err = ctx.tenantDB.WithReadOnlyTx(ctxB, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctxB,
			"SELECT COUNT(*) FROM application_components WHERE tenant_id = $1 AND id = $2",
			tenantB.Value(), componentIDTenantA,
		).Scan(&countTenantB)
	})
	require.NoError(t, err)
	assert.Equal(t, 0, countTenantB)

	var countTenantA int
	err = ctx.tenantDB.WithReadOnlyTx(ctxA, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctxA,
			"SELECT COUNT(*) FROM application_components WHERE tenant_id = $1 AND id = $2",
			tenantA.Value(), componentIDTenantA,
		).Scan(&countTenantA)
	})
	require.NoError(t, err)
	assert.Equal(t, 1, countTenantA)

	var nameTenantA string
	err = ctx.tenantDB.WithReadOnlyTx(ctxA, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctxA,
			"SELECT name FROM application_components WHERE tenant_id = $1 AND id = $2",
			tenantA.Value(), componentIDTenantA,
		).Scan(&nameTenantA)
	})
	require.NoError(t, err)
	assert.Equal(t, "Component A", nameTenantA)

	var nameTenantB string
	err = ctx.tenantDB.WithReadOnlyTx(ctxB, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctxB,
			"SELECT name FROM application_components WHERE tenant_id = $1 AND id = $2",
			tenantB.Value(), componentIDTenantB,
		).Scan(&nameTenantB)
	})
	require.NoError(t, err)
	assert.Equal(t, "Component B", nameTenantB)
}

func TestTenantIsolation_EventStoreData(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cleanup := setupTestDB(t)
	defer cleanup()

	tenantA, err := sharedvo.NewTenantID("tenant-a")
	require.NoError(t, err)

	tenantB, err := sharedvo.NewTenantID("tenant-b")
	require.NoError(t, err)

	aggregateID := ctx.uniqueID("aggregate-1")

	ctxA := sharedctx.WithTenant(context.Background(), tenantA)
	ctxB := sharedctx.WithTenant(context.Background(), tenantB)

	eventData := map[string]interface{}{
		"name":        "Test Component",
		"description": "Test Description",
	}
	eventJSON, err := json.Marshal(eventData)
	require.NoError(t, err)

	tx, err := ctx.tenantDB.BeginTxWithTenant(ctxA, nil)
	require.NoError(t, err)

	_, err = tx.ExecContext(ctxA,
		"INSERT INTO events (tenant_id, aggregate_id, event_type, event_data, version, occurred_at, actor_id, actor_email) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
		tenantA.Value(), aggregateID, "ComponentCreated", eventJSON, 1, time.Now(), "test-user-id", "test@example.com",
	)
	require.NoError(t, err)

	err = tx.Commit()
	require.NoError(t, err)

	var countTenantB int
	err = ctx.tenantDB.WithReadOnlyTx(ctxB, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctxB,
			"SELECT COUNT(*) FROM events WHERE tenant_id = $1 AND aggregate_id = $2",
			tenantB.Value(), aggregateID,
		).Scan(&countTenantB)
	})
	require.NoError(t, err)
	assert.Equal(t, 0, countTenantB)

	var countTenantA int
	err = ctx.tenantDB.WithReadOnlyTx(ctxA, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctxA,
			"SELECT COUNT(*) FROM events WHERE tenant_id = $1 AND aggregate_id = $2",
			tenantA.Value(), aggregateID,
		).Scan(&countTenantA)
	})
	require.NoError(t, err)
	assert.Equal(t, 1, countTenantA)

	var eventType string
	err = ctx.tenantDB.WithReadOnlyTx(ctxA, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctxA,
			"SELECT event_type FROM events WHERE tenant_id = $1 AND aggregate_id = $2",
			tenantA.Value(), aggregateID,
		).Scan(&eventType)
	})
	require.NoError(t, err)
	assert.Equal(t, "ComponentCreated", eventType)
}

func TestTenantIsolation_MultipleTablesConsistency(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cleanup := setupTestDB(t)
	defer cleanup()

	tenantA, err := sharedvo.NewTenantID("tenant-a")
	require.NoError(t, err)

	tenantB, err := sharedvo.NewTenantID("tenant-b")
	require.NoError(t, err)

	capabilityIDTenantA := ctx.uniqueID("cap-a")
	capabilityIDTenantB := ctx.uniqueID("cap-b")

	ctxA := sharedctx.WithTenant(context.Background(), tenantA)
	ctxB := sharedctx.WithTenant(context.Background(), tenantB)

	_, err = ctx.tenantDB.ExecContext(ctxA,
		"INSERT INTO capabilities (id, tenant_id, name, description, level, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
		capabilityIDTenantA, tenantA.Value(), "Capability A", "Tenant A Capability", "L1", time.Now(),
	)
	require.NoError(t, err)

	_, err = ctx.tenantDB.ExecContext(ctxB,
		"INSERT INTO capabilities (id, tenant_id, name, description, level, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
		capabilityIDTenantB, tenantB.Value(), "Capability B", "Tenant B Capability", "L1", time.Now(),
	)
	require.NoError(t, err)

	var results []string
	err = ctx.tenantDB.WithReadOnlyTx(ctxA, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctxA,
			"SELECT name FROM capabilities WHERE tenant_id = $1 AND id IN ($2, $3)",
			tenantA.Value(), capabilityIDTenantA, capabilityIDTenantB,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
				return err
			}
			results = append(results, name)
		}
		return rows.Err()
	})
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "Capability A", results[0])

	results = nil
	err = ctx.tenantDB.WithReadOnlyTx(ctxB, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctxB,
			"SELECT name FROM capabilities WHERE tenant_id = $1 AND id IN ($2, $3)",
			tenantB.Value(), capabilityIDTenantA, capabilityIDTenantB,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
				return err
			}
			results = append(results, name)
		}
		return rows.Err()
	})
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "Capability B", results[0])
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
		"INSERT INTO application_components (id, tenant_id, name, description, created_at) VALUES ($1, $2, $3, $4, $5)",
		componentID, defaultTenant.Value(), "Default Component", "Default Tenant Component", time.Now(),
	)
	require.NoError(t, err)

	var name string
	err = ctx.tenantDB.WithReadOnlyTx(ctxWithDefault, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctxWithDefault,
			"SELECT name FROM application_components WHERE tenant_id = $1 AND id = $2",
			defaultTenant.Value(), componentID,
		).Scan(&name)
	})
	require.NoError(t, err)
	assert.Equal(t, "Default Component", name)
}

func TestTenantContext_TransactionIsolation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cleanup := setupTestDB(t)
	defer cleanup()

	tenantA, err := sharedvo.NewTenantID("tenant-a")
	require.NoError(t, err)

	tenantB, err := sharedvo.NewTenantID("tenant-b")
	require.NoError(t, err)

	componentID := ctx.uniqueID("tx-comp")

	ctxA := sharedctx.WithTenant(context.Background(), tenantA)
	ctxB := sharedctx.WithTenant(context.Background(), tenantB)

	txA, err := ctx.tenantDB.BeginTxWithTenant(ctxA, nil)
	require.NoError(t, err)

	_, err = txA.ExecContext(ctxA,
		"INSERT INTO application_components (id, tenant_id, name, description, created_at) VALUES ($1, $2, $3, $4, $5)",
		componentID, tenantA.Value(), "Transaction Component", "Test", time.Now(),
	)
	require.NoError(t, err)

	var countBeforeCommit int
	err = ctx.tenantDB.WithReadOnlyTx(ctxA, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctxA,
			"SELECT COUNT(*) FROM application_components WHERE tenant_id = $1 AND id = $2",
			tenantA.Value(), componentID,
		).Scan(&countBeforeCommit)
	})
	require.NoError(t, err)
	assert.Equal(t, 0, countBeforeCommit)

	err = txA.Commit()
	require.NoError(t, err)

	var countAfterCommit int
	err = ctx.tenantDB.WithReadOnlyTx(ctxA, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctxA,
			"SELECT COUNT(*) FROM application_components WHERE tenant_id = $1 AND id = $2",
			tenantA.Value(), componentID,
		).Scan(&countAfterCommit)
	})
	require.NoError(t, err)
	assert.Equal(t, 1, countAfterCommit)

	var countTenantB int
	err = ctx.tenantDB.WithReadOnlyTx(ctxB, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctxB,
			"SELECT COUNT(*) FROM application_components WHERE tenant_id = $1 AND id = $2",
			tenantB.Value(), componentID,
		).Scan(&countTenantB)
	})
	require.NoError(t, err)
	assert.Equal(t, 0, countTenantB)
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
		"INSERT INTO application_components (id, tenant_id, name, description, created_at) VALUES ($1, $2, $3, $4, $5)",
		componentID, tenantA.Value(), "Read Only Test", "Test", time.Now(),
	)
	require.NoError(t, err)

	var name string
	err = ctx.tenantDB.WithReadOnlyTx(ctxA, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctxA,
			"SELECT name FROM application_components WHERE tenant_id = $1 AND id = $2",
			tenantA.Value(), componentID,
		).Scan(&name)
	})
	require.NoError(t, err)
	assert.Equal(t, "Read Only Test", name)

	err = ctx.tenantDB.WithReadOnlyTx(ctxA, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctxA,
			"UPDATE application_components SET name = $1 WHERE tenant_id = $2 AND id = $3",
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

	tenantA, err := sharedvo.NewTenantID("tenant-a")
	require.NoError(t, err)

	tenantB, err := sharedvo.NewTenantID("tenant-b")
	require.NoError(t, err)

	aggregateID := ctx.uniqueID("versioned-agg")

	ctxA := sharedctx.WithTenant(context.Background(), tenantA)
	ctxB := sharedctx.WithTenant(context.Background(), tenantB)

	eventData := map[string]interface{}{"test": "data"}
	eventJSON, err := json.Marshal(eventData)
	require.NoError(t, err)

	txA, err := ctx.tenantDB.BeginTxWithTenant(ctxA, nil)
	require.NoError(t, err)

	_, err = txA.ExecContext(ctxA,
		"INSERT INTO events (tenant_id, aggregate_id, event_type, event_data, version, occurred_at, actor_id, actor_email) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
		tenantA.Value(), aggregateID, "Event1", eventJSON, 1, time.Now(), "test-user-id", "test@example.com",
	)
	require.NoError(t, err)

	err = txA.Commit()
	require.NoError(t, err)

	txB, err := ctx.tenantDB.BeginTxWithTenant(ctxB, nil)
	require.NoError(t, err)

	_, err = txB.ExecContext(ctxB,
		"INSERT INTO events (tenant_id, aggregate_id, event_type, event_data, version, occurred_at, actor_id, actor_email) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
		tenantB.Value(), aggregateID, "Event1", eventJSON, 1, time.Now(), "test-user-id", "test@example.com",
	)
	require.NoError(t, err)

	err = txB.Commit()
	require.NoError(t, err)

	var versionA int
	err = ctx.tenantDB.WithReadOnlyTx(ctxA, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctxA,
			"SELECT version FROM events WHERE tenant_id = $1 AND aggregate_id = $2",
			tenantA.Value(), aggregateID,
		).Scan(&versionA)
	})
	require.NoError(t, err)
	assert.Equal(t, 1, versionA)

	var versionB int
	err = ctx.tenantDB.WithReadOnlyTx(ctxB, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctxB,
			"SELECT version FROM events WHERE tenant_id = $1 AND aggregate_id = $2",
			tenantB.Value(), aggregateID,
		).Scan(&versionB)
	})
	require.NoError(t, err)
	assert.Equal(t, 1, versionB)
}
