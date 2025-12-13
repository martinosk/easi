//go:build integration
// +build integration

package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"easi/backend/internal/capabilitymapping/application/handlers"
	"easi/backend/internal/capabilitymapping/application/projectors"
	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/infrastructure/repositories"
	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/infrastructure/eventstore"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/shared/events"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type businessDomainTestContext struct {
	db                   *sql.DB
	testID               string
	createdDomainIDs     []string
	createdCapabilityIDs []string
}

func (ctx *businessDomainTestContext) setTenantContext(t *testing.T) {
	_, err := ctx.db.Exec(fmt.Sprintf("SET app.current_tenant = '%s'", testTenantID()))
	require.NoError(t, err)
}

func setupBusinessDomainTestDB(t *testing.T) (*businessDomainTestContext, func()) {
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

	ctx := &businessDomainTestContext{
		db:                   db,
		testID:               testID,
		createdDomainIDs:     make([]string, 0),
		createdCapabilityIDs: make([]string, 0),
	}

	cleanup := func() {
		ctx.setTenantContext(t)
		for _, id := range ctx.createdDomainIDs {
			db.Exec("DELETE FROM domain_capability_assignments WHERE business_domain_id = $1", id)
			db.Exec("DELETE FROM business_domains WHERE id = $1", id)
			db.Exec("DELETE FROM events WHERE aggregate_id = $1", id)
		}
		for _, id := range ctx.createdCapabilityIDs {
			db.Exec("DELETE FROM capabilities WHERE id = $1", id)
			db.Exec("DELETE FROM events WHERE aggregate_id = $1", id)
		}
		db.Close()
	}

	return ctx, cleanup
}

func (ctx *businessDomainTestContext) trackDomainID(id string) {
	ctx.createdDomainIDs = append(ctx.createdDomainIDs, id)
}

func (ctx *businessDomainTestContext) trackCapabilityID(id string) {
	ctx.createdCapabilityIDs = append(ctx.createdCapabilityIDs, id)
}

func (ctx *businessDomainTestContext) createTestDomain(t *testing.T, id, name, description string) {
	ctx.setTenantContext(t)
	_, err := ctx.db.Exec(
		"INSERT INTO business_domains (id, name, description, capability_count, tenant_id, created_at) VALUES ($1, $2, $3, 0, $4, NOW())",
		id, name, description, testTenantID(),
	)
	require.NoError(t, err)
	ctx.trackDomainID(id)
}

func (ctx *businessDomainTestContext) createTestCapability(t *testing.T, id, name, level string) {
	ctx.setTenantContext(t)
	_, err := ctx.db.Exec(
		"INSERT INTO capabilities (id, name, description, level, tenant_id, maturity_level, status, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())",
		id, name, "", level, testTenantID(), "Genesis", "Active",
	)
	require.NoError(t, err)
	ctx.trackCapabilityID(id)
}

func (ctx *businessDomainTestContext) createTestCapabilityWithParent(t *testing.T, id, name, level, parentID string) {
	ctx.setTenantContext(t)
	_, err := ctx.db.Exec(
		"INSERT INTO capabilities (id, name, description, level, parent_id, tenant_id, maturity_level, status, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())",
		id, name, "", level, parentID, testTenantID(), "Genesis", "Active",
	)
	require.NoError(t, err)
	ctx.trackCapabilityID(id)
}

func (ctx *businessDomainTestContext) createTestRealization(t *testing.T, id, capabilityID, componentID, componentName, level, origin string) {
	ctx.setTenantContext(t)
	_, err := ctx.db.Exec(
		"INSERT INTO capability_realizations (id, capability_id, component_id, component_name, realization_level, origin, tenant_id, linked_at) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())",
		id, capabilityID, componentID, componentName, level, origin, testTenantID(),
	)
	require.NoError(t, err)
}

func (ctx *businessDomainTestContext) createTestComponent(t *testing.T, id, name string) {
	ctx.setTenantContext(t)
	_, err := ctx.db.Exec(
		"INSERT INTO application_components (id, name, description, tenant_id, created_at) VALUES ($1, $2, $3, $4, NOW())",
		id, name, "", testTenantID(),
	)
	require.NoError(t, err)
}

func setupBusinessDomainHandlers(db *sql.DB) *BusinessDomainHandlers {
	tenantDB := database.NewTenantAwareDB(db)

	eventStore := eventstore.NewPostgresEventStore(tenantDB)
	commandBus := cqrs.NewInMemoryCommandBus()
	hateoas := sharedAPI.NewHATEOASLinks("/api/v1")

	eventBus := events.NewInMemoryEventBus()
	eventStore.SetEventBus(eventBus)

	domainRM := readmodels.NewBusinessDomainReadModel(tenantDB)
	assignmentRM := readmodels.NewDomainCapabilityAssignmentReadModel(tenantDB)
	capabilityRM := readmodels.NewCapabilityReadModel(tenantDB)

	domainProjector := projectors.NewBusinessDomainProjector(domainRM)
	eventBus.Subscribe("BusinessDomainCreated", domainProjector)
	eventBus.Subscribe("BusinessDomainUpdated", domainProjector)
	eventBus.Subscribe("BusinessDomainDeleted", domainProjector)
	eventBus.Subscribe("CapabilityAssignedToDomain", domainProjector)
	eventBus.Subscribe("CapabilityUnassignedFromDomain", domainProjector)

	assignmentProjector := projectors.NewBusinessDomainAssignmentProjector(assignmentRM, domainRM, capabilityRM)
	eventBus.Subscribe("CapabilityAssignedToDomain", assignmentProjector)
	eventBus.Subscribe("CapabilityUnassignedFromDomain", assignmentProjector)

	domainRepo := repositories.NewBusinessDomainRepository(eventStore)
	assignmentRepo := repositories.NewBusinessDomainAssignmentRepository(eventStore)

	commandBus.Register("CreateBusinessDomain", handlers.NewCreateBusinessDomainHandler(domainRepo, domainRM))
	commandBus.Register("UpdateBusinessDomain", handlers.NewUpdateBusinessDomainHandler(domainRepo, domainRM))
	commandBus.Register("DeleteBusinessDomain", handlers.NewDeleteBusinessDomainHandler(domainRepo, assignmentRM))
	commandBus.Register("AssignCapabilityToDomain", handlers.NewAssignCapabilityToDomainHandler(assignmentRepo, domainRM, capabilityRM, assignmentRM))
	commandBus.Register("UnassignCapabilityFromDomain", handlers.NewUnassignCapabilityFromDomainHandler(assignmentRepo))

	realizationRM := readmodels.NewRealizationReadModel(tenantDB)

	readModels := &BusinessDomainReadModels{
		Domain:      domainRM,
		Assignment:  assignmentRM,
		Capability:  capabilityRM,
		Realization: realizationRM,
	}

	return NewBusinessDomainHandlers(commandBus, readModels, hateoas)
}

func TestCreateBusinessDomain_Integration(t *testing.T) {
	testCtx, cleanup := setupBusinessDomainTestDB(t)
	defer cleanup()

	handler := setupBusinessDomainHandlers(testCtx.db)

	reqBody := CreateBusinessDomainRequest{
		Name:        "Customer Experience",
		Description: "Customer-facing capabilities",
	}
	body, _ := json.Marshal(reqBody)

	w, req := makeRequest(t, http.MethodPost, "/api/v1/business-domains", body, nil)

	handler.CreateBusinessDomain(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Header().Get("Location"), "/api/v1/business-domains/")

	var response readmodels.BusinessDomainDTO
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "Customer Experience", response.Name)
	assert.Equal(t, "Customer-facing capabilities", response.Description)
	assert.Equal(t, 0, response.CapabilityCount)
	assert.NotNil(t, response.Links)
	assert.Contains(t, response.Links, "self")
	assert.Contains(t, response.Links, "delete")

	testCtx.trackDomainID(response.ID)

	testCtx.setTenantContext(t)
	var eventData string
	err = testCtx.db.QueryRow(
		"SELECT event_data FROM events WHERE aggregate_id = $1 AND event_type = 'BusinessDomainCreated'",
		response.ID,
	).Scan(&eventData)
	require.NoError(t, err)
	assert.Contains(t, eventData, "Customer Experience")
}

func TestGetAllBusinessDomains_Integration(t *testing.T) {
	testCtx, cleanup := setupBusinessDomainTestDB(t)
	defer cleanup()

	handler := setupBusinessDomainHandlers(testCtx.db)

	id1 := fmt.Sprintf("test-domain-1-%d", time.Now().UnixNano())
	id2 := fmt.Sprintf("test-domain-2-%d", time.Now().UnixNano())
	testCtx.createTestDomain(t, id1, "Sales", "Sales capabilities")
	testCtx.createTestDomain(t, id2, "Marketing", "Marketing capabilities")

	w, req := makeRequest(t, http.MethodGet, "/api/v1/business-domains", nil, nil)

	handler.GetAllBusinessDomains(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response sharedAPI.PaginatedResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	dataBytes, _ := json.Marshal(response.Data)
	var domains []readmodels.BusinessDomainDTO
	json.Unmarshal(dataBytes, &domains)

	foundDomains := 0
	for _, domain := range domains {
		if domain.ID == id1 || domain.ID == id2 {
			foundDomains++
			assert.NotNil(t, domain.Links)
			assert.Contains(t, domain.Links, "self")
		}
	}
	assert.Equal(t, 2, foundDomains)
	assert.NotNil(t, response.Links)
	assert.False(t, response.Pagination.HasMore)
}

func TestGetBusinessDomainByID_Integration(t *testing.T) {
	testCtx, cleanup := setupBusinessDomainTestDB(t)
	defer cleanup()

	handler := setupBusinessDomainHandlers(testCtx.db)

	domainID := fmt.Sprintf("test-domain-%d", time.Now().UnixNano())
	testCtx.createTestDomain(t, domainID, "Operations", "Operational capabilities")

	w, req := makeRequest(t, http.MethodGet, "/api/v1/business-domains/"+domainID, nil, map[string]string{"id": domainID})

	handler.GetBusinessDomainByID(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response readmodels.BusinessDomainDTO
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, domainID, response.ID)
	assert.Equal(t, "Operations", response.Name)
	assert.NotNil(t, response.Links)
	assert.Contains(t, response.Links, "self")
}

func TestGetBusinessDomainByID_NotFound_Integration(t *testing.T) {
	testCtx, cleanup := setupBusinessDomainTestDB(t)
	defer cleanup()

	handler := setupBusinessDomainHandlers(testCtx.db)

	nonExistentID := fmt.Sprintf("non-existent-%d", time.Now().UnixNano())
	w, req := makeRequest(t, http.MethodGet, "/api/v1/business-domains/"+nonExistentID, nil, map[string]string{"id": nonExistentID})

	handler.GetBusinessDomainByID(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateBusinessDomain_Integration(t *testing.T) {
	testCtx, cleanup := setupBusinessDomainTestDB(t)
	defer cleanup()

	handler := setupBusinessDomainHandlers(testCtx.db)

	createReqBody := CreateBusinessDomainRequest{
		Name:        "Original Name",
		Description: "Original description",
	}
	createBody, _ := json.Marshal(createReqBody)

	createW, createReq := makeRequest(t, http.MethodPost, "/api/v1/business-domains", createBody, nil)
	handler.CreateBusinessDomain(createW, createReq)
	assert.Equal(t, http.StatusCreated, createW.Code)

	var createdDomain readmodels.BusinessDomainDTO
	json.NewDecoder(createW.Body).Decode(&createdDomain)
	testCtx.trackDomainID(createdDomain.ID)

	time.Sleep(100 * time.Millisecond)

	updateReqBody := UpdateBusinessDomainRequest{
		Name:        "Updated Name",
		Description: "Updated description",
	}
	updateBody, _ := json.Marshal(updateReqBody)

	updateW, updateReq := makeRequest(t, http.MethodPut, "/api/v1/business-domains/"+createdDomain.ID, updateBody, map[string]string{"id": createdDomain.ID})
	handler.UpdateBusinessDomain(updateW, updateReq)

	assert.Equal(t, http.StatusOK, updateW.Code)

	var response readmodels.BusinessDomainDTO
	err := json.NewDecoder(updateW.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", response.Name)
	assert.Equal(t, "Updated description", response.Description)

	testCtx.setTenantContext(t)
	var eventData string
	err = testCtx.db.QueryRow(
		"SELECT event_data FROM events WHERE aggregate_id = $1 AND event_type = 'BusinessDomainUpdated'",
		createdDomain.ID,
	).Scan(&eventData)
	require.NoError(t, err)
	assert.Contains(t, eventData, "Updated Name")
}

func TestUpdateBusinessDomain_ValidationError_Integration(t *testing.T) {
	testCtx, cleanup := setupBusinessDomainTestDB(t)
	defer cleanup()

	handler := setupBusinessDomainHandlers(testCtx.db)

	domainID := fmt.Sprintf("test-domain-%d", time.Now().UnixNano())
	testCtx.createTestDomain(t, domainID, "Test Domain", "Description")

	updateReqBody := UpdateBusinessDomainRequest{
		Name:        "",
		Description: "Updated description",
	}
	updateBody, _ := json.Marshal(updateReqBody)

	w, req := makeRequest(t, http.MethodPut, "/api/v1/business-domains/"+domainID, updateBody, map[string]string{"id": domainID})
	handler.UpdateBusinessDomain(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateBusinessDomain_DuplicateName_Integration(t *testing.T) {
	testCtx, cleanup := setupBusinessDomainTestDB(t)
	defer cleanup()

	handler := setupBusinessDomainHandlers(testCtx.db)

	id1 := fmt.Sprintf("test-domain-1-%d", time.Now().UnixNano())
	id2 := fmt.Sprintf("test-domain-2-%d", time.Now().UnixNano())
	testCtx.createTestDomain(t, id1, "Existing Domain", "Description 1")
	testCtx.createTestDomain(t, id2, "Other Domain", "Description 2")

	updateReqBody := UpdateBusinessDomainRequest{
		Name:        "Existing Domain",
		Description: "Updated description",
	}
	updateBody, _ := json.Marshal(updateReqBody)

	w, req := makeRequest(t, http.MethodPut, "/api/v1/business-domains/"+id2, updateBody, map[string]string{"id": id2})
	handler.UpdateBusinessDomain(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestDeleteBusinessDomain_Integration(t *testing.T) {
	testCtx, cleanup := setupBusinessDomainTestDB(t)
	defer cleanup()

	handler := setupBusinessDomainHandlers(testCtx.db)

	createReqBody := CreateBusinessDomainRequest{
		Name:        "Domain To Delete",
		Description: "This will be deleted",
	}
	createBody, _ := json.Marshal(createReqBody)

	createW, createReq := makeRequest(t, http.MethodPost, "/api/v1/business-domains", createBody, nil)
	handler.CreateBusinessDomain(createW, createReq)
	assert.Equal(t, http.StatusCreated, createW.Code)

	var createdDomain readmodels.BusinessDomainDTO
	json.NewDecoder(createW.Body).Decode(&createdDomain)
	testCtx.trackDomainID(createdDomain.ID)

	time.Sleep(100 * time.Millisecond)

	deleteW, deleteReq := makeRequest(t, http.MethodDelete, "/api/v1/business-domains/"+createdDomain.ID, nil, map[string]string{"id": createdDomain.ID})
	handler.DeleteBusinessDomain(deleteW, deleteReq)

	assert.Equal(t, http.StatusNoContent, deleteW.Code)

	time.Sleep(100 * time.Millisecond)

	testCtx.setTenantContext(t)
	var count int
	err := testCtx.db.QueryRow(
		"SELECT COUNT(*) FROM business_domains WHERE id = $1",
		createdDomain.ID,
	).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestDeleteBusinessDomain_NotFound_Integration(t *testing.T) {
	testCtx, cleanup := setupBusinessDomainTestDB(t)
	defer cleanup()

	handler := setupBusinessDomainHandlers(testCtx.db)

	nonExistentID := fmt.Sprintf("non-existent-%d", time.Now().UnixNano())
	w, req := makeRequest(t, http.MethodDelete, "/api/v1/business-domains/"+nonExistentID, nil, map[string]string{"id": nonExistentID})

	handler.DeleteBusinessDomain(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDeleteBusinessDomain_HasCapabilities_Integration(t *testing.T) {
	testCtx, cleanup := setupBusinessDomainTestDB(t)
	defer cleanup()

	handler := setupBusinessDomainHandlers(testCtx.db)

	domainID := fmt.Sprintf("test-domain-%d", time.Now().UnixNano())
	capID := fmt.Sprintf("test-cap-%d", time.Now().UnixNano())

	testCtx.createTestDomain(t, domainID, "Test Domain", "Description")
	testCtx.createTestCapability(t, capID, "Test Capability", "L1")

	testCtx.setTenantContext(t)
	assignmentID := fmt.Sprintf("test-assignment-%d", time.Now().UnixNano())
	_, err := testCtx.db.Exec(
		"INSERT INTO domain_capability_assignments (assignment_id, business_domain_id, business_domain_name, capability_id, capability_code, capability_name, capability_level, tenant_id, assigned_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())",
		assignmentID, domainID, "Test Domain", capID, "BIZ-001", "Test Capability", "L1", testTenantID(),
	)
	require.NoError(t, err)

	_, err = testCtx.db.Exec(
		"UPDATE business_domains SET capability_count = 1 WHERE id = $1",
		domainID,
	)
	require.NoError(t, err)

	w, req := makeRequest(t, http.MethodDelete, "/api/v1/business-domains/"+domainID, nil, map[string]string{"id": domainID})
	handler.DeleteBusinessDomain(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestGetCapabilitiesInDomain_Integration(t *testing.T) {
	testCtx, cleanup := setupBusinessDomainTestDB(t)
	defer cleanup()

	handler := setupBusinessDomainHandlers(testCtx.db)

	domainID := fmt.Sprintf("test-domain-%d", time.Now().UnixNano())
	cap1ID := fmt.Sprintf("test-cap-1-%d", time.Now().UnixNano())
	cap2ID := fmt.Sprintf("test-cap-2-%d", time.Now().UnixNano())

	testCtx.createTestDomain(t, domainID, "Test Domain", "Description")
	testCtx.createTestCapability(t, cap1ID, "Capability 1", "L1")
	testCtx.createTestCapability(t, cap2ID, "Capability 2", "L1")

	testCtx.setTenantContext(t)
	assignment1ID := fmt.Sprintf("test-assignment-1-%d", time.Now().UnixNano())
	_, err := testCtx.db.Exec(
		"INSERT INTO domain_capability_assignments (assignment_id, business_domain_id, business_domain_name, capability_id, capability_code, capability_name, capability_level, tenant_id, assigned_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())",
		assignment1ID, domainID, "Test Domain", cap1ID, "BIZ-001", "Capability 1", "L1", testTenantID(),
	)
	require.NoError(t, err)

	assignment2ID := fmt.Sprintf("test-assignment-2-%d", time.Now().UnixNano())
	_, err = testCtx.db.Exec(
		"INSERT INTO domain_capability_assignments (assignment_id, business_domain_id, business_domain_name, capability_id, capability_code, capability_name, capability_level, tenant_id, assigned_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())",
		assignment2ID, domainID, "Test Domain", cap2ID, "BIZ-002", "Capability 2", "L1", testTenantID(),
	)
	require.NoError(t, err)

	w, req := makeRequest(t, http.MethodGet, "/api/v1/business-domains/"+domainID+"/capabilities", nil, map[string]string{"id": domainID})
	handler.GetCapabilitiesInDomain(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response sharedAPI.PaginatedResponse
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	dataBytes, _ := json.Marshal(response.Data)
	var capabilities []CapabilityInDomainDTO
	json.Unmarshal(dataBytes, &capabilities)

	assert.Equal(t, 2, len(capabilities))
	assert.NotNil(t, response.Links)
	assert.Contains(t, response.Links, "domain")
}

func TestGetCapabilitiesInDomain_DomainNotFound_Integration(t *testing.T) {
	testCtx, cleanup := setupBusinessDomainTestDB(t)
	defer cleanup()

	handler := setupBusinessDomainHandlers(testCtx.db)

	nonExistentID := fmt.Sprintf("non-existent-%d", time.Now().UnixNano())
	w, req := makeRequest(t, http.MethodGet, "/api/v1/business-domains/"+nonExistentID+"/capabilities", nil, map[string]string{"id": nonExistentID})

	handler.GetCapabilitiesInDomain(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAssignCapabilityToDomain_Integration(t *testing.T) {
	testCtx, cleanup := setupBusinessDomainTestDB(t)
	defer cleanup()

	handler := setupBusinessDomainHandlers(testCtx.db)

	domainID := fmt.Sprintf("test-domain-%d", time.Now().UnixNano())
	capID := fmt.Sprintf("test-cap-%d", time.Now().UnixNano())

	testCtx.createTestDomain(t, domainID, "Test Domain", "Description")
	testCtx.createTestCapability(t, capID, "Test Capability", "L1")

	time.Sleep(100 * time.Millisecond)

	reqBody := AssignCapabilityRequest{
		CapabilityID: capID,
	}
	body, _ := json.Marshal(reqBody)

	w, req := makeRequest(t, http.MethodPost, "/api/v1/business-domains/"+domainID+"/capabilities", body, map[string]string{"id": domainID})
	handler.AssignCapabilityToDomain(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Header().Get("Location"), "/api/v1/business-domains/"+domainID+"/capabilities/"+capID)

	time.Sleep(100 * time.Millisecond)

	testCtx.setTenantContext(t)
	var count int
	err := testCtx.db.QueryRow(
		"SELECT COUNT(*) FROM domain_capability_assignments WHERE business_domain_id = $1 AND capability_id = $2",
		domainID, capID,
	).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestAssignCapabilityToDomain_CapabilityNotFound_Integration(t *testing.T) {
	testCtx, cleanup := setupBusinessDomainTestDB(t)
	defer cleanup()

	handler := setupBusinessDomainHandlers(testCtx.db)

	domainID := fmt.Sprintf("test-domain-%d", time.Now().UnixNano())
	testCtx.createTestDomain(t, domainID, "Test Domain", "Description")

	nonExistentCapID := fmt.Sprintf("non-existent-%d", time.Now().UnixNano())
	reqBody := AssignCapabilityRequest{
		CapabilityID: nonExistentCapID,
	}
	body, _ := json.Marshal(reqBody)

	w, req := makeRequest(t, http.MethodPost, "/api/v1/business-domains/"+domainID+"/capabilities", body, map[string]string{"id": domainID})
	handler.AssignCapabilityToDomain(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestRemoveCapabilityFromDomain_Integration(t *testing.T) {
	testCtx, cleanup := setupBusinessDomainTestDB(t)
	defer cleanup()

	handler := setupBusinessDomainHandlers(testCtx.db)

	domainID := fmt.Sprintf("test-domain-%d", time.Now().UnixNano())
	capID := fmt.Sprintf("test-cap-%d", time.Now().UnixNano())

	testCtx.createTestDomain(t, domainID, "Test Domain", "Description")
	testCtx.createTestCapability(t, capID, "Test Capability", "L1")

	testCtx.setTenantContext(t)
	assignmentID := fmt.Sprintf("test-assignment-%d", time.Now().UnixNano())
	_, err := testCtx.db.Exec(
		"INSERT INTO domain_capability_assignments (assignment_id, business_domain_id, business_domain_name, capability_id, capability_code, capability_name, capability_level, tenant_id, assigned_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())",
		assignmentID, domainID, "Test Domain", capID, "BIZ-001", "Test Capability", "L1", testTenantID(),
	)
	require.NoError(t, err)

	w, req := makeRequest(t, http.MethodDelete, fmt.Sprintf("/api/v1/business-domains/%s/capabilities/%s", domainID, capID), nil, map[string]string{"domainId": domainID, "capabilityId": capID})
	handler.RemoveCapabilityFromDomain(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	time.Sleep(100 * time.Millisecond)

	testCtx.setTenantContext(t)
	var count int
	err = testCtx.db.QueryRow(
		"SELECT COUNT(*) FROM domain_capability_assignments WHERE assignment_id = $1",
		assignmentID,
	).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestRemoveCapabilityFromDomain_NotFound_Integration(t *testing.T) {
	testCtx, cleanup := setupBusinessDomainTestDB(t)
	defer cleanup()

	handler := setupBusinessDomainHandlers(testCtx.db)

	domainID := fmt.Sprintf("test-domain-%d", time.Now().UnixNano())
	capID := fmt.Sprintf("test-cap-%d", time.Now().UnixNano())

	testCtx.createTestDomain(t, domainID, "Test Domain", "Description")
	testCtx.createTestCapability(t, capID, "Test Capability", "L1")

	w, req := makeRequest(t, http.MethodDelete, fmt.Sprintf("/api/v1/business-domains/%s/capabilities/%s", domainID, capID), nil, map[string]string{"domainId": domainID, "capabilityId": capID})
	handler.RemoveCapabilityFromDomain(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetDomainsForCapability_Integration(t *testing.T) {
	testCtx, cleanup := setupBusinessDomainTestDB(t)
	defer cleanup()

	handler := setupBusinessDomainHandlers(testCtx.db)

	domain1ID := fmt.Sprintf("test-domain-1-%d", time.Now().UnixNano())
	domain2ID := fmt.Sprintf("test-domain-2-%d", time.Now().UnixNano())
	capID := fmt.Sprintf("test-cap-%d", time.Now().UnixNano())

	testCtx.createTestDomain(t, domain1ID, "Domain 1", "Description 1")
	testCtx.createTestDomain(t, domain2ID, "Domain 2", "Description 2")
	testCtx.createTestCapability(t, capID, "Test Capability", "L1")

	testCtx.setTenantContext(t)
	assignment1ID := fmt.Sprintf("test-assignment-1-%d", time.Now().UnixNano())
	_, err := testCtx.db.Exec(
		"INSERT INTO domain_capability_assignments (assignment_id, business_domain_id, business_domain_name, capability_id, capability_code, capability_name, capability_level, tenant_id, assigned_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())",
		assignment1ID, domain1ID, "Domain 1", capID, "BIZ-001", "Test Capability", "L1", testTenantID(),
	)
	require.NoError(t, err)

	assignment2ID := fmt.Sprintf("test-assignment-2-%d", time.Now().UnixNano())
	_, err = testCtx.db.Exec(
		"INSERT INTO domain_capability_assignments (assignment_id, business_domain_id, business_domain_name, capability_id, capability_code, capability_name, capability_level, tenant_id, assigned_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())",
		assignment2ID, domain2ID, "Domain 2", capID, "BIZ-001", "Test Capability", "L1", testTenantID(),
	)
	require.NoError(t, err)

	w, req := makeRequest(t, http.MethodGet, "/api/v1/capabilities/"+capID+"/business-domains", nil, map[string]string{"id": capID})
	handler.GetDomainsForCapability(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response sharedAPI.CollectionResponse
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	dataBytes, _ := json.Marshal(response.Data)
	var domains []DomainForCapabilityDTO
	json.Unmarshal(dataBytes, &domains)

	assert.Equal(t, 2, len(domains))
	assert.NotNil(t, response.Links)
	assert.Contains(t, response.Links, "capability")
}

func TestGetDomainsForCapability_CapabilityNotFound_Integration(t *testing.T) {
	testCtx, cleanup := setupBusinessDomainTestDB(t)
	defer cleanup()

	handler := setupBusinessDomainHandlers(testCtx.db)

	nonExistentID := fmt.Sprintf("non-existent-%d", time.Now().UnixNano())
	w, req := makeRequest(t, http.MethodGet, "/api/v1/capabilities/"+nonExistentID+"/business-domains", nil, map[string]string{"id": nonExistentID})

	handler.GetDomainsForCapability(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetCapabilityRealizationsByDomain_Integration(t *testing.T) {
	testCtx, cleanup := setupBusinessDomainTestDB(t)
	defer cleanup()

	handler := setupBusinessDomainHandlers(testCtx.db)

	domainID := fmt.Sprintf("test-domain-%d", time.Now().UnixNano())
	capL1ID := fmt.Sprintf("test-cap-l1-%d", time.Now().UnixNano())
	capL2ID := fmt.Sprintf("test-cap-l2-%d", time.Now().UnixNano())
	componentID := fmt.Sprintf("test-comp-%d", time.Now().UnixNano())
	realizationID := fmt.Sprintf("test-real-%d", time.Now().UnixNano())

	testCtx.createTestDomain(t, domainID, "Test Domain", "Description")
	testCtx.createTestCapability(t, capL1ID, "L1 Capability", "L1")
	testCtx.createTestCapabilityWithParent(t, capL2ID, "L2 Capability", "L2", capL1ID)
	testCtx.createTestComponent(t, componentID, "Test System")
	testCtx.createTestRealization(t, realizationID, capL2ID, componentID, "Test System", "Full", "Direct")

	testCtx.setTenantContext(t)
	assignmentID := fmt.Sprintf("test-assignment-%d", time.Now().UnixNano())
	_, err := testCtx.db.Exec(
		"INSERT INTO domain_capability_assignments (assignment_id, business_domain_id, business_domain_name, capability_id, capability_name, capability_level, tenant_id, assigned_at) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())",
		assignmentID, domainID, "Test Domain", capL1ID, "L1 Capability", "L1", testTenantID(),
	)
	require.NoError(t, err)

	w, req := makeRequest(t, http.MethodGet, "/api/v1/business-domains/"+domainID+"/capability-realizations?depth=2", nil, map[string]string{"id": domainID})
	handler.GetCapabilityRealizationsByDomain(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response sharedAPI.CollectionResponse
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	dataBytes, _ := json.Marshal(response.Data)
	var groups []CapabilityRealizationsGroupDTO
	json.Unmarshal(dataBytes, &groups)

	assert.GreaterOrEqual(t, len(groups), 1)

	var foundL2 bool
	for _, group := range groups {
		if group.CapabilityID == capL2ID {
			foundL2 = true
			assert.Equal(t, "L2", group.Level)
			assert.Equal(t, 1, len(group.Realizations))
			if len(group.Realizations) > 0 {
				assert.Equal(t, componentID, group.Realizations[0].ComponentID)
			}
		}
	}
	assert.True(t, foundL2, "L2 capability should be included in results")
}

func TestGetCapabilityRealizationsByDomain_EmptyDomain_Integration(t *testing.T) {
	testCtx, cleanup := setupBusinessDomainTestDB(t)
	defer cleanup()

	handler := setupBusinessDomainHandlers(testCtx.db)

	domainID := fmt.Sprintf("test-domain-%d", time.Now().UnixNano())
	testCtx.createTestDomain(t, domainID, "Empty Domain", "No capabilities assigned")

	w, req := makeRequest(t, http.MethodGet, "/api/v1/business-domains/"+domainID+"/capability-realizations", nil, map[string]string{"id": domainID})
	handler.GetCapabilityRealizationsByDomain(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response sharedAPI.CollectionResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	dataBytes, _ := json.Marshal(response.Data)
	var groups []CapabilityRealizationsGroupDTO
	json.Unmarshal(dataBytes, &groups)

	assert.Equal(t, 0, len(groups))
}

func TestGetCapabilityRealizationsByDomain_DomainNotFound_Integration(t *testing.T) {
	testCtx, cleanup := setupBusinessDomainTestDB(t)
	defer cleanup()

	handler := setupBusinessDomainHandlers(testCtx.db)

	nonExistentID := fmt.Sprintf("non-existent-%d", time.Now().UnixNano())
	w, req := makeRequest(t, http.MethodGet, "/api/v1/business-domains/"+nonExistentID+"/capability-realizations", nil, map[string]string{"id": nonExistentID})

	handler.GetCapabilityRealizationsByDomain(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
