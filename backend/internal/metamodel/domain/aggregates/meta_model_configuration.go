package aggregates

import (
	"easi/backend/internal/metamodel/domain/events"
	"easi/backend/internal/metamodel/domain/valueobjects"
	"easi/backend/internal/shared/eventsourcing"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

type MetaModelConfiguration struct {
	domain.AggregateRoot
	tenantID            sharedvo.TenantID
	maturityScaleConfig valueobjects.MaturityScaleConfig
	createdAt           valueobjects.Timestamp
	modifiedAt          valueobjects.Timestamp
	modifiedBy          valueobjects.UserEmail
}

func NewMetaModelConfiguration(tenantID sharedvo.TenantID, createdBy valueobjects.UserEmail) (*MetaModelConfiguration, error) {
	aggregate := &MetaModelConfiguration{
		AggregateRoot: domain.NewAggregateRoot(),
	}

	defaultConfig := valueobjects.DefaultMaturityScaleConfig()
	sections := maturityScaleConfigToEventData(defaultConfig)

	event := events.NewMetaModelConfigurationCreated(
		aggregate.ID(),
		tenantID.Value(),
		sections,
		createdBy.Value(),
	)

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
	}
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
