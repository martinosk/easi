-- Migration: Add Release 1.1.0
-- Description: Adds release notes for version 1.1.0

INSERT INTO releases.releases (version, release_date, notes, created_at) VALUES
('1.1.0', '2026-05-03', '## What''s New in v1.1.0

### Major
- **Click-on-handle creates related entities.** Click any free handle on a canvas entity to open a picker of valid related entity types, then create the new entity, link it to the source, and place it on the current view in one gesture. Available from Component, Capability, Acquired Entity, Vendor, and Internal Team handles. The picker is driven entirely by what the backend advertises, so it always matches server-side rules (e.g. an L4 capability does not offer "child capability"). Component handles can create related Components (Triggers or Serves), Capabilities (parent or realization), and Origin entities (Acquired Entity, Vendor, Internal Team). Works in both regular and dynamic-view modes.
- **Redesigned context menus.** Context menus across the canvas, navigation tree, business-domain pages, and the new create-related picker now use a radial layout for up to six items (linear above), with an icon and a short description on every item.

### Minor
- **New `_links["x-related"]` HATEOAS array on entities.** Component, Capability, AcquiredEntity, Vendor, and InternalTeam responses now include an `x-related` array under `_links`, with one entry per available relation. Each entry declares `href`, `methods`, `title`, `targetType`, and `relationType`. Entries advertising `POST` indicate that a related entity can be created from this one.
- **New `/reference/x-related-links` endpoint** documents the `RelatedLink` shape in the OpenAPI schema.', CURRENT_TIMESTAMP)
ON CONFLICT (version) DO UPDATE SET
  release_date = EXCLUDED.release_date,
  notes = EXCLUDED.notes;
