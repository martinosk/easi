//go:build integration
// +build integration

package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	sharedctx "easi/backend/internal/shared/context"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIndexUsage_EventsTenantQuery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cleanup := setupTestDB(t)
	defer cleanup()

	tenant, err := sharedvo.NewTenantID("perf-tenant")
	require.NoError(t, err)

	tenantCtx := sharedctx.WithTenant(context.Background(), tenant)

	var explainOutput string
	err = ctx.tenantDB.WithReadOnlyTx(tenantCtx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(tenantCtx,
			"EXPLAIN SELECT * FROM events WHERE tenant_id = $1 AND aggregate_id = $2 ORDER BY version ASC",
			tenant.Value(), "test-aggregate-id",
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		var lines []string
		for rows.Next() {
			var line string
			if err := rows.Scan(&line); err != nil {
				return err
			}
			lines = append(lines, line)
		}
		explainOutput = strings.Join(lines, "\n")
		return rows.Err()
	})
	require.NoError(t, err)

	usesIndex := strings.Contains(explainOutput, "Index") || strings.Contains(explainOutput, "idx_events_tenant")
	assert.True(t, usesIndex, "Query should use tenant index. Got: %s", explainOutput)
}

func TestIndexUsage_ApplicationComponentsTenantQuery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cleanup := setupTestDB(t)
	defer cleanup()

	tenant, err := sharedvo.NewTenantID("perf-tenant")
	require.NoError(t, err)

	tenantCtx := sharedctx.WithTenant(context.Background(), tenant)

	var explainOutput string
	err = ctx.tenantDB.WithReadOnlyTx(tenantCtx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(tenantCtx,
			"EXPLAIN SELECT * FROM application_components WHERE tenant_id = $1 ORDER BY created_at DESC",
			tenant.Value(),
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		var lines []string
		for rows.Next() {
			var line string
			if err := rows.Scan(&line); err != nil {
				return err
			}
			lines = append(lines, line)
		}
		explainOutput = strings.Join(lines, "\n")
		return rows.Err()
	})
	require.NoError(t, err)

	usesIndex := strings.Contains(explainOutput, "Index") || strings.Contains(explainOutput, "idx_application_components")
	assert.True(t, usesIndex, "Query should use tenant index. Got: %s", explainOutput)
}

func TestIndexUsage_CapabilitiesTenantQuery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cleanup := setupTestDB(t)
	defer cleanup()

	tenant, err := sharedvo.NewTenantID("perf-tenant")
	require.NoError(t, err)

	tenantCtx := sharedctx.WithTenant(context.Background(), tenant)

	var explainOutput string
	err = ctx.tenantDB.WithReadOnlyTx(tenantCtx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(tenantCtx,
			"EXPLAIN SELECT * FROM capabilities WHERE tenant_id = $1 ORDER BY level, name",
			tenant.Value(),
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		var lines []string
		for rows.Next() {
			var line string
			if err := rows.Scan(&line); err != nil {
				return err
			}
			lines = append(lines, line)
		}
		explainOutput = strings.Join(lines, "\n")
		return rows.Err()
	})
	require.NoError(t, err)

	usesIndex := strings.Contains(explainOutput, "Index") || strings.Contains(explainOutput, "idx_capabilities")
	assert.True(t, usesIndex, "Query should use tenant index. Got: %s", explainOutput)
}

func TestRLSPerformance_WithTenantContext(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cleanup := setupTestDB(t)
	defer cleanup()

	tenant, err := sharedvo.NewTenantID("perf-tenant")
	require.NoError(t, err)

	tenantCtx := sharedctx.WithTenant(context.Background(), tenant)

	const numOperations = 100
	start := time.Now()

	for i := 0; i < numOperations; i++ {
		err = ctx.tenantDB.WithReadOnlyTx(tenantCtx, func(tx *sql.Tx) error {
			var count int
			return tx.QueryRowContext(tenantCtx,
				"SELECT COUNT(*) FROM events WHERE tenant_id = $1",
				tenant.Value(),
			).Scan(&count)
		})
		require.NoError(t, err)
	}

	elapsed := time.Since(start)
	avgLatency := elapsed / numOperations

	t.Logf("RLS Query Performance: %d operations in %v (avg: %v per operation)", numOperations, elapsed, avgLatency)
	assert.Less(t, avgLatency, 50*time.Millisecond, "Average query latency should be under 50ms")
}

func TestRLSPerformance_TenantContextSwitch(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cleanup := setupTestDB(t)
	defer cleanup()

	tenants := make([]sharedvo.TenantID, 10)
	for i := 0; i < 10; i++ {
		tenant, err := sharedvo.NewTenantID(fmt.Sprintf("perf-tenant-%d", i))
		require.NoError(t, err)
		tenants[i] = tenant
	}

	const numSwitches = 50
	start := time.Now()

	for i := 0; i < numSwitches; i++ {
		tenant := tenants[i%len(tenants)]
		tenantCtx := sharedctx.WithTenant(context.Background(), tenant)

		err := ctx.tenantDB.WithReadOnlyTx(tenantCtx, func(tx *sql.Tx) error {
			var count int
			return tx.QueryRowContext(tenantCtx,
				"SELECT COUNT(*) FROM events WHERE tenant_id = $1",
				tenant.Value(),
			).Scan(&count)
		})
		require.NoError(t, err)
	}

	elapsed := time.Since(start)
	avgSwitchTime := elapsed / numSwitches

	t.Logf("Tenant Context Switch Performance: %d switches in %v (avg: %v per switch)", numSwitches, elapsed, avgSwitchTime)
	assert.Less(t, avgSwitchTime, 100*time.Millisecond, "Average context switch latency should be under 100ms")
}

func TestRLSPerformance_BulkInsert(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cleanup := setupTestDB(t)
	defer cleanup()

	tenant, err := sharedvo.NewTenantID("bulk-perf-tenant")
	require.NoError(t, err)

	tenantCtx := sharedctx.WithTenant(context.Background(), tenant)
	aggregateID := ctx.uniqueID("bulk-agg")

	eventData := map[string]interface{}{"test": "data"}
	eventJSON, err := json.Marshal(eventData)
	require.NoError(t, err)

	const numEvents = 50
	start := time.Now()

	tx, err := ctx.tenantDB.BeginTxWithTenant(tenantCtx, nil)
	require.NoError(t, err)

	for i := 0; i < numEvents; i++ {
		_, err = tx.ExecContext(tenantCtx,
			"INSERT INTO events (tenant_id, aggregate_id, event_type, event_data, version, occurred_at) VALUES ($1, $2, $3, $4, $5, $6)",
			tenant.Value(), fmt.Sprintf("%s-%d", aggregateID, i), "TestEvent", eventJSON, 1, time.Now(),
		)
		require.NoError(t, err)
	}

	err = tx.Commit()
	require.NoError(t, err)

	elapsed := time.Since(start)
	avgInsertTime := elapsed / numEvents

	t.Logf("RLS Bulk Insert Performance: %d inserts in %v (avg: %v per insert)", numEvents, elapsed, avgInsertTime)
	assert.Less(t, avgInsertTime, 10*time.Millisecond, "Average insert latency should be under 10ms")

	ctx.db.Exec(fmt.Sprintf("DELETE FROM events WHERE aggregate_id LIKE '%s%%'", aggregateID))
}

func BenchmarkRLSQuery(b *testing.B) {
	dbHost := getEnv("INTEGRATION_TEST_DB_HOST", "localhost")
	dbPort := getEnv("INTEGRATION_TEST_DB_PORT", "5432")
	dbUser := getEnv("INTEGRATION_TEST_DB_USER", "easi_app")
	dbPassword := getEnv("INTEGRATION_TEST_DB_PASSWORD", "localdev")
	dbName := getEnv("INTEGRATION_TEST_DB_NAME", "easi")
	dbSSLMode := getEnv("INTEGRATION_TEST_DB_SSLMODE", "disable")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbPassword, dbName, dbSSLMode)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		b.Skip("Database not available")
		return
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		b.Skip("Database not available")
		return
	}

	tenantDB := NewTenantAwareDB(db)
	tenant, _ := sharedvo.NewTenantID("bench-tenant")
	tenantCtx := sharedctx.WithTenant(context.Background(), tenant)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := tenantDB.WithReadOnlyTx(tenantCtx, func(tx *sql.Tx) error {
			var count int
			return tx.QueryRowContext(tenantCtx,
				"SELECT COUNT(*) FROM events WHERE tenant_id = $1",
				tenant.Value(),
			).Scan(&count)
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRLSInsert(b *testing.B) {
	dbHost := getEnv("INTEGRATION_TEST_DB_HOST", "localhost")
	dbPort := getEnv("INTEGRATION_TEST_DB_PORT", "5432")
	dbUser := getEnv("INTEGRATION_TEST_DB_USER", "easi_app")
	dbPassword := getEnv("INTEGRATION_TEST_DB_PASSWORD", "localdev")
	dbName := getEnv("INTEGRATION_TEST_DB_NAME", "easi")
	dbSSLMode := getEnv("INTEGRATION_TEST_DB_SSLMODE", "disable")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbPassword, dbName, dbSSLMode)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		b.Skip("Database not available")
		return
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		b.Skip("Database not available")
		return
	}

	tenantDB := NewTenantAwareDB(db)
	tenant, _ := sharedvo.NewTenantID("bench-tenant")
	tenantCtx := sharedctx.WithTenant(context.Background(), tenant)
	testID := fmt.Sprintf("bench-%d", time.Now().UnixNano())

	eventData := map[string]interface{}{"test": "data"}
	eventJSON, _ := json.Marshal(eventData)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		tx, err := tenantDB.BeginTxWithTenant(tenantCtx, nil)
		if err != nil {
			b.Fatal(err)
		}

		_, err = tx.ExecContext(tenantCtx,
			"INSERT INTO events (tenant_id, aggregate_id, event_type, event_data, version, occurred_at) VALUES ($1, $2, $3, $4, $5, $6)",
			tenant.Value(), fmt.Sprintf("%s-%d", testID, i), "BenchEvent", eventJSON, 1, time.Now(),
		)
		if err != nil {
			tx.Rollback()
			b.Fatal(err)
		}

		if err := tx.Commit(); err != nil {
			b.Fatal(err)
		}
	}

	b.StopTimer()
	db.Exec(fmt.Sprintf("DELETE FROM events WHERE aggregate_id LIKE '%s%%'", testID))
}
