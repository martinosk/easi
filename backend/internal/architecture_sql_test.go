//go:build !integration

package internal_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var allowedSchemaAccess = map[string]string{
	"auth -> platform": "tenant domain checking for authentication",
}

var schemaQualifiedPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\bFROM\s+(\w+)\.(\w+)`),
	regexp.MustCompile(`(?i)\bJOIN\s+(\w+)\.(\w+)`),
	regexp.MustCompile(`(?i)\bINSERT\s+INTO\s+(\w+)\.(\w+)`),
	regexp.MustCompile(`(?i)\bUPDATE\s+(\w+)\.(\w+)`),
	regexp.MustCompile(`(?i)\bDELETE\s+FROM\s+(\w+)\.(\w+)`),
}

var unqualifiedPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\bFROM\s+([a-zA-Z_]\w*)`),
	regexp.MustCompile(`(?i)\bJOIN\s+([a-zA-Z_]\w*)`),
	regexp.MustCompile(`(?i)\bINSERT\s+INTO\s+([a-zA-Z_]\w*)`),
	regexp.MustCompile(`(?i)\bUPDATE\s+([a-zA-Z_]\w*)`),
	regexp.MustCompile(`(?i)\bDELETE\s+FROM\s+([a-zA-Z_]\w*)`),
}

var ctePattern = regexp.MustCompile(`(?i)\bWITH\s+(?:RECURSIVE\s+)?(\w+)\s+AS\s*\(`)

var sqlKeywords = map[string]bool{
	"select": true, "set": true, "where": true, "values": true,
	"not": true, "null": true, "exists": true, "only": true,
	"table": true, "index": true, "if": true, "as": true,
}

type schemaTableRef struct {
	schema string
	table  string
}

type fileContext struct {
	relPath string
	ownerBC string
}

type sqlAnalyzer struct {
	sql string
}

func (a sqlAnalyzer) qualifiedTables() []schemaTableRef {
	var refs []schemaTableRef
	for _, re := range schemaQualifiedPatterns {
		for _, match := range re.FindAllStringSubmatch(a.sql, -1) {
			refs = append(refs, schemaTableRef{
				schema: strings.ToLower(match[1]),
				table:  strings.ToLower(match[2]),
			})
		}
	}
	return refs
}

func (a sqlAnalyzer) knownIdentifiers() map[string]bool {
	known := make(map[string]bool)
	for _, re := range schemaQualifiedPatterns {
		for _, match := range re.FindAllStringSubmatch(a.sql, -1) {
			known[strings.ToLower(match[1])] = true
			known[strings.ToLower(match[2])] = true
		}
	}
	for _, match := range ctePattern.FindAllStringSubmatch(a.sql, -1) {
		known[strings.ToLower(match[1])] = true
	}
	return known
}

func (a sqlAnalyzer) unqualifiedTables() []string {
	known := a.knownIdentifiers()
	if len(known) == 0 {
		return nil
	}

	var unqualified []string
	for _, re := range unqualifiedPatterns {
		for _, match := range re.FindAllStringSubmatch(a.sql, -1) {
			table := strings.ToLower(match[1])
			if !sqlKeywords[table] && !known[table] {
				unqualified = append(unqualified, table)
			}
		}
	}
	return unqualified
}

func extractStringLiterals(node ast.Node) []string {
	var literals []string
	ast.Inspect(node, func(n ast.Node) bool {
		lit, ok := n.(*ast.BasicLit)
		if !ok {
			return true
		}
		if lit.Kind.String() != "STRING" {
			return true
		}
		val := lit.Value
		if strings.HasPrefix(val, "`") && strings.HasSuffix(val, "`") {
			literals = append(literals, val[1:len(val)-1])
		} else if strings.HasPrefix(val, `"`) && strings.HasSuffix(val, `"`) {
			literals = append(literals, val[1:len(val)-1])
		}
		return true
	})
	return literals
}

type sqlViolation struct {
	relPath string
	message string
}

type sqlScanResult struct {
	violations           []sqlViolation
	usedAllowlistEntries map[string]bool
	errors               []string
}

type schemaOwnershipScanner struct {
	internalDir string
}

func (s *schemaOwnershipScanner) collectFiles() ([]string, error) {
	patterns := []string{
		"*/application/readmodels/*.go",
		"*/application/projectors/*.go",
		"*/infrastructure/repositories/*.go",
		"*/infrastructure/repository/*.go",
		"*/infrastructure/eventstore/*.go",
	}
	var files []string
	for _, pattern := range patterns {
		matches, err := filepath.Glob(filepath.Join(s.internalDir, pattern))
		if err != nil {
			return nil, err
		}
		files = append(files, matches...)
	}
	return filterProductionFiles(files), nil
}

func filterProductionFiles(paths []string) []string {
	var result []string
	for _, p := range paths {
		if !strings.HasSuffix(p, "_test.go") {
			result = append(result, p)
		}
	}
	return result
}

func (s *schemaOwnershipScanner) checkFile(path string) ([]sqlViolation, map[string]bool, error) {
	relPath, err := filepath.Rel(s.internalDir, path)
	if err != nil {
		return nil, nil, err
	}

	fc := fileContext{
		relPath: filepath.ToSlash(relPath),
		ownerBC: strings.SplitN(filepath.ToSlash(relPath), "/", 2)[0],
	}
	if sharedPackages[fc.ownerBC] {
		return nil, nil, nil
	}

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, nil, err
	}

	return s.findSchemaViolations(node, fc)
}

func (s *schemaOwnershipScanner) findSchemaViolations(node *ast.File, fc fileContext) ([]sqlViolation, map[string]bool, error) {
	var violations []sqlViolation
	usedEntries := make(map[string]bool)

	for _, literal := range extractStringLiterals(node) {
		analyzer := sqlAnalyzer{sql: literal}

		for _, name := range analyzer.unqualifiedTables() {
			violations = append(violations, sqlViolation{
				relPath: fc.relPath,
				message: "unqualified table reference '" + name + "' — must use schema.table format",
			})
		}

		v, used := checkCrossBCAccess(analyzer.qualifiedTables(), fc)
		violations = append(violations, v...)
		for k := range used {
			usedEntries[k] = true
		}
	}

	return violations, usedEntries, nil
}

func checkCrossBCAccess(refs []schemaTableRef, fc fileContext) ([]sqlViolation, map[string]bool) {
	var violations []sqlViolation
	usedEntries := make(map[string]bool)
	ownedSchemas := map[string]bool{"infrastructure": true, "shared": true, fc.ownerBC: true}

	for _, ref := range refs {
		if ownedSchemas[ref.schema] {
			continue
		}

		allowlistKey := fc.ownerBC + " -> " + ref.schema
		if _, ok := allowedSchemaAccess[allowlistKey]; ok {
			usedEntries[allowlistKey] = true
			continue
		}

		violations = append(violations, sqlViolation{
			relPath: fc.relPath,
			message: "cross-BC schema access: references " + ref.schema + "." + ref.table + " (file is in " + fc.ownerBC + ")",
		})
	}

	return violations, usedEntries
}

func (s *schemaOwnershipScanner) scan(files []string) sqlScanResult {
	result := sqlScanResult{usedAllowlistEntries: make(map[string]bool)}
	for _, path := range files {
		violations, usedEntries, err := s.checkFile(path)
		if err != nil {
			result.errors = append(result.errors, path+": "+err.Error())
			continue
		}
		result.violations = append(result.violations, violations...)
		for k := range usedEntries {
			result.usedAllowlistEntries[k] = true
		}
	}
	return result
}

func TestReadModelsOnlyReferenceOwnedTables(t *testing.T) {
	internalDir, err := filepath.Abs(".")
	if err != nil {
		t.Fatalf("failed to get absolute path: %v", err)
	}

	scanner := &schemaOwnershipScanner{internalDir: internalDir}

	files, err := scanner.collectFiles()
	if err != nil {
		t.Fatalf("failed to collect files: %v", err)
	}

	scan := scanner.scan(files)

	for _, e := range scan.errors {
		t.Errorf("failed to check: %s", e)
	}
	for _, v := range scan.violations {
		t.Errorf("SQL OWNERSHIP VIOLATION: %s — %s", v.relPath, v.message)
	}

	t.Run("NoStaleSchemaAllowlistEntries", func(t *testing.T) {
		for entry, reason := range allowedSchemaAccess {
			if !scan.usedAllowlistEntries[entry] {
				t.Errorf("STALE SCHEMA ALLOWLIST ENTRY: %q (reason: %s) — violation no longer exists, remove this entry", entry, reason)
			}
		}
	})
}
