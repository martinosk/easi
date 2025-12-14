package readmodels

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
)

type UserDTO struct {
	ID           uuid.UUID  `json:"id"`
	Email        string     `json:"email"`
	Name         *string    `json:"name,omitempty"`
	Role         string     `json:"role"`
	Status       string     `json:"status"`
	ExternalID   *string    `json:"externalId,omitempty"`
	InvitationID *uuid.UUID `json:"invitationId,omitempty"`
	CreatedAt    time.Time  `json:"createdAt"`
	LastLoginAt  *time.Time `json:"lastLoginAt,omitempty"`
}

type UserReadModel struct {
	db *database.TenantAwareDB
}

func NewUserReadModel(db *database.TenantAwareDB) *UserReadModel {
	return &UserReadModel{db: db}
}

func (rm *UserReadModel) GetByID(ctx context.Context, id uuid.UUID) (*UserDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var dto UserDTO
	var notFound bool

	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		err := tx.QueryRowContext(ctx,
			`SELECT id, email, name, role, status, external_id, invitation_id, created_at, last_login_at
			 FROM users WHERE tenant_id = $1 AND id = $2`,
			tenantID.Value(), id,
		).Scan(&dto.ID, &dto.Email, &dto.Name, &dto.Role, &dto.Status,
			&dto.ExternalID, &dto.InvitationID, &dto.CreatedAt, &dto.LastLoginAt)

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

func (rm *UserReadModel) GetByEmail(ctx context.Context, email string) (*UserDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var dto UserDTO
	var notFound bool

	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		err := tx.QueryRowContext(ctx,
			`SELECT id, email, name, role, status, external_id, invitation_id, created_at, last_login_at
			 FROM users WHERE tenant_id = $1 AND email = $2`,
			tenantID.Value(), email,
		).Scan(&dto.ID, &dto.Email, &dto.Name, &dto.Role, &dto.Status,
			&dto.ExternalID, &dto.InvitationID, &dto.CreatedAt, &dto.LastLoginAt)

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

func (rm *UserReadModel) Insert(ctx context.Context, dto UserDTO) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		`INSERT INTO users (id, tenant_id, email, name, role, status, external_id, invitation_id, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		dto.ID, tenantID.Value(), dto.Email, dto.Name, dto.Role, dto.Status, dto.ExternalID, dto.InvitationID, dto.CreatedAt,
	)
	return err
}

func (rm *UserReadModel) UpdateLastLogin(ctx context.Context, id uuid.UUID, lastLoginAt time.Time) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		`UPDATE users SET last_login_at = $1, updated_at = CURRENT_TIMESTAMP WHERE tenant_id = $2 AND id = $3`,
		lastLoginAt, tenantID.Value(), id,
	)
	return err
}
