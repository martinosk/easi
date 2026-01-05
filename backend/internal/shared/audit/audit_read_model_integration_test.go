//go:build integration
// +build integration

package audit

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testContext struct {
	db       *sql.DB
	tenantDB *database.TenantAwareDB
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
	tenantDB := database.NewTenantAwareDB(db)

	ctx := &testContext{
		db:       db,
		tenantDB: tenantDB,
		testID:   testID,
	}

	cleanup := func() {
		db.Exec(fmt.Sprintf("DELETE FROM events WHERE aggregate_id LIKE '%s%%'", testID))
		db.Close()
	}

	return ctx, cleanup
}

func (tc *testContext) uniqueID(suffix string) string {
	return fmt.Sprintf("%s-%s", tc.testID, suffix)
}

func TestAuditHistory_IncludesFitScoreEventsForComponent(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cleanup := setupTestDB(t)
	defer cleanup()

	tenantID, err := sharedvo.NewTenantID("test-tenant")
	require.NoError(t, err)

	tenantCtx := sharedctx.WithTenant(context.Background(), tenantID)

	componentID := ctx.uniqueID("component")
	fitScoreAggregateID := ctx.uniqueID("fitscore")
	pillarID := ctx.uniqueID("pillar")

	componentEventData := map[string]any{
		"id":          componentID,
		"name":        "Test Component",
		"description": "A test component",
	}
	componentEventJSON, err := json.Marshal(componentEventData)
	require.NoError(t, err)

	tx, err := ctx.tenantDB.BeginTxWithTenant(tenantCtx, nil)
	require.NoError(t, err)

	_, err = tx.ExecContext(tenantCtx,
		`INSERT INTO events (tenant_id, aggregate_id, event_type, event_data, version, occurred_at, actor_id, actor_email)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		tenantID.Value(), componentID, "ApplicationComponentCreated", componentEventJSON, 1, time.Now(), "user-1", "user@test.com",
	)
	require.NoError(t, err)

	fitScoreEventData := map[string]any{
		"id":          fitScoreAggregateID,
		"componentId": componentID,
		"pillarId":    pillarID,
		"pillarName":  "Digital Transformation",
		"score":       4,
		"rationale":   "Good fit",
		"scoredBy":    "architect@test.com",
	}
	fitScoreEventJSON, err := json.Marshal(fitScoreEventData)
	require.NoError(t, err)

	_, err = tx.ExecContext(tenantCtx,
		`INSERT INTO events (tenant_id, aggregate_id, event_type, event_data, version, occurred_at, actor_id, actor_email)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		tenantID.Value(), fitScoreAggregateID, "ApplicationFitScoreSet", fitScoreEventJSON, 1, time.Now().Add(time.Second), "user-2", "architect@test.com",
	)
	require.NoError(t, err)

	fitScoreUpdateEventData := map[string]any{
		"id":           fitScoreAggregateID,
		"componentId":  componentID,
		"score":        5,
		"rationale":    "Excellent fit after review",
		"oldScore":     4,
		"oldRationale": "Good fit",
		"updatedBy":    "architect@test.com",
	}
	fitScoreUpdateEventJSON, err := json.Marshal(fitScoreUpdateEventData)
	require.NoError(t, err)

	_, err = tx.ExecContext(tenantCtx,
		`INSERT INTO events (tenant_id, aggregate_id, event_type, event_data, version, occurred_at, actor_id, actor_email)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		tenantID.Value(), fitScoreAggregateID, "ApplicationFitScoreUpdated", fitScoreUpdateEventJSON, 2, time.Now().Add(2*time.Second), "user-2", "architect@test.com",
	)
	require.NoError(t, err)

	err = tx.Commit()
	require.NoError(t, err)

	readModel := NewAuditHistoryReadModel(ctx.tenantDB)

	entries, hasMore, _, err := readModel.GetHistoryByAggregateID(tenantCtx, componentID, 50, "")
	require.NoError(t, err)
	assert.False(t, hasMore)

	assert.Len(t, entries, 3, "Expected 3 events: component created + fit score set + fit score updated")

	eventTypes := make([]string, len(entries))
	for i, entry := range entries {
		eventTypes[i] = entry.EventType
	}

	assert.Contains(t, eventTypes, "ApplicationComponentCreated", "Should include component created event")
	assert.Contains(t, eventTypes, "ApplicationFitScoreSet", "Should include fit score set event")
	assert.Contains(t, eventTypes, "ApplicationFitScoreUpdated", "Should include fit score updated event")

	for _, entry := range entries {
		if entry.EventType == "ApplicationFitScoreSet" {
			assert.Equal(t, fitScoreAggregateID, entry.AggregateID)
			assert.Equal(t, componentID, entry.EventData["componentId"])
			assert.Equal(t, float64(4), entry.EventData["score"])
		}
		if entry.EventType == "ApplicationFitScoreUpdated" {
			assert.Equal(t, fitScoreAggregateID, entry.AggregateID)
			assert.Equal(t, componentID, entry.EventData["componentId"])
			assert.Equal(t, float64(5), entry.EventData["score"])
		}
	}
}

func TestAuditHistory_DoesNotIncludeUnrelatedFitScoreEvents(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cleanup := setupTestDB(t)
	defer cleanup()

	tenantID, err := sharedvo.NewTenantID("test-tenant")
	require.NoError(t, err)

	tenantCtx := sharedctx.WithTenant(context.Background(), tenantID)

	componentA := ctx.uniqueID("component-a")
	componentB := ctx.uniqueID("component-b")
	fitScoreForA := ctx.uniqueID("fitscore-a")
	fitScoreForB := ctx.uniqueID("fitscore-b")

	tx, err := ctx.tenantDB.BeginTxWithTenant(tenantCtx, nil)
	require.NoError(t, err)

	componentAEventData, _ := json.Marshal(map[string]any{"id": componentA, "name": "Component A"})
	_, err = tx.ExecContext(tenantCtx,
		`INSERT INTO events (tenant_id, aggregate_id, event_type, event_data, version, occurred_at, actor_id, actor_email)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		tenantID.Value(), componentA, "ApplicationComponentCreated", componentAEventData, 1, time.Now(), "user-1", "user@test.com",
	)
	require.NoError(t, err)

	componentBEventData, _ := json.Marshal(map[string]any{"id": componentB, "name": "Component B"})
	_, err = tx.ExecContext(tenantCtx,
		`INSERT INTO events (tenant_id, aggregate_id, event_type, event_data, version, occurred_at, actor_id, actor_email)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		tenantID.Value(), componentB, "ApplicationComponentCreated", componentBEventData, 1, time.Now(), "user-1", "user@test.com",
	)
	require.NoError(t, err)

	fitScoreAEventData, _ := json.Marshal(map[string]any{
		"id": fitScoreForA, "componentId": componentA, "score": 4,
	})
	_, err = tx.ExecContext(tenantCtx,
		`INSERT INTO events (tenant_id, aggregate_id, event_type, event_data, version, occurred_at, actor_id, actor_email)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		tenantID.Value(), fitScoreForA, "ApplicationFitScoreSet", fitScoreAEventData, 1, time.Now(), "user-2", "architect@test.com",
	)
	require.NoError(t, err)

	fitScoreBEventData, _ := json.Marshal(map[string]any{
		"id": fitScoreForB, "componentId": componentB, "score": 3,
	})
	_, err = tx.ExecContext(tenantCtx,
		`INSERT INTO events (tenant_id, aggregate_id, event_type, event_data, version, occurred_at, actor_id, actor_email)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		tenantID.Value(), fitScoreForB, "ApplicationFitScoreSet", fitScoreBEventData, 1, time.Now(), "user-2", "architect@test.com",
	)
	require.NoError(t, err)

	err = tx.Commit()
	require.NoError(t, err)

	readModel := NewAuditHistoryReadModel(ctx.tenantDB)

	entriesA, _, _, err := readModel.GetHistoryByAggregateID(tenantCtx, componentA, 50, "")
	require.NoError(t, err)

	assert.Len(t, entriesA, 2, "Component A should have 2 events: created + its fit score")

	for _, entry := range entriesA {
		if entry.EventType == "ApplicationFitScoreSet" {
			assert.Equal(t, componentA, entry.EventData["componentId"], "Should only include fit score for component A")
		}
	}

	entriesB, _, _, err := readModel.GetHistoryByAggregateID(tenantCtx, componentB, 50, "")
	require.NoError(t, err)

	assert.Len(t, entriesB, 2, "Component B should have 2 events: created + its fit score")

	for _, entry := range entriesB {
		if entry.EventType == "ApplicationFitScoreSet" {
			assert.Equal(t, componentB, entry.EventData["componentId"], "Should only include fit score for component B")
		}
	}
}
