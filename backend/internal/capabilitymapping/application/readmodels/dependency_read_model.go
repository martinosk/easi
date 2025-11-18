package readmodels

import (
	"context"
	"database/sql"
	"time"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
)

type DependencyDTO struct {
	ID                 string            `json:"id"`
	SourceCapabilityID string            `json:"sourceCapabilityId"`
	TargetCapabilityID string            `json:"targetCapabilityId"`
	DependencyType     string            `json:"dependencyType"`
	Description        string            `json:"description,omitempty"`
	CreatedAt          time.Time         `json:"createdAt"`
	Links              map[string]string `json:"_links,omitempty"`
}

type DependencyReadModel struct {
	db *database.TenantAwareDB
}

func NewDependencyReadModel(db *database.TenantAwareDB) *DependencyReadModel {
	return &DependencyReadModel{db: db}
}

func (rm *DependencyReadModel) Insert(ctx context.Context, dto DependencyDTO) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"INSERT INTO capability_dependencies (id, tenant_id, source_capability_id, target_capability_id, dependency_type, description, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		dto.ID, tenantID.Value(), dto.SourceCapabilityID, dto.TargetCapabilityID, dto.DependencyType, dto.Description, dto.CreatedAt,
	)
	return err
}

func (rm *DependencyReadModel) Delete(ctx context.Context, id string) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"DELETE FROM capability_dependencies WHERE tenant_id = $1 AND id = $2",
		tenantID.Value(), id,
	)
	return err
}

func (rm *DependencyReadModel) GetAll(ctx context.Context) ([]DependencyDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var dependencies []DependencyDTO
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx,
			"SELECT id, source_capability_id, target_capability_id, dependency_type, description, created_at FROM capability_dependencies WHERE tenant_id = $1 ORDER BY created_at DESC",
			tenantID.Value(),
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var dto DependencyDTO
			if err := rows.Scan(&dto.ID, &dto.SourceCapabilityID, &dto.TargetCapabilityID, &dto.DependencyType, &dto.Description, &dto.CreatedAt); err != nil {
				return err
			}
			dependencies = append(dependencies, dto)
		}

		return rows.Err()
	})

	return dependencies, err
}

func (rm *DependencyReadModel) GetOutgoing(ctx context.Context, capabilityID string) ([]DependencyDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var dependencies []DependencyDTO
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx,
			"SELECT id, source_capability_id, target_capability_id, dependency_type, description, created_at FROM capability_dependencies WHERE tenant_id = $1 AND source_capability_id = $2 ORDER BY created_at DESC",
			tenantID.Value(), capabilityID,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var dto DependencyDTO
			if err := rows.Scan(&dto.ID, &dto.SourceCapabilityID, &dto.TargetCapabilityID, &dto.DependencyType, &dto.Description, &dto.CreatedAt); err != nil {
				return err
			}
			dependencies = append(dependencies, dto)
		}

		return rows.Err()
	})

	return dependencies, err
}

func (rm *DependencyReadModel) GetIncoming(ctx context.Context, capabilityID string) ([]DependencyDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var dependencies []DependencyDTO
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx,
			"SELECT id, source_capability_id, target_capability_id, dependency_type, description, created_at FROM capability_dependencies WHERE tenant_id = $1 AND target_capability_id = $2 ORDER BY created_at DESC",
			tenantID.Value(), capabilityID,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var dto DependencyDTO
			if err := rows.Scan(&dto.ID, &dto.SourceCapabilityID, &dto.TargetCapabilityID, &dto.DependencyType, &dto.Description, &dto.CreatedAt); err != nil {
				return err
			}
			dependencies = append(dependencies, dto)
		}

		return rows.Err()
	})

	return dependencies, err
}

func (rm *DependencyReadModel) GetByID(ctx context.Context, id string) (*DependencyDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var dto DependencyDTO
	var notFound bool

	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		err := tx.QueryRowContext(ctx,
			"SELECT id, source_capability_id, target_capability_id, dependency_type, description, created_at FROM capability_dependencies WHERE tenant_id = $1 AND id = $2",
			tenantID.Value(), id,
		).Scan(&dto.ID, &dto.SourceCapabilityID, &dto.TargetCapabilityID, &dto.DependencyType, &dto.Description, &dto.CreatedAt)

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
