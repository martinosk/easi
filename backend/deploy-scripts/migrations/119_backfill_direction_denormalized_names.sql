INSERT INTO architecturedirection.reference_name_cache (tenant_id, entity_type, entity_id, name)
SELECT tenant_id, 'capability', id, name
FROM capabilitymapping.capabilities
ON CONFLICT (tenant_id, entity_type, entity_id) DO UPDATE SET name = EXCLUDED.name;

INSERT INTO architecturedirection.reference_name_cache (tenant_id, entity_type, entity_id, name)
SELECT tenant_id, 'business_domain', id, name
FROM capabilitymapping.business_domains
ON CONFLICT (tenant_id, entity_type, entity_id) DO UPDATE SET name = EXCLUDED.name;

INSERT INTO architecturedirection.reference_name_cache (tenant_id, entity_type, entity_id, name)
SELECT tenant_id, 'application', id, name
FROM architecturemodeling.application_components
WHERE is_deleted = FALSE
ON CONFLICT (tenant_id, entity_type, entity_id) DO UPDATE SET name = EXCLUDED.name;

INSERT INTO architecturedirection.capability_domain_cache (tenant_id, capability_id, business_domain_id)
SELECT tenant_id, capability_id, business_domain_id
FROM capabilitymapping.domain_capability_assignments
ON CONFLICT (tenant_id, capability_id) DO UPDATE SET business_domain_id = EXCLUDED.business_domain_id;

UPDATE architecturedirection.direction_source_capabilities dsc
SET capability_name = c.name
FROM capabilitymapping.capabilities c
WHERE dsc.tenant_id = c.tenant_id AND dsc.capability_id = c.id
  AND dsc.capability_name IS NULL;

UPDATE architecturedirection.direction_source_capabilities dsc
SET business_domain_id = dca.business_domain_id,
    business_domain_name = bd.name
FROM capabilitymapping.domain_capability_assignments dca
JOIN capabilitymapping.business_domains bd
  ON bd.tenant_id = dca.tenant_id AND bd.id = dca.business_domain_id
WHERE dsc.tenant_id = dca.tenant_id AND dsc.capability_id = dca.capability_id
  AND dsc.business_domain_id IS NULL;

UPDATE architecturedirection.standard_applications sa
SET application_name = ac.name
FROM architecturemodeling.application_components ac
WHERE sa.tenant_id = ac.tenant_id AND sa.application_id = ac.id
  AND sa.application_name IS NULL;

UPDATE architecturedirection.standard_application_history sah
SET application_name = ac.name
FROM architecturemodeling.application_components ac
WHERE sah.tenant_id = ac.tenant_id AND sah.application_id = ac.id
  AND sah.application_name IS NULL;

UPDATE architecturedirection.standard_application_history sah
SET previous_application_name = ac.name
FROM architecturemodeling.application_components ac
WHERE sah.tenant_id = ac.tenant_id AND sah.previous_application_id = ac.id
  AND sah.previous_application_name IS NULL;
