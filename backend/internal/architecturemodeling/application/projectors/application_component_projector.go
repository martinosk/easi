package projectors

import (
	"context"
	"encoding/json"
	"fmt"
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

type expertAddedAdapter struct {
	e events.ApplicationComponentExpertAdded
}

func (a expertAddedAdapter) getComponentID() string { return a.e.ComponentID }
func (a expertAddedAdapter) getExpertName() string  { return a.e.ExpertName }
func (a expertAddedAdapter) getExpertRole() string  { return a.e.ExpertRole }
func (a expertAddedAdapter) getContactInfo() string { return a.e.ContactInfo }
func (a expertAddedAdapter) getAddedAt() time.Time  { return a.e.AddedAt }

type expertRemovedAdapter struct {
	e events.ApplicationComponentExpertRemoved
}

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
		wrappedErr := fmt.Errorf("unmarshal %s event data: %w", eventName, err)
		log.Printf("failed to unmarshal %s event: %v", eventName, wrappedErr)
		return nil, wrappedErr
	}
	return &event, nil
}

func projectEvent[T any](ctx context.Context, eventData []byte, eventName string, fn func(context.Context, *T) error) error {
	event, err := unmarshalEvent[T](eventData, eventName)
	if err != nil {
		return err
	}
	return fn(ctx, event)
}

func (p *ApplicationComponentProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		wrappedErr := fmt.Errorf("marshal %s event for aggregate %s: %w", event.EventType(), event.AggregateID(), err)
		log.Printf("failed to marshal event data: %v", wrappedErr)
		return wrappedErr
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
	return projectEvent(ctx, eventData, "ApplicationComponentCreated", func(ctx context.Context, event *events.ApplicationComponentCreated) error {
		return p.readModel.Insert(ctx, readmodels.ApplicationComponentDTO{
			ID: event.ID, Name: event.Name, Description: event.Description, CreatedAt: event.CreatedAt,
		})
	})
}

func (p *ApplicationComponentProjector) projectUpdated(ctx context.Context, eventData []byte) error {
	return projectEvent(ctx, eventData, "ApplicationComponentUpdated", func(ctx context.Context, event *events.ApplicationComponentUpdated) error {
		return p.readModel.Update(ctx, event.ID, event.Name, event.Description)
	})
}

func (p *ApplicationComponentProjector) projectDeleted(ctx context.Context, eventData []byte) error {
	return projectEvent(ctx, eventData, "ApplicationComponentDeleted", func(ctx context.Context, event *events.ApplicationComponentDeleted) error {
		return p.readModel.MarkAsDeleted(ctx, event.ID, event.DeletedAt)
	})
}

func (p *ApplicationComponentProjector) projectExpertAdded(ctx context.Context, eventData []byte) error {
	return projectEvent(ctx, eventData, "ApplicationComponentExpertAdded", func(ctx context.Context, event *events.ApplicationComponentExpertAdded) error {
		return p.readModel.AddExpert(ctx, toExpertInfo(expertAddedAdapter{*event}))
	})
}

func (p *ApplicationComponentProjector) projectExpertRemoved(ctx context.Context, eventData []byte) error {
	return projectEvent(ctx, eventData, "ApplicationComponentExpertRemoved", func(ctx context.Context, event *events.ApplicationComponentExpertRemoved) error {
		return p.readModel.RemoveExpert(ctx, toExpertInfo(expertRemovedAdapter{*event}))
	})
}
