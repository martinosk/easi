package readmodels

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/stewardship/application/home"
)

type Queries struct {
	store tenantStore
}

func NewQueries(db *database.TenantAwareDB) *Queries {
	return &Queries{store: tenantStore{db: db}}
}

func (q *Queries) CallerFacts(ctx context.Context, caller home.Caller, now time.Time) (home.CallerFacts, error) {
	var facts home.CallerFacts
	err := q.store.read(ctx, "read home facts of user "+caller.UserID, func(tx tenantQuery) error {
		if caller.UserID != "" {
			if err := readUserFacts(tx, caller.UserID, &facts); err != nil {
				return err
			}
		}
		if caller.Email == "" {
			return nil
		}
		grants, err := list(tx, grantsQuery, grantColumns,
			strings.ToLower(caller.Email), now.UTC(), string(home.SubjectCapability), string(home.SubjectApplication))
		facts.Grants = grants
		return err
	})
	return facts, err
}

func readUserFacts(tx tenantQuery, userID string, facts *home.CallerFacts) (err error) {
	if facts.Stewardships, err = list(tx, stewardshipsQuery, stewardshipColumns, userID); err != nil {
		return err
	}
	if facts.ArchitectedDomains, err = list(tx, architectedDomainsQuery, domainColumns, userID); err != nil {
		return err
	}
	if facts.EAOwnedCapabilities, err = list(tx, eaOwnedCapabilitiesQuery, ownedCapabilityColumns, userID); err != nil {
		return err
	}
	facts.Ownerships, err = list(tx, applicationOwnershipsQuery, ownershipColumns, userID)
	return err
}

func list[T any](tx tenantQuery, query string, columns func(*T) []any, args ...any) ([]T, error) {
	var found []T
	err := tx.each(query, func(rows *sql.Rows) error {
		var row T
		if err := rows.Scan(columns(&row)...); err != nil {
			return err
		}
		found = append(found, row)
		return nil
	}, args...)
	return found, err
}

func stewardshipColumns(s *home.StewardshipAnchor) []any {
	return []any{&s.Domain.ID, &s.Domain.Name, &s.Concern}
}

func domainColumns(d *home.Domain) []any {
	return []any{&d.ID, &d.Name}
}

func ownedCapabilityColumns(c *home.OwnedCapability) []any {
	return []any{&c.ID, &c.Name, &c.Level}
}

func ownershipColumns(o *home.ApplicationOwnership) []any {
	return []any{&o.ApplicationID, &o.Name, &o.State}
}

func grantColumns(g *home.EditGrant) []any {
	return []any{&g.Subject.Type, &g.Subject.ID, &g.Name, &g.Level, &g.ExpiresAt}
}

const stewardshipsQuery = `
	SELECT s.domain_id, d.name, s.concern
	FROM stewardship.stewardships s
	JOIN stewardship.domain_cache d ON d.tenant_id = s.tenant_id AND d.domain_id = s.domain_id
	WHERE s.tenant_id = $1 AND s.steward_id = $2
	ORDER BY d.name, s.concern`

const architectedDomainsQuery = `
	SELECT domain_id, name FROM stewardship.domain_cache
	WHERE tenant_id = $1 AND domain_architect_id = $2
	ORDER BY name`

const eaOwnedCapabilitiesQuery = `
	SELECT capability_id, name, level
	FROM stewardship.capability_cache
	WHERE tenant_id = $1 AND ea_owner = $2`

const applicationOwnershipsQuery = `
	SELECT component_id, name, ownership_state
	FROM stewardship.application_cache
	WHERE tenant_id = $1 AND owner_kind = 'user' AND owner_id = $2`

const grantsQuery = `
	SELECT CASE WHEN c.capability_id IS NOT NULL THEN $4::text ELSE $5::text END,
	       g.artifact_id, COALESCE(c.name, a.name), COALESCE(c.level, ''), g.expires_at
	FROM stewardship.edit_grant_cache g
	LEFT JOIN stewardship.capability_cache c
	       ON g.artifact_type = 'capability' AND c.tenant_id = $1 AND c.capability_id = g.artifact_id
	LEFT JOIN stewardship.application_cache a
	       ON g.artifact_type = 'component' AND a.tenant_id = $1 AND a.component_id = g.artifact_id
	WHERE g.tenant_id = $1 AND lower(g.grantee_email) = $2 AND g.expires_at > $3::timestamp
	  AND (c.capability_id IS NOT NULL OR a.component_id IS NOT NULL)`
