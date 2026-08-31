-- Migration: 142_backfill_accessdelegation_artifact_name_cache
-- Purpose: Seeds accessdelegation.artifact_name_cache from the owning contexts' tables (spec 209)
-- so grant artifact names are complete immediately after deployment; the projector keeps the
-- cache current from that point on. Cross-schema reads are allowed in a backfill migration.
-- Soft-deleted artifacts are skipped: they resolve to "Deleted artifact" exactly as before.
-- Idempotent.

INSERT INTO accessdelegation.artifact_name_cache (tenant_id, artifact_type, artifact_id, name)
SELECT tenant_id, 'capability', id, name
FROM capabilitymapping.capabilities
UNION ALL
SELECT tenant_id, 'domain', id, name
FROM capabilitymapping.business_domains
UNION ALL
SELECT tenant_id, 'component', id, name
FROM architecturemodeling.application_components
WHERE is_deleted = FALSE
UNION ALL
SELECT tenant_id, 'vendor', id, name
FROM architecturemodeling.vendors
WHERE is_deleted = FALSE
UNION ALL
SELECT tenant_id, 'acquired_entity', id, name
FROM architecturemodeling.acquired_entities
WHERE is_deleted = FALSE
UNION ALL
SELECT tenant_id, 'internal_team', id, name
FROM architecturemodeling.internal_teams
WHERE is_deleted = FALSE
UNION ALL
SELECT tenant_id, 'view', id, name
FROM architectureviews.architecture_views
WHERE is_deleted = FALSE
ON CONFLICT (tenant_id, artifact_type, artifact_id) DO UPDATE SET name = EXCLUDED.name;
