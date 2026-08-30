-- Migration: 149_merge_platform_schema_into_auth
-- Purpose: Merge the Platform bounded context into Auth (spec 209 amendment, user decision 2026-08-30).
-- The Auth tenant caches (migration 139) existed only to satisfy the events-only rule against a
-- context that published a single event and owned three near-static tables. Merging Platform's
-- tenant tables directly into the auth schema restores live tenant reads -- including immediate
-- effect of tenant suspension at login -- and removes the caches entirely.

ALTER TABLE platform.tenants SET SCHEMA auth;
ALTER TABLE platform.tenant_domains SET SCHEMA auth;
ALTER TABLE platform.tenant_oidc_configs SET SCHEMA auth;

DROP TABLE IF EXISTS auth.tenant_oidc_cache;
DROP TABLE IF EXISTS auth.tenant_domain_cache;
DROP TABLE IF EXISTS auth.tenant_cache;

DROP SCHEMA platform;
