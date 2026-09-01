-- Migration: 158_add_am_component_hosting
-- Purpose: spec 215 — hosting classification on application components
-- (on-premises/cloud/saas/third-party-hosted/unknown). The NOT NULL DEFAULT
-- backfills every existing row to 'unknown'; the value itself is projected
-- from the ApplicationHostingClassified event.

ALTER TABLE architecturemodeling.application_components
    ADD COLUMN IF NOT EXISTS hosting VARCHAR(20) NOT NULL DEFAULT 'unknown';

CREATE INDEX IF NOT EXISTS idx_application_components_hosting
    ON architecturemodeling.application_components(tenant_id, hosting);
