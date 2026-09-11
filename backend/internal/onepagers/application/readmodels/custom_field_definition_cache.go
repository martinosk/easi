package readmodels

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
)

type OptionRecord struct {
	ID     string `json:"id"`
	Label  string `json:"label"`
	Active bool   `json:"active"`
}

type CustomFieldRecord struct {
	ID       string         `json:"id"`
	Name     string         `json:"name"`
	Type     string         `json:"type"`
	HelpText string         `json:"helpText"`
	Active   bool           `json:"active"`
	Options  []OptionRecord `json:"options,omitempty"`
	Min      *float64       `json:"min,omitempty"`
	Max      *float64       `json:"max,omitempty"`
}

type CustomFieldDefinitions []CustomFieldRecord

func (d CustomFieldDefinitions) ByID(fieldID string) (CustomFieldRecord, bool) {
	for _, field := range d {
		if field.ID == fieldID {
			return field, true
		}
	}
	return CustomFieldRecord{}, false
}

type SubjectDefinition struct {
	SubjectType string
	Field       CustomFieldRecord
}

type CustomFieldDefinitionCacheReadModel struct {
	db *database.TenantAwareDB
}

func NewCustomFieldDefinitionCacheReadModel(db *database.TenantAwareDB) *CustomFieldDefinitionCacheReadModel {
	return &CustomFieldDefinitionCacheReadModel{db: db}
}

func (rm *CustomFieldDefinitionCacheReadModel) Save(ctx context.Context, definition SubjectDefinition) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}
	field := definition.Field
	encoded, err := json.Marshal(field)
	if err != nil {
		return fmt.Errorf("encode custom field definition %s: %w", field.ID, err)
	}
	_, err = rm.db.ExecContext(ctx,
		`INSERT INTO onepagers.custom_field_definition_cache (tenant_id, field_id, subject_type, definition, pending_transfer)
		VALUES ($1, $2, $3, $4::jsonb, FALSE)
		ON CONFLICT (tenant_id, field_id) DO UPDATE
		SET subject_type = EXCLUDED.subject_type, definition = EXCLUDED.definition, pending_transfer = FALSE`,
		tenantID.Value(), field.ID, definition.SubjectType, encoded,
	)
	if err != nil {
		return fmt.Errorf("cache custom field definition %s: %w", field.ID, err)
	}
	return nil
}

func (rm *CustomFieldDefinitionCacheReadModel) Get(ctx context.Context, fieldID string) (*CustomFieldRecord, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}
	var raw []byte
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		scanErr := tx.QueryRowContext(ctx,
			`SELECT definition FROM onepagers.custom_field_definition_cache WHERE tenant_id = $1 AND field_id = $2`,
			tenantID.Value(), fieldID,
		).Scan(&raw)
		if scanErr == sql.ErrNoRows {
			return nil
		}
		return scanErr
	})
	if err != nil {
		return nil, fmt.Errorf("read cached custom field definition %s: %w", fieldID, err)
	}
	if raw == nil {
		return nil, nil
	}
	var field CustomFieldRecord
	if err := json.Unmarshal(raw, &field); err != nil {
		return nil, fmt.Errorf("decode cached custom field definition %s: %w", fieldID, err)
	}
	return &field, nil
}

func (rm *CustomFieldDefinitionCacheReadModel) ForSubjectType(ctx context.Context, subjectType string) (CustomFieldDefinitions, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}
	var definitions CustomFieldDefinitions
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx,
			`SELECT definition FROM onepagers.custom_field_definition_cache
			WHERE tenant_id = $1 AND subject_type = $2 ORDER BY field_id`,
			tenantID.Value(), subjectType,
		)
		if err != nil {
			return err
		}
		defer func() { _ = rows.Close() }()
		for rows.Next() {
			var raw []byte
			if err := rows.Scan(&raw); err != nil {
				return err
			}
			var field CustomFieldRecord
			if err := json.Unmarshal(raw, &field); err != nil {
				return err
			}
			definitions = append(definitions, field)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("read cached custom field definitions for %s: %w", subjectType, err)
	}
	return definitions, nil
}

func (rm *CustomFieldDefinitionCacheReadModel) PendingTransfers(ctx context.Context) ([]SubjectDefinition, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}
	var pending []SubjectDefinition
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx,
			`SELECT subject_type, definition FROM onepagers.custom_field_definition_cache
			WHERE tenant_id = $1 AND pending_transfer ORDER BY subject_type, field_id`,
			tenantID.Value(),
		)
		if err != nil {
			return err
		}
		defer func() { _ = rows.Close() }()
		for rows.Next() {
			var item SubjectDefinition
			var raw []byte
			if err := rows.Scan(&item.SubjectType, &raw); err != nil {
				return err
			}
			if err := json.Unmarshal(raw, &item.Field); err != nil {
				return err
			}
			pending = append(pending, item)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("read custom field definitions pending transfer: %w", err)
	}
	return pending, nil
}

func (rm *CustomFieldDefinitionCacheReadModel) MarkTransferred(ctx context.Context, fieldID string) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}
	_, err = rm.db.ExecContext(ctx,
		`UPDATE onepagers.custom_field_definition_cache SET pending_transfer = FALSE WHERE tenant_id = $1 AND field_id = $2`,
		tenantID.Value(), fieldID,
	)
	if err != nil {
		return fmt.Errorf("mark custom field definition %s transferred: %w", fieldID, err)
	}
	return nil
}
