-- Migration: Add One-Pager Subject Index
-- Spec: 189_OnePagerQualityList
-- Description: Denormalized read model powering the global One-Pager Quality master list.
--   One row per subject across the six subject types, carrying the stored name, creator,
--   created/last-updated timestamps, and the completeness inputs (required_count / filled_count)
--   spanning required custom fields and required included built-in fields. Reverses D6:
--   completeness is materialized so the list can be sorted/keyset-paginated by completeness.
--
-- Backfill (mandatory for a new read model):
--   * name            -- from the six supplier read-model tables
--   * creator / dates -- from infrastructure.events (version = 1 for the creation event's
--                        actor + occurred_at; MAX(occurred_at) over the subject's own aggregate
--                        for last_updated_at), falling back to the subject row's created_at
--   * required_count / filled_count for CUSTOM fields -- from one_pager_configurations (JSONB)
--                        and one_pager_facts
--   * required_count / filled_count for BUILT-IN ATTRIBUTE fields -- evaluated per subject type
--                        against the subject's own read-model columns using 186's fill predicate
--                        (text/date/maturity present; experts list non-empty; maturity and name
--                        are always filled).
--
-- Documented simplification: required RELATION built-in fields (spec 188 — e.g.
--   realizing-applications, included-capabilities) are NOT counted here. Their fill depends on
--   cross-context relation tables and the EA composition service, which are impractical to
--   replicate correctly in pure SQL. They are non-default (an admin must both include AND require
--   a relation built-in, which emits BuiltInFieldRequirementChanged). The runtime projector
--   recomputes each subject's full completeness through the BuiltInFieldSource ports (including
--   relations) on the next relevant event, and recomputes every subject of a type the moment a
--   built-in's requirement changes — so post-deploy the list converges to the correct value.

CREATE TABLE IF NOT EXISTS onepagers.one_pager_subject_index (
    tenant_id VARCHAR(50) NOT NULL,
    subject_type VARCHAR(50) NOT NULL,
    subject_id VARCHAR(255) NOT NULL,
    name TEXT NOT NULL,
    creator_actor_id VARCHAR(255) NOT NULL,
    creator_email VARCHAR(500) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    last_updated_at TIMESTAMP NOT NULL,
    required_count INTEGER NOT NULL DEFAULT 0,
    filled_count INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (tenant_id, subject_type, subject_id)
);

CREATE INDEX IF NOT EXISTS idx_one_pager_subject_index_name
    ON onepagers.one_pager_subject_index (tenant_id, LOWER(name), subject_type, subject_id);

CREATE INDEX IF NOT EXISTS idx_one_pager_subject_index_creator
    ON onepagers.one_pager_subject_index (tenant_id, LOWER(creator_email), subject_type, subject_id);

CREATE INDEX IF NOT EXISTS idx_one_pager_subject_index_created
    ON onepagers.one_pager_subject_index (tenant_id, created_at, subject_type, subject_id);

CREATE INDEX IF NOT EXISTS idx_one_pager_subject_index_updated
    ON onepagers.one_pager_subject_index (tenant_id, last_updated_at, subject_type, subject_id);

CREATE INDEX IF NOT EXISTS idx_one_pager_subject_index_completeness
    ON onepagers.one_pager_subject_index (
        tenant_id,
        (CASE WHEN required_count = 0 THEN 2 WHEN filled_count >= required_count THEN 1 ELSE 0 END),
        (required_count - filled_count),
        subject_type,
        subject_id
    );

ALTER TABLE onepagers.one_pager_subject_index ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation_policy ON onepagers.one_pager_subject_index;
CREATE POLICY tenant_isolation_policy ON onepagers.one_pager_subject_index
    FOR ALL TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));

DO $$
BEGIN
    IF EXISTS (SELECT FROM pg_catalog.pg_user WHERE usename = 'easi_app') THEN
        EXECUTE 'GRANT SELECT, INSERT, UPDATE, DELETE ON onepagers.one_pager_subject_index TO easi_app';
    END IF;
    IF EXISTS (SELECT FROM pg_catalog.pg_user WHERE usename = 'easi_admin') THEN
        EXECUTE 'GRANT ALL PRIVILEGES ON onepagers.one_pager_subject_index TO easi_admin';
    END IF;
END $$;

-- ============================================================================
-- Backfill
-- ============================================================================

-- Capability
INSERT INTO onepagers.one_pager_subject_index
    (tenant_id, subject_type, subject_id, name, creator_actor_id, creator_email, created_at, last_updated_at, required_count, filled_count)
SELECT
    s.tenant_id, 'capability', s.id, s.name,
    COALESCE(v1.actor_id, ''), COALESCE(v1.actor_email, ''),
    COALESCE(v1.occurred_at, s.created_at),
    COALESCE(last_evt.updated_at, s.created_at),
    cfg.req_custom
        + (CASE WHEN 'description' = ANY(cfg.req_builtins) THEN 1 ELSE 0 END)
        + (CASE WHEN 'maturity' = ANY(cfg.req_builtins) THEN 1 ELSE 0 END)
        + (CASE WHEN 'experts' = ANY(cfg.req_builtins) THEN 1 ELSE 0 END)
        + (CASE WHEN 'name' = ANY(cfg.req_builtins) THEN 1 ELSE 0 END),
    COALESCE(facts.filled_custom, 0)
        + (CASE WHEN 'description' = ANY(cfg.req_builtins) AND s.description IS NOT NULL AND s.description <> '' THEN 1 ELSE 0 END)
        + (CASE WHEN 'maturity' = ANY(cfg.req_builtins) THEN 1 ELSE 0 END)
        + (CASE WHEN 'experts' = ANY(cfg.req_builtins) AND EXISTS (SELECT 1 FROM capabilitymapping.capability_experts e WHERE e.tenant_id = s.tenant_id AND e.capability_id = s.id) THEN 1 ELSE 0 END)
        + (CASE WHEN 'name' = ANY(cfg.req_builtins) THEN 1 ELSE 0 END)
FROM capabilitymapping.capabilities s
CROSS JOIN LATERAL (
    SELECT
        COALESCE((SELECT array_agg(f->>'id') FROM jsonb_array_elements(c.configuration->'builtInFields') f WHERE (f->>'required')::boolean), ARRAY[]::text[]) AS req_builtins,
        COALESCE((SELECT count(*)::int FROM jsonb_array_elements(c.configuration->'customFields') cf WHERE (cf->>'active')::boolean AND (cf->>'required')::boolean), 0) AS req_custom,
        COALESCE((SELECT array_agg(cf->>'id') FROM jsonb_array_elements(c.configuration->'customFields') cf WHERE (cf->>'active')::boolean AND (cf->>'required')::boolean), ARRAY[]::text[]) AS req_custom_ids
    FROM onepagers.one_pager_configurations c
    WHERE c.tenant_id = s.tenant_id AND c.subject_type = 'capability'
    LIMIT 1
) cfg
LEFT JOIN LATERAL (
    SELECT count(*)::int AS filled_custom
    FROM onepagers.one_pager_facts pf
    WHERE pf.tenant_id = s.tenant_id AND pf.subject_type = 'capability' AND pf.subject_id = s.id
      AND pf.value IS NOT NULL AND pf.field_id = ANY(cfg.req_custom_ids)
) facts ON TRUE
LEFT JOIN LATERAL (
    SELECT e.actor_id, e.actor_email, e.occurred_at
    FROM infrastructure.events e
    WHERE e.aggregate_id = s.id AND e.tenant_id = s.tenant_id AND e.version = 1
    LIMIT 1
) v1 ON TRUE
LEFT JOIN LATERAL (
    SELECT MAX(e.occurred_at) AS updated_at
    FROM infrastructure.events e
    WHERE e.aggregate_id = s.id AND e.tenant_id = s.tenant_id
) last_evt ON TRUE
ON CONFLICT (tenant_id, subject_type, subject_id) DO NOTHING;

-- Enterprise Capability
INSERT INTO onepagers.one_pager_subject_index
    (tenant_id, subject_type, subject_id, name, creator_actor_id, creator_email, created_at, last_updated_at, required_count, filled_count)
SELECT
    s.tenant_id, 'enterprise-capability', s.id, s.name,
    COALESCE(v1.actor_id, ''), COALESCE(v1.actor_email, ''),
    COALESCE(v1.occurred_at, s.created_at),
    COALESCE(last_evt.updated_at, s.created_at),
    cfg.req_custom
        + (CASE WHEN 'description' = ANY(cfg.req_builtins) THEN 1 ELSE 0 END)
        + (CASE WHEN 'category' = ANY(cfg.req_builtins) THEN 1 ELSE 0 END)
        + (CASE WHEN 'name' = ANY(cfg.req_builtins) THEN 1 ELSE 0 END),
    COALESCE(facts.filled_custom, 0)
        + (CASE WHEN 'description' = ANY(cfg.req_builtins) AND s.description IS NOT NULL AND s.description <> '' THEN 1 ELSE 0 END)
        + (CASE WHEN 'category' = ANY(cfg.req_builtins) AND s.category IS NOT NULL AND s.category <> '' THEN 1 ELSE 0 END)
        + (CASE WHEN 'name' = ANY(cfg.req_builtins) THEN 1 ELSE 0 END)
FROM enterprisearchitecture.enterprise_capabilities s
CROSS JOIN LATERAL (
    SELECT
        COALESCE((SELECT array_agg(f->>'id') FROM jsonb_array_elements(c.configuration->'builtInFields') f WHERE (f->>'required')::boolean), ARRAY[]::text[]) AS req_builtins,
        COALESCE((SELECT count(*)::int FROM jsonb_array_elements(c.configuration->'customFields') cf WHERE (cf->>'active')::boolean AND (cf->>'required')::boolean), 0) AS req_custom,
        COALESCE((SELECT array_agg(cf->>'id') FROM jsonb_array_elements(c.configuration->'customFields') cf WHERE (cf->>'active')::boolean AND (cf->>'required')::boolean), ARRAY[]::text[]) AS req_custom_ids
    FROM onepagers.one_pager_configurations c
    WHERE c.tenant_id = s.tenant_id AND c.subject_type = 'enterprise-capability'
    LIMIT 1
) cfg
LEFT JOIN LATERAL (
    SELECT count(*)::int AS filled_custom
    FROM onepagers.one_pager_facts pf
    WHERE pf.tenant_id = s.tenant_id AND pf.subject_type = 'enterprise-capability' AND pf.subject_id = s.id
      AND pf.value IS NOT NULL AND pf.field_id = ANY(cfg.req_custom_ids)
) facts ON TRUE
LEFT JOIN LATERAL (
    SELECT e.actor_id, e.actor_email, e.occurred_at
    FROM infrastructure.events e
    WHERE e.aggregate_id = s.id AND e.tenant_id = s.tenant_id AND e.version = 1
    LIMIT 1
) v1 ON TRUE
LEFT JOIN LATERAL (
    SELECT MAX(e.occurred_at) AS updated_at
    FROM infrastructure.events e
    WHERE e.aggregate_id = s.id AND e.tenant_id = s.tenant_id
) last_evt ON TRUE
ON CONFLICT (tenant_id, subject_type, subject_id) DO NOTHING;

-- Application Component
INSERT INTO onepagers.one_pager_subject_index
    (tenant_id, subject_type, subject_id, name, creator_actor_id, creator_email, created_at, last_updated_at, required_count, filled_count)
SELECT
    s.tenant_id, 'application', s.id, s.name,
    COALESCE(v1.actor_id, ''), COALESCE(v1.actor_email, ''),
    COALESCE(v1.occurred_at, s.created_at),
    COALESCE(last_evt.updated_at, s.created_at),
    cfg.req_custom
        + (CASE WHEN 'description' = ANY(cfg.req_builtins) THEN 1 ELSE 0 END)
        + (CASE WHEN 'experts' = ANY(cfg.req_builtins) THEN 1 ELSE 0 END)
        + (CASE WHEN 'name' = ANY(cfg.req_builtins) THEN 1 ELSE 0 END),
    COALESCE(facts.filled_custom, 0)
        + (CASE WHEN 'description' = ANY(cfg.req_builtins) AND s.description IS NOT NULL AND s.description <> '' THEN 1 ELSE 0 END)
        + (CASE WHEN 'experts' = ANY(cfg.req_builtins) AND EXISTS (SELECT 1 FROM architecturemodeling.application_component_experts e WHERE e.tenant_id = s.tenant_id AND e.component_id = s.id) THEN 1 ELSE 0 END)
        + (CASE WHEN 'name' = ANY(cfg.req_builtins) THEN 1 ELSE 0 END)
FROM architecturemodeling.application_components s
CROSS JOIN LATERAL (
    SELECT
        COALESCE((SELECT array_agg(f->>'id') FROM jsonb_array_elements(c.configuration->'builtInFields') f WHERE (f->>'required')::boolean), ARRAY[]::text[]) AS req_builtins,
        COALESCE((SELECT count(*)::int FROM jsonb_array_elements(c.configuration->'customFields') cf WHERE (cf->>'active')::boolean AND (cf->>'required')::boolean), 0) AS req_custom,
        COALESCE((SELECT array_agg(cf->>'id') FROM jsonb_array_elements(c.configuration->'customFields') cf WHERE (cf->>'active')::boolean AND (cf->>'required')::boolean), ARRAY[]::text[]) AS req_custom_ids
    FROM onepagers.one_pager_configurations c
    WHERE c.tenant_id = s.tenant_id AND c.subject_type = 'application'
    LIMIT 1
) cfg
LEFT JOIN LATERAL (
    SELECT count(*)::int AS filled_custom
    FROM onepagers.one_pager_facts pf
    WHERE pf.tenant_id = s.tenant_id AND pf.subject_type = 'application' AND pf.subject_id = s.id
      AND pf.value IS NOT NULL AND pf.field_id = ANY(cfg.req_custom_ids)
) facts ON TRUE
LEFT JOIN LATERAL (
    SELECT e.actor_id, e.actor_email, e.occurred_at
    FROM infrastructure.events e
    WHERE e.aggregate_id = s.id AND e.tenant_id = s.tenant_id AND e.version = 1
    LIMIT 1
) v1 ON TRUE
LEFT JOIN LATERAL (
    SELECT MAX(e.occurred_at) AS updated_at
    FROM infrastructure.events e
    WHERE e.aggregate_id = s.id AND e.tenant_id = s.tenant_id
) last_evt ON TRUE
WHERE s.is_deleted = FALSE
ON CONFLICT (tenant_id, subject_type, subject_id) DO NOTHING;

-- Acquired Entity
INSERT INTO onepagers.one_pager_subject_index
    (tenant_id, subject_type, subject_id, name, creator_actor_id, creator_email, created_at, last_updated_at, required_count, filled_count)
SELECT
    s.tenant_id, 'acquired-entity', s.id, s.name,
    COALESCE(v1.actor_id, ''), COALESCE(v1.actor_email, ''),
    COALESCE(v1.occurred_at, s.created_at),
    COALESCE(last_evt.updated_at, s.created_at),
    cfg.req_custom
        + (CASE WHEN 'acquisition-date' = ANY(cfg.req_builtins) THEN 1 ELSE 0 END)
        + (CASE WHEN 'integration-status' = ANY(cfg.req_builtins) THEN 1 ELSE 0 END)
        + (CASE WHEN 'name' = ANY(cfg.req_builtins) THEN 1 ELSE 0 END),
    COALESCE(facts.filled_custom, 0)
        + (CASE WHEN 'acquisition-date' = ANY(cfg.req_builtins) AND s.acquisition_date IS NOT NULL THEN 1 ELSE 0 END)
        + (CASE WHEN 'integration-status' = ANY(cfg.req_builtins) AND s.integration_status IS NOT NULL AND s.integration_status <> '' THEN 1 ELSE 0 END)
        + (CASE WHEN 'name' = ANY(cfg.req_builtins) THEN 1 ELSE 0 END)
FROM architecturemodeling.acquired_entities s
CROSS JOIN LATERAL (
    SELECT
        COALESCE((SELECT array_agg(f->>'id') FROM jsonb_array_elements(c.configuration->'builtInFields') f WHERE (f->>'required')::boolean), ARRAY[]::text[]) AS req_builtins,
        COALESCE((SELECT count(*)::int FROM jsonb_array_elements(c.configuration->'customFields') cf WHERE (cf->>'active')::boolean AND (cf->>'required')::boolean), 0) AS req_custom,
        COALESCE((SELECT array_agg(cf->>'id') FROM jsonb_array_elements(c.configuration->'customFields') cf WHERE (cf->>'active')::boolean AND (cf->>'required')::boolean), ARRAY[]::text[]) AS req_custom_ids
    FROM onepagers.one_pager_configurations c
    WHERE c.tenant_id = s.tenant_id AND c.subject_type = 'acquired-entity'
    LIMIT 1
) cfg
LEFT JOIN LATERAL (
    SELECT count(*)::int AS filled_custom
    FROM onepagers.one_pager_facts pf
    WHERE pf.tenant_id = s.tenant_id AND pf.subject_type = 'acquired-entity' AND pf.subject_id = s.id
      AND pf.value IS NOT NULL AND pf.field_id = ANY(cfg.req_custom_ids)
) facts ON TRUE
LEFT JOIN LATERAL (
    SELECT e.actor_id, e.actor_email, e.occurred_at
    FROM infrastructure.events e
    WHERE e.aggregate_id = s.id AND e.tenant_id = s.tenant_id AND e.version = 1
    LIMIT 1
) v1 ON TRUE
LEFT JOIN LATERAL (
    SELECT MAX(e.occurred_at) AS updated_at
    FROM infrastructure.events e
    WHERE e.aggregate_id = s.id AND e.tenant_id = s.tenant_id
) last_evt ON TRUE
WHERE s.is_deleted = FALSE
ON CONFLICT (tenant_id, subject_type, subject_id) DO NOTHING;

-- Vendor
INSERT INTO onepagers.one_pager_subject_index
    (tenant_id, subject_type, subject_id, name, creator_actor_id, creator_email, created_at, last_updated_at, required_count, filled_count)
SELECT
    s.tenant_id, 'vendor', s.id, s.name,
    COALESCE(v1.actor_id, ''), COALESCE(v1.actor_email, ''),
    COALESCE(v1.occurred_at, s.created_at),
    COALESCE(last_evt.updated_at, s.created_at),
    cfg.req_custom
        + (CASE WHEN 'implementation-partner' = ANY(cfg.req_builtins) THEN 1 ELSE 0 END)
        + (CASE WHEN 'notes' = ANY(cfg.req_builtins) THEN 1 ELSE 0 END)
        + (CASE WHEN 'name' = ANY(cfg.req_builtins) THEN 1 ELSE 0 END),
    COALESCE(facts.filled_custom, 0)
        + (CASE WHEN 'implementation-partner' = ANY(cfg.req_builtins) AND s.implementation_partner IS NOT NULL AND s.implementation_partner <> '' THEN 1 ELSE 0 END)
        + (CASE WHEN 'notes' = ANY(cfg.req_builtins) AND s.notes IS NOT NULL AND s.notes <> '' THEN 1 ELSE 0 END)
        + (CASE WHEN 'name' = ANY(cfg.req_builtins) THEN 1 ELSE 0 END)
FROM architecturemodeling.vendors s
CROSS JOIN LATERAL (
    SELECT
        COALESCE((SELECT array_agg(f->>'id') FROM jsonb_array_elements(c.configuration->'builtInFields') f WHERE (f->>'required')::boolean), ARRAY[]::text[]) AS req_builtins,
        COALESCE((SELECT count(*)::int FROM jsonb_array_elements(c.configuration->'customFields') cf WHERE (cf->>'active')::boolean AND (cf->>'required')::boolean), 0) AS req_custom,
        COALESCE((SELECT array_agg(cf->>'id') FROM jsonb_array_elements(c.configuration->'customFields') cf WHERE (cf->>'active')::boolean AND (cf->>'required')::boolean), ARRAY[]::text[]) AS req_custom_ids
    FROM onepagers.one_pager_configurations c
    WHERE c.tenant_id = s.tenant_id AND c.subject_type = 'vendor'
    LIMIT 1
) cfg
LEFT JOIN LATERAL (
    SELECT count(*)::int AS filled_custom
    FROM onepagers.one_pager_facts pf
    WHERE pf.tenant_id = s.tenant_id AND pf.subject_type = 'vendor' AND pf.subject_id = s.id
      AND pf.value IS NOT NULL AND pf.field_id = ANY(cfg.req_custom_ids)
) facts ON TRUE
LEFT JOIN LATERAL (
    SELECT e.actor_id, e.actor_email, e.occurred_at
    FROM infrastructure.events e
    WHERE e.aggregate_id = s.id AND e.tenant_id = s.tenant_id AND e.version = 1
    LIMIT 1
) v1 ON TRUE
LEFT JOIN LATERAL (
    SELECT MAX(e.occurred_at) AS updated_at
    FROM infrastructure.events e
    WHERE e.aggregate_id = s.id AND e.tenant_id = s.tenant_id
) last_evt ON TRUE
WHERE s.is_deleted = FALSE
ON CONFLICT (tenant_id, subject_type, subject_id) DO NOTHING;

-- Internal Team
INSERT INTO onepagers.one_pager_subject_index
    (tenant_id, subject_type, subject_id, name, creator_actor_id, creator_email, created_at, last_updated_at, required_count, filled_count)
SELECT
    s.tenant_id, 'internal-team', s.id, s.name,
    COALESCE(v1.actor_id, ''), COALESCE(v1.actor_email, ''),
    COALESCE(v1.occurred_at, s.created_at),
    COALESCE(last_evt.updated_at, s.created_at),
    cfg.req_custom
        + (CASE WHEN 'department' = ANY(cfg.req_builtins) THEN 1 ELSE 0 END)
        + (CASE WHEN 'contact-person' = ANY(cfg.req_builtins) THEN 1 ELSE 0 END)
        + (CASE WHEN 'name' = ANY(cfg.req_builtins) THEN 1 ELSE 0 END),
    COALESCE(facts.filled_custom, 0)
        + (CASE WHEN 'department' = ANY(cfg.req_builtins) AND s.department IS NOT NULL AND s.department <> '' THEN 1 ELSE 0 END)
        + (CASE WHEN 'contact-person' = ANY(cfg.req_builtins) AND s.contact_person IS NOT NULL AND s.contact_person <> '' THEN 1 ELSE 0 END)
        + (CASE WHEN 'name' = ANY(cfg.req_builtins) THEN 1 ELSE 0 END)
FROM architecturemodeling.internal_teams s
CROSS JOIN LATERAL (
    SELECT
        COALESCE((SELECT array_agg(f->>'id') FROM jsonb_array_elements(c.configuration->'builtInFields') f WHERE (f->>'required')::boolean), ARRAY[]::text[]) AS req_builtins,
        COALESCE((SELECT count(*)::int FROM jsonb_array_elements(c.configuration->'customFields') cf WHERE (cf->>'active')::boolean AND (cf->>'required')::boolean), 0) AS req_custom,
        COALESCE((SELECT array_agg(cf->>'id') FROM jsonb_array_elements(c.configuration->'customFields') cf WHERE (cf->>'active')::boolean AND (cf->>'required')::boolean), ARRAY[]::text[]) AS req_custom_ids
    FROM onepagers.one_pager_configurations c
    WHERE c.tenant_id = s.tenant_id AND c.subject_type = 'internal-team'
    LIMIT 1
) cfg
LEFT JOIN LATERAL (
    SELECT count(*)::int AS filled_custom
    FROM onepagers.one_pager_facts pf
    WHERE pf.tenant_id = s.tenant_id AND pf.subject_type = 'internal-team' AND pf.subject_id = s.id
      AND pf.value IS NOT NULL AND pf.field_id = ANY(cfg.req_custom_ids)
) facts ON TRUE
LEFT JOIN LATERAL (
    SELECT e.actor_id, e.actor_email, e.occurred_at
    FROM infrastructure.events e
    WHERE e.aggregate_id = s.id AND e.tenant_id = s.tenant_id AND e.version = 1
    LIMIT 1
) v1 ON TRUE
LEFT JOIN LATERAL (
    SELECT MAX(e.occurred_at) AS updated_at
    FROM infrastructure.events e
    WHERE e.aggregate_id = s.id AND e.tenant_id = s.tenant_id
) last_evt ON TRUE
WHERE s.is_deleted = FALSE
ON CONFLICT (tenant_id, subject_type, subject_id) DO NOTHING;
