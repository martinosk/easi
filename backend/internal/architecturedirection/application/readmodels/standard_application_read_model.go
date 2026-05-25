package readmodels

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/types"
)

const standardApplicationPerECUniqueConstraint = "uq_standard_applications_per_ec"

var ErrStandardApplicationAlreadyExists = errors.New("a standard application already exists for this enterprise capability")

type StandardApplicationDTO struct {
	ID                     string      `json:"id"`
	EnterpriseCapabilityID string      `json:"enterpriseCapabilityId"`
	ApplicationID          string      `json:"applicationId"`
	ApplicationStale       bool        `json:"applicationStale"`
	ApplicationName        *string     `json:"applicationName"`
	Narrative              string      `json:"narrative"`
	SetAt                  time.Time   `json:"setAt"`
	UpdatedAt              *time.Time  `json:"updatedAt,omitempty"`
	Links                  types.Links `json:"_links,omitempty"`
}

type StandardApplicationHistoryEntryDTO struct {
	ApplicationID           string  `json:"applicationId"`
	PreviousApplicationID   string  `json:"previousApplicationId,omitempty"`
	ApplicationName         *string `json:"applicationName"`
	PreviousApplicationName *string `json:"previousApplicationName,omitempty"`
	Narrative               string  `json:"narrative"`
	SetAt                   time.Time `json:"setAt"`
}

type StandardApplicationHistoryDTO struct {
	StandardApplicationID  string                               `json:"standardApplicationId"`
	EnterpriseCapabilityID string                               `json:"enterpriseCapabilityId"`
	Entries                []StandardApplicationHistoryEntryDTO `json:"entries"`
	Links                  types.Links                          `json:"_links,omitempty"`
}

type UpsertStandardApplicationParams struct {
	ID                     string
	EnterpriseCapabilityID string
	ApplicationID          string
	Narrative              string
	SetAt                  time.Time
}

type AppendStandardApplicationHistoryParams struct {
	StandardApplicationID string
	ApplicationID         string
	PreviousApplicationID string
	Narrative             string
	SetAt                 time.Time
}

type StandardApplicationReadModel struct {
	db *database.TenantAwareDB
}

func NewStandardApplicationReadModel(db *database.TenantAwareDB) *StandardApplicationReadModel {
	return &StandardApplicationReadModel{db: db}
}

func (rm *StandardApplicationReadModel) UpsertCurrent(ctx context.Context, p UpsertStandardApplicationParams) error {
	tenantID, err := standardTenantOf(ctx)
	if err != nil {
		return err
	}
	_, err = rm.db.ExecContext(ctx,
		`INSERT INTO architecturedirection.standard_applications
		 (id, tenant_id, enterprise_capability_id, application_id, narrative, set_at, application_stale, application_name)
		 SELECT $1, $2, $3, $4, $5, $6, FALSE, rnc.name
		 FROM (SELECT 1) AS stub
		 LEFT JOIN architecturedirection.reference_name_cache rnc
		   ON rnc.tenant_id = $2 AND rnc.entity_type = 'application' AND rnc.entity_id = $4
		 ON CONFLICT (tenant_id, id) DO UPDATE SET
		   application_id = EXCLUDED.application_id,
		   application_name = EXCLUDED.application_name,
		   narrative = EXCLUDED.narrative,
		   set_at = EXCLUDED.set_at,
		   application_stale = FALSE,
		   updated_at = CURRENT_TIMESTAMP`,
		p.ID, tenantID, p.EnterpriseCapabilityID, p.ApplicationID, p.Narrative, p.SetAt,
	)
	return mapStandardApplicationInsertError(err)
}

func mapStandardApplicationInsertError(err error) error {
	if isStandardApplicationPerECUniqueViolation(err) {
		return ErrStandardApplicationAlreadyExists
	}
	return err
}

func isStandardApplicationPerECUniqueViolation(err error) bool {
	var pqErr *pq.Error
	if !errors.As(err, &pqErr) {
		return false
	}
	return string(pqErr.Code) == pgUniqueViolation && pqErr.Constraint == standardApplicationPerECUniqueConstraint
}

func (rm *StandardApplicationReadModel) AppendHistory(ctx context.Context, p AppendStandardApplicationHistoryParams) error {
	tenantID, err := standardTenantOf(ctx)
	if err != nil {
		return err
	}
	previous := sql.NullString{}
	if p.PreviousApplicationID != "" {
		previous = sql.NullString{String: p.PreviousApplicationID, Valid: true}
	}
	tx, err := rm.db.BeginTxWithTenant(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var nextSequence int
	if err := tx.QueryRowContext(ctx,
		`SELECT COALESCE(MAX(sequence), 0) + 1
		 FROM architecturedirection.standard_application_history
		 WHERE tenant_id = $1 AND standard_application_id = $2`,
		tenantID, p.StandardApplicationID,
	).Scan(&nextSequence); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO architecturedirection.standard_application_history
		 (tenant_id, standard_application_id, sequence, application_id, previous_application_id, narrative, set_at,
		  application_name, previous_application_name)
		 SELECT $1, $2, $3, $4, $5, $6, $7, cur.name, prev.name
		 FROM (SELECT 1) AS stub
		 LEFT JOIN architecturedirection.reference_name_cache cur
		   ON cur.tenant_id = $1 AND cur.entity_type = 'application' AND cur.entity_id = $4
		 LEFT JOIN architecturedirection.reference_name_cache prev
		   ON prev.tenant_id = $1 AND prev.entity_type = 'application' AND prev.entity_id = $5`,
		tenantID, p.StandardApplicationID, nextSequence, p.ApplicationID, previous, p.Narrative, p.SetAt,
	); err != nil {
		return err
	}
	return tx.Commit()
}

func (rm *StandardApplicationReadModel) tenantExec(ctx context.Context, query string, argsFn func(tenantID string) []any) error {
	tenantID, err := standardTenantOf(ctx)
	if err != nil {
		return err
	}
	_, err = rm.db.ExecContext(ctx, query, argsFn(tenantID)...)
	return err
}

func (rm *StandardApplicationReadModel) CacheApplicationName(ctx context.Context, applicationID, name string) error {
	return rm.tenantExec(ctx,
		`INSERT INTO architecturedirection.reference_name_cache (tenant_id, entity_type, entity_id, name)
		 VALUES ($1, 'application', $2, $3)
		 ON CONFLICT (tenant_id, entity_type, entity_id) DO UPDATE SET name = EXCLUDED.name`,
		func(t string) []any { return []any{t, applicationID, name} },
	)
}

func (rm *StandardApplicationReadModel) UpdateApplicationName(ctx context.Context, applicationID, name string) error {
	return rm.tenantExec(ctx,
		`UPDATE architecturedirection.standard_applications SET application_name = $1
		 WHERE tenant_id = $2 AND application_id = $3`,
		func(t string) []any { return []any{name, t, applicationID} },
	)
}

func (rm *StandardApplicationReadModel) MarkApplicationStale(ctx context.Context, applicationID string) error {
	return rm.tenantExec(ctx,
		`UPDATE architecturedirection.standard_applications SET application_stale = TRUE
		 WHERE tenant_id = $1 AND application_id = $2 AND application_stale = FALSE`,
		func(t string) []any { return []any{t, applicationID} },
	)
}

func (rm *StandardApplicationReadModel) FindAggregateIDForEnterpriseCapability(ctx context.Context, ecID string) (string, bool, error) {
	tenantID, err := standardTenantOf(ctx)
	if err != nil {
		return "", false, err
	}
	var id string
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx,
			`SELECT id FROM architecturedirection.standard_applications
			 WHERE tenant_id = $1 AND enterprise_capability_id = $2`,
			tenantID, ecID,
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

func (rm *StandardApplicationReadModel) GetCurrentByEnterpriseCapability(ctx context.Context, ecID string) (*StandardApplicationDTO, error) {
	tenantID, err := standardTenantOf(ctx)
	if err != nil {
		return nil, err
	}
	var dto *StandardApplicationDTO
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		row := tx.QueryRowContext(ctx,
			`SELECT id, enterprise_capability_id, application_id, application_stale, application_name, narrative, set_at, updated_at
			 FROM architecturedirection.standard_applications
			 WHERE tenant_id = $1 AND enterprise_capability_id = $2`,
			tenantID, ecID,
		)
		fetched, scanErr := scanStandardApplication(row)
		if scanErr == sql.ErrNoRows {
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

func (rm *StandardApplicationReadModel) GetHistoryByAggregateID(ctx context.Context, id string) (*StandardApplicationHistoryDTO, error) {
	tenantID, err := standardTenantOf(ctx)
	if err != nil {
		return nil, err
	}
	var history *StandardApplicationHistoryDTO
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		var ecID string
		err := tx.QueryRowContext(ctx,
			`SELECT enterprise_capability_id FROM architecturedirection.standard_applications
			 WHERE tenant_id = $1 AND id = $2`,
			tenantID, id,
		).Scan(&ecID)
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		entries, entriesErr := loadHistoryEntries(ctx, tx, tenantID, id)
		if entriesErr != nil {
			return entriesErr
		}
		history = &StandardApplicationHistoryDTO{
			StandardApplicationID:  id,
			EnterpriseCapabilityID: ecID,
			Entries:                entries,
		}
		return nil
	})
	return history, err
}

func loadHistoryEntries(ctx context.Context, tx *sql.Tx, tenantID, aggregateID string) ([]StandardApplicationHistoryEntryDTO, error) {
	rows, err := tx.QueryContext(ctx,
		`SELECT application_id, previous_application_id, application_name, previous_application_name, narrative, set_at
		 FROM architecturedirection.standard_application_history
		 WHERE tenant_id = $1 AND standard_application_id = $2
		 ORDER BY sequence DESC`,
		tenantID, aggregateID,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	entries := []StandardApplicationHistoryEntryDTO{}
	for rows.Next() {
		entry, scanErr := scanHistoryEntry(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		entries = append(entries, entry)
	}
	return entries, rows.Err()
}

type standardApplicationRowScanner interface {
	Scan(dest ...any) error
}

func scanStandardApplication(row standardApplicationRowScanner) (StandardApplicationDTO, error) {
	var dto StandardApplicationDTO
	var updatedAt sql.NullTime
	err := row.Scan(&dto.ID, &dto.EnterpriseCapabilityID, &dto.ApplicationID, &dto.ApplicationStale,
		&dto.ApplicationName, &dto.Narrative, &dto.SetAt, &updatedAt)
	if err != nil {
		return dto, err
	}
	if updatedAt.Valid {
		dto.UpdatedAt = &updatedAt.Time
	}
	return dto, nil
}

func scanHistoryEntry(row standardApplicationRowScanner) (StandardApplicationHistoryEntryDTO, error) {
	var entry StandardApplicationHistoryEntryDTO
	var previous sql.NullString
	if err := row.Scan(&entry.ApplicationID, &previous, &entry.ApplicationName, &entry.PreviousApplicationName,
		&entry.Narrative, &entry.SetAt); err != nil {
		return entry, err
	}
	if previous.Valid {
		entry.PreviousApplicationID = previous.String
	}
	return entry, nil
}

func standardTenantOf(ctx context.Context) (string, error) {
	t, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return "", err
	}
	return t.Value(), nil
}
