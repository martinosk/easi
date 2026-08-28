-- Migration: Add Release 2.3.0
-- Description: Adds release notes for version 2.3.0

INSERT INTO releases.releases (version, release_date, notes, created_at) VALUES
('2.3.0', '2026-08-28', '## What''s New in v2.3.0

### Minor
- Journey milestones can now be reordered by drag-and-drop or keyboard in the milestone drawer, letting architects arrange them into the sequence the plan will actually execute.
- A milestone whose target period contradicts its position in the sequence now shows a schedule-conflict marker naming the period it is out of step with.', CURRENT_TIMESTAMP)
ON CONFLICT (version) DO UPDATE SET
  release_date = EXCLUDED.release_date,
  notes = EXCLUDED.notes;
