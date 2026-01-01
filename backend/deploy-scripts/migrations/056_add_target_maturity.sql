-- Migration: Add Target Maturity to Enterprise Capabilities
-- Spec: 102_Standardization_Analysis_pending.md
-- Description: Adds target_maturity column for maturity gap analysis

ALTER TABLE enterprise_capabilities
ADD COLUMN IF NOT EXISTS target_maturity INTEGER CHECK (target_maturity IS NULL OR (target_maturity >= 0 AND target_maturity <= 99));

CREATE INDEX IF NOT EXISTS idx_ec_target_maturity ON enterprise_capabilities(tenant_id, target_maturity) WHERE target_maturity IS NOT NULL;
