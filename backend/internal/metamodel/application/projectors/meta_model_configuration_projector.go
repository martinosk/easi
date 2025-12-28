package projectors

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"easi/backend/internal/metamodel/application/readmodels"
	"easi/backend/internal/metamodel/domain/events"
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
	case "MetaModelConfigurationCreated":
		return p.handleCreated(ctx, eventData)
	case "MaturityScaleConfigUpdated":
		return p.handleUpdated(ctx, eventData)
	case "MaturityScaleConfigReset":
		return p.handleReset(ctx, eventData)
	case "StrategyPillarAdded":
		return p.handlePillarAdded(ctx, eventData)
	case "StrategyPillarUpdated":
		return p.handlePillarUpdated(ctx, eventData)
	case "StrategyPillarRemoved":
		return p.handlePillarRemoved(ctx, eventData)
	}
	return nil
}

func (p *MetaModelConfigurationProjector) handleCreated(ctx context.Context, eventData []byte) error {
	var event events.MetaModelConfigurationCreated
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal MetaModelConfigurationCreated event: %v", err)
		return err
	}

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
}

func (p *MetaModelConfigurationProjector) handleUpdated(ctx context.Context, eventData []byte) error {
	var event events.MaturityScaleConfigUpdated
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal MaturityScaleConfigUpdated event: %v", err)
		return err
	}
	params := readmodels.UpdateParams{
		ID:         event.ID,
		Sections:   toSectionDTOs(event.NewSections),
		Version:    event.Version,
		IsDefault:  false,
		ModifiedAt: event.ModifiedAt,
		ModifiedBy: event.ModifiedBy,
	}
	return p.readModel.Update(ctx, params)
}

func (p *MetaModelConfigurationProjector) handleReset(ctx context.Context, eventData []byte) error {
	var event events.MaturityScaleConfigReset
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal MaturityScaleConfigReset event: %v", err)
		return err
	}
	params := readmodels.UpdateParams{
		ID:         event.ID,
		Sections:   toSectionDTOs(event.Sections),
		Version:    event.Version,
		IsDefault:  true,
		ModifiedAt: event.ModifiedAt,
		ModifiedBy: event.ModifiedBy,
	}
	return p.readModel.Update(ctx, params)
}

type pillarEventData struct {
	ID         string
	Version    int
	ModifiedAt time.Time
	ModifiedBy string
}

func (p *MetaModelConfigurationProjector) handlePillarAdded(ctx context.Context, eventData []byte) error {
	var event events.StrategyPillarAdded
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal StrategyPillarAdded event: %v", err)
		return err
	}

	return p.updatePillarsWithModification(ctx, pillarEventData{
		ID:         event.ID,
		Version:    event.Version,
		ModifiedAt: event.ModifiedAt,
		ModifiedBy: event.ModifiedBy,
	}, func(pillars []readmodels.StrategyPillarDTO) []readmodels.StrategyPillarDTO {
		return append(pillars, readmodels.StrategyPillarDTO{
			ID:          event.PillarID,
			Name:        event.Name,
			Description: event.Description,
			Active:      true,
		})
	})
}

func (p *MetaModelConfigurationProjector) handlePillarUpdated(ctx context.Context, eventData []byte) error {
	var event events.StrategyPillarUpdated
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal StrategyPillarUpdated event: %v", err)
		return err
	}

	return p.updatePillarsWithModification(ctx, pillarEventData{
		ID:         event.ID,
		Version:    event.Version,
		ModifiedAt: event.ModifiedAt,
		ModifiedBy: event.ModifiedBy,
	}, func(pillars []readmodels.StrategyPillarDTO) []readmodels.StrategyPillarDTO {
		for i, pillar := range pillars {
			if pillar.ID == event.PillarID {
				pillars[i].Name = event.NewName
				pillars[i].Description = event.NewDescription
				break
			}
		}
		return pillars
	})
}

func (p *MetaModelConfigurationProjector) handlePillarRemoved(ctx context.Context, eventData []byte) error {
	var event events.StrategyPillarRemoved
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal StrategyPillarRemoved event: %v", err)
		return err
	}

	return p.updatePillarsWithModification(ctx, pillarEventData{
		ID:         event.ID,
		Version:    event.Version,
		ModifiedAt: event.ModifiedAt,
		ModifiedBy: event.ModifiedBy,
	}, func(pillars []readmodels.StrategyPillarDTO) []readmodels.StrategyPillarDTO {
		for i, pillar := range pillars {
			if pillar.ID == event.PillarID {
				pillars[i].Active = false
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
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Active:      p.Active,
		}
	}
	return result
}
