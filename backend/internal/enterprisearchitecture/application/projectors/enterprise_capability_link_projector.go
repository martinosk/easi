package projectors

import (
	"context"
	"encoding/json"
	"log"

	"easi/backend/internal/enterprisearchitecture/application/readmodels"
	"easi/backend/internal/enterprisearchitecture/domain/events"
	domain "easi/backend/internal/shared/eventsourcing"
)

type EnterpriseCapabilityLinkProjector struct {
	readModel *readmodels.EnterpriseCapabilityLinkReadModel
}

func NewEnterpriseCapabilityLinkProjector(readModel *readmodels.EnterpriseCapabilityLinkReadModel) *EnterpriseCapabilityLinkProjector {
	return &EnterpriseCapabilityLinkProjector{
		readModel: readModel,
	}
}

func (p *EnterpriseCapabilityLinkProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		log.Printf("Failed to marshal event data: %v", err)
		return err
	}
	return p.ProjectEvent(ctx, event.EventType(), eventData)
}

func (p *EnterpriseCapabilityLinkProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	handlers := map[string]func(context.Context, []byte) error{
		"EnterpriseCapabilityLinked":   p.handleLinked,
		"EnterpriseCapabilityUnlinked": p.handleUnlinked,
	}

	if handler, exists := handlers[eventType]; exists {
		return handler(ctx, eventData)
	}
	return nil
}

func (p *EnterpriseCapabilityLinkProjector) handleLinked(ctx context.Context, eventData []byte) error {
	var event events.EnterpriseCapabilityLinked
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal EnterpriseCapabilityLinked event: %v", err)
		return err
	}

	dto := readmodels.EnterpriseCapabilityLinkDTO{
		ID:                     event.ID,
		EnterpriseCapabilityID: event.EnterpriseCapabilityID,
		DomainCapabilityID:     event.DomainCapabilityID,
		LinkedBy:               event.LinkedBy,
		LinkedAt:               event.LinkedAt,
	}
	return p.readModel.Insert(ctx, dto)
}

func (p *EnterpriseCapabilityLinkProjector) handleUnlinked(ctx context.Context, eventData []byte) error {
	var event events.EnterpriseCapabilityUnlinked
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal EnterpriseCapabilityUnlinked event: %v", err)
		return err
	}
	return p.readModel.Delete(ctx, event.ID)
}
