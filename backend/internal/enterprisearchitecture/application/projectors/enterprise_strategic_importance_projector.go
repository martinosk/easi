package projectors

import (
	"context"
	"encoding/json"
	"log"

	"easi/backend/internal/enterprisearchitecture/application/readmodels"
	"easi/backend/internal/enterprisearchitecture/domain/events"
	domain "easi/backend/internal/shared/eventsourcing"
)

type EnterpriseStrategicImportanceProjector struct {
	readModel *readmodels.EnterpriseStrategicImportanceReadModel
}

func NewEnterpriseStrategicImportanceProjector(readModel *readmodels.EnterpriseStrategicImportanceReadModel) *EnterpriseStrategicImportanceProjector {
	return &EnterpriseStrategicImportanceProjector{
		readModel: readModel,
	}
}

func (p *EnterpriseStrategicImportanceProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		log.Printf("Failed to marshal event data: %v", err)
		return err
	}
	return p.ProjectEvent(ctx, event.EventType(), eventData)
}

func (p *EnterpriseStrategicImportanceProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	handlers := map[string]func(context.Context, []byte) error{
		"EnterpriseStrategicImportanceSet":     p.handleSet,
		"EnterpriseStrategicImportanceUpdated": p.handleUpdated,
		"EnterpriseStrategicImportanceRemoved": p.handleRemoved,
	}

	if handler, exists := handlers[eventType]; exists {
		return handler(ctx, eventData)
	}
	return nil
}

func (p *EnterpriseStrategicImportanceProjector) handleSet(ctx context.Context, eventData []byte) error {
	var event events.EnterpriseStrategicImportanceSet
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal EnterpriseStrategicImportanceSet event: %v", err)
		return err
	}

	dto := readmodels.EnterpriseStrategicImportanceDTO{
		ID:                     event.ID,
		EnterpriseCapabilityID: event.EnterpriseCapabilityID,
		PillarID:               event.PillarID,
		PillarName:             event.PillarName,
		Importance:             event.Importance,
		Rationale:              event.Rationale,
		SetAt:                  event.SetAt,
	}
	return p.readModel.Insert(ctx, dto)
}

func (p *EnterpriseStrategicImportanceProjector) handleUpdated(ctx context.Context, eventData []byte) error {
	var event events.EnterpriseStrategicImportanceUpdated
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal EnterpriseStrategicImportanceUpdated event: %v", err)
		return err
	}
	return p.readModel.Update(ctx, event.ID, event.Importance, event.Rationale)
}

func (p *EnterpriseStrategicImportanceProjector) handleRemoved(ctx context.Context, eventData []byte) error {
	var event events.EnterpriseStrategicImportanceRemoved
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal EnterpriseStrategicImportanceRemoved event: %v", err)
		return err
	}
	return p.readModel.Delete(ctx, event.ID)
}
