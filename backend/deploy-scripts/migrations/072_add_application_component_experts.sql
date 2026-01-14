-- Migration: Add Application Component Experts
-- Spec: 114_ApplicationComponent_Experts
-- Description: Add experts table for application components with role autocomplete support

CREATE TABLE IF NOT EXISTS application_component_experts (
    component_id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(50) NOT NULL,
    expert_name VARCHAR(500) NOT NULL,
    expert_role VARCHAR(500) NOT NULL,
    contact_info VARCHAR(500) NOT NULL,
    added_at TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_application_component_experts_component
    ON application_component_experts(tenant_id, component_id);

CREATE INDEX IF NOT EXISTS idx_application_component_experts_role
    ON application_component_experts(tenant_id, expert_role);

ALTER TABLE application_component_experts ENABLE ROW LEVEL SECURITY;

CREATE POLICY application_component_experts_tenant_isolation ON application_component_experts
    USING (tenant_id = current_setting('app.current_tenant_id', TRUE));
