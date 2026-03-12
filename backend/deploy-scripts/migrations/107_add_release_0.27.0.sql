-- Migration: Add Release 0.27.0
-- Description: Adds release notes for version 0.27.0

INSERT INTO releases (version, release_date, notes, created_at) VALUES
('0.27.0', '2026-03-12', '## What''s New in v0.27.0

### Major

- **Auto-Generate View from Entity** — Right-click any entity on the Architecture Canvas to instantly generate a focused view containing that entity and everything connected to it. No more manually adding dozens of related elements one by one.

- **Smarter Architecture Assistant** — The AI assistant now understands the EASI domain model (capability hierarchies, strategy pillars, TIME classification, fit analysis) and gives contextually accurate answers instead of generic responses.

- **Massively Expanded Assistant Tool Coverage** — The assistant can now work with ~80 operations across the full platform: enterprise capabilities, strategy analysis, fit scores, capability dependencies, component origins (vendors, acquired entities, internal teams), value streams, business domains, and more.

- **Agent Permission Ceiling** — The AI assistant is now restricted to architecture-related operations only. Even admin users'' assistants cannot access user management, access delegation, or audit APIs, preventing privilege escalation through AI workflows.

### Bugs

- Fixed theme styling issue on the frontend', CURRENT_TIMESTAMP)
ON CONFLICT (version) DO UPDATE SET
  release_date = EXCLUDED.release_date,
  notes = EXCLUDED.notes;
