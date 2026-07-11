package migrations

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubstituteEnvVars(t *testing.T) {
	t.Run("no placeholders returns content unchanged", func(t *testing.T) {
		result, err := substituteEnvVars("SELECT 1;")
		require.NoError(t, err)
		assert.Equal(t, "SELECT 1;", result)
	})

	t.Run("substitutes a set variable", func(t *testing.T) {
		t.Setenv("EASI_TEST_PASSWORD", "s3cret")
		result, err := substituteEnvVars("ALTER USER x WITH PASSWORD '${EASI_TEST_PASSWORD}';")
		require.NoError(t, err)
		assert.Equal(t, "ALTER USER x WITH PASSWORD 's3cret';", result)
	})

	t.Run("substitutes repeated occurrences of the same variable", func(t *testing.T) {
		t.Setenv("EASI_TEST_PASSWORD", "s3cret")
		result, err := substituteEnvVars("${EASI_TEST_PASSWORD} ${EASI_TEST_PASSWORD}")
		require.NoError(t, err)
		assert.Equal(t, "s3cret s3cret", result)
	})

	t.Run("fails when variable is unset", func(t *testing.T) {
		_, err := substituteEnvVars("PASSWORD '${EASI_TEST_UNSET_VAR}'")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "EASI_TEST_UNSET_VAR")
	})

	t.Run("fails when variable is set to empty string", func(t *testing.T) {
		t.Setenv("EASI_TEST_EMPTY_VAR", "")
		_, err := substituteEnvVars("PASSWORD '${EASI_TEST_EMPTY_VAR}'")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "EASI_TEST_EMPTY_VAR")
	})

	t.Run("never returns the literal placeholder on error", func(t *testing.T) {
		result, err := substituteEnvVars("PASSWORD '${EASI_TEST_UNSET_VAR}'")
		require.Error(t, err)
		assert.Empty(t, result)
	})

	t.Run("reports every distinct missing variable once", func(t *testing.T) {
		t.Setenv("EASI_TEST_SET_VAR", "value")
		_, err := substituteEnvVars("${EASI_TEST_MISSING_ONE} ${EASI_TEST_SET_VAR} ${EASI_TEST_MISSING_TWO} ${EASI_TEST_MISSING_ONE}")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "EASI_TEST_MISSING_ONE")
		assert.Contains(t, err.Error(), "EASI_TEST_MISSING_TWO")
		assert.Equal(t, 1, countOccurrences(err.Error(), "EASI_TEST_MISSING_ONE"))
	})
}

func countOccurrences(haystack, needle string) int {
	count := 0
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			count++
			i += len(needle) - 1
		}
	}
	return count
}

func TestValidateMigrationFilename(t *testing.T) {
	t.Run("accepts a well-formed filename", func(t *testing.T) {
		assert.NoError(t, validateMigrationFilename("001_init_schema.sql"))
	})

	t.Run("rejects a non-sql extension", func(t *testing.T) {
		assert.Error(t, validateMigrationFilename("001_init_schema.txt"))
	})

	t.Run("rejects path traversal", func(t *testing.T) {
		assert.Error(t, validateMigrationFilename("../../etc/passwd.sql"))
	})

	t.Run("rejects filenames with spaces", func(t *testing.T) {
		assert.Error(t, validateMigrationFilename("001 init schema.sql"))
	})
}
