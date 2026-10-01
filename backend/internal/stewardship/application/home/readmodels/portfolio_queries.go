package readmodels

import (
	"context"
	"database/sql"

	"github.com/lib/pq"

	"easi/backend/internal/stewardship/application/home"
)

const portfolioCountsQuery = `
	WITH RECURSIVE domain_tree(capability_id, depth) AS (
		SELECT c.capability_id, 0
		FROM stewardship.domain_assignment_cache a
		JOIN stewardship.capability_cache c ON c.tenant_id = $1 AND c.capability_id = a.capability_id AND c.parent_id IS NULL
		WHERE a.tenant_id = $1 AND a.domain_id = ANY($2::text[])
		UNION
		SELECT c.capability_id, t.depth + 1
		FROM domain_tree t
		JOIN stewardship.capability_cache c ON c.tenant_id = $1 AND c.parent_id = t.capability_id
		WHERE t.depth < 8
	),
	portfolio_capabilities AS (
		SELECT c.capability_id, c.status
		FROM stewardship.capability_cache c
		WHERE c.tenant_id = $1
		  AND ($4 OR c.capability_id = ANY($3::text[]) OR c.capability_id IN (SELECT capability_id FROM domain_tree))
	),
	portfolio_realizations AS (
		SELECT r.capability_id, r.component_id, r.capability_id IN (SELECT capability_id FROM portfolio_capabilities) AS of_capability
		FROM stewardship.realization_cache r
		JOIN stewardship.capability_cache c ON c.tenant_id = $1 AND c.capability_id = r.capability_id
		JOIN stewardship.application_cache a ON a.tenant_id = $1 AND a.component_id = r.component_id
		WHERE r.tenant_id = $1
		  AND ($4 OR r.capability_id IN (SELECT capability_id FROM portfolio_capabilities) OR r.component_id = ANY($5::text[]))
	)
	SELECT 'status', status, COUNT(*) FROM portfolio_capabilities GROUP BY status
	UNION ALL
	SELECT 'applications', '', COUNT(*) FROM stewardship.application_cache a
	WHERE a.tenant_id = $1
	  AND ($4 OR a.component_id = ANY($5::text[])
	       OR a.component_id IN (SELECT component_id FROM portfolio_realizations WHERE of_capability))
	UNION ALL
	SELECT 'grade', COALESCE(t.grade, ''), COUNT(*)
	FROM portfolio_realizations r
	LEFT JOIN stewardship.time_assessment_cache t
	       ON t.tenant_id = $1 AND t.capability_id = r.capability_id AND t.component_id = r.component_id
	GROUP BY 2`

const domainNamesQuery = `
	WITH RECURSIVE ancestors(capability_id, parent_id, depth) AS (
		SELECT c.capability_id, c.parent_id, 0
		FROM stewardship.capability_cache c
		WHERE c.tenant_id = $1
		  AND (c.capability_id = ANY($3::text[])
		       OR c.capability_id IN (SELECT r.capability_id FROM stewardship.realization_cache r
		                              WHERE r.tenant_id = $1 AND r.component_id = ANY($5::text[])))
		UNION ALL
		SELECT p.capability_id, p.parent_id, u.depth + 1
		FROM ancestors u
		JOIN stewardship.capability_cache p ON p.tenant_id = $1 AND p.capability_id = u.parent_id
		WHERE u.depth < 8
	),
	scope_domains AS (
		SELECT a.domain_id
		FROM ancestors u
		JOIN stewardship.domain_assignment_cache a ON a.tenant_id = $1 AND a.capability_id = u.capability_id
		WHERE u.parent_id IS NULL
		UNION
		SELECT unnest($2::text[])
	)
	SELECT d.name FROM stewardship.domain_cache d
	WHERE d.tenant_id = $1 AND ($4 OR d.domain_id IN (SELECT domain_id FROM scope_domains))
	ORDER BY d.name`

func (q *Queries) Portfolio(ctx context.Context, seed home.PortfolioSeed) (home.PortfolioFacts, error) {
	var facts home.PortfolioFacts
	args := []any{pq.Array(seed.DomainIDs), pq.Array(seed.CapabilityIDs), seed.All, pq.Array(seed.ApplicationIDs)}
	err := q.store.read(ctx, "compose home portfolio", func(tx tenantQuery) error {
		if err := tx.each(portfolioCountsQuery, func(rows *sql.Rows) error { return scanPortfolioCount(&facts, rows) }, args...); err != nil {
			return err
		}
		return tx.each(domainNamesQuery, func(rows *sql.Rows) error { return scanDomainName(&facts, rows) }, args...)
	})
	return facts, err
}

func scanPortfolioCount(f *home.PortfolioFacts, rows *sql.Rows) error {
	var (
		kind, key string
		count     int
	)
	if err := rows.Scan(&kind, &key, &count); err != nil {
		return err
	}
	switch kind {
	case "status":
		return f.Statuses.Add(key, count)
	case "applications":
		f.Applications = count
	case "grade":
		return f.Grades.Add(key, count)
	}
	return nil
}

func scanDomainName(f *home.PortfolioFacts, rows *sql.Rows) error {
	var name string
	if err := rows.Scan(&name); err != nil {
		return err
	}
	f.DomainNames = append(f.DomainNames, name)
	return nil
}

const subjectGradesQuery = `
	SELECT s.subject_type, s.subject_id, t.grade, COUNT(*)
	FROM stewardship.realization_cache r
	JOIN stewardship.time_assessment_cache t
	     ON t.tenant_id = $1 AND t.capability_id = r.capability_id AND t.component_id = r.component_id
	JOIN stewardship.capability_cache c ON c.tenant_id = $1 AND c.capability_id = r.capability_id
	JOIN stewardship.application_cache a ON a.tenant_id = $1 AND a.component_id = r.component_id
	CROSS JOIN LATERAL (VALUES ('capability', r.capability_id), ('application', r.component_id)) AS s(subject_type, subject_id)
	WHERE r.tenant_id = $1
	  AND ((s.subject_type = 'capability' AND s.subject_id = ANY($2::text[]))
	       OR (s.subject_type = 'application' AND s.subject_id = ANY($3::text[])))
	GROUP BY 1, 2, 3`

func (q *Queries) SubjectGrades(ctx context.Context, subjects home.Subjects) (map[home.Subject]home.GradeCounts, error) {
	grades := map[home.Subject]home.GradeCounts{}
	err := q.store.read(ctx, "read TIME grades of home subjects", func(tx tenantQuery) error {
		return tx.each(subjectGradesQuery, func(rows *sql.Rows) error {
			return countSubjectGrade(grades, rows)
		}, pq.Array(subjects.IDsOf(home.SubjectCapability)), pq.Array(subjects.IDsOf(home.SubjectApplication)))
	})
	return grades, err
}

func countSubjectGrade(grades map[home.Subject]home.GradeCounts, rows *sql.Rows) error {
	var (
		subject home.Subject
		grade   string
		count   int
	)
	if err := rows.Scan(&subject.Type, &subject.ID, &grade, &count); err != nil {
		return err
	}
	counts := grades[subject]
	if err := counts.Add(grade, count); err != nil {
		return err
	}
	grades[subject] = counts
	return nil
}
