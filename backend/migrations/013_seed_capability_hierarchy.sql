-- Migration: Seed Capability Hierarchy
-- Spec: 034_Capability_Tree_Sidebar_pending.md
-- Description: Creates sample capability hierarchy for testing the capability tree sidebar

-- ============================================================================
-- Seed Data for Capability Hierarchy
-- ============================================================================

-- L1 Capabilities (Root Level)
INSERT INTO capabilities (id, tenant_id, name, description, parent_id, level, maturity_level, status, created_at)
VALUES
    ('cap-l1-customer-mgmt', 'default', 'Customer Management', 'Capabilities related to customer lifecycle and relationship management', NULL, 'L1', 'Developing', 'Active', NOW()),
    ('cap-l1-finance', 'default', 'Finance', 'Financial operations and management capabilities', NULL, 'L1', 'Defined', 'Active', NOW()),
    ('cap-l1-product-mgmt', 'default', 'Product Management', 'Product lifecycle and catalog management capabilities', NULL, 'L1', 'Initial', 'Active', NOW())
ON CONFLICT (tenant_id, id) DO NOTHING;

-- L2 Capabilities (Second Level)
INSERT INTO capabilities (id, tenant_id, name, description, parent_id, level, maturity_level, status, created_at)
VALUES
    ('cap-l2-onboarding', 'default', 'Customer Onboarding', 'New customer registration and onboarding processes', 'cap-l1-customer-mgmt', 'L2', 'Developing', 'Active', NOW()),
    ('cap-l2-support', 'default', 'Customer Support', 'Customer service and support capabilities', 'cap-l1-customer-mgmt', 'L2', 'Managed', 'Active', NOW()),
    ('cap-l2-billing', 'default', 'Billing', 'Invoice generation and payment processing', 'cap-l1-finance', 'L2', 'Defined', 'Active', NOW()),
    ('cap-l2-reporting', 'default', 'Financial Reporting', 'Financial reports and analytics', 'cap-l1-finance', 'L2', 'Developing', 'Active', NOW()),
    ('cap-l2-catalog', 'default', 'Product Catalog', 'Product information and catalog management', 'cap-l1-product-mgmt', 'L2', 'Initial', 'Active', NOW())
ON CONFLICT (tenant_id, id) DO NOTHING;

-- L3 Capabilities (Third Level)
INSERT INTO capabilities (id, tenant_id, name, description, parent_id, level, maturity_level, status, created_at)
VALUES
    ('cap-l3-ticketing', 'default', 'Ticketing System', 'Support ticket management and tracking', 'cap-l2-support', 'L3', 'Managed', 'Active', NOW()),
    ('cap-l3-kyc', 'default', 'KYC Verification', 'Know Your Customer identity verification', 'cap-l2-onboarding', 'L3', 'Developing', 'Active', NOW()),
    ('cap-l3-pricing', 'default', 'Pricing Management', 'Product pricing rules and management', 'cap-l2-catalog', 'L3', 'Initial', 'Active', NOW())
ON CONFLICT (tenant_id, id) DO NOTHING;

-- ============================================================================
-- Migration complete
-- ============================================================================
