package readmodels

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"easi/backend/internal/infrastructure/database"
)

type StewardshipRow struct {
	ID          string
	DomainID    string
	Concern     string
	StewardID   string
	StewardName *string
	AssignedBy  string
	AssignedAt  time.Time
}

type StewardshipReadModel struct {
	store tenantScopedDB
}

func NewStewardshipReadModel(db *database.TenantAwareDB) *StewardshipReadModel {
	return &StewardshipReadModel{store: tenantScopedDB{db: db}}
}

func (rm *StewardshipReadModel) Upsert(ctx context.Context, row StewardshipRow) error {
	return rm.store.exec(ctx, "record stewardship "+row.ID, `
		INSERT INTO stewardship.stewardships (tenant_id, id, domain_id, concern, steward_id, assigned_by, assigned_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (tenant_id, id) DO UPDATE SET
			steward_id = EXCLUDED.steward_id, assigned_by = EXCLUDED.assigned_by, assigned_at = EXCLUDED.assigned_at`,
		row.ID, row.DomainID, row.Concern, row.StewardID, row.AssignedBy, row.AssignedAt.UTC())
}

func (rm *StewardshipReadModel) Delete(ctx context.Context, id string) error {
	return rm.store.exec(ctx, "remove stewardship "+id, `
		DELETE FROM stewardship.stewardships WHERE tenant_id = $1 AND id = $2`,
		id)
}

func (rm *StewardshipReadModel) FindLiveStewardshipID(ctx context.Context, domainID, concern string) (string, bool, error) {
	var id string
	err := rm.store.read(ctx, "find "+concern+" stewardship of domain "+domainID, func(tx *sql.Tx, tenantID string) error {
		return tx.QueryRowContext(ctx, `
			SELECT id FROM stewardship.stewardships WHERE tenant_id = $1 AND domain_id = $2 AND concern = $3`,
			tenantID, domainID, concern,
		).Scan(&id)
	})
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	return id, err == nil, err
}

func (rm *StewardshipReadModel) ForDomain(ctx context.Context, domainID string) ([]StewardshipRow, error) {
	var rows []StewardshipRow
	err := rm.store.read(ctx, "list stewardships of domain "+domainID, func(tx *sql.Tx, tenantID string) error {
		result, err := tx.QueryContext(ctx, `
			SELECT s.id, s.domain_id, s.concern, s.steward_id, NULLIF(COALESCE(NULLIF(u.name, ''), u.email), ''),
			       s.assigned_by, s.assigned_at
			FROM stewardship.stewardships s
			LEFT JOIN stewardship.user_cache u ON u.tenant_id = s.tenant_id AND u.user_id = s.steward_id
			WHERE s.tenant_id = $1 AND s.domain_id = $2`,
			tenantID, domainID)
		if err != nil {
			return err
		}
		defer func() { _ = result.Close() }()
		for result.Next() {
			var (
				row  StewardshipRow
				name sql.NullString
			)
			if err := result.Scan(&row.ID, &row.DomainID, &row.Concern, &row.StewardID, &name, &row.AssignedBy, &row.AssignedAt); err != nil {
				return err
			}
			row.StewardName = nullableString(name)
			row.AssignedAt = row.AssignedAt.UTC()
			rows = append(rows, row)
		}
		return result.Err()
	})
	return rows, err
}

func (rm *StewardshipReadModel) LiveConcernsOfDomain(ctx context.Context, domainID string) ([]string, error) {
	rows, err := rm.ForDomain(ctx, domainID)
	if err != nil {
		return nil, err
	}
	concerns := make([]string, len(rows))
	for i, row := range rows {
		concerns[i] = row.Concern
	}
	return concerns, nil
}
