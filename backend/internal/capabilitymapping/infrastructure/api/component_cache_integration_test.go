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
	db                   *sql.DB
	testID               string
	createdComponentIDs  []string
	componentCacheRM     *readmodels.ComponentCacheReadModel
	eventBus             events.EventBus
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
			db.Exec("DELETE FROM capability_component_cache WHERE id = $1", id)
		}
		db.Close()
	}

	return ctx, cleanup
}

func (ctx *componentCacheTestContext) setTenantContext(t *testing.T) {
	_, err := ctx.db.Exec(fmt.Sprintf("SET app.current_tenant = '%s'", testTenantID()))
	require.NoError(t, err)
}

func (ctx *componentCacheTestContext) trackComponentID(id string) {
	ctx.createdComponentIDs = append(ctx.createdComponentIDs, id)
}

func (ctx *componentCacheTestContext) publishComponentCreated(t *testing.T, id, name string) {
	event := archEvents.NewApplicationComponentCreated(id, name, "test description")
	err := ctx.eventBus.Publish(tenantContext(), []domain.DomainEvent{event})
	require.NoError(t, err)
	ctx.trackComponentID(id)

	time.Sleep(50 * time.Millisecond)
}

func (ctx *componentCacheTestContext) publishComponentUpdated(t *testing.T, id, name string) {
	event := archEvents.NewApplicationComponentUpdated(id, name, "updated description")
	err := ctx.eventBus.Publish(tenantContext(), []domain.DomainEvent{event})
	require.NoError(t, err)

	time.Sleep(50 * time.Millisecond)
}

func (ctx *componentCacheTestContext) publishComponentDeleted(t *testing.T, id, name string) {
	event := archEvents.NewApplicationComponentDeleted(id, name)
	err := ctx.eventBus.Publish(tenantContext(), []domain.DomainEvent{event})
	require.NoError(t, err)

	time.Sleep(50 * time.Millisecond)
}

func TestComponentCache_PopulatedByCreatedEvent_Integration(t *testing.T) {
	testCtx, cleanup := setupComponentCacheTestDB(t)
	defer cleanup()

	componentID := uuid.New().String()
	componentName := "Test Component Created"

	testCtx.publishComponentCreated(t, componentID, componentName)

	component, err := testCtx.componentCacheRM.GetByID(tenantContext(), componentID)
	require.NoError(t, err)
	require.NotNil(t, component, "Component should be in cache after Created event")

	assert.Equal(t, componentID, component.ID)
	assert.Equal(t, componentName, component.Name)
}

func TestComponentCache_UpdatedByUpdateEvent_Integration(t *testing.T) {
	testCtx, cleanup := setupComponentCacheTestDB(t)
	defer cleanup()

	componentID := uuid.New().String()
	originalName := "Original Name"
	updatedName := "Updated Name"

	testCtx.publishComponentCreated(t, componentID, originalName)

	component, err := testCtx.componentCacheRM.GetByID(tenantContext(), componentID)
	require.NoError(t, err)
	require.NotNil(t, component)
	assert.Equal(t, originalName, component.Name)

	testCtx.publishComponentUpdated(t, componentID, updatedName)

	component, err = testCtx.componentCacheRM.GetByID(tenantContext(), componentID)
	require.NoError(t, err)
	require.NotNil(t, component)
	assert.Equal(t, updatedName, component.Name)
}

func TestComponentCache_RemovedByDeleteEvent_Integration(t *testing.T) {
	testCtx, cleanup := setupComponentCacheTestDB(t)
	defer cleanup()

	componentID := uuid.New().String()
	componentName := "Component To Delete"

	testCtx.publishComponentCreated(t, componentID, componentName)

	component, err := testCtx.componentCacheRM.GetByID(tenantContext(), componentID)
	require.NoError(t, err)
	require.NotNil(t, component, "Component should exist before delete")

	testCtx.publishComponentDeleted(t, componentID, componentName)

	component, err = testCtx.componentCacheRM.GetByID(tenantContext(), componentID)
	require.NoError(t, err)
	assert.Nil(t, component, "Component should be removed from cache after Delete event")
}

func TestComponentCache_MultipleComponents_Integration(t *testing.T) {
	testCtx, cleanup := setupComponentCacheTestDB(t)
	defer cleanup()

	comp1ID := uuid.New().String()
	comp2ID := uuid.New().String()
	comp3ID := uuid.New().String()

	testCtx.publishComponentCreated(t, comp1ID, "Component 1")
	testCtx.publishComponentCreated(t, comp2ID, "Component 2")
	testCtx.publishComponentCreated(t, comp3ID, "Component 3")

	comp1, err := testCtx.componentCacheRM.GetByID(tenantContext(), comp1ID)
	require.NoError(t, err)
	require.NotNil(t, comp1)
	assert.Equal(t, "Component 1", comp1.Name)

	comp2, err := testCtx.componentCacheRM.GetByID(tenantContext(), comp2ID)
	require.NoError(t, err)
	require.NotNil(t, comp2)
	assert.Equal(t, "Component 2", comp2.Name)

	comp3, err := testCtx.componentCacheRM.GetByID(tenantContext(), comp3ID)
	require.NoError(t, err)
	require.NotNil(t, comp3)
	assert.Equal(t, "Component 3", comp3.Name)

	testCtx.publishComponentDeleted(t, comp2ID, "Component 2")

	comp1, err = testCtx.componentCacheRM.GetByID(tenantContext(), comp1ID)
	require.NoError(t, err)
	assert.NotNil(t, comp1, "Component 1 should still exist")

	comp2, err = testCtx.componentCacheRM.GetByID(tenantContext(), comp2ID)
	require.NoError(t, err)
	assert.Nil(t, comp2, "Component 2 should be deleted")

	comp3, err = testCtx.componentCacheRM.GetByID(tenantContext(), comp3ID)
	require.NoError(t, err)
	assert.NotNil(t, comp3, "Component 3 should still exist")
}

func TestComponentCache_NonExistentComponent_ReturnsNil_Integration(t *testing.T) {
	testCtx, cleanup := setupComponentCacheTestDB(t)
	defer cleanup()

	nonExistentID := uuid.New().String()

	component, err := testCtx.componentCacheRM.GetByID(tenantContext(), nonExistentID)
	require.NoError(t, err)
	assert.Nil(t, component, "Should return nil for non-existent component")
}

func TestComponentCache_UpdateNonExistent_CreatesEntry_Integration(t *testing.T) {
	testCtx, cleanup := setupComponentCacheTestDB(t)
	defer cleanup()

	componentID := uuid.New().String()
	componentName := "Created via Update"

	testCtx.trackComponentID(componentID)
	testCtx.publishComponentUpdated(t, componentID, componentName)

	component, err := testCtx.componentCacheRM.GetByID(tenantContext(), componentID)
	require.NoError(t, err)
	require.NotNil(t, component, "Update event should create entry if not exists (upsert behavior)")
	assert.Equal(t, componentName, component.Name)
}
