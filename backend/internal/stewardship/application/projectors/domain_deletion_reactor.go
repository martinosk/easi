package projectors

import (
	"context"
	"fmt"

	"easi/backend/internal/shared/cqrs"
	domain "easi/backend/internal/shared/eventsourcing"
	"easi/backend/internal/stewardship/application/commands"
)

const DomainDeletedActor = "system:domain-deleted"

type LiveConcernsReader interface {
	LiveConcernsOfDomain(ctx context.Context, domainID string) ([]string, error)
}

type DomainCacheForgetter interface {
	Delete(ctx context.Context, domainID string) error
}

type DomainDeletionReactor struct {
	stewardships LiveConcernsReader
	commandBus   cqrs.CommandBus
	domains      DomainCacheForgetter
}

func NewDomainDeletionReactor(stewardships LiveConcernsReader, commandBus cqrs.CommandBus, domains DomainCacheForgetter) *DomainDeletionReactor {
	return &DomainDeletionReactor{stewardships: stewardships, commandBus: commandBus, domains: domains}
}

func (r *DomainDeletionReactor) Handle(ctx context.Context, event domain.DomainEvent) error {
	payload, err := decodePayload[idPayload](event)
	if err != nil {
		return err
	}
	domainID := subjectID(event, payload.ID)
	if err := r.releaseEveryStewardship(ctx, domainID); err != nil {
		return err
	}
	return r.domains.Delete(ctx, domainID)
}

func (r *DomainDeletionReactor) releaseEveryStewardship(ctx context.Context, domainID string) error {
	concerns, err := r.stewardships.LiveConcernsOfDomain(ctx, domainID)
	if err != nil {
		return fmt.Errorf("list stewardships of deleted domain %s: %w", domainID, err)
	}
	for _, concern := range concerns {
		cmd := &commands.ReleaseSteward{DomainID: domainID, Concern: concern, ReleasedBy: DomainDeletedActor}
		if _, err := r.commandBus.Dispatch(ctx, cmd); err != nil {
			return fmt.Errorf("release %s stewardship of deleted domain %s: %w", concern, domainID, err)
		}
	}
	return nil
}
