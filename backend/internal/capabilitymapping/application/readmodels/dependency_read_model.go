package readmodels

import (
	"context"
	"database/sql"
	"time"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/types"
)

type DependencyDTO struct {
	ID                 string      `json:"id"`
	SourceCapabilityID string      `json:"sourceCapabilityId"`
	TargetCapabilityID string      `json:"targetCapabilityId"`
	DependencyType     string      `json:"dependencyType"`
	Description        string      `json:"description,omitempty"`
	CreatedAt          time.Time   `json:"createdAt"`
	Links              types.Links `json:"_links,omitempty"`
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
		"INSERT INTO capabilitymapping.capability_dependencies (id, tenant_id, source_capability_id, target_capability_id, dependency_type, description, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7)",
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
		"DELETE FROM capabilitymapping.capability_dependencies WHERE tenant_id = $1 AND id = $2",
		tenantID.Value(), id,
	)
	return err
}

func (rm *DependencyReadModel) GetAll(ctx context.Context) ([]DependencyDTO, error) {
	return rm.queryDependencies(ctx, "SELECT id, source_capability_id, target_capability_id, dependency_type, description, created_at FROM capabilitymapping.capability_dependencies WHERE tenant_id = $1 ORDER BY created_at DESC")
}

func (rm *DependencyReadModel) GetOutgoing(ctx context.Context, capabilityID string) ([]DependencyDTO, error) {
	return rm.queryDependenciesWithParam(ctx, "SELECT id, source_capability_id, target_capability_id, dependency_type, description, created_at FROM capabilitymapping.capability_dependencies WHERE tenant_id = $1 AND source_capability_id = $2 ORDER BY created_at DESC", capabilityID)
}

func (rm *DependencyReadModel) GetIncoming(ctx context.Context, capabilityID string) ([]DependencyDTO, error) {
	return rm.queryDependenciesWithParam(ctx, "SELECT id, source_capability_id, target_capability_id, dependency_type, description, created_at FROM capabilitymapping.capability_dependencies WHERE tenant_id = $1 AND target_capability_id = $2 ORDER BY created_at DESC", capabilityID)
}

func (rm *DependencyReadModel) queryDependencies(ctx context.Context, query string) ([]DependencyDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	return rm.executeDependencyQuery(ctx, query, tenantID.Value())
}

func (rm *DependencyReadModel) queryDependenciesWithParam(ctx context.Context, query, param string) ([]DependencyDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	return rm.executeDependencyQuery(ctx, query, tenantID.Value(), param)
}

func (rm *DependencyReadModel) executeDependencyQuery(ctx context.Context, query string, args ...interface{}) ([]DependencyDTO, error) {
	var dependencies []DependencyDTO
	err := rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, query, args...)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			dto, err := rm.scanDependencyRow(rows)
			if err != nil {
				return err
			}
			dependencies = append(dependencies, dto)
		}

		return rows.Err()
	})

	return dependencies, err
}

func (rm *DependencyReadModel) scanDependencyRow(rows *sql.Rows) (DependencyDTO, error) {
	var dto DependencyDTO
	err := rows.Scan(&dto.ID, &dto.SourceCapabilityID, &dto.TargetCapabilityID, &dto.DependencyType, &dto.Description, &dto.CreatedAt)
	return dto, err
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
			"SELECT id, source_capability_id, target_capability_id, dependency_type, description, created_at FROM capabilitymapping.capability_dependencies WHERE tenant_id = $1 AND id = $2",
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
