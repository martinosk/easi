-- Migration: Add Release 2.5.0
-- Description: Adds release notes for version 2.5.0

INSERT INTO releases.releases (version, release_date, notes, created_at) VALUES
('2.5.0', '2026-09-14', '## What''s New in v2.5.0

### Major
- Added application composition: an application can now be modelled as a suite, platform or monolith that contains other applications as parts, declared per part as a composition (the part cannot outlive its parent) or an aggregation (the part stands on its own). Containment is capped at two levels: a part has exactly one parent and cannot accept parts of its own.
- Parts are attached on the Architecture Canvas either by dragging a connection between two applications and choosing "Part of (composition)" or "Part of (aggregation)" in the connection dialog, or by clicking a parent''s handle and creating a new composed or aggregated part directly. Parent and part are joined by a UML containment edge with a filled diamond for composition and a hollow diamond for aggregation, and the edge''s context menu offers "Detach Part".
- Deleting a parent deletes its composed parts and releases its aggregated parts as standalone applications; the delete confirmation in the navigation tree and on the canvas lists exactly what will be deleted and what will be released.
- The API exposes each application''s parent (with kind) and its parts, with `x-attach-to` and `x-detach` affordances on the component representation and new attach and detach operations addressed at the component.

### Minor
- The applications list annotates each part with the application it belongs to and each parent with its parts.

### Bugs
- Context menus on the Architecture Canvas now close when clicking anywhere on the canvas background instead of staying open until another menu was opened.', CURRENT_TIMESTAMP)
ON CONFLICT (version) DO UPDATE SET
  release_date = EXCLUDED.release_date,
  notes = EXCLUDED.notes;
