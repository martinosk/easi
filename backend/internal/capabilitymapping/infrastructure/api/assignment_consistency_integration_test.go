//go:build integration
// +build integration

package api

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/application/handlers"
	"easi/backend/internal/capabilitymapping/application/projectors"
	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/infrastructure/repositories"
	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/infrastructure/eventstore"
	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/shared/events"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type assignmentConsistencyTestContext struct {
	db                   *sql.DB
	testID               string
	createdCapabilityIDs []string
	createdDomainIDs     []string
	commandBus           *cqrs.InMemoryCommandBus
	capabilityRM         *readmodels.CapabilityReadModel
	assignmentRM         *readmodels.DomainCapabilityAssignmentReadModel
	domainRM             *readmodels.BusinessDomainReadModel
}

func setupAssignmentConsistencyTestDB(t *testing.T) (*assignmentConsistencyTestContext, func()) {
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
	eventStore := eventstore.NewPostgresEventStore(tenantDB)
	commandBus := cqrs.NewInMemoryCommandBus()
	eventBus := events.NewInMemoryEventBus()
	eventStore.SetEventBus(eventBus)

	capabilityRM := readmodels.NewCapabilityReadModel(tenantDB)
	domainRM := readmodels.NewBusinessDomainReadModel(tenantDB)
	assignmentRM := readmodels.NewDomainCapabilityAssignmentReadModel(tenantDB)

	capabilityProjector := projectors.NewCapabilityProjector(capabilityRM, assignmentRM)
	eventBus.Subscribe("CapabilityCreated", capabilityProjector)
	eventBus.Subscribe("CapabilityUpdated", capabilityProjector)
	eventBus.Subscribe("CapabilityParentChanged", capabilityProjector)

	domainProjector := projectors.NewBusinessDomainProjector(domainRM)
	eventBus.Subscribe("BusinessDomainCreated", domainProjector)
	eventBus.Subscribe("BusinessDomainUpdated", domainProjector)
	eventBus.Subscribe("CapabilityAssignedToDomain", domainProjector)
	eventBus.Subscribe("CapabilityUnassignedFromDomain", domainProjector)

	assignmentProjector := projectors.NewBusinessDomainAssignmentProjector(assignmentRM, domainRM, capabilityRM)
	eventBus.Subscribe("CapabilityAssignedToDomain", assignmentProjector)
	eventBus.Subscribe("CapabilityUnassignedFromDomain", assignmentProjector)

	capabilityRepo := repositories.NewCapabilityRepository(eventStore)
	domainRepo := repositories.NewBusinessDomainRepository(eventStore)
	assignmentRepo := repositories.NewBusinessDomainAssignmentRepository(eventStore)

	commandBus.Register("CreateCapability", handlers.NewCreateCapabilityHandler(capabilityRepo))
	commandBus.Register("ChangeCapabilityParent", handlers.NewChangeCapabilityParentHandler(capabilityRepo, capabilityRM))
	commandBus.Register("CreateBusinessDomain", handlers.NewCreateBusinessDomainHandler(domainRepo, domainRM))
	commandBus.Register("AssignCapabilityToDomain", handlers.NewAssignCapabilityToDomainHandler(assignmentRepo, domainRM, capabilityRM, assignmentRM))
	commandBus.Register("UnassignCapabilityFromDomain", handlers.NewUnassignCapabilityFromDomainHandler(assignmentRepo))

	onCapabilityParentChangedHandler := handlers.NewOnCapabilityParentChangedHandler(commandBus, assignmentRM, capabilityRM)
	eventBus.Subscribe("CapabilityParentChanged", onCapabilityParentChangedHandler)

	ctx := &assignmentConsistencyTestContext{
		db:                   db,
		testID:               testID,
		createdCapabilityIDs: make([]string, 0),
		createdDomainIDs:     make([]string, 0),
		commandBus:           commandBus,
		capabilityRM:         capabilityRM,
		assignmentRM:         assignmentRM,
		domainRM:             domainRM,
	}

	cleanup := func() {
		ctx.setTenantContext(t)
		for _, id := range ctx.createdDomainIDs {
			db.Exec("DELETE FROM domain_capability_assignments WHERE business_domain_id = $1", id)
			db.Exec("DELETE FROM business_domains WHERE id = $1", id)
			db.Exec("DELETE FROM events WHERE aggregate_id = $1", id)
		}
		for _, id := range ctx.createdCapabilityIDs {
			db.Exec("DELETE FROM domain_capability_assignments WHERE capability_id = $1", id)
			db.Exec("DELETE FROM capabilities WHERE id = $1", id)
			db.Exec("DELETE FROM events WHERE aggregate_id = $1", id)
		}
		db.Close()
	}

	return ctx, cleanup
}

func (ctx *assignmentConsistencyTestContext) setTenantContext(t *testing.T) {
	_, err := ctx.db.Exec(fmt.Sprintf("SET app.current_tenant = '%s'", testTenantID()))
	require.NoError(t, err)
}

func (ctx *assignmentConsistencyTestContext) trackCapabilityID(id string) {
	ctx.createdCapabilityIDs = append(ctx.createdCapabilityIDs, id)
}

func (ctx *assignmentConsistencyTestContext) trackDomainID(id string) {
	ctx.createdDomainIDs = append(ctx.createdDomainIDs, id)
}

func (ctx *assignmentConsistencyTestContext) createCapability(t *testing.T, name, level, parentID string) string {
	ctx.setTenantContext(t)

	id := fmt.Sprintf("cap-%s-%d", name, time.Now().UnixNano())

	_, err := ctx.db.Exec(
		"INSERT INTO capabilities (id, name, description, level, parent_id, tenant_id, maturity_level, status, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())",
		id, name, "", level, nullString(parentID), testTenantID(), "Genesis", "Active",
	)
	require.NoError(t, err)
	ctx.trackCapabilityID(id)
	return id
}

func (ctx *assignmentConsistencyTestContext) createCapabilityWithEvents(t *testing.T, name, level, parentID string) string {
	ctx.setTenantContext(t)

	id := uuid.New().String()

	_, err := ctx.db.Exec(
		"INSERT INTO capabilities (id, name, description, level, parent_id, tenant_id, maturity_level, status, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())",
		id, name, "", level, nullString(parentID), testTenantID(), "Genesis", "Active",
	)
	require.NoError(t, err)

	var parentIDJSON string
	if parentID != "" {
		parentIDJSON = fmt.Sprintf(`,"parentId":"%s"`, parentID)
	}
	eventData := fmt.Sprintf(`{"id":"%s","name":"%s","description":"","level":"%s"%s,"createdAt":"%s"}`,
		id, name, level, parentIDJSON, time.Now().Format(time.RFC3339Nano))
	_, err = ctx.db.Exec(
		"INSERT INTO events (tenant_id, aggregate_id, event_type, event_data, version, occurred_at) VALUES ($1, $2, $3, $4, $5, NOW())",
		testTenantID(), id, "CapabilityCreated", eventData, 1,
	)
	require.NoError(t, err)

	ctx.trackCapabilityID(id)
	return id
}

func (ctx *assignmentConsistencyTestContext) createDomain(t *testing.T, name string) string {
	ctx.setTenantContext(t)

	id := uuid.New().String()

	_, err := ctx.db.Exec(
		"INSERT INTO business_domains (id, name, description, capability_count, tenant_id, created_at) VALUES ($1, $2, $3, 0, $4, NOW())",
		id, name, "", testTenantID(),
	)
	require.NoError(t, err)
	ctx.trackDomainID(id)
	return id
}

func (ctx *assignmentConsistencyTestContext) assignCapabilityToDomain(t *testing.T, domainID, capabilityID string) string {
	ctx.setTenantContext(t)

	cap, err := ctx.capabilityRM.GetByID(tenantContext(), capabilityID)
	require.NoError(t, err)
	require.NotNil(t, cap)

	dom, err := ctx.domainRM.GetByID(tenantContext(), domainID)
	require.NoError(t, err)
	require.NotNil(t, dom)

	assignmentID := fmt.Sprintf("assign-%d", time.Now().UnixNano())

	_, err = ctx.db.Exec(
		"INSERT INTO domain_capability_assignments (assignment_id, business_domain_id, business_domain_name, capability_id, capability_name, capability_level, tenant_id, assigned_at) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())",
		assignmentID, domainID, dom.Name, capabilityID, cap.Name, cap.Level, testTenantID(),
	)
	require.NoError(t, err)

	_, err = ctx.db.Exec(
		"UPDATE business_domains SET capability_count = capability_count + 1 WHERE id = $1 AND tenant_id = $2",
		domainID, testTenantID(),
	)
	require.NoError(t, err)

	return assignmentID
}

func (ctx *assignmentConsistencyTestContext) assignCapabilityToDomainWithEvents(t *testing.T, domainID, capabilityID string) string {
	ctx.setTenantContext(t)

	cap, err := ctx.capabilityRM.GetByID(tenantContext(), capabilityID)
	require.NoError(t, err)
	require.NotNil(t, cap)

	dom, err := ctx.domainRM.GetByID(tenantContext(), domainID)
	require.NoError(t, err)
	require.NotNil(t, dom)

	assignmentID := fmt.Sprintf("assign-%d", time.Now().UnixNano())

	_, err = ctx.db.Exec(
		"INSERT INTO domain_capability_assignments (assignment_id, business_domain_id, business_domain_name, capability_id, capability_name, capability_level, tenant_id, assigned_at) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())",
		assignmentID, domainID, dom.Name, capabilityID, cap.Name, cap.Level, testTenantID(),
	)
	require.NoError(t, err)

	_, err = ctx.db.Exec(
		"UPDATE business_domains SET capability_count = capability_count + 1 WHERE id = $1 AND tenant_id = $2",
		domainID, testTenantID(),
	)
	require.NoError(t, err)

	eventData := fmt.Sprintf(`{"id":"%s","businessDomainId":"%s","capabilityId":"%s","assignedAt":"%s"}`,
		assignmentID, domainID, capabilityID, time.Now().Format(time.RFC3339Nano))
	_, err = ctx.db.Exec(
		"INSERT INTO events (tenant_id, aggregate_id, event_type, event_data, version, occurred_at) VALUES ($1, $2, $3, $4, $5, NOW())",
		testTenantID(), assignmentID, "CapabilityAssignedToDomain", eventData, 1,
	)
	require.NoError(t, err)

	return assignmentID
}

func nullString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func TestL1ToL2_ReassignsToDomain_Integration(t *testing.T) {
	testCtx, cleanup := setupAssignmentConsistencyTestDB(t)
	defer cleanup()

	l1Child := testCtx.createCapabilityWithEvents(t, "ChildCap", "L1", "")
	l1Parent := testCtx.createCapabilityWithEvents(t, "ParentCap", "L1", "")
	domainID := testCtx.createDomain(t, "TestDomain")

	testCtx.assignCapabilityToDomainWithEvents(t, domainID, l1Child)

	time.Sleep(100 * time.Millisecond)

	assignmentsBefore, err := testCtx.assignmentRM.GetByDomainID(tenantContext(), domainID)
	require.NoError(t, err)
	assert.Len(t, assignmentsBefore, 1)
	assert.Equal(t, l1Child, assignmentsBefore[0].CapabilityID)

	testCtx.setTenantContext(t)
	_, err = testCtx.db.Exec(
		"UPDATE capabilities SET parent_id = $1, level = 'L2' WHERE id = $2 AND tenant_id = $3",
		l1Parent, l1Child, testTenantID(),
	)
	require.NoError(t, err)

	tenantDB := database.NewTenantAwareDB(testCtx.db)
	eventStore := eventstore.NewPostgresEventStore(tenantDB)
	eventBus := events.NewInMemoryEventBus()
	eventStore.SetEventBus(eventBus)

	capabilityProjector := projectors.NewCapabilityProjector(testCtx.capabilityRM, testCtx.assignmentRM)
	eventBus.Subscribe("CapabilityParentChanged", capabilityProjector)

	domainProjector := projectors.NewBusinessDomainProjector(testCtx.domainRM)
	eventBus.Subscribe("CapabilityAssignedToDomain", domainProjector)
	eventBus.Subscribe("CapabilityUnassignedFromDomain", domainProjector)

	assignmentProjector := projectors.NewBusinessDomainAssignmentProjector(testCtx.assignmentRM, testCtx.domainRM, testCtx.capabilityRM)
	eventBus.Subscribe("CapabilityAssignedToDomain", assignmentProjector)
	eventBus.Subscribe("CapabilityUnassignedFromDomain", assignmentProjector)

	commandBus := cqrs.NewInMemoryCommandBus()
	assignmentRepo := repositories.NewBusinessDomainAssignmentRepository(eventStore)
	commandBus.Register("AssignCapabilityToDomain", handlers.NewAssignCapabilityToDomainHandler(assignmentRepo, testCtx.domainRM, testCtx.capabilityRM, testCtx.assignmentRM))
	commandBus.Register("UnassignCapabilityFromDomain", handlers.NewUnassignCapabilityFromDomainHandler(assignmentRepo))

	onCapabilityParentChangedHandler := handlers.NewOnCapabilityParentChangedHandler(commandBus, testCtx.assignmentRM, testCtx.capabilityRM)
	eventBus.Subscribe("CapabilityParentChanged", onCapabilityParentChangedHandler)

	capabilityRepo := repositories.NewCapabilityRepository(eventStore)
	changeParentHandler := handlers.NewChangeCapabilityParentHandler(capabilityRepo, testCtx.capabilityRM)
	commandBus.Register("ChangeCapabilityParent", changeParentHandler)

	cmd := &commands.ChangeCapabilityParent{
		CapabilityID: l1Child,
		NewParentID:  l1Parent,
	}
	err = commandBus.Dispatch(tenantContext(), cmd)
	require.NoError(t, err)

	time.Sleep(200 * time.Millisecond)

	assignmentsAfter, err := testCtx.assignmentRM.GetByDomainID(tenantContext(), domainID)
	require.NoError(t, err)

	assert.Len(t, assignmentsAfter, 1, "Should have exactly 1 assignment")
	assert.Equal(t, l1Parent, assignmentsAfter[0].CapabilityID, "The L1 parent should now be assigned")
}

func TestL1ToL3_FindsL1Ancestor_Integration(t *testing.T) {
	testCtx, cleanup := setupAssignmentConsistencyTestDB(t)
	defer cleanup()

	l1Root := testCtx.createCapabilityWithEvents(t, "L1Root", "L1", "")
	l2Middle := testCtx.createCapabilityWithEvents(t, "L2Middle", "L2", l1Root)
	l1Child := testCtx.createCapabilityWithEvents(t, "ChildCap", "L1", "")
	domainID := testCtx.createDomain(t, "TestDomain")

	testCtx.assignCapabilityToDomainWithEvents(t, domainID, l1Child)

	time.Sleep(100 * time.Millisecond)

	assignmentsBefore, err := testCtx.assignmentRM.GetByDomainID(tenantContext(), domainID)
	require.NoError(t, err)
	assert.Len(t, assignmentsBefore, 1)
	assert.Equal(t, l1Child, assignmentsBefore[0].CapabilityID)

	testCtx.setTenantContext(t)
	_, err = testCtx.db.Exec(
		"UPDATE capabilities SET parent_id = $1, level = 'L3' WHERE id = $2 AND tenant_id = $3",
		l2Middle, l1Child, testTenantID(),
	)
	require.NoError(t, err)

	tenantDB := database.NewTenantAwareDB(testCtx.db)
	eventStore := eventstore.NewPostgresEventStore(tenantDB)
	eventBus := events.NewInMemoryEventBus()
	eventStore.SetEventBus(eventBus)

	capabilityProjector := projectors.NewCapabilityProjector(testCtx.capabilityRM, testCtx.assignmentRM)
	eventBus.Subscribe("CapabilityParentChanged", capabilityProjector)

	domainProjector := projectors.NewBusinessDomainProjector(testCtx.domainRM)
	eventBus.Subscribe("CapabilityAssignedToDomain", domainProjector)
	eventBus.Subscribe("CapabilityUnassignedFromDomain", domainProjector)

	assignmentProjector := projectors.NewBusinessDomainAssignmentProjector(testCtx.assignmentRM, testCtx.domainRM, testCtx.capabilityRM)
	eventBus.Subscribe("CapabilityAssignedToDomain", assignmentProjector)
	eventBus.Subscribe("CapabilityUnassignedFromDomain", assignmentProjector)

	commandBus := cqrs.NewInMemoryCommandBus()
	assignmentRepo := repositories.NewBusinessDomainAssignmentRepository(eventStore)
	commandBus.Register("AssignCapabilityToDomain", handlers.NewAssignCapabilityToDomainHandler(assignmentRepo, testCtx.domainRM, testCtx.capabilityRM, testCtx.assignmentRM))
	commandBus.Register("UnassignCapabilityFromDomain", handlers.NewUnassignCapabilityFromDomainHandler(assignmentRepo))

	onCapabilityParentChangedHandler := handlers.NewOnCapabilityParentChangedHandler(commandBus, testCtx.assignmentRM, testCtx.capabilityRM)
	eventBus.Subscribe("CapabilityParentChanged", onCapabilityParentChangedHandler)

	capabilityRepo := repositories.NewCapabilityRepository(eventStore)
	changeParentHandler := handlers.NewChangeCapabilityParentHandler(capabilityRepo, testCtx.capabilityRM)
	commandBus.Register("ChangeCapabilityParent", changeParentHandler)

	cmd := &commands.ChangeCapabilityParent{
		CapabilityID: l1Child,
		NewParentID:  l2Middle,
	}
	err = commandBus.Dispatch(tenantContext(), cmd)
	require.NoError(t, err)

	time.Sleep(200 * time.Millisecond)

	assignmentsAfter, err := testCtx.assignmentRM.GetByDomainID(tenantContext(), domainID)
	require.NoError(t, err)

	assert.Len(t, assignmentsAfter, 1, "Should have exactly 1 assignment")
	assert.Equal(t, l1Root, assignmentsAfter[0].CapabilityID, "The L1 root ancestor should now be assigned")
}

func TestL1ToL2_ParentAlreadyAssigned_NoDuplicate_Integration(t *testing.T) {
	testCtx, cleanup := setupAssignmentConsistencyTestDB(t)
	defer cleanup()

	l1Child := testCtx.createCapabilityWithEvents(t, "ChildCap", "L1", "")
	l1Parent := testCtx.createCapabilityWithEvents(t, "ParentCap", "L1", "")
	domainID := testCtx.createDomain(t, "TestDomain")

	testCtx.assignCapabilityToDomainWithEvents(t, domainID, l1Child)
	testCtx.assignCapabilityToDomainWithEvents(t, domainID, l1Parent)

	time.Sleep(100 * time.Millisecond)

	assignmentsBefore, err := testCtx.assignmentRM.GetByDomainID(tenantContext(), domainID)
	require.NoError(t, err)
	assert.Len(t, assignmentsBefore, 2)

	testCtx.setTenantContext(t)
	_, err = testCtx.db.Exec(
		"UPDATE capabilities SET parent_id = $1, level = 'L2' WHERE id = $2 AND tenant_id = $3",
		l1Parent, l1Child, testTenantID(),
	)
	require.NoError(t, err)

	tenantDB := database.NewTenantAwareDB(testCtx.db)
	eventStore := eventstore.NewPostgresEventStore(tenantDB)
	eventBus := events.NewInMemoryEventBus()
	eventStore.SetEventBus(eventBus)

	capabilityProjector := projectors.NewCapabilityProjector(testCtx.capabilityRM, testCtx.assignmentRM)
	eventBus.Subscribe("CapabilityParentChanged", capabilityProjector)

	domainProjector := projectors.NewBusinessDomainProjector(testCtx.domainRM)
	eventBus.Subscribe("CapabilityAssignedToDomain", domainProjector)
	eventBus.Subscribe("CapabilityUnassignedFromDomain", domainProjector)

	assignmentProjector := projectors.NewBusinessDomainAssignmentProjector(testCtx.assignmentRM, testCtx.domainRM, testCtx.capabilityRM)
	eventBus.Subscribe("CapabilityAssignedToDomain", assignmentProjector)
	eventBus.Subscribe("CapabilityUnassignedFromDomain", assignmentProjector)

	commandBus := cqrs.NewInMemoryCommandBus()
	assignmentRepo := repositories.NewBusinessDomainAssignmentRepository(eventStore)
	commandBus.Register("AssignCapabilityToDomain", handlers.NewAssignCapabilityToDomainHandler(assignmentRepo, testCtx.domainRM, testCtx.capabilityRM, testCtx.assignmentRM))
	commandBus.Register("UnassignCapabilityFromDomain", handlers.NewUnassignCapabilityFromDomainHandler(assignmentRepo))

	onCapabilityParentChangedHandler := handlers.NewOnCapabilityParentChangedHandler(commandBus, testCtx.assignmentRM, testCtx.capabilityRM)
	eventBus.Subscribe("CapabilityParentChanged", onCapabilityParentChangedHandler)

	capabilityRepo := repositories.NewCapabilityRepository(eventStore)
	changeParentHandler := handlers.NewChangeCapabilityParentHandler(capabilityRepo, testCtx.capabilityRM)
	commandBus.Register("ChangeCapabilityParent", changeParentHandler)

	cmd := &commands.ChangeCapabilityParent{
		CapabilityID: l1Child,
		NewParentID:  l1Parent,
	}
	err = commandBus.Dispatch(tenantContext(), cmd)
	require.NoError(t, err)

	time.Sleep(200 * time.Millisecond)

	assignmentsAfter, err := testCtx.assignmentRM.GetByDomainID(tenantContext(), domainID)
	require.NoError(t, err)

	assert.Len(t, assignmentsAfter, 1, "Should have exactly 1 assignment (parent was already assigned)")
	assert.Equal(t, l1Parent, assignmentsAfter[0].CapabilityID, "Only the L1 parent should remain assigned")
}
