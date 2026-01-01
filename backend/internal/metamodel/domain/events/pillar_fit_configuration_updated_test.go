package events

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPillarFitConfigurationUpdated(t *testing.T) {
	params := UpdatePillarFitConfigParams{
		PillarEventParams: PillarEventParams{
			ConfigID:   "config-123",
			TenantID:   "tenant-456",
			Version:    2,
			PillarID:   "pillar-789",
			ModifiedBy: "admin@example.com",
		},
		FitScoringEnabled: true,
		FitCriteria:       "Evaluate based on strategic alignment",
	}

	event := NewPillarFitConfigurationUpdated(params)

	assert.Equal(t, params.ConfigID, event.ID)
	assert.Equal(t, params.TenantID, event.TenantID)
	assert.Equal(t, params.Version, event.Version)
	assert.Equal(t, params.PillarID, event.PillarID)
	assert.Equal(t, params.FitScoringEnabled, event.FitScoringEnabled)
	assert.Equal(t, params.FitCriteria, event.FitCriteria)
	assert.Equal(t, params.ModifiedBy, event.ModifiedBy)
	assert.NotZero(t, event.ModifiedAt)
}

func TestPillarFitConfigurationUpdated_EventType(t *testing.T) {
	event := NewPillarFitConfigurationUpdated(UpdatePillarFitConfigParams{
		PillarEventParams: PillarEventParams{ConfigID: "test"},
	})

	assert.Equal(t, "PillarFitConfigurationUpdated", event.EventType())
}

func TestPillarFitConfigurationUpdated_EventData(t *testing.T) {
	params := UpdatePillarFitConfigParams{
		PillarEventParams: PillarEventParams{
			ConfigID:   "config-123",
			TenantID:   "tenant-456",
			Version:    2,
			PillarID:   "pillar-789",
			ModifiedBy: "admin@example.com",
		},
		FitScoringEnabled: true,
		FitCriteria:       "Criteria description",
	}
	event := NewPillarFitConfigurationUpdated(params)
	data := event.EventData()

	assert.Equal(t, "config-123", data["id"])
	assert.Equal(t, "tenant-456", data["tenantId"])
	assert.Equal(t, 2, data["version"])
	assert.Equal(t, "pillar-789", data["pillarId"])
	assert.Equal(t, true, data["fitScoringEnabled"])
	assert.Equal(t, "Criteria description", data["fitCriteria"])
	assert.Equal(t, "admin@example.com", data["modifiedBy"])
	assert.NotNil(t, data["modifiedAt"])
}

func TestPillarFitConfigurationUpdated_DisabledScoring(t *testing.T) {
	params := UpdatePillarFitConfigParams{
		PillarEventParams: PillarEventParams{
			ConfigID: "config-123",
			PillarID: "pillar-789",
		},
		FitScoringEnabled: false,
		FitCriteria:       "",
	}

	event := NewPillarFitConfigurationUpdated(params)

	assert.False(t, event.FitScoringEnabled)
	assert.Empty(t, event.FitCriteria)
}
