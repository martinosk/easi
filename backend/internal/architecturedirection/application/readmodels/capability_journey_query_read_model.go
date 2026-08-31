package readmodels

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"

	"easi/backend/internal/architecturedirection/domain/valueobjects"
)

func (rm *CapabilityJourneyReadModel) FindActiveJourneyIDForCapability(ctx context.Context, capabilityID string, kinds []string) (string, bool, error) {
	tenantID, err := tenantOf(ctx)
	if err != nil {
		return "", false, err
	}
	var id string
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx,
			`SELECT id FROM architecturedirection.capability_journeys
			 WHERE tenant_id = $1 AND capability_id = $2 AND status IN ('planned', 'in-flight')
			   AND kind = ANY($3)`,
			tenantID, capabilityID, pq.Array(kinds),
		).Scan(&id)
	})
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return id, true, nil
}

func (rm *CapabilityJourneyReadModel) GetActiveByCapabilityID(ctx context.Context, capabilityID string) ([]CapabilityJourneyDTO, error) {
	return rm.fetchMany(ctx,
		`SELECT `+journeyColumns+` FROM `+journeySource+`
		 WHERE cj.tenant_id = $1 AND cj.capability_id = $2 AND cj.status IN ('planned', 'in-flight')
		 ORDER BY cj.planned_at DESC`,
		capabilityID,
	)
}

func (rm *CapabilityJourneyReadModel) GetByID(ctx context.Context, journeyID string) (*CapabilityJourneyDTO, error) {
	return rm.fetchOne(ctx,
		`SELECT `+journeyColumns+` FROM `+journeySource+` WHERE cj.tenant_id = $1 AND cj.id = $2`,
		func(tenantID string) []any { return []any{tenantID, journeyID} },
	)
}

func (rm *CapabilityJourneyReadModel) GetHistoryByCapabilityID(ctx context.Context, capabilityID string) ([]CapabilityJourneyDTO, error) {
	return rm.fetchMany(ctx,
		`SELECT `+journeyColumns+` FROM `+journeySource+`
		 WHERE cj.tenant_id = $1 AND cj.capability_id = $2 ORDER BY cj.planned_at DESC`,
		capabilityID,
	)
}

func (rm *CapabilityJourneyReadModel) fetchMany(ctx context.Context, query, capabilityID string) ([]CapabilityJourneyDTO, error) {
	tenantID, err := tenantOf(ctx)
	if err != nil {
		return nil, err
	}
	journeys := []CapabilityJourneyDTO{}
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		loaded, err := rm.queryJourneys(ctx, tx, query, tenantID, capabilityID)
		journeys = loaded
		return err
	})
	return journeys, err
}

func (rm *CapabilityJourneyReadModel) GetCurrentByCapabilityIDs(ctx context.Context, capabilityIDs []string) ([]CapabilityJourneyDTO, error) {
	if len(capabilityIDs) == 0 {
		return []CapabilityJourneyDTO{}, nil
	}
	tenantID, err := tenantOf(ctx)
	if err != nil {
		return nil, err
	}
	return rm.currentJourneys(ctx, "tenant_id = $1 AND capability_id = ANY($2)", tenantID, pq.Array(capabilityIDs))
}

func (rm *CapabilityJourneyReadModel) GetAllCurrent(ctx context.Context) ([]CapabilityJourneyDTO, error) {
	tenantID, err := tenantOf(ctx)
	if err != nil {
		return nil, err
	}
	return rm.currentJourneys(ctx, "tenant_id = $1", tenantID)
}

func (rm *CapabilityJourneyReadModel) currentJourneys(ctx context.Context, where string, args ...any) ([]CapabilityJourneyDTO, error) {
	journeys := []CapabilityJourneyDTO{}
	err := rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		loaded, err := rm.queryJourneys(ctx, tx,
			`SELECT `+journeyColumns+` FROM (
			   SELECT ranked.*, ROW_NUMBER() OVER (
			     PARTITION BY capability_id,
			       (CASE WHEN status IN ('planned', 'in-flight') THEN 1 ELSE 0 END),
			       (CASE WHEN status IN ('planned', 'in-flight') AND kind = '`+valueobjects.JourneyKindMaturity+`' THEN 1 ELSE 0 END)
			     ORDER BY planned_at DESC
			   ) AS rn
			   FROM architecturedirection.capability_journeys ranked
			   WHERE `+where+`
			 ) cj `+journeyMaturityJoin+`
			 WHERE cj.rn = 1
			 ORDER BY cj.capability_id, cj.planned_at DESC`,
			args...,
		)
		journeys = loaded
		return err
	})
	return journeys, err
}

func (rm *CapabilityJourneyReadModel) queryJourneys(ctx context.Context, tx *sql.Tx, query string, args ...any) ([]CapabilityJourneyDTO, error) {
	tenantID, err := tenantOf(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	journeys, err := scanAllJourneys(rows)
	if err != nil {
		return nil, err
	}
	for i := range journeys {
		if err := rm.hydrateJourney(ctx, tx, tenantID, &journeys[i]); err != nil {
			return nil, err
		}
	}
	return journeys, nil
}

func scanAllJourneys(rows *sql.Rows) ([]CapabilityJourneyDTO, error) {
	defer func() { _ = rows.Close() }()
	journeys := []CapabilityJourneyDTO{}
	for rows.Next() {
		dto, err := scanJourney(rows)
		if err != nil {
			return nil, err
		}
		journeys = append(journeys, dto)
	}
	return journeys, rows.Err()
}

const journeyColumns = `cj.id, cj.capability_id, COALESCE(cj.capability_name, ''), cj.capability_stale, cj.kind, cj.status,
	cj.progress, cj.target_year, cj.target_quarter, cj.note,
	cj.to_component_id, COALESCE(cj.to_component_name, ''), cj.to_component_stale,
	COALESCE(cj.target_domain_id, ''), COALESCE(cj.target_domain_name, ''), cj.target_domain_stale,
	COALESCE(cj.target_parent_id, ''), COALESCE(cj.target_parent_name, ''), cj.target_parent_stale, cj.resulting_name,
	cj.target_maturity, cnc.maturity_value,
	cj.planned_by, cj.planned_by_name, cj.planned_at, cj.updated_at, cj.started_at, cj.completed_at, cj.abandoned_at`

const journeyMaturityJoin = `LEFT JOIN architecturedirection.capability_node_cache cnc
	 ON cnc.tenant_id = cj.tenant_id AND cnc.capability_id = cj.capability_id`

const journeySource = `architecturedirection.capability_journeys cj ` + journeyMaturityJoin

type journeyRowScanner interface {
	Scan(dest ...any) error
}

type journeyScanBuffer struct {
	progress                                       sql.NullInt64
	targetYear, targetQuarter                      sql.NullInt64
	move                                           JourneyMoveDTO
	targetMaturity, currentMaturity                sql.NullInt64
	updatedAt, startedAt, completedAt, abandonedAt sql.NullTime
}

func scanJourney(row journeyRowScanner) (CapabilityJourneyDTO, error) {
	var dto CapabilityJourneyDTO
	var raw journeyScanBuffer
	if err := row.Scan(journeyScanTargets(&dto, &raw)...); err != nil {
		return dto, err
	}
	applyJourneyScanBuffer(&dto, raw)
	dto.FromApplications = []JourneyApplicationRefDTO{}
	dto.Milestones = []CapabilityJourneyMilestoneDTO{}
	return dto, nil
}

func journeyScanTargets(dto *CapabilityJourneyDTO, raw *journeyScanBuffer) []any {
	return []any{
		&dto.ID, &dto.CapabilityID, &dto.CapabilityName, &dto.CapabilityStale, &dto.Kind, &dto.Status,
		&raw.progress, &raw.targetYear, &raw.targetQuarter, &dto.Note,
		&dto.ToApplication.ComponentID, &dto.ToApplication.ComponentName, &dto.ToApplication.Stale,
		&raw.move.TargetDomainID, &raw.move.TargetDomainName, &raw.move.TargetDomainStale,
		&raw.move.TargetParentID, &raw.move.TargetParentName, &raw.move.TargetParentStale, &raw.move.ResultingName,
		&raw.targetMaturity, &raw.currentMaturity,
		&dto.PlannedBy, &dto.PlannedByName, &dto.PlannedAt, &raw.updatedAt, &raw.startedAt, &raw.completedAt, &raw.abandonedAt,
	}
}

func applyJourneyScanBuffer(dto *CapabilityJourneyDTO, raw journeyScanBuffer) {
	applyJourneyProgress(dto, raw.progress)
	applyJourneyTargetPeriod(dto, raw.targetYear, raw.targetQuarter)
	applyJourneyMove(dto, raw.move)
	applyJourneyMaturity(dto, raw.targetMaturity, raw.currentMaturity)
	applyJourneyTimestamps(dto, raw)
}

func applyJourneyProgress(dto *CapabilityJourneyDTO, progress sql.NullInt64) {
	if !progress.Valid {
		return
	}
	v := int(progress.Int64)
	dto.Progress = &v
}

func applyJourneyTargetPeriod(dto *CapabilityJourneyDTO, year, quarter sql.NullInt64) {
	if !year.Valid || !quarter.Valid {
		return
	}
	dto.TargetPeriod = &TargetPeriodDTO{Year: int(year.Int64), Quarter: int(quarter.Int64)}
}

func applyJourneyMove(dto *CapabilityJourneyDTO, move JourneyMoveDTO) {
	if dto.Kind != valueobjects.JourneyKindMove {
		return
	}
	dto.Move = &move
}

func applyJourneyMaturity(dto *CapabilityJourneyDTO, target, current sql.NullInt64) {
	if dto.Kind != valueobjects.JourneyKindMaturity || !target.Valid {
		return
	}
	targetValue, currentValue := int(target.Int64), int(current.Int64)
	dto.Maturity = &JourneyMaturityDTO{
		TargetMaturity:  targetValue,
		CurrentMaturity: currentValue,
		MaturityGap:     targetValue - currentValue,
	}
}

func applyJourneyTimestamps(dto *CapabilityJourneyDTO, raw journeyScanBuffer) {
	if raw.updatedAt.Valid {
		dto.UpdatedAt = &raw.updatedAt.Time
	}
	if raw.startedAt.Valid {
		dto.StartedAt = &raw.startedAt.Time
	}
	if raw.completedAt.Valid {
		dto.CompletedAt = &raw.completedAt.Time
	}
	if raw.abandonedAt.Valid {
		dto.AbandonedAt = &raw.abandonedAt.Time
	}
}

func (rm *CapabilityJourneyReadModel) fetchOne(ctx context.Context, query string, argsFn func(tenantID string) []any) (*CapabilityJourneyDTO, error) {
	tenantID, err := tenantOf(ctx)
	if err != nil {
		return nil, err
	}
	var dto *CapabilityJourneyDTO
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		row := tx.QueryRowContext(ctx, query, argsFn(tenantID)...)
		fetched, scanErr := scanJourney(row)
		if errors.Is(scanErr, sql.ErrNoRows) {
			return nil
		}
		if scanErr != nil {
			return scanErr
		}
		if err := rm.hydrateJourney(ctx, tx, tenantID, &fetched); err != nil {
			return err
		}
		dto = &fetched
		return nil
	})
	return dto, err
}

func (rm *CapabilityJourneyReadModel) hydrateJourney(ctx context.Context, tx *sql.Tx, tenantID string, dto *CapabilityJourneyDTO) error {
	sources, err := loadJourneySources(ctx, tx, tenantID, dto.ID)
	if err != nil {
		return err
	}
	dto.FromApplications = sources
	milestones, err := loadJourneyMilestones(ctx, tx, tenantID, dto.ID)
	if err != nil {
		return err
	}
	dto.Milestones = milestones
	return nil
}

func loadJourneySources(ctx context.Context, tx *sql.Tx, tenantID, journeyID string) ([]JourneyApplicationRefDTO, error) {
	rows, err := tx.QueryContext(ctx,
		`SELECT component_id, COALESCE(component_name, ''), component_stale
		 FROM architecturedirection.capability_journey_sources
		 WHERE tenant_id = $1 AND journey_id = $2 ORDER BY position`,
		tenantID, journeyID,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := []JourneyApplicationRefDTO{}
	for rows.Next() {
		var ref JourneyApplicationRefDTO
		if err := rows.Scan(&ref.ComponentID, &ref.ComponentName, &ref.Stale); err != nil {
			return nil, err
		}
		out = append(out, ref)
	}
	return out, rows.Err()
}

func loadJourneyMilestones(ctx context.Context, tx *sql.Tx, tenantID, journeyID string) ([]CapabilityJourneyMilestoneDTO, error) {
	rows, err := tx.QueryContext(ctx,
		`SELECT milestone_id, label, target_year, target_quarter, status
		 FROM architecturedirection.capability_journey_milestones
		 WHERE tenant_id = $1 AND journey_id = $2 ORDER BY position`,
		tenantID, journeyID,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := []CapabilityJourneyMilestoneDTO{}
	for rows.Next() {
		var m CapabilityJourneyMilestoneDTO
		var year, quarter sql.NullInt64
		if err := rows.Scan(&m.ID, &m.Label, &year, &quarter, &m.Status); err != nil {
			return nil, err
		}
		if year.Valid && quarter.Valid {
			m.TargetPeriod = &TargetPeriodDTO{Year: int(year.Int64), Quarter: int(quarter.Int64)}
		}
		out = append(out, m)
	}
	return out, rows.Err()
}
