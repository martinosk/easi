package projectors

import (
	"context"
	"encoding/json"
	"fmt"
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
		wrappedErr := fmt.Errorf("marshal %s event for aggregate %s: %w", event.EventType(), event.AggregateID(), err)
		log.Printf("failed to marshal event data: %v", wrappedErr)
		return wrappedErr
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
	return handleProjection(ctx, eventData, func(ctx context.Context, event events.EnterpriseStrategicImportanceSet) error {
		return p.readModel.Insert(ctx, readmodels.EnterpriseStrategicImportanceDTO{
			ID:                     event.ID,
			EnterpriseCapabilityID: event.EnterpriseCapabilityID,
			PillarID:               event.PillarID,
			PillarName:             event.PillarName,
			Importance:             event.Importance,
			Rationale:              event.Rationale,
			SetAt:                  event.SetAt,
		})
	})
}

func (p *EnterpriseStrategicImportanceProjector) handleUpdated(ctx context.Context, eventData []byte) error {
	return handleProjection(ctx, eventData, func(ctx context.Context, event events.EnterpriseStrategicImportanceUpdated) error {
		return p.readModel.Update(ctx, event.ID, event.Importance, event.Rationale)
	})
}

func (p *EnterpriseStrategicImportanceProjector) handleRemoved(ctx context.Context, eventData []byte) error {
	return handleProjection(ctx, eventData, func(ctx context.Context, event events.EnterpriseStrategicImportanceRemoved) error {
		return p.readModel.Delete(ctx, event.ID)
	})
}
