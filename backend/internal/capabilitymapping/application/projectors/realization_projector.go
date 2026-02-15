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
	Update(ctx context.Context, update readmodels.RealizationUpdate) error
	Delete(ctx context.Context, id string) error
	DeleteBySourceRealizationID(ctx context.Context, sourceRealizationID string) error
	DeleteInheritedBySourceRealizationIDAndCapabilities(ctx context.Context, deletion readmodels.InheritedRealizationDeletion) error
	DeleteByComponentID(ctx context.Context, componentID string) error
	UpdateSourceCapabilityName(ctx context.Context, update readmodels.NameUpdate) error
	UpdateComponentName(ctx context.Context, update readmodels.NameUpdate) error
}

type RealizationProjector struct {
	readModel        RealizationProjectorReadModel
	componentGateway architecturemodeling.ComponentGateway
	handlers         map[string]eventHandlerFunc
}

type eventHandlerFunc func(ctx context.Context, eventData []byte) error

func NewRealizationProjector(
	readModel RealizationProjectorReadModel,
	componentGateway architecturemodeling.ComponentGateway,
) *RealizationProjector {
	p := &RealizationProjector{
		readModel:        readModel,
		componentGateway: componentGateway,
	}
	p.handlers = map[string]eventHandlerFunc{
		"SystemLinkedToCapability":           p.handleSystemLinked,
		"SystemRealizationUpdated":           p.handleRealizationUpdated,
		"SystemRealizationDeleted":           p.handleRealizationDeleted,
		"CapabilityRealizationsInherited":    p.handleCapabilityRealizationsInherited,
		"CapabilityRealizationsUninherited":  p.handleCapabilityRealizationsUninherited,
		"CapabilityUpdated":                  p.handleCapabilityUpdated,
		archPL.ApplicationComponentUpdated:   p.handleApplicationComponentUpdated,
		archPL.ApplicationComponentDeleted:   p.handleApplicationComponentDeleted,
	}
	return p
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
	if handler, ok := p.handlers[eventType]; ok {
		return handler(ctx, eventData)
	}
	return nil
}

func (p *RealizationProjector) handleSystemLinked(ctx context.Context, eventData []byte) error {
	event, err := unmarshalEvent[events.SystemLinkedToCapability](eventData)
	if err != nil {
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
	event, err := unmarshalEvent[events.SystemRealizationUpdated](eventData)
	if err != nil {
		return err
	}
	return p.readModel.Update(ctx, readmodels.RealizationUpdate{
		ID:               event.ID,
		RealizationLevel: event.RealizationLevel,
		Notes:            event.Notes,
	})
}

func (p *RealizationProjector) handleRealizationDeleted(ctx context.Context, eventData []byte) error {
	event, err := unmarshalEvent[events.SystemRealizationDeleted](eventData)
	if err != nil {
		return err
	}
	if err := p.readModel.DeleteBySourceRealizationID(ctx, event.ID); err != nil {
		return err
	}
	return p.readModel.Delete(ctx, event.ID)
}

func (p *RealizationProjector) handleCapabilityRealizationsInherited(ctx context.Context, eventData []byte) error {
	event, err := unmarshalEvent[events.CapabilityRealizationsInherited](eventData)
	if err != nil {
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
	event, err := unmarshalEvent[events.CapabilityRealizationsUninherited](eventData)
	if err != nil {
		return err
	}

	for _, removal := range event.Removals {
		if err := p.readModel.DeleteInheritedBySourceRealizationIDAndCapabilities(ctx, readmodels.InheritedRealizationDeletion{
			SourceRealizationID: removal.SourceRealizationID,
			CapabilityIDs:       removal.CapabilityIDs,
		}); err != nil {
			return err
		}
	}

	return nil
}

func (p *RealizationProjector) handleCapabilityUpdated(ctx context.Context, eventData []byte) error {
	event, err := unmarshalEvent[events.CapabilityUpdated](eventData)
	if err != nil {
		return err
	}
	return p.readModel.UpdateSourceCapabilityName(ctx, readmodels.NameUpdate{ID: event.ID, Name: event.Name})
}

type applicationComponentUpdatedEvent struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type applicationComponentDeletedEvent struct {
	ID string `json:"id"`
}

func (p *RealizationProjector) handleApplicationComponentUpdated(ctx context.Context, eventData []byte) error {
	event, err := unmarshalEvent[applicationComponentUpdatedEvent](eventData)
	if err != nil {
		return err
	}
	return p.readModel.UpdateComponentName(ctx, readmodels.NameUpdate{ID: event.ID, Name: event.Name})
}

func (p *RealizationProjector) handleApplicationComponentDeleted(ctx context.Context, eventData []byte) error {
	event, err := unmarshalEvent[applicationComponentDeletedEvent](eventData)
	if err != nil {
		return err
	}
	return p.readModel.DeleteByComponentID(ctx, event.ID)
}
