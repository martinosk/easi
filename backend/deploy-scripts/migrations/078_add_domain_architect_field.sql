-- Migration: Add Domain Architect Field
-- Spec: 117_Portfolio_Metadata_Foundation
-- Description: Add domain architect user ID to business domains

ALTER TABLE business_domains
ADD COLUMN IF NOT EXISTS domain_architect_id VARCHAR(36);
