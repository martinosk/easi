package projectors

import (
	"context"
	"encoding/json"
	"log"

	archPL "easi/backend/internal/architecturemodeling/publishedlanguage"
	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/domain/events"
	"easi/backend/internal/capabilitymapping/infrastructure/architecturemodeling"
	domain "easi/backend/internal/shared/eventsourcing"
)

type RealizationProjectorReadModel interface {
	Insert(ctx context.Context, dto readmodels.RealizationDTO) error
	InsertInherited(ctx context.Context, dto readmodels.RealizationDTO) error
	Update(ctx context.Context, id, realizationLevel, notes string) error
	Delete(ctx context.Context, id string) error
	DeleteBySourceRealizationID(ctx context.Context, sourceRealizationID string) error
	DeleteInheritedBySourceRealizationIDAndCapabilities(ctx context.Context, sourceRealizationID string, capabilityIDs []string) error
	DeleteByComponentID(ctx context.Context, componentID string) error
	UpdateSourceCapabilityName(ctx context.Context, capabilityID, capabilityName string) error
	UpdateComponentName(ctx context.Context, componentID, componentName string) error
}

type RealizationProjector struct {
	readModel        RealizationProjectorReadModel
	componentGateway architecturemodeling.ComponentGateway
}

func NewRealizationProjector(
	readModel RealizationProjectorReadModel,
	componentGateway architecturemodeling.ComponentGateway,
) *RealizationProjector {
	return &RealizationProjector{
		readModel:        readModel,
		componentGateway: componentGateway,
	}
}

func (p *RealizationProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		log.Printf("Failed to marshal event data: %v", err)
		return err
	}
	return p.ProjectEvent(ctx, event.EventType(), eventData)
}

func (p *RealizationProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	switch eventType {
	case "SystemLinkedToCapability":
		return p.handleSystemLinked(ctx, eventData)
	case "SystemRealizationUpdated":
		return p.handleRealizationUpdated(ctx, eventData)
	case "SystemRealizationDeleted":
		return p.handleRealizationDeleted(ctx, eventData)
	case "CapabilityRealizationsInherited":
		return p.handleCapabilityRealizationsInherited(ctx, eventData)
	case "CapabilityRealizationsUninherited":
		return p.handleCapabilityRealizationsUninherited(ctx, eventData)
	case "CapabilityUpdated":
		return p.handleCapabilityUpdated(ctx, eventData)
	case archPL.ApplicationComponentUpdated:
		return p.handleApplicationComponentUpdated(ctx, eventData)
	case archPL.ApplicationComponentDeleted:
		return p.handleApplicationComponentDeleted(ctx, eventData)
	}
	return nil
}

func (p *RealizationProjector) handleSystemLinked(ctx context.Context, eventData []byte) error {
	var event events.SystemLinkedToCapability
	if err := unmarshalEvent(eventData, &event, "SystemLinkedToCapability"); err != nil {
		return err
	}

	componentName := event.ComponentName
	if componentName == "" {
		componentName = p.lookupComponentName(ctx, event.ComponentID)
	}

	dto := readmodels.RealizationDTO{
		ID:               event.ID,
		CapabilityID:     event.CapabilityID,
		ComponentID:      event.ComponentID,
		ComponentName:    componentName,
		RealizationLevel: event.RealizationLevel,
		Notes:            event.Notes,
		Origin:           "Direct",
		LinkedAt:         event.LinkedAt,
	}

	if err := p.readModel.Insert(ctx, dto); err != nil {
		return err
	}

	return nil
}

func (p *RealizationProjector) lookupComponentName(ctx context.Context, componentID string) string {
	component, err := p.componentGateway.GetByID(ctx, componentID)
	if err != nil || component == nil {
		return ""
	}
	return component.Name
}

func (p *RealizationProjector) handleRealizationUpdated(ctx context.Context, eventData []byte) error {
	var event events.SystemRealizationUpdated
	if err := unmarshalEvent(eventData, &event, "SystemRealizationUpdated"); err != nil {
		return err
	}
	return p.readModel.Update(ctx, event.ID, event.RealizationLevel, event.Notes)
}

func (p *RealizationProjector) handleRealizationDeleted(ctx context.Context, eventData []byte) error {
	var event events.SystemRealizationDeleted
	if err := unmarshalEvent(eventData, &event, "SystemRealizationDeleted"); err != nil {
		return err
	}
	if err := p.readModel.DeleteBySourceRealizationID(ctx, event.ID); err != nil {
		return err
	}
	return p.readModel.Delete(ctx, event.ID)
}

func unmarshalEvent(data []byte, v interface{}, eventType string) error {
	if err := json.Unmarshal(data, v); err != nil {
		log.Printf("Failed to unmarshal %s event: %v", eventType, err)
		return err
	}
	return nil
}

func (p *RealizationProjector) handleCapabilityRealizationsInherited(ctx context.Context, eventData []byte) error {
	var event events.CapabilityRealizationsInherited
	if err := unmarshalEvent(eventData, &event, "CapabilityRealizationsInherited"); err != nil {
		return err
	}

	for _, realization := range event.InheritedRealizations {
		dto := readmodels.RealizationDTO{
			CapabilityID:         realization.CapabilityID,
			ComponentID:          realization.ComponentID,
			ComponentName:        realization.ComponentName,
			RealizationLevel:     realization.RealizationLevel,
			Notes:                realization.Notes,
			Origin:               realization.Origin,
			SourceRealizationID:  realization.SourceRealizationID,
			SourceCapabilityID:   realization.SourceCapabilityID,
			SourceCapabilityName: realization.SourceCapabilityName,
			LinkedAt:             realization.LinkedAt,
		}
		if err := p.readModel.InsertInherited(ctx, dto); err != nil {
			return err
		}
	}

	return nil
}

func (p *RealizationProjector) handleCapabilityRealizationsUninherited(ctx context.Context, eventData []byte) error {
	var event events.CapabilityRealizationsUninherited
	if err := unmarshalEvent(eventData, &event, "CapabilityRealizationsUninherited"); err != nil {
		return err
	}

	for _, removal := range event.Removals {
		if err := p.readModel.DeleteInheritedBySourceRealizationIDAndCapabilities(ctx, removal.SourceRealizationID, removal.CapabilityIDs); err != nil {
			return err
		}
	}

	return nil
}

func (p *RealizationProjector) handleCapabilityUpdated(ctx context.Context, eventData []byte) error {
	var event events.CapabilityUpdated
	if err := unmarshalEvent(eventData, &event, "CapabilityUpdated"); err != nil {
		return err
	}
	return p.readModel.UpdateSourceCapabilityName(ctx, event.ID, event.Name)
}

type applicationComponentUpdatedEvent struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type applicationComponentDeletedEvent struct {
	ID string `json:"id"`
}

func (p *RealizationProjector) handleApplicationComponentUpdated(ctx context.Context, eventData []byte) error {
	var event applicationComponentUpdatedEvent
	if err := unmarshalEvent(eventData, &event, "ApplicationComponentUpdated"); err != nil {
		return err
	}
	return p.readModel.UpdateComponentName(ctx, event.ID, event.Name)
}

func (p *RealizationProjector) handleApplicationComponentDeleted(ctx context.Context, eventData []byte) error {
	var event applicationComponentDeletedEvent
	if err := unmarshalEvent(eventData, &event, "ApplicationComponentDeleted"); err != nil {
		return err
	}
	return p.readModel.DeleteByComponentID(ctx, event.ID)
}
