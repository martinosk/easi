-- Migration: 157_backfill_am_user_names_cache
-- Purpose: spec 214 — one-time seed of architecturemodeling.user_names from
-- auth.users so the cache starts complete; the UserCreated projector only
-- covers post-deploy users. Cross-schema read is allowed here because the
-- filename marks this migration as a backfill (spec 209 rule).

INSERT INTO architecturemodeling.user_names (tenant_id, user_id, name, email)
SELECT tenant_id, id::text, COALESCE(name, ''), email
FROM auth.users
ON CONFLICT (tenant_id, user_id) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email;
