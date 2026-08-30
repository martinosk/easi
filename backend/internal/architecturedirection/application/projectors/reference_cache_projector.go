package projectors

import (
	"context"
	"encoding/json"
	"fmt"

	"easi/backend/internal/architecturedirection/application/readmodels"
	amPL "easi/backend/internal/architecturemodeling/publishedlanguage"
	cmPL "easi/backend/internal/capabilitymapping/publishedlanguage"
	domain "easi/backend/internal/shared/eventsourcing"
)

type ReferenceCacheStore interface {
	SaveReference(ctx context.Context, entity readmodels.ReferenceEntity, entityID, name string) error
	RemoveReference(ctx context.Context, entity readmodels.ReferenceEntity, entityID string) error
}

type referenceCacheTarget struct {
	entity  readmodels.ReferenceEntity
	removed bool
}

var referenceCacheTargets = map[string]referenceCacheTarget{
	amPL.ApplicationComponentCreated: {entity: readmodels.ReferenceEntityApplication},
	amPL.ApplicationComponentUpdated: {entity: readmodels.ReferenceEntityApplication},
	amPL.ApplicationComponentDeleted: {entity: readmodels.ReferenceEntityApplication, removed: true},
	cmPL.BusinessDomainCreated:       {entity: readmodels.ReferenceEntityBusinessDomain},
	cmPL.BusinessDomainUpdated:       {entity: readmodels.ReferenceEntityBusinessDomain},
	cmPL.BusinessDomainDeleted:       {entity: readmodels.ReferenceEntityBusinessDomain, removed: true},
}

type ReferenceCacheProjector struct {
	cache ReferenceCacheStore
}

func NewReferenceCacheProjector(cache ReferenceCacheStore) *ReferenceCacheProjector {
	return &ReferenceCacheProjector{cache: cache}
}

func (p *ReferenceCacheProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	return dispatchReferenceEvent(ctx, event, p.ProjectEvent)
}

func (p *ReferenceCacheProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	target, tracked := referenceCacheTargets[eventType]
	if !tracked {
		return nil
	}
	var payload struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal(eventData, &payload); err != nil {
		return fmt.Errorf("unmarshal %s payload: %w", eventType, err)
	}
	if payload.ID == "" {
		return nil
	}
	if err := p.apply(ctx, target, payload.ID, payload.Name); err != nil {
		return fmt.Errorf("project %s into reference cache: %w", eventType, err)
	}
	return nil
}

func (p *ReferenceCacheProjector) apply(ctx context.Context, target referenceCacheTarget, entityID, name string) error {
	if target.removed {
		return p.cache.RemoveReference(ctx, target.entity, entityID)
	}
	return p.cache.SaveReference(ctx, target.entity, entityID, name)
}
