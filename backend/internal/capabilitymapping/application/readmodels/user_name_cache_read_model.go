package readmodels

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"

	"github.com/google/uuid"
)

type UserNameCacheReadModel struct {
	db *database.TenantAwareDB
}

func NewUserNameCacheReadModel(db *database.TenantAwareDB) *UserNameCacheReadModel {
	return &UserNameCacheReadModel{db: db}
}

func (rm *UserNameCacheReadModel) Upsert(ctx context.Context, id, name, email string) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx, `
		INSERT INTO capabilitymapping.user_names (tenant_id, user_id, name, email)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (tenant_id, user_id) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email
	`, tenantID.Value(), id, name, email)
	if err != nil {
		return fmt.Errorf("upsert user name cache entry for user %s: %w", id, err)
	}
	return nil
}

func (rm *UserNameCacheReadModel) ResolveEAOwner(ctx context.Context, value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", valueobjects.ErrEAOwnerNotUser
	}
	if _, err := uuid.Parse(trimmed); err == nil {
		return trimmed, nil
	}

	ids, err := rm.findUserIDsByNameOrEmail(ctx, trimmed)
	if err != nil {
		return "", fmt.Errorf("resolve EA owner %q: %w", trimmed, err)
	}

	switch len(ids) {
	case 0:
		return "", valueobjects.ErrEAOwnerNotUser
	case 1:
		return ids[0], nil
	default:
		return "", valueobjects.ErrEAOwnerAmbiguous
	}
}

func (rm *UserNameCacheReadModel) findUserIDsByNameOrEmail(ctx context.Context, value string) ([]string, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var ids []string
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx,
			`SELECT user_id FROM capabilitymapping.user_names
			 WHERE tenant_id = $1 AND (LOWER(name) = LOWER($2) OR LOWER(email) = LOWER($2))
			 LIMIT 2`,
			tenantID.Value(), value,
		)
		if err != nil {
			return err
		}
		defer func() { _ = rows.Close() }()

		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				return err
			}
			ids = append(ids, id)
		}
		return rows.Err()
	})
	return ids, err
}
