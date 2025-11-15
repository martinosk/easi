package migrations

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *sql.DB {
	// Use environment variables or defaults for test database
	dbHost := getEnv("TEST_DB_HOST", "localhost")
	dbPort := getEnv("TEST_DB_PORT", "5432")
	dbUser := getEnv("TEST_DB_USER", "easi")
	dbPassword := getEnv("TEST_DB_PASSWORD", "easi")
	dbName := getEnv("TEST_DB_NAME", "easi_test")

	connStr := "host=" + dbHost + " port=" + dbPort + " user=" + dbUser +
		" password=" + dbPassword + " dbname=" + dbName + " sslmode=disable"

	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err)

	err = db.Ping()
	require.NoError(t, err)

	return db
}

func cleanupTestDB(t *testing.T, db *sql.DB) {
	// Drop the migrations table
	_, err := db.Exec("DROP TABLE IF EXISTS schema_migrations CASCADE")
	require.NoError(t, err)

	db.Close()
}

func createTestMigrationFiles(t *testing.T, dir string, files map[string]string) {
	for filename, content := range files {
		filePath := filepath.Join(dir, filename)
		err := os.WriteFile(filePath, []byte(content), 0644)
		require.NoError(t, err)
	}
}

func TestRunner_CreateMigrationsTable(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	runner := NewRunner(db, "./test_migrations")
	err := runner.createMigrationsTable()
	require.NoError(t, err)

	// Verify table exists
	var tableName string
	err = db.QueryRow(`
		SELECT table_name
		FROM information_schema.tables
		WHERE table_name = 'schema_migrations'
	`).Scan(&tableName)
	require.NoError(t, err)
	assert.Equal(t, "schema_migrations", tableName)
}

func TestRunner_GetMigrationFiles(t *testing.T) {
	// Create temporary directory for test migrations
	tmpDir, err := os.MkdirTemp("", "migrations_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create test migration files
	testFiles := map[string]string{
		"001_first.sql":  "CREATE TABLE test1 (id SERIAL PRIMARY KEY);",
		"002_second.sql": "CREATE TABLE test2 (id SERIAL PRIMARY KEY);",
		"README.md":      "This is a readme",
		"003_third.sql":  "CREATE TABLE test3 (id SERIAL PRIMARY KEY);",
	}
	createTestMigrationFiles(t, tmpDir, testFiles)

	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	runner := NewRunner(db, tmpDir)
	files, err := runner.getMigrationFiles()
	require.NoError(t, err)

	// Should only return .sql files, sorted
	expected := []string{"001_first.sql", "002_second.sql", "003_third.sql"}
	assert.Equal(t, expected, files)
}

func TestRunner_ExecuteMigration(t *testing.T) {
	// Create temporary directory for test migrations
	tmpDir, err := os.MkdirTemp("", "migrations_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create a test migration
	testFiles := map[string]string{
		"001_create_test_table.sql": "CREATE TABLE test_migration_table (id SERIAL PRIMARY KEY, name VARCHAR(100));",
	}
	createTestMigrationFiles(t, tmpDir, testFiles)

	db := setupTestDB(t)
	defer func() {
		// Clean up test table
		db.Exec("DROP TABLE IF EXISTS test_migration_table")
		cleanupTestDB(t, db)
	}()

	runner := NewRunner(db, tmpDir)

	// Create migrations table
	err = runner.createMigrationsTable()
	require.NoError(t, err)

	// Execute migration
	err = runner.executeMigration("001_create_test_table.sql")
	require.NoError(t, err)

	// Verify table was created
	var tableName string
	err = db.QueryRow(`
		SELECT table_name
		FROM information_schema.tables
		WHERE table_name = 'test_migration_table'
	`).Scan(&tableName)
	require.NoError(t, err)
	assert.Equal(t, "test_migration_table", tableName)

	// Verify migration was recorded
	var migrationName string
	err = db.QueryRow("SELECT migration_name FROM schema_migrations WHERE migration_name = $1",
		"001_create_test_table.sql").Scan(&migrationName)
	require.NoError(t, err)
	assert.Equal(t, "001_create_test_table.sql", migrationName)
}

func TestRunner_GetExecutedMigrations(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	runner := NewRunner(db, "./test_migrations")

	// Create migrations table
	err := runner.createMigrationsTable()
	require.NoError(t, err)

	// Insert some executed migrations
	_, err = db.Exec("INSERT INTO schema_migrations (migration_name) VALUES ($1), ($2)",
		"001_first.sql", "002_second.sql")
	require.NoError(t, err)

	// Get executed migrations
	executed, err := runner.getExecutedMigrations()
	require.NoError(t, err)

	// Verify results
	assert.True(t, executed["001_first.sql"])
	assert.True(t, executed["002_second.sql"])
	assert.False(t, executed["003_third.sql"])
}

func TestRunner_Run_ExecutesAllMigrations(t *testing.T) {
	// Create temporary directory for test migrations
	tmpDir, err := os.MkdirTemp("", "migrations_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create test migrations
	testFiles := map[string]string{
		"001_first.sql":  "CREATE TABLE migration_test_1 (id SERIAL PRIMARY KEY);",
		"002_second.sql": "CREATE TABLE migration_test_2 (id SERIAL PRIMARY KEY);",
		"003_third.sql":  "CREATE TABLE migration_test_3 (id SERIAL PRIMARY KEY);",
	}
	createTestMigrationFiles(t, tmpDir, testFiles)

	db := setupTestDB(t)
	defer func() {
		db.Exec("DROP TABLE IF EXISTS migration_test_1")
		db.Exec("DROP TABLE IF EXISTS migration_test_2")
		db.Exec("DROP TABLE IF EXISTS migration_test_3")
		cleanupTestDB(t, db)
	}()

	runner := NewRunner(db, tmpDir)

	// Run migrations
	err = runner.Run()
	require.NoError(t, err)

	// Verify all tables were created
	var count int
	err = db.QueryRow(`
		SELECT COUNT(*)
		FROM information_schema.tables
		WHERE table_name IN ('migration_test_1', 'migration_test_2', 'migration_test_3')
	`).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 3, count)

	// Verify all migrations were recorded
	err = db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 3, count)
}

func TestRunner_Run_SkipsExecutedMigrations(t *testing.T) {
	// Create temporary directory for test migrations
	tmpDir, err := os.MkdirTemp("", "migrations_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create test migrations
	testFiles := map[string]string{
		"001_first.sql":  "CREATE TABLE migration_skip_test_1 (id SERIAL PRIMARY KEY);",
		"002_second.sql": "CREATE TABLE migration_skip_test_2 (id SERIAL PRIMARY KEY);",
	}
	createTestMigrationFiles(t, tmpDir, testFiles)

	db := setupTestDB(t)
	defer func() {
		db.Exec("DROP TABLE IF EXISTS migration_skip_test_1")
		db.Exec("DROP TABLE IF EXISTS migration_skip_test_2")
		cleanupTestDB(t, db)
	}()

	runner := NewRunner(db, tmpDir)

	// Run migrations first time
	err = runner.Run()
	require.NoError(t, err)

	// Run migrations again - should skip already executed ones
	err = runner.Run()
	require.NoError(t, err)

	// Verify migrations were only recorded once
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 2, count)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
