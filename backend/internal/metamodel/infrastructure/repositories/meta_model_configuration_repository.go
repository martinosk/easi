package repositories

import (
	"errors"

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
		"MetaModelConfigurationCreated": func(data map[string]interface{}) (domain.DomainEvent, error) {
			id, err := repository.GetRequiredString(data, "id")
			if err != nil {
				return nil, err
			}
			tenantID, err := repository.GetRequiredString(data, "tenantId")
			if err != nil {
				return nil, err
			}
			createdBy, err := repository.GetRequiredString(data, "createdBy")
			if err != nil {
				return nil, err
			}
			createdAt, err := repository.GetRequiredTime(data, "createdAt")
			if err != nil {
				return nil, err
			}

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
			return evt, nil
		},
		"MaturityScaleConfigUpdated": func(data map[string]interface{}) (domain.DomainEvent, error) {
			id, err := repository.GetRequiredString(data, "id")
			if err != nil {
				return nil, err
			}
			tenantID, err := repository.GetRequiredString(data, "tenantId")
			if err != nil {
				return nil, err
			}
			version, err := repository.GetRequiredInt(data, "version")
			if err != nil {
				return nil, err
			}
			modifiedBy, err := repository.GetRequiredString(data, "modifiedBy")
			if err != nil {
				return nil, err
			}
			modifiedAt, err := repository.GetRequiredTime(data, "modifiedAt")
			if err != nil {
				return nil, err
			}

			sectionsRaw, _ := data["newSections"].([]interface{})
			sections := deserializeSections(sectionsRaw)

			evt := events.NewMaturityScaleConfigUpdated(id, tenantID, version, sections, modifiedBy)
			evt.ModifiedAt = modifiedAt
			return evt, nil
		},
		"MaturityScaleConfigReset": func(data map[string]interface{}) (domain.DomainEvent, error) {
			id, err := repository.GetRequiredString(data, "id")
			if err != nil {
				return nil, err
			}
			tenantID, err := repository.GetRequiredString(data, "tenantId")
			if err != nil {
				return nil, err
			}
			version, err := repository.GetRequiredInt(data, "version")
			if err != nil {
				return nil, err
			}
			modifiedBy, err := repository.GetRequiredString(data, "modifiedBy")
			if err != nil {
				return nil, err
			}
			modifiedAt, err := repository.GetRequiredTime(data, "modifiedAt")
			if err != nil {
				return nil, err
			}

			sectionsRaw, _ := data["sections"].([]interface{})
			sections := deserializeSections(sectionsRaw)

			evt := events.NewMaturityScaleConfigReset(id, tenantID, version, sections, modifiedBy)
			evt.ModifiedAt = modifiedAt
			return evt, nil
		},
		"StrategyPillarAdded": func(data map[string]interface{}) (domain.DomainEvent, error) {
			id, err := repository.GetRequiredString(data, "id")
			if err != nil {
				return nil, err
			}
			tenantID, err := repository.GetRequiredString(data, "tenantId")
			if err != nil {
				return nil, err
			}
			version, err := repository.GetRequiredInt(data, "version")
			if err != nil {
				return nil, err
			}
			pillarID, err := repository.GetRequiredString(data, "pillarId")
			if err != nil {
				return nil, err
			}
			name, err := repository.GetRequiredString(data, "name")
			if err != nil {
				return nil, err
			}
			description, err := repository.GetOptionalString(data, "description", "")
			if err != nil {
				return nil, err
			}
			modifiedBy, err := repository.GetRequiredString(data, "modifiedBy")
			if err != nil {
				return nil, err
			}
			modifiedAt, err := repository.GetRequiredTime(data, "modifiedAt")
			if err != nil {
				return nil, err
			}

			evt := events.NewStrategyPillarAdded(events.AddPillarParams{
				PillarEventParams: events.PillarEventParams{
					ConfigID:   id,
					TenantID:   tenantID,
					Version:    version,
					PillarID:   pillarID,
					ModifiedBy: modifiedBy,
				},
				Name:        name,
				Description: description,
			})
			evt.ModifiedAt = modifiedAt
			return evt, nil
		},
		"StrategyPillarUpdated": func(data map[string]interface{}) (domain.DomainEvent, error) {
			id, err := repository.GetRequiredString(data, "id")
			if err != nil {
				return nil, err
			}
			tenantID, err := repository.GetRequiredString(data, "tenantId")
			if err != nil {
				return nil, err
			}
			version, err := repository.GetRequiredInt(data, "version")
			if err != nil {
				return nil, err
			}
			pillarID, err := repository.GetRequiredString(data, "pillarId")
			if err != nil {
				return nil, err
			}
			newName, err := repository.GetRequiredString(data, "newName")
			if err != nil {
				return nil, err
			}
			newDescription, err := repository.GetOptionalString(data, "newDescription", "")
			if err != nil {
				return nil, err
			}
			modifiedBy, err := repository.GetRequiredString(data, "modifiedBy")
			if err != nil {
				return nil, err
			}
			modifiedAt, err := repository.GetRequiredTime(data, "modifiedAt")
			if err != nil {
				return nil, err
			}

			evt := events.NewStrategyPillarUpdated(events.UpdatePillarParams{
				PillarEventParams: events.PillarEventParams{
					ConfigID:   id,
					TenantID:   tenantID,
					Version:    version,
					PillarID:   pillarID,
					ModifiedBy: modifiedBy,
				},
				NewName:        newName,
				NewDescription: newDescription,
			})
			evt.ModifiedAt = modifiedAt
			return evt, nil
		},
		"StrategyPillarRemoved": func(data map[string]interface{}) (domain.DomainEvent, error) {
			id, err := repository.GetRequiredString(data, "id")
			if err != nil {
				return nil, err
			}
			tenantID, err := repository.GetRequiredString(data, "tenantId")
			if err != nil {
				return nil, err
			}
			version, err := repository.GetRequiredInt(data, "version")
			if err != nil {
				return nil, err
			}
			pillarID, err := repository.GetRequiredString(data, "pillarId")
			if err != nil {
				return nil, err
			}
			modifiedBy, err := repository.GetRequiredString(data, "modifiedBy")
			if err != nil {
				return nil, err
			}
			modifiedAt, err := repository.GetRequiredTime(data, "modifiedAt")
			if err != nil {
				return nil, err
			}

			evt := events.NewStrategyPillarRemoved(events.PillarEventParams{
				ConfigID:   id,
				TenantID:   tenantID,
				Version:    version,
				PillarID:   pillarID,
				ModifiedBy: modifiedBy,
			})
			evt.ModifiedAt = modifiedAt
			return evt, nil
		},
		"PillarFitConfigurationUpdated": func(data map[string]interface{}) (domain.DomainEvent, error) {
			id, err := repository.GetRequiredString(data, "id")
			if err != nil {
				return nil, err
			}
			tenantID, err := repository.GetRequiredString(data, "tenantId")
			if err != nil {
				return nil, err
			}
			version, err := repository.GetRequiredInt(data, "version")
			if err != nil {
				return nil, err
			}
			pillarID, err := repository.GetRequiredString(data, "pillarId")
			if err != nil {
				return nil, err
			}
			fitScoringEnabled, err := repository.GetRequiredBool(data, "fitScoringEnabled")
			if err != nil {
				return nil, err
			}
			fitCriteria, err := repository.GetOptionalString(data, "fitCriteria", "")
			if err != nil {
				return nil, err
			}
			modifiedBy, err := repository.GetRequiredString(data, "modifiedBy")
			if err != nil {
				return nil, err
			}
			modifiedAt, err := repository.GetRequiredTime(data, "modifiedAt")
			if err != nil {
				return nil, err
			}

			evt := events.NewPillarFitConfigurationUpdated(events.UpdatePillarFitConfigParams{
				PillarEventParams: events.PillarEventParams{
					ConfigID:   id,
					TenantID:   tenantID,
					Version:    version,
					PillarID:   pillarID,
					ModifiedBy: modifiedBy,
				},
				FitScoringEnabled: fitScoringEnabled,
				FitCriteria:       fitCriteria,
			})
			evt.ModifiedAt = modifiedAt
			return evt, nil
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
