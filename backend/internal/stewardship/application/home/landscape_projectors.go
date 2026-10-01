package home

import (
	"context"
	"time"

	adPL "easi/backend/internal/accessdelegation/publishedlanguage"
	directionPL "easi/backend/internal/architecturedirection/publishedlanguage"
	amPL "easi/backend/internal/architecturemodeling/publishedlanguage"
	capPL "easi/backend/internal/capabilitymapping/publishedlanguage"
)

type ApplicationCacheWriter interface {
	SaveApplication(ctx context.Context, componentID ComponentID, name string) error
	SetOwnership(ctx context.Context, ownership CachedOwnership) error
	DeleteApplication(ctx context.Context, componentID ComponentID) error
}

type RealizationCacheWriter interface {
	SaveRealization(ctx context.Context, realization CachedRealization) error
	DeleteRealization(ctx context.Context, realizationID RealizationID) error
	DeleteRealizationsOfCapability(ctx context.Context, capabilityID CapabilityID) error
	DeleteRealizationsOfComponent(ctx context.Context, componentID ComponentID) error
}

type TimeAssessmentCacheWriter interface {
	SaveTimeAssessment(ctx context.Context, assessment CachedTimeAssessment) error
	DeleteTimeAssessment(ctx context.Context, capabilityID CapabilityID, componentID ComponentID) error
}

type EditGrantCacheWriter interface {
	SaveEditGrant(ctx context.Context, grant CachedEditGrant) error
	DeleteEditGrant(ctx context.Context, grantID EditGrantID) error
}

type applicationPayload struct {
	ID   ComponentID `json:"id"`
	Name string      `json:"name"`
}

type ownershipPayload struct {
	ComponentID    ComponentID `json:"componentId"`
	OwnerKind      string      `json:"ownerKind"`
	OwnerID        string      `json:"ownerId"`
	OwnershipState string      `json:"ownershipState"`
}

type realizationPayload struct {
	ID           RealizationID `json:"id"`
	CapabilityID CapabilityID  `json:"capabilityId"`
	ComponentID  ComponentID   `json:"componentId"`
}

type timeAssessmentPayload struct {
	CapabilityID CapabilityID `json:"capabilityId"`
	ComponentID  ComponentID  `json:"componentId"`
	Grade        string       `json:"grade"`
	OccurredOn   time.Time    `json:"occurredOn"`
}

type editGrantPayload struct {
	ID           EditGrantID `json:"id"`
	ArtifactType string      `json:"artifactType"`
	ArtifactID   string      `json:"artifactId"`
	GranteeEmail string      `json:"granteeEmail"`
	ExpiresAt    time.Time   `json:"expiresAt"`
}

type ApplicationCacheProjector struct{ projections }

func NewApplicationCacheProjector(cache ApplicationCacheWriter) *ApplicationCacheProjector {
	setOwnership := func(state func(ownershipPayload) string) projection {
		return on(func(ctx context.Context, p ownershipPayload) error {
			return cache.SetOwnership(ctx, CachedOwnership{ComponentID: p.ComponentID, State: state(p), OwnerKind: p.OwnerKind, OwnerID: p.OwnerID})
		})
	}
	recordedState := func(p ownershipPayload) string { return p.OwnershipState }
	return &ApplicationCacheProjector{projections{
		amPL.ApplicationComponentCreated: on(func(ctx context.Context, p applicationPayload) error {
			return cache.SaveApplication(ctx, p.ID, p.Name)
		}),
		amPL.ApplicationComponentUpdated: on(func(ctx context.Context, p applicationPayload) error {
			return cache.SaveApplication(ctx, p.ID, p.Name)
		}),
		amPL.ApplicationComponentDeleted: on(func(ctx context.Context, p applicationPayload) error {
			return cache.DeleteApplication(ctx, p.ID)
		}),
		amPL.ApplicationOwnerNominated:     setOwnership(func(ownershipPayload) string { return ownershipNominated }),
		amPL.ApplicationOwnershipConfirmed: setOwnership(recordedState),
		amPL.ApplicationOwnerAssigned:      setOwnership(recordedState),
		amPL.ApplicationOwnershipCleared: on(func(ctx context.Context, p ownershipPayload) error {
			return cache.SetOwnership(ctx, CachedOwnership{ComponentID: p.ComponentID, State: ownershipUnknown})
		}),
	}}
}

type RealizationCacheProjector struct{ projections }

func NewRealizationCacheProjector(cache RealizationCacheWriter) *RealizationCacheProjector {
	return &RealizationCacheProjector{projections{
		capPL.SystemLinkedToCapability: on(func(ctx context.Context, p realizationPayload) error {
			return cache.SaveRealization(ctx, CachedRealization(p))
		}),
		capPL.SystemRealizationDeleted: on(func(ctx context.Context, p idPayload[RealizationID]) error {
			return cache.DeleteRealization(ctx, p.ID)
		}),
		capPL.CapabilityDeleted: on(func(ctx context.Context, p idPayload[CapabilityID]) error {
			return cache.DeleteRealizationsOfCapability(ctx, p.ID)
		}),
		amPL.ApplicationComponentDeleted: on(func(ctx context.Context, p idPayload[ComponentID]) error {
			return cache.DeleteRealizationsOfComponent(ctx, p.ID)
		}),
	}}
}

type TimeAssessmentCacheProjector struct{ projections }

func NewTimeAssessmentCacheProjector(cache TimeAssessmentCacheWriter) *TimeAssessmentCacheProjector {
	return &TimeAssessmentCacheProjector{projections{
		directionPL.TimeAssessmentRecorded: on(func(ctx context.Context, p timeAssessmentPayload) error {
			return cache.SaveTimeAssessment(ctx, CachedTimeAssessment{CapabilityID: p.CapabilityID, ComponentID: p.ComponentID, Grade: p.Grade, AssessedAt: p.OccurredOn})
		}),
		directionPL.TimeAssessmentRemoved: on(func(ctx context.Context, p timeAssessmentPayload) error {
			return cache.DeleteTimeAssessment(ctx, p.CapabilityID, p.ComponentID)
		}),
	}}
}

type EditGrantCacheProjector struct{ projections }

func NewEditGrantCacheProjector(cache EditGrantCacheWriter) *EditGrantCacheProjector {
	forget := on(func(ctx context.Context, p idPayload[EditGrantID]) error { return cache.DeleteEditGrant(ctx, p.ID) })
	return &EditGrantCacheProjector{projections{
		adPL.EditGrantActivated: on(func(ctx context.Context, p editGrantPayload) error {
			return cache.SaveEditGrant(ctx, CachedEditGrant(p))
		}),
		adPL.EditGrantRevoked: forget,
		adPL.EditGrantExpired: forget,
	}}
}
