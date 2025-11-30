-- Migration: Remove capability_level CHECK constraint from domain_capability_assignments
-- Spec: 059_BusinessDomain_Assignment_Consistency_pending.md
-- Reason: Domain invariants must be enforced by the domain model, not the database

ALTER TABLE domain_capability_assignments DROP CONSTRAINT IF EXISTS domain_capability_assignments_capability_level_check;
