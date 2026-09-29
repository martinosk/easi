package projectors

import (
	"context"

	domain "easi/backend/internal/shared/eventsourcing"
	"easi/backend/internal/stewardship/application/readmodels"
	"easi/backend/internal/stewardship/domain/events"
	pl "easi/backend/internal/stewardship/publishedlanguage"
)

type StewardshipStore interface {
	Upsert(ctx context.Context, row readmodels.StewardshipRow) error
	Delete(ctx context.Context, id string) error
}

type StewardshipProjector struct {
	store StewardshipStore
}

func NewStewardshipProjector(store StewardshipStore) *StewardshipProjector {
	return &StewardshipProjector{store: store}
}

func (p *StewardshipProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	switch event.EventType() {
	case pl.StewardAssigned:
		assigned, err := decodePayload[events.StewardAssigned](event)
		if err != nil {
			return err
		}
		return p.store.Upsert(ctx, readmodels.StewardshipRow{
			ID:         subjectID(event, assigned.ID),
			DomainID:   assigned.DomainID,
			Concern:    assigned.Concern,
			StewardID:  assigned.StewardID,
			AssignedBy: assigned.AssignedBy,
			AssignedAt: assigned.AssignedAt,
		})
	case pl.StewardReleased:
		released, err := decodePayload[idPayload](event)
		if err != nil {
			return err
		}
		return p.store.Delete(ctx, subjectID(event, released.ID))
	}
	return nil
}
