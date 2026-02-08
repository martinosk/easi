package projectors

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	authCommands "easi/backend/internal/auth/application/commands"
	"easi/backend/internal/shared/cqrs"
	domain "easi/backend/internal/shared/eventsourcing"
)

type InvitationAutoCreateProjector struct {
	commandBus cqrs.CommandBus
}

func NewInvitationAutoCreateProjector(commandBus cqrs.CommandBus) *InvitationAutoCreateProjector {
	return &InvitationAutoCreateProjector{commandBus: commandBus}
}

func (p *InvitationAutoCreateProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		return fmt.Errorf("marshal event data: %w", err)
	}

	var data struct {
		GranteeEmail string `json:"granteeEmail"`
		GrantorID    string `json:"grantorId"`
		GrantorEmail string `json:"grantorEmail"`
	}
	if err := json.Unmarshal(eventData, &data); err != nil {
		log.Printf("[WARN] invitation-auto-create: failed to unmarshal event: %v", err)
		return nil
	}

	cmd := &authCommands.CreateInvitation{
		Email:        data.GranteeEmail,
		Role:         "stakeholder",
		InviterID:    data.GrantorID,
		InviterEmail: data.GrantorEmail,
	}

	if _, err := p.commandBus.Dispatch(ctx, cmd); err != nil {
		log.Printf("[WARN] invitation-auto-create: failed to create invitation for %s: %v", data.GranteeEmail, err)
		return nil
	}

	log.Printf("[AUDIT] invitation-auto-created email=%s inviter=%s reason=edit-grant", data.GranteeEmail, data.GrantorEmail)
	return nil
}
