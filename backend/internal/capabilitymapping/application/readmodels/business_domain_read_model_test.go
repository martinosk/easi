//go:build integration

package readmodels

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"easi/backend/internal/infrastructure/database"
	sharedcontext "easi/backend/internal/shared/context"
	sharedvo "easi/backend/internal/shared/domain/valueobjects"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) (*sql.DB, func()) {
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

	cleanup := func() {
		db.Close()
	}

	return db, cleanup
}

func getEnv(key, defaultValue string) string {
	return defaultValue
}

func tenantContext() context.Context {
	return sharedcontext.WithTenant(context.Background(), sharedvo.DefaultTenantID())
}

func setTenantContext(t *testing.T, db *sql.DB) {
	_, err := db.Exec("SET app.current_tenant = 'default'")
	require.NoError(t, err)
}

func TestBusinessDomainReadModel_Insert(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	tenantDB := database.NewTenantAwareDB(db)
	readModel := NewBusinessDomainReadModel(tenantDB)

	domainID := fmt.Sprintf("bd-test-%d", time.Now().UnixNano())
	dto := BusinessDomainDTO{
		ID:          domainID,
		Name:        "Finance",
		Description: "Financial operations and planning",
		CreatedAt:   time.Now().UTC(),
	}

	ctx := tenantContext()
	err := readModel.Insert(ctx, dto)
	require.NoError(t, err)

	setTenantContext(t, db)
	var name, description string
	var createdAt time.Time
	err = db.QueryRow(
		"SELECT name, description, created_at FROM business_domains WHERE id = $1",
		domainID,
	).Scan(&name, &description, &createdAt)
	require.NoError(t, err)

	assert.Equal(t, "Finance", name)
	assert.Equal(t, "Financial operations and planning", description)
	assert.WithinDuration(t, dto.CreatedAt, createdAt, time.Second)

	_, err = db.Exec("DELETE FROM business_domains WHERE id = $1", domainID)
	require.NoError(t, err)
}

func TestBusinessDomainReadModel_Update(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	tenantDB := database.NewTenantAwareDB(db)
	readModel := NewBusinessDomainReadModel(tenantDB)

	domainID := fmt.Sprintf("bd-test-%d", time.Now().UnixNano())
	setTenantContext(t, db)
	_, err := db.Exec(
		"INSERT INTO business_domains (id, tenant_id, name, description, created_at) VALUES ($1, $2, $3, $4, $5)",
		domainID, "default", "Original Name", "Original Description", time.Now().UTC(),
	)
	require.NoError(t, err)

	ctx := tenantContext()
	err = readModel.Update(ctx, domainID, "Updated Name", "Updated Description")
	require.NoError(t, err)

	setTenantContext(t, db)
	var name, description string
	err = db.QueryRow(
		"SELECT name, description FROM business_domains WHERE id = $1",
		domainID,
	).Scan(&name, &description)
	require.NoError(t, err)

	assert.Equal(t, "Updated Name", name)
	assert.Equal(t, "Updated Description", description)

	_, err = db.Exec("DELETE FROM business_domains WHERE id = $1", domainID)
	require.NoError(t, err)
}

func TestBusinessDomainReadModel_Delete(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	tenantDB := database.NewTenantAwareDB(db)
	readModel := NewBusinessDomainReadModel(tenantDB)

	domainID := fmt.Sprintf("bd-test-%d", time.Now().UnixNano())
	setTenantContext(t, db)
	_, err := db.Exec(
		"INSERT INTO business_domains (id, tenant_id, name, description, created_at) VALUES ($1, $2, $3, $4, $5)",
		domainID, "default", "To Delete", "Will be removed", time.Now().UTC(),
	)
	require.NoError(t, err)

	ctx := tenantContext()
	err = readModel.Delete(ctx, domainID)
	require.NoError(t, err)

	setTenantContext(t, db)
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM business_domains WHERE id = $1", domainID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestBusinessDomainReadModel_IncrementCapabilityCount(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	tenantDB := database.NewTenantAwareDB(db)
	readModel := NewBusinessDomainReadModel(tenantDB)

	domainID := fmt.Sprintf("bd-test-%d", time.Now().UnixNano())
	setTenantContext(t, db)
	_, err := db.Exec(
		"INSERT INTO business_domains (id, tenant_id, name, description, capability_count, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
		domainID, "default", "Test Domain", "Description", 0, time.Now().UTC(),
	)
	require.NoError(t, err)

	ctx := tenantContext()
	err = readModel.IncrementCapabilityCount(ctx, domainID)
	require.NoError(t, err)

	setTenantContext(t, db)
	var count int
	err = db.QueryRow("SELECT capability_count FROM business_domains WHERE id = $1", domainID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	err = readModel.IncrementCapabilityCount(ctx, domainID)
	require.NoError(t, err)

	setTenantContext(t, db)
	err = db.QueryRow("SELECT capability_count FROM business_domains WHERE id = $1", domainID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 2, count)

	_, err = db.Exec("DELETE FROM business_domains WHERE id = $1", domainID)
	require.NoError(t, err)
}

func TestBusinessDomainReadModel_DecrementCapabilityCount(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	tenantDB := database.NewTenantAwareDB(db)
	readModel := NewBusinessDomainReadModel(tenantDB)

	domainID := fmt.Sprintf("bd-test-%d", time.Now().UnixNano())
	setTenantContext(t, db)
	_, err := db.Exec(
		"INSERT INTO business_domains (id, tenant_id, name, description, capability_count, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
		domainID, "default", "Test Domain", "Description", 2, time.Now().UTC(),
	)
	require.NoError(t, err)

	ctx := tenantContext()
	err = readModel.DecrementCapabilityCount(ctx, domainID)
	require.NoError(t, err)

	setTenantContext(t, db)
	var count int
	err = db.QueryRow("SELECT capability_count FROM business_domains WHERE id = $1", domainID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	_, err = db.Exec("DELETE FROM business_domains WHERE id = $1", domainID)
	require.NoError(t, err)
}

func TestBusinessDomainReadModel_GetAll(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	tenantDB := database.NewTenantAwareDB(db)
	readModel := NewBusinessDomainReadModel(tenantDB)

	domain1ID := fmt.Sprintf("bd-test-1-%d", time.Now().UnixNano())
	domain2ID := fmt.Sprintf("bd-test-2-%d", time.Now().UnixNano())
	setTenantContext(t, db)
	_, err := db.Exec(
		"INSERT INTO business_domains (id, tenant_id, name, description, capability_count, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
		domain1ID, "default", "Finance", "Finance domain", 3, time.Now().UTC(),
	)
	require.NoError(t, err)
	_, err = db.Exec(
		"INSERT INTO business_domains (id, tenant_id, name, description, capability_count, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
		domain2ID, "default", "Operations", "Operations domain", 5, time.Now().UTC(),
	)
	require.NoError(t, err)

	ctx := tenantContext()
	domains, err := readModel.GetAll(ctx)
	require.NoError(t, err)

	var found1, found2 bool
	for _, d := range domains {
		if d.ID == domain1ID {
			found1 = true
			assert.Equal(t, "Finance", d.Name)
			assert.Equal(t, 3, d.CapabilityCount)
		}
		if d.ID == domain2ID {
			found2 = true
			assert.Equal(t, "Operations", d.Name)
			assert.Equal(t, 5, d.CapabilityCount)
		}
	}
	assert.True(t, found1, "Domain 1 not found in results")
	assert.True(t, found2, "Domain 2 not found in results")

	_, err = db.Exec("DELETE FROM business_domains WHERE id IN ($1, $2)", domain1ID, domain2ID)
	require.NoError(t, err)
}

func TestBusinessDomainReadModel_GetByID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	tenantDB := database.NewTenantAwareDB(db)
	readModel := NewBusinessDomainReadModel(tenantDB)

	domainID := fmt.Sprintf("bd-test-%d", time.Now().UnixNano())
	setTenantContext(t, db)
	_, err := db.Exec(
		"INSERT INTO business_domains (id, tenant_id, name, description, capability_count, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
		domainID, "default", "Customer Experience", "CX domain", 7, time.Now().UTC(),
	)
	require.NoError(t, err)

	ctx := tenantContext()
	domain, err := readModel.GetByID(ctx, domainID)
	require.NoError(t, err)
	require.NotNil(t, domain)

	assert.Equal(t, domainID, domain.ID)
	assert.Equal(t, "Customer Experience", domain.Name)
	assert.Equal(t, "CX domain", domain.Description)
	assert.Equal(t, 7, domain.CapabilityCount)

	_, err = db.Exec("DELETE FROM business_domains WHERE id = $1", domainID)
	require.NoError(t, err)
}

func TestBusinessDomainReadModel_GetByID_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	tenantDB := database.NewTenantAwareDB(db)
	readModel := NewBusinessDomainReadModel(tenantDB)

	ctx := tenantContext()
	domain, err := readModel.GetByID(ctx, "bd-nonexistent")
	require.NoError(t, err)
	assert.Nil(t, domain)
}

func TestBusinessDomainReadModel_GetByName(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	tenantDB := database.NewTenantAwareDB(db)
	readModel := NewBusinessDomainReadModel(tenantDB)

	domainID := fmt.Sprintf("bd-test-%d", time.Now().UnixNano())
	setTenantContext(t, db)
	_, err := db.Exec(
		"INSERT INTO business_domains (id, tenant_id, name, description, capability_count, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
		domainID, "default", "Technology", "Tech domain", 4, time.Now().UTC(),
	)
	require.NoError(t, err)

	ctx := tenantContext()
	domain, err := readModel.GetByName(ctx, "Technology")
	require.NoError(t, err)
	require.NotNil(t, domain)

	assert.Equal(t, domainID, domain.ID)
	assert.Equal(t, "Technology", domain.Name)

	_, err = db.Exec("DELETE FROM business_domains WHERE id = $1", domainID)
	require.NoError(t, err)
}

func TestBusinessDomainReadModel_GetByName_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	tenantDB := database.NewTenantAwareDB(db)
	readModel := NewBusinessDomainReadModel(tenantDB)

	ctx := tenantContext()
	domain, err := readModel.GetByName(ctx, "NonExistentDomain")
	require.NoError(t, err)
	assert.Nil(t, domain)
}

func TestBusinessDomainReadModel_NameExists(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	tenantDB := database.NewTenantAwareDB(db)
	readModel := NewBusinessDomainReadModel(tenantDB)

	domainID := fmt.Sprintf("bd-test-%d", time.Now().UnixNano())
	setTenantContext(t, db)
	_, err := db.Exec(
		"INSERT INTO business_domains (id, tenant_id, name, description, created_at) VALUES ($1, $2, $3, $4, $5)",
		domainID, "default", "Marketing", "Marketing domain", time.Now().UTC(),
	)
	require.NoError(t, err)

	ctx := tenantContext()
	exists, err := readModel.NameExists(ctx, "Marketing", "")
	require.NoError(t, err)
	assert.True(t, exists)

	exists, err = readModel.NameExists(ctx, "NonExistent", "")
	require.NoError(t, err)
	assert.False(t, exists)

	exists, err = readModel.NameExists(ctx, "Marketing", domainID)
	require.NoError(t, err)
	assert.False(t, exists, "Should exclude self when checking name uniqueness")

	_, err = db.Exec("DELETE FROM business_domains WHERE id = $1", domainID)
	require.NoError(t, err)
}
