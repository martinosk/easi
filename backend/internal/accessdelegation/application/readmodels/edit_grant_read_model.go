package readmodels

import (
	"context"
	"database/sql"
	"time"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/types"
)

type EditGrantDTO struct {
	ID                string      `json:"id"`
	GrantorID         string      `json:"grantorId"`
	GrantorEmail      string      `json:"grantorEmail"`
	GranteeEmail      string      `json:"granteeEmail"`
	ArtifactType      string      `json:"artifactType"`
	ArtifactID        string      `json:"artifactId"`
	ArtifactName      string      `json:"artifactName,omitempty"`
	Scope             string      `json:"scope"`
	Status            string      `json:"status"`
	Reason            *string     `json:"reason,omitempty"`
	InvitationCreated bool        `json:"invitationCreated,omitempty"`
	CreatedAt         time.Time   `json:"createdAt"`
	ExpiresAt         time.Time   `json:"expiresAt"`
	RevokedAt         *time.Time  `json:"revokedAt,omitempty"`
	Links             types.Links `json:"_links,omitempty"`
}

type EditGrantStatusUpdate struct {
	ID        string
	Status    string
	RevokedAt *time.Time
}

type EditGrantReadModel struct {
	db *database.TenantAwareDB
}

func NewEditGrantReadModel(db *database.TenantAwareDB) *EditGrantReadModel {
	return &EditGrantReadModel{db: db}
}

func (rm *EditGrantReadModel) getTenantID(ctx context.Context) (string, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return "", err
	}
	return tenantID.Value(), nil
}

func (rm *EditGrantReadModel) Insert(ctx context.Context, dto EditGrantDTO) error {
	tenantID, err := rm.getTenantID(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		`INSERT INTO accessdelegation.edit_grants (id, tenant_id, grantor_id, grantor_email, grantee_email, artifact_type, artifact_id, scope, status, reason, created_at, expires_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		dto.ID, tenantID, dto.GrantorID, dto.GrantorEmail, dto.GranteeEmail,
		dto.ArtifactType, dto.ArtifactID, dto.Scope, dto.Status, dto.Reason,
		dto.CreatedAt, dto.ExpiresAt,
	)
	return err
}

func (rm *EditGrantReadModel) UpdateStatus(ctx context.Context, update EditGrantStatusUpdate) error {
	tenantID, err := rm.getTenantID(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		`UPDATE accessdelegation.edit_grants SET status = $1, revoked_at = $2 WHERE tenant_id = $3 AND id = $4`,
		update.Status, update.RevokedAt, tenantID, update.ID,
	)
	return err
}

func (rm *EditGrantReadModel) GetByID(ctx context.Context, id string) (*EditGrantDTO, error) {
	return rm.queryOneWithTenant(ctx,
		`SELECT id, grantor_id, grantor_email, grantee_email, artifact_type, artifact_id, scope, status, reason, created_at, expires_at, revoked_at
		 FROM accessdelegation.edit_grants WHERE tenant_id = $1 AND id = $2`, id)
}

func (rm *EditGrantReadModel) GetByGrantorID(ctx context.Context, grantorID string) ([]EditGrantDTO, error) {
	return rm.queryManyWithTenant(ctx,
		`SELECT id, grantor_id, grantor_email, grantee_email, artifact_type, artifact_id, scope, status, reason, created_at, expires_at, revoked_at
		 FROM accessdelegation.edit_grants WHERE tenant_id = $1 AND grantor_id = $2
		 ORDER BY created_at DESC`, grantorID)
}

func (rm *EditGrantReadModel) GetByGranteeEmail(ctx context.Context, email string) ([]EditGrantDTO, error) {
	return rm.queryManyWithTenant(ctx,
		`SELECT id, grantor_id, grantor_email, grantee_email, artifact_type, artifact_id, scope, status, reason, created_at, expires_at, revoked_at
		 FROM accessdelegation.edit_grants WHERE tenant_id = $1 AND grantee_email = $2 AND status = 'active' AND expires_at > NOW()
		 ORDER BY created_at DESC`, email)
}

func (rm *EditGrantReadModel) GetActiveForArtifact(ctx context.Context, artifactType, artifactID string) ([]EditGrantDTO, error) {
	return rm.queryManyWithTenant(ctx,
		`SELECT id, grantor_id, grantor_email, grantee_email, artifact_type, artifact_id, scope, status, reason, created_at, expires_at, revoked_at
		 FROM accessdelegation.edit_grants WHERE tenant_id = $1 AND artifact_type = $2 AND artifact_id = $3 AND status = 'active' AND expires_at > NOW()
		 ORDER BY created_at DESC`, artifactType, artifactID)
}

func (rm *EditGrantReadModel) HasActiveGrant(ctx context.Context, granteeEmail, artifactType, artifactID string) (bool, error) {
	tenantID, err := rm.getTenantID(ctx)
	if err != nil {
		return false, err
	}

	var exists bool
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx,
			`SELECT EXISTS(SELECT 1 FROM accessdelegation.edit_grants WHERE tenant_id = $1 AND grantee_email = $2 AND artifact_type = $3 AND artifact_id = $4 AND status = 'active' AND expires_at > NOW())`,
			tenantID, granteeEmail, artifactType, artifactID,
		).Scan(&exists)
	})

	return exists, err
}

func (rm *EditGrantReadModel) GetGrantedArtifactIDs(ctx context.Context, granteeEmail, artifactType string) (map[string]bool, error) {
	ids, err := rm.queryStringColumn(ctx,
		`SELECT artifact_id FROM accessdelegation.edit_grants WHERE tenant_id = $1 AND grantee_email = $2 AND artifact_type = $3 AND status = 'active' AND expires_at > NOW()`,
		granteeEmail, artifactType,
	)
	if err != nil {
		return nil, err
	}

	result := make(map[string]bool, len(ids))
	for _, id := range ids {
		result[id] = true
	}
	return result, nil
}

func (rm *EditGrantReadModel) ResolveEditGrants(ctx context.Context, email string) (map[string]map[string]bool, error) {
	tenantID, err := rm.getTenantID(ctx)
	if err != nil {
		return nil, err
	}

	result := make(map[string]map[string]bool)
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx,
			`SELECT artifact_type, artifact_id FROM accessdelegation.edit_grants WHERE tenant_id = $1 AND grantee_email = $2 AND status = 'active' AND expires_at > NOW()`,
			tenantID, email,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var artifactType, artifactID string
			if err := rows.Scan(&artifactType, &artifactID); err != nil {
				return err
			}
			if result[artifactType] == nil {
				result[artifactType] = make(map[string]bool)
			}
			result[artifactType][artifactID] = true
		}
		return rows.Err()
	})

	return result, err
}

func (rm *EditGrantReadModel) GetActiveGrantIDsForArtifact(ctx context.Context, artifactType, artifactID string) ([]string, error) {
	return rm.queryStringColumn(ctx,
		`SELECT id FROM accessdelegation.edit_grants WHERE tenant_id = $1 AND artifact_type = $2 AND artifact_id = $3 AND status = 'active'`,
		artifactType, artifactID,
	)
}

func (rm *EditGrantReadModel) queryStringColumn(ctx context.Context, query string, extraArgs ...interface{}) ([]string, error) {
	tenantID, err := rm.getTenantID(ctx)
	if err != nil {
		return nil, err
	}

	args := append([]interface{}{tenantID}, extraArgs...)
	var results []string
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, query, args...)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var val string
			if err := rows.Scan(&val); err != nil {
				return err
			}
			results = append(results, val)
		}
		return rows.Err()
	})

	return results, err
}

func (rm *EditGrantReadModel) queryOneWithTenant(ctx context.Context, query string, extraArgs ...interface{}) (*EditGrantDTO, error) {
	tenantID, err := rm.getTenantID(ctx)
	if err != nil {
		return nil, err
	}
	args := append([]interface{}{tenantID}, extraArgs...)
	return rm.queryOne(ctx, query, args...)
}

func (rm *EditGrantReadModel) queryOne(ctx context.Context, query string, args ...interface{}) (*EditGrantDTO, error) {
	var dto EditGrantDTO
	var notFound bool

	err := rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		err := tx.QueryRowContext(ctx, query, args...).Scan(
			&dto.ID, &dto.GrantorID, &dto.GrantorEmail, &dto.GranteeEmail,
			&dto.ArtifactType, &dto.ArtifactID, &dto.Scope, &dto.Status, &dto.Reason,
			&dto.CreatedAt, &dto.ExpiresAt, &dto.RevokedAt,
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

func (rm *EditGrantReadModel) queryManyWithTenant(ctx context.Context, query string, extraArgs ...interface{}) ([]EditGrantDTO, error) {
	tenantID, err := rm.getTenantID(ctx)
	if err != nil {
		return nil, err
	}
	args := append([]interface{}{tenantID}, extraArgs...)
	return rm.queryMany(ctx, query, args...)
}

func (rm *EditGrantReadModel) queryMany(ctx context.Context, query string, args ...interface{}) ([]EditGrantDTO, error) {
	var grants []EditGrantDTO

	err := rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, query, args...)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var dto EditGrantDTO
			if err := rows.Scan(
				&dto.ID, &dto.GrantorID, &dto.GrantorEmail, &dto.GranteeEmail,
				&dto.ArtifactType, &dto.ArtifactID, &dto.Scope, &dto.Status, &dto.Reason,
				&dto.CreatedAt, &dto.ExpiresAt, &dto.RevokedAt,
			); err != nil {
				return err
			}
			grants = append(grants, dto)
		}
		return rows.Err()
	})

	return grants, err
}
