-- Migration: Fix Release v0.20.0 version format
-- Description: Removes the 'v' prefix from version 0.20.0 to match semver validation

UPDATE releases SET version = '0.20.0' WHERE version = 'v0.20.0';
