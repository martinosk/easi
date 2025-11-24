package projectors

import (
	"context"
	"encoding/json"
	"log"

	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/domain/events"
	"easi/backend/internal/shared/domain"
)

type RealizationProjector struct {
	readModel           *readmodels.RealizationReadModel
	capabilityReadModel *readmodels.CapabilityReadModel
}

func NewRealizationProjector(
	readModel *readmodels.RealizationReadModel,
	capabilityReadModel *readmodels.CapabilityReadModel,
) *RealizationProjector {
	return &RealizationProjector{
		readModel:           readModel,
		capabilityReadModel: capabilityReadModel,
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
	}
	return nil
}

func (p *RealizationProjector) handleSystemLinked(ctx context.Context, eventData []byte) error {
	var event events.SystemLinkedToCapability
	if err := unmarshalEvent(eventData, &event, "SystemLinkedToCapability"); err != nil {
		return err
	}

	dto := readmodels.RealizationDTO{
		ID:               event.ID,
		CapabilityID:     event.CapabilityID,
		ComponentID:      event.ComponentID,
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

	for _, realization := range realizations {
		sourceID := realization.ID
		if realization.Origin == "Inherited" && realization.SourceRealizationID != "" {
			sourceID = realization.SourceRealizationID
		}

		source := readmodels.RealizationDTO{
			ID:           sourceID,
			CapabilityID: event.NewParentID,
			ComponentID:  realization.ComponentID,
			LinkedAt:     realization.LinkedAt,
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
		CapabilityID:        source.CapabilityID,
		ComponentID:         source.ComponentID,
		RealizationLevel:    "Full",
		Origin:              "Inherited",
		SourceRealizationID: source.ID,
		LinkedAt:            source.LinkedAt,
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
