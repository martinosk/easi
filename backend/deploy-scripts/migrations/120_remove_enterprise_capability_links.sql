-- Spec 172 — Direction Is the Association.
-- Standalone enterprise-capability linking is removed: a domain capability is
-- associated with an enterprise capability only through the sources of its
-- active direction. Link rows are hard-deleted (dropping the tables) — there is
-- no migration of historical links into directions.

DROP TABLE IF EXISTS enterprisearchitecture.enterprise_capability_links;
DROP TABLE IF EXISTS enterprisearchitecture.capability_link_blocking;

-- includedCapabilityCount and domainCount are now derived from the composition
-- algorithm at read time; the persisted counters are obsolete.
ALTER TABLE enterprisearchitecture.enterprise_capabilities DROP COLUMN IF EXISTS link_count;
ALTER TABLE enterprisearchitecture.enterprise_capabilities DROP COLUMN IF EXISTS domain_count;
