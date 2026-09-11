-- Migration: Backfill the OnePagers custom-field definition cache (spec 217)
-- Description: Seeds the cache added by 160 from the custom-field definitions still embedded in
--              the legacy OnePagers configuration documents, so every existing field keeps its
--              identity, type, options and bounds. Required-ness stays on the configuration and
--              is not part of the definition. Rows are marked pending_transfer so the backend
--              hands them to MetaModel on startup; MetaModel's published events then confirm them.

INSERT INTO onepagers.custom_field_definition_cache (tenant_id, field_id, subject_type, definition, pending_transfer)
SELECT c.tenant_id, field->>'id', c.subject_type, field - 'required', TRUE
FROM onepagers.one_pager_configurations c,
     jsonb_array_elements(COALESCE(c.configuration->'customFields', '[]'::jsonb)) AS field
WHERE field->>'id' IS NOT NULL AND field ? 'name'
ON CONFLICT (tenant_id, field_id) DO NOTHING;
