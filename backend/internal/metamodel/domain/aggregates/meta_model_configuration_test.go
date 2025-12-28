package aggregates

import (
	"testing"

	"easi/backend/internal/metamodel/domain/events"
	"easi/backend/internal/metamodel/domain/valueobjects"
	domain "easi/backend/internal/shared/eventsourcing"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMetaModelConfiguration(t *testing.T) {
	tenantID, _ := sharedvo.NewTenantID("tenant-123")
	createdBy, _ := valueobjects.NewUserEmail("admin@example.com")

	config, err := NewMetaModelConfiguration(tenantID, createdBy)

	require.NoError(t, err)
	assert.NotEmpty(t, config.ID())
	assert.Equal(t, tenantID, config.TenantID())
	assert.Equal(t, valueobjects.DefaultMaturityScaleConfig(), config.MaturityScaleConfig())
	assert.Equal(t, 1, config.Version())
}

func TestNewMetaModelConfiguration_RaisesCreatedEvent(t *testing.T) {
	tenantID, _ := sharedvo.NewTenantID("tenant-123")
	createdBy, _ := valueobjects.NewUserEmail("admin@example.com")

	config, err := NewMetaModelConfiguration(tenantID, createdBy)

	require.NoError(t, err)
	changes := config.GetUncommittedChanges()
	assert.Len(t, changes, 1)

	createdEvent, ok := changes[0].(events.MetaModelConfigurationCreated)
	assert.True(t, ok)
	assert.Equal(t, config.ID(), createdEvent.ID)
	assert.Equal(t, tenantID.Value(), createdEvent.TenantID)
	assert.Equal(t, createdBy.Value(), createdEvent.CreatedBy)
	assert.Len(t, createdEvent.Sections, 4)
}

func TestMetaModelConfiguration_UpdateMaturityScale(t *testing.T) {
	tenantID, _ := sharedvo.NewTenantID("tenant-123")
	createdBy, _ := valueobjects.NewUserEmail("admin@example.com")
	config, _ := NewMetaModelConfiguration(tenantID, createdBy)
	config.MarkChangesAsCommitted()

	newConfig := createCustomMaturityScaleConfig()
	modifiedBy, _ := valueobjects.NewUserEmail("editor@example.com")

	err := config.UpdateMaturityScale(newConfig, modifiedBy)

	require.NoError(t, err)
	assert.Equal(t, newConfig, config.MaturityScaleConfig())
	assert.Equal(t, 2, config.Version())

	changes := config.GetUncommittedChanges()
	assert.Len(t, changes, 1)

	updatedEvent, ok := changes[0].(events.MaturityScaleConfigUpdated)
	assert.True(t, ok)
	assert.Equal(t, config.ID(), updatedEvent.ID)
	assert.Equal(t, 2, updatedEvent.Version)
	assert.Equal(t, modifiedBy.Value(), updatedEvent.ModifiedBy)
}

func TestMetaModelConfiguration_ResetToDefaults(t *testing.T) {
	tenantID, _ := sharedvo.NewTenantID("tenant-123")
	createdBy, _ := valueobjects.NewUserEmail("admin@example.com")
	config, _ := NewMetaModelConfiguration(tenantID, createdBy)
	config.MarkChangesAsCommitted()

	newConfig := createCustomMaturityScaleConfig()
	modifiedBy, _ := valueobjects.NewUserEmail("editor@example.com")
	_ = config.UpdateMaturityScale(newConfig, modifiedBy)
	config.MarkChangesAsCommitted()

	resetBy, _ := valueobjects.NewUserEmail("admin@example.com")
	err := config.ResetToDefaults(resetBy)

	require.NoError(t, err)
	assert.Equal(t, valueobjects.DefaultMaturityScaleConfig(), config.MaturityScaleConfig())
	assert.Equal(t, 3, config.Version())

	changes := config.GetUncommittedChanges()
	assert.Len(t, changes, 1)

	resetEvent, ok := changes[0].(events.MaturityScaleConfigReset)
	assert.True(t, ok)
	assert.Equal(t, config.ID(), resetEvent.ID)
	assert.Equal(t, 3, resetEvent.Version)
	assert.Equal(t, resetBy.Value(), resetEvent.ModifiedBy)
}

func TestLoadMetaModelConfigurationFromHistory(t *testing.T) {
	createdEvent := events.NewMetaModelConfigurationCreated(events.CreateConfigParams{
		ID:       "config-uuid",
		TenantID: "tenant-123",
		Sections: []events.MaturitySectionData{
			{Order: 1, Name: "Genesis", MinValue: 0, MaxValue: 24},
			{Order: 2, Name: "Custom Built", MinValue: 25, MaxValue: 49},
			{Order: 3, Name: "Product", MinValue: 50, MaxValue: 74},
			{Order: 4, Name: "Commodity", MinValue: 75, MaxValue: 99},
		},
		Pillars: []events.StrategyPillarData{
			{ID: "pillar-1", Name: "Always On", Description: "Core capabilities", Active: true},
			{ID: "pillar-2", Name: "Grow", Description: "Growth capabilities", Active: true},
			{ID: "pillar-3", Name: "Transform", Description: "Transformation capabilities", Active: true},
		},
		CreatedBy: "admin@example.com",
	})

	history := []domain.DomainEvent{createdEvent}

	config, err := LoadMetaModelConfigurationFromHistory(history)

	require.NoError(t, err)
	assert.Equal(t, "config-uuid", config.ID())
	assert.Equal(t, "tenant-123", config.TenantID().Value())
	assert.Equal(t, 1, config.Version())
}

func TestLoadMetaModelConfigurationFromHistory_WithUpdates(t *testing.T) {
	createdEvent := events.NewMetaModelConfigurationCreated(events.CreateConfigParams{
		ID:       "config-uuid",
		TenantID: "tenant-123",
		Sections: []events.MaturitySectionData{
			{Order: 1, Name: "Genesis", MinValue: 0, MaxValue: 24},
			{Order: 2, Name: "Custom Built", MinValue: 25, MaxValue: 49},
			{Order: 3, Name: "Product", MinValue: 50, MaxValue: 74},
			{Order: 4, Name: "Commodity", MinValue: 75, MaxValue: 99},
		},
		Pillars: []events.StrategyPillarData{
			{ID: "pillar-1", Name: "Always On", Description: "Core capabilities", Active: true},
			{ID: "pillar-2", Name: "Grow", Description: "Growth capabilities", Active: true},
			{ID: "pillar-3", Name: "Transform", Description: "Transformation capabilities", Active: true},
		},
		CreatedBy: "admin@example.com",
	})

	updatedEvent := events.NewMaturityScaleConfigUpdated(
		"config-uuid",
		"tenant-123",
		2,
		[]events.MaturitySectionData{
			{Order: 1, Name: "Emerging", MinValue: 0, MaxValue: 30},
			{Order: 2, Name: "Growing", MinValue: 31, MaxValue: 55},
			{Order: 3, Name: "Mature", MinValue: 56, MaxValue: 80},
			{Order: 4, Name: "Declining", MinValue: 81, MaxValue: 99},
		},
		"editor@example.com",
	)

	history := []domain.DomainEvent{createdEvent, updatedEvent}

	config, err := LoadMetaModelConfigurationFromHistory(history)

	require.NoError(t, err)
	assert.Equal(t, "config-uuid", config.ID())
	assert.Equal(t, 2, config.Version())
	assert.Equal(t, "Emerging", config.MaturityScaleConfig().Sections()[0].Name().Value())
}

func TestMetaModelConfiguration_AddStrategyPillar(t *testing.T) {
	tenantID, _ := sharedvo.NewTenantID("tenant-123")
	createdBy, _ := valueobjects.NewUserEmail("admin@example.com")
	config, _ := NewMetaModelConfiguration(tenantID, createdBy)
	config.MarkChangesAsCommitted()

	pillarName, _ := valueobjects.NewPillarName("Innovation")
	pillarDesc, _ := valueobjects.NewPillarDescription("Innovation capabilities")
	modifiedBy, _ := valueobjects.NewUserEmail("editor@example.com")

	err := config.AddStrategyPillar(pillarName, pillarDesc, modifiedBy)

	require.NoError(t, err)
	assert.Equal(t, 4, config.StrategyPillarsConfig().CountActive())
	assert.Equal(t, 2, config.Version())

	changes := config.GetUncommittedChanges()
	assert.Len(t, changes, 1)

	addedEvent, ok := changes[0].(events.StrategyPillarAdded)
	assert.True(t, ok)
	assert.Equal(t, "Innovation", addedEvent.Name)
	assert.Equal(t, modifiedBy.Value(), addedEvent.ModifiedBy)
}

func TestMetaModelConfiguration_AddStrategyPillar_DuplicateName(t *testing.T) {
	tenantID, _ := sharedvo.NewTenantID("tenant-123")
	createdBy, _ := valueobjects.NewUserEmail("admin@example.com")
	config, _ := NewMetaModelConfiguration(tenantID, createdBy)
	config.MarkChangesAsCommitted()

	pillarName, _ := valueobjects.NewPillarName("always on")
	pillarDesc, _ := valueobjects.NewPillarDescription("")
	modifiedBy, _ := valueobjects.NewUserEmail("editor@example.com")

	err := config.AddStrategyPillar(pillarName, pillarDesc, modifiedBy)

	assert.Error(t, err)
	assert.Equal(t, valueobjects.ErrPillarNameDuplicate, err)
}

func TestMetaModelConfiguration_UpdateStrategyPillar(t *testing.T) {
	tenantID, _ := sharedvo.NewTenantID("tenant-123")
	createdBy, _ := valueobjects.NewUserEmail("admin@example.com")
	config, _ := NewMetaModelConfiguration(tenantID, createdBy)
	config.MarkChangesAsCommitted()

	pillarID := config.StrategyPillarsConfig().Pillars()[0].ID()
	newName, _ := valueobjects.NewPillarName("Updated Pillar")
	newDesc, _ := valueobjects.NewPillarDescription("Updated description")
	modifiedBy, _ := valueobjects.NewUserEmail("editor@example.com")

	err := config.UpdateStrategyPillar(pillarID, newName, newDesc, modifiedBy)

	require.NoError(t, err)
	assert.Equal(t, 2, config.Version())
	found, _ := config.StrategyPillarsConfig().FindByID(pillarID)
	assert.Equal(t, "Updated Pillar", found.Name().Value())
}

func TestMetaModelConfiguration_RemoveStrategyPillar(t *testing.T) {
	tenantID, _ := sharedvo.NewTenantID("tenant-123")
	createdBy, _ := valueobjects.NewUserEmail("admin@example.com")
	config, _ := NewMetaModelConfiguration(tenantID, createdBy)
	config.MarkChangesAsCommitted()

	pillarID := config.StrategyPillarsConfig().Pillars()[0].ID()
	modifiedBy, _ := valueobjects.NewUserEmail("editor@example.com")

	err := config.RemoveStrategyPillar(pillarID, modifiedBy)

	require.NoError(t, err)
	assert.Equal(t, 2, config.StrategyPillarsConfig().CountActive())
	assert.Equal(t, 2, config.Version())

	changes := config.GetUncommittedChanges()
	assert.Len(t, changes, 1)

	removedEvent, ok := changes[0].(events.StrategyPillarRemoved)
	assert.True(t, ok)
	assert.Equal(t, pillarID.Value(), removedEvent.PillarID)
}

func TestMetaModelConfiguration_RemoveStrategyPillar_LastActive(t *testing.T) {
	tenantID, _ := sharedvo.NewTenantID("tenant-123")
	createdBy, _ := valueobjects.NewUserEmail("admin@example.com")
	config, _ := NewMetaModelConfiguration(tenantID, createdBy)
	config.MarkChangesAsCommitted()

	pillars := config.StrategyPillarsConfig().Pillars()
	modifiedBy, _ := valueobjects.NewUserEmail("editor@example.com")

	_ = config.RemoveStrategyPillar(pillars[0].ID(), modifiedBy)
	config.MarkChangesAsCommitted()
	_ = config.RemoveStrategyPillar(pillars[1].ID(), modifiedBy)
	config.MarkChangesAsCommitted()

	err := config.RemoveStrategyPillar(pillars[2].ID(), modifiedBy)

	assert.Error(t, err)
	assert.Equal(t, valueobjects.ErrCannotRemoveLastActivePillar, err)
}

func createCustomMaturityScaleConfig() valueobjects.MaturityScaleConfig {
	order1, _ := valueobjects.NewSectionOrder(1)
	name1, _ := valueobjects.NewSectionName("Emerging")
	min1, _ := valueobjects.NewMaturityValue(0)
	max1, _ := valueobjects.NewMaturityValue(30)
	section1, _ := valueobjects.NewMaturitySection(order1, name1, min1, max1)

	order2, _ := valueobjects.NewSectionOrder(2)
	name2, _ := valueobjects.NewSectionName("Growing")
	min2, _ := valueobjects.NewMaturityValue(31)
	max2, _ := valueobjects.NewMaturityValue(55)
	section2, _ := valueobjects.NewMaturitySection(order2, name2, min2, max2)

	order3, _ := valueobjects.NewSectionOrder(3)
	name3, _ := valueobjects.NewSectionName("Mature")
	min3, _ := valueobjects.NewMaturityValue(56)
	max3, _ := valueobjects.NewMaturityValue(80)
	section3, _ := valueobjects.NewMaturitySection(order3, name3, min3, max3)

	order4, _ := valueobjects.NewSectionOrder(4)
	name4, _ := valueobjects.NewSectionName("Declining")
	min4, _ := valueobjects.NewMaturityValue(81)
	max4, _ := valueobjects.NewMaturityValue(99)
	section4, _ := valueobjects.NewMaturitySection(order4, name4, min4, max4)

	config, _ := valueobjects.NewMaturityScaleConfig([4]valueobjects.MaturitySection{section1, section2, section3, section4})
	return config
}
