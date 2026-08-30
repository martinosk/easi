package readmodels

import (
	"context"
	"database/sql"
	"fmt"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"

	"github.com/lib/pq"
)

type RelationEntry struct {
	EntryID     string
	RelatedType string
	RelatedID   string
	RelatedName string
	EdgeID      string
}

type RelationTarget struct {
	EntryID   string
	RelatedID string
}

type RelationQuery struct {
	SubjectType string
	SubjectIDs  []string
	EntryIDs    []string
}

type RelationReference struct {
	SubjectID   string
	EntryID     string
	RelatedType string
	RelatedID   string
	Label       string
}

type SubjectRelationCacheReadModel struct {
	db *database.TenantAwareDB
}

func NewSubjectRelationCacheReadModel(db *database.TenantAwareDB) *SubjectRelationCacheReadModel {
	return &SubjectRelationCacheReadModel{db: db}
}

func (rm *SubjectRelationCacheReadModel) Save(ctx context.Context, subject SubjectKey, entry RelationEntry) error {
	return rm.exec(ctx,
		fmt.Sprintf("cache %s relation %s of %s %s", entry.EntryID, entry.RelatedID, subject.SubjectType, subject.SubjectID),
		`INSERT INTO onepagers.subject_relation_cache
		(tenant_id, subject_type, subject_id, entry_id, related_type, related_id, related_name, edge_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (tenant_id, subject_type, subject_id, entry_id, related_id) DO UPDATE
		SET related_type = EXCLUDED.related_type, related_name = EXCLUDED.related_name, edge_id = EXCLUDED.edge_id`,
		subject.SubjectType, subject.SubjectID, entry.EntryID,
		entry.RelatedType, entry.RelatedID, entry.RelatedName, entry.EdgeID,
	)
}

func (rm *SubjectRelationCacheReadModel) Replace(ctx context.Context, subject SubjectKey, entryID string, entries []RelationEntry) error {
	if err := rm.exec(ctx,
		fmt.Sprintf("clear %s relations of %s %s", entryID, subject.SubjectType, subject.SubjectID),
		`DELETE FROM onepagers.subject_relation_cache
		WHERE tenant_id = $1 AND subject_type = $2 AND subject_id = $3 AND entry_id = $4`,
		subject.SubjectType, subject.SubjectID, entryID,
	); err != nil {
		return err
	}
	for _, entry := range entries {
		if err := rm.Save(ctx, subject, entry); err != nil {
			return err
		}
	}
	return nil
}

func (rm *SubjectRelationCacheReadModel) DeleteByEdge(ctx context.Context, edgeID string) error {
	if edgeID == "" {
		return nil
	}
	return rm.exec(ctx,
		fmt.Sprintf("delete cached relations of edge %s", edgeID),
		`DELETE FROM onepagers.subject_relation_cache WHERE tenant_id = $1 AND edge_id = $2`,
		edgeID,
	)
}

func (rm *SubjectRelationCacheReadModel) DeleteEdgeForSubjects(ctx context.Context, edgeID string, subjectIDs []string) error {
	if edgeID == "" || len(subjectIDs) == 0 {
		return nil
	}
	return rm.exec(ctx,
		fmt.Sprintf("delete cached relations of edge %s for %d subjects", edgeID, len(subjectIDs)),
		`DELETE FROM onepagers.subject_relation_cache
		WHERE tenant_id = $1 AND edge_id = $2 AND (subject_id = ANY($3) OR related_id = ANY($3))`,
		edgeID, pq.Array(subjectIDs),
	)
}

func (rm *SubjectRelationCacheReadModel) DeleteByRelated(ctx context.Context, target RelationTarget) error {
	return rm.exec(ctx,
		fmt.Sprintf("delete cached %s relations pointing at %s", target.EntryID, target.RelatedID),
		`DELETE FROM onepagers.subject_relation_cache
		WHERE tenant_id = $1 AND entry_id = $2 AND related_id = $3`,
		target.EntryID, target.RelatedID,
	)
}

func (rm *SubjectRelationCacheReadModel) DeleteSubject(ctx context.Context, subject SubjectKey) error {
	return rm.exec(ctx,
		fmt.Sprintf("delete cached relations of %s %s", subject.SubjectType, subject.SubjectID),
		`DELETE FROM onepagers.subject_relation_cache
		WHERE tenant_id = $1 AND ((subject_type = $2 AND subject_id = $3) OR (related_type = $2 AND related_id = $3))`,
		subject.SubjectType, subject.SubjectID,
	)
}

func (rm *SubjectRelationCacheReadModel) RenameRelated(ctx context.Context, target RelationTarget, name string) error {
	return rm.exec(ctx,
		fmt.Sprintf("rename cached %s relation target %s", target.EntryID, target.RelatedID),
		`UPDATE onepagers.subject_relation_cache SET related_name = $4
		WHERE tenant_id = $1 AND entry_id = $2 AND related_id = $3`,
		target.EntryID, target.RelatedID, name,
	)
}

const relationReferenceSelect = `SELECT r.subject_id, r.entry_id, r.related_type, r.related_id,
	COALESCE(NULLIF(i.name, ''), r.related_name) AS label
FROM onepagers.subject_relation_cache r
LEFT JOIN onepagers.one_pager_subject_index i
	ON i.tenant_id = r.tenant_id AND i.subject_type = r.related_type AND i.subject_id = r.related_id
WHERE r.tenant_id = $1 AND r.subject_type = $2 AND r.subject_id = ANY($3) AND r.entry_id = ANY($4)
ORDER BY r.subject_id, r.entry_id, label, r.related_id`

func (rm *SubjectRelationCacheReadModel) References(ctx context.Context, query RelationQuery) (map[string][]RelationReference, error) {
	references := map[string][]RelationReference{}
	if len(query.SubjectIDs) == 0 || len(query.EntryIDs) == 0 {
		return references, nil
	}
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, queryErr := tx.QueryContext(ctx, relationReferenceSelect,
			tenantID.Value(), query.SubjectType, pq.Array(query.SubjectIDs), pq.Array(query.EntryIDs))
		if queryErr != nil {
			return queryErr
		}
		defer func() { _ = rows.Close() }()
		return scanRelationReferences(rows, references)
	})
	if err != nil {
		return nil, fmt.Errorf("query cached %s relations: %w", query.SubjectType, err)
	}
	return references, nil
}

func scanRelationReferences(rows *sql.Rows, references map[string][]RelationReference) error {
	for rows.Next() {
		var reference RelationReference
		if err := rows.Scan(&reference.SubjectID, &reference.EntryID, &reference.RelatedType, &reference.RelatedID, &reference.Label); err != nil {
			return err
		}
		references[reference.SubjectID] = append(references[reference.SubjectID], reference)
	}
	return rows.Err()
}

func (rm *SubjectRelationCacheReadModel) CountSubjectsWithEntry(ctx context.Context, subjectType, entryID string) (int, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return 0, err
	}
	var count int
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx,
			`SELECT COUNT(DISTINCT subject_id) FROM onepagers.subject_relation_cache
			WHERE tenant_id = $1 AND subject_type = $2 AND entry_id = $3`,
			tenantID.Value(), subjectType, entryID,
		).Scan(&count)
	})
	if err != nil {
		return 0, fmt.Errorf("count %s subjects with a %s relation: %w", subjectType, entryID, err)
	}
	return count, nil
}

func (rm *SubjectRelationCacheReadModel) exec(ctx context.Context, description, query string, args ...any) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}
	if _, err := rm.db.ExecContext(ctx, query, append([]any{tenantID.Value()}, args...)...); err != nil {
		return fmt.Errorf("%s: %w", description, err)
	}
	return nil
}
