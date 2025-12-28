-- Migration: Add Strategy Pillars Column
-- Spec: 098_StrategyPillars_Settings
-- Description: Adds strategy_pillars column to meta_model_configurations table

ALTER TABLE meta_model_configurations
ADD COLUMN IF NOT EXISTS strategy_pillars JSONB NOT NULL DEFAULT '[]'::jsonb;

-- Initialize existing rows with default pillars
UPDATE meta_model_configurations
SET strategy_pillars = '[
    {"id": "pillar-always-on", "name": "Always On", "description": "Core capabilities that must always be operational", "active": true},
    {"id": "pillar-grow", "name": "Grow", "description": "Capabilities driving business growth", "active": true},
    {"id": "pillar-transform", "name": "Transform", "description": "Capabilities enabling digital transformation", "active": true}
]'::jsonb
WHERE strategy_pillars = '[]'::jsonb;
