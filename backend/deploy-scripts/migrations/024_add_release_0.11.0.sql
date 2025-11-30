-- Migration: Add Release 0.11.0
-- Description: Adds release notes for version 0.11.0

INSERT INTO releases (version, release_date, notes, created_at) VALUES
('0.11.0', '2025-11-30', '## What''s New in 0.11.0

### Enhanced Business Domain Visualization
- **Nested Capability View**: View L2, L3, and L4 capabilities nested inside their L1 parents with color-coded levels (Blue/Purple/Pink/Orange)
- **Depth Selector**: Control how deep to visualize - from L1 only up to L1-L2-L3-L4
- **Collapsible Sidebars**: Both the Business Domains list and Capability Explorer sidebars can now be collapsed to maximize the grid view area
- **Improved Drag Reordering**: Fixed drag-and-drop ordering so capabilities maintain correct positions when rearranged

### Automatic Domain Assignment Transfer
- **Smart Re-parenting**: When an L1 capability gets assigned a parent (becoming L2), its business domain assignments automatically transfer to the new L1 parent
- **Maintains Consistency**: Business domains always contain only L1 capabilities - the system handles hierarchy changes automatically

### UI Improvements
- Refined visual styling throughout the Business Domains page
', CURRENT_TIMESTAMP)
ON CONFLICT (version) DO UPDATE SET
  release_date = EXCLUDED.release_date,
  notes = EXCLUDED.notes;
