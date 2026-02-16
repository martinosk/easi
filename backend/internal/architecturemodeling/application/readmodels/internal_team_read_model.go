package readmodels

import (
	"context"
	"database/sql"
	"time"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/types"
)

type InternalTeamDTO struct {
	ID            string      `json:"id"`
	Name          string      `json:"name"`
	Department    string      `json:"department,omitempty"`
	ContactPerson string      `json:"contactPerson,omitempty"`
	Notes         string      `json:"notes,omitempty"`
	CreatedAt     time.Time   `json:"createdAt"`
	UpdatedAt     *time.Time  `json:"updatedAt,omitempty"`
	Links         types.Links `json:"_links,omitempty"`
}

type InternalTeamReadModel struct {
	db *database.TenantAwareDB
}

func NewInternalTeamReadModel(db *database.TenantAwareDB) *InternalTeamReadModel {
	return &InternalTeamReadModel{db: db}
}

func (rm *InternalTeamReadModel) Insert(ctx context.Context, dto InternalTeamDTO) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"DELETE FROM architecturemodeling.internal_teams WHERE tenant_id = $1 AND id = $2",
		tenantID.Value(), dto.ID,
	)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		`INSERT INTO architecturemodeling.internal_teams
		(id, tenant_id, name, department, contact_person, notes, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		dto.ID, tenantID.Value(), dto.Name, dto.Department, dto.ContactPerson, dto.Notes, dto.CreatedAt,
	)
	return err
}

type InternalTeamUpdate struct {
	ID            string
	Name          string
	Department    string
	ContactPerson string
	Notes         string
}

func (rm *InternalTeamReadModel) Update(ctx context.Context, update InternalTeamUpdate) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"UPDATE architecturemodeling.internal_teams SET name = $1, department = $2, contact_person = $3, notes = $4, updated_at = CURRENT_TIMESTAMP WHERE tenant_id = $5 AND id = $6",
		update.Name, update.Department, update.ContactPerson, update.Notes, tenantID.Value(), update.ID,
	)
	return err
}

func (rm *InternalTeamReadModel) MarkAsDeleted(ctx context.Context, id string, deletedAt time.Time) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"UPDATE architecturemodeling.internal_teams SET is_deleted = TRUE, deleted_at = $1 WHERE tenant_id = $2 AND id = $3",
		deletedAt, tenantID.Value(), id,
	)
	return err
}

func (rm *InternalTeamReadModel) GetByID(ctx context.Context, id string) (*InternalTeamDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var dto InternalTeamDTO
	var notFound bool

	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		err := tx.QueryRowContext(ctx,
			"SELECT id, name, department, contact_person, notes, created_at, updated_at FROM architecturemodeling.internal_teams WHERE tenant_id = $1 AND id = $2 AND is_deleted = FALSE",
			tenantID.Value(), id,
		).Scan(&dto.ID, &dto.Name, &dto.Department, &dto.ContactPerson, &dto.Notes, &dto.CreatedAt, &dto.UpdatedAt)

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

func (rm *InternalTeamReadModel) GetAll(ctx context.Context) ([]InternalTeamDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	teams := make([]InternalTeamDTO, 0)
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx,
			"SELECT id, name, department, contact_person, notes, created_at, updated_at FROM architecturemodeling.internal_teams WHERE tenant_id = $1 AND is_deleted = FALSE ORDER BY LOWER(name) ASC",
			tenantID.Value(),
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var dto InternalTeamDTO
			if err := rows.Scan(&dto.ID, &dto.Name, &dto.Department, &dto.ContactPerson, &dto.Notes, &dto.CreatedAt, &dto.UpdatedAt); err != nil {
				return err
			}
			teams = append(teams, dto)
		}

		return rows.Err()
	})

	return teams, err
}

func (rm *InternalTeamReadModel) GetAllPaginated(ctx context.Context, limit int, afterCursor string, afterName string) ([]InternalTeamDTO, bool, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, false, err
	}

	queryLimit := limit + 1
	teams := make([]InternalTeamDTO, 0)
	query, args := internalTeamPageQuery(tenantID.Value(), queryLimit, afterCursor, afterName)

	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, query, args...)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var dto InternalTeamDTO
			if err := rows.Scan(&dto.ID, &dto.Name, &dto.Department, &dto.ContactPerson, &dto.Notes, &dto.CreatedAt, &dto.UpdatedAt); err != nil {
				return err
			}
			teams = append(teams, dto)
		}

		return rows.Err()
	})

	if err != nil {
		return nil, false, err
	}

	hasMore := len(teams) > limit
	if hasMore {
		teams = teams[:limit]
	}

	return teams, hasMore, nil
}

func internalTeamPageQuery(tenantID string, queryLimit int, afterCursor, afterName string) (string, []any) {
	const selectCols = "id, name, department, contact_person, notes, created_at, updated_at"
	const base = "SELECT " + selectCols + " FROM architecturemodeling.internal_teams WHERE tenant_id = $1 AND is_deleted = FALSE"
	const order = " ORDER BY LOWER(name) ASC, id ASC"
	if afterCursor == "" {
		return base + order + " LIMIT $2", []any{tenantID, queryLimit}
	}
	return base + " AND (LOWER(name) > LOWER($2) OR (LOWER(name) = LOWER($2) AND id > $3))" + order + " LIMIT $4",
		[]any{tenantID, afterName, afterCursor, queryLimit}
}
