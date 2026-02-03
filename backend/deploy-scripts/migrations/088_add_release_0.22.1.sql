
INSERT INTO releases (version, release_date, notes, created_at) VALUES
('0.22.1', '2026-02-03', '## What''s New in v0.22.1

### Major
- Reintroduced auto-layout for architecture views to quickly organize components
- Updated business domain screens to respect HATEOAS permissions for creating/editing strategic importance and showing the "Add domain" action
- Simplified stakeholder experience by hiding the capability overview on business domains

### Bugs
- Fixed multi-select dragging so grouped items move correctly', CURRENT_TIMESTAMP)
ON CONFLICT (version) DO UPDATE SET
  release_date = EXCLUDED.release_date,
  notes = EXCLUDED.notes;