//go:build integration

package readmodels

import (
	"fmt"
	"testing"
	"time"

	"easi/backend/internal/infrastructure/database"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDomainCapabilityAssignmentReadModel_Insert(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	tenantDB := database.NewTenantAwareDB(db)
	readModel := NewDomainCapabilityAssignmentReadModel(tenantDB)

	assignmentID := fmt.Sprintf("assign-test-%d", time.Now().UnixNano())
	domainID := fmt.Sprintf("bd-test-%d", time.Now().UnixNano())
	capabilityID := fmt.Sprintf("cap-test-%d", time.Now().UnixNano())

	dto := AssignmentDTO{
		AssignmentID:       assignmentID,
		BusinessDomainID:   domainID,
		BusinessDomainName: "Finance",
		CapabilityID:       capabilityID,
		CapabilityCode:     "FIN-01",
		CapabilityName:     "Financial Reporting",
		CapabilityLevel:    "L1",
		AssignedAt:         time.Now().UTC(),
	}

	ctx := tenantContext()
	err := readModel.Insert(ctx, dto)
	require.NoError(t, err)

	setTenantContext(t, db)
	var businessDomainID, capID string
	err = db.QueryRow(
		"SELECT business_domain_id, capability_id FROM domain_capability_assignments WHERE assignment_id = $1",
		assignmentID,
	).Scan(&businessDomainID, &capID)
	require.NoError(t, err)

	assert.Equal(t, domainID, businessDomainID)
	assert.Equal(t, capabilityID, capID)

	_, err = db.Exec("DELETE FROM domain_capability_assignments WHERE assignment_id = $1", assignmentID)
	require.NoError(t, err)
}

func TestDomainCapabilityAssignmentReadModel_Delete(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	tenantDB := database.NewTenantAwareDB(db)
	readModel := NewDomainCapabilityAssignmentReadModel(tenantDB)

	assignmentID := fmt.Sprintf("assign-test-%d", time.Now().UnixNano())
	domainID := fmt.Sprintf("bd-test-%d", time.Now().UnixNano())
	capabilityID := fmt.Sprintf("cap-test-%d", time.Now().UnixNano())

	setTenantContext(t, db)
	_, err := db.Exec(
		"INSERT INTO domain_capability_assignments (assignment_id, tenant_id, business_domain_id, business_domain_name, capability_id, capability_code, capability_name, capability_level, assigned_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)",
		assignmentID, "default", domainID, "Finance", capabilityID, "FIN-01", "Financial Reporting", "L1", time.Now().UTC(),
	)
	require.NoError(t, err)

	ctx := tenantContext()
	err = readModel.Delete(ctx, assignmentID)
	require.NoError(t, err)

	setTenantContext(t, db)
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM domain_capability_assignments WHERE assignment_id = $1", assignmentID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestDomainCapabilityAssignmentReadModel_GetByDomainID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	tenantDB := database.NewTenantAwareDB(db)
	readModel := NewDomainCapabilityAssignmentReadModel(tenantDB)

	domainID := fmt.Sprintf("bd-test-%d", time.Now().UnixNano())
	assignment1ID := fmt.Sprintf("assign-test-1-%d", time.Now().UnixNano())
	assignment2ID := fmt.Sprintf("assign-test-2-%d", time.Now().UnixNano())
	cap1ID := fmt.Sprintf("cap-test-1-%d", time.Now().UnixNano())
	cap2ID := fmt.Sprintf("cap-test-2-%d", time.Now().UnixNano())

	setTenantContext(t, db)
	_, err := db.Exec(
		"INSERT INTO domain_capability_assignments (assignment_id, tenant_id, business_domain_id, business_domain_name, capability_id, capability_code, capability_name, capability_level, assigned_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)",
		assignment1ID, "default", domainID, "Finance", cap1ID, "FIN-01", "Financial Reporting", "L1", time.Now().UTC(),
	)
	require.NoError(t, err)
	_, err = db.Exec(
		"INSERT INTO domain_capability_assignments (assignment_id, tenant_id, business_domain_id, business_domain_name, capability_id, capability_code, capability_name, capability_level, assigned_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)",
		assignment2ID, "default", domainID, "Finance", cap2ID, "FIN-02", "Budget Planning", "L1", time.Now().UTC(),
	)
	require.NoError(t, err)

	ctx := tenantContext()
	assignments, err := readModel.GetByDomainID(ctx, domainID)
	require.NoError(t, err)

	assert.Len(t, assignments, 2)
	found1 := false
	found2 := false
	for _, a := range assignments {
		if a.AssignmentID == assignment1ID {
			found1 = true
			assert.Equal(t, "Financial Reporting", a.CapabilityName)
		}
		if a.AssignmentID == assignment2ID {
			found2 = true
			assert.Equal(t, "Budget Planning", a.CapabilityName)
		}
	}
	assert.True(t, found1, "Assignment 1 not found")
	assert.True(t, found2, "Assignment 2 not found")

	_, err = db.Exec("DELETE FROM domain_capability_assignments WHERE assignment_id IN ($1, $2)", assignment1ID, assignment2ID)
	require.NoError(t, err)
}

func TestDomainCapabilityAssignmentReadModel_GetByCapabilityID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	tenantDB := database.NewTenantAwareDB(db)
	readModel := NewDomainCapabilityAssignmentReadModel(tenantDB)

	capabilityID := fmt.Sprintf("cap-test-%d", time.Now().UnixNano())
	assignment1ID := fmt.Sprintf("assign-test-1-%d", time.Now().UnixNano())
	assignment2ID := fmt.Sprintf("assign-test-2-%d", time.Now().UnixNano())
	domain1ID := fmt.Sprintf("bd-test-1-%d", time.Now().UnixNano())
	domain2ID := fmt.Sprintf("bd-test-2-%d", time.Now().UnixNano())

	setTenantContext(t, db)
	_, err := db.Exec(
		"INSERT INTO domain_capability_assignments (assignment_id, tenant_id, business_domain_id, business_domain_name, capability_id, capability_code, capability_name, capability_level, assigned_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)",
		assignment1ID, "default", domain1ID, "Finance", capabilityID, "FIN-01", "Financial Reporting", "L1", time.Now().UTC(),
	)
	require.NoError(t, err)
	_, err = db.Exec(
		"INSERT INTO domain_capability_assignments (assignment_id, tenant_id, business_domain_id, business_domain_name, capability_id, capability_code, capability_name, capability_level, assigned_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)",
		assignment2ID, "default", domain2ID, "Operations", capabilityID, "FIN-01", "Financial Reporting", "L1", time.Now().UTC(),
	)
	require.NoError(t, err)

	ctx := tenantContext()
	assignments, err := readModel.GetByCapabilityID(ctx, capabilityID)
	require.NoError(t, err)

	assert.Len(t, assignments, 2)
	found1 := false
	found2 := false
	for _, a := range assignments {
		if a.BusinessDomainID == domain1ID {
			found1 = true
			assert.Equal(t, "Finance", a.BusinessDomainName)
		}
		if a.BusinessDomainID == domain2ID {
			found2 = true
			assert.Equal(t, "Operations", a.BusinessDomainName)
		}
	}
	assert.True(t, found1, "Domain 1 assignment not found")
	assert.True(t, found2, "Domain 2 assignment not found")

	_, err = db.Exec("DELETE FROM domain_capability_assignments WHERE assignment_id IN ($1, $2)", assignment1ID, assignment2ID)
	require.NoError(t, err)
}

func TestDomainCapabilityAssignmentReadModel_GetByDomainAndCapability(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	tenantDB := database.NewTenantAwareDB(db)
	readModel := NewDomainCapabilityAssignmentReadModel(tenantDB)

	assignmentID := fmt.Sprintf("assign-test-%d", time.Now().UnixNano())
	domainID := fmt.Sprintf("bd-test-%d", time.Now().UnixNano())
	capabilityID := fmt.Sprintf("cap-test-%d", time.Now().UnixNano())

	setTenantContext(t, db)
	_, err := db.Exec(
		"INSERT INTO domain_capability_assignments (assignment_id, tenant_id, business_domain_id, business_domain_name, capability_id, capability_code, capability_name, capability_level, assigned_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)",
		assignmentID, "default", domainID, "Finance", capabilityID, "FIN-01", "Financial Reporting", "L1", time.Now().UTC(),
	)
	require.NoError(t, err)

	ctx := tenantContext()
	assignment, err := readModel.GetByDomainAndCapability(ctx, domainID, capabilityID)
	require.NoError(t, err)
	require.NotNil(t, assignment)

	assert.Equal(t, assignmentID, assignment.AssignmentID)
	assert.Equal(t, domainID, assignment.BusinessDomainID)
	assert.Equal(t, capabilityID, assignment.CapabilityID)

	_, err = db.Exec("DELETE FROM domain_capability_assignments WHERE assignment_id = $1", assignmentID)
	require.NoError(t, err)
}

func TestDomainCapabilityAssignmentReadModel_GetByDomainAndCapability_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	tenantDB := database.NewTenantAwareDB(db)
	readModel := NewDomainCapabilityAssignmentReadModel(tenantDB)

	ctx := tenantContext()
	assignment, err := readModel.GetByDomainAndCapability(ctx, "bd-nonexistent", "cap-nonexistent")
	require.NoError(t, err)
	assert.Nil(t, assignment)
}

func TestDomainCapabilityAssignmentReadModel_AssignmentExists(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	tenantDB := database.NewTenantAwareDB(db)
	readModel := NewDomainCapabilityAssignmentReadModel(tenantDB)

	assignmentID := fmt.Sprintf("assign-test-%d", time.Now().UnixNano())
	domainID := fmt.Sprintf("bd-test-%d", time.Now().UnixNano())
	capabilityID := fmt.Sprintf("cap-test-%d", time.Now().UnixNano())

	setTenantContext(t, db)
	_, err := db.Exec(
		"INSERT INTO domain_capability_assignments (assignment_id, tenant_id, business_domain_id, business_domain_name, capability_id, capability_code, capability_name, capability_level, assigned_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)",
		assignmentID, "default", domainID, "Finance", capabilityID, "FIN-01", "Financial Reporting", "L1", time.Now().UTC(),
	)
	require.NoError(t, err)

	ctx := tenantContext()
	exists, err := readModel.AssignmentExists(ctx, domainID, capabilityID)
	require.NoError(t, err)
	assert.True(t, exists)

	exists, err = readModel.AssignmentExists(ctx, "bd-nonexistent", capabilityID)
	require.NoError(t, err)
	assert.False(t, exists)

	_, err = db.Exec("DELETE FROM domain_capability_assignments WHERE assignment_id = $1", assignmentID)
	require.NoError(t, err)
}
