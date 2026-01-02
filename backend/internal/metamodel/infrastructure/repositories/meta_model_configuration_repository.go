package repositories

import (
	"errors"
	"time"

	"easi/backend/internal/infrastructure/eventstore"
	"easi/backend/internal/metamodel/domain/aggregates"
	"easi/backend/internal/metamodel/domain/events"
	domain "easi/backend/internal/shared/eventsourcing"
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

var metaModelEventDeserializers = repository.NewEventDeserializers(
	map[string]repository.EventDeserializerFunc{
		"MetaModelConfigurationCreated": func(data map[string]interface{}) domain.DomainEvent {
			id, _ := data["id"].(string)
			tenantID, _ := data["tenantId"].(string)
			createdBy, _ := data["createdBy"].(string)
			createdAtStr, _ := data["createdAt"].(string)
			createdAt, _ := time.Parse(time.RFC3339Nano, createdAtStr)

			sectionsRaw, _ := data["sections"].([]interface{})
			sections := deserializeSections(sectionsRaw)

			pillarsRaw, _ := data["pillars"].([]interface{})
			pillars := deserializePillars(pillarsRaw)

			evt := events.NewMetaModelConfigurationCreated(events.CreateConfigParams{
				ID:        id,
				TenantID:  tenantID,
				Sections:  sections,
				Pillars:   pillars,
				CreatedBy: createdBy,
			})
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
		"StrategyPillarAdded": func(data map[string]interface{}) domain.DomainEvent {
			id, _ := data["id"].(string)
			tenantID, _ := data["tenantId"].(string)
			version, _ := data["version"].(float64)
			pillarID, _ := data["pillarId"].(string)
			name, _ := data["name"].(string)
			description, _ := data["description"].(string)
			modifiedBy, _ := data["modifiedBy"].(string)
			modifiedAtStr, _ := data["modifiedAt"].(string)
			modifiedAt, _ := time.Parse(time.RFC3339Nano, modifiedAtStr)

			evt := events.NewStrategyPillarAdded(events.AddPillarParams{
				PillarEventParams: events.PillarEventParams{
					ConfigID:   id,
					TenantID:   tenantID,
					Version:    int(version),
					PillarID:   pillarID,
					ModifiedBy: modifiedBy,
				},
				Name:        name,
				Description: description,
			})
			evt.ModifiedAt = modifiedAt
			return evt
		},
		"StrategyPillarUpdated": func(data map[string]interface{}) domain.DomainEvent {
			id, _ := data["id"].(string)
			tenantID, _ := data["tenantId"].(string)
			version, _ := data["version"].(float64)
			pillarID, _ := data["pillarId"].(string)
			newName, _ := data["newName"].(string)
			newDescription, _ := data["newDescription"].(string)
			modifiedBy, _ := data["modifiedBy"].(string)
			modifiedAtStr, _ := data["modifiedAt"].(string)
			modifiedAt, _ := time.Parse(time.RFC3339Nano, modifiedAtStr)

			evt := events.NewStrategyPillarUpdated(events.UpdatePillarParams{
				PillarEventParams: events.PillarEventParams{
					ConfigID:   id,
					TenantID:   tenantID,
					Version:    int(version),
					PillarID:   pillarID,
					ModifiedBy: modifiedBy,
				},
				NewName:        newName,
				NewDescription: newDescription,
			})
			evt.ModifiedAt = modifiedAt
			return evt
		},
		"StrategyPillarRemoved": func(data map[string]interface{}) domain.DomainEvent {
			id, _ := data["id"].(string)
			tenantID, _ := data["tenantId"].(string)
			version, _ := data["version"].(float64)
			pillarID, _ := data["pillarId"].(string)
			modifiedBy, _ := data["modifiedBy"].(string)
			modifiedAtStr, _ := data["modifiedAt"].(string)
			modifiedAt, _ := time.Parse(time.RFC3339Nano, modifiedAtStr)

			evt := events.NewStrategyPillarRemoved(events.PillarEventParams{
				ConfigID:   id,
				TenantID:   tenantID,
				Version:    int(version),
				PillarID:   pillarID,
				ModifiedBy: modifiedBy,
			})
			evt.ModifiedAt = modifiedAt
			return evt
		},
		"PillarFitConfigurationUpdated": func(data map[string]interface{}) domain.DomainEvent {
			id, _ := data["id"].(string)
			tenantID, _ := data["tenantId"].(string)
			version, _ := data["version"].(float64)
			pillarID, _ := data["pillarId"].(string)
			fitScoringEnabled, _ := data["fitScoringEnabled"].(bool)
			fitCriteria, _ := data["fitCriteria"].(string)
			modifiedBy, _ := data["modifiedBy"].(string)
			modifiedAtStr, _ := data["modifiedAt"].(string)
			modifiedAt, _ := time.Parse(time.RFC3339Nano, modifiedAtStr)

			evt := events.NewPillarFitConfigurationUpdated(events.UpdatePillarFitConfigParams{
				PillarEventParams: events.PillarEventParams{
					ConfigID:   id,
					TenantID:   tenantID,
					Version:    int(version),
					PillarID:   pillarID,
					ModifiedBy: modifiedBy,
				},
				FitScoringEnabled: fitScoringEnabled,
				FitCriteria:       fitCriteria,
			})
			evt.ModifiedAt = modifiedAt
			return evt
		},
	},
)

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

func deserializePillars(pillarsRaw []interface{}) []events.StrategyPillarData {
	pillars := make([]events.StrategyPillarData, 0, len(pillarsRaw))
	for _, p := range pillarsRaw {
		pillarMap, ok := p.(map[string]interface{})
		if !ok {
			continue
		}
		id, _ := pillarMap["id"].(string)
		name, _ := pillarMap["name"].(string)
		description, _ := pillarMap["description"].(string)
		active, _ := pillarMap["active"].(bool)

		pillars = append(pillars, events.StrategyPillarData{
			ID:          id,
			Name:        name,
			Description: description,
			Active:      active,
		})
	}
	return pillars
}
