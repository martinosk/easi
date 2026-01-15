-- Migration: Fix Release 0.20.3 date
-- Description: Corrects the release date from 2025 to 2026

UPDATE releases SET release_date = '2026-01-14' WHERE version = '0.20.3';
