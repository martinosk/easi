//go:build !integration

package internal_test

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const modulePrefix = "easi/backend/internal/"

var sharedPackages = map[string]bool{
	"shared":         true,
	"infrastructure": true,
	"testing":        true,
}

var freeImportPackages = map[string]bool{
	"platform/infrastructure/api": true,
}

var allowedCrossBCImports = map[string]string{
	"platform -> auth/application/commands":        "platform-internal",
	"platform -> auth/application/handlers":        "platform-internal",
	"platform -> auth/infrastructure/repositories": "platform-internal",
}

type architectureScanner struct {
	internalDir string
}

func isGoSourceFile(info os.FileInfo, path string) bool {
	return !info.IsDir() && strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go")
}

func isProductionGoFile(info os.FileInfo, path string) bool {
	if !isGoSourceFile(info, path) {
		return false
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return true
	}
	firstLine := strings.SplitN(string(content), "\n", 2)[0]
	return !strings.Contains(firstLine, "//go:build integration")
}

func isAllowedCrossBCImport(ownerBC, importSuffix string) bool {
	importedBC := strings.SplitN(importSuffix, "/", 2)[0]
	if importedBC == ownerBC || sharedPackages[importedBC] {
		return true
	}
	if freeImportPackages[importSuffix] {
		return true
	}
	if strings.Contains(importSuffix, "/publishedlanguage/contracts") {
		return false
	}
	return strings.Contains(importSuffix, "/publishedlanguage")
}

type importViolation struct {
	relPath      string
	importSuffix string
	importedBC   string
}

type importScanResult struct {
	violations           []importViolation
	usedAllowlistEntries map[string]bool
}

func (s architectureScanner) checkFileImports(path string) ([]importViolation, map[string]bool, error) {
	relPath, err := filepath.Rel(s.internalDir, path)
	if err != nil {
		return nil, nil, err
	}
	relPath = filepath.ToSlash(relPath)

	ownerBC := strings.SplitN(relPath, "/", 2)[0]
	if sharedPackages[ownerBC] {
		return nil, nil, nil
	}

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
	if err != nil {
		return nil, nil, err
	}

	var violations []importViolation
	usedEntries := make(map[string]bool)

	for _, imp := range node.Imports {
		importPath := strings.Trim(imp.Path.Value, `"`)
		if !strings.HasPrefix(importPath, modulePrefix) {
			continue
		}

		suffix := importPath[len(modulePrefix):]
		if isAllowedCrossBCImport(ownerBC, suffix) {
			continue
		}

		allowlistKey := ownerBC + " -> " + suffix
		if _, ok := allowedCrossBCImports[allowlistKey]; ok {
			usedEntries[allowlistKey] = true
			continue
		}

		importedBC := strings.SplitN(suffix, "/", 2)[0]
		violations = append(violations, importViolation{relPath, suffix, importedBC})
	}

	return violations, usedEntries, nil
}

func (s architectureScanner) scanImports() (importScanResult, error) {
	result := importScanResult{usedAllowlistEntries: make(map[string]bool)}

	err := filepath.Walk(s.internalDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !isProductionGoFile(info, path) {
			return nil
		}
		violations, usedEntries, fileErr := s.checkFileImports(path)
		if fileErr != nil {
			return fileErr
		}
		result.violations = append(result.violations, violations...)
		for k := range usedEntries {
			result.usedAllowlistEntries[k] = true
		}
		return nil
	})

	return result, err
}

type purityViolation struct {
	relPath    string
	importPath string
}

func (s architectureScanner) checkContractsPurity(path string) ([]purityViolation, error) {
	relPath, err := filepath.Rel(s.internalDir, path)
	if err != nil {
		return nil, err
	}
	relPath = filepath.ToSlash(relPath)

	if !strings.Contains(relPath, "/publishedlanguage/contracts/") {
		return nil, nil
	}

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
	if err != nil {
		return nil, err
	}

	var violations []purityViolation
	for _, imp := range node.Imports {
		importPath := strings.Trim(imp.Path.Value, `"`)
		if strings.Contains(importPath, ".") || strings.Contains(importPath, modulePrefix) {
			violations = append(violations, purityViolation{relPath, importPath})
		}
	}
	return violations, nil
}

func (s architectureScanner) scanContractsPurity() ([]purityViolation, error) {
	var violations []purityViolation

	err := filepath.Walk(s.internalDir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !isGoSourceFile(info, path) {
			return nil
		}
		fileViolations, checkErr := s.checkContractsPurity(path)
		if checkErr != nil {
			return checkErr
		}
		violations = append(violations, fileViolations...)
		return nil
	})

	return violations, err
}

func TestNoCrossBoundedContextImports(t *testing.T) {
	internalDir, err := filepath.Abs(".")
	if err != nil {
		t.Fatalf("failed to get absolute path: %v", err)
	}

	scanner := architectureScanner{internalDir: internalDir}
	scan, err := scanner.scanImports()
	if err != nil {
		t.Fatalf("failed to scan imports: %v", err)
	}

	for _, v := range scan.violations {
		t.Errorf("CROSS-BC VIOLATION: %s imports %s (from %s, only publishedlanguage allowed)", v.relPath, v.importSuffix, v.importedBC)
	}

	t.Run("NoStaleAllowlistEntries", func(t *testing.T) {
		for entry, spec := range allowedCrossBCImports {
			if !scan.usedAllowlistEntries[entry] {
				t.Errorf("STALE ALLOWLIST ENTRY: %q (was for %s) — violation no longer exists, remove this entry", entry, spec)
			}
		}
	})
}

func TestPublishedLanguageContractsPurity(t *testing.T) {
	internalDir, err := filepath.Abs(".")
	if err != nil {
		t.Fatalf("failed to get absolute path: %v", err)
	}

	scanner := architectureScanner{internalDir: internalDir}
	violations, err := scanner.scanContractsPurity()
	if err != nil {
		t.Fatalf("failed to scan contracts purity: %v", err)
	}

	for _, v := range violations {
		t.Errorf("PURITY VIOLATION: %s imports %q — contracts packages must only use stdlib", v.relPath, v.importPath)
	}
}
