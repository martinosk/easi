package projectors

import (
	"context"

	capPL "easi/backend/internal/capabilitymapping/publishedlanguage"
	domain "easi/backend/internal/shared/eventsourcing"
	"easi/backend/internal/stewardship/application/readmodels"
)

type DomainCacheWriter interface {
	Upsert(ctx context.Context, domain readmodels.CachedDomain) error
}

type DomainCacheProjector struct {
	cache DomainCacheWriter
}

func NewDomainCacheProjector(cache DomainCacheWriter) *DomainCacheProjector {
	return &DomainCacheProjector{cache: cache}
}

type businessDomainPayload struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	DomainArchitectID string `json:"domainArchitectId"`
}

func (p *DomainCacheProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	if event.EventType() != capPL.BusinessDomainCreated && event.EventType() != capPL.BusinessDomainUpdated {
		return nil
	}
	payload, err := decodePayload[businessDomainPayload](event)
	if err != nil {
		return err
	}
	return p.cache.Upsert(ctx, readmodels.CachedDomain{
		ID:                subjectID(event, payload.ID),
		Name:              payload.Name,
		DomainArchitectID: payload.DomainArchitectID,
	})
}
