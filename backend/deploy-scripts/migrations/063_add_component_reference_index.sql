-- Migration: Add component reference index to events table
-- Description: Creates index for querying events that reference a component ID
-- This enables showing fit score events in component audit history

CREATE INDEX IF NOT EXISTS idx_events_component_ref
ON events ((event_data->>'componentId'))
WHERE event_data->>'componentId' IS NOT NULL;
