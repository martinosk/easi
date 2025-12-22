package readmodels

import (
	"context"
	"database/sql"
	"strconv"
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

type UserEventData struct {
	ID           string
	Email        string
	Name         string
	Role         string
	Status       string
	ExternalID   string
	InvitationID string
	CreatedAt    string
}

type UserPaginationFilter struct {
	Limit          int
	AfterCursor    string
	AfterTimestamp int64
	StatusFilter   string
	RoleFilter     string
}

type UserReadModel struct {
	db *database.TenantAwareDB
}

func NewUserReadModel(db *database.TenantAwareDB) *UserReadModel {
	return &UserReadModel{db: db}
}

func (rm *UserReadModel) GetByID(ctx context.Context, id uuid.UUID) (*UserDTO, error) {
	return rm.getByField(ctx, "id", id)
}

func (rm *UserReadModel) GetByEmail(ctx context.Context, email string) (*UserDTO, error) {
	return rm.getByField(ctx, "email", email)
}

func (rm *UserReadModel) getByField(ctx context.Context, field string, value interface{}) (*UserDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var dto UserDTO
	var notFound bool

	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		query := `SELECT id, email, name, role, status, external_id, invitation_id, created_at, last_login_at
			 FROM users WHERE tenant_id = $1 AND ` + field + ` = $2`

		err := tx.QueryRowContext(ctx, query, tenantID.Value(), value).
			Scan(&dto.ID, &dto.Email, &dto.Name, &dto.Role, &dto.Status,
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

func (rm *UserReadModel) InsertFromEvent(ctx context.Context, data UserEventData) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	userID, err := uuid.Parse(data.ID)
	if err != nil {
		return err
	}

	createdAt, err := time.Parse(time.RFC3339, data.CreatedAt)
	if err != nil {
		return err
	}

	namePtr := toStringPtr(data.Name)
	externalIDPtr := toStringPtr(data.ExternalID)
	invitationIDPtr := parseUUIDPtr(data.InvitationID)

	now := time.Now().UTC()
	_, err = rm.db.ExecContext(ctx,
		`INSERT INTO users (id, tenant_id, email, name, role, status, external_id, invitation_id, created_at, last_login_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		userID, tenantID.Value(), data.Email, namePtr, data.Role, data.Status, externalIDPtr, invitationIDPtr, createdAt, now,
	)
	return err
}

func toStringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func parseUUIDPtr(s string) *uuid.UUID {
	if s == "" {
		return nil
	}
	if parsed, err := uuid.Parse(s); err == nil {
		return &parsed
	}
	return nil
}

func (rm *UserReadModel) UpdateLastLogin(ctx context.Context, id uuid.UUID, lastLoginAt time.Time) error {
	return rm.updateField(ctx, id.String(), "last_login_at", lastLoginAt)
}

func (rm *UserReadModel) GetByIDString(ctx context.Context, id string) (*UserDTO, error) {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return nil, nil
	}
	return rm.GetByID(ctx, parsedID)
}

func (rm *UserReadModel) CountActiveAdmins(ctx context.Context) (int, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return 0, err
	}

	var count int
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM users WHERE tenant_id = $1 AND role = 'admin' AND status = 'active'`,
			tenantID.Value(),
		).Scan(&count)
	})

	return count, err
}

func (rm *UserReadModel) IsLastActiveAdmin(ctx context.Context, userID string) (bool, error) {
	user, err := rm.GetByIDString(ctx, userID)
	if err != nil {
		return false, err
	}

	if !rm.isActiveAdmin(user) {
		return false, nil
	}

	count, err := rm.CountActiveAdmins(ctx)
	if err != nil {
		return false, err
	}

	return count == 1, nil
}

func (rm *UserReadModel) isActiveAdmin(user *UserDTO) bool {
	return user != nil && user.Role == "admin" && user.Status == "active"
}

func (rm *UserReadModel) GetAllPaginated(ctx context.Context, filter UserPaginationFilter) ([]UserDTO, bool, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, false, err
	}

	var users []UserDTO
	var hasMore bool

	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		query, args := rm.buildPaginatedQuery(tenantID.Value(), filter)

		rows, err := tx.QueryContext(ctx, query, args...)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var dto UserDTO
			if err := rows.Scan(&dto.ID, &dto.Email, &dto.Name, &dto.Role, &dto.Status,
				&dto.ExternalID, &dto.InvitationID, &dto.CreatedAt, &dto.LastLoginAt); err != nil {
				return err
			}
			users = append(users, dto)
		}

		return rows.Err()
	})

	if err != nil {
		return nil, false, err
	}

	if len(users) > filter.Limit {
		hasMore = true
		users = users[:filter.Limit]
	}

	return users, hasMore, nil
}

func (rm *UserReadModel) buildPaginatedQuery(tenantID string, filter UserPaginationFilter) (string, []interface{}) {
	baseQuery := `SELECT id, email, name, role, status, external_id, invitation_id, created_at, last_login_at
		FROM users WHERE tenant_id = $1`
	args := []interface{}{tenantID}
	argIndex := 2

	if filter.StatusFilter != "" {
		baseQuery += " AND status = $" + strconv.Itoa(argIndex)
		args = append(args, filter.StatusFilter)
		argIndex++
	}

	if filter.RoleFilter != "" {
		baseQuery += " AND role = $" + strconv.Itoa(argIndex)
		args = append(args, filter.RoleFilter)
		argIndex++
	}

	if filter.AfterCursor != "" && filter.AfterTimestamp > 0 {
		baseQuery += " AND (created_at < to_timestamp($" + strconv.Itoa(argIndex) + ") OR (created_at = to_timestamp($" + strconv.Itoa(argIndex) + ") AND id < $" + strconv.Itoa(argIndex+1) + "))"
		args = append(args, filter.AfterTimestamp, filter.AfterCursor)
		argIndex += 2
	}

	baseQuery += " ORDER BY created_at DESC, id DESC LIMIT $" + strconv.Itoa(argIndex)
	args = append(args, filter.Limit+1)

	return baseQuery, args
}

func (rm *UserReadModel) UpdateRole(ctx context.Context, id string, role string) error {
	return rm.updateField(ctx, id, "role", role)
}

func (rm *UserReadModel) UpdateStatus(ctx context.Context, id string, status string) error {
	return rm.updateField(ctx, id, "status", status)
}

func (rm *UserReadModel) updateField(ctx context.Context, id string, field string, value interface{}) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		`UPDATE users SET `+field+` = $1, updated_at = CURRENT_TIMESTAMP WHERE tenant_id = $2 AND id = $3`,
		value, tenantID.Value(), id,
	)
	return err
}
