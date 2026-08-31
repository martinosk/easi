-- Spec 211: maturity journeys.
-- Description: adds the nullable target_maturity a maturity journey declares, and re-scopes the
--              single-active-journey index so a capability may run one application journey and one
--              maturity journey at the same time, but never two of either.
--              Current maturity is read from architecturedirection.capability_node_cache at query
--              time and the gap is derived there, so nothing is backfilled: journeys planned before
--              this migration have no target maturity and no gap.

ALTER TABLE architecturedirection.capability_journeys
    ADD COLUMN IF NOT EXISTS target_maturity SMALLINT;

DROP INDEX IF EXISTS architecturedirection.uq_capability_journeys_single_active;

CREATE UNIQUE INDEX IF NOT EXISTS uq_capability_journeys_single_active
    ON architecturedirection.capability_journeys(tenant_id, capability_id, (kind = 'maturity'))
    WHERE status IN ('planned', 'in-flight');
