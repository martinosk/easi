package readmodels

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
)

type MaturitySectionDTO struct {
	Order    int    `json:"order"`
	Name     string `json:"name"`
	MinValue int    `json:"minValue"`
	MaxValue int    `json:"maxValue"`
}

type MetaModelConfigurationDTO struct {
	ID         string               `json:"id"`
	TenantID   string               `json:"tenantId"`
	Sections   []MaturitySectionDTO `json:"sections"`
	Version    int                  `json:"version"`
	CreatedAt  time.Time            `json:"createdAt"`
	ModifiedAt time.Time            `json:"modifiedAt"`
	ModifiedBy string               `json:"modifiedBy"`
	Links      map[string]string    `json:"_links,omitempty"`
}

type MetaModelConfigurationReadModel struct {
	db *database.TenantAwareDB
}

type UpdateParams struct {
	ID         string
	Sections   []MaturitySectionDTO
	Version    int
	ModifiedAt time.Time
	ModifiedBy string
}

func NewMetaModelConfigurationReadModel(db *database.TenantAwareDB) *MetaModelConfigurationReadModel {
	return &MetaModelConfigurationReadModel{db: db}
}

func (rm *MetaModelConfigurationReadModel) Insert(ctx context.Context, dto MetaModelConfigurationDTO) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	sectionsJSON, err := json.Marshal(dto.Sections)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		`INSERT INTO meta_model_configurations
		(id, tenant_id, sections, version, created_at, modified_at, modified_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		dto.ID, tenantID.Value(), sectionsJSON, dto.Version, dto.CreatedAt, dto.ModifiedAt, dto.ModifiedBy,
	)
	return err
}

func (rm *MetaModelConfigurationReadModel) Update(ctx context.Context, params UpdateParams) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	sectionsJSON, err := json.Marshal(params.Sections)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		`UPDATE meta_model_configurations
		SET sections = $1, version = $2, modified_at = $3, modified_by = $4
		WHERE tenant_id = $5 AND id = $6`,
		sectionsJSON, params.Version, params.ModifiedAt, params.ModifiedBy, tenantID.Value(), params.ID,
	)
	return err
}

func (rm *MetaModelConfigurationReadModel) GetByID(ctx context.Context, id string) (*MetaModelConfigurationDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var dto MetaModelConfigurationDTO
	var sectionsJSON []byte
	var notFound bool

	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		err := tx.QueryRowContext(ctx,
			`SELECT id, tenant_id, sections, version, created_at, modified_at, modified_by
			FROM meta_model_configurations
			WHERE tenant_id = $1 AND id = $2`,
			tenantID.Value(), id,
		).Scan(&dto.ID, &dto.TenantID, &sectionsJSON, &dto.Version, &dto.CreatedAt, &dto.ModifiedAt, &dto.ModifiedBy)

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

	if err := json.Unmarshal(sectionsJSON, &dto.Sections); err != nil {
		return nil, err
	}

	return &dto, nil
}

func (rm *MetaModelConfigurationReadModel) GetByTenantID(ctx context.Context) (*MetaModelConfigurationDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var dto MetaModelConfigurationDTO
	var sectionsJSON []byte
	var notFound bool

	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		err := tx.QueryRowContext(ctx,
			`SELECT id, tenant_id, sections, version, created_at, modified_at, modified_by
			FROM meta_model_configurations
			WHERE tenant_id = $1`,
			tenantID.Value(),
		).Scan(&dto.ID, &dto.TenantID, &sectionsJSON, &dto.Version, &dto.CreatedAt, &dto.ModifiedAt, &dto.ModifiedBy)

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

	if err := json.Unmarshal(sectionsJSON, &dto.Sections); err != nil {
		return nil, err
	}

	return &dto, nil
}

func (rm *MetaModelConfigurationReadModel) GetConfigIDByTenantID(ctx context.Context, tenantID string) (string, error) {
	var id string
	var notFound bool

	err := rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		err := tx.QueryRowContext(ctx,
			`SELECT id FROM meta_model_configurations WHERE tenant_id = $1`,
			tenantID,
		).Scan(&id)

		if err == sql.ErrNoRows {
			notFound = true
			return nil
		}
		return err
	})

	if err != nil {
		return "", err
	}
	if notFound {
		return "", nil
	}

	return id, nil
}
