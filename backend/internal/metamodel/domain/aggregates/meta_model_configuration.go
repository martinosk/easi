package aggregates

import (
	"easi/backend/internal/metamodel/domain/events"
	"easi/backend/internal/metamodel/domain/valueobjects"
	domain "easi/backend/internal/shared/eventsourcing"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

type MetaModelConfiguration struct {
	domain.AggregateRoot
	tenantID              sharedvo.TenantID
	maturityScaleConfig   valueobjects.MaturityScaleConfig
	strategyPillarsConfig valueobjects.StrategyPillarsConfig
	createdAt             valueobjects.Timestamp
	modifiedAt            valueobjects.Timestamp
	modifiedBy            valueobjects.UserEmail
}

func NewMetaModelConfiguration(tenantID sharedvo.TenantID, createdBy valueobjects.UserEmail) (*MetaModelConfiguration, error) {
	aggregate := &MetaModelConfiguration{
		AggregateRoot: domain.NewAggregateRoot(),
	}

	defaultMaturityConfig := valueobjects.DefaultMaturityScaleConfig()
	defaultPillarsConfig := valueobjects.DefaultStrategyPillarsConfig()
	sections := maturityScaleConfigToEventData(defaultMaturityConfig)
	pillars := strategyPillarsConfigToEventData(defaultPillarsConfig)

	event := events.NewMetaModelConfigurationCreated(events.CreateConfigParams{
		ID:        aggregate.ID(),
		TenantID:  tenantID.Value(),
		Sections:  sections,
		Pillars:   pillars,
		CreatedBy: createdBy.Value(),
	})

	aggregate.apply(event)
	aggregate.RaiseEvent(event)

	return aggregate, nil
}

func LoadMetaModelConfigurationFromHistory(eventHistory []domain.DomainEvent) (*MetaModelConfiguration, error) {
	aggregate := &MetaModelConfiguration{
		AggregateRoot: domain.NewAggregateRoot(),
	}

	aggregate.LoadFromHistory(eventHistory, func(event domain.DomainEvent) {
		aggregate.apply(event)
	})

	return aggregate, nil
}

func (m *MetaModelConfiguration) UpdateMaturityScale(config valueobjects.MaturityScaleConfig, modifiedBy valueobjects.UserEmail) error {
	sections := maturityScaleConfigToEventData(config)
	event := events.NewMaturityScaleConfigUpdated(
		m.ID(),
		m.tenantID.Value(),
		m.Version()+1,
		sections,
		modifiedBy.Value(),
	)
	m.applyAndRaise(event)
	return nil
}

func (m *MetaModelConfiguration) ResetToDefaults(modifiedBy valueobjects.UserEmail) error {
	sections := maturityScaleConfigToEventData(valueobjects.DefaultMaturityScaleConfig())
	event := events.NewMaturityScaleConfigReset(
		m.ID(),
		m.tenantID.Value(),
		m.Version()+1,
		sections,
		modifiedBy.Value(),
	)
	m.applyAndRaise(event)
	return nil
}

func (m *MetaModelConfiguration) applyAndRaise(event domain.DomainEvent) {
	m.apply(event)
	m.RaiseEvent(event)
}

func (m *MetaModelConfiguration) apply(event domain.DomainEvent) {
	switch e := event.(type) {
	case events.MetaModelConfigurationCreated:
		m.AggregateRoot = domain.NewAggregateRootWithID(e.ID)
		m.tenantID, _ = sharedvo.NewTenantID(e.TenantID)
		m.maturityScaleConfig = eventDataToMaturityScaleConfig(e.Sections)
		m.strategyPillarsConfig = eventDataToStrategyPillarsConfig(e.Pillars)
		m.createdAt, _ = valueobjects.NewTimestamp(e.CreatedAt)
		m.modifiedAt, _ = valueobjects.NewTimestamp(e.CreatedAt)
		m.modifiedBy, _ = valueobjects.NewUserEmail(e.CreatedBy)
	case events.MaturityScaleConfigUpdated:
		m.maturityScaleConfig = eventDataToMaturityScaleConfig(e.NewSections)
		m.modifiedAt, _ = valueobjects.NewTimestamp(e.ModifiedAt)
		m.modifiedBy, _ = valueobjects.NewUserEmail(e.ModifiedBy)
	case events.MaturityScaleConfigReset:
		m.maturityScaleConfig = eventDataToMaturityScaleConfig(e.Sections)
		m.modifiedAt, _ = valueobjects.NewTimestamp(e.ModifiedAt)
		m.modifiedBy, _ = valueobjects.NewUserEmail(e.ModifiedBy)
	case events.StrategyPillarAdded:
		m.applyPillarAdded(e)
	case events.StrategyPillarUpdated:
		m.applyPillarUpdated(e)
	case events.StrategyPillarRemoved:
		m.applyPillarRemoved(e)
	}
}

func (m *MetaModelConfiguration) applyPillarAdded(e events.StrategyPillarAdded) {
	pillarID, _ := valueobjects.NewStrategyPillarIDFromString(e.PillarID)
	pillarName, _ := valueobjects.NewPillarName(e.Name)
	pillarDesc, _ := valueobjects.NewPillarDescription(e.Description)
	pillar, _ := valueobjects.NewStrategyPillar(pillarID, pillarName, pillarDesc)
	m.strategyPillarsConfig, _ = m.strategyPillarsConfig.WithAddedPillar(pillar)
	m.modifiedAt, _ = valueobjects.NewTimestamp(e.ModifiedAt)
	m.modifiedBy, _ = valueobjects.NewUserEmail(e.ModifiedBy)
}

func (m *MetaModelConfiguration) applyPillarUpdated(e events.StrategyPillarUpdated) {
	pillarID, _ := valueobjects.NewStrategyPillarIDFromString(e.PillarID)
	newName, _ := valueobjects.NewPillarName(e.NewName)
	newDesc, _ := valueobjects.NewPillarDescription(e.NewDescription)
	m.strategyPillarsConfig, _ = m.strategyPillarsConfig.WithUpdatedPillar(pillarID, newName, newDesc)
	m.modifiedAt, _ = valueobjects.NewTimestamp(e.ModifiedAt)
	m.modifiedBy, _ = valueobjects.NewUserEmail(e.ModifiedBy)
}

func (m *MetaModelConfiguration) applyPillarRemoved(e events.StrategyPillarRemoved) {
	pillarID, _ := valueobjects.NewStrategyPillarIDFromString(e.PillarID)
	m.strategyPillarsConfig, _ = m.strategyPillarsConfig.WithRemovedPillar(pillarID)
	m.modifiedAt, _ = valueobjects.NewTimestamp(e.ModifiedAt)
	m.modifiedBy, _ = valueobjects.NewUserEmail(e.ModifiedBy)
}

func (m *MetaModelConfiguration) TenantID() sharedvo.TenantID {
	return m.tenantID
}

func (m *MetaModelConfiguration) MaturityScaleConfig() valueobjects.MaturityScaleConfig {
	return m.maturityScaleConfig
}

func (m *MetaModelConfiguration) CreatedAt() valueobjects.Timestamp {
	return m.createdAt
}

func (m *MetaModelConfiguration) ModifiedAt() valueobjects.Timestamp {
	return m.modifiedAt
}

func (m *MetaModelConfiguration) ModifiedBy() valueobjects.UserEmail {
	return m.modifiedBy
}

func (m *MetaModelConfiguration) StrategyPillarsConfig() valueobjects.StrategyPillarsConfig {
	return m.strategyPillarsConfig
}

func (m *MetaModelConfiguration) AddStrategyPillar(name valueobjects.PillarName, description valueobjects.PillarDescription, modifiedBy valueobjects.UserEmail) error {
	pillarID := valueobjects.NewStrategyPillarID()
	pillar, err := valueobjects.NewStrategyPillar(pillarID, name, description)
	if err != nil {
		return err
	}

	if _, err := m.strategyPillarsConfig.WithAddedPillar(pillar); err != nil {
		return err
	}

	event := events.NewStrategyPillarAdded(events.AddPillarParams{
		PillarEventParams: events.PillarEventParams{
			ConfigID:   m.ID(),
			TenantID:   m.tenantID.Value(),
			Version:    m.Version() + 1,
			PillarID:   pillarID.Value(),
			ModifiedBy: modifiedBy.Value(),
		},
		Name:        name.Value(),
		Description: description.Value(),
	})
	m.applyAndRaise(event)
	return nil
}

func (m *MetaModelConfiguration) UpdateStrategyPillar(id valueobjects.StrategyPillarID, name valueobjects.PillarName, description valueobjects.PillarDescription, modifiedBy valueobjects.UserEmail) error {
	if _, err := m.strategyPillarsConfig.WithUpdatedPillar(id, name, description); err != nil {
		return err
	}

	event := events.NewStrategyPillarUpdated(events.UpdatePillarParams{
		PillarEventParams: events.PillarEventParams{
			ConfigID:   m.ID(),
			TenantID:   m.tenantID.Value(),
			Version:    m.Version() + 1,
			PillarID:   id.Value(),
			ModifiedBy: modifiedBy.Value(),
		},
		NewName:        name.Value(),
		NewDescription: description.Value(),
	})
	m.applyAndRaise(event)
	return nil
}

func (m *MetaModelConfiguration) RemoveStrategyPillar(id valueobjects.StrategyPillarID, modifiedBy valueobjects.UserEmail) error {
	if _, err := m.strategyPillarsConfig.WithRemovedPillar(id); err != nil {
		return err
	}

	event := events.NewStrategyPillarRemoved(events.PillarEventParams{
		ConfigID:   m.ID(),
		TenantID:   m.tenantID.Value(),
		Version:    m.Version() + 1,
		PillarID:   id.Value(),
		ModifiedBy: modifiedBy.Value(),
	})
	m.applyAndRaise(event)
	return nil
}

func maturityScaleConfigToEventData(config valueobjects.MaturityScaleConfig) []events.MaturitySectionData {
	sections := config.Sections()
	data := make([]events.MaturitySectionData, 4)
	for i, section := range sections {
		data[i] = events.MaturitySectionData{
			Order:    section.Order().Value(),
			Name:     section.Name().Value(),
			MinValue: section.MinValue().Value(),
			MaxValue: section.MaxValue().Value(),
		}
	}
	return data
}

func eventDataToMaturityScaleConfig(data []events.MaturitySectionData) valueobjects.MaturityScaleConfig {
	var sections [4]valueobjects.MaturitySection
	for i, d := range data {
		order, _ := valueobjects.NewSectionOrder(d.Order)
		name, _ := valueobjects.NewSectionName(d.Name)
		minValue, _ := valueobjects.NewMaturityValue(d.MinValue)
		maxValue, _ := valueobjects.NewMaturityValue(d.MaxValue)
		sections[i], _ = valueobjects.NewMaturitySection(order, name, minValue, maxValue)
	}
	config, _ := valueobjects.NewMaturityScaleConfig(sections)
	return config
}

func strategyPillarsConfigToEventData(config valueobjects.StrategyPillarsConfig) []events.StrategyPillarData {
	pillars := config.Pillars()
	data := make([]events.StrategyPillarData, len(pillars))
	for i, pillar := range pillars {
		data[i] = events.StrategyPillarData{
			ID:          pillar.ID().Value(),
			Name:        pillar.Name().Value(),
			Description: pillar.Description().Value(),
			Active:      pillar.IsActive(),
		}
	}
	return data
}

func eventDataToStrategyPillarsConfig(data []events.StrategyPillarData) valueobjects.StrategyPillarsConfig {
	pillars := make([]valueobjects.StrategyPillar, len(data))
	for i, d := range data {
		id, _ := valueobjects.NewStrategyPillarIDFromString(d.ID)
		name, _ := valueobjects.NewPillarName(d.Name)
		desc, _ := valueobjects.NewPillarDescription(d.Description)
		if d.Active {
			pillars[i], _ = valueobjects.NewStrategyPillar(id, name, desc)
		} else {
			pillars[i], _ = valueobjects.NewInactiveStrategyPillar(id, name, desc)
		}
	}
	config, _ := valueobjects.NewStrategyPillarsConfig(pillars)
	return config
}
