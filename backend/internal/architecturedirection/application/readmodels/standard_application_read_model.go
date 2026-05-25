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
	Narrative              string      `json:"narrative"`
	SetAt                  time.Time   `json:"setAt"`
	UpdatedAt              *time.Time  `json:"updatedAt,omitempty"`
	Links                  types.Links `json:"_links,omitempty"`
}

type StandardApplicationHistoryEntryDTO struct {
	ApplicationID         string    `json:"applicationId"`
	PreviousApplicationID string    `json:"previousApplicationId,omitempty"`
	Narrative             string    `json:"narrative"`
	SetAt                 time.Time `json:"setAt"`
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
		 (id, tenant_id, enterprise_capability_id, application_id, narrative, set_at, application_stale)
		 VALUES ($1, $2, $3, $4, $5, $6, FALSE)
		 ON CONFLICT (tenant_id, id) DO UPDATE SET
		   application_id = EXCLUDED.application_id,
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
		 (tenant_id, standard_application_id, sequence, application_id, previous_application_id, narrative, set_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		tenantID, p.StandardApplicationID, nextSequence, p.ApplicationID, previous, p.Narrative, p.SetAt,
	); err != nil {
		return err
	}
	return tx.Commit()
}

func (rm *StandardApplicationReadModel) MarkApplicationStale(ctx context.Context, applicationID string) error {
	tenantID, err := standardTenantOf(ctx)
	if err != nil {
		return err
	}
	_, err = rm.db.ExecContext(ctx,
		`UPDATE architecturedirection.standard_applications SET application_stale = TRUE
		 WHERE tenant_id = $1 AND application_id = $2 AND application_stale = FALSE`,
		tenantID, applicationID,
	)
	return err
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
			`SELECT id, enterprise_capability_id, application_id, application_stale, narrative, set_at, updated_at
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
		`SELECT application_id, previous_application_id, narrative, set_at
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
		&dto.Narrative, &dto.SetAt, &updatedAt)
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
	if err := row.Scan(&entry.ApplicationID, &previous, &entry.Narrative, &entry.SetAt); err != nil {
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
