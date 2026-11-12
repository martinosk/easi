-- Migration: 172_backfill_architectureviews_view_membership
-- Purpose: spec 234 — one-time repair of view membership. Appends the
-- ComponentRemovedFromView events that failed application cleanup should have
-- produced, makes every non-deleted view's application placements match the
-- applications its history lists, and drops placements of deleted capabilities,
-- origin entities and views. Cross-schema reads and event-store writes are
-- allowed because this is a backfill (spec 209, spec 234 Design Decision 6).
-- DML only. Idempotent.

WITH latest AS (
    SELECT DISTINCT ON (e.tenant_id, e.aggregate_id, e.event_data->>'componentId')
           e.tenant_id, e.aggregate_id AS view_id, e.event_data->>'componentId' AS component_id, e.event_type
    FROM infrastructure.events e
    JOIN architectureviews.architecture_views v
      ON v.tenant_id = e.tenant_id AND v.id = e.aggregate_id AND v.is_deleted IS NOT TRUE
    WHERE e.event_type IN ('ComponentAddedToView', 'ComponentRemovedFromView')
    ORDER BY e.tenant_id, e.aggregate_id, e.event_data->>'componentId', e.version DESC
),
orphaned AS (
    SELECT l.tenant_id, l.view_id, l.component_id,
           ROW_NUMBER() OVER (PARTITION BY l.tenant_id, l.view_id ORDER BY l.component_id) AS offset_in_view
    FROM latest l
    WHERE l.event_type = 'ComponentAddedToView'
      AND NOT EXISTS (
          SELECT 1 FROM architecturemodeling.application_components ac
          WHERE ac.tenant_id = l.tenant_id AND ac.id = l.component_id AND ac.is_deleted IS NOT TRUE
      )
),
head AS (
    SELECT e.tenant_id, e.aggregate_id AS view_id, MAX(e.version) AS version
    FROM infrastructure.events e
    WHERE (e.tenant_id, e.aggregate_id) IN (SELECT tenant_id, view_id FROM orphaned)
    GROUP BY e.tenant_id, e.aggregate_id
)
INSERT INTO infrastructure.events (tenant_id, aggregate_id, event_type, event_data, version, occurred_at, actor_id, actor_email)
SELECT o.tenant_id,
       o.view_id,
       'ComponentRemovedFromView',
       jsonb_build_object(
           'viewId', o.view_id,
           'componentId', o.component_id,
           'removedAt', to_char(NOW() AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS.US"Z"')
       ),
       h.version + o.offset_in_view,
       NOW() AT TIME ZONE 'UTC',
       'system',
       'system@easi.app'
FROM orphaned o
JOIN head h ON h.tenant_id = o.tenant_id AND h.view_id = o.view_id;

WITH latest AS (
    SELECT DISTINCT ON (e.tenant_id, e.aggregate_id, e.event_data->>'componentId')
           e.tenant_id, e.aggregate_id AS view_id, e.event_data->>'componentId' AS component_id, e.event_type
    FROM infrastructure.events e
    WHERE e.event_type IN ('ComponentAddedToView', 'ComponentRemovedFromView')
    ORDER BY e.tenant_id, e.aggregate_id, e.event_data->>'componentId', e.version DESC
)
DELETE FROM architectureviews.view_element_positions p
USING architectureviews.architecture_views v
WHERE v.tenant_id = p.tenant_id
  AND v.id = p.view_id
  AND v.is_deleted IS NOT TRUE
  AND p.element_type = 'component'
  AND NOT EXISTS (
      SELECT 1 FROM latest l
      WHERE l.tenant_id = p.tenant_id
        AND l.view_id = p.view_id
        AND l.component_id = p.element_id
        AND l.event_type = 'ComponentAddedToView'
  );

WITH latest AS (
    SELECT DISTINCT ON (e.tenant_id, e.aggregate_id, e.event_data->>'componentId')
           e.tenant_id, e.aggregate_id AS view_id, e.event_data->>'componentId' AS component_id, e.event_type,
           COALESCE((e.event_data->>'x')::DOUBLE PRECISION, 0) AS x,
           COALESCE((e.event_data->>'y')::DOUBLE PRECISION, 0) AS y
    FROM infrastructure.events e
    JOIN architectureviews.architecture_views v
      ON v.tenant_id = e.tenant_id AND v.id = e.aggregate_id AND v.is_deleted IS NOT TRUE
    WHERE e.event_type IN ('ComponentAddedToView', 'ComponentRemovedFromView')
    ORDER BY e.tenant_id, e.aggregate_id, e.event_data->>'componentId', e.version DESC
)
INSERT INTO architectureviews.view_element_positions (view_id, tenant_id, element_id, element_type, x, y, created_at, updated_at)
SELECT l.view_id, l.tenant_id, l.component_id, 'component', l.x, l.y, NOW(), NOW()
FROM latest l
WHERE l.event_type = 'ComponentAddedToView'
ON CONFLICT (tenant_id, view_id, element_id, element_type) DO NOTHING;

DELETE FROM architectureviews.view_element_positions p
WHERE p.element_type = 'capability'
  AND NOT EXISTS (
      SELECT 1 FROM capabilitymapping.capabilities c
      WHERE c.tenant_id = p.tenant_id AND c.id = p.element_id
  );

DELETE FROM architectureviews.view_element_positions p
WHERE p.element_type = 'origin_entity'
  AND NOT EXISTS (
      SELECT 1 FROM architecturemodeling.vendors o
      WHERE o.tenant_id = p.tenant_id AND o.id = p.element_id AND o.is_deleted IS NOT TRUE
  )
  AND NOT EXISTS (
      SELECT 1 FROM architecturemodeling.acquired_entities o
      WHERE o.tenant_id = p.tenant_id AND o.id = p.element_id AND o.is_deleted IS NOT TRUE
  )
  AND NOT EXISTS (
      SELECT 1 FROM architecturemodeling.internal_teams o
      WHERE o.tenant_id = p.tenant_id AND o.id = p.element_id AND o.is_deleted IS NOT TRUE
  );

DELETE FROM architectureviews.view_element_positions p
WHERE NOT EXISTS (
    SELECT 1 FROM architectureviews.architecture_views v
    WHERE v.tenant_id = p.tenant_id AND v.id = p.view_id AND v.is_deleted IS NOT TRUE
);
