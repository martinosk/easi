-- Migration: Drop the Architecture Direction business-domain-name cache (spec 212 follow-up)
-- Description: The cache was written by its own projector and read only by
--              domain_capability_metadata's projector, which spec 212 deleted along with the
--              table it fed. Nothing reads it any more; capability_node_cache already carries
--              business_domain_name for the surviving readers. OnePagers keeps its own
--              independent cache of the same name in its own schema — untouched here.

DROP TABLE IF EXISTS architecturedirection.business_domain_name_cache;
