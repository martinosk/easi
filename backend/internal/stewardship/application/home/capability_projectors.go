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
	ID       string `json:"id"`
	Name     string `json:"name"`
	ParentID string `json:"parentId"`
	Level    string `json:"level"`
	Status   string `json:"status"`
	EAOwner  string `json:"eaOwner"`
}

type capabilityMovePayload struct {
	CapabilityID string `json:"capabilityId"`
	NewParentID  string `json:"newParentId"`
	NewLevel     string `json:"newLevel"`
}

type domainAssignmentPayload struct {
	BusinessDomainID string `json:"businessDomainId"`
	CapabilityID     string `json:"capabilityId"`
}

func (p domainAssignmentPayload) assignment() CachedAssignment {
	return CachedAssignment{CapabilityID: CapabilityID(p.CapabilityID), DomainID: DomainID(p.BusinessDomainID)}
}

type idPayload struct {
	ID string `json:"id"`
}

type CapabilityCacheProjector struct{ projections }

func NewCapabilityCacheProjector(cache CapabilityCacheWriter) *CapabilityCacheProjector {
	return &CapabilityCacheProjector{projections{
		capPL.CapabilityCreated: on(func(ctx context.Context, p capabilityPayload) error {
			return cache.SaveCapability(ctx, CachedCapability{ID: p.ID, Name: p.Name, Level: p.Level, ParentID: p.ParentID})
		}),
		capPL.CapabilityUpdated: on(func(ctx context.Context, p capabilityPayload) error {
			return cache.RenameCapability(ctx, CapabilityID(p.ID), p.Name)
		}),
		capPL.CapabilityMetadataUpdated: on(func(ctx context.Context, p capabilityPayload) error {
			return cache.SetCapabilityMetadata(ctx, CachedCapabilityMetadata{CapabilityID: CapabilityID(p.ID), Status: p.Status, EAOwner: p.EAOwner})
		}),
		capPL.CapabilityParentChanged: on(func(ctx context.Context, p capabilityMovePayload) error {
			return cache.MoveCapability(ctx, CachedPlacement{CapabilityID: CapabilityID(p.CapabilityID), ParentID: p.NewParentID, Level: p.NewLevel})
		}),
		capPL.CapabilityLevelChanged: on(func(ctx context.Context, p capabilityMovePayload) error {
			return cache.SetCapabilityLevel(ctx, CapabilityID(p.CapabilityID), p.NewLevel)
		}),
		capPL.CapabilityDeleted: on(func(ctx context.Context, p idPayload) error {
			return cache.DeleteCapability(ctx, CapabilityID(p.ID))
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
		capPL.CapabilityDeleted: on(func(ctx context.Context, p idPayload) error {
			return cache.DeleteAssignmentsOfCapability(ctx, CapabilityID(p.ID))
		}),
		capPL.BusinessDomainDeleted: on(func(ctx context.Context, p idPayload) error {
			return cache.DeleteAssignmentsToDomain(ctx, DomainID(p.ID))
		}),
	}}
}
