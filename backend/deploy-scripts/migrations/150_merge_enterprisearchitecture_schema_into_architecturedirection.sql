-- Migration: 150_merge_enterprisearchitecture_schema_into_architecturedirection
-- Purpose: Relocate EnterpriseArchitecture into Architecture Direction (spec 210, roadmap SD1 / H1-1).
-- EnterpriseArchitecture's language belongs to Architecture Direction, so its read-model tables
-- re-parent into the architecturedirection schema. The global event store is untouched: stored
-- events never move and published event type strings are unchanged.
--
-- architecturedirection.enterprise_capability_cache (migration 137) existed only because the
-- enterprise capability read model lived in another context. That table is now local, so the
-- cache is redundant and is dropped along with its projector.
--
-- Migration 151 re-seeds the two backfills that addressed the enterprisearchitecture schema.

ALTER TABLE enterprisearchitecture.enterprise_capabilities SET SCHEMA architecturedirection;
ALTER TABLE enterprisearchitecture.enterprise_strategic_importance SET SCHEMA architecturedirection;
ALTER TABLE enterprisearchitecture.domain_capability_metadata SET SCHEMA architecturedirection;
ALTER TABLE enterprisearchitecture.business_domain_name_cache SET SCHEMA architecturedirection;
ALTER TABLE enterprisearchitecture.ea_strategy_pillar_cache SET SCHEMA architecturedirection;
ALTER TABLE enterprisearchitecture.ea_realization_cache SET SCHEMA architecturedirection;
ALTER TABLE enterprisearchitecture.ea_importance_cache SET SCHEMA architecturedirection;
ALTER TABLE enterprisearchitecture.ea_fit_score_cache SET SCHEMA architecturedirection;

DROP TABLE IF EXISTS architecturedirection.enterprise_capability_cache;

DROP SCHEMA enterprisearchitecture;
