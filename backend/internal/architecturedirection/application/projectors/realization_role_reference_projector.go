package projectors

import (
	"context"

	amPL "easi/backend/internal/architecturemodeling/publishedlanguage"
	cmPL "easi/backend/internal/capabilitymapping/publishedlanguage"
	domain "easi/backend/internal/shared/eventsourcing"
)

type RealizationRoleReferenceStore interface {
	DeleteByCapabilityID(ctx context.Context, capabilityID string) error
	DeleteByComponentID(ctx context.Context, componentID string) error
	CacheCapabilityName(ctx context.Context, capabilityID, name string) error
	UpdateCapabilityName(ctx context.Context, capabilityID, name string) error
	CacheComponentName(ctx context.Context, componentID, name string) error
	UpdateComponentName(ctx context.Context, componentID, name string) error
}

type RealizationRoleReferenceProjector struct {
	readModel RealizationRoleReferenceStore
}

func NewRealizationRoleReferenceProjector(readModel RealizationRoleReferenceStore) *RealizationRoleReferenceProjector {
	return &RealizationRoleReferenceProjector{readModel: readModel}
}

func (p *RealizationRoleReferenceProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	return dispatchReferenceEvent(ctx, event, p.ProjectEvent)
}

func (p *RealizationRoleReferenceProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	switch eventType {
	case cmPL.CapabilityDeleted:
		return dispatchByReferenceID(ctx, eventData, p.readModel.DeleteByCapabilityID)
	case amPL.ApplicationComponentDeleted:
		return dispatchByReferenceID(ctx, eventData, p.readModel.DeleteByComponentID)
	case cmPL.CapabilityCreated, cmPL.CapabilityUpdated:
		return dispatchReferenceNameChange(ctx, eventData, p.readModel.CacheCapabilityName, p.readModel.UpdateCapabilityName)
	case amPL.ApplicationComponentCreated, amPL.ApplicationComponentUpdated:
		return dispatchReferenceNameChange(ctx, eventData, p.readModel.CacheComponentName, p.readModel.UpdateComponentName)
	default:
		return nil
	}
}
