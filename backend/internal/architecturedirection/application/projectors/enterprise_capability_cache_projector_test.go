package projectors

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"easi/backend/internal/architecturedirection/application/readmodels"
	eaPL "easi/backend/internal/enterprisearchitecture/publishedlanguage"
)

type recordingEnterpriseCapabilityCache struct {
	inserted         []readmodels.EnterpriseCapabilityCacheDTO
	renamed          []readmodels.EnterpriseCapabilityCacheDTO
	deleted          []string
	targetMaturities map[string]int
}

func (c *recordingEnterpriseCapabilityCache) Insert(_ context.Context, dto readmodels.EnterpriseCapabilityCacheDTO) error {
	c.inserted = append(c.inserted, dto)
	return nil
}

func (c *recordingEnterpriseCapabilityCache) UpdateDetails(_ context.Context, dto readmodels.EnterpriseCapabilityCacheDTO) error {
	c.renamed = append(c.renamed, dto)
	return nil
}

func (c *recordingEnterpriseCapabilityCache) Delete(_ context.Context, id string) error {
	c.deleted = append(c.deleted, id)
	return nil
}

func (c *recordingEnterpriseCapabilityCache) UpdateTargetMaturity(_ context.Context, id string, targetMaturity int) error {
	if c.targetMaturities == nil {
		c.targetMaturities = map[string]int{}
	}
	c.targetMaturities[id] = targetMaturity
	return nil
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

func TestEnterpriseCapabilityCacheProjector_Deleted_RemovesRow(t *testing.T) {
	cache := &recordingEnterpriseCapabilityCache{}
	projector := NewEnterpriseCapabilityCacheProjector(cache)

	projectECEvent(t, projector, eaPL.EnterpriseCapabilityDeleted, map[string]any{"id": "ec-1"})

	assert.Equal(t, []string{"ec-1"}, cache.deleted)
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
	assert.Empty(t, cache.deleted)
}
