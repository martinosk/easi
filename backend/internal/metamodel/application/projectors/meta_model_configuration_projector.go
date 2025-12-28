package projectors

import (
	"context"
	"encoding/json"
	"log"

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
		ID:         event.ID,
		TenantID:   event.TenantID,
		Sections:   toSectionDTOs(event.Sections),
		Version:    1,
		IsDefault:  true,
		CreatedAt:  event.CreatedAt,
		ModifiedAt: event.CreatedAt,
		ModifiedBy: event.CreatedBy,
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
