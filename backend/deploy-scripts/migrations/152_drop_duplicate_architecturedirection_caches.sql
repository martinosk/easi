-- Migration: Drop the duplicate Architecture Direction caches (spec 212)
-- Description: The TIME suggestion query was the last reader of ea_realization_cache and
--              domain_capability_metadata. It now reads realization_cache and
--              capability_node_cache, which carry the same facts, are fed by the same events
--              and are already backfilled by migrations 138 and 146. Component names come from
--              reference_name_cache. Nothing else references the two dropped tables.

DROP TABLE IF EXISTS architecturedirection.ea_realization_cache;
DROP TABLE IF EXISTS architecturedirection.domain_capability_metadata;
