package readmodels

import (
	"context"
	"database/sql"
	"fmt"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"

	"github.com/lib/pq"
)

const (
	SortCompleteness = "completeness"
	SortCreator      = "creator"
	SortName         = "name"
	SortCreated      = "created"
	SortUpdated      = "updated"

	OrderAsc  = "asc"
	OrderDesc = "desc"
)

type SubjectIndexQuery struct {
	SubjectTypes []string
	Sort         string
	Order        string
	Limit        int
	After        *SubjectIndexRecord
}

type OnePagerSubjectIndexReadModel struct {
	db *database.TenantAwareDB
}

func NewOnePagerSubjectIndexReadModel(db *database.TenantAwareDB) *OnePagerSubjectIndexReadModel {
	return &OnePagerSubjectIndexReadModel{db: db}
}

func (rm *OnePagerSubjectIndexReadModel) Upsert(ctx context.Context, record SubjectIndexRecord) error {
	return rm.exec(ctx,
		fmt.Sprintf("upsert subject index row for %s %s", record.SubjectType, record.SubjectID),
		`INSERT INTO onepagers.one_pager_subject_index
		(tenant_id, subject_type, subject_id, name, creator_actor_id, creator_email, created_at, last_updated_at, required_count, filled_count)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (tenant_id, subject_type, subject_id) DO UPDATE
		SET name = EXCLUDED.name, creator_actor_id = EXCLUDED.creator_actor_id, creator_email = EXCLUDED.creator_email,
			created_at = EXCLUDED.created_at, last_updated_at = EXCLUDED.last_updated_at,
			required_count = EXCLUDED.required_count, filled_count = EXCLUDED.filled_count`,
		record.SubjectType, record.SubjectID, record.Name,
		record.CreatorActorID, record.CreatorEmail, record.CreatedAt, record.LastUpdatedAt,
		record.RequiredCount, record.FilledCount,
	)
}

func (rm *OnePagerSubjectIndexReadModel) Delete(ctx context.Context, subject SubjectKey) error {
	return rm.exec(ctx,
		fmt.Sprintf("delete subject index row for %s %s", subject.SubjectType, subject.SubjectID),
		`DELETE FROM onepagers.one_pager_subject_index WHERE tenant_id = $1 AND subject_type = $2 AND subject_id = $3`,
		subject.SubjectType, subject.SubjectID,
	)
}

func (rm *OnePagerSubjectIndexReadModel) ApplySubjectChange(ctx context.Context, change SubjectChange) error {
	return rm.exec(ctx,
		fmt.Sprintf("apply subject change for %s %s", change.Subject.SubjectType, change.Subject.SubjectID),
		`UPDATE onepagers.one_pager_subject_index
		SET name = CASE WHEN $4 = '' THEN name ELSE $4 END,
			required_count = $5, filled_count = $6, last_updated_at = $7
		WHERE tenant_id = $1 AND subject_type = $2 AND subject_id = $3`,
		change.Subject.SubjectType, change.Subject.SubjectID, change.Name,
		change.Counts.Required, change.Counts.Filled, change.OccurredAt,
	)
}

func (rm *OnePagerSubjectIndexReadModel) ApplyCompleteness(ctx context.Context, subjectType string, required int, filledBySubject map[string]int) error {
	if len(filledBySubject) == 0 {
		return nil
	}
	subjectIDs := make([]string, 0, len(filledBySubject))
	filledCounts := make([]int64, 0, len(filledBySubject))
	for subjectID, filled := range filledBySubject {
		subjectIDs = append(subjectIDs, subjectID)
		filledCounts = append(filledCounts, int64(filled))
	}
	return rm.exec(ctx,
		fmt.Sprintf("apply completeness for %d %s subjects", len(subjectIDs), subjectType),
		`UPDATE onepagers.one_pager_subject_index AS idx
		SET required_count = $3, filled_count = filled.count
		FROM unnest($4::text[], $5::int[]) AS filled(subject_id, count)
		WHERE idx.tenant_id = $1 AND idx.subject_type = $2 AND idx.subject_id = filled.subject_id`,
		subjectType, required, pq.Array(subjectIDs), pq.Array(filledCounts),
	)
}

func (rm *OnePagerSubjectIndexReadModel) SubjectIDs(ctx context.Context, subjectType string) ([]string, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var subjectIDs []string
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx,
			`SELECT subject_id FROM onepagers.one_pager_subject_index WHERE tenant_id = $1 AND subject_type = $2`,
			tenantID.Value(), subjectType,
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
			subjectIDs = append(subjectIDs, id)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("list %s subject index ids: %w", subjectType, err)
	}
	return subjectIDs, nil
}

func (rm *OnePagerSubjectIndexReadModel) exec(ctx context.Context, description, query string, args ...any) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}
	if _, err := rm.db.ExecContext(ctx, query, append([]any{tenantID.Value()}, args...)...); err != nil {
		return fmt.Errorf("%s: %w", description, err)
	}
	return nil
}
