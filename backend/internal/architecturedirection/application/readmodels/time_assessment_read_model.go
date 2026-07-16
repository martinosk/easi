package readmodels

import (
	"database/sql"
	"errors"
	"time"

	"context"

	"github.com/lib/pq"

	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/shared/types"
)

const timeAssessmentPerPairUniqueConstraint = "uq_time_assessments_per_pair"
const timeAssessmentStaleThreshold = "12 months"

var ErrTimeAssessmentAlreadyExists = errors.New("a time assessment already exists for this capability and component pair")

type TimeAssessmentDTO struct {
	ID             string      `json:"id"`
	CapabilityID   string      `json:"capabilityId"`
	CapabilityName string      `json:"capabilityName"`
	ComponentID    string      `json:"componentId"`
	ComponentName  string      `json:"componentName"`
	Grade          string      `json:"grade"`
	Rationale      string      `json:"rationale"`
	AssessedBy     string      `json:"assessedBy"`
	AssessedByName string      `json:"assessedByName"`
	AssessedAt     time.Time   `json:"assessedAt"`
	Stale          bool        `json:"stale"`
	Links          types.Links `json:"_links,omitempty"`
}

type TimeGradeCounts struct {
	Invest    int `json:"Invest"`
	Tolerate  int `json:"Tolerate"`
	Migrate   int `json:"Migrate"`
	Eliminate int `json:"Eliminate"`
}

type TimeAssessmentRollupDTO struct {
	ComponentID string          `json:"componentId"`
	Counts      TimeGradeCounts `json:"counts"`
}

type UpsertTimeAssessmentParams struct {
	ID            string
	CapabilityID  string
	ComponentID   string
	RealizationID string
	Grade         string
	Rationale     string
	AssessedBy    string
	AssessedAt    time.Time
}

type TimeAssessmentReadModel struct {
	db *database.TenantAwareDB
}

func NewTimeAssessmentReadModel(db *database.TenantAwareDB) *TimeAssessmentReadModel {
	return &TimeAssessmentReadModel{db: db}
}

func (rm *TimeAssessmentReadModel) UpsertCurrent(ctx context.Context, p UpsertTimeAssessmentParams) error {
	tenantID, err := tenantOf(ctx)
	if err != nil {
		return err
	}
	_, err = rm.db.ExecContext(ctx,
		`INSERT INTO architecturedirection.time_assessments
		 (id, tenant_id, capability_id, component_id, realization_id, grade, rationale, assessed_by, assessed_by_name, assessed_at, capability_name, component_name)
		 SELECT $1, $2, $3, $4, $5, $6, $7, $8, COALESCE(usr.name, $8), $9, cap.name, comp.name
		 FROM (SELECT 1) AS stub
		 LEFT JOIN architecturedirection.reference_name_cache cap
		   ON cap.tenant_id = $2 AND cap.entity_type = 'capability' AND cap.entity_id = $3
		 LEFT JOIN architecturedirection.reference_name_cache comp
		   ON comp.tenant_id = $2 AND comp.entity_type = 'application' AND comp.entity_id = $4
		 LEFT JOIN architecturedirection.reference_name_cache usr
		   ON usr.tenant_id = $2 AND usr.entity_type = 'user' AND usr.entity_id = $8
		 ON CONFLICT (tenant_id, id) DO UPDATE SET
		   capability_id = EXCLUDED.capability_id,
		   component_id = EXCLUDED.component_id,
		   realization_id = EXCLUDED.realization_id,
		   grade = EXCLUDED.grade,
		   rationale = EXCLUDED.rationale,
		   assessed_by = EXCLUDED.assessed_by,
		   assessed_by_name = EXCLUDED.assessed_by_name,
		   assessed_at = EXCLUDED.assessed_at,
		   capability_name = EXCLUDED.capability_name,
		   component_name = EXCLUDED.component_name,
		   updated_at = CURRENT_TIMESTAMP`,
		p.ID, tenantID, p.CapabilityID, p.ComponentID, p.RealizationID, p.Grade, p.Rationale, p.AssessedBy, p.AssessedAt,
	)
	return mapTimeAssessmentInsertError(err)
}

func mapTimeAssessmentInsertError(err error) error {
	if isTimeAssessmentPairUniqueViolation(err) {
		return ErrTimeAssessmentAlreadyExists
	}
	return err
}

func isTimeAssessmentPairUniqueViolation(err error) bool {
	var pqErr *pq.Error
	if !errors.As(err, &pqErr) {
		return false
	}
	return string(pqErr.Code) == pgUniqueViolation && pqErr.Constraint == timeAssessmentPerPairUniqueConstraint
}

func (rm *TimeAssessmentReadModel) Delete(ctx context.Context, id string) error {
	return rm.tenantExec(ctx,
		`DELETE FROM architecturedirection.time_assessments WHERE tenant_id = $1 AND id = $2`,
		func(t string) []any { return []any{t, id} },
	)
}

func (rm *TimeAssessmentReadModel) DeleteByCapabilityID(ctx context.Context, capabilityID string) error {
	return rm.tenantExec(ctx,
		`DELETE FROM architecturedirection.time_assessments WHERE tenant_id = $1 AND capability_id = $2`,
		func(t string) []any { return []any{t, capabilityID} },
	)
}

func (rm *TimeAssessmentReadModel) DeleteByComponentID(ctx context.Context, componentID string) error {
	return rm.tenantExec(ctx,
		`DELETE FROM architecturedirection.time_assessments WHERE tenant_id = $1 AND component_id = $2`,
		func(t string) []any { return []any{t, componentID} },
	)
}

func (rm *TimeAssessmentReadModel) CacheCapabilityName(ctx context.Context, capabilityID, name string) error {
	return rm.cacheReferenceName(ctx, "capability", capabilityID, name)
}

func (rm *TimeAssessmentReadModel) UpdateCapabilityName(ctx context.Context, capabilityID, name string) error {
	return rm.tenantExec(ctx,
		`UPDATE architecturedirection.time_assessments SET capability_name = $1
		 WHERE tenant_id = $2 AND capability_id = $3`,
		func(t string) []any { return []any{name, t, capabilityID} },
	)
}

func (rm *TimeAssessmentReadModel) CacheComponentName(ctx context.Context, componentID, name string) error {
	return rm.cacheReferenceName(ctx, "application", componentID, name)
}

func (rm *TimeAssessmentReadModel) CacheUserName(ctx context.Context, email, name string) error {
	return rm.cacheReferenceName(ctx, "user", email, name)
}

func (rm *TimeAssessmentReadModel) cacheReferenceName(ctx context.Context, entityType, entityID, name string) error {
	return rm.tenantExec(ctx,
		`INSERT INTO architecturedirection.reference_name_cache (tenant_id, entity_type, entity_id, name)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (tenant_id, entity_type, entity_id) DO UPDATE SET name = EXCLUDED.name`,
		func(t string) []any { return []any{t, entityType, entityID, name} },
	)
}

func (rm *TimeAssessmentReadModel) UpdateComponentName(ctx context.Context, componentID, name string) error {
	return rm.tenantExec(ctx,
		`UPDATE architecturedirection.time_assessments SET component_name = $1
		 WHERE tenant_id = $2 AND component_id = $3`,
		func(t string) []any { return []any{name, t, componentID} },
	)
}

func (rm *TimeAssessmentReadModel) tenantExec(ctx context.Context, query string, argsFn func(tenantID string) []any) error {
	tenantID, err := tenantOf(ctx)
	if err != nil {
		return err
	}
	_, err = rm.db.ExecContext(ctx, query, argsFn(tenantID)...)
	return err
}

func (rm *TimeAssessmentReadModel) FindPairByRealizationID(ctx context.Context, realizationID string) (string, string, bool, error) {
	var capabilityID, componentID string
	found, err := rm.findSingleRow(ctx,
		`SELECT capability_id, component_id FROM architecturedirection.time_assessments
		 WHERE tenant_id = $1 AND realization_id = $2`,
		[]any{realizationID}, &capabilityID, &componentID)
	return capabilityID, componentID, found, err
}

func (rm *TimeAssessmentReadModel) FindAggregateIDForPair(ctx context.Context, capabilityID, componentID string) (string, bool, error) {
	var id string
	found, err := rm.findSingleRow(ctx,
		`SELECT id FROM architecturedirection.time_assessments
		 WHERE tenant_id = $1 AND capability_id = $2 AND component_id = $3`,
		[]any{capabilityID, componentID}, &id)
	return id, found, err
}

func (rm *TimeAssessmentReadModel) findSingleRow(ctx context.Context, query string, args []any, dest ...any) (bool, error) {
	tenantID, err := tenantOf(ctx)
	if err != nil {
		return false, err
	}
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx, query, append([]any{tenantID}, args...)...).Scan(dest...)
	})
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

const timeAssessmentSelectColumns = `id, capability_id, COALESCE(capability_name, ''), component_id, COALESCE(component_name, ''),
	grade, rationale, assessed_by, assessed_by_name, assessed_at,
	(assessed_at < now() - interval '` + timeAssessmentStaleThreshold + `')`

func (rm *TimeAssessmentReadModel) GetByPair(ctx context.Context, capabilityID, componentID string) (*TimeAssessmentDTO, error) {
	tenantID, err := tenantOf(ctx)
	if err != nil {
		return nil, err
	}
	var dto *TimeAssessmentDTO
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		row := tx.QueryRowContext(ctx,
			`SELECT `+timeAssessmentSelectColumns+`
			 FROM architecturedirection.time_assessments
			 WHERE tenant_id = $1 AND capability_id = $2 AND component_id = $3`,
			tenantID, capabilityID, componentID,
		)
		fetched, scanErr := scanTimeAssessment(row)
		if errors.Is(scanErr, sql.ErrNoRows) {
			return nil
		}
		if scanErr != nil {
			return scanErr
		}
		dto = &fetched
		return nil
	})
	return dto, err
}

func (rm *TimeAssessmentReadModel) GetByCapabilityIDs(ctx context.Context, capabilityIDs []string) ([]TimeAssessmentDTO, error) {
	if len(capabilityIDs) == 0 {
		return []TimeAssessmentDTO{}, nil
	}
	tenantID, err := tenantOf(ctx)
	if err != nil {
		return nil, err
	}
	return rm.assessments(ctx, "tenant_id = $1 AND capability_id = ANY($2)", tenantID, pq.Array(capabilityIDs))
}

func (rm *TimeAssessmentReadModel) GetAll(ctx context.Context) ([]TimeAssessmentDTO, error) {
	tenantID, err := tenantOf(ctx)
	if err != nil {
		return nil, err
	}
	return rm.assessments(ctx, "tenant_id = $1", tenantID)
}

func (rm *TimeAssessmentReadModel) assessments(ctx context.Context, where string, args ...any) ([]TimeAssessmentDTO, error) {
	assessments := []TimeAssessmentDTO{}
	err := rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx,
			`SELECT `+timeAssessmentSelectColumns+`
			 FROM architecturedirection.time_assessments
			 WHERE `+where+`
			 ORDER BY assessed_at DESC`,
			args...,
		)
		if err != nil {
			return err
		}
		defer func() { _ = rows.Close() }()

		for rows.Next() {
			dto, scanErr := scanTimeAssessment(rows)
			if scanErr != nil {
				return scanErr
			}
			assessments = append(assessments, dto)
		}
		return rows.Err()
	})
	return assessments, err
}

func (rm *TimeAssessmentReadModel) GetRollupsByComponentIDs(ctx context.Context, componentIDs []string) ([]TimeAssessmentRollupDTO, error) {
	if len(componentIDs) == 0 {
		return []TimeAssessmentRollupDTO{}, nil
	}
	tenantID, err := tenantOf(ctx)
	if err != nil {
		return nil, err
	}
	var rollups []TimeAssessmentRollupDTO
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		grouped, groupErr := groupTimeAssessmentCountsByComponent(ctx, tx, tenantID, componentIDs)
		if groupErr != nil {
			return groupErr
		}
		rollups = grouped
		return nil
	})
	if err != nil {
		return nil, err
	}
	return rollups, nil
}

func groupTimeAssessmentCountsByComponent(ctx context.Context, tx *sql.Tx, tenantID string, componentIDs []string) ([]TimeAssessmentRollupDTO, error) {
	rows, err := tx.QueryContext(ctx,
		`SELECT component_id, grade, COUNT(*)
		 FROM architecturedirection.time_assessments
		 WHERE tenant_id = $1 AND component_id = ANY($2)
		 GROUP BY component_id, grade`,
		tenantID, pq.Array(componentIDs),
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	byComponent := map[string]*TimeAssessmentRollupDTO{}
	order := []string{}
	for rows.Next() {
		componentID, grade, count, scanErr := scanTimeAssessmentGradeCount(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		entry, exists := byComponent[componentID]
		if !exists {
			entry = &TimeAssessmentRollupDTO{ComponentID: componentID}
			byComponent[componentID] = entry
			order = append(order, componentID)
		}
		applyGradeCount(&entry.Counts, grade, count)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return rollupsInOrder(byComponent, order), nil
}

func scanTimeAssessmentGradeCount(rows *sql.Rows) (string, string, int, error) {
	var componentID, grade string
	var count int
	err := rows.Scan(&componentID, &grade, &count)
	return componentID, grade, count, err
}

func rollupsInOrder(byComponent map[string]*TimeAssessmentRollupDTO, order []string) []TimeAssessmentRollupDTO {
	rollups := make([]TimeAssessmentRollupDTO, 0, len(order))
	for _, componentID := range order {
		rollups = append(rollups, *byComponent[componentID])
	}
	return rollups
}

func applyGradeCount(counts *TimeGradeCounts, grade string, count int) {
	switch grade {
	case "Invest":
		counts.Invest = count
	case "Tolerate":
		counts.Tolerate = count
	case "Migrate":
		counts.Migrate = count
	case "Eliminate":
		counts.Eliminate = count
	}
}

type timeAssessmentRowScanner interface {
	Scan(dest ...any) error
}

func scanTimeAssessment(row timeAssessmentRowScanner) (TimeAssessmentDTO, error) {
	var dto TimeAssessmentDTO
	err := row.Scan(&dto.ID, &dto.CapabilityID, &dto.CapabilityName, &dto.ComponentID, &dto.ComponentName,
		&dto.Grade, &dto.Rationale, &dto.AssessedBy, &dto.AssessedByName, &dto.AssessedAt, &dto.Stale)
	return dto, err
}
