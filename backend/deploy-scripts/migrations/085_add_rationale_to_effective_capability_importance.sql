-- Migration: Add Rationale to Effective Capability Importance
-- Description: Denormalize rationale into effective_capability_importance table
-- to avoid complex joins in strategic fit analysis queries.

ALTER TABLE effective_capability_importance
ADD COLUMN rationale TEXT NOT NULL DEFAULT '';

-- Migration complete
