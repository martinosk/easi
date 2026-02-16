package projectors

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"easi/backend/internal/metamodel/application/readmodels"
	"easi/backend/internal/metamodel/domain/events"
	mmPL "easi/backend/internal/metamodel/publishedlanguage"
	domain "easi/backend/internal/shared/eventsourcing"
)

type MetaModelConfigurationProjector struct {
	readModel *readmodels.MetaModelConfigurationReadModel
}

func NewMetaModelConfigurationProjector(readModel *readmodels.MetaModelConfigurationReadModel) *MetaModelConfigurationProjector {
	return &MetaModelConfigurationProjector{
		readModel: readModel,
	}
}

func (p *MetaModelConfigurationProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		log.Printf("Failed to marshal event data: %v", err)
		return err
	}
	return p.ProjectEvent(ctx, event.EventType(), eventData)
}

func (p *MetaModelConfigurationProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	switch eventType {
	case mmPL.MetaModelConfigurationCreated:
		return p.handleCreated(ctx, eventData)
	case mmPL.MaturityScaleConfigUpdated:
		return p.handleUpdated(ctx, eventData)
	case mmPL.MaturityScaleConfigReset:
		return p.handleReset(ctx, eventData)
	case mmPL.StrategyPillarAdded:
		return p.handlePillarAdded(ctx, eventData)
	case mmPL.StrategyPillarUpdated:
		return p.handlePillarUpdated(ctx, eventData)
	case mmPL.StrategyPillarRemoved:
		return p.handlePillarRemoved(ctx, eventData)
	case mmPL.PillarFitConfigurationUpdated:
		return p.handlePillarFitConfigurationUpdated(ctx, eventData)
	}
	return nil
}

func unmarshalAndProject[T any](eventData []byte, eventName string, project func(*T) error) error {
	var event T
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal %s event: %v", eventName, err)
		return err
	}
	return project(&event)
}

func (p *MetaModelConfigurationProjector) handleCreated(ctx context.Context, eventData []byte) error {
	return unmarshalAndProject(eventData, "MetaModelConfigurationCreated", func(event *events.MetaModelConfigurationCreated) error {
		dto := readmodels.MetaModelConfigurationDTO{
			ID:              event.ID,
			TenantID:        event.TenantID,
			Sections:        toSectionDTOs(event.Sections),
			StrategyPillars: toPillarDTOs(event.Pillars),
			Version:         1,
			IsDefault:       true,
			CreatedAt:       event.CreatedAt,
			ModifiedAt:      event.CreatedAt,
			ModifiedBy:      event.CreatedBy,
		}
		return p.readModel.Insert(ctx, dto)
	})
}

func (p *MetaModelConfigurationProjector) handleUpdated(ctx context.Context, eventData []byte) error {
	return unmarshalAndProject(eventData, "MaturityScaleConfigUpdated", func(event *events.MaturityScaleConfigUpdated) error {
		return p.updateMaturitySections(ctx, maturityUpdateData{
			ID: event.ID, Sections: toSectionDTOs(event.NewSections), Version: event.Version, IsDefault: false, ModifiedAt: event.ModifiedAt, ModifiedBy: event.ModifiedBy,
		})
	})
}

func (p *MetaModelConfigurationProjector) handleReset(ctx context.Context, eventData []byte) error {
	return unmarshalAndProject(eventData, "MaturityScaleConfigReset", func(event *events.MaturityScaleConfigReset) error {
		return p.updateMaturitySections(ctx, maturityUpdateData{
			ID: event.ID, Sections: toSectionDTOs(event.Sections), Version: event.Version, IsDefault: true, ModifiedAt: event.ModifiedAt, ModifiedBy: event.ModifiedBy,
		})
	})
}

type maturityUpdateData struct {
	ID         string
	Sections   []readmodels.MaturitySectionDTO
	Version    int
	IsDefault  bool
	ModifiedAt time.Time
	ModifiedBy string
}

func (p *MetaModelConfigurationProjector) updateMaturitySections(ctx context.Context, data maturityUpdateData) error {
	return p.readModel.Update(ctx, readmodels.UpdateParams{
		ID:         data.ID,
		Sections:   data.Sections,
		Version:    data.Version,
		IsDefault:  data.IsDefault,
		ModifiedAt: data.ModifiedAt,
		ModifiedBy: data.ModifiedBy,
	})
}

type pillarEventData struct {
	ID         string
	Version    int
	ModifiedAt time.Time
	ModifiedBy string
}

func (p *MetaModelConfigurationProjector) handlePillarAdded(ctx context.Context, eventData []byte) error {
	return unmarshalAndProject(eventData, "StrategyPillarAdded", func(event *events.StrategyPillarAdded) error {
		data := pillarEventData{ID: event.ID, Version: event.Version, ModifiedAt: event.ModifiedAt, ModifiedBy: event.ModifiedBy}
		return p.updatePillarsWithModification(ctx, data, func(pillars []readmodels.StrategyPillarDTO) []readmodels.StrategyPillarDTO {
			return append(pillars, readmodels.StrategyPillarDTO{ID: event.PillarID, Name: event.Name, Description: event.Description, Active: true})
		})
	})
}

func (p *MetaModelConfigurationProjector) handlePillarUpdated(ctx context.Context, eventData []byte) error {
	return unmarshalAndProject(eventData, "StrategyPillarUpdated", func(event *events.StrategyPillarUpdated) error {
		data := pillarEventData{ID: event.ID, Version: event.Version, ModifiedAt: event.ModifiedAt, ModifiedBy: event.ModifiedBy}
		return p.updatePillarByID(ctx, data, event.PillarID, func(pillar *readmodels.StrategyPillarDTO) {
			pillar.Name, pillar.Description = event.NewName, event.NewDescription
		})
	})
}

func (p *MetaModelConfigurationProjector) handlePillarRemoved(ctx context.Context, eventData []byte) error {
	return unmarshalAndProject(eventData, "StrategyPillarRemoved", func(event *events.StrategyPillarRemoved) error {
		data := pillarEventData{ID: event.ID, Version: event.Version, ModifiedAt: event.ModifiedAt, ModifiedBy: event.ModifiedBy}
		return p.updatePillarByID(ctx, data, event.PillarID, func(pillar *readmodels.StrategyPillarDTO) {
			pillar.Active = false
		})
	})
}

func (p *MetaModelConfigurationProjector) handlePillarFitConfigurationUpdated(ctx context.Context, eventData []byte) error {
	return unmarshalAndProject(eventData, "PillarFitConfigurationUpdated", func(event *events.PillarFitConfigurationUpdated) error {
		data := pillarEventData{ID: event.ID, Version: event.Version, ModifiedAt: event.ModifiedAt, ModifiedBy: event.ModifiedBy}
		return p.updatePillarByID(ctx, data, event.PillarID, func(pillar *readmodels.StrategyPillarDTO) {
			pillar.FitScoringEnabled, pillar.FitCriteria, pillar.FitType = event.FitScoringEnabled, event.FitCriteria, event.FitType
		})
	})
}

func (p *MetaModelConfigurationProjector) updatePillarByID(
	ctx context.Context,
	eventData pillarEventData,
	pillarID string,
	modify func(*readmodels.StrategyPillarDTO),
) error {
	return p.updatePillarsWithModification(ctx, eventData, func(pillars []readmodels.StrategyPillarDTO) []readmodels.StrategyPillarDTO {
		for i := range pillars {
			if pillars[i].ID == pillarID {
				modify(&pillars[i])
				break
			}
		}
		return pillars
	})
}

func (p *MetaModelConfigurationProjector) updatePillarsWithModification(
	ctx context.Context,
	eventData pillarEventData,
	modify func([]readmodels.StrategyPillarDTO) []readmodels.StrategyPillarDTO,
) error {
	config, err := p.readModel.GetByID(ctx, eventData.ID)
	if err != nil {
		return err
	}
	if config == nil {
		log.Printf("Configuration not found for ID: %s", eventData.ID)
		return nil
	}

	config.StrategyPillars = modify(config.StrategyPillars)

	params := readmodels.UpdateParams{
		ID:              eventData.ID,
		Sections:        config.Sections,
		StrategyPillars: config.StrategyPillars,
		Version:         eventData.Version,
		IsDefault:       config.IsDefault,
		ModifiedAt:      eventData.ModifiedAt,
		ModifiedBy:      eventData.ModifiedBy,
	}
	return p.readModel.Update(ctx, params)
}

func toSectionDTOs(sections []events.MaturitySectionData) []readmodels.MaturitySectionDTO {
	result := make([]readmodels.MaturitySectionDTO, len(sections))
	for i, s := range sections {
		result[i] = readmodels.MaturitySectionDTO{
			Order:    s.Order,
			Name:     s.Name,
			MinValue: s.MinValue,
			MaxValue: s.MaxValue,
		}
	}
	return result
}

func toPillarDTOs(pillars []events.StrategyPillarData) []readmodels.StrategyPillarDTO {
	result := make([]readmodels.StrategyPillarDTO, len(pillars))
	for i, p := range pillars {
		result[i] = readmodels.StrategyPillarDTO{
			ID:                p.ID,
			Name:              p.Name,
			Description:       p.Description,
			Active:            p.Active,
			FitScoringEnabled: p.FitScoringEnabled,
			FitCriteria:       p.FitCriteria,
			FitType:           p.FitType,
		}
	}
	return result
}
