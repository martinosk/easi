package projectors

import (
	"context"
	"encoding/json"
	"log"

	"easi/backend/internal/enterprisearchitecture/application/readmodels"
	"easi/backend/internal/enterprisearchitecture/domain/events"
	domain "easi/backend/internal/shared/eventsourcing"
)

type EnterpriseCapabilityProjector struct {
	readModel *readmodels.EnterpriseCapabilityReadModel
}

func NewEnterpriseCapabilityProjector(readModel *readmodels.EnterpriseCapabilityReadModel) *EnterpriseCapabilityProjector {
	return &EnterpriseCapabilityProjector{
		readModel: readModel,
	}
}

func (p *EnterpriseCapabilityProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		log.Printf("Failed to marshal event data: %v", err)
		return err
	}
	return p.ProjectEvent(ctx, event.EventType(), eventData)
}

func (p *EnterpriseCapabilityProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	handlers := map[string]func(context.Context, []byte) error{
		"EnterpriseCapabilityCreated":  p.handleCreated,
		"EnterpriseCapabilityUpdated":  p.handleUpdated,
		"EnterpriseCapabilityDeleted":  p.handleDeleted,
		"EnterpriseCapabilityLinked":   p.handleLinked,
		"EnterpriseCapabilityUnlinked": p.handleUnlinked,
	}

	if handler, exists := handlers[eventType]; exists {
		return handler(ctx, eventData)
	}
	return nil
}

func (p *EnterpriseCapabilityProjector) handleCreated(ctx context.Context, eventData []byte) error {
	var event events.EnterpriseCapabilityCreated
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal EnterpriseCapabilityCreated event: %v", err)
		return err
	}

	dto := readmodels.EnterpriseCapabilityDTO{
		ID:          event.ID,
		Name:        event.Name,
		Description: event.Description,
		Category:    event.Category,
		Active:      event.Active,
		CreatedAt:   event.CreatedAt,
	}
	return p.readModel.Insert(ctx, dto)
}

func (p *EnterpriseCapabilityProjector) handleUpdated(ctx context.Context, eventData []byte) error {
	var event events.EnterpriseCapabilityUpdated
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal EnterpriseCapabilityUpdated event: %v", err)
		return err
	}
	return p.readModel.Update(ctx, readmodels.UpdateCapabilityParams{
		ID:          event.ID,
		Name:        event.Name,
		Description: event.Description,
		Category:    event.Category,
	})
}

func (p *EnterpriseCapabilityProjector) handleDeleted(ctx context.Context, eventData []byte) error {
	var event events.EnterpriseCapabilityDeleted
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal EnterpriseCapabilityDeleted event: %v", err)
		return err
	}
	return p.readModel.Delete(ctx, event.ID)
}

func (p *EnterpriseCapabilityProjector) handleLinked(ctx context.Context, eventData []byte) error {
	var event events.EnterpriseCapabilityLinked
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal EnterpriseCapabilityLinked event: %v", err)
		return err
	}
	if err := p.readModel.IncrementLinkCount(ctx, event.EnterpriseCapabilityID); err != nil {
		return err
	}
	return p.readModel.RecalculateDomainCount(ctx, event.EnterpriseCapabilityID)
}

func (p *EnterpriseCapabilityProjector) handleUnlinked(ctx context.Context, eventData []byte) error {
	var event events.EnterpriseCapabilityUnlinked
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal EnterpriseCapabilityUnlinked event: %v", err)
		return err
	}
	if err := p.readModel.DecrementLinkCount(ctx, event.EnterpriseCapabilityID); err != nil {
		return err
	}
	return p.readModel.RecalculateDomainCount(ctx, event.EnterpriseCapabilityID)
}
