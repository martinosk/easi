package readmodels

import (
	"context"

	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/stewardship/application/home"
)

type Caches struct {
	store tenantStore
}

func NewCaches(db *database.TenantAwareDB) *Caches {
	return &Caches{store: tenantStore{db: db}}
}

func (c *Caches) SaveCapability(ctx context.Context, capability home.CachedCapability) error {
	return c.store.exec(ctx, "cache capability "+string(capability.ID), `
		INSERT INTO stewardship.capability_cache (tenant_id, capability_id, name, level, parent_id, status)
		VALUES ($1, $2, $3, $4, NULLIF($5, ''), 'Active')
		ON CONFLICT (tenant_id, capability_id) DO UPDATE SET
			name = EXCLUDED.name, level = EXCLUDED.level, parent_id = EXCLUDED.parent_id`,
		capability.ID, capability.Name, capability.Level, capability.ParentID)
}

func (c *Caches) RenameCapability(ctx context.Context, capabilityID home.CapabilityID, name string) error {
	return c.store.exec(ctx, "rename cached capability "+string(capabilityID), `
		UPDATE stewardship.capability_cache SET name = $3 WHERE tenant_id = $1 AND capability_id = $2`,
		capabilityID, name)
}

func (c *Caches) SetCapabilityMetadata(ctx context.Context, metadata home.CachedCapabilityMetadata) error {
	return c.store.exec(ctx, "set cached metadata of capability "+string(metadata.CapabilityID), `
		UPDATE stewardship.capability_cache SET status = COALESCE(NULLIF($3, ''), status), ea_owner = NULLIF($4, '')
		WHERE tenant_id = $1 AND capability_id = $2`,
		metadata.CapabilityID, metadata.Status, metadata.EAOwner)
}

func (c *Caches) MoveCapability(ctx context.Context, placement home.CachedPlacement) error {
	return c.store.exec(ctx, "move cached capability "+string(placement.CapabilityID), `
		UPDATE stewardship.capability_cache SET parent_id = NULLIF($3, ''), level = COALESCE(NULLIF($4, ''), level)
		WHERE tenant_id = $1 AND capability_id = $2`,
		placement.CapabilityID, placement.ParentID, placement.Level)
}

func (c *Caches) SetCapabilityLevel(ctx context.Context, capabilityID home.CapabilityID, level string) error {
	return c.store.exec(ctx, "set cached level of capability "+string(capabilityID), `
		UPDATE stewardship.capability_cache SET level = $3 WHERE tenant_id = $1 AND capability_id = $2`,
		capabilityID, level)
}

func (c *Caches) DeleteCapability(ctx context.Context, capabilityID home.CapabilityID) error {
	return c.store.exec(ctx, "forget cached capability "+string(capabilityID), `
		DELETE FROM stewardship.capability_cache WHERE tenant_id = $1 AND capability_id = $2`,
		capabilityID)
}

func (c *Caches) AssignToDomain(ctx context.Context, assignment home.CachedAssignment) error {
	return c.store.exec(ctx, "cache assignment of capability "+string(assignment.CapabilityID)+" to domain "+string(assignment.DomainID), `
		INSERT INTO stewardship.domain_assignment_cache (tenant_id, capability_id, domain_id)
		VALUES ($1, $2, $3) ON CONFLICT (tenant_id, capability_id, domain_id) DO NOTHING`,
		assignment.CapabilityID, assignment.DomainID)
}

func (c *Caches) UnassignFromDomain(ctx context.Context, assignment home.CachedAssignment) error {
	return c.store.exec(ctx, "forget assignment of capability "+string(assignment.CapabilityID)+" to domain "+string(assignment.DomainID), `
		DELETE FROM stewardship.domain_assignment_cache WHERE tenant_id = $1 AND capability_id = $2 AND domain_id = $3`,
		assignment.CapabilityID, assignment.DomainID)
}

func (c *Caches) DeleteAssignmentsOfCapability(ctx context.Context, capabilityID home.CapabilityID) error {
	return c.store.exec(ctx, "forget domain assignments of capability "+string(capabilityID), `
		DELETE FROM stewardship.domain_assignment_cache WHERE tenant_id = $1 AND capability_id = $2`,
		capabilityID)
}

func (c *Caches) DeleteAssignmentsToDomain(ctx context.Context, domainID home.DomainID) error {
	return c.store.exec(ctx, "forget capability assignments to domain "+string(domainID), `
		DELETE FROM stewardship.domain_assignment_cache WHERE tenant_id = $1 AND domain_id = $2`,
		domainID)
}
