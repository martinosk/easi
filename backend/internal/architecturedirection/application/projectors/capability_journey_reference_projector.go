package projectors

import (
	"context"

	amPL "easi/backend/internal/architecturemodeling/publishedlanguage"
	authPL "easi/backend/internal/auth/publishedlanguage"
	cmPL "easi/backend/internal/capabilitymapping/publishedlanguage"
	domain "easi/backend/internal/shared/eventsourcing"
)

type CapabilityJourneyReferenceStore interface {
	CacheReferenceName(ctx context.Context, entityType, entityID, name string) error
	UpdateCapabilityName(ctx context.Context, capabilityID, name string) error
	MarkCapabilityStale(ctx context.Context, capabilityID string) error
	UpdateComponentName(ctx context.Context, componentID, name string) error
	MarkComponentStale(ctx context.Context, componentID string) error
	UpdateDomainName(ctx context.Context, domainID, name string) error
	MarkDomainStale(ctx context.Context, domainID string) error
	UpdatePlannedByName(ctx context.Context, email, name string) error
}

type CapabilityJourneyReferenceProjector struct {
	readModel CapabilityJourneyReferenceStore
}

func NewCapabilityJourneyReferenceProjector(readModel CapabilityJourneyReferenceStore) *CapabilityJourneyReferenceProjector {
	return &CapabilityJourneyReferenceProjector{readModel: readModel}
}

func (p *CapabilityJourneyReferenceProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	return dispatchReferenceEvent(ctx, event, p.ProjectEvent)
}

func (p *CapabilityJourneyReferenceProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	switch eventType {
	case cmPL.CapabilityCreated, cmPL.CapabilityUpdated:
		return dispatchReferenceNameChange(ctx, eventData, p.cacheCapabilityName, p.readModel.UpdateCapabilityName)
	case cmPL.CapabilityDeleted:
		return dispatchByReferenceID(ctx, eventData, p.readModel.MarkCapabilityStale)
	case amPL.ApplicationComponentCreated, amPL.ApplicationComponentUpdated:
		return dispatchReferenceNameChange(ctx, eventData, p.cacheComponentName, p.readModel.UpdateComponentName)
	case amPL.ApplicationComponentDeleted:
		return dispatchByReferenceID(ctx, eventData, p.readModel.MarkComponentStale)
	case cmPL.BusinessDomainCreated, cmPL.BusinessDomainUpdated:
		return dispatchReferenceNameChange(ctx, eventData, p.cacheDomainName, p.readModel.UpdateDomainName)
	case cmPL.BusinessDomainDeleted:
		return dispatchByReferenceID(ctx, eventData, p.readModel.MarkDomainStale)
	case authPL.UserCreated:
		return dispatchReferenceUserCreated(ctx, eventData, p.readModel.UpdatePlannedByName)
	default:
		return nil
	}
}

func (p *CapabilityJourneyReferenceProjector) cacheCapabilityName(ctx context.Context, id, name string) error {
	return p.readModel.CacheReferenceName(ctx, "capability", id, name)
}

func (p *CapabilityJourneyReferenceProjector) cacheComponentName(ctx context.Context, id, name string) error {
	return p.readModel.CacheReferenceName(ctx, "application", id, name)
}

func (p *CapabilityJourneyReferenceProjector) cacheDomainName(ctx context.Context, id, name string) error {
	return p.readModel.CacheReferenceName(ctx, "business_domain", id, name)
}
