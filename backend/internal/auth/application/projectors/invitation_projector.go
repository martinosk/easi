package projectors

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"easi/backend/internal/auth/application/readmodels"
	"easi/backend/internal/auth/domain/events"
	"easi/backend/internal/shared/eventsourcing"
)

type InvitationProjector struct {
	readModel *readmodels.InvitationReadModel
}

type statusTransitionEvent struct {
	ID         string     `json:"id"`
	AcceptedAt *time.Time `json:"acceptedAt,omitempty"`
	RevokedAt  *time.Time `json:"revokedAt,omitempty"`
}

func NewInvitationProjector(readModel *readmodels.InvitationReadModel) *InvitationProjector {
	return &InvitationProjector{
		readModel: readModel,
	}
}

func (p *InvitationProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		log.Printf("Failed to marshal event data: %v", err)
		return err
	}
	return p.ProjectEvent(ctx, event.EventType(), eventData)
}

func (p *InvitationProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	switch eventType {
	case "InvitationCreated":
		return p.handleInvitationCreated(ctx, eventData)
	case "InvitationAccepted":
		return p.handleInvitationAccepted(ctx, eventData)
	case "InvitationRevoked":
		return p.handleInvitationRevoked(ctx, eventData)
	case "InvitationExpired":
		return p.handleInvitationExpired(ctx, eventData)
	}
	return nil
}

func (p *InvitationProjector) handleInvitationCreated(ctx context.Context, eventData []byte) error {
	var event events.InvitationCreated
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal InvitationCreated event: %v", err)
		return err
	}

	var invitedBy *string
	if event.InviterID != "" {
		invitedBy = &event.InviterID
	}

	dto := readmodels.InvitationDTO{
		ID:        event.ID,
		Email:     event.Email,
		Role:      event.Role,
		Status:    "pending",
		InvitedBy: invitedBy,
		CreatedAt: event.CreatedAt,
		ExpiresAt: event.ExpiresAt,
	}

	return p.readModel.Insert(ctx, dto)
}

func (p *InvitationProjector) handleStatusTransition(ctx context.Context, eventData []byte, eventType, status string) error {
	var event statusTransitionEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal %s event: %v", eventType, err)
		return err
	}

	return p.readModel.UpdateStatus(ctx, readmodels.StatusUpdate{
		ID:         event.ID,
		Status:     status,
		AcceptedAt: event.AcceptedAt,
		RevokedAt:  event.RevokedAt,
	})
}

func (p *InvitationProjector) handleInvitationAccepted(ctx context.Context, eventData []byte) error {
	return p.handleStatusTransition(ctx, eventData, "InvitationAccepted", "accepted")
}

func (p *InvitationProjector) handleInvitationRevoked(ctx context.Context, eventData []byte) error {
	return p.handleStatusTransition(ctx, eventData, "InvitationRevoked", "revoked")
}

func (p *InvitationProjector) handleInvitationExpired(ctx context.Context, eventData []byte) error {
	return p.handleStatusTransition(ctx, eventData, "InvitationExpired", "expired")
}
