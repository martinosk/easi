package readmodels

import (
	"context"

	"easi/backend/internal/stewardship/application/home"
)

func (c *Caches) SaveApplication(ctx context.Context, componentID, name string) error {
	return c.store.exec(ctx, "cache application "+componentID, `
		INSERT INTO stewardship.application_cache (tenant_id, component_id, name)
		VALUES ($1, $2, $3)
		ON CONFLICT (tenant_id, component_id) DO UPDATE SET name = EXCLUDED.name`,
		componentID, name)
}

func (c *Caches) SetOwnership(ctx context.Context, ownership home.CachedOwnership) error {
	return c.store.exec(ctx, "set cached ownership of application "+ownership.ComponentID, `
		UPDATE stewardship.application_cache
		SET ownership_state = $3, owner_kind = NULLIF($4, ''), owner_id = NULLIF($5, '')
		WHERE tenant_id = $1 AND component_id = $2`,
		ownership.ComponentID, ownership.State, ownership.OwnerKind, ownership.OwnerID)
}

func (c *Caches) DeleteApplication(ctx context.Context, componentID string) error {
	return c.store.exec(ctx, "forget cached application "+componentID, `
		DELETE FROM stewardship.application_cache WHERE tenant_id = $1 AND component_id = $2`,
		componentID)
}

func (c *Caches) SaveRealization(ctx context.Context, realization home.CachedRealization) error {
	return c.store.exec(ctx, "cache realization "+realization.ID, `
		INSERT INTO stewardship.realization_cache (tenant_id, realization_id, capability_id, component_id)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (tenant_id, realization_id) DO UPDATE SET
			capability_id = EXCLUDED.capability_id, component_id = EXCLUDED.component_id`,
		realization.ID, realization.CapabilityID, realization.ComponentID)
}

func (c *Caches) DeleteRealization(ctx context.Context, realizationID string) error {
	return c.store.exec(ctx, "forget cached realization "+realizationID, `
		DELETE FROM stewardship.realization_cache WHERE tenant_id = $1 AND realization_id = $2`,
		realizationID)
}

func (c *Caches) DeleteRealizationsOfCapability(ctx context.Context, capabilityID string) error {
	return c.store.exec(ctx, "forget cached realizations of capability "+capabilityID, `
		DELETE FROM stewardship.realization_cache WHERE tenant_id = $1 AND capability_id = $2`,
		capabilityID)
}

func (c *Caches) DeleteRealizationsOfComponent(ctx context.Context, componentID string) error {
	return c.store.exec(ctx, "forget cached realizations of application "+componentID, `
		DELETE FROM stewardship.realization_cache WHERE tenant_id = $1 AND component_id = $2`,
		componentID)
}

func (c *Caches) SaveTimeAssessment(ctx context.Context, assessment home.CachedTimeAssessment) error {
	return c.store.exec(ctx, "cache TIME assessment of "+assessment.CapabilityID+"/"+assessment.ComponentID, `
		INSERT INTO stewardship.time_assessment_cache (tenant_id, capability_id, component_id, grade, assessed_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (tenant_id, capability_id, component_id) DO UPDATE SET
			grade = EXCLUDED.grade, assessed_at = EXCLUDED.assessed_at`,
		assessment.CapabilityID, assessment.ComponentID, assessment.Grade, assessment.AssessedAt.UTC())
}

func (c *Caches) DeleteTimeAssessment(ctx context.Context, capabilityID, componentID string) error {
	return c.store.exec(ctx, "forget cached TIME assessment of "+capabilityID+"/"+componentID, `
		DELETE FROM stewardship.time_assessment_cache WHERE tenant_id = $1 AND capability_id = $2 AND component_id = $3`,
		capabilityID, componentID)
}

func (c *Caches) SaveEditGrant(ctx context.Context, grant home.CachedEditGrant) error {
	return c.store.exec(ctx, "cache edit grant "+grant.ID, `
		INSERT INTO stewardship.edit_grant_cache (tenant_id, grant_id, artifact_type, artifact_id, grantee_email, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (tenant_id, grant_id) DO UPDATE SET
			artifact_type = EXCLUDED.artifact_type, artifact_id = EXCLUDED.artifact_id,
			grantee_email = EXCLUDED.grantee_email, expires_at = EXCLUDED.expires_at`,
		grant.ID, grant.ArtifactType, grant.ArtifactID, grant.GranteeEmail, grant.ExpiresAt.UTC())
}

func (c *Caches) DeleteEditGrant(ctx context.Context, grantID string) error {
	return c.store.exec(ctx, "forget cached edit grant "+grantID, `
		DELETE FROM stewardship.edit_grant_cache WHERE tenant_id = $1 AND grant_id = $2`,
		grantID)
}
