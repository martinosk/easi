package home

import (
	"context"

	capPL "easi/backend/internal/capabilitymapping/publishedlanguage"
)

type CapabilityCacheWriter interface {
	SaveCapability(ctx context.Context, capability CachedCapability) error
	RenameCapability(ctx context.Context, capabilityID CapabilityID, name string) error
	SetCapabilityMetadata(ctx context.Context, metadata CachedCapabilityMetadata) error
	MoveCapability(ctx context.Context, placement CachedPlacement) error
	SetCapabilityLevel(ctx context.Context, capabilityID CapabilityID, level string) error
	DeleteCapability(ctx context.Context, capabilityID CapabilityID) error
}

type DomainAssignmentCacheWriter interface {
	AssignToDomain(ctx context.Context, assignment CachedAssignment) error
	UnassignFromDomain(ctx context.Context, assignment CachedAssignment) error
	DeleteAssignmentsOfCapability(ctx context.Context, capabilityID CapabilityID) error
	DeleteAssignmentsToDomain(ctx context.Context, domainID DomainID) error
}

type capabilityPayload struct {
	ID       CapabilityID `json:"id"`
	Name     string       `json:"name"`
	ParentID CapabilityID `json:"parentId"`
	Level    string       `json:"level"`
	Status   string       `json:"status"`
	EAOwner  string       `json:"eaOwner"`
}

type capabilityMovePayload struct {
	CapabilityID CapabilityID `json:"capabilityId"`
	NewParentID  CapabilityID `json:"newParentId"`
	NewLevel     string       `json:"newLevel"`
}

type domainAssignmentPayload struct {
	BusinessDomainID DomainID     `json:"businessDomainId"`
	CapabilityID     CapabilityID `json:"capabilityId"`
}

func (p domainAssignmentPayload) assignment() CachedAssignment {
	return CachedAssignment{CapabilityID: p.CapabilityID, DomainID: p.BusinessDomainID}
}

type idPayload[T ~string] struct {
	ID T `json:"id"`
}

type CapabilityCacheProjector struct{ projections }

func NewCapabilityCacheProjector(cache CapabilityCacheWriter) *CapabilityCacheProjector {
	return &CapabilityCacheProjector{projections{
		capPL.CapabilityCreated: on(func(ctx context.Context, p capabilityPayload) error {
			return cache.SaveCapability(ctx, CachedCapability{ID: p.ID, Name: p.Name, Level: p.Level, ParentID: p.ParentID})
		}),
		capPL.CapabilityUpdated: on(func(ctx context.Context, p capabilityPayload) error {
			return cache.RenameCapability(ctx, p.ID, p.Name)
		}),
		capPL.CapabilityMetadataUpdated: on(func(ctx context.Context, p capabilityPayload) error {
			return cache.SetCapabilityMetadata(ctx, CachedCapabilityMetadata{CapabilityID: p.ID, Status: p.Status, EAOwner: p.EAOwner})
		}),
		capPL.CapabilityParentChanged: on(func(ctx context.Context, p capabilityMovePayload) error {
			return cache.MoveCapability(ctx, CachedPlacement{CapabilityID: p.CapabilityID, ParentID: p.NewParentID, Level: p.NewLevel})
		}),
		capPL.CapabilityLevelChanged: on(func(ctx context.Context, p capabilityMovePayload) error {
			return cache.SetCapabilityLevel(ctx, p.CapabilityID, p.NewLevel)
		}),
		capPL.CapabilityDeleted: on(func(ctx context.Context, p idPayload[CapabilityID]) error {
			return cache.DeleteCapability(ctx, p.ID)
		}),
	}}
}

type DomainAssignmentCacheProjector struct{ projections }

func NewDomainAssignmentCacheProjector(cache DomainAssignmentCacheWriter) *DomainAssignmentCacheProjector {
	return &DomainAssignmentCacheProjector{projections{
		capPL.CapabilityAssignedToDomain: on(func(ctx context.Context, p domainAssignmentPayload) error {
			return cache.AssignToDomain(ctx, p.assignment())
		}),
		capPL.CapabilityUnassignedFromDomain: on(func(ctx context.Context, p domainAssignmentPayload) error {
			return cache.UnassignFromDomain(ctx, p.assignment())
		}),
		capPL.CapabilityDeleted: on(func(ctx context.Context, p idPayload[CapabilityID]) error {
			return cache.DeleteAssignmentsOfCapability(ctx, p.ID)
		}),
		capPL.BusinessDomainDeleted: on(func(ctx context.Context, p idPayload[DomainID]) error {
			return cache.DeleteAssignmentsToDomain(ctx, p.ID)
		}),
	}}
}
