//go:build integration
// +build integration

package readmodels

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
		db.Exec(fmt.Sprintf("DELETE FROM infrastructure.events WHERE aggregate_id LIKE '%s%%'", testID))
		db.Close()
	}

	return ctx, cleanup
}

func (tc *testContext) uniqueID(suffix string) string {
	return fmt.Sprintf("%s-%s", tc.testID, suffix)
}

type eventRow struct {
	aggregateID string
	eventType   string
	data        map[string]any
	version     int
	occurredAt  time.Time
	actorID     string
	actorEmail  string
}

type eventInserter struct {
	t        *testing.T
	tx       *sql.Tx
	ctx      context.Context
	tenantID sharedvo.TenantID
}

func (ei eventInserter) insert(row eventRow) {
	ei.t.Helper()

	dataJSON, err := json.Marshal(row.data)
	require.NoError(ei.t, err)

	_, err = ei.tx.ExecContext(ei.ctx,
		`INSERT INTO infrastructure.events (tenant_id, aggregate_id, event_type, event_data, version, occurred_at, actor_id, actor_email)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		ei.tenantID.Value(), row.aggregateID, row.eventType, dataJSON, row.version, row.occurredAt, row.actorID, row.actorEmail,
	)
	require.NoError(ei.t, err)
}

type fitScoreExpectation struct {
	eventType   string
	aggregateID string
	componentID string
	score       float64
}

func assertFitScoreEntry(t *testing.T, entries []AuditEntry, want fitScoreExpectation) {
	t.Helper()
	for _, entry := range entries {
		if entry.EventType != want.eventType {
			continue
		}
		assert.Equal(t, want.aggregateID, entry.AggregateID)
		assert.Equal(t, want.componentID, entry.EventData["componentId"])
		assert.Equal(t, want.score, entry.EventData["score"])
	}
}

func assertFitScoreComponent(t *testing.T, entries []AuditEntry, componentID, message string) {
	t.Helper()
	for _, entry := range entries {
		if entry.EventType == "ApplicationFitScoreSet" {
			assert.Equal(t, componentID, entry.EventData["componentId"], message)
		}
	}
}

type seedContext struct {
	t        *testing.T
	tc       *testContext
	ctx      context.Context
	tenantID sharedvo.TenantID
}

func newSeedContext(t *testing.T) (seedContext, func()) {
	t.Helper()

	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tc, cleanup := setupTestDB(t)

	tenantID, err := sharedvo.NewTenantID("test-tenant")
	require.NoError(t, err)

	tenantCtx := sharedctx.WithTenant(context.Background(), tenantID)

	return seedContext{t: t, tc: tc, ctx: tenantCtx, tenantID: tenantID}, cleanup
}

func (s seedContext) withTransaction(seed func(eventInserter)) {
	s.t.Helper()

	tx, err := s.tc.tenantDB.BeginTxWithTenant(s.ctx, nil)
	require.NoError(s.t, err)

	seed(eventInserter{t: s.t, tx: tx, ctx: s.ctx, tenantID: s.tenantID})

	require.NoError(s.t, tx.Commit())
}

type componentFitScoreIDs struct {
	componentID         string
	fitScoreAggregateID string
	pillarID            string
}

func (s seedContext) seedComponentFitScoreHistory(ids componentFitScoreIDs) {
	s.t.Helper()

	now := time.Now()
	s.withTransaction(func(inserter eventInserter) {
		inserter.insert(eventRow{
			aggregateID: ids.componentID,
			eventType:   "ApplicationComponentCreated",
			data: map[string]any{
				"id":          ids.componentID,
				"name":        "Test Component",
				"description": "A test component",
			},
			version:    1,
			occurredAt: now,
			actorID:    "user-1",
			actorEmail: "user@test.com",
		})

		inserter.insert(eventRow{
			aggregateID: ids.fitScoreAggregateID,
			eventType:   "ApplicationFitScoreSet",
			data: map[string]any{
				"id":          ids.fitScoreAggregateID,
				"componentId": ids.componentID,
				"pillarId":    ids.pillarID,
				"pillarName":  "Digital Transformation",
				"score":       4,
				"rationale":   "Good fit",
				"scoredBy":    "architect@test.com",
			},
			version:    1,
			occurredAt: now.Add(time.Second),
			actorID:    "user-2",
			actorEmail: "architect@test.com",
		})

		inserter.insert(eventRow{
			aggregateID: ids.fitScoreAggregateID,
			eventType:   "ApplicationFitScoreUpdated",
			data: map[string]any{
				"id":           ids.fitScoreAggregateID,
				"componentId":  ids.componentID,
				"score":        5,
				"rationale":    "Excellent fit after review",
				"oldScore":     4,
				"oldRationale": "Good fit",
				"updatedBy":    "architect@test.com",
			},
			version:    2,
			occurredAt: now.Add(2 * time.Second),
			actorID:    "user-2",
			actorEmail: "architect@test.com",
		})
	})
}

func TestAuditHistory_IncludesFitScoreEventsForComponent(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	seed, cleanup := newSeedContext(t)
	defer cleanup()

	ids := componentFitScoreIDs{
		componentID:         seed.tc.uniqueID("component"),
		fitScoreAggregateID: seed.tc.uniqueID("fitscore"),
		pillarID:            seed.tc.uniqueID("pillar"),
	}

	seed.seedComponentFitScoreHistory(ids)

	componentID := ids.componentID
	fitScoreAggregateID := ids.fitScoreAggregateID

	readModel := NewAuditHistoryReadModel(seed.tc.tenantDB)

	entries, hasMore, _, err := readModel.GetHistoryByAggregateID(seed.ctx, componentID, 50, "")
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

	assertFitScoreEntry(t, entries, fitScoreExpectation{
		eventType: "ApplicationFitScoreSet", aggregateID: fitScoreAggregateID, componentID: componentID, score: float64(4),
	})
	assertFitScoreEntry(t, entries, fitScoreExpectation{
		eventType: "ApplicationFitScoreUpdated", aggregateID: fitScoreAggregateID, componentID: componentID, score: float64(5),
	})
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

	now := time.Now()
	inserter := eventInserter{t: t, tx: tx, ctx: tenantCtx, tenantID: tenantID}

	inserter.insert(eventRow{
		aggregateID: componentA,
		eventType:   "ApplicationComponentCreated",
		data:        map[string]any{"id": componentA, "name": "Component A"},
		version:     1,
		occurredAt:  now,
		actorID:     "user-1",
		actorEmail:  "user@test.com",
	})

	inserter.insert(eventRow{
		aggregateID: componentB,
		eventType:   "ApplicationComponentCreated",
		data:        map[string]any{"id": componentB, "name": "Component B"},
		version:     1,
		occurredAt:  now,
		actorID:     "user-1",
		actorEmail:  "user@test.com",
	})

	inserter.insert(eventRow{
		aggregateID: fitScoreForA,
		eventType:   "ApplicationFitScoreSet",
		data:        map[string]any{"id": fitScoreForA, "componentId": componentA, "score": 4},
		version:     1,
		occurredAt:  now,
		actorID:     "user-2",
		actorEmail:  "architect@test.com",
	})

	inserter.insert(eventRow{
		aggregateID: fitScoreForB,
		eventType:   "ApplicationFitScoreSet",
		data:        map[string]any{"id": fitScoreForB, "componentId": componentB, "score": 3},
		version:     1,
		occurredAt:  now,
		actorID:     "user-2",
		actorEmail:  "architect@test.com",
	})

	err = tx.Commit()
	require.NoError(t, err)

	readModel := NewAuditHistoryReadModel(ctx.tenantDB)

	entriesA, _, _, err := readModel.GetHistoryByAggregateID(tenantCtx, componentA, 50, "")
	require.NoError(t, err)

	assert.Len(t, entriesA, 2, "Component A should have 2 events: created + its fit score")
	assertFitScoreComponent(t, entriesA, componentA, "Should only include fit score for component A")

	entriesB, _, _, err := readModel.GetHistoryByAggregateID(tenantCtx, componentB, 50, "")
	require.NoError(t, err)

	assert.Len(t, entriesB, 2, "Component B should have 2 events: created + its fit score")
	assertFitScoreComponent(t, entriesB, componentB, "Should only include fit score for component B")
}
