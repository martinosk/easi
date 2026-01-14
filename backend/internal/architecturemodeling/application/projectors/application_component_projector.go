package projectors

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"easi/backend/internal/architecturemodeling/application/readmodels"
	"easi/backend/internal/architecturemodeling/domain/events"
	domain "easi/backend/internal/shared/eventsourcing"
)

type expertEvent interface {
	getComponentID() string
	getExpertName() string
	getExpertRole() string
	getContactInfo() string
	getAddedAt() time.Time
}

type expertAddedAdapter struct{ e events.ApplicationComponentExpertAdded }

func (a expertAddedAdapter) getComponentID() string { return a.e.ComponentID }
func (a expertAddedAdapter) getExpertName() string  { return a.e.ExpertName }
func (a expertAddedAdapter) getExpertRole() string  { return a.e.ExpertRole }
func (a expertAddedAdapter) getContactInfo() string { return a.e.ContactInfo }
func (a expertAddedAdapter) getAddedAt() time.Time  { return a.e.AddedAt }

type expertRemovedAdapter struct{ e events.ApplicationComponentExpertRemoved }

func (a expertRemovedAdapter) getComponentID() string { return a.e.ComponentID }
func (a expertRemovedAdapter) getExpertName() string  { return a.e.ExpertName }
func (a expertRemovedAdapter) getExpertRole() string  { return a.e.ExpertRole }
func (a expertRemovedAdapter) getContactInfo() string { return a.e.ContactInfo }
func (a expertRemovedAdapter) getAddedAt() time.Time  { return time.Time{} }

func toExpertInfo(e expertEvent) readmodels.ExpertInfo {
	return readmodels.ExpertInfo{
		ComponentID: e.getComponentID(),
		Name:        e.getExpertName(),
		Role:        e.getExpertRole(),
		Contact:     e.getContactInfo(),
		AddedAt:     e.getAddedAt(),
	}
}

type ApplicationComponentProjector struct {
	readModel *readmodels.ApplicationComponentReadModel
}

func NewApplicationComponentProjector(readModel *readmodels.ApplicationComponentReadModel) *ApplicationComponentProjector {
	return &ApplicationComponentProjector{
		readModel: readModel,
	}
}

func unmarshalEvent[T any](eventData []byte, eventName string) (*T, error) {
	var event T
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal %s event: %v", eventName, err)
		return nil, err
	}
	return &event, nil
}

func (p *ApplicationComponentProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		log.Printf("Failed to marshal event data: %v", err)
		return err
	}
	return p.ProjectEvent(ctx, event.EventType(), eventData)
}

func (p *ApplicationComponentProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	switch eventType {
	case "ApplicationComponentCreated":
		return p.projectCreated(ctx, eventData)
	case "ApplicationComponentUpdated":
		return p.projectUpdated(ctx, eventData)
	case "ApplicationComponentDeleted":
		return p.projectDeleted(ctx, eventData)
	case "ApplicationComponentExpertAdded":
		return p.projectExpertAdded(ctx, eventData)
	case "ApplicationComponentExpertRemoved":
		return p.projectExpertRemoved(ctx, eventData)
	}
	return nil
}

func (p *ApplicationComponentProjector) projectCreated(ctx context.Context, eventData []byte) error {
	event, err := unmarshalEvent[events.ApplicationComponentCreated](eventData, "ApplicationComponentCreated")
	if err != nil {
		return err
	}

	return p.readModel.Insert(ctx, readmodels.ApplicationComponentDTO{
		ID:          event.ID,
		Name:        event.Name,
		Description: event.Description,
		CreatedAt:   event.CreatedAt,
	})
}

func (p *ApplicationComponentProjector) projectUpdated(ctx context.Context, eventData []byte) error {
	event, err := unmarshalEvent[events.ApplicationComponentUpdated](eventData, "ApplicationComponentUpdated")
	if err != nil {
		return err
	}
	return p.readModel.Update(ctx, event.ID, event.Name, event.Description)
}

func (p *ApplicationComponentProjector) projectDeleted(ctx context.Context, eventData []byte) error {
	event, err := unmarshalEvent[events.ApplicationComponentDeleted](eventData, "ApplicationComponentDeleted")
	if err != nil {
		return err
	}
	return p.readModel.MarkAsDeleted(ctx, event.ID, event.DeletedAt)
}

func (p *ApplicationComponentProjector) projectExpertAdded(ctx context.Context, eventData []byte) error {
	event, err := unmarshalEvent[events.ApplicationComponentExpertAdded](eventData, "ApplicationComponentExpertAdded")
	if err != nil {
		return err
	}
	return p.readModel.AddExpert(ctx, toExpertInfo(expertAddedAdapter{*event}))
}

func (p *ApplicationComponentProjector) projectExpertRemoved(ctx context.Context, eventData []byte) error {
	event, err := unmarshalEvent[events.ApplicationComponentExpertRemoved](eventData, "ApplicationComponentExpertRemoved")
	if err != nil {
		return err
	}
	return p.readModel.RemoveExpert(ctx, toExpertInfo(expertRemovedAdapter{*event}))
}
