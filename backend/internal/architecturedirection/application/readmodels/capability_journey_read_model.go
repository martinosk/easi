package readmodels

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"

	"easi/backend/internal/infrastructure/database"
)

const activeJourneyUniqueConstraint = "uq_capability_journeys_single_active"

var ErrActiveCapabilityJourneyExists = errors.New("an active journey already exists for this capability")
var ErrUnknownJourneyTimestampColumn = errors.New("unknown journey timestamp column")

type CapabilityJourneyReadModel struct {
	db *database.TenantAwareDB
}

func NewCapabilityJourneyReadModel(db *database.TenantAwareDB) *CapabilityJourneyReadModel {
	return &CapabilityJourneyReadModel{db: db}
}

func (rm *CapabilityJourneyReadModel) InsertJourney(ctx context.Context, p InsertJourneyParams) error {
	return rm.withTx(ctx, func(tx *sql.Tx, tenantID string) error {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO architecturedirection.capability_journeys
			 (tenant_id, id, capability_id, kind, status, target_year, target_quarter, note,
			  planned_by, planned_by_name, planned_at, capability_name, to_component_id, to_component_name,
			  target_domain_id, target_domain_name, target_parent_id, target_parent_name, resulting_name, target_maturity)
			 SELECT $1, $2, $3, $4, 'planned', $5, $6, $7, $8, COALESCE(usr.name, ''), $9,
			   COALESCE(cap.name, ''), $10, COALESCE(comp.name, ''),
			   NULLIF($11, ''), COALESCE(dom.name, ''), NULLIF($12, ''), COALESCE(parent.name, ''), $13, $14
			 FROM (SELECT 1) AS stub
			 LEFT JOIN architecturedirection.reference_name_cache cap
			   ON cap.tenant_id = $1 AND cap.entity_type = 'capability' AND cap.entity_id = $3
			 LEFT JOIN architecturedirection.reference_name_cache comp
			   ON comp.tenant_id = $1 AND comp.entity_type = 'application' AND comp.entity_id = $10
			 LEFT JOIN architecturedirection.reference_name_cache dom
			   ON dom.tenant_id = $1 AND dom.entity_type = 'business_domain' AND dom.entity_id = NULLIF($11, '')
			 LEFT JOIN architecturedirection.reference_name_cache parent
			   ON parent.tenant_id = $1 AND parent.entity_type = 'capability' AND parent.entity_id = NULLIF($12, '')
			 LEFT JOIN architecturedirection.reference_name_cache usr
			   ON usr.tenant_id = $1 AND usr.entity_type = 'user' AND usr.entity_id = $8`,
			tenantID, p.ID, p.CapabilityID, p.Kind, nullableInt(p.TargetYear), nullableInt(p.TargetQuarter), p.Note,
			p.PlannedBy, p.PlannedAt, p.ToComponentID, p.TargetDomainID, p.TargetParentID, p.ResultingName,
			nullableInt(p.TargetMaturity),
		); err != nil {
			return mapJourneyInsertError(err)
		}
		return journeySourceWriter{tx: tx, tenantID: tenantID}.replace(ctx, p.ID, p.FromComponentIDs)
	})
}

func isActiveJourneyUniqueViolation(err error) bool {
	var pqErr *pq.Error
	if !errors.As(err, &pqErr) {
		return false
	}
	return string(pqErr.Code) == pgUniqueViolation && pqErr.Constraint == activeJourneyUniqueConstraint
}

func mapJourneyInsertError(err error) error {
	if isActiveJourneyUniqueViolation(err) {
		return ErrActiveCapabilityJourneyExists
	}
	return err
}

func (rm *CapabilityJourneyReadModel) UpdateStatus(ctx context.Context, p UpdateJourneyStatusParams) error {
	if _, ok := journeyTimestampColumns[p.Column]; !ok {
		return ErrUnknownJourneyTimestampColumn
	}
	return rm.tenantExec(ctx,
		`UPDATE architecturedirection.capability_journeys SET status = $1, `+string(p.Column)+` = $2, updated_at = CURRENT_TIMESTAMP
		 WHERE tenant_id = $3 AND id = $4`,
		func(t string) []any { return []any{p.Status, p.OccurredAt, t, p.JourneyID} },
	)
}

func (rm *CapabilityJourneyReadModel) UpdateProgress(ctx context.Context, journeyID string, progress int) error {
	return rm.tenantExec(ctx,
		`UPDATE architecturedirection.capability_journeys SET progress = $1, updated_at = CURRENT_TIMESTAMP
		 WHERE tenant_id = $2 AND id = $3`,
		func(t string) []any { return []any{progress, t, journeyID} },
	)
}

func (rm *CapabilityJourneyReadModel) UpdateDetails(ctx context.Context, p UpdateJourneyDetailsParams) error {
	return rm.tenantExec(ctx,
		`UPDATE architecturedirection.capability_journeys
		 SET note = $1, target_year = $2, target_quarter = $3, resulting_name = $4, updated_at = CURRENT_TIMESTAMP
		 WHERE tenant_id = $5 AND id = $6`,
		func(t string) []any {
			return []any{p.Note, nullableInt(p.TargetYear), nullableInt(p.TargetQuarter), p.ResultingName, t, p.JourneyID}
		},
	)
}

func (rm *CapabilityJourneyReadModel) ReplaceSources(ctx context.Context, journeyID string, componentIDs []string) error {
	return rm.withTx(ctx, func(tx *sql.Tx, tenantID string) error {
		if err := (journeySourceWriter{tx: tx, tenantID: tenantID}).replace(ctx, journeyID, componentIDs); err != nil {
			return err
		}
		_, err := tx.ExecContext(ctx,
			`UPDATE architecturedirection.capability_journeys SET updated_at = CURRENT_TIMESTAMP WHERE tenant_id = $1 AND id = $2`,
			tenantID, journeyID,
		)
		return err
	})
}

type journeySourceWriter struct {
	tx       *sql.Tx
	tenantID string
}

func (w journeySourceWriter) replace(ctx context.Context, journeyID string, componentIDs []string) error {
	if _, err := w.tx.ExecContext(ctx,
		`DELETE FROM architecturedirection.capability_journey_sources WHERE tenant_id = $1 AND journey_id = $2`,
		w.tenantID, journeyID,
	); err != nil {
		return err
	}
	for position, componentID := range componentIDs {
		if err := w.insertOne(ctx, journeyID, componentID, position); err != nil {
			return err
		}
	}
	return nil
}

func (w journeySourceWriter) insertOne(ctx context.Context, journeyID, componentID string, position int) error {
	_, err := w.tx.ExecContext(ctx,
		`INSERT INTO architecturedirection.capability_journey_sources
		 (tenant_id, journey_id, component_id, position, component_name)
		 SELECT $1, $2, $3, $4, COALESCE(comp.name, '')
		 FROM (SELECT 1) AS stub
		 LEFT JOIN architecturedirection.reference_name_cache comp
		   ON comp.tenant_id = $1 AND comp.entity_type = 'application' AND comp.entity_id = $3`,
		w.tenantID, journeyID, componentID, position,
	)
	return err
}

func (rm *CapabilityJourneyReadModel) AddMilestone(ctx context.Context, p JourneyMilestoneUpsertParams) error {
	return rm.withTx(ctx, func(tx *sql.Tx, tenantID string) error {
		var nextPosition int
		if err := tx.QueryRowContext(ctx,
			`SELECT COALESCE(MAX(position) + 1, 0) FROM architecturedirection.capability_journey_milestones
			 WHERE tenant_id = $1 AND journey_id = $2`,
			tenantID, p.JourneyID,
		).Scan(&nextPosition); err != nil {
			return err
		}
		_, err := tx.ExecContext(ctx,
			`INSERT INTO architecturedirection.capability_journey_milestones
			 (tenant_id, journey_id, milestone_id, position, label, target_year, target_quarter, status, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			tenantID, p.JourneyID, p.MilestoneID, nextPosition, p.Label, nullableInt(p.TargetYear), nullableInt(p.TargetQuarter), p.Status, p.UpdatedAt,
		)
		return err
	})
}

func (rm *CapabilityJourneyReadModel) UpdateMilestone(ctx context.Context, p JourneyMilestoneUpsertParams) error {
	return rm.tenantExec(ctx,
		`UPDATE architecturedirection.capability_journey_milestones
		 SET label = $1, target_year = $2, target_quarter = $3, status = $4, updated_at = $5
		 WHERE tenant_id = $6 AND journey_id = $7 AND milestone_id = $8`,
		func(t string) []any {
			return []any{p.Label, nullableInt(p.TargetYear), nullableInt(p.TargetQuarter), p.Status, p.UpdatedAt, t, p.JourneyID, p.MilestoneID}
		},
	)
}

func (rm *CapabilityJourneyReadModel) RemoveMilestone(ctx context.Context, journeyID, milestoneID string) error {
	return rm.withTx(ctx, func(tx *sql.Tx, tenantID string) error {
		var removedPosition int
		err := tx.QueryRowContext(ctx,
			`DELETE FROM architecturedirection.capability_journey_milestones
			 WHERE tenant_id = $1 AND journey_id = $2 AND milestone_id = $3
			 RETURNING position`,
			tenantID, journeyID, milestoneID,
		).Scan(&removedPosition)
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx,
			`UPDATE architecturedirection.capability_journey_milestones
			 SET position = position - 1
			 WHERE tenant_id = $1 AND journey_id = $2 AND position > $3`,
			tenantID, journeyID, removedPosition,
		)
		return err
	})
}

func (rm *CapabilityJourneyReadModel) ReorderMilestones(ctx context.Context, journeyID string, milestoneIDs []string) error {
	return rm.withTx(ctx, func(tx *sql.Tx, tenantID string) error {
		for position, milestoneID := range milestoneIDs {
			if _, err := tx.ExecContext(ctx,
				`UPDATE architecturedirection.capability_journey_milestones
				 SET position = $1
				 WHERE tenant_id = $2 AND journey_id = $3 AND milestone_id = $4`,
				position, tenantID, journeyID, milestoneID,
			); err != nil {
				return fmt.Errorf("reorder milestone %s on journey %s: %w", milestoneID, journeyID, err)
			}
		}
		return nil
	})
}

func (rm *CapabilityJourneyReadModel) tenantExec(ctx context.Context, query string, argsFn func(tenantID string) []any) error {
	tenantID, err := tenantOf(ctx)
	if err != nil {
		return err
	}
	_, err = rm.db.ExecContext(ctx, query, argsFn(tenantID)...)
	return err
}

func (rm *CapabilityJourneyReadModel) tenantTx(ctx context.Context, fn func(tx *sql.Tx, tenantID string) error) error {
	return rm.withTx(ctx, fn)
}

func (rm *CapabilityJourneyReadModel) withTx(ctx context.Context, fn func(tx *sql.Tx, tenantID string) error) error {
	tenantID, err := tenantOf(ctx)
	if err != nil {
		return err
	}
	tx, err := rm.db.BeginTxWithTenant(ctx, nil)
	if err != nil {
		return err
	}
	if err := fn(tx, tenantID); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func nullableInt(v *int) any {
	if v == nil {
		return nil
	}
	return *v
}
