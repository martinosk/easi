// Test to reproduce the bug where child capabilities don't inherit business domain from parent

package readmodels

import (
	"context"
	"database/sql"
	"testing"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/eventsourcing/valueobjects"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDomainCapabilityMetadata_ChildInheritsFromParent(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := sharedctx.WithTenant(context.Background(), valueobjects.MustNewTenantID("test-tenant"))

	db, err := sql.Open("postgres", "host=localhost port=5432 user=easi password=easi dbname=easi sslmode=disable")
	require.NoError(t, err)
	defer db.Close()

	tenantDB := database.NewTenantAwareDB(db)
	readModel := NewDomainCapabilityMetadataReadModel(tenantDB)

	// Cleanup
	defer func() {
		_, _ = db.Exec("DELETE FROM domain_capability_metadata WHERE tenant_id = 'test-tenant'")
	}()

	// Test scenario: Create parent L1, then child L2, then assign parent to domain
	
	// Step 1: Create L1 parent
	parentDTO := DomainCapabilityMetadataDTO{
		CapabilityID:    "parent-l1-id",
		CapabilityName:  "Parent L1",
		CapabilityLevel: "L1",
		ParentID:        "",
		L1CapabilityID:  "parent-l1-id",
	}
	err = readModel.Insert(ctx, parentDTO)
	require.NoError(t, err)

	// Step 2: Create L2 child (simulating what projector does when child is created)
	// At this point, parent exists but doesn't have business domain yet
	parentMeta, err := readModel.GetByID(ctx, "parent-l1-id")
	require.NoError(t, err)
	require.NotNil(t, parentMeta)
	
	childDTO := DomainCapabilityMetadataDTO{
		CapabilityID:       "child-l2-id",
		CapabilityName:     "Child L2",
		CapabilityLevel:    "L2",
		ParentID:           "parent-l1-id",
		L1CapabilityID:     parentMeta.L1CapabilityID,  // Should inherit from parent
		BusinessDomainID:   parentMeta.BusinessDomainID, // Should be empty
		BusinessDomainName: parentMeta.BusinessDomainName, // Should be empty
	}
	err = readModel.Insert(ctx, childDTO)
	require.NoError(t, err)

	// Verify child was created with correct L1 ID
	childMeta, err := readModel.GetByID(ctx, "child-l2-id")
	require.NoError(t, err)
	assert.Equal(t, "parent-l1-id", childMeta.L1CapabilityID, "Child should inherit L1 ID from parent")
	assert.Equal(t, "", childMeta.BusinessDomainID)

	// Step 3: Assign parent to business domain (simulating what projector does)
	err = readModel.UpdateBusinessDomainForL1Subtree(ctx, "parent-l1-id", "domain-id", "Domain Name")
	require.NoError(t, err)

	// Verify parent got domain
	parentMeta, err = readModel.GetByID(ctx, "parent-l1-id")
	require.NoError(t, err)
	assert.Equal(t, "domain-id", parentMeta.BusinessDomainID)
	assert.Equal(t, "Domain Name", parentMeta.BusinessDomainName)

	// Verify child also got domain
	childMeta, err = readModel.GetByID(ctx, "child-l2-id")
	require.NoError(t, err)
	assert.Equal(t, "domain-id", childMeta.BusinessDomainID, "Child should inherit business domain when parent is assigned")
	assert.Equal(t, "Domain Name", childMeta.BusinessDomainName)
}
