-- Migration: Backfill strategy pillar caches from MetaModel
-- Description: Populate CapabilityMapping and EnterpriseArchitecture pillar caches with existing MetaModel data

-- Backfill CapabilityMapping cache
INSERT INTO cm_strategy_pillar_cache (id, tenant_id, name, description, active, fit_scoring_enabled, fit_criteria, fit_type)
SELECT 
    pillar->>'id' as id,
    tenant_id,
    pillar->>'name' as name,
    COALESCE(pillar->>'description', '') as description,
    COALESCE((pillar->>'active')::boolean, true) as active,
    COALESCE((pillar->>'fitScoringEnabled')::boolean, false) as fit_scoring_enabled,
    pillar->>'fitCriteria' as fit_criteria,
    pillar->>'fitType' as fit_type
FROM meta_model_configurations,
     jsonb_array_elements(strategy_pillars) as pillar
ON CONFLICT (id, tenant_id) DO UPDATE
SET 
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    active = EXCLUDED.active,
    fit_scoring_enabled = EXCLUDED.fit_scoring_enabled,
    fit_criteria = EXCLUDED.fit_criteria,
    fit_type = EXCLUDED.fit_type;

-- Backfill EnterpriseArchitecture cache
INSERT INTO ea_strategy_pillar_cache (id, tenant_id, name, description, active, fit_scoring_enabled, fit_criteria, fit_type)
SELECT 
    pillar->>'id' as id,
    tenant_id,
    pillar->>'name' as name,
    COALESCE(pillar->>'description', '') as description,
    COALESCE((pillar->>'active')::boolean, true) as active,
    COALESCE((pillar->>'fitScoringEnabled')::boolean, false) as fit_scoring_enabled,
    pillar->>'fitCriteria' as fit_criteria,
    pillar->>'fitType' as fit_type
FROM meta_model_configurations,
     jsonb_array_elements(strategy_pillars) as pillar
ON CONFLICT (id, tenant_id) DO UPDATE
SET 
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    active = EXCLUDED.active,
    fit_scoring_enabled = EXCLUDED.fit_scoring_enabled,
    fit_criteria = EXCLUDED.fit_criteria,
    fit_type = EXCLUDED.fit_type;
