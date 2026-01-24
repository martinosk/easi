-- Migration: Add Release 0.21.1
-- Description: Adds release notes for version 0.21.1

INSERT INTO releases (version, release_date, notes, created_at) VALUES
('0.21.1', '2026-01-24', '## What''s New in v0.21.1

### Major

- **Portfolio metadata foundation** - Added support for tracking application origins through three new entity types:
  - **Acquired Entities**: Track applications that came from M&A activity, including acquisition date and integration status
  - **Vendors**: Track applications purchased from external software vendors
  - **Internal Teams**: Track applications built by internal development teams
- **Origin relationships visualization** - Draw relationships on the canvas from origin entities to application components to visualize where your applications come from
- **Component origins view** - View all origin relationships for an application component in its details panel
- **Domain architect assignment** - Assign architects to business domains to identify ownership and accountability

### Bugs

- Fixed a bug where re-linking a component to a previously linked origin entity (after clearing) caused a 500 error

### API

- New endpoints for Acquired Entities: `POST/GET/PUT/DELETE /acquired-entities`
- New endpoints for Vendors: `POST/GET/PUT/DELETE /vendors`
- New endpoints for Internal Teams: `POST/GET/PUT/DELETE /internal-teams`
- New endpoint: `POST /components/{id}/origin/acquired-via` - Link component to acquired entity
- New endpoint: `POST /components/{id}/origin/purchased-from` - Link component to vendor
- New endpoint: `POST /components/{id}/origin/built-by` - Link component to internal team
- New endpoint: `GET /components/{id}/origins` - Get all origin relationships for a component
- New endpoint: `DELETE /origin-relationships/{type}/{id}` - Remove origin relationships
- Updated `PUT /business-domains/{id}` to support domain architect assignment', CURRENT_TIMESTAMP)
ON CONFLICT (version) DO UPDATE SET
  release_date = EXCLUDED.release_date,
  notes = EXCLUDED.notes;
