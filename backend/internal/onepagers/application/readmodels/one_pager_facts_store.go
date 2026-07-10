package readmodels

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/onepagers/domain/valueobjects"
	sharedctx "easi/backend/internal/shared/context"
)

type OnePagerFactsReadModel struct {
	db *database.TenantAwareDB
}

func NewOnePagerFactsReadModel(db *database.TenantAwareDB) *OnePagerFactsReadModel {
	return &OnePagerFactsReadModel{db: db}
}

func (rm *OnePagerFactsReadModel) Upsert(ctx context.Context, record FactRecord) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	valueJSON, err := json.Marshal(record.Value)
	if err != nil {
		return fmt.Errorf("marshal one-pager fact value for field %s: %w", record.FieldID, err)
	}

	_, err = rm.db.ExecContext(ctx,
		`INSERT INTO onepagers.one_pager_facts
		(tenant_id, subject_type, subject_id, field_id, facts_id, value, value_type, display_text, modified_at, modified_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (tenant_id, subject_type, subject_id, field_id) DO UPDATE
		SET facts_id = EXCLUDED.facts_id, value = EXCLUDED.value, value_type = EXCLUDED.value_type,
			display_text = EXCLUDED.display_text, modified_at = EXCLUDED.modified_at, modified_by = EXCLUDED.modified_by`,
		tenantID.Value(), record.SubjectType, record.SubjectID, record.FieldID, record.FactsID,
		valueJSON, record.ValueType, record.DisplayText, record.ModifiedAt, record.ModifiedBy,
	)
	if err != nil {
		return fmt.Errorf("upsert one-pager fact for field %s: %w", record.FieldID, err)
	}
	return nil
}

func (rm *OnePagerFactsReadModel) Clear(ctx context.Context, params ClearFactParams) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		`UPDATE onepagers.one_pager_facts
		SET value = NULL, value_type = NULL, display_text = NULL, modified_at = $1, modified_by = $2
		WHERE tenant_id = $3 AND subject_type = $4 AND subject_id = $5 AND field_id = $6`,
		params.ModifiedAt, params.ModifiedBy, tenantID.Value(), params.SubjectType, params.SubjectID, params.FieldID,
	)
	if err != nil {
		return fmt.Errorf("clear one-pager fact for field %s: %w", params.FieldID, err)
	}
	return nil
}

func (rm *OnePagerFactsReadModel) DeleteForSubject(ctx context.Context, subject SubjectKey) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		`DELETE FROM onepagers.one_pager_facts
		WHERE tenant_id = $1 AND subject_type = $2 AND subject_id = $3`,
		tenantID.Value(), subject.SubjectType, subject.SubjectID,
	)
	if err != nil {
		return fmt.Errorf("delete one-pager facts for %s %s: %w", subject.SubjectType, subject.SubjectID, err)
	}
	return nil
}

func (rm *OnePagerFactsReadModel) FactsIDForSubject(ctx context.Context, subject SubjectKey) (string, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return "", err
	}

	var factsID string
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		err := tx.QueryRowContext(ctx,
			`SELECT facts_id FROM onepagers.one_pager_facts
			WHERE tenant_id = $1 AND subject_type = $2 AND subject_id = $3
			LIMIT 1`,
			tenantID.Value(), subject.SubjectType, subject.SubjectID,
		).Scan(&factsID)
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	})
	if err != nil {
		return "", fmt.Errorf("query one-pager facts ID for %s %s: %w", subject.SubjectType, subject.SubjectID, err)
	}
	return factsID, nil
}

func (rm *OnePagerFactsReadModel) GetForSubject(ctx context.Context, subject SubjectKey) ([]FactRecord, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var records []FactRecord
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx,
			`SELECT facts_id, tenant_id, subject_type, subject_id, field_id, value, value_type, display_text, modified_at, modified_by
			FROM onepagers.one_pager_facts
			WHERE tenant_id = $1 AND subject_type = $2 AND subject_id = $3 AND value IS NOT NULL
			ORDER BY field_id`,
			tenantID.Value(), subject.SubjectType, subject.SubjectID,
		)
		if err != nil {
			return err
		}
		defer func() { _ = rows.Close() }()

		records, err = scanFactRecords(rows)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("query one-pager facts for %s %s: %w", subject.SubjectType, subject.SubjectID, err)
	}
	return records, nil
}

func scanFactRecords(rows *sql.Rows) ([]FactRecord, error) {
	var records []FactRecord
	for rows.Next() {
		var record FactRecord
		var valueJSON []byte
		if err := rows.Scan(
			&record.FactsID, &record.TenantID, &record.SubjectType, &record.SubjectID, &record.FieldID,
			&valueJSON, &record.ValueType, &record.DisplayText, &record.ModifiedAt, &record.ModifiedBy,
		); err != nil {
			return nil, err
		}
		var envelope valueobjects.ValueEnvelope
		if err := json.Unmarshal(valueJSON, &envelope); err != nil {
			return nil, fmt.Errorf("unmarshal one-pager fact value for field %s: %w", record.FieldID, err)
		}
		record.Value = &envelope
		records = append(records, record)
	}
	return records, rows.Err()
}
