package projectors

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"easi/backend/internal/architecturedirection/application/readmodels"
	"easi/backend/internal/architecturedirection/domain/aggregates"
	"easi/backend/internal/architecturedirection/domain/services"
	"easi/backend/internal/architecturedirection/domain/valueobjects"
	eaPL "easi/backend/internal/enterprisearchitecture/publishedlanguage"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

type recordingEnterpriseCapabilityCache struct {
	rows             map[string]*readmodels.EnterpriseCapabilityCacheDTO
	inserted         []readmodels.EnterpriseCapabilityCacheDTO
	renamed          []readmodels.EnterpriseCapabilityCacheDTO
	deactivated      []string
	targetMaturities map[string]int
}

func (c *recordingEnterpriseCapabilityCache) Insert(_ context.Context, dto readmodels.EnterpriseCapabilityCacheDTO) error {
	c.inserted = append(c.inserted, dto)
	if c.rows == nil {
		c.rows = map[string]*readmodels.EnterpriseCapabilityCacheDTO{}
	}
	row := dto
	c.rows[dto.ID] = &row
	return nil
}

func (c *recordingEnterpriseCapabilityCache) UpdateDetails(_ context.Context, dto readmodels.EnterpriseCapabilityCacheDTO) error {
	c.renamed = append(c.renamed, dto)
	if row, ok := c.rows[dto.ID]; ok {
		row.Name = dto.Name
		row.Category = dto.Category
	}
	return nil
}

func (c *recordingEnterpriseCapabilityCache) Deactivate(_ context.Context, id string) error {
	c.deactivated = append(c.deactivated, id)
	if row, ok := c.rows[id]; ok {
		row.Active = false
	}
	return nil
}

func (c *recordingEnterpriseCapabilityCache) UpdateTargetMaturity(_ context.Context, id string, targetMaturity int) error {
	if c.targetMaturities == nil {
		c.targetMaturities = map[string]int{}
	}
	c.targetMaturities[id] = targetMaturity
	return nil
}

func (c *recordingEnterpriseCapabilityCache) GetByID(_ context.Context, id string) (*readmodels.EnterpriseCapabilityCacheDTO, error) {
	row, ok := c.rows[id]
	if !ok {
		return nil, nil
	}
	found := *row
	return &found, nil
}

func projectECEvent(t *testing.T, p *EnterpriseCapabilityCacheProjector, eventType string, payload map[string]any) {
	t.Helper()
	data, err := json.Marshal(payload)
	require.NoError(t, err)
	require.NoError(t, p.ProjectEvent(context.Background(), eventType, data))
}

func TestEnterpriseCapabilityCacheProjector_Created_InsertsIdentity(t *testing.T) {
	cache := &recordingEnterpriseCapabilityCache{}
	projector := NewEnterpriseCapabilityCacheProjector(cache)

	projectECEvent(t, projector, eaPL.EnterpriseCapabilityCreated, map[string]any{
		"id": "ec-1", "name": "Payments", "description": "d", "category": "Finance", "active": true,
	})

	require.Len(t, cache.inserted, 1)
	assert.Equal(t, readmodels.EnterpriseCapabilityCacheDTO{ID: "ec-1", Name: "Payments", Category: "Finance", Active: true}, cache.inserted[0])
}

func TestEnterpriseCapabilityCacheProjector_Updated_RenamesAndRecategorises(t *testing.T) {
	cache := &recordingEnterpriseCapabilityCache{}
	projector := NewEnterpriseCapabilityCacheProjector(cache)

	projectECEvent(t, projector, eaPL.EnterpriseCapabilityUpdated, map[string]any{
		"id": "ec-1", "name": "Payments & Billing", "description": "d", "category": "Revenue",
	})

	require.Len(t, cache.renamed, 1)
	assert.Equal(t, "Payments & Billing", cache.renamed[0].Name)
	assert.Equal(t, "Revenue", cache.renamed[0].Category)
	assert.Equal(t, "ec-1", cache.renamed[0].ID)
}

func TestEnterpriseCapabilityCacheProjector_Deleted_DeactivatesRowInsteadOfRemovingIt(t *testing.T) {
	cache := &recordingEnterpriseCapabilityCache{}
	projector := NewEnterpriseCapabilityCacheProjector(cache)

	projectECEvent(t, projector, eaPL.EnterpriseCapabilityCreated, map[string]any{
		"id": "ec-1", "name": "Payments", "description": "d", "category": "Finance", "active": true,
	})
	projectECEvent(t, projector, eaPL.EnterpriseCapabilityDeleted, map[string]any{"id": "ec-1"})

	assert.Equal(t, []string{"ec-1"}, cache.deactivated)
	row, err := cache.GetByID(context.Background(), "ec-1")
	require.NoError(t, err)
	require.NotNil(t, row, "the cache row must remain so existence checks still find the enterprise capability")
	assert.False(t, row.Active)
}

func TestEnterpriseCapabilityCacheProjector_DeletedEC_CaptureDirectionIsRejectedAsInactiveNot404(t *testing.T) {
	ecID := uuid.New().String()
	sourceID := uuid.New().String()
	cache := &recordingEnterpriseCapabilityCache{}
	projector := NewEnterpriseCapabilityCacheProjector(cache)

	projectECEvent(t, projector, eaPL.EnterpriseCapabilityCreated, map[string]any{
		"id": ecID, "name": "Payments", "description": "d", "category": "Finance", "active": true,
	})
	projectECEvent(t, projector, eaPL.EnterpriseCapabilityDeleted, map[string]any{"id": ecID})

	refs := &services.ReferenceChecker{
		EnterpriseCapabilityExists: func(ctx context.Context, id string) (bool, error) {
			row, err := cache.GetByID(ctx, id)
			return row != nil, err
		},
		EnterpriseCapabilityIsActive: func(ctx context.Context, id string) (bool, error) {
			row, err := cache.GetByID(ctx, id)
			return row != nil && row.Active, err
		},
		PhysicalCapabilityExists: func(context.Context, string) (bool, error) { return true, nil },
	}
	policy := services.NewDirectionReferenceService(refs, noActiveDirectionLookup{}, noSourceConflicts{})

	ecRef, err := valueobjects.NewEnterpriseCapabilityRef(ecID)
	require.NoError(t, err)
	sourceRef, err := valueobjects.NewPhysicalCapabilityRef(sourceID)
	require.NoError(t, err)
	directionType, err := valueobjects.NewDirectionType(valueobjects.DirectionTypeStay)
	require.NoError(t, err)
	horizon, err := valueobjects.NewHorizon(valueobjects.HorizonNext)
	require.NoError(t, err)
	narrative, err := sharedvo.NewDescription("Keep payroll on the standard platform.")
	require.NoError(t, err)

	err = policy.VerifyCanCapture(context.Background(), aggregates.DraftParams{
		EnterpriseCapabilityID: ecRef,
		Type:                   directionType,
		SourceCapabilityIDs:    []valueobjects.PhysicalCapabilityRef{sourceRef},
		Horizon:                horizon,
		Narrative:              narrative,
	})

	require.ErrorIs(t, err, services.ErrEnterpriseCapabilityInactive, "a soft-deleted EC must be rejected as inactive (409), not as missing (404) (R4)")
	require.NotErrorIs(t, err, services.ErrReferencedEntityNotFound)
}

type noActiveDirectionLookup struct{}

func (noActiveDirectionLookup) HasActiveDirectionForEnterpriseCapability(context.Context, string) (bool, error) {
	return false, nil
}

type noSourceConflicts struct{}

func (noSourceConflicts) FirstSourceConflict(context.Context, string, []string) (*services.SourceConflict, error) {
	return nil, nil
}

func TestEnterpriseCapabilityCacheProjector_TargetMaturitySet_UpdatesTarget(t *testing.T) {
	cache := &recordingEnterpriseCapabilityCache{}
	projector := NewEnterpriseCapabilityCacheProjector(cache)

	projectECEvent(t, projector, eaPL.EnterpriseCapabilityTargetMaturitySet, map[string]any{"id": "ec-1", "targetMaturity": 75})

	assert.Equal(t, 75, cache.targetMaturities["ec-1"])
}

func TestEnterpriseCapabilityCacheProjector_IgnoresUnrelatedEvents(t *testing.T) {
	cache := &recordingEnterpriseCapabilityCache{}
	projector := NewEnterpriseCapabilityCacheProjector(cache)

	projectECEvent(t, projector, eaPL.EnterpriseStrategicImportanceSet, map[string]any{"id": "imp-1"})

	assert.Empty(t, cache.inserted)
	assert.Empty(t, cache.renamed)
	assert.Empty(t, cache.deactivated)
}
