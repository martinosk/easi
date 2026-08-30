package onepagers_test

import (
	"go/ast"
	"go/build/constraint"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const internalPrefix = "easi/backend/internal/"

var allowedImportPrefixes = []string{
	"easi/backend/internal/onepagers/",
	"easi/backend/internal/shared/",
	"easi/backend/internal/infrastructure/database",
	"easi/backend/internal/infrastructure/eventstore",
}

func isAllowedImport(importPath string) bool {
	if !strings.HasPrefix(importPath, internalPrefix) {
		return true
	}
	for _, prefix := range allowedImportPrefixes {
		if strings.HasPrefix(importPath, prefix) {
			return true
		}
	}
	segments := strings.Split(strings.TrimPrefix(importPath, internalPrefix), "/")
	return len(segments) >= 2 && segments[1] == "publishedlanguage"
}

func isIntegrationTaggedFile(path string) (bool, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	compiles, err := compilesWithoutIntegrationTag(content)
	if err != nil {
		return false, err
	}
	return !compiles, nil
}

func compilesWithoutIntegrationTag(content []byte) (bool, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "", content, parser.PackageClauseOnly|parser.ParseComments)
	if err != nil {
		return false, err
	}
	return evalBuildConstraintsBeforePackageClause(node), nil
}

func evalBuildConstraintsBeforePackageClause(node *ast.File) bool {
	for _, group := range node.Comments {
		if group.Pos() >= node.Package {
			break
		}
		if !evalCommentGroup(group) {
			return false
		}
	}
	return true
}

func evalCommentGroup(group *ast.CommentGroup) bool {
	for _, comment := range group.List {
		if compiles, ok := evalBuildConstraint(comment); ok && !compiles {
			return false
		}
	}
	return true
}

func evalBuildConstraint(comment *ast.Comment) (compiles bool, isConstraint bool) {
	if !constraint.IsGoBuild(comment.Text) {
		return false, false
	}
	expr, err := constraint.Parse(comment.Text)
	if err != nil {
		return false, false
	}
	return expr.Eval(noDefaultBuildTagsSet), true
}

func noDefaultBuildTagsSet(string) bool {
	return false
}

func forbiddenImports(path string) ([]string, error) {
	integration, err := isIntegrationTaggedFile(path)
	if err != nil || integration {
		return nil, err
	}
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
	if err != nil {
		return nil, err
	}
	var forbidden []string
	for _, imp := range file.Imports {
		importPath := strings.Trim(imp.Path.Value, `"`)
		if !isAllowedImport(importPath) {
			forbidden = append(forbidden, importPath)
		}
	}
	return forbidden, nil
}

func TestOnePagersExposesNoPublishedLanguage(t *testing.T) {
	root, err := filepath.Abs(".")
	if err != nil {
		t.Fatalf("failed to resolve onepagers root: %v", err)
	}

	if _, err := os.Stat(filepath.Join(root, "publishedlanguage")); !os.IsNotExist(err) {
		t.Error("BOUNDARY VIOLATION: internal/onepagers must not expose a publishedlanguage package — onepagers publishes no events to other contexts; its event types are internal aggregate mechanics")
	}
}

func TestArchitectureBoundary(t *testing.T) {
	root, err := filepath.Abs(".")
	if err != nil {
		t.Fatalf("failed to resolve onepagers root: %v", err)
	}

	if walkErr := filepath.WalkDir(root, boundaryCheck(t, root)); walkErr != nil {
		t.Fatalf("failed to scan internal/onepagers imports: %v", walkErr)
	}
}

func boundaryCheck(t *testing.T, root string) fs.WalkDirFunc {
	return func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		return reportForbiddenImports(t, root, path)
	}
}

func reportForbiddenImports(t *testing.T, root, path string) error {
	relPath, err := filepath.Rel(root, path)
	if err != nil {
		return err
	}
	forbidden, err := forbiddenImports(path)
	if err != nil {
		return err
	}
	for _, importPath := range forbidden {
		t.Errorf(
			"BOUNDARY VIOLATION: internal/onepagers/%s imports %q — onepagers may only import stdlib, third-party, internal/shared, internal/onepagers, other contexts' publishedlanguage packages, and the shared eventstore/database infrastructure",
			filepath.ToSlash(relPath), importPath,
		)
	}
	return nil
}
