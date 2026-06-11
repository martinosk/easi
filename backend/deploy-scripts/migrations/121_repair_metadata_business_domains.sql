-- Spec 172 follow-up data repair.
-- enterprisearchitecture.domain_capability_metadata rows were missing (or carried a
-- raw domain ID as) business_domain_id/business_domain_name: the projector resolved
-- domain names from the metadata table itself, which never contains a name the first
-- time a domain is used. The projector now resolves names from the capabilitymapping
-- read model; this migration repairs rows written before that fix.

UPDATE enterprisearchitecture.domain_capability_metadata dcm
SET business_domain_id   = assignment.business_domain_id,
    business_domain_name = assignment.name
FROM (
    SELECT dca.tenant_id, target.l1_capability_id, dca.business_domain_id, bd.name
    FROM capabilitymapping.domain_capability_assignments dca
    JOIN enterprisearchitecture.domain_capability_metadata target
        ON target.tenant_id = dca.tenant_id AND target.capability_id = dca.capability_id
    JOIN capabilitymapping.business_domains bd
        ON bd.tenant_id = dca.tenant_id AND bd.id = dca.business_domain_id
) assignment
WHERE dcm.tenant_id = assignment.tenant_id
  AND dcm.l1_capability_id = assignment.l1_capability_id;
