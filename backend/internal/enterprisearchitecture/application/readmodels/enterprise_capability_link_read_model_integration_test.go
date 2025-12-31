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
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"

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

func TestCheckHierarchyConflict_AncestorLinkedToDifferent(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	tenantDB := database.NewTenantAwareDB(db)
	readModel := NewEnterpriseCapabilityLinkReadModel(tenantDB)
	ctx := tenantContext()

	setTenantContext(t, db)

	grandparentID := fmt.Sprintf("cap-gp-%d", time.Now().UnixNano())
	parentID := fmt.Sprintf("cap-p-%d", time.Now().UnixNano())
	childID := fmt.Sprintf("cap-c-%d", time.Now().UnixNano())
	enterpriseCapID1 := fmt.Sprintf("ec-1-%d", time.Now().UnixNano())
	enterpriseCapID2 := fmt.Sprintf("ec-2-%d", time.Now().UnixNano())

	_, err := db.Exec(`
		INSERT INTO capabilities (id, tenant_id, name, level, parent_id, status, created_at)
		VALUES
			($1, 'default', 'Grandparent', 'L1', NULL, 'active', NOW()),
			($2, 'default', 'Parent', 'L1', $1, 'active', NOW()),
			($3, 'default', 'Child', 'L1', $2, 'active', NOW())
	`, grandparentID, parentID, childID)
	require.NoError(t, err)

	_, err = db.Exec(`
		INSERT INTO enterprise_capabilities (id, tenant_id, name, description, category, is_active, created_at)
		VALUES
			($1, 'default', 'EC1', 'EC1 desc', 'Test', true, NOW()),
			($2, 'default', 'EC2', 'EC2 desc', 'Test', true, NOW())
	`, enterpriseCapID1, enterpriseCapID2)
	require.NoError(t, err)

	linkID := fmt.Sprintf("link-%d", time.Now().UnixNano())
	_, err = db.Exec(`
		INSERT INTO enterprise_capability_links (id, tenant_id, enterprise_capability_id, domain_capability_id, linked_by, linked_at)
		VALUES ($1, 'default', $2, $3, 'test@example.com', NOW())
	`, linkID, enterpriseCapID1, parentID)
	require.NoError(t, err)

	conflict, err := readModel.CheckHierarchyConflict(ctx, childID, enterpriseCapID2)
	require.NoError(t, err)
	require.NotNil(t, conflict)

	assert.Equal(t, parentID, conflict.ConflictingCapabilityID)
	assert.Equal(t, "Parent", conflict.ConflictingCapabilityName)
	assert.Equal(t, enterpriseCapID1, conflict.LinkedToCapabilityID)
	assert.Equal(t, "EC1", conflict.LinkedToCapabilityName)
	assert.True(t, conflict.IsAncestor)

	_, err = db.Exec("DELETE FROM enterprise_capability_links WHERE id = $1", linkID)
	require.NoError(t, err)
	_, err = db.Exec("DELETE FROM enterprise_capabilities WHERE id IN ($1, $2)", enterpriseCapID1, enterpriseCapID2)
	require.NoError(t, err)
	_, err = db.Exec("DELETE FROM capabilities WHERE id IN ($1, $2, $3)", grandparentID, parentID, childID)
	require.NoError(t, err)
}

func TestCheckHierarchyConflict_DescendantLinkedToDifferent(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	tenantDB := database.NewTenantAwareDB(db)
	readModel := NewEnterpriseCapabilityLinkReadModel(tenantDB)
	ctx := tenantContext()

	setTenantContext(t, db)

	parentID := fmt.Sprintf("cap-p-%d", time.Now().UnixNano())
	childID := fmt.Sprintf("cap-c-%d", time.Now().UnixNano())
	enterpriseCapID1 := fmt.Sprintf("ec-1-%d", time.Now().UnixNano())
	enterpriseCapID2 := fmt.Sprintf("ec-2-%d", time.Now().UnixNano())

	_, err := db.Exec(`
		INSERT INTO capabilities (id, tenant_id, name, level, parent_id, status, created_at)
		VALUES
			($1, 'default', 'Parent', 'L1', NULL, 'active', NOW()),
			($2, 'default', 'Child', 'L1', $1, 'active', NOW())
	`, parentID, childID)
	require.NoError(t, err)

	_, err = db.Exec(`
		INSERT INTO enterprise_capabilities (id, tenant_id, name, description, category, is_active, created_at)
		VALUES
			($1, 'default', 'EC1', 'EC1 desc', 'Test', true, NOW()),
			($2, 'default', 'EC2', 'EC2 desc', 'Test', true, NOW())
	`, enterpriseCapID1, enterpriseCapID2)
	require.NoError(t, err)

	linkID := fmt.Sprintf("link-%d", time.Now().UnixNano())
	_, err = db.Exec(`
		INSERT INTO enterprise_capability_links (id, tenant_id, enterprise_capability_id, domain_capability_id, linked_by, linked_at)
		VALUES ($1, 'default', $2, $3, 'test@example.com', NOW())
	`, linkID, enterpriseCapID1, childID)
	require.NoError(t, err)

	conflict, err := readModel.CheckHierarchyConflict(ctx, parentID, enterpriseCapID2)
	require.NoError(t, err)
	require.NotNil(t, conflict)

	assert.Equal(t, childID, conflict.ConflictingCapabilityID)
	assert.Equal(t, "Child", conflict.ConflictingCapabilityName)
	assert.Equal(t, enterpriseCapID1, conflict.LinkedToCapabilityID)
	assert.Equal(t, "EC1", conflict.LinkedToCapabilityName)
	assert.False(t, conflict.IsAncestor)

	_, err = db.Exec("DELETE FROM enterprise_capability_links WHERE id = $1", linkID)
	require.NoError(t, err)
	_, err = db.Exec("DELETE FROM enterprise_capabilities WHERE id IN ($1, $2)", enterpriseCapID1, enterpriseCapID2)
	require.NoError(t, err)
	_, err = db.Exec("DELETE FROM capabilities WHERE id IN ($1, $2)", parentID, childID)
	require.NoError(t, err)
}

func TestCheckHierarchyConflict_SameEnterpriseCapability_NoConflict(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	tenantDB := database.NewTenantAwareDB(db)
	readModel := NewEnterpriseCapabilityLinkReadModel(tenantDB)
	ctx := tenantContext()

	setTenantContext(t, db)

	parentID := fmt.Sprintf("cap-p-%d", time.Now().UnixNano())
	childID := fmt.Sprintf("cap-c-%d", time.Now().UnixNano())
	enterpriseCapID := fmt.Sprintf("ec-%d", time.Now().UnixNano())

	_, err := db.Exec(`
		INSERT INTO capabilities (id, tenant_id, name, level, parent_id, status, created_at)
		VALUES
			($1, 'default', 'Parent', 'L1', NULL, 'active', NOW()),
			($2, 'default', 'Child', 'L1', $1, 'active', NOW())
	`, parentID, childID)
	require.NoError(t, err)

	_, err = db.Exec(`
		INSERT INTO enterprise_capabilities (id, tenant_id, name, description, category, is_active, created_at)
		VALUES ($1, 'default', 'EC', 'EC desc', 'Test', true, NOW())
	`, enterpriseCapID)
	require.NoError(t, err)

	linkID := fmt.Sprintf("link-%d", time.Now().UnixNano())
	_, err = db.Exec(`
		INSERT INTO enterprise_capability_links (id, tenant_id, enterprise_capability_id, domain_capability_id, linked_by, linked_at)
		VALUES ($1, 'default', $2, $3, 'test@example.com', NOW())
	`, linkID, enterpriseCapID, childID)
	require.NoError(t, err)

	conflict, err := readModel.CheckHierarchyConflict(ctx, parentID, enterpriseCapID)
	require.NoError(t, err)
	assert.Nil(t, conflict)

	_, err = db.Exec("DELETE FROM enterprise_capability_links WHERE id = $1", linkID)
	require.NoError(t, err)
	_, err = db.Exec("DELETE FROM enterprise_capabilities WHERE id = $1", enterpriseCapID)
	require.NoError(t, err)
	_, err = db.Exec("DELETE FROM capabilities WHERE id IN ($1, $2)", parentID, childID)
	require.NoError(t, err)
}

func TestCheckHierarchyConflict_DeepHierarchy_HandlesDepthLimit(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	tenantDB := database.NewTenantAwareDB(db)
	readModel := NewEnterpriseCapabilityLinkReadModel(tenantDB)
	ctx := tenantContext()

	setTenantContext(t, db)

	capabilityIDs := make([]string, 12)
	for i := 0; i < 12; i++ {
		capabilityIDs[i] = fmt.Sprintf("cap-deep-%d-%d", i, time.Now().UnixNano())
	}

	enterpriseCapID1 := fmt.Sprintf("ec-1-%d", time.Now().UnixNano())
	enterpriseCapID2 := fmt.Sprintf("ec-2-%d", time.Now().UnixNano())

	_, err := db.Exec(`
		INSERT INTO enterprise_capabilities (id, tenant_id, name, description, category, is_active, created_at)
		VALUES
			($1, 'default', 'EC1', 'EC1 desc', 'Test', true, NOW()),
			($2, 'default', 'EC2', 'EC2 desc', 'Test', true, NOW())
	`, enterpriseCapID1, enterpriseCapID2)
	require.NoError(t, err)

	var parentID *string
	for i := 0; i < 12; i++ {
		if i == 0 {
			_, err = db.Exec(`
				INSERT INTO capabilities (id, tenant_id, name, level, parent_id, status, created_at)
				VALUES ($1, 'default', $2, 'L1', NULL, 'active', NOW())
			`, capabilityIDs[i], fmt.Sprintf("Level %d", i))
		} else {
			_, err = db.Exec(`
				INSERT INTO capabilities (id, tenant_id, name, level, parent_id, status, created_at)
				VALUES ($1, 'default', $2, 'L1', $3, 'active', NOW())
			`, capabilityIDs[i], fmt.Sprintf("Level %d", i), *parentID)
		}
		require.NoError(t, err)
		parentID = &capabilityIDs[i]
	}

	linkID := fmt.Sprintf("link-%d", time.Now().UnixNano())
	_, err = db.Exec(`
		INSERT INTO enterprise_capability_links (id, tenant_id, enterprise_capability_id, domain_capability_id, linked_by, linked_at)
		VALUES ($1, 'default', $2, $3, 'test@example.com', NOW())
	`, linkID, enterpriseCapID1, capabilityIDs[8])
	require.NoError(t, err)

	conflict, err := readModel.CheckHierarchyConflict(ctx, capabilityIDs[11], enterpriseCapID2)
	require.NoError(t, err)
	require.NotNil(t, conflict)
	assert.True(t, conflict.IsAncestor)

	conflict, err = readModel.CheckHierarchyConflict(ctx, capabilityIDs[0], enterpriseCapID2)
	require.NoError(t, err)
	require.NotNil(t, conflict)
	assert.False(t, conflict.IsAncestor)

	_, err = db.Exec("DELETE FROM enterprise_capability_links WHERE id = $1", linkID)
	require.NoError(t, err)
	_, err = db.Exec("DELETE FROM enterprise_capabilities WHERE id IN ($1, $2)", enterpriseCapID1, enterpriseCapID2)
	require.NoError(t, err)
	for i := len(capabilityIDs) - 1; i >= 0; i-- {
		_, err = db.Exec("DELETE FROM capabilities WHERE id = $1", capabilityIDs[i])
		require.NoError(t, err)
	}
}

func TestCheckHierarchyConflict_NoRelationship_NoConflict(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	tenantDB := database.NewTenantAwareDB(db)
	readModel := NewEnterpriseCapabilityLinkReadModel(tenantDB)
	ctx := tenantContext()

	setTenantContext(t, db)

	cap1ID := fmt.Sprintf("cap-1-%d", time.Now().UnixNano())
	cap2ID := fmt.Sprintf("cap-2-%d", time.Now().UnixNano())
	enterpriseCapID1 := fmt.Sprintf("ec-1-%d", time.Now().UnixNano())
	enterpriseCapID2 := fmt.Sprintf("ec-2-%d", time.Now().UnixNano())

	_, err := db.Exec(`
		INSERT INTO capabilities (id, tenant_id, name, level, parent_id, status, created_at)
		VALUES
			($1, 'default', 'Cap1', 'L1', NULL, 'active', NOW()),
			($2, 'default', 'Cap2', 'L1', NULL, 'active', NOW())
	`, cap1ID, cap2ID)
	require.NoError(t, err)

	_, err = db.Exec(`
		INSERT INTO enterprise_capabilities (id, tenant_id, name, description, category, is_active, created_at)
		VALUES
			($1, 'default', 'EC1', 'EC1 desc', 'Test', true, NOW()),
			($2, 'default', 'EC2', 'EC2 desc', 'Test', true, NOW())
	`, enterpriseCapID1, enterpriseCapID2)
	require.NoError(t, err)

	linkID := fmt.Sprintf("link-%d", time.Now().UnixNano())
	_, err = db.Exec(`
		INSERT INTO enterprise_capability_links (id, tenant_id, enterprise_capability_id, domain_capability_id, linked_by, linked_at)
		VALUES ($1, 'default', $2, $3, 'test@example.com', NOW())
	`, linkID, enterpriseCapID1, cap2ID)
	require.NoError(t, err)

	conflict, err := readModel.CheckHierarchyConflict(ctx, cap1ID, enterpriseCapID2)
	require.NoError(t, err)
	assert.Nil(t, conflict)

	_, err = db.Exec("DELETE FROM enterprise_capability_links WHERE id = $1", linkID)
	require.NoError(t, err)
	_, err = db.Exec("DELETE FROM enterprise_capabilities WHERE id IN ($1, $2)", enterpriseCapID1, enterpriseCapID2)
	require.NoError(t, err)
	_, err = db.Exec("DELETE FROM capabilities WHERE id IN ($1, $2)", cap1ID, cap2ID)
	require.NoError(t, err)
}
