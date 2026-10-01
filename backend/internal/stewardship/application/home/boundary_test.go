package home

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const homePackage = "easi/backend/internal/stewardship/application/home"

func isGoSource(entry os.DirEntry) bool {
	return !entry.IsDir() && strings.HasSuffix(entry.Name(), ".go")
}

func importsOf(t *testing.T, path string) []string {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
	require.NoError(t, err)
	imports := make([]string, len(file.Imports))
	for i, imp := range file.Imports {
		imports[i], _ = strconv.Unquote(imp.Path.Value)
	}
	return imports
}

func TestWriteSideDoesNotImportTheHomeReadSide(t *testing.T) {
	for _, dir := range []string{"../handlers", "../projectors", "../commands", "../../domain"} {
		err := filepath.WalkDir(dir, func(path string, entry os.DirEntry, err error) error {
			if err != nil || !isGoSource(entry) {
				return err
			}
			for _, importPath := range importsOf(t, path) {
				if strings.HasPrefix(importPath, homePackage) {
					t.Errorf("%s imports the home read side (%s); command handlers and reactors read only the domain and user caches (spec 228 rule 12)", path, importPath)
				}
			}
			return nil
		})
		require.NoError(t, err)
	}
}
