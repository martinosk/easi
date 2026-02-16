package projectors

import (
	"context"
	"encoding/json"
	"log"

	"easi/backend/internal/auth/application/readmodels"
	authPL "easi/backend/internal/auth/publishedlanguage"
	domain "easi/backend/internal/shared/eventsourcing"
)

type UserProjector struct {
	readModel *readmodels.UserReadModel
}

func NewUserProjector(readModel *readmodels.UserReadModel) *UserProjector {
	return &UserProjector{
		readModel: readModel,
	}
}

func (p *UserProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		log.Printf("Failed to marshal event data: %v", err)
		return err
	}
	return p.ProjectEvent(ctx, event.EventType(), eventData)
}

func (p *UserProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	switch eventType {
	case authPL.UserCreated:
		return p.handleUserCreated(ctx, eventData)
	case authPL.UserRoleChanged:
		return p.handleUserRoleChanged(ctx, eventData)
	case authPL.UserDisabled:
		return p.handleUserDisabled(ctx, eventData)
	case authPL.UserEnabled:
		return p.handleUserEnabled(ctx, eventData)
	}
	return nil
}

type userCreatedEvent struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	Name         string `json:"name"`
	Role         string `json:"role"`
	Status       string `json:"status"`
	ExternalID   string `json:"externalId"`
	InvitationID string `json:"invitationId"`
	CreatedAt    string `json:"createdAt"`
}

func (p *UserProjector) handleUserCreated(ctx context.Context, eventData []byte) error {
	var event userCreatedEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal UserCreated event: %v", err)
		return err
	}

	return p.readModel.InsertFromEvent(ctx, readmodels.UserEventData{
		ID:           event.ID,
		Email:        event.Email,
		Name:         event.Name,
		Role:         event.Role,
		Status:       event.Status,
		ExternalID:   event.ExternalID,
		InvitationID: event.InvitationID,
		CreatedAt:    event.CreatedAt,
	})
}

type userRoleChangedEvent struct {
	ID      string `json:"id"`
	NewRole string `json:"newRole"`
}

func (p *UserProjector) handleUserRoleChanged(ctx context.Context, eventData []byte) error {
	var event userRoleChangedEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal UserRoleChanged event: %v", err)
		return err
	}

	return p.readModel.UpdateRole(ctx, event.ID, event.NewRole)
}

type userStatusEvent struct {
	ID string `json:"id"`
}

func (p *UserProjector) handleUserDisabled(ctx context.Context, eventData []byte) error {
	var event userStatusEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal UserDisabled event: %v", err)
		return err
	}

	return p.readModel.UpdateStatus(ctx, event.ID, "disabled")
}

func (p *UserProjector) handleUserEnabled(ctx context.Context, eventData []byte) error {
	var event userStatusEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal UserEnabled event: %v", err)
		return err
	}

	return p.readModel.UpdateStatus(ctx, event.ID, "active")
}
