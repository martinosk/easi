//go:build integration

package fixtures

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/infrastructure/eventstore"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/shared/events"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

type TestContext struct {
	T          *testing.T
	DB         *sql.DB
	TenantDB   *database.TenantAwareDB
	EventStore *eventstore.PostgresEventStore
	CommandBus *cqrs.InMemoryCommandBus
	EventBus   *events.InMemoryEventBus
	Ctx        context.Context
	TenantID   sharedvo.TenantID
	cleanupIDs []string
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func NewTestContext(t *testing.T) *TestContext {
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

	tenantDB := database.NewTenantAwareDB(db)
	eventStore := eventstore.NewPostgresEventStore(tenantDB)
	commandBus := cqrs.NewInMemoryCommandBus()
	eventBus := events.NewInMemoryEventBus()
	eventStore.SetEventBus(eventBus)

	tenantID := sharedvo.DefaultTenantID()
	ctx := sharedctx.WithTenant(context.Background(), tenantID)

	_, err = db.Exec(fmt.Sprintf("SET app.current_tenant = '%s'", tenantID.Value()))
	require.NoError(t, err)

	tc := &TestContext{
		T:          t,
		DB:         db,
		TenantDB:   tenantDB,
		EventStore: eventStore,
		CommandBus: commandBus,
		EventBus:   eventBus,
		Ctx:        ctx,
		TenantID:   tenantID,
		cleanupIDs: make([]string, 0),
	}

	t.Cleanup(func() {
		tc.cleanup()
		db.Close()
	})

	return tc
}

func (tc *TestContext) TrackID(id string) {
	tc.cleanupIDs = append(tc.cleanupIDs, id)
}

func (tc *TestContext) cleanup() {
	tc.setTenantContext()
	for _, id := range tc.cleanupIDs {
		tc.DB.Exec("DELETE FROM capabilitymapping.domain_capability_metadata WHERE capability_id = $1", id)
		tc.DB.Exec("DELETE FROM capabilitymapping.domain_capability_assignments WHERE capability_id = $1 OR business_domain_id = $1", id)
		tc.DB.Exec("DELETE FROM capabilitymapping.strategy_importance WHERE capability_id = $1 OR business_domain_id = $1", id)
		tc.DB.Exec("DELETE FROM capabilitymapping.effective_capability_importance WHERE capability_id = $1 OR business_domain_id = $1", id)
		tc.DB.Exec("DELETE FROM capabilitymapping.capability_realizations WHERE capability_id = $1 OR component_id = $1", id)
		tc.DB.Exec("DELETE FROM capabilitymapping.application_fit_scores WHERE component_id = $1", id)
		tc.DB.Exec("DELETE FROM enterprisearchitecture.enterprise_capability_links WHERE domain_capability_id = $1 OR enterprise_capability_id = $1", id)
		tc.DB.Exec("DELETE FROM enterprisearchitecture.capability_link_blocking WHERE domain_capability_id = $1 OR blocked_by_capability_id = $1", id)
		tc.DB.Exec("DELETE FROM capabilitymapping.capabilities WHERE id = $1", id)
		tc.DB.Exec("DELETE FROM capabilitymapping.business_domains WHERE id = $1", id)
		tc.DB.Exec("DELETE FROM architecturemodeling.application_components WHERE id = $1", id)
		tc.DB.Exec("DELETE FROM enterprisearchitecture.enterprise_capabilities WHERE id = $1", id)
		tc.DB.Exec("DELETE FROM infrastructure.events WHERE aggregate_id = $1", id)
	}
}

func (tc *TestContext) setTenantContext() {
	_, err := tc.DB.Exec(fmt.Sprintf("SET app.current_tenant = '%s'", tc.TenantID.Value()))
	require.NoError(tc.T, err)
}

func (tc *TestContext) Dispatch(cmd cqrs.Command) (cqrs.CommandResult, error) {
	return tc.CommandBus.Dispatch(tc.Ctx, cmd)
}

func (tc *TestContext) MustDispatch(cmd cqrs.Command) cqrs.CommandResult {
	result, err := tc.Dispatch(cmd)
	require.NoError(tc.T, err)
	return result
}
