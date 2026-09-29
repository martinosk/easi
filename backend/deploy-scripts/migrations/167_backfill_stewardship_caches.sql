-- Migration: 167_backfill_stewardship_caches
-- Purpose: spec 226 — one-time seed of the Stewardship domain and user caches
-- from capabilitymapping.business_domains and auth.users so steward names,
-- statuses and domain architects resolve for everything created before
-- deployment; the cache projectors keep them current from then on.
-- Cross-schema reads are allowed because this is a backfill (spec 209). Idempotent.

INSERT INTO stewardship.domain_cache (tenant_id, domain_id, name, domain_architect_id)
SELECT tenant_id, id, name, NULLIF(domain_architect_id, '')
FROM capabilitymapping.business_domains
ON CONFLICT (tenant_id, domain_id) DO UPDATE SET
    name = EXCLUDED.name,
    domain_architect_id = EXCLUDED.domain_architect_id;

INSERT INTO stewardship.user_cache (tenant_id, user_id, name, email, active)
SELECT tenant_id, id::text, COALESCE(name, ''), email, status = 'active'
FROM auth.users
ON CONFLICT (tenant_id, user_id) DO UPDATE SET
    name = EXCLUDED.name,
    email = EXCLUDED.email,
    active = EXCLUDED.active;
