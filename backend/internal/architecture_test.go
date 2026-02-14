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
	"enterprisearchitecture -> capabilitymapping/infrastructure/metamodel": "spec-135",
	"enterprisearchitecture -> auth/domain/valueobjects":                   "spec-135",
	"enterprisearchitecture -> auth/infrastructure/session":                "spec-135",
	"capabilitymapping -> auth/domain/valueobjects":                        "spec-135",
	"capabilitymapping -> auth/infrastructure/session":                     "spec-135",
	"architecturemodeling -> auth/domain/valueobjects":                     "spec-135",
	"architectureviews -> auth/domain/valueobjects":                        "spec-135",
	"architectureviews -> auth/application/readmodels":                     "spec-138",
	"accessdelegation -> architecturemodeling/application/readmodels":      "spec-138",
	"accessdelegation -> architectureviews/application/readmodels":         "spec-138",
	"accessdelegation -> capabilitymapping/application/readmodels":         "spec-138",
	"accessdelegation -> auth/application/readmodels":                      "spec-138",
	"importing -> architecturemodeling/application/commands":               "spec-138",
	"importing -> capabilitymapping/application/commands":                  "spec-138",
	"importing -> valuestreams/application/commands":                       "spec-138",
	"accessdelegation -> auth/infrastructure/api":                          "spec-135",
	"metamodel -> auth/domain/valueobjects":                                "spec-135",
	"metamodel -> auth/infrastructure/session":                             "spec-135",
	"valuestreams -> auth/domain/valueobjects":                             "spec-135",
	"viewlayouts -> auth/domain/valueobjects":                              "spec-135",
	"platform -> auth/application/commands":                                "spec-135",
	"platform -> auth/application/handlers":                                "spec-135",
	"platform -> auth/infrastructure/repositories":                         "spec-135",
}

func isProductionGoFile(info os.FileInfo, path string) bool {
	return !info.IsDir() && strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go")
}

func isAllowedCrossBCImport(ownerBC, importSuffix string) bool {
	importedBC := strings.SplitN(importSuffix, "/", 2)[0]
	if importedBC == ownerBC || sharedPackages[importedBC] {
		return true
	}
	if freeImportPackages[importSuffix] {
		return true
	}
	return strings.Contains(importSuffix, "/publishedlanguage")
}

type importViolation struct {
	relPath      string
	importSuffix string
	importedBC   string
}

func checkFileImports(path, internalDir string) ([]importViolation, map[string]bool, error) {
	relPath, err := filepath.Rel(internalDir, path)
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

type importScanResult struct {
	violations           []importViolation
	usedAllowlistEntries map[string]bool
}

func scanAllImports(internalDir string) (importScanResult, error) {
	result := importScanResult{usedAllowlistEntries: make(map[string]bool)}

	err := filepath.Walk(internalDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !isProductionGoFile(info, path) {
			return nil
		}
		violations, usedEntries, fileErr := checkFileImports(path, internalDir)
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

func TestNoCrossBoundedContextImports(t *testing.T) {
	internalDir, err := filepath.Abs(".")
	if err != nil {
		t.Fatalf("failed to get absolute path: %v", err)
	}

	scan, err := scanAllImports(internalDir)
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
