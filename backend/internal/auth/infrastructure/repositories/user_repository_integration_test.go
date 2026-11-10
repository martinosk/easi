//go:build integration

package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"easi/backend/internal/testing/testdb"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const userRepositoryTestTenant = "auth-user-repo-test"

func openSingleConnectionDB(t *testing.T) *sql.DB {
	t.Helper()
	if testing.Short() {
		t.Skip("Skipping integration test")
	}
	db := testdb.Open(t)
	db.SetMaxOpenConns(1)
	return db
}

func asTenant(t *testing.T, db *sql.DB, query string, args ...any) {
	t.Helper()
	tx, err := db.Begin()
	require.NoError(t, err)
	defer func() { _ = tx.Rollback() }()

	_, err = tx.Exec("SELECT set_config('app.current_tenant', $1, true)", userRepositoryTestTenant)
	require.NoError(t, err)
	_, err = tx.Exec(query, args...)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
}

func seedUser(t *testing.T, db *sql.DB) string {
	t.Helper()
	_, err := db.Exec(`INSERT INTO auth.tenants (id, name, status) VALUES ($1, 'User Repo Test', 'active') ON CONFLICT (id) DO NOTHING`, userRepositoryTestTenant)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM auth.tenants WHERE id = $1", userRepositoryTestTenant)
	})

	email := fmt.Sprintf("repo-%d@acme.com", time.Now().UnixNano())
	asTenant(t, db,
		`INSERT INTO auth.users (id, tenant_id, email, role, status, created_at)
		 VALUES (gen_random_uuid(), $1, $2, 'architect', 'active', NOW())`,
		userRepositoryTestTenant, email)
	t.Cleanup(func() {
		asTenant(t, db, "DELETE FROM auth.users WHERE tenant_id = $1 AND email = $2", userRepositoryTestTenant, email)
	})
	return email
}

func currentTenantSetting(t *testing.T, db *sql.DB) string {
	t.Helper()
	var setting sql.NullString
	require.NoError(t, db.QueryRow("SELECT current_setting('app.current_tenant', true)").Scan(&setting))
	return setting.String
}

func TestUserRepository_GetByEmail_FindsUserOfTheGivenTenant(t *testing.T) {
	db := openSingleConnectionDB(t)
	email := seedUser(t, db)

	user, err := NewUserRepository(db).GetByEmail(context.Background(), userRepositoryTestTenant, email)

	require.NoError(t, err)
	assert.Equal(t, email, user.Email)
	assert.Equal(t, "architect", user.Role)
	assert.Equal(t, "active", user.Status)
}

func TestUserRepository_GetByEmail_DoesNotFindUserOfAnotherTenant(t *testing.T) {
	db := openSingleConnectionDB(t)
	email := seedUser(t, db)

	_, err := NewUserRepository(db).GetByEmail(context.Background(), "some-other-tenant", email)

	assert.ErrorIs(t, err, ErrUserNotFound)
}

func TestUserRepository_GetByEmail_LeavesNoTenantOnThePooledConnection(t *testing.T) {
	db := openSingleConnectionDB(t)
	email := seedUser(t, db)

	_, err := NewUserRepository(db).GetByEmail(context.Background(), userRepositoryTestTenant, email)
	require.NoError(t, err)

	assert.Empty(t, currentTenantSetting(t, db), "the tenant must not outlive the lookup on a reused connection")
}

func TestUserRepository_GetByEmail_LeavesNoTenantWhenUserIsMissing(t *testing.T) {
	db := openSingleConnectionDB(t)

	_, err := NewUserRepository(db).GetByEmail(context.Background(), userRepositoryTestTenant, "nobody@acme.com")
	require.ErrorIs(t, err, ErrUserNotFound)

	assert.Empty(t, currentTenantSetting(t, db))
}
