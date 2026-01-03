-- Migration: Add actor columns to events table
-- Description: Adds actor tracking columns to support audit history
-- Backfills existing events with the first admin user per tenant

ALTER TABLE events ADD COLUMN IF NOT EXISTS actor_id VARCHAR(255);
ALTER TABLE events ADD COLUMN IF NOT EXISTS actor_email VARCHAR(500);

UPDATE events e
SET
    actor_id = first_admin.id::text,
    actor_email = first_admin.email
FROM (
    SELECT DISTINCT ON (tenant_id)
        tenant_id,
        id,
        email
    FROM users
    WHERE role = 'admin' AND status = 'active'
    ORDER BY tenant_id, created_at ASC
) first_admin
WHERE e.tenant_id = first_admin.tenant_id
  AND e.actor_id IS NULL;

ALTER TABLE events ALTER COLUMN actor_id SET NOT NULL;
ALTER TABLE events ALTER COLUMN actor_email SET NOT NULL;

CREATE INDEX IF NOT EXISTS idx_events_tenant_actor ON events(tenant_id, actor_id);
CREATE INDEX IF NOT EXISTS idx_events_tenant_occurred ON events(tenant_id, occurred_at);
