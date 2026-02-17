//go:build integration
// +build integration

package projectors

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/domain/events"
	"easi/backend/internal/infrastructure/database"
	sharedcontext "easi/backend/internal/shared/context"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type realizationProjectorIntegrationContext struct {
	db *sql.DB
}

func setupRealizationProjectorIntegrationDB(t *testing.T) (*realizationProjectorIntegrationContext, func()) {
	dbHost := getEnvOrDefault("INTEGRATION_TEST_DB_HOST", "localhost")
	dbPort := getEnvOrDefault("INTEGRATION_TEST_DB_PORT", "5432")
	dbUser := getEnvOrDefault("INTEGRATION_TEST_DB_USER", "easi_app")
	dbPassword := getEnvOrDefault("INTEGRATION_TEST_DB_PASSWORD", "localdev")
	dbName := getEnvOrDefault("INTEGRATION_TEST_DB_NAME", "easi")
	dbSSLMode := getEnvOrDefault("INTEGRATION_TEST_DB_SSLMODE", "disable")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbPassword, dbName, dbSSLMode)
	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err)
	require.NoError(t, db.Ping())

	ctx := &realizationProjectorIntegrationContext{db: db}
	cleanup := func() {
		db.Close()
	}
	return ctx, cleanup
}

func (ctx *realizationProjectorIntegrationContext) setTenantContext(t *testing.T) {
	_, err := ctx.db.Exec("SET app.current_tenant = 'default'")
	require.NoError(t, err)
}

func tenantContext() context.Context {
	ctx := sharedcontext.WithTenant(context.Background(), sharedvo.DefaultTenantID())
	ctx = sharedcontext.WithActor(ctx, sharedcontext.Actor{ID: "test-user-id", Email: "test@example.com"})
	return ctx
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func TestRealizationProjectorIntegration_AppliesInheritedEvent(t *testing.T) {
	testCtx, cleanup := setupRealizationProjectorIntegrationDB(t)
	defer cleanup()
	testCtx.setTenantContext(t)

	tenantDB := database.NewTenantAwareDB(testCtx.db)
	realizationRM := readmodels.NewRealizationReadModel(tenantDB)
	componentCacheRM := readmodels.NewComponentCacheReadModel(tenantDB)
	projector := NewRealizationProjector(realizationRM, componentCacheRM)

	event := events.CapabilityRealizationsInherited{
		CapabilityID: "cap-source",
		InheritedRealizations: []events.InheritedRealization{
			{
				CapabilityID:         "cap-parent-a",
				ComponentID:          "comp-a",
				ComponentName:        "Component A",
				RealizationLevel:     "Full",
				Origin:               "Inherited",
				SourceRealizationID:  "real-source-1",
				SourceCapabilityID:   "cap-source",
				SourceCapabilityName: "Source",
				LinkedAt:             time.Now().UTC(),
			},
			{
				CapabilityID:         "cap-parent-b",
				ComponentID:          "comp-b",
				ComponentName:        "Component B",
				RealizationLevel:     "Full",
				Origin:               "Inherited",
				SourceRealizationID:  "real-source-1",
				SourceCapabilityID:   "cap-source",
				SourceCapabilityName: "Source",
				LinkedAt:             time.Now().UTC(),
			},
		},
	}

	eventData, err := json.Marshal(event)
	require.NoError(t, err)

	err = projector.ProjectEvent(tenantContext(), "CapabilityRealizationsInherited", eventData)
	require.NoError(t, err)

	var inheritedCount int
	err = testCtx.db.QueryRow(
		"SELECT COUNT(*) FROM capabilitymapping.capability_realizations WHERE tenant_id = $1 AND origin = 'Inherited' AND source_realization_id = $2",
		"default", "real-source-1",
	).Scan(&inheritedCount)
	require.NoError(t, err)
	assert.Equal(t, 2, inheritedCount)

	_, err = testCtx.db.Exec(
		"DELETE FROM capabilitymapping.capability_realizations WHERE tenant_id = $1 AND source_realization_id = $2",
		"default",
		"real-source-1",
	)
	require.NoError(t, err)
}

func TestRealizationProjectorIntegration_AppliesUninheritedEvent(t *testing.T) {
	testCtx, cleanup := setupRealizationProjectorIntegrationDB(t)
	defer cleanup()
	testCtx.setTenantContext(t)

	tenantDB := database.NewTenantAwareDB(testCtx.db)
	realizationRM := readmodels.NewRealizationReadModel(tenantDB)
	componentCacheRM := readmodels.NewComponentCacheReadModel(tenantDB)
	projector := NewRealizationProjector(realizationRM, componentCacheRM)

	keepID := uuid.New().String()
	removeAID := uuid.New().String()
	removeBID := uuid.New().String()

	_, err := testCtx.db.Exec(
		`INSERT INTO capabilitymapping.capability_realizations
			(id, tenant_id, capability_id, component_id, component_name, realization_level, notes, origin, source_realization_id, source_capability_id, source_capability_name, linked_at)
		VALUES
			($1, $2, 'cap-old-a', 'comp-a', 'Component A', 'Full', '', 'Inherited', 'real-source-remove', 'cap-source', 'Source', NOW()),
			($3, $2, 'cap-old-b', 'comp-b', 'Component B', 'Full', '', 'Inherited', 'real-source-remove', 'cap-source', 'Source', NOW()),
			($4, $2, 'cap-keep', 'comp-c', 'Component C', 'Full', '', 'Inherited', 'real-source-keep', 'cap-source', 'Source', NOW())`,
		removeAID, "default", removeBID, keepID,
	)
	require.NoError(t, err)

	event := events.CapabilityRealizationsUninherited{
		CapabilityID: "cap-source",
		Removals: []events.RealizationInheritanceRemoval{
			{
				SourceRealizationID: "real-source-remove",
				CapabilityIDs:       []string{"cap-old-a", "cap-old-b"},
			},
		},
	}

	eventData, err := json.Marshal(event)
	require.NoError(t, err)

	err = projector.ProjectEvent(tenantContext(), "CapabilityRealizationsUninherited", eventData)
	require.NoError(t, err)

	var removedCount int
	err = testCtx.db.QueryRow(
		"SELECT COUNT(*) FROM capabilitymapping.capability_realizations WHERE tenant_id = $1 AND source_realization_id = $2",
		"default", "real-source-remove",
	).Scan(&removedCount)
	require.NoError(t, err)
	assert.Equal(t, 0, removedCount)

	var keptCount int
	err = testCtx.db.QueryRow(
		"SELECT COUNT(*) FROM capabilitymapping.capability_realizations WHERE tenant_id = $1 AND source_realization_id = $2",
		"default", "real-source-keep",
	).Scan(&keptCount)
	require.NoError(t, err)
	assert.Equal(t, 1, keptCount)

	_, err = testCtx.db.Exec(
		"DELETE FROM capabilitymapping.capability_realizations WHERE tenant_id = $1 AND id IN ($2, $3, $4)",
		"default", removeAID, removeBID, keepID,
	)
	require.NoError(t, err)
}
