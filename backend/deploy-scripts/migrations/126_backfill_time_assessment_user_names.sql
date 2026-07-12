INSERT INTO architecturedirection.reference_name_cache (tenant_id, entity_type, entity_id, name)
SELECT DISTINCT ON (tenant_id, email) tenant_id, 'user', email, name
FROM auth.users
WHERE name IS NOT NULL AND name <> ''
ORDER BY tenant_id, email, updated_at DESC
ON CONFLICT (tenant_id, entity_type, entity_id) DO UPDATE SET name = EXCLUDED.name;

UPDATE architecturedirection.time_assessments ta
SET assessed_by_name = u.name
FROM auth.users u
WHERE ta.tenant_id = u.tenant_id AND ta.assessed_by = u.email
  AND u.name IS NOT NULL AND u.name <> ''
  AND ta.assessed_by_name = ta.assessed_by;
