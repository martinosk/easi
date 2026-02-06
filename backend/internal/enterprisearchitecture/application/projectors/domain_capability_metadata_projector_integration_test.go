//go:build integration

package projectors

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"easi/backend/internal/enterprisearchitecture/application/readmodels"
	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/eventsourcing/valueobjects"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testTenant = "projector-test-tenant"

func setupProjectorTest(t *testing.T) (context.Context, *DomainCapabilityMetadataProjector, *readmodels.DomainCapabilityMetadataReadModel, *sql.DB) {
	t.Helper()

	db, err := sql.Open("postgres", "host=localhost port=5432 user=easi password=easi dbname=easi sslmode=disable")
	require.NoError(t, err)

	tenantDB := database.NewTenantAwareDB(db)
	metadataRM := readmodels.NewDomainCapabilityMetadataReadModel(tenantDB)
	capabilityRM := readmodels.NewEnterpriseCapabilityReadModel(tenantDB)
	linkRM := readmodels.NewEnterpriseCapabilityLinkReadModel(tenantDB)
	projector := NewDomainCapabilityMetadataProjector(metadataRM, capabilityRM, linkRM)

	ctx := sharedctx.WithTenant(context.Background(), valueobjects.MustNewTenantID(testTenant))

	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM domain_capability_metadata WHERE tenant_id = $1", testTenant)
		_, _ = db.Exec("DELETE FROM domain_capability_assignments WHERE tenant_id = $1", testTenant)
		db.Close()
	})

	return ctx, projector, metadataRM, db
}

func TestMetadataProjector_AssignToDomain_SetsBusinessDomainOnMetadata(t *testing.T) {
	ctx, projector, metadataRM, db := setupProjectorTest(t)

	l1ID := uuid.New().String()
	l2ID := uuid.New().String()
	domainID := uuid.New().String()

	err := metadataRM.Insert(ctx, readmodels.DomainCapabilityMetadataDTO{
		CapabilityID:    l1ID,
		CapabilityName:  "L1 Capability",
		CapabilityLevel: "L1",
		L1CapabilityID:  l1ID,
	})
	require.NoError(t, err)

	err = metadataRM.Insert(ctx, readmodels.DomainCapabilityMetadataDTO{
		CapabilityID:    l2ID,
		CapabilityName:  "L2 Capability",
		CapabilityLevel: "L2",
		ParentID:        l1ID,
		L1CapabilityID:  l1ID,
	})
	require.NoError(t, err)

	_, err = db.Exec(
		`INSERT INTO domain_capability_assignments
		 (assignment_id, tenant_id, business_domain_id, business_domain_name, capability_id, capability_name, capability_level, assigned_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		uuid.New().String(), testTenant, domainID, "Test Domain", l1ID, "L1 Capability", "L1", time.Now().UTC(),
	)
	require.NoError(t, err)

	eventData, err := json.Marshal(map[string]interface{}{
		"id":               uuid.New().String(),
		"businessDomainId": domainID,
		"capabilityId":     l1ID,
		"assignedAt":       time.Now().UTC(),
	})
	require.NoError(t, err)

	err = projector.ProjectEvent(ctx, "CapabilityAssignedToDomain", eventData)
	require.NoError(t, err)

	l1Meta, err := metadataRM.GetByID(ctx, l1ID)
	require.NoError(t, err)
	assert.Equal(t, domainID, l1Meta.BusinessDomainID, "L1 metadata should have business domain set after assignment")
	assert.Equal(t, "Test Domain", l1Meta.BusinessDomainName)

	l2Meta, err := metadataRM.GetByID(ctx, l2ID)
	require.NoError(t, err)
	assert.Equal(t, domainID, l2Meta.BusinessDomainID, "L2 child should inherit business domain from L1 parent")
	assert.Equal(t, "Test Domain", l2Meta.BusinessDomainName)
}

func TestMetadataProjector_UnassignFromDomain_ClearsBusinessDomainOnMetadata(t *testing.T) {
	ctx, projector, metadataRM, _ := setupProjectorTest(t)

	l1ID := uuid.New().String()
	l2ID := uuid.New().String()
	domainID := uuid.New().String()

	err := metadataRM.Insert(ctx, readmodels.DomainCapabilityMetadataDTO{
		CapabilityID:       l1ID,
		CapabilityName:     "L1 Capability",
		CapabilityLevel:    "L1",
		L1CapabilityID:     l1ID,
		BusinessDomainID:   domainID,
		BusinessDomainName: "Test Domain",
	})
	require.NoError(t, err)

	err = metadataRM.Insert(ctx, readmodels.DomainCapabilityMetadataDTO{
		CapabilityID:       l2ID,
		CapabilityName:     "L2 Capability",
		CapabilityLevel:    "L2",
		ParentID:           l1ID,
		L1CapabilityID:     l1ID,
		BusinessDomainID:   domainID,
		BusinessDomainName: "Test Domain",
	})
	require.NoError(t, err)

	eventData, err := json.Marshal(map[string]interface{}{
		"id":               uuid.New().String(),
		"businessDomainId": domainID,
		"capabilityId":     l1ID,
		"unassignedAt":     time.Now().UTC(),
	})
	require.NoError(t, err)

	err = projector.ProjectEvent(ctx, "CapabilityUnassignedFromDomain", eventData)
	require.NoError(t, err)

	l1Meta, err := metadataRM.GetByID(ctx, l1ID)
	require.NoError(t, err)
	assert.Empty(t, l1Meta.BusinessDomainID, "L1 metadata should have business domain cleared after unassignment")
	assert.Empty(t, l1Meta.BusinessDomainName)

	l2Meta, err := metadataRM.GetByID(ctx, l2ID)
	require.NoError(t, err)
	assert.Empty(t, l2Meta.BusinessDomainID, "L2 child should have business domain cleared when L1 parent is unassigned")
	assert.Empty(t, l2Meta.BusinessDomainName)
}
