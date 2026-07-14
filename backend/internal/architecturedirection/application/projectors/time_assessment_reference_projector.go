package projectors

import (
	"context"

	amPL "easi/backend/internal/architecturemodeling/publishedlanguage"
	authPL "easi/backend/internal/auth/publishedlanguage"
	cmPL "easi/backend/internal/capabilitymapping/publishedlanguage"
	domain "easi/backend/internal/shared/eventsourcing"
)

type TimeAssessmentReferenceStore interface {
	DeleteByCapabilityID(ctx context.Context, capabilityID string) error
	DeleteByComponentID(ctx context.Context, componentID string) error
	CacheCapabilityName(ctx context.Context, capabilityID, name string) error
	UpdateCapabilityName(ctx context.Context, capabilityID, name string) error
	CacheComponentName(ctx context.Context, componentID, name string) error
	UpdateComponentName(ctx context.Context, componentID, name string) error
	CacheUserName(ctx context.Context, email, name string) error
}

type TimeAssessmentReferenceProjector struct {
	readModel TimeAssessmentReferenceStore
}

func NewTimeAssessmentReferenceProjector(readModel TimeAssessmentReferenceStore) *TimeAssessmentReferenceProjector {
	return &TimeAssessmentReferenceProjector{readModel: readModel}
}

func (p *TimeAssessmentReferenceProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	return dispatchReferenceEvent(ctx, event, p.ProjectEvent)
}

func (p *TimeAssessmentReferenceProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	switch eventType {
	case cmPL.CapabilityDeleted:
		return dispatchByReferenceID(ctx, eventData, p.readModel.DeleteByCapabilityID)
	case amPL.ApplicationComponentDeleted:
		return dispatchByReferenceID(ctx, eventData, p.readModel.DeleteByComponentID)
	case cmPL.CapabilityCreated, cmPL.CapabilityUpdated:
		return dispatchReferenceNameChange(ctx, eventData, p.readModel.CacheCapabilityName, p.readModel.UpdateCapabilityName)
	case amPL.ApplicationComponentCreated, amPL.ApplicationComponentUpdated:
		return dispatchReferenceNameChange(ctx, eventData, p.readModel.CacheComponentName, p.readModel.UpdateComponentName)
	case authPL.UserCreated:
		return dispatchReferenceUserCreated(ctx, eventData, p.readModel.CacheUserName)
	default:
		return nil
	}
}
