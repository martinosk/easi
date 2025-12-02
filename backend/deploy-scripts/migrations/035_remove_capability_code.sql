-- Migration: Remove capability_code from domain_capability_assignments
-- Description: The capability_code field was redundantly storing the capability ID.
-- This migration removes the unused column and updates ordering to use capability_name.

ALTER TABLE domain_capability_assignments DROP COLUMN IF EXISTS capability_code;
