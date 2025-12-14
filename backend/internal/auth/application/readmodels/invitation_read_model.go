package readmodels

import (
	"context"
	"database/sql"
	"time"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
)

type InvitationDTO struct {
	ID         string            `json:"id"`
	Email      string            `json:"email"`
	Role       string            `json:"role"`
	Status     string            `json:"status"`
	InvitedBy  *string           `json:"invitedBy,omitempty"`
	CreatedAt  time.Time         `json:"createdAt"`
	ExpiresAt  time.Time         `json:"expiresAt"`
	AcceptedAt *time.Time        `json:"acceptedAt,omitempty"`
	RevokedAt  *time.Time        `json:"revokedAt,omitempty"`
	Links      map[string]string `json:"_links,omitempty"`
}

type StatusUpdate struct {
	ID         string
	Status     string
	AcceptedAt *time.Time
	RevokedAt  *time.Time
}

type InvitationReadModel struct {
	db *database.TenantAwareDB
}

type paginationQuery struct {
	afterCursor    string
	afterTimestamp int64
	limit          int
}

func NewInvitationReadModel(db *database.TenantAwareDB) *InvitationReadModel {
	return &InvitationReadModel{db: db}
}

func (rm *InvitationReadModel) getTenantID(ctx context.Context) (string, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return "", err
	}
	return tenantID.Value(), nil
}

func (rm *InvitationReadModel) Insert(ctx context.Context, dto InvitationDTO) error {
	tenantID, err := rm.getTenantID(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		`INSERT INTO invitations (id, tenant_id, email, role, status, invited_by, created_at, expires_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		dto.ID, tenantID, dto.Email, dto.Role, dto.Status, dto.InvitedBy, dto.CreatedAt, dto.ExpiresAt,
	)
	return err
}

func (rm *InvitationReadModel) UpdateStatus(ctx context.Context, update StatusUpdate) error {
	tenantID, err := rm.getTenantID(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		`UPDATE invitations SET status = $1, accepted_at = $2, revoked_at = $3 WHERE tenant_id = $4 AND id = $5`,
		update.Status, update.AcceptedAt, update.RevokedAt, tenantID, update.ID,
	)
	return err
}

func (rm *InvitationReadModel) GetByID(ctx context.Context, id string) (*InvitationDTO, error) {
	return rm.queryOneWithTenant(ctx,
		`SELECT id, email, role, status, invited_by, created_at, expires_at, accepted_at, revoked_at
		 FROM invitations WHERE tenant_id = $1 AND id = $2`, id)
}

func (rm *InvitationReadModel) GetPendingByEmail(ctx context.Context, email string) (*InvitationDTO, error) {
	return rm.queryOneWithTenant(ctx,
		`SELECT id, email, role, status, invited_by, created_at, expires_at, accepted_at, revoked_at
		 FROM invitations
		 WHERE tenant_id = $1 AND email = $2 AND status = 'pending' AND expires_at > NOW()
		 ORDER BY created_at DESC
		 LIMIT 1`, email)
}

func (rm *InvitationReadModel) queryOneWithTenant(ctx context.Context, query string, extraArgs ...interface{}) (*InvitationDTO, error) {
	tenantID, err := rm.getTenantID(ctx)
	if err != nil {
		return nil, err
	}
	args := append([]interface{}{tenantID}, extraArgs...)
	return rm.queryOne(ctx, query, args...)
}

func (rm *InvitationReadModel) queryOne(ctx context.Context, query string, args ...interface{}) (*InvitationDTO, error) {
	var dto InvitationDTO
	var notFound bool

	err := rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		err := tx.QueryRowContext(ctx, query, args...).Scan(
			&dto.ID, &dto.Email, &dto.Role, &dto.Status, &dto.InvitedBy,
			&dto.CreatedAt, &dto.ExpiresAt, &dto.AcceptedAt, &dto.RevokedAt,
		)
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

func (rm *InvitationReadModel) GetAll(ctx context.Context) ([]InvitationDTO, error) {
	return rm.queryManyWithTenant(ctx,
		`SELECT id, email, role, status, invited_by, created_at, expires_at, accepted_at, revoked_at
		 FROM invitations WHERE tenant_id = $1
		 ORDER BY created_at DESC`)
}

func (rm *InvitationReadModel) GetAllPaginated(ctx context.Context, limit int, afterCursor string, afterTimestamp int64) ([]InvitationDTO, bool, error) {
	tenantID, err := rm.getTenantID(ctx)
	if err != nil {
		return nil, false, err
	}

	fetchLimit := limit + 1
	pq := paginationQuery{afterCursor: afterCursor, afterTimestamp: afterTimestamp, limit: fetchLimit}
	invitations, err := rm.queryPaginated(ctx, tenantID, pq)
	if err != nil {
		return nil, false, err
	}

	hasMore := len(invitations) > limit
	if hasMore {
		invitations = invitations[:limit]
	}

	return invitations, hasMore, nil
}

func (rm *InvitationReadModel) queryPaginated(ctx context.Context, tenantID string, pq paginationQuery) ([]InvitationDTO, error) {
	if pq.afterCursor == "" {
		return rm.queryMany(ctx,
			`SELECT id, email, role, status, invited_by, created_at, expires_at, accepted_at, revoked_at
			 FROM invitations WHERE tenant_id = $1
			 ORDER BY created_at DESC, id DESC
			 LIMIT $2`,
			tenantID, pq.limit,
		)
	}

	return rm.queryMany(ctx,
		`SELECT id, email, role, status, invited_by, created_at, expires_at, accepted_at, revoked_at
		 FROM invitations
		 WHERE tenant_id = $1
		 AND (created_at < to_timestamp($2) OR (created_at = to_timestamp($2) AND id < $3))
		 ORDER BY created_at DESC, id DESC
		 LIMIT $4`,
		tenantID, pq.afterTimestamp, pq.afterCursor, pq.limit,
	)
}

func (rm *InvitationReadModel) queryManyWithTenant(ctx context.Context, query string, extraArgs ...interface{}) ([]InvitationDTO, error) {
	tenantID, err := rm.getTenantID(ctx)
	if err != nil {
		return nil, err
	}
	args := append([]interface{}{tenantID}, extraArgs...)
	return rm.queryMany(ctx, query, args...)
}

func (rm *InvitationReadModel) queryMany(ctx context.Context, query string, args ...interface{}) ([]InvitationDTO, error) {
	var invitations []InvitationDTO

	err := rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, query, args...)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var dto InvitationDTO
			if err := rows.Scan(
				&dto.ID, &dto.Email, &dto.Role, &dto.Status, &dto.InvitedBy,
				&dto.CreatedAt, &dto.ExpiresAt, &dto.AcceptedAt, &dto.RevokedAt,
			); err != nil {
				return err
			}
			invitations = append(invitations, dto)
		}
		return rows.Err()
	})

	return invitations, err
}

func (rm *InvitationReadModel) ExistsPendingForEmail(ctx context.Context, email string) (bool, error) {
	tenantID, err := rm.getTenantID(ctx)
	if err != nil {
		return false, err
	}

	var exists bool
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx,
			`SELECT EXISTS(SELECT 1 FROM invitations WHERE tenant_id = $1 AND email = $2 AND status = 'pending' AND expires_at > NOW())`,
			tenantID, email,
		).Scan(&exists)
	})

	return exists, err
}
