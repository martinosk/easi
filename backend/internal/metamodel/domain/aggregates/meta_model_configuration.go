package aggregates

import (
	"fmt"
	"time"

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

	if err := aggregate.apply(event); err != nil {
		return nil, err
	}
	aggregate.RaiseEvent(event)

	return aggregate, nil
}

func LoadMetaModelConfigurationFromHistory(eventHistory []domain.DomainEvent) (*MetaModelConfiguration, error) {
	aggregate := &MetaModelConfiguration{
		AggregateRoot: domain.NewAggregateRoot(),
	}

	var applyErr error
	aggregate.LoadFromHistory(eventHistory, func(event domain.DomainEvent) {
		if applyErr != nil {
			return
		}
		applyErr = aggregate.apply(event)
	})
	if applyErr != nil {
		return nil, applyErr
	}

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
	return m.applyAndRaise(event)
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
	return m.applyAndRaise(event)
}

func (m *MetaModelConfiguration) applyAndRaise(event domain.DomainEvent) error {
	if err := m.apply(event); err != nil {
		return err
	}
	m.RaiseEvent(event)
	return nil
}

func (m *MetaModelConfiguration) apply(event domain.DomainEvent) error {
	switch e := event.(type) {
	case events.MetaModelConfigurationCreated:
		return m.applyCreated(e)
	case events.MaturityScaleConfigUpdated:
		return m.applyMaturityScaleUpdated(e)
	case events.MaturityScaleConfigReset:
		return m.applyMaturityScaleReset(e)
	case events.StrategyPillarAdded:
		return m.applyPillarAdded(e)
	case events.StrategyPillarUpdated:
		return m.applyPillarUpdated(e)
	case events.StrategyPillarRemoved:
		return m.applyPillarRemoved(e)
	case events.PillarFitConfigurationUpdated:
		return m.applyPillarFitConfigurationUpdated(e)
	}
	return nil
}

func (m *MetaModelConfiguration) applyCreated(e events.MetaModelConfigurationCreated) error {
	tenantID, err := sharedvo.NewTenantID(e.TenantID)
	if err != nil {
		return fmt.Errorf("%w: tenant ID %q: %v", domain.ErrCorruptedEvent, e.TenantID, err)
	}
	maturityConfig, err := eventDataToMaturityScaleConfigSafe(e.Sections)
	if err != nil {
		return fmt.Errorf("%w: maturity scale config: %v", domain.ErrCorruptedEvent, err)
	}
	pillarsConfig, err := eventDataToStrategyPillarsConfigSafe(e.Pillars)
	if err != nil {
		return fmt.Errorf("%w: strategy pillars config: %v", domain.ErrCorruptedEvent, err)
	}
	createdAt, err := valueobjects.NewTimestamp(e.CreatedAt)
	if err != nil {
		return fmt.Errorf("%w: created at: %v", domain.ErrCorruptedEvent, err)
	}
	modifiedAt, err := valueobjects.NewTimestamp(e.CreatedAt)
	if err != nil {
		return fmt.Errorf("%w: modified at: %v", domain.ErrCorruptedEvent, err)
	}
	modifiedBy, err := valueobjects.NewUserEmail(e.CreatedBy)
	if err != nil {
		return fmt.Errorf("%w: created by %q: %v", domain.ErrCorruptedEvent, e.CreatedBy, err)
	}
	m.AggregateRoot = domain.NewAggregateRootWithID(e.ID)
	m.tenantID = tenantID
	m.maturityScaleConfig = maturityConfig
	m.strategyPillarsConfig = pillarsConfig
	m.createdAt = createdAt
	m.modifiedAt = modifiedAt
	m.modifiedBy = modifiedBy
	return nil
}

func (m *MetaModelConfiguration) applyMaturityScaleUpdated(e events.MaturityScaleConfigUpdated) error {
	maturityConfig, err := eventDataToMaturityScaleConfigSafe(e.NewSections)
	if err != nil {
		return fmt.Errorf("%w: maturity scale config: %v", domain.ErrCorruptedEvent, err)
	}
	m.maturityScaleConfig = maturityConfig
	return m.applyModificationMetadata(e.ModifiedAt, e.ModifiedBy)
}

func (m *MetaModelConfiguration) applyMaturityScaleReset(e events.MaturityScaleConfigReset) error {
	maturityConfig, err := eventDataToMaturityScaleConfigSafe(e.Sections)
	if err != nil {
		return fmt.Errorf("%w: maturity scale config: %v", domain.ErrCorruptedEvent, err)
	}
	m.maturityScaleConfig = maturityConfig
	return m.applyModificationMetadata(e.ModifiedAt, e.ModifiedBy)
}

func (m *MetaModelConfiguration) applyPillarAdded(e events.StrategyPillarAdded) error {
	pillarID, err := valueobjects.NewStrategyPillarIDFromString(e.PillarID)
	if err != nil {
		return fmt.Errorf("%w: pillar ID %q: %v", domain.ErrCorruptedEvent, e.PillarID, err)
	}
	pillarName, err := valueobjects.NewPillarName(e.Name)
	if err != nil {
		return fmt.Errorf("%w: pillar name %q: %v", domain.ErrCorruptedEvent, e.Name, err)
	}
	pillarDesc, err := valueobjects.NewPillarDescription(e.Description)
	if err != nil {
		return fmt.Errorf("%w: pillar description: %v", domain.ErrCorruptedEvent, err)
	}
	pillar, err := valueobjects.NewStrategyPillar(pillarID, pillarName, pillarDesc)
	if err != nil {
		return fmt.Errorf("%w: strategy pillar: %v", domain.ErrCorruptedEvent, err)
	}
	config, err := m.strategyPillarsConfig.WithAddedPillar(pillar)
	if err != nil {
		return fmt.Errorf("%w: adding pillar: %v", domain.ErrCorruptedEvent, err)
	}
	m.strategyPillarsConfig = config
	return m.applyModificationMetadata(e.ModifiedAt, e.ModifiedBy)
}

func (m *MetaModelConfiguration) applyPillarUpdated(e events.StrategyPillarUpdated) error {
	pillarID, err := valueobjects.NewStrategyPillarIDFromString(e.PillarID)
	if err != nil {
		return fmt.Errorf("%w: pillar ID %q: %v", domain.ErrCorruptedEvent, e.PillarID, err)
	}
	newName, err := valueobjects.NewPillarName(e.NewName)
	if err != nil {
		return fmt.Errorf("%w: pillar name %q: %v", domain.ErrCorruptedEvent, e.NewName, err)
	}
	newDesc, err := valueobjects.NewPillarDescription(e.NewDescription)
	if err != nil {
		return fmt.Errorf("%w: pillar description: %v", domain.ErrCorruptedEvent, err)
	}
	config, err := m.strategyPillarsConfig.WithUpdatedPillar(pillarID, newName, newDesc)
	if err != nil {
		return fmt.Errorf("%w: updating pillar: %v", domain.ErrCorruptedEvent, err)
	}
	m.strategyPillarsConfig = config
	return m.applyModificationMetadata(e.ModifiedAt, e.ModifiedBy)
}

func (m *MetaModelConfiguration) applyPillarRemoved(e events.StrategyPillarRemoved) error {
	pillarID, err := valueobjects.NewStrategyPillarIDFromString(e.PillarID)
	if err != nil {
		return fmt.Errorf("%w: pillar ID %q: %v", domain.ErrCorruptedEvent, e.PillarID, err)
	}
	config, err := m.strategyPillarsConfig.WithRemovedPillar(pillarID)
	if err != nil {
		return fmt.Errorf("%w: removing pillar: %v", domain.ErrCorruptedEvent, err)
	}
	m.strategyPillarsConfig = config
	return m.applyModificationMetadata(e.ModifiedAt, e.ModifiedBy)
}

func (m *MetaModelConfiguration) applyPillarFitConfigurationUpdated(e events.PillarFitConfigurationUpdated) error {
	pillarID, err := valueobjects.NewStrategyPillarIDFromString(e.PillarID)
	if err != nil {
		return fmt.Errorf("%w: pillar ID %q: %v", domain.ErrCorruptedEvent, e.PillarID, err)
	}
	criteria, err := valueobjects.NewFitCriteria(e.FitCriteria)
	if err != nil {
		return fmt.Errorf("%w: fit criteria %q: %v", domain.ErrCorruptedEvent, e.FitCriteria, err)
	}
	fitType, err := valueobjects.NewFitType(e.FitType)
	if err != nil {
		return fmt.Errorf("%w: fit type %q: %v", domain.ErrCorruptedEvent, e.FitType, err)
	}
	config, err := m.strategyPillarsConfig.WithUpdatedPillarFitConfiguration(pillarID, e.FitScoringEnabled, criteria, fitType)
	if err != nil {
		return fmt.Errorf("%w: updating pillar fit configuration: %v", domain.ErrCorruptedEvent, err)
	}
	m.strategyPillarsConfig = config
	return m.applyModificationMetadata(e.ModifiedAt, e.ModifiedBy)
}

func (m *MetaModelConfiguration) applyModificationMetadata(modifiedAtRaw time.Time, modifiedByRaw string) error {
	modifiedAt, err := valueobjects.NewTimestamp(modifiedAtRaw)
	if err != nil {
		return fmt.Errorf("%w: modified at: %v", domain.ErrCorruptedEvent, err)
	}
	modifiedBy, err := valueobjects.NewUserEmail(modifiedByRaw)
	if err != nil {
		return fmt.Errorf("%w: modified by %q: %v", domain.ErrCorruptedEvent, modifiedByRaw, err)
	}
	m.modifiedAt = modifiedAt
	m.modifiedBy = modifiedBy
	return nil
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
	return m.applyAndRaise(event)
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
	return m.applyAndRaise(event)
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
	return m.applyAndRaise(event)
}

func (m *MetaModelConfiguration) UpdatePillarFitConfiguration(id valueobjects.StrategyPillarID, fitConfig valueobjects.FitConfigurationParams, modifiedBy valueobjects.UserEmail) error {
	if _, err := m.strategyPillarsConfig.WithUpdatedPillarFitConfiguration(id, fitConfig.Enabled(), fitConfig.Criteria(), fitConfig.FitType()); err != nil {
		return err
	}

	event := events.NewPillarFitConfigurationUpdated(events.UpdatePillarFitConfigParams{
		PillarEventParams: events.PillarEventParams{
			ConfigID:   m.ID(),
			TenantID:   m.tenantID.Value(),
			Version:    m.Version() + 1,
			PillarID:   id.Value(),
			ModifiedBy: modifiedBy.Value(),
		},
		FitScoringEnabled: fitConfig.Enabled(),
		FitCriteria:       fitConfig.Criteria().Value(),
		FitType:           fitConfig.FitType().Value(),
	})
	return m.applyAndRaise(event)
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
	config, _ := eventDataToMaturityScaleConfigSafe(data)
	return config
}

func eventDataToMaturityScaleConfigSafe(data []events.MaturitySectionData) (valueobjects.MaturityScaleConfig, error) {
	var sections [4]valueobjects.MaturitySection
	for i, d := range data {
		order, err := valueobjects.NewSectionOrder(d.Order)
		if err != nil {
			return valueobjects.MaturityScaleConfig{}, fmt.Errorf("section %d order: %v", i, err)
		}
		name, err := valueobjects.NewSectionName(d.Name)
		if err != nil {
			return valueobjects.MaturityScaleConfig{}, fmt.Errorf("section %d name %q: %v", i, d.Name, err)
		}
		minValue, err := valueobjects.NewMaturityValue(d.MinValue)
		if err != nil {
			return valueobjects.MaturityScaleConfig{}, fmt.Errorf("section %d min value: %v", i, err)
		}
		maxValue, err := valueobjects.NewMaturityValue(d.MaxValue)
		if err != nil {
			return valueobjects.MaturityScaleConfig{}, fmt.Errorf("section %d max value: %v", i, err)
		}
		sections[i], err = valueobjects.NewMaturitySection(order, name, minValue, maxValue)
		if err != nil {
			return valueobjects.MaturityScaleConfig{}, fmt.Errorf("section %d: %v", i, err)
		}
	}
	config, err := valueobjects.NewMaturityScaleConfig(sections)
	if err != nil {
		return valueobjects.MaturityScaleConfig{}, err
	}
	return config, nil
}

func strategyPillarsConfigToEventData(config valueobjects.StrategyPillarsConfig) []events.StrategyPillarData {
	pillars := config.Pillars()
	data := make([]events.StrategyPillarData, len(pillars))
	for i, pillar := range pillars {
		data[i] = events.StrategyPillarData{
			ID:                pillar.ID().Value(),
			Name:              pillar.Name().Value(),
			Description:       pillar.Description().Value(),
			Active:            pillar.IsActive(),
			FitScoringEnabled: pillar.FitScoringEnabled(),
			FitCriteria:       pillar.FitCriteria().Value(),
			FitType:           pillar.FitType().Value(),
		}
	}
	return data
}

func eventDataToStrategyPillarsConfig(data []events.StrategyPillarData) valueobjects.StrategyPillarsConfig {
	config, _ := eventDataToStrategyPillarsConfigSafe(data)
	return config
}

func eventDataToStrategyPillarsConfigSafe(data []events.StrategyPillarData) (valueobjects.StrategyPillarsConfig, error) {
	pillars := make([]valueobjects.StrategyPillar, len(data))
	for i, d := range data {
		pillar, err := convertPillarEventData(d)
		if err != nil {
			return valueobjects.StrategyPillarsConfig{}, fmt.Errorf("pillar %d: %v", i, err)
		}
		pillars[i] = pillar
	}
	config, err := valueobjects.NewStrategyPillarsConfig(pillars)
	if err != nil {
		return valueobjects.StrategyPillarsConfig{}, err
	}
	return config, nil
}

func convertPillarEventData(d events.StrategyPillarData) (valueobjects.StrategyPillar, error) {
	id, err := valueobjects.NewStrategyPillarIDFromString(d.ID)
	if err != nil {
		return valueobjects.StrategyPillar{}, fmt.Errorf("ID %q: %v", d.ID, err)
	}
	name, err := valueobjects.NewPillarName(d.Name)
	if err != nil {
		return valueobjects.StrategyPillar{}, fmt.Errorf("name %q: %v", d.Name, err)
	}
	desc, err := valueobjects.NewPillarDescription(d.Description)
	if err != nil {
		return valueobjects.StrategyPillar{}, fmt.Errorf("description: %v", err)
	}
	criteria, err := valueobjects.NewFitCriteria(d.FitCriteria)
	if err != nil {
		return valueobjects.StrategyPillar{}, fmt.Errorf("fit criteria %q: %v", d.FitCriteria, err)
	}
	fitType, err := valueobjects.NewFitType(d.FitType)
	if err != nil {
		return valueobjects.StrategyPillar{}, fmt.Errorf("fit type %q: %v", d.FitType, err)
	}
	var pillar valueobjects.StrategyPillar
	if d.Active {
		pillar, err = valueobjects.NewStrategyPillar(id, name, desc)
	} else {
		pillar, err = valueobjects.NewInactiveStrategyPillar(id, name, desc)
	}
	if err != nil {
		return valueobjects.StrategyPillar{}, err
	}
	return pillar.WithFitConfiguration(d.FitScoringEnabled, criteria, fitType), nil
}
