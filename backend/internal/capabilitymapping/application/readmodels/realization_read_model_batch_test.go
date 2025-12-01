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

func TestRealizationReadModel_GetByCapabilityIDs_ReturnsRealizationsForMultipleCapabilities(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	tenantDB := database.NewTenantAwareDB(db)
	readModel := NewRealizationReadModel(tenantDB)

	cap1ID := fmt.Sprintf("cap-1-%d", time.Now().UnixNano())
	cap2ID := fmt.Sprintf("cap-2-%d", time.Now().UnixNano())
	comp1ID := fmt.Sprintf("comp-1-%d", time.Now().UnixNano())
	comp2ID := fmt.Sprintf("comp-2-%d", time.Now().UnixNano())
	real1ID := fmt.Sprintf("real-1-%d", time.Now().UnixNano())
	real2ID := fmt.Sprintf("real-2-%d", time.Now().UnixNano())
	real3ID := fmt.Sprintf("real-3-%d", time.Now().UnixNano())

	setTenantContext(t, db)

	_, err := db.Exec(
		"INSERT INTO application_components (id, tenant_id, name, description, created_at) VALUES ($1, $2, $3, $4, $5)",
		comp1ID, "default", "Order Service", "Handles orders", time.Now().UTC(),
	)
	require.NoError(t, err)
	defer db.Exec("DELETE FROM application_components WHERE id = $1", comp1ID)

	_, err = db.Exec(
		"INSERT INTO application_components (id, tenant_id, name, description, created_at) VALUES ($1, $2, $3, $4, $5)",
		comp2ID, "default", "Payment Gateway", "Processes payments", time.Now().UTC(),
	)
	require.NoError(t, err)
	defer db.Exec("DELETE FROM application_components WHERE id = $1", comp2ID)

	_, err = db.Exec(
		"INSERT INTO capability_realizations (id, tenant_id, capability_id, component_id, realization_level, notes, origin, linked_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
		real1ID, "default", cap1ID, comp1ID, "Full", "Primary implementation", "Direct", time.Now().UTC(),
	)
	require.NoError(t, err)
	defer db.Exec("DELETE FROM capability_realizations WHERE id = $1", real1ID)

	_, err = db.Exec(
		"INSERT INTO capability_realizations (id, tenant_id, capability_id, component_id, realization_level, notes, origin, linked_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
		real2ID, "default", cap2ID, comp2ID, "Partial", "", "Direct", time.Now().UTC(),
	)
	require.NoError(t, err)
	defer db.Exec("DELETE FROM capability_realizations WHERE id = $1", real2ID)

	_, err = db.Exec(
		"INSERT INTO capability_realizations (id, tenant_id, capability_id, component_id, realization_level, notes, origin, linked_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
		real3ID, "default", cap2ID, comp1ID, "Planned", "Future implementation", "Direct", time.Now().UTC(),
	)
	require.NoError(t, err)
	defer db.Exec("DELETE FROM capability_realizations WHERE id = $1", real3ID)

	ctx := tenantContext()
	realizations, err := readModel.GetByCapabilityIDs(ctx, []string{cap1ID, cap2ID})
	require.NoError(t, err)

	assert.Len(t, realizations, 3, "Should return realizations for both capabilities")

	realizationsByCapability := make(map[string][]RealizationDTO)
	for _, r := range realizations {
		realizationsByCapability[r.CapabilityID] = append(realizationsByCapability[r.CapabilityID], r)
	}

	assert.Len(t, realizationsByCapability[cap1ID], 1, "Cap1 should have 1 realization")
	assert.Len(t, realizationsByCapability[cap2ID], 2, "Cap2 should have 2 realizations")
}

func TestRealizationReadModel_GetByCapabilityIDs_IncludesComponentName(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	tenantDB := database.NewTenantAwareDB(db)
	readModel := NewRealizationReadModel(tenantDB)

	capID := fmt.Sprintf("cap-%d", time.Now().UnixNano())
	compID := fmt.Sprintf("comp-%d", time.Now().UnixNano())
	realID := fmt.Sprintf("real-%d", time.Now().UnixNano())

	setTenantContext(t, db)

	_, err := db.Exec(
		"INSERT INTO application_components (id, tenant_id, name, description, created_at) VALUES ($1, $2, $3, $4, $5)",
		compID, "default", "Billing System", "Handles billing", time.Now().UTC(),
	)
	require.NoError(t, err)
	defer db.Exec("DELETE FROM application_components WHERE id = $1", compID)

	_, err = db.Exec(
		"INSERT INTO capability_realizations (id, tenant_id, capability_id, component_id, realization_level, notes, origin, linked_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
		realID, "default", capID, compID, "Full", "", "Direct", time.Now().UTC(),
	)
	require.NoError(t, err)
	defer db.Exec("DELETE FROM capability_realizations WHERE id = $1", realID)

	ctx := tenantContext()
	realizations, err := readModel.GetByCapabilityIDs(ctx, []string{capID})
	require.NoError(t, err)
	require.Len(t, realizations, 1)

	assert.Equal(t, "Billing System", realizations[0].ComponentName, "Component name must be denormalized")
}

func TestRealizationReadModel_GetByCapabilityIDs_IncludesSourceCapabilityForInherited(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	tenantDB := database.NewTenantAwareDB(db)
	readModel := NewRealizationReadModel(tenantDB)

	parentCapID := fmt.Sprintf("parent-cap-%d", time.Now().UnixNano())
	childCapID := fmt.Sprintf("child-cap-%d", time.Now().UnixNano())
	compID := fmt.Sprintf("comp-%d", time.Now().UnixNano())
	realDirectID := fmt.Sprintf("real-direct-%d", time.Now().UnixNano())
	realInheritedID := fmt.Sprintf("real-inherited-%d", time.Now().UnixNano())

	setTenantContext(t, db)

	_, err := db.Exec(
		"INSERT INTO capabilities (id, tenant_id, name, description, level, maturity_level, status, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
		childCapID, "default", "Order Processing", "Processes orders", "L3", "Genesis", "Active", time.Now().UTC(),
	)
	require.NoError(t, err)
	defer db.Exec("DELETE FROM capabilities WHERE id = $1", childCapID)

	_, err = db.Exec(
		"INSERT INTO application_components (id, tenant_id, name, description, created_at) VALUES ($1, $2, $3, $4, $5)",
		compID, "default", "Order Service", "Handles orders", time.Now().UTC(),
	)
	require.NoError(t, err)
	defer db.Exec("DELETE FROM application_components WHERE id = $1", compID)

	_, err = db.Exec(
		"INSERT INTO capability_realizations (id, tenant_id, capability_id, component_id, realization_level, notes, origin, source_realization_id, linked_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)",
		realDirectID, "default", childCapID, compID, "Full", "", "Direct", nil, time.Now().UTC(),
	)
	require.NoError(t, err)
	defer db.Exec("DELETE FROM capability_realizations WHERE id = $1", realDirectID)

	_, err = db.Exec(
		"INSERT INTO capability_realizations (id, tenant_id, capability_id, component_id, realization_level, notes, origin, source_realization_id, linked_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)",
		realInheritedID, "default", parentCapID, compID, "Full", "", "Inherited", realDirectID, time.Now().UTC(),
	)
	require.NoError(t, err)
	defer db.Exec("DELETE FROM capability_realizations WHERE id = $1", realInheritedID)

	ctx := tenantContext()
	realizations, err := readModel.GetByCapabilityIDs(ctx, []string{parentCapID})
	require.NoError(t, err)
	require.Len(t, realizations, 1)

	inherited := realizations[0]
	assert.Equal(t, "Inherited", inherited.Origin)
	assert.Equal(t, childCapID, inherited.SourceCapabilityID, "Source capability ID must be populated")
	assert.Equal(t, "Order Processing", inherited.SourceCapabilityName, "Source capability name must be denormalized for inherited realizations")
}

func TestRealizationReadModel_GetByCapabilityIDs_EmptyListReturnsEmpty(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	tenantDB := database.NewTenantAwareDB(db)
	readModel := NewRealizationReadModel(tenantDB)

	ctx := tenantContext()
	realizations, err := readModel.GetByCapabilityIDs(ctx, []string{})
	require.NoError(t, err)
	assert.Empty(t, realizations)
}

func TestRealizationReadModel_GetByCapabilityIDs_NonExistentCapabilitiesReturnEmpty(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	tenantDB := database.NewTenantAwareDB(db)
	readModel := NewRealizationReadModel(tenantDB)

	ctx := tenantContext()
	realizations, err := readModel.GetByCapabilityIDs(ctx, []string{"non-existent-1", "non-existent-2"})
	require.NoError(t, err)
	assert.Empty(t, realizations)
}

func TestRealizationReadModel_GetByCapabilityIDs_OrdersByLinkedAtDescending(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	tenantDB := database.NewTenantAwareDB(db)
	readModel := NewRealizationReadModel(tenantDB)

	capID := fmt.Sprintf("cap-%d", time.Now().UnixNano())
	comp1ID := fmt.Sprintf("comp-1-%d", time.Now().UnixNano())
	comp2ID := fmt.Sprintf("comp-2-%d", time.Now().UnixNano())
	real1ID := fmt.Sprintf("real-1-%d", time.Now().UnixNano())
	real2ID := fmt.Sprintf("real-2-%d", time.Now().UnixNano())

	earlierTime := time.Now().UTC().Add(-1 * time.Hour)
	laterTime := time.Now().UTC()

	setTenantContext(t, db)

	_, err := db.Exec(
		"INSERT INTO application_components (id, tenant_id, name, description, created_at) VALUES ($1, $2, $3, $4, $5)",
		comp1ID, "default", "First Component", "", time.Now().UTC(),
	)
	require.NoError(t, err)
	defer db.Exec("DELETE FROM application_components WHERE id = $1", comp1ID)

	_, err = db.Exec(
		"INSERT INTO application_components (id, tenant_id, name, description, created_at) VALUES ($1, $2, $3, $4, $5)",
		comp2ID, "default", "Second Component", "", time.Now().UTC(),
	)
	require.NoError(t, err)
	defer db.Exec("DELETE FROM application_components WHERE id = $1", comp2ID)

	_, err = db.Exec(
		"INSERT INTO capability_realizations (id, tenant_id, capability_id, component_id, realization_level, notes, origin, linked_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
		real1ID, "default", capID, comp1ID, "Full", "", "Direct", earlierTime,
	)
	require.NoError(t, err)
	defer db.Exec("DELETE FROM capability_realizations WHERE id = $1", real1ID)

	_, err = db.Exec(
		"INSERT INTO capability_realizations (id, tenant_id, capability_id, component_id, realization_level, notes, origin, linked_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
		real2ID, "default", capID, comp2ID, "Full", "", "Direct", laterTime,
	)
	require.NoError(t, err)
	defer db.Exec("DELETE FROM capability_realizations WHERE id = $1", real2ID)

	ctx := tenantContext()
	realizations, err := readModel.GetByCapabilityIDs(ctx, []string{capID})
	require.NoError(t, err)
	require.Len(t, realizations, 2)

	assert.Equal(t, "Second Component", realizations[0].ComponentName, "Most recent realization should be first")
	assert.Equal(t, "First Component", realizations[1].ComponentName, "Older realization should be second")
}
