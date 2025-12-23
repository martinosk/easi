package projectors

import (
	"context"
	"encoding/json"
	"log"

	archReadmodels "easi/backend/internal/architecturemodeling/application/readmodels"
	archEvents "easi/backend/internal/architecturemodeling/domain/events"
	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/domain/events"
	"easi/backend/internal/shared/eventsourcing"
)

type RealizationProjector struct {
	readModel           *readmodels.RealizationReadModel
	capabilityReadModel *readmodels.CapabilityReadModel
	componentReadModel  *archReadmodels.ApplicationComponentReadModel
}

func NewRealizationProjector(
	readModel *readmodels.RealizationReadModel,
	capabilityReadModel *readmodels.CapabilityReadModel,
	componentReadModel *archReadmodels.ApplicationComponentReadModel,
) *RealizationProjector {
	return &RealizationProjector{
		readModel:           readModel,
		capabilityReadModel: capabilityReadModel,
		componentReadModel:  componentReadModel,
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
	case "CapabilityParentChanged":
		return p.handleCapabilityParentChanged(ctx, eventData)
	case "CapabilityUpdated":
		return p.handleCapabilityUpdated(ctx, eventData)
	case "ApplicationComponentUpdated":
		return p.handleApplicationComponentUpdated(ctx, eventData)
	}
	return nil
}

func (p *RealizationProjector) handleSystemLinked(ctx context.Context, eventData []byte) error {
	var event events.SystemLinkedToCapability
	if err := unmarshalEvent(eventData, &event, "SystemLinkedToCapability"); err != nil {
		return err
	}

	componentName := p.lookupComponentName(ctx, event.ComponentID)

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

	return p.createInheritedRealizationsForAncestors(ctx, dto)
}

func (p *RealizationProjector) lookupComponentName(ctx context.Context, componentID string) string {
	if p.componentReadModel == nil {
		return ""
	}
	component, err := p.componentReadModel.GetByID(ctx, componentID)
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

func (p *RealizationProjector) createInheritedRealizationsForAncestors(ctx context.Context, source readmodels.RealizationDTO) error {
	capability, err := p.capabilityReadModel.GetByID(ctx, source.CapabilityID)
	if err != nil {
		return err
	}
	if capability == nil || capability.ParentID == "" {
		return nil
	}

	source.SourceCapabilityID = source.CapabilityID
	source.SourceCapabilityName = capability.Name
	nextSource := source
	nextSource.CapabilityID = capability.ParentID
	return p.propagateInheritedRealizations(ctx, nextSource)
}

func (p *RealizationProjector) handleCapabilityParentChanged(ctx context.Context, eventData []byte) error {
	var event events.CapabilityParentChanged
	if err := unmarshalEvent(eventData, &event, "CapabilityParentChanged"); err != nil {
		return err
	}

	if event.NewParentID == "" {
		return nil
	}

	realizations, err := p.readModel.GetByCapabilityID(ctx, event.CapabilityID)
	if err != nil {
		return err
	}

	capability, err := p.capabilityReadModel.GetByID(ctx, event.CapabilityID)
	if err != nil {
		return err
	}

	for _, realization := range realizations {
		sourceID := realization.ID
		sourceCapabilityID := event.CapabilityID
		sourceCapabilityName := ""
		if capability != nil {
			sourceCapabilityName = capability.Name
		}

		if realization.Origin == "Inherited" && realization.SourceRealizationID != "" {
			sourceID = realization.SourceRealizationID
			sourceCapabilityID = realization.SourceCapabilityID
			sourceCapabilityName = realization.SourceCapabilityName
		}

		source := readmodels.RealizationDTO{
			ID:                   sourceID,
			CapabilityID:         event.NewParentID,
			ComponentID:          realization.ComponentID,
			ComponentName:        realization.ComponentName,
			SourceCapabilityID:   sourceCapabilityID,
			SourceCapabilityName: sourceCapabilityName,
			LinkedAt:             realization.LinkedAt,
		}

		if err := p.propagateInheritedRealizations(ctx, source); err != nil {
			return err
		}
	}

	return nil
}

func (p *RealizationProjector) propagateInheritedRealizations(ctx context.Context, source readmodels.RealizationDTO) error {
	capability, err := p.capabilityReadModel.GetByID(ctx, source.CapabilityID)
	if err != nil {
		return err
	}
	if capability == nil {
		return nil
	}

	inheritedDTO := readmodels.RealizationDTO{
		CapabilityID:         source.CapabilityID,
		ComponentID:          source.ComponentID,
		ComponentName:        source.ComponentName,
		RealizationLevel:     "Full",
		Origin:               "Inherited",
		SourceRealizationID:  source.ID,
		SourceCapabilityID:   source.SourceCapabilityID,
		SourceCapabilityName: source.SourceCapabilityName,
		LinkedAt:             source.LinkedAt,
	}

	if err := p.readModel.InsertInherited(ctx, inheritedDTO); err != nil {
		return err
	}

	if capability.ParentID == "" {
		return nil
	}

	nextSource := source
	nextSource.CapabilityID = capability.ParentID
	return p.propagateInheritedRealizations(ctx, nextSource)
}

func (p *RealizationProjector) handleCapabilityUpdated(ctx context.Context, eventData []byte) error {
	var event events.CapabilityUpdated
	if err := unmarshalEvent(eventData, &event, "CapabilityUpdated"); err != nil {
		return err
	}
	return p.readModel.UpdateSourceCapabilityName(ctx, event.ID, event.Name)
}

func (p *RealizationProjector) handleApplicationComponentUpdated(ctx context.Context, eventData []byte) error {
	var event archEvents.ApplicationComponentUpdated
	if err := unmarshalEvent(eventData, &event, "ApplicationComponentUpdated"); err != nil {
		return err
	}
	return p.readModel.UpdateComponentName(ctx, event.ID, event.Name)
}
