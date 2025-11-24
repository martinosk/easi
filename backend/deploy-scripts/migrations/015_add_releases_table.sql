-- Migration: Add Releases Table
-- Spec: 043_Release_Notes_pending.md
-- Description: Creates the releases table for storing release notes
-- Note: This table is system-wide, not tenant-specific

CREATE TABLE IF NOT EXISTS releases (
    version VARCHAR(50) PRIMARY KEY,
    release_date TIMESTAMP NOT NULL,
    notes TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_releases_release_date ON releases(release_date DESC);
