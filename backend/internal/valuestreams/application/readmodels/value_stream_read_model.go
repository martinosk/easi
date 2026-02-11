package readmodels

import (
	"context"
	"database/sql"
	"time"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/types"
)

type ValueStreamDTO struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	StageCount  int         `json:"stageCount"`
	CreatedAt   time.Time   `json:"createdAt"`
	UpdatedAt   *time.Time  `json:"updatedAt,omitempty"`
	Links       types.Links `json:"_links,omitempty"`
}

type ValueStreamUpdate struct {
	Name        string
	Description string
}

type ValueStreamReadModel struct {
	db *database.TenantAwareDB
}

func NewValueStreamReadModel(db *database.TenantAwareDB) *ValueStreamReadModel {
	return &ValueStreamReadModel{db: db}
}

func (rm *ValueStreamReadModel) Insert(ctx context.Context, dto ValueStreamDTO) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"INSERT INTO value_streams (id, tenant_id, name, description, stage_count, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
		dto.ID, tenantID.Value(), dto.Name, dto.Description, 0, dto.CreatedAt,
	)
	return err
}

func (rm *ValueStreamReadModel) Update(ctx context.Context, id string, update ValueStreamUpdate) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"UPDATE value_streams SET name = $1, description = $2, updated_at = CURRENT_TIMESTAMP WHERE tenant_id = $3 AND id = $4",
		update.Name, update.Description, tenantID.Value(), id,
	)
	return err
}

func (rm *ValueStreamReadModel) Delete(ctx context.Context, id string) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}
	_, err = rm.db.ExecContext(ctx, "DELETE FROM value_streams WHERE tenant_id = $1 AND id = $2", tenantID.Value(), id)
	return err
}

func (rm *ValueStreamReadModel) GetAll(ctx context.Context) ([]ValueStreamDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var streams []ValueStreamDTO
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx,
			"SELECT id, name, description, stage_count, created_at, updated_at FROM value_streams WHERE tenant_id = $1 ORDER BY name",
			tenantID.Value(),
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			dto, err := scanValueStream(rows)
			if err != nil {
				return err
			}
			streams = append(streams, *dto)
		}

		return rows.Err()
	})

	return streams, err
}

type scanner interface {
	Scan(dest ...interface{}) error
}

func scanValueStream(s scanner) (*ValueStreamDTO, error) {
	var dto ValueStreamDTO
	var updatedAt sql.NullTime

	if err := s.Scan(&dto.ID, &dto.Name, &dto.Description, &dto.StageCount, &dto.CreatedAt, &updatedAt); err != nil {
		return nil, err
	}

	if updatedAt.Valid {
		dto.UpdatedAt = &updatedAt.Time
	}

	return &dto, nil
}

func (rm *ValueStreamReadModel) GetByID(ctx context.Context, id string) (*ValueStreamDTO, error) {
	return rm.getByCondition(ctx, "id = $2", id)
}

func (rm *ValueStreamReadModel) getByCondition(ctx context.Context, whereClause string, arg interface{}) (*ValueStreamDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var dto *ValueStreamDTO
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		row := tx.QueryRowContext(ctx,
			"SELECT id, name, description, stage_count, created_at, updated_at FROM value_streams WHERE tenant_id = $1 AND "+whereClause,
			tenantID.Value(), arg,
		)

		result, err := scanValueStream(row)
		if err == sql.ErrNoRows {
			return nil
		}
		if err != nil {
			return err
		}
		dto = result
		return nil
	})

	return dto, err
}

func (rm *ValueStreamReadModel) NameExists(ctx context.Context, name, excludeID string) (bool, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return false, err
	}

	var count int
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		if excludeID != "" {
			return tx.QueryRowContext(ctx,
				"SELECT COUNT(*) FROM value_streams WHERE tenant_id = $1 AND name = $2 AND id != $3",
				tenantID.Value(), name, excludeID,
			).Scan(&count)
		}
		return tx.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM value_streams WHERE tenant_id = $1 AND name = $2",
			tenantID.Value(), name,
		).Scan(&count)
	})

	if err != nil {
		return false, err
	}

	return count > 0, nil
}
