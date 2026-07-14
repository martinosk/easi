package readmodels

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	sharedctx "easi/backend/internal/shared/context"

	"github.com/lib/pq"
)

func (rm *OnePagerSubjectIndexReadModel) Page(ctx context.Context, query SubjectIndexQuery) ([]SubjectIndexRecord, bool, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, false, err
	}
	if len(query.SubjectTypes) == 0 {
		return []SubjectIndexRecord{}, false, nil
	}

	pageSQL, args := buildPageSQL(tenantID.Value(), query)

	var records []SubjectIndexRecord
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, pageSQL, args...)
		if err != nil {
			return err
		}
		defer func() { _ = rows.Close() }()
		records, err = scanSubjectIndexRows(rows)
		return err
	})
	if err != nil {
		return nil, false, fmt.Errorf("query one-pager subject index page: %w", err)
	}

	if len(records) > query.Limit {
		return records[:query.Limit], true, nil
	}
	return records, false, nil
}

func scanSubjectIndexRows(rows *sql.Rows) ([]SubjectIndexRecord, error) {
	var records []SubjectIndexRecord
	for rows.Next() {
		var record SubjectIndexRecord
		if err := rows.Scan(
			&record.SubjectType, &record.SubjectID, &record.Name,
			&record.CreatorActorID, &record.CreatorEmail, &record.CreatedAt, &record.LastUpdatedAt,
			&record.RequiredCount, &record.FilledCount,
		); err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, rows.Err()
}

type keysetColumn struct {
	expr string
	asc  bool
	val  any
}

const projectedColumns = `SELECT subject_type, subject_id, name, creator_actor_id, creator_email, created_at, last_updated_at, required_count, filled_count
FROM (
	SELECT *, LOWER(name) AS name_lower, LOWER(creator_email) AS email_lower,
		(CASE WHEN required_count = 0 THEN 2 WHEN filled_count >= required_count THEN 1 ELSE 0 END) AS bucket,
		(required_count - filled_count) AS missing
	FROM onepagers.one_pager_subject_index
	WHERE tenant_id = $1 AND subject_type = ANY($2)
) t`

func buildPageSQL(tenantID string, query SubjectIndexQuery) (string, []any) {
	args := []any{tenantID, pq.Array(query.SubjectTypes)}
	columns := sortColumns(query)

	var builder strings.Builder
	builder.WriteString(projectedColumns)
	if query.After != nil {
		builder.WriteString(" WHERE ")
		builder.WriteString(keysetPredicate(columns, &args))
	}
	builder.WriteString(" ORDER BY ")
	builder.WriteString(orderByClause(columns))

	args = append(args, query.Limit+1)
	fmt.Fprintf(&builder, " LIMIT $%d", len(args))
	return builder.String(), args
}

func sortColumns(query SubjectIndexQuery) []keysetColumn {
	asc := query.Order != OrderDesc
	primary := primarySortColumns(query.Sort, asc, query.After)
	return append(primary,
		keysetColumn{"subject_type", asc, cursorValue(query.After, func(r *SubjectIndexRecord) any { return r.SubjectType })},
		keysetColumn{"subject_id", asc, cursorValue(query.After, func(r *SubjectIndexRecord) any { return r.SubjectID })},
	)
}

func primarySortColumns(sort string, asc bool, after *SubjectIndexRecord) []keysetColumn {
	switch sort {
	case SortName:
		return []keysetColumn{{"name_lower", asc, cursorValue(after, func(r *SubjectIndexRecord) any { return strings.ToLower(r.Name) })}}
	case SortCreator:
		return []keysetColumn{{"email_lower", asc, cursorValue(after, func(r *SubjectIndexRecord) any { return strings.ToLower(r.CreatorEmail) })}}
	case SortCreated:
		return []keysetColumn{{"created_at", asc, cursorValue(after, func(r *SubjectIndexRecord) any { return r.CreatedAt })}}
	case SortUpdated:
		return []keysetColumn{{"last_updated_at", asc, cursorValue(after, func(r *SubjectIndexRecord) any { return r.LastUpdatedAt })}}
	default:
		return []keysetColumn{
			{"bucket", asc, cursorValue(after, func(r *SubjectIndexRecord) any { return r.CompletenessBucket() })},
			{"missing", !asc, cursorValue(after, func(r *SubjectIndexRecord) any { return r.RequiredCount - r.FilledCount })},
		}
	}
}

func cursorValue(after *SubjectIndexRecord, get func(*SubjectIndexRecord) any) any {
	if after == nil {
		return nil
	}
	return get(after)
}

func keysetPredicate(columns []keysetColumn, args *[]any) string {
	terms := make([]string, 0, len(columns))
	for i := range columns {
		var clause strings.Builder
		clause.WriteString("(")
		for j := 0; j < i; j++ {
			fmt.Fprintf(&clause, "%s = %s AND ", columns[j].expr, placeholder(columns[j].val, args))
		}
		fmt.Fprintf(&clause, "%s %s %s", columns[i].expr, comparator(columns[i].asc), placeholder(columns[i].val, args))
		clause.WriteString(")")
		terms = append(terms, clause.String())
	}
	return strings.Join(terms, " OR ")
}

func orderByClause(columns []keysetColumn) string {
	parts := make([]string, len(columns))
	for i, column := range columns {
		parts[i] = column.expr + " " + direction(column.asc)
	}
	return strings.Join(parts, ", ")
}

func placeholder(value any, args *[]any) string {
	*args = append(*args, value)
	return fmt.Sprintf("$%d", len(*args))
}

func comparator(asc bool) string {
	if asc {
		return ">"
	}
	return "<"
}

func direction(asc bool) string {
	if asc {
		return "ASC"
	}
	return "DESC"
}
