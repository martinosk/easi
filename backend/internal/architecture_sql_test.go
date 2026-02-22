//go:build !integration

// SQL Schema Guardrail
//
// Scanned locations (approvedLocationPatterns):
//   - */application/readmodels/*.go    — read model query definitions
//   - */application/projectors/*.go    — event projectors that write to read models
//   - */infrastructure/repositories/*.go — aggregate persistence
//   - */infrastructure/repository/*.go  — aggregate persistence (singular variant)
//   - */infrastructure/eventstore/*.go  — event store implementations
//   - */infrastructure/adapters/*.go    — infrastructure adapters
//   - */infrastructure/metamodel/*.go   — cross-BC metamodel gateways
//
// Intentionally excluded:
//   - shared/, infrastructure/, testing/ — shared packages have no BC ownership
//   - Domain layer (domain/) — must never contain SQL
//   - Application services — orchestrate via repositories, never issue SQL directly
//   - Migration files — schema DDL, not runtime queries
//
// How to extend:
//   Add a new glob pattern to approvedLocationPatterns. Both TestSQLSchemaOwnership
//   (positive ownership scan) and TestNoSQLOutsideApprovedLocations (negative guard)
//   will automatically pick it up.

package internal_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

var allowedSchemaAccess = map[string]string{
	"auth -> platform":               "tenant domain checking for authentication",
	"capabilitymapping -> metamodel": "strategy pillar configuration gateway",
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

var sqlStatementPattern = regexp.MustCompile(`(?i)(\bSELECT\b.*\bFROM\b|\bINSERT\s+INTO\b|\bUPDATE\b.*\bSET\b|\bDELETE\s+FROM\b)`)

var sqlKeywords = map[string]bool{
	"select": true, "set": true, "where": true, "values": true,
	"not": true, "null": true, "exists": true, "only": true,
	"table": true, "index": true, "if": true, "as": true,
}

var approvedLocationPatterns = []string{
	"*/application/readmodels/*.go",
	"*/application/projectors/*.go",
	"*/infrastructure/repositories/*.go",
	"*/infrastructure/repository/*.go",
	"*/infrastructure/eventstore/*.go",
	"*/infrastructure/adapters/*.go",
	"*/infrastructure/metamodel/*.go",
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

func (a sqlAnalyzer) looksLikeSQL() bool {
	return sqlStatementPattern.MatchString(a.sql)
}

func (a sqlAnalyzer) unqualifiedTables() []string {
	known := a.knownIdentifiers()
	if len(known) == 0 && !a.looksLikeSQL() {
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

func unquoteStringLiteral(lit *ast.BasicLit) (string, bool) {
	val := lit.Value
	if strings.HasPrefix(val, "`") && strings.HasSuffix(val, "`") {
		return val[1 : len(val)-1], true
	}
	if unquoted, err := strconv.Unquote(val); err == nil {
		return unquoted, true
	}
	return "", false
}

func extractStringLiterals(node ast.Node) []string {
	var literals []string
	ast.Inspect(node, func(n ast.Node) bool {
		lit, ok := n.(*ast.BasicLit)
		if !ok || lit.Kind.String() != "STRING" {
			return true
		}
		if s, ok := unquoteStringLiteral(lit); ok {
			literals = append(literals, s)
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
	var files []string
	for _, pattern := range approvedLocationPatterns {
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

func (a sqlAnalyzer) containsSchemaQualifiedRef() bool {
	for _, re := range schemaQualifiedPatterns {
		if re.MatchString(a.sql) {
			return true
		}
	}
	return false
}

type sqlLocationGuard struct {
	internalDir   string
	approvedFiles map[string]bool
}

func (s *schemaOwnershipScanner) locationGuard() sqlLocationGuard {
	approved := make(map[string]bool)
	for _, pattern := range approvedLocationPatterns {
		matches, err := filepath.Glob(filepath.Join(s.internalDir, pattern))
		if err != nil {
			continue
		}
		for _, m := range matches {
			approved[filepath.Clean(m)] = true
		}
	}
	return sqlLocationGuard{internalDir: s.internalDir, approvedFiles: approved}
}

func fileContainsSchemaQualifiedSQL(path string) bool {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return false
	}
	for _, literal := range extractStringLiterals(node) {
		if (sqlAnalyzer{sql: literal}).containsSchemaQualifiedRef() {
			return true
		}
	}
	return false
}

func (g *sqlLocationGuard) countSQLFiles() int {
	count := 0
	for path := range g.approvedFiles {
		if fileContainsSchemaQualifiedSQL(path) {
			count++
		}
	}
	return count
}

func (g *sqlLocationGuard) checkFile(path string) (string, bool) {
	relPath, err := filepath.Rel(g.internalDir, path)
	if err != nil {
		return "", false
	}
	ownerBC := strings.SplitN(filepath.ToSlash(relPath), "/", 2)[0]
	if sharedPackages[ownerBC] || g.approvedFiles[filepath.Clean(path)] {
		return "", false
	}

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return "", false
	}

	for _, literal := range extractStringLiterals(node) {
		if (sqlAnalyzer{sql: literal}).containsSchemaQualifiedRef() {
			return filepath.ToSlash(relPath), true
		}
	}
	return "", false
}

func mustInternalDir(t *testing.T) string {
	t.Helper()
	dir, err := filepath.Abs(".")
	if err != nil {
		t.Fatalf("failed to get absolute path: %v", err)
	}
	return dir
}

type unqualifiedTableScanResult struct {
	violations   []sqlViolation
	filesScanned int
}

func (s *schemaOwnershipScanner) scanAllForUnqualifiedTables() (unqualifiedTableScanResult, error) {
	var result unqualifiedTableScanResult

	err := filepath.Walk(s.internalDir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() && info.Name() == "migrations" {
			return filepath.SkipDir
		}
		if info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		result.filesScanned++
		result.violations = append(result.violations, s.findUnqualifiedTables(path)...)
		return nil
	})

	return result, err
}

func (s *schemaOwnershipScanner) findUnqualifiedTables(path string) []sqlViolation {
	relPath, err := filepath.Rel(s.internalDir, path)
	if err != nil {
		return nil
	}

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil
	}

	var violations []sqlViolation
	for _, literal := range extractStringLiterals(node) {
		analyzer := sqlAnalyzer{sql: literal}
		for _, name := range analyzer.unqualifiedTables() {
			violations = append(violations, sqlViolation{
				relPath: filepath.ToSlash(relPath),
				message: "unqualified table reference '" + name + "' — must use schema.table format",
			})
		}
	}
	return violations
}

func TestAllGoFilesUseSchemaQualifiedSQL(t *testing.T) {
	scanner := &schemaOwnershipScanner{internalDir: mustInternalDir(t)}
	result, err := scanner.scanAllForUnqualifiedTables()
	if err != nil {
		t.Fatalf("failed to walk internal dir: %v", err)
	}

	for _, v := range result.violations {
		t.Errorf("UNQUALIFIED TABLE: %s — %s", v.relPath, v.message)
	}

	t.Run("SmokeCheck", func(t *testing.T) {
		if result.filesScanned == 0 {
			t.Error("scanned 0 files — detection mechanism may be broken")
		}
	})
}

func TestSQLSchemaOwnership(t *testing.T) {
	scanner := &schemaOwnershipScanner{internalDir: mustInternalDir(t)}

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

func TestNoSQLOutsideApprovedLocations(t *testing.T) {
	scanner := &schemaOwnershipScanner{internalDir: mustInternalDir(t)}
	guard := scanner.locationGuard()
	var violations []string

	walkErr := filepath.Walk(scanner.internalDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !isProductionGoFile(info, path) {
			return nil
		}
		if relPath, found := guard.checkFile(path); found {
			violations = append(violations, relPath)
		}
		return nil
	})
	if walkErr != nil {
		t.Fatalf("failed to walk internal dir: %v", walkErr)
	}

	for _, v := range violations {
		t.Errorf("SQL OUTSIDE APPROVED LOCATION: %s — move SQL to an approved location or extend approvedLocationPatterns", v)
	}

	t.Run("DetectionSmokeCheck", func(t *testing.T) {
		if guard.countSQLFiles() == 0 {
			t.Error("location guard detected 0 SQL-containing files in approved locations — detection mechanism may be broken")
		}
	})
}
