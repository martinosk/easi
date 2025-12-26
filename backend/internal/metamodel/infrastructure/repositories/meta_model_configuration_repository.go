package repositories

import (
	"context"
	"errors"
	"time"

	"easi/backend/internal/infrastructure/eventstore"
	"easi/backend/internal/metamodel/domain/aggregates"
	"easi/backend/internal/metamodel/domain/events"
	"easi/backend/internal/shared/eventsourcing"
	"easi/backend/internal/shared/infrastructure/repository"
)

var ErrMetaModelConfigurationNotFound = errors.New("meta model configuration not found")

type MetaModelConfigurationRepository struct {
	*repository.EventSourcedRepository[*aggregates.MetaModelConfiguration]
}

func NewMetaModelConfigurationRepository(eventStore eventstore.EventStore) *MetaModelConfigurationRepository {
	return &MetaModelConfigurationRepository{
		EventSourcedRepository: repository.NewEventSourcedRepository(
			eventStore,
			metaModelEventDeserializers,
			aggregates.LoadMetaModelConfigurationFromHistory,
			ErrMetaModelConfigurationNotFound,
		),
	}
}

func (r *MetaModelConfigurationRepository) GetByTenantID(ctx context.Context, tenantID string) (*aggregates.MetaModelConfiguration, error) {
	return nil, errors.New("not implemented - requires read model lookup")
}

var metaModelEventDeserializers = repository.EventDeserializers{
	"MetaModelConfigurationCreated": func(data map[string]interface{}) domain.DomainEvent {
		id, _ := data["id"].(string)
		tenantID, _ := data["tenantId"].(string)
		createdBy, _ := data["createdBy"].(string)
		createdAtStr, _ := data["createdAt"].(string)
		createdAt, _ := time.Parse(time.RFC3339Nano, createdAtStr)

		sectionsRaw, _ := data["sections"].([]interface{})
		sections := deserializeSections(sectionsRaw)

		evt := events.NewMetaModelConfigurationCreated(id, tenantID, sections, createdBy)
		evt.CreatedAt = createdAt
		return evt
	},
	"MaturityScaleConfigUpdated": func(data map[string]interface{}) domain.DomainEvent {
		id, _ := data["id"].(string)
		tenantID, _ := data["tenantId"].(string)
		version, _ := data["version"].(float64)
		modifiedBy, _ := data["modifiedBy"].(string)
		modifiedAtStr, _ := data["modifiedAt"].(string)
		modifiedAt, _ := time.Parse(time.RFC3339Nano, modifiedAtStr)

		sectionsRaw, _ := data["newSections"].([]interface{})
		sections := deserializeSections(sectionsRaw)

		evt := events.NewMaturityScaleConfigUpdated(id, tenantID, int(version), sections, modifiedBy)
		evt.ModifiedAt = modifiedAt
		return evt
	},
	"MaturityScaleConfigReset": func(data map[string]interface{}) domain.DomainEvent {
		id, _ := data["id"].(string)
		tenantID, _ := data["tenantId"].(string)
		version, _ := data["version"].(float64)
		modifiedBy, _ := data["modifiedBy"].(string)
		modifiedAtStr, _ := data["modifiedAt"].(string)
		modifiedAt, _ := time.Parse(time.RFC3339Nano, modifiedAtStr)

		sectionsRaw, _ := data["sections"].([]interface{})
		sections := deserializeSections(sectionsRaw)

		evt := events.NewMaturityScaleConfigReset(id, tenantID, int(version), sections, modifiedBy)
		evt.ModifiedAt = modifiedAt
		return evt
	},
}

func deserializeSections(sectionsRaw []interface{}) []events.MaturitySectionData {
	sections := make([]events.MaturitySectionData, 0, len(sectionsRaw))
	for _, s := range sectionsRaw {
		sectionMap, ok := s.(map[string]interface{})
		if !ok {
			continue
		}
		order, _ := sectionMap["order"].(float64)
		name, _ := sectionMap["name"].(string)
		minValue, _ := sectionMap["minValue"].(float64)
		maxValue, _ := sectionMap["maxValue"].(float64)

		sections = append(sections, events.MaturitySectionData{
			Order:    int(order),
			Name:     name,
			MinValue: int(minValue),
			MaxValue: int(maxValue),
		})
	}
	return sections
}
