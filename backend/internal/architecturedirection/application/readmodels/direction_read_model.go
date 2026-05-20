package readmodels

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/lib/pq"

	"easi/backend/internal/architecturedirection/domain/services"
	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/types"
)

const pgUniqueViolation = "23505"
const activeDirectionUniqueConstraint = "uq_directions_active_per_ec"

type DirectionID string
type CapabilityID string

type DirectionField string

const (
	DirectionFieldStatus    DirectionField = "status"
	DirectionFieldNarrative DirectionField = "narrative"
	DirectionFieldHorizon   DirectionField = "horizon"
)

type DirectionPlacementDTO struct {
	TargetBusinessDomainID string `json:"targetBusinessDomainId"`
	ResultingName          string `json:"resultingName,omitempty"`
}

type DirectionSourceCapabilityDTO struct {
	ID    string `json:"id"`
	Stale bool   `json:"stale"`
}

type DirectionDTO struct {
	ID                     string                         `json:"id"`
	EnterpriseCapabilityID string                         `json:"enterpriseCapabilityId"`
	Type                   string                         `json:"type"`
	Status                 string                         `json:"status"`
	Horizon                string                         `json:"horizon"`
	Narrative              string                         `json:"narrative,omitempty"`
	SourceCapabilities     []DirectionSourceCapabilityDTO `json:"sourceCapabilities"`
	Placements             []DirectionPlacementDTO        `json:"placements"`
	CreatedAt              time.Time                      `json:"createdAt"`
	UpdatedAt              *time.Time                     `json:"updatedAt,omitempty"`
	Links                  types.Links                    `json:"_links,omitempty"`
}

type DirectionReadModel struct {
	db *database.TenantAwareDB
}

func NewDirectionReadModel(db *database.TenantAwareDB) *DirectionReadModel {
	return &DirectionReadModel{db: db}
}

type InsertDirectionParams struct {
	ID                     DirectionID
	EnterpriseCapabilityID string
	Type                   string
	Status                 string
	Horizon                string
	Narrative              string
	SourceCapabilityIDs    []CapabilityID
	Placements             []DirectionPlacementDTO
	CreatedAt              time.Time
}

type FieldUpdate struct {
	DirectionID DirectionID
	Field       DirectionField
	Value       string
}

type PlacementsUpdate struct {
	DirectionID DirectionID
	Placements  []DirectionPlacementDTO
}

type SourceCapabilitiesUpdate struct {
	DirectionID         DirectionID
	SourceCapabilityIDs []CapabilityID
}

func (rm *DirectionReadModel) Insert(ctx context.Context, p InsertDirectionParams) error {
	placementsJSON, err := json.Marshal(p.Placements)
	if err != nil {
		return err
	}
	return rm.withTx(ctx, func(tx *sql.Tx, tenantID string) error {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO architecturedirection.directions
			 (id, tenant_id, enterprise_capability_id, type, status, horizon, narrative, placements, created_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8::jsonb, $9)
			 ON CONFLICT (tenant_id, id) DO UPDATE SET
			   enterprise_capability_id = EXCLUDED.enterprise_capability_id,
			   type = EXCLUDED.type,
			   status = EXCLUDED.status,
			   horizon = EXCLUDED.horizon,
			   narrative = EXCLUDED.narrative,
			   placements = EXCLUDED.placements,
			   created_at = EXCLUDED.created_at`,
			string(p.ID), tenantID, p.EnterpriseCapabilityID, p.Type, p.Status, p.Horizon, p.Narrative, string(placementsJSON), p.CreatedAt,
		); err != nil {
			return mapInsertError(err)
		}
		return sourceRowReplacement{tx: tx, tenantID: tenantID, directionID: p.ID, sourceCapabilityIDs: p.SourceCapabilityIDs}.execute(ctx)
	})
}

func mapInsertError(err error) error {
	if isActiveDirectionUniqueViolation(err) {
		return services.ErrActiveDirectionAlreadyExists
	}
	return err
}

func isActiveDirectionUniqueViolation(err error) bool {
	var pqErr *pq.Error
	if !errors.As(err, &pqErr) {
		return false
	}
	return string(pqErr.Code) == pgUniqueViolation && pqErr.Constraint == activeDirectionUniqueConstraint
}

func (rm *DirectionReadModel) UpdateField(ctx context.Context, u FieldUpdate) error {
	tenantID, err := tenantOf(ctx)
	if err != nil {
		return err
	}
	_, err = rm.db.ExecContext(ctx,
		`UPDATE architecturedirection.directions SET `+string(u.Field)+` = $1, updated_at = CURRENT_TIMESTAMP
		 WHERE tenant_id = $2 AND id = $3`,
		u.Value, tenantID, string(u.DirectionID),
	)
	return err
}

func (rm *DirectionReadModel) UpdatePlacements(ctx context.Context, u PlacementsUpdate) error {
	placementsJSON, err := json.Marshal(u.Placements)
	if err != nil {
		return err
	}
	tenantID, err := tenantOf(ctx)
	if err != nil {
		return err
	}
	_, err = rm.db.ExecContext(ctx,
		`UPDATE architecturedirection.directions SET placements = $1::jsonb, updated_at = CURRENT_TIMESTAMP
		 WHERE tenant_id = $2 AND id = $3`,
		string(placementsJSON), tenantID, string(u.DirectionID),
	)
	return err
}

func (rm *DirectionReadModel) ReplaceSourceCapabilities(ctx context.Context, u SourceCapabilitiesUpdate) error {
	return rm.withTx(ctx, func(tx *sql.Tx, tenantID string) error {
		replacement := sourceRowReplacement{tx: tx, tenantID: tenantID, directionID: u.DirectionID, sourceCapabilityIDs: u.SourceCapabilityIDs}
		if err := replacement.execute(ctx); err != nil {
			return err
		}
		_, err := tx.ExecContext(ctx,
			`UPDATE architecturedirection.directions SET updated_at = CURRENT_TIMESTAMP WHERE tenant_id = $1 AND id = $2`,
			tenantID, string(u.DirectionID),
		)
		return err
	})
}

func (rm *DirectionReadModel) MarkSourceCapabilityStale(ctx context.Context, capabilityID CapabilityID) error {
	tenantID, err := tenantOf(ctx)
	if err != nil {
		return err
	}
	_, err = rm.db.ExecContext(ctx,
		`UPDATE architecturedirection.direction_source_capabilities SET stale = TRUE
		 WHERE tenant_id = $1 AND capability_id = $2 AND stale = FALSE`,
		tenantID, string(capabilityID),
	)
	return err
}

func (rm *DirectionReadModel) GetByID(ctx context.Context, id DirectionID) (*DirectionDTO, error) {
	return rm.fetchDirection(ctx, fetchSpec{
		query: `SELECT ` + directionCols + ` FROM architecturedirection.directions WHERE tenant_id = $1 AND id = $2`,
		args:  func(tenantID string) []interface{} { return []interface{}{tenantID, string(id)} },
	})
}

func (rm *DirectionReadModel) GetActiveByEnterpriseCapabilityID(ctx context.Context, enterpriseCapabilityID string) (*DirectionDTO, error) {
	return rm.fetchDirection(ctx, fetchSpec{
		query: `SELECT ` + directionCols + ` FROM architecturedirection.directions
		        WHERE tenant_id = $1 AND enterprise_capability_id = $2 AND status != 'rejected'
		        ORDER BY created_at DESC LIMIT 1`,
		args: func(tenantID string) []interface{} { return []interface{}{tenantID, enterpriseCapabilityID} },
	})
}

func (rm *DirectionReadModel) HasActiveDirectionForEnterpriseCapability(ctx context.Context, enterpriseCapabilityID string) (bool, error) {
	tenantID, err := tenantOf(ctx)
	if err != nil {
		return false, err
	}
	var exists bool
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx,
			`SELECT EXISTS(SELECT 1 FROM architecturedirection.directions
			 WHERE tenant_id = $1 AND enterprise_capability_id = $2 AND status != 'rejected')`,
			tenantID, enterpriseCapabilityID,
		).Scan(&exists)
	})
	return exists, err
}

type fetchSpec struct {
	query string
	args  func(tenantID string) []interface{}
}

func tenantOf(ctx context.Context) (string, error) {
	t, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return "", err
	}
	return t.Value(), nil
}

func (rm *DirectionReadModel) withTx(ctx context.Context, fn func(tx *sql.Tx, tenantID string) error) error {
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

type sourceRowReplacement struct {
	tx                  *sql.Tx
	tenantID            string
	directionID         DirectionID
	sourceCapabilityIDs []CapabilityID
}

func (r sourceRowReplacement) execute(ctx context.Context) error {
	if _, err := r.tx.ExecContext(ctx,
		`DELETE FROM architecturedirection.direction_source_capabilities WHERE tenant_id = $1 AND direction_id = $2`,
		r.tenantID, string(r.directionID),
	); err != nil {
		return err
	}
	for _, sid := range r.sourceCapabilityIDs {
		if _, err := r.tx.ExecContext(ctx,
			`INSERT INTO architecturedirection.direction_source_capabilities
			 (tenant_id, direction_id, capability_id) VALUES ($1, $2, $3)`,
			r.tenantID, string(r.directionID), string(sid),
		); err != nil {
			return err
		}
	}
	return nil
}

func (rm *DirectionReadModel) fetchDirection(ctx context.Context, spec fetchSpec) (*DirectionDTO, error) {
	tenantID, err := tenantOf(ctx)
	if err != nil {
		return nil, err
	}
	var dto *DirectionDTO
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		row := tx.QueryRowContext(ctx, spec.query, spec.args(tenantID)...)
		direction, scanErr := scanDirection(row)
		if scanErr == sql.ErrNoRows {
			return nil
		}
		if scanErr != nil {
			return scanErr
		}
		sources, srcErr := loadSourcesForDirection(ctx, tx, sourceFetchKey{tenantID: tenantID, directionID: DirectionID(direction.ID)})
		if srcErr != nil {
			return srcErr
		}
		direction.SourceCapabilities = sources
		dto = &direction
		return nil
	})
	return dto, err
}

const directionCols = "id, enterprise_capability_id, type, status, horizon, narrative, placements, created_at, updated_at"

type directionRowScanner interface {
	Scan(dest ...any) error
}

func scanDirection(row directionRowScanner) (DirectionDTO, error) {
	var dto DirectionDTO
	var narrative sql.NullString
	var placementsJSON sql.NullString
	var updatedAt sql.NullTime
	err := row.Scan(&dto.ID, &dto.EnterpriseCapabilityID, &dto.Type, &dto.Status, &dto.Horizon,
		&narrative, &placementsJSON, &dto.CreatedAt, &updatedAt)
	if err != nil {
		return dto, err
	}
	if narrative.Valid {
		dto.Narrative = narrative.String
	}
	if updatedAt.Valid {
		dto.UpdatedAt = &updatedAt.Time
	}
	if placementsJSON.Valid && placementsJSON.String != "" {
		if err := json.Unmarshal([]byte(placementsJSON.String), &dto.Placements); err != nil {
			return dto, err
		}
	}
	if dto.Placements == nil {
		dto.Placements = []DirectionPlacementDTO{}
	}
	return dto, nil
}

type sourceFetchKey struct {
	tenantID    string
	directionID DirectionID
}

func loadSourcesForDirection(ctx context.Context, tx *sql.Tx, key sourceFetchKey) ([]DirectionSourceCapabilityDTO, error) {
	rows, err := tx.QueryContext(ctx,
		`SELECT capability_id, stale FROM architecturedirection.direction_source_capabilities
		 WHERE tenant_id = $1 AND direction_id = $2 ORDER BY capability_id`,
		key.tenantID, string(key.directionID),
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	out := []DirectionSourceCapabilityDTO{}
	for rows.Next() {
		var s DirectionSourceCapabilityDTO
		if err := rows.Scan(&s.ID, &s.Stale); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

