package projectors

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"easi/backend/internal/accessdelegation/application/readmodels"
	"easi/backend/internal/accessdelegation/domain/events"
	domain "easi/backend/internal/shared/eventsourcing"
)

type EditGrantProjector struct {
	readModel *readmodels.EditGrantReadModel
}

func NewEditGrantProjector(readModel *readmodels.EditGrantReadModel) *EditGrantProjector {
	return &EditGrantProjector{readModel: readModel}
}

func (p *EditGrantProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		log.Printf("Failed to marshal event data: %v", err)
		return err
	}
	return p.ProjectEvent(ctx, event.EventType(), eventData)
}

func (p *EditGrantProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	switch eventType {
	case "EditGrantActivated":
		return p.handleEditGrantActivated(ctx, eventData)
	case "EditGrantRevoked":
		return p.handleEditGrantRevoked(ctx, eventData)
	case "EditGrantExpired":
		return p.handleEditGrantExpired(ctx, eventData)
	}
	return nil
}

func (p *EditGrantProjector) handleEditGrantActivated(ctx context.Context, eventData []byte) error {
	var event events.EditGrantActivated
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal EditGrantActivated event: %v", err)
		return err
	}

	var reason *string
	if event.Reason != "" {
		reason = &event.Reason
	}

	dto := readmodels.EditGrantDTO{
		ID:           event.ID,
		GrantorID:    event.GrantorID,
		GrantorEmail: event.GrantorEmail,
		GranteeEmail: event.GranteeEmail,
		ArtifactType: event.ArtifactType,
		ArtifactID:   event.ArtifactID,
		Scope:        event.Scope,
		Status:       "active",
		Reason:       reason,
		CreatedAt:    event.CreatedAt,
		ExpiresAt:    event.ExpiresAt,
	}

	return p.readModel.Insert(ctx, dto)
}

type statusTransitionEvent struct {
	ID        string     `json:"id"`
	RevokedAt *time.Time `json:"revokedAt,omitempty"`
}

func (p *EditGrantProjector) handleStatusTransition(ctx context.Context, eventData []byte, eventType, status string) error {
	var event statusTransitionEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal %s event: %v", eventType, err)
		return err
	}

	return p.readModel.UpdateStatus(ctx, readmodels.EditGrantStatusUpdate{
		ID:        event.ID,
		Status:    status,
		RevokedAt: event.RevokedAt,
	})
}

func (p *EditGrantProjector) handleEditGrantRevoked(ctx context.Context, eventData []byte) error {
	return p.handleStatusTransition(ctx, eventData, "EditGrantRevoked", "revoked")
}

func (p *EditGrantProjector) handleEditGrantExpired(ctx context.Context, eventData []byte) error {
	return p.handleStatusTransition(ctx, eventData, "EditGrantExpired", "expired")
}
