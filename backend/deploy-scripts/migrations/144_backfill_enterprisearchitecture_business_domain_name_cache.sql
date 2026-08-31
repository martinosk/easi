-- Migration: Backfill Enterprise Architecture business domain name cache (spec 209 WP4)
-- Description: Populates enterprisearchitecture.business_domain_name_cache from
--              capabilitymapping.business_domains so lookups are complete immediately
--              after deployment. BusinessDomainNameCacheProjector keeps it current from
--              this point on. Idempotent.

INSERT INTO enterprisearchitecture.business_domain_name_cache (tenant_id, business_domain_id, name)
SELECT tenant_id, id, name
FROM capabilitymapping.business_domains
ON CONFLICT (tenant_id, business_domain_id) DO UPDATE SET
    name = EXCLUDED.name;
