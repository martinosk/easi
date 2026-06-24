//go:build integration
// +build integration

package api

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	archEvents "easi/backend/internal/architecturemodeling/domain/events"
	"easi/backend/internal/capabilitymapping/application/projectors"
	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/shared/events"
	domain "easi/backend/internal/shared/eventsourcing"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type componentCacheTestContext struct {
	db                  *sql.DB
	testID              string
	createdComponentIDs []string
	componentCacheRM    *readmodels.ComponentCacheReadModel
	eventBus            events.EventBus
}

func setupComponentCacheTestDB(t *testing.T) (*componentCacheTestContext, func()) {
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

	testID := fmt.Sprintf("%s-%d", t.Name(), time.Now().UnixNano())

	tenantDB := database.NewTenantAwareDB(db)
	eventBus := events.NewInMemoryEventBus()

	componentCacheRM := readmodels.NewComponentCacheReadModel(tenantDB)

	componentCacheProjector := projectors.NewComponentCacheProjector(componentCacheRM)
	eventBus.Subscribe("ApplicationComponentCreated", componentCacheProjector)
	eventBus.Subscribe("ApplicationComponentUpdated", componentCacheProjector)
	eventBus.Subscribe("ApplicationComponentDeleted", componentCacheProjector)

	ctx := &componentCacheTestContext{
		db:                  db,
		testID:              testID,
		createdComponentIDs: make([]string, 0),
		componentCacheRM:    componentCacheRM,
		eventBus:            eventBus,
	}

	cleanup := func() {
		ctx.setTenantContext(t)
		for _, id := range ctx.createdComponentIDs {
			db.Exec("DELETE FROM capabilitymapping.capability_component_cache WHERE id = $1", id)
		}
		db.Close()
	}

	return ctx, cleanup
}

type cachedComponent struct {
	id   string
	name string
}

func newCachedComponent(name string) cachedComponent {
	return cachedComponent{id: uuid.New().String(), name: name}
}

func newComponentCacheTest(t *testing.T) *componentCacheTestContext {
	testCtx, cleanup := setupComponentCacheTestDB(t)
	t.Cleanup(cleanup)
	return testCtx
}

func (ctx *componentCacheTestContext) requireComponent(t *testing.T, want cachedComponent, msg string) {
	component, err := ctx.componentCacheRM.GetByID(tenantContext(), want.id)
	require.NoError(t, err)
	require.NotNil(t, component, msg)
	assert.Equal(t, want.id, component.ID)
	assert.Equal(t, want.name, component.Name)
}

func (ctx *componentCacheTestContext) requireNoComponent(t *testing.T, id, msg string) {
	component, err := ctx.componentCacheRM.GetByID(tenantContext(), id)
	require.NoError(t, err)
	assert.Nil(t, component, msg)
}

func (ctx *componentCacheTestContext) setTenantContext(t *testing.T) {
	_, err := ctx.db.Exec(fmt.Sprintf("SET app.current_tenant = '%s'", testTenantID()))
	require.NoError(t, err)
}

func (ctx *componentCacheTestContext) trackComponentID(id string) {
	ctx.createdComponentIDs = append(ctx.createdComponentIDs, id)
}

func (ctx *componentCacheTestContext) publishEvent(t *testing.T, event domain.DomainEvent) {
	err := ctx.eventBus.Publish(tenantContext(), []domain.DomainEvent{event})
	require.NoError(t, err)

	time.Sleep(50 * time.Millisecond)
}

func (ctx *componentCacheTestContext) publishComponentCreated(t *testing.T, c cachedComponent) {
	ctx.trackComponentID(c.id)
	ctx.publishEvent(t, archEvents.NewApplicationComponentCreated(c.id, c.name, "test description"))
}

func (ctx *componentCacheTestContext) publishComponentUpdated(t *testing.T, c cachedComponent) {
	ctx.publishEvent(t, archEvents.NewApplicationComponentUpdated(c.id, c.name, "updated description"))
}

func (ctx *componentCacheTestContext) publishComponentDeleted(t *testing.T, c cachedComponent) {
	ctx.publishEvent(t, archEvents.NewApplicationComponentDeleted(c.id, c.name))
}

func TestComponentCache_PopulatedByCreatedEvent_Integration(t *testing.T) {
	testCtx := newComponentCacheTest(t)

	component := newCachedComponent("Test Component Created")

	testCtx.publishComponentCreated(t, component)

	testCtx.requireComponent(t, component, "Component should be in cache after Created event")
}

func TestComponentCache_UpdatedByUpdateEvent_Integration(t *testing.T) {
	testCtx := newComponentCacheTest(t)

	original := newCachedComponent("Original Name")
	updated := cachedComponent{id: original.id, name: "Updated Name"}

	testCtx.publishComponentCreated(t, original)
	testCtx.requireComponent(t, original, "Component should be in cache after Created event")

	testCtx.publishComponentUpdated(t, updated)
	testCtx.requireComponent(t, updated, "Component should reflect updated name")
}

func TestComponentCache_RemovedByDeleteEvent_Integration(t *testing.T) {
	testCtx := newComponentCacheTest(t)

	component := newCachedComponent("Component To Delete")

	testCtx.publishComponentCreated(t, component)
	testCtx.requireComponent(t, component, "Component should exist before delete")

	testCtx.publishComponentDeleted(t, component)
	testCtx.requireNoComponent(t, component.id, "Component should be removed from cache after Delete event")
}

func TestComponentCache_MultipleComponents_Integration(t *testing.T) {
	testCtx := newComponentCacheTest(t)

	comp1 := newCachedComponent("Component 1")
	comp2 := newCachedComponent("Component 2")
	comp3 := newCachedComponent("Component 3")

	testCtx.publishComponentCreated(t, comp1)
	testCtx.publishComponentCreated(t, comp2)
	testCtx.publishComponentCreated(t, comp3)

	testCtx.requireComponent(t, comp1, "Component 1 should exist")
	testCtx.requireComponent(t, comp2, "Component 2 should exist")
	testCtx.requireComponent(t, comp3, "Component 3 should exist")

	testCtx.publishComponentDeleted(t, comp2)

	testCtx.requireComponent(t, comp1, "Component 1 should still exist")
	testCtx.requireNoComponent(t, comp2.id, "Component 2 should be deleted")
	testCtx.requireComponent(t, comp3, "Component 3 should still exist")
}

func TestComponentCache_NonExistentComponent_ReturnsNil_Integration(t *testing.T) {
	testCtx := newComponentCacheTest(t)

	nonExistentID := uuid.New().String()

	testCtx.requireNoComponent(t, nonExistentID, "Should return nil for non-existent component")
}

func TestComponentCache_UpdateNonExistent_CreatesEntry_Integration(t *testing.T) {
	testCtx := newComponentCacheTest(t)

	component := newCachedComponent("Created via Update")

	testCtx.trackComponentID(component.id)
	testCtx.publishComponentUpdated(t, component)

	testCtx.requireComponent(t, component, "Update event should create entry if not exists (upsert behavior)")
}
