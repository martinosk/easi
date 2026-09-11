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

type AttributeOptionRecord struct {
	ID     string `json:"id"`
	Label  string `json:"label"`
	Active bool   `json:"active"`
}

type SubjectAttributeRecord struct {
	ID       string                  `json:"id"`
	Name     string                  `json:"name"`
	Type     string                  `json:"type"`
	HelpText string                  `json:"helpText"`
	Active   bool                    `json:"active"`
	Options  []AttributeOptionRecord `json:"options,omitempty"`
	Min      *float64                `json:"min,omitempty"`
	Max      *float64                `json:"max,omitempty"`
}

type SubjectAttributeSchemaRecord struct {
	ID          string
	TenantID    string
	SubjectType string
	Attributes  []SubjectAttributeRecord
	Version     int
	CreatedAt   time.Time
	ModifiedAt  time.Time
	ModifiedBy  string
}

type UpdateSchemaParams struct {
	ID         string
	Attributes []SubjectAttributeRecord
	Version    int
	ModifiedAt time.Time
	ModifiedBy string
}

type SubjectAttributeSchemaReadModel struct {
	db *database.TenantAwareDB
}

func NewSubjectAttributeSchemaReadModel(db *database.TenantAwareDB) *SubjectAttributeSchemaReadModel {
	return &SubjectAttributeSchemaReadModel{db: db}
}

func (rm *SubjectAttributeSchemaReadModel) Insert(ctx context.Context, record SubjectAttributeSchemaRecord) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}
	attributesJSON, err := json.Marshal(record.Attributes)
	if err != nil {
		return fmt.Errorf("marshal subject attribute schema %s: %w", record.ID, err)
	}
	_, err = rm.db.ExecContext(ctx,
		`INSERT INTO metamodel.subject_attribute_schemas
		(id, tenant_id, subject_type, attributes, version, created_at, modified_at, modified_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		record.ID, tenantID.Value(), record.SubjectType, attributesJSON,
		record.Version, record.CreatedAt, record.ModifiedAt, record.ModifiedBy,
	)
	if err != nil {
		return fmt.Errorf("insert subject attribute schema %s: %w", record.ID, err)
	}
	return nil
}

func (rm *SubjectAttributeSchemaReadModel) Update(ctx context.Context, params UpdateSchemaParams) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}
	attributesJSON, err := json.Marshal(params.Attributes)
	if err != nil {
		return fmt.Errorf("marshal subject attribute schema %s: %w", params.ID, err)
	}
	_, err = rm.db.ExecContext(ctx,
		`UPDATE metamodel.subject_attribute_schemas
		SET attributes = $1, version = $2, modified_at = $3, modified_by = $4
		WHERE tenant_id = $5 AND id = $6`,
		attributesJSON, params.Version, params.ModifiedAt, params.ModifiedBy, tenantID.Value(), params.ID,
	)
	if err != nil {
		return fmt.Errorf("update subject attribute schema %s: %w", params.ID, err)
	}
	return nil
}

const selectSchema = `SELECT id, tenant_id, subject_type, attributes, version, created_at, modified_at, modified_by
	FROM metamodel.subject_attribute_schemas WHERE tenant_id = $1 AND `

func (rm *SubjectAttributeSchemaReadModel) GetByID(ctx context.Context, id string) (*SubjectAttributeSchemaRecord, error) {
	return rm.getByQuery(ctx, selectSchema+"id = $2", id)
}

func (rm *SubjectAttributeSchemaReadModel) GetBySubjectType(ctx context.Context, subjectType string) (*SubjectAttributeSchemaRecord, error) {
	return rm.getByQuery(ctx, selectSchema+"subject_type = $2", subjectType)
}

func (rm *SubjectAttributeSchemaReadModel) getByQuery(ctx context.Context, query, arg string) (*SubjectAttributeSchemaRecord, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}
	var record SubjectAttributeSchemaRecord
	var attributesJSON []byte
	var notFound bool
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		scanErr := tx.QueryRowContext(ctx, query, tenantID.Value(), arg).Scan(
			&record.ID, &record.TenantID, &record.SubjectType, &attributesJSON,
			&record.Version, &record.CreatedAt, &record.ModifiedAt, &record.ModifiedBy,
		)
		if scanErr == sql.ErrNoRows {
			notFound = true
			return nil
		}
		return scanErr
	})
	if err != nil {
		return nil, fmt.Errorf("query subject attribute schema: %w", err)
	}
	if notFound {
		return nil, nil
	}
	if err := json.Unmarshal(attributesJSON, &record.Attributes); err != nil {
		return nil, fmt.Errorf("unmarshal subject attribute schema %s: %w", record.ID, err)
	}
	if record.Attributes == nil {
		record.Attributes = []SubjectAttributeRecord{}
	}
	return &record, nil
}
