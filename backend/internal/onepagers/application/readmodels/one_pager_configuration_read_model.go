package readmodels

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
)

type ConfigurationRecord struct {
	ID          string
	TenantID    string
	SubjectType string
	Document    ConfigurationDocument
	Version     int
	CreatedAt   time.Time
	ModifiedAt  time.Time
	ModifiedBy  string
}

type UpdateParams struct {
	ID         string
	Document   ConfigurationDocument
	Version    int
	ModifiedAt time.Time
	ModifiedBy string
}

type OnePagerConfigurationReadModel struct {
	db *database.TenantAwareDB
}

func NewOnePagerConfigurationReadModel(db *database.TenantAwareDB) *OnePagerConfigurationReadModel {
	return &OnePagerConfigurationReadModel{db: db}
}

func (rm *OnePagerConfigurationReadModel) Insert(ctx context.Context, record ConfigurationRecord) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	documentJSON, err := json.Marshal(record.Document)
	if err != nil {
		return fmt.Errorf("marshal one-pager configuration document %s: %w", record.ID, err)
	}

	_, err = rm.db.ExecContext(ctx,
		`INSERT INTO onepagers.one_pager_configurations
		(id, tenant_id, subject_type, configuration, version, created_at, modified_at, modified_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		record.ID, tenantID.Value(), record.SubjectType, documentJSON,
		record.Version, record.CreatedAt, record.ModifiedAt, record.ModifiedBy,
	)
	if err != nil {
		return fmt.Errorf("insert one-pager configuration %s: %w", record.ID, err)
	}
	return nil
}

func (rm *OnePagerConfigurationReadModel) Update(ctx context.Context, params UpdateParams) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	documentJSON, err := json.Marshal(params.Document)
	if err != nil {
		return fmt.Errorf("marshal one-pager configuration document %s: %w", params.ID, err)
	}

	_, err = rm.db.ExecContext(ctx,
		`UPDATE onepagers.one_pager_configurations
		SET configuration = $1, version = $2, modified_at = $3, modified_by = $4
		WHERE tenant_id = $5 AND id = $6`,
		documentJSON, params.Version, params.ModifiedAt, params.ModifiedBy, tenantID.Value(), params.ID,
	)
	if err != nil {
		return fmt.Errorf("update one-pager configuration %s: %w", params.ID, err)
	}
	return nil
}

func (rm *OnePagerConfigurationReadModel) GetByID(ctx context.Context, id string) (*ConfigurationRecord, error) {
	return rm.getByQuery(ctx,
		`SELECT id, tenant_id, subject_type, configuration, version, created_at, modified_at, modified_by
		FROM onepagers.one_pager_configurations
		WHERE tenant_id = $1 AND id = $2`,
		id,
	)
}

func (rm *OnePagerConfigurationReadModel) GetBySubjectType(ctx context.Context, subjectType string) (*ConfigurationRecord, error) {
	return rm.getByQuery(ctx,
		`SELECT id, tenant_id, subject_type, configuration, version, created_at, modified_at, modified_by
		FROM onepagers.one_pager_configurations
		WHERE tenant_id = $1 AND subject_type = $2`,
		subjectType,
	)
}

func (rm *OnePagerConfigurationReadModel) ConfigurationExists(ctx context.Context, subjectType string) (bool, error) {
	record, err := rm.GetBySubjectType(ctx, subjectType)
	if err != nil {
		return false, err
	}
	return record != nil, nil
}

func (rm *OnePagerConfigurationReadModel) getByQuery(ctx context.Context, query, arg string) (*ConfigurationRecord, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var record ConfigurationRecord
	var documentJSON []byte
	var notFound bool

	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		err := tx.QueryRowContext(ctx, query, tenantID.Value(), arg).Scan(
			&record.ID, &record.TenantID, &record.SubjectType, &documentJSON,
			&record.Version, &record.CreatedAt, &record.ModifiedAt, &record.ModifiedBy,
		)
		if err == sql.ErrNoRows {
			notFound = true
			return nil
		}
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("query one-pager configuration: %w", err)
	}
	if notFound {
		return nil, nil
	}

	if err := json.Unmarshal(documentJSON, &record.Document); err != nil {
		return nil, fmt.Errorf("unmarshal one-pager configuration document %s: %w", record.ID, err)
	}
	return &record, nil
}
