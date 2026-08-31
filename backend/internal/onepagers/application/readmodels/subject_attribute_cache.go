package readmodels

import (
	"context"
	"database/sql"
	"fmt"

	sharedctx "easi/backend/internal/shared/context"

	"github.com/lib/pq"
)

func (rm *OnePagerSubjectIndexReadModel) MergeAttributes(ctx context.Context, subject SubjectKey, attributes SubjectAttributes) error {
	if len(attributes) == 0 {
		return nil
	}
	encoded, err := attributes.encode()
	if err != nil {
		return err
	}
	return rm.exec(ctx,
		fmt.Sprintf("merge cached attributes of %s %s", subject.SubjectType, subject.SubjectID),
		`UPDATE onepagers.one_pager_subject_index
		SET built_in_fields = built_in_fields || $4::jsonb
		WHERE tenant_id = $1 AND subject_type = $2 AND subject_id = $3`,
		subject.SubjectType, subject.SubjectID, encoded,
	)
}

func (rm *OnePagerSubjectIndexReadModel) ApplyExpertChange(ctx context.Context, subject SubjectKey, expert SubjectExpert, added bool) error {
	row, err := rm.AttributeRow(ctx, subject)
	if err != nil || row == nil {
		return err
	}
	current, err := row.Attributes.Experts()
	if err != nil {
		return err
	}

	experts := withoutExpert(current, expert)
	if added {
		experts = append(experts, expert)
	}
	attributes := SubjectAttributes{}
	if err := attributes.Set(ExpertsAttribute, experts); err != nil {
		return err
	}
	return rm.MergeAttributes(ctx, subject, attributes)
}

func withoutExpert(experts []SubjectExpert, removed SubjectExpert) []SubjectExpert {
	kept := make([]SubjectExpert, 0, len(experts))
	for _, expert := range experts {
		if expert != removed {
			kept = append(kept, expert)
		}
	}
	return kept
}

func (rm *OnePagerSubjectIndexReadModel) AttributeRow(ctx context.Context, subject SubjectKey) (*SubjectAttributeRow, error) {
	rows, err := rm.AttributeRows(ctx, subject.SubjectType, []string{subject.SubjectID})
	if err != nil || len(rows) == 0 {
		return nil, err
	}
	return &rows[0], nil
}

func (rm *OnePagerSubjectIndexReadModel) AttributeRows(ctx context.Context, subjectType string, subjectIDs []string) ([]SubjectAttributeRow, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	query := `SELECT subject_id, name, built_in_fields FROM onepagers.one_pager_subject_index
		WHERE tenant_id = $1 AND subject_type = $2`
	args := []any{tenantID.Value(), subjectType}
	if subjectIDs != nil {
		query += " AND subject_id = ANY($3)"
		args = append(args, pq.Array(subjectIDs))
	}

	var rows []SubjectAttributeRow
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		result, queryErr := tx.QueryContext(ctx, query, args...)
		if queryErr != nil {
			return queryErr
		}
		defer func() { _ = result.Close() }()
		rows, queryErr = scanSubjectAttributeRows(result)
		return queryErr
	})
	if err != nil {
		return nil, fmt.Errorf("query cached attributes of %s subjects: %w", subjectType, err)
	}
	return rows, nil
}

func scanSubjectAttributeRows(result *sql.Rows) ([]SubjectAttributeRow, error) {
	rows := []SubjectAttributeRow{}
	for result.Next() {
		var row SubjectAttributeRow
		var raw []byte
		if err := result.Scan(&row.SubjectID, &row.Name, &raw); err != nil {
			return nil, err
		}
		attributes, err := decodeSubjectAttributes(raw)
		if err != nil {
			return nil, err
		}
		row.Attributes = attributes
		rows = append(rows, row)
	}
	return rows, result.Err()
}

func (rm *OnePagerSubjectIndexReadModel) Exists(ctx context.Context, subject SubjectKey) (bool, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return false, err
	}
	var exists bool
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx,
			`SELECT EXISTS(SELECT 1 FROM onepagers.one_pager_subject_index
			WHERE tenant_id = $1 AND subject_type = $2 AND subject_id = $3)`,
			tenantID.Value(), subject.SubjectType, subject.SubjectID,
		).Scan(&exists)
	})
	if err != nil {
		return false, fmt.Errorf("check existence of %s %s: %w", subject.SubjectType, subject.SubjectID, err)
	}
	return exists, nil
}

func (rm *OnePagerSubjectIndexReadModel) CountSubjects(ctx context.Context, subjectType string) (int, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return 0, err
	}
	var count int
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM onepagers.one_pager_subject_index WHERE tenant_id = $1 AND subject_type = $2`,
			tenantID.Value(), subjectType,
		).Scan(&count)
	})
	if err != nil {
		return 0, fmt.Errorf("count %s subjects: %w", subjectType, err)
	}
	return count, nil
}
