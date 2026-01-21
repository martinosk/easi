package readmodels

import (
	"context"
	"database/sql"
	"time"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/types"
)

type AcquiredEntityDTO struct {
	ID                string      `json:"id"`
	Name              string      `json:"name"`
	AcquisitionDate   *time.Time  `json:"acquisitionDate,omitempty"`
	IntegrationStatus string      `json:"integrationStatus,omitempty"`
	Notes             string      `json:"notes,omitempty"`
	CreatedAt         time.Time   `json:"createdAt"`
	Links             types.Links `json:"_links,omitempty"`
}

type AcquiredEntityReadModel struct {
	db *database.TenantAwareDB
}

func NewAcquiredEntityReadModel(db *database.TenantAwareDB) *AcquiredEntityReadModel {
	return &AcquiredEntityReadModel{db: db}
}

func (rm *AcquiredEntityReadModel) Insert(ctx context.Context, dto AcquiredEntityDTO) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"INSERT INTO acquired_entities (id, tenant_id, name, acquisition_date, integration_status, notes, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		dto.ID, tenantID.Value(), dto.Name, dto.AcquisitionDate, dto.IntegrationStatus, dto.Notes, dto.CreatedAt,
	)
	return err
}

func (rm *AcquiredEntityReadModel) Update(ctx context.Context, id, name string, acquisitionDate *time.Time, integrationStatus, notes string) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"UPDATE acquired_entities SET name = $1, acquisition_date = $2, integration_status = $3, notes = $4, updated_at = CURRENT_TIMESTAMP WHERE tenant_id = $5 AND id = $6",
		name, acquisitionDate, integrationStatus, notes, tenantID.Value(), id,
	)
	return err
}

func (rm *AcquiredEntityReadModel) MarkAsDeleted(ctx context.Context, id string, deletedAt time.Time) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"UPDATE acquired_entities SET is_deleted = TRUE, deleted_at = $1 WHERE tenant_id = $2 AND id = $3",
		deletedAt, tenantID.Value(), id,
	)
	return err
}

func (rm *AcquiredEntityReadModel) GetByID(ctx context.Context, id string) (*AcquiredEntityDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var dto AcquiredEntityDTO
	var notFound bool

	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		err := tx.QueryRowContext(ctx,
			"SELECT id, name, acquisition_date, integration_status, notes, created_at FROM acquired_entities WHERE tenant_id = $1 AND id = $2 AND is_deleted = FALSE",
			tenantID.Value(), id,
		).Scan(&dto.ID, &dto.Name, &dto.AcquisitionDate, &dto.IntegrationStatus, &dto.Notes, &dto.CreatedAt)

		if err == sql.ErrNoRows {
			notFound = true
			return nil
		}
		return err
	})

	if err != nil {
		return nil, err
	}
	if notFound {
		return nil, nil
	}

	return &dto, nil
}

func (rm *AcquiredEntityReadModel) GetAll(ctx context.Context) ([]AcquiredEntityDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	entities := make([]AcquiredEntityDTO, 0)
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx,
			"SELECT id, name, acquisition_date, integration_status, notes, created_at FROM acquired_entities WHERE tenant_id = $1 AND is_deleted = FALSE ORDER BY LOWER(name) ASC",
			tenantID.Value(),
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var dto AcquiredEntityDTO
			if err := rows.Scan(&dto.ID, &dto.Name, &dto.AcquisitionDate, &dto.IntegrationStatus, &dto.Notes, &dto.CreatedAt); err != nil {
				return err
			}
			entities = append(entities, dto)
		}

		return rows.Err()
	})

	return entities, err
}

func (rm *AcquiredEntityReadModel) GetAllPaginated(ctx context.Context, limit int, afterCursor string, afterName string) ([]AcquiredEntityDTO, bool, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, false, err
	}

	queryLimit := limit + 1
	entities := make([]AcquiredEntityDTO, 0)

	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		var rows *sql.Rows
		var err error

		if afterCursor == "" {
			rows, err = tx.QueryContext(ctx,
				"SELECT id, name, acquisition_date, integration_status, notes, created_at FROM acquired_entities WHERE tenant_id = $1 AND is_deleted = FALSE ORDER BY LOWER(name) ASC, id ASC LIMIT $2",
				tenantID.Value(), queryLimit,
			)
		} else {
			rows, err = tx.QueryContext(ctx,
				"SELECT id, name, acquisition_date, integration_status, notes, created_at FROM acquired_entities WHERE tenant_id = $1 AND is_deleted = FALSE AND (LOWER(name) > LOWER($2) OR (LOWER(name) = LOWER($2) AND id > $3)) ORDER BY LOWER(name) ASC, id ASC LIMIT $4",
				tenantID.Value(), afterName, afterCursor, queryLimit,
			)
		}
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var dto AcquiredEntityDTO
			if err := rows.Scan(&dto.ID, &dto.Name, &dto.AcquisitionDate, &dto.IntegrationStatus, &dto.Notes, &dto.CreatedAt); err != nil {
				return err
			}
			entities = append(entities, dto)
		}

		return rows.Err()
	})

	if err != nil {
		return nil, false, err
	}

	hasMore := len(entities) > limit
	if hasMore {
		entities = entities[:limit]
	}

	return entities, hasMore, nil
}
