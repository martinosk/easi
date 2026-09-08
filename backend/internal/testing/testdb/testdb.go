//go:build integration

package testdb

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

type role struct {
	userVar         string
	passwordVar     string
	defaultUser     string
	defaultPassword string
}

var (
	appRole = role{
		userVar:         "INTEGRATION_TEST_DB_USER",
		passwordVar:     "INTEGRATION_TEST_DB_PASSWORD",
		defaultUser:     "easi_app",
		defaultPassword: "localdev",
	}
	adminRole = role{
		userVar:         "INTEGRATION_TEST_DB_ADMIN_USER",
		passwordVar:     "INTEGRATION_TEST_DB_ADMIN_PASSWORD",
		defaultUser:     "easi",
		defaultPassword: "easi",
	}
)

func Open(t *testing.T) *sql.DB {
	t.Helper()
	return open(t, appRole)
}

func OpenAdmin(t *testing.T) *sql.DB {
	t.Helper()
	return open(t, adminRole)
}

func connStringFor(r role) string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		envOr("INTEGRATION_TEST_DB_HOST", "localhost"),
		envOr("INTEGRATION_TEST_DB_PORT", "5432"),
		envOr(r.userVar, r.defaultUser),
		envOr(r.passwordVar, r.defaultPassword),
		envOr("INTEGRATION_TEST_DB_NAME", "easi"),
		envOr("INTEGRATION_TEST_DB_SSLMODE", "disable"),
	)
}

func open(t *testing.T, r role) *sql.DB {
	t.Helper()
	db, err := sql.Open("postgres", connStringFor(r))
	require.NoError(t, err)
	require.NoError(t, db.Ping())
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
