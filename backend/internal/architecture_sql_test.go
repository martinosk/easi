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

var tableOwnership = map[string]string{
	"events":    "infrastructure",
	"snapshots": "infrastructure",

	"sessions": "shared",

	"application_components":        "architecturemodeling",
	"component_relations":           "architecturemodeling",
	"application_component_experts": "architecturemodeling",
	"acquired_entities":             "architecturemodeling",
	"vendors":                       "architecturemodeling",
	"internal_teams":                "architecturemodeling",
	"acquired_via_relationships":    "architecturemodeling",
	"purchased_from_relationships":  "architecturemodeling",
	"built_by_relationships":        "architecturemodeling",

	"architecture_views":       "architectureviews",
	"view_element_positions":   "architectureviews",
	"view_component_positions": "architectureviews",
	"view_preferences":         "architectureviews",

	"capabilities":                   "capabilitymapping",
	"capability_dependencies":        "capabilitymapping",
	"capability_realizations":        "capabilitymapping",
	"capability_experts":             "capabilitymapping",
	"capability_tags":                "capabilitymapping",
	"capability_component_cache":     "capabilitymapping",
	"domain_capability_assignments":  "capabilitymapping",
	"effective_capability_importance": "capabilitymapping",
	"application_fit_scores":         "capabilitymapping",
	"cm_strategy_pillar_cache":       "capabilitymapping",
	"strategy_importance":            "capabilitymapping",
	"domain_composition_view":        "capabilitymapping",
	"business_domains":               "capabilitymapping",

	"enterprise_capabilities":        "enterprisearchitecture",
	"enterprise_capability_links":    "enterprisearchitecture",
	"enterprise_strategic_importance": "enterprisearchitecture",
	"domain_capability_metadata":     "enterprisearchitecture",
	"capability_link_blocking":       "enterprisearchitecture",
	"ea_strategy_pillar_cache":       "enterprisearchitecture",

	"layout_containers": "viewlayouts",
	"element_positions": "viewlayouts",

	"import_sessions": "importing",

	"tenants":             "platform",
	"tenant_domains":      "platform",
	"tenant_oidc_configs": "platform",
	"users":               "auth",
	"invitations":         "auth",

	"edit_grants": "accessdelegation",

	"meta_model_configurations": "metamodel",

	"releases": "releases",

	"value_streams":                   "valuestreams",
	"value_stream_stages":             "valuestreams",
	"value_stream_stage_capabilities": "valuestreams",
	"value_stream_capability_cache":   "valuestreams",
}

var allowedSQLCrossAccess = map[string]string{
	"enterprisearchitecture/maturity_analysis_read_model.go -> capabilities":                           "spec-136",
	"enterprisearchitecture/time_suggestion_read_model.go -> capability_realizations":                  "spec-136",
	"enterprisearchitecture/time_suggestion_read_model.go -> capabilities":                             "spec-136",
	"enterprisearchitecture/time_suggestion_read_model.go -> effective_capability_importance":           "spec-136",
	"enterprisearchitecture/time_suggestion_read_model.go -> application_fit_scores":                   "spec-136",
	"enterprisearchitecture/domain_capability_metadata_read_model.go -> domain_capability_assignments": "spec-136",
	"capabilitymapping/strategic_fit_analysis_read_model.go -> domain_capability_metadata":             "spec-137",
	"auth/tenant_domain_checker.go -> tenant_domains":                                                 "spec-138",
}

var sqlTablePatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\bFROM\s+(\w+)`),
	regexp.MustCompile(`(?i)\bJOIN\s+(\w+)`),
	regexp.MustCompile(`(?i)\bINSERT\s+INTO\s+(\w+)`),
	regexp.MustCompile(`(?i)\bUPDATE\s+(\w+)`),
	regexp.MustCompile(`(?i)\bDELETE\s+FROM\s+(\w+)`),
}

var sqlKeywords = map[string]bool{
	"select": true, "set": true, "where": true, "values": true,
	"not": true, "null": true, "exists": true, "only": true,
	"table": true, "index": true, "if": true, "as": true,
}

func extractTablesFromSQL(sql string) map[string]bool {
	tables := make(map[string]bool)
	for _, re := range sqlTablePatterns {
		for _, match := range re.FindAllStringSubmatch(sql, -1) {
			table := strings.ToLower(match[1])
			if !sqlKeywords[table] {
				tables[table] = true
			}
		}
	}
	return tables
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
	relPath    string
	table      string
	tableOwner string
	ownerBC    string
}

type sqlScanResult struct {
	violations           []sqlViolation
	usedAllowlistEntries map[string]bool
	errors               []string
}

type tableOwnershipScanner struct {
	internalDir string
}

func (s *tableOwnershipScanner) collectFiles() ([]string, error) {
	patterns := []string{
		"*/application/readmodels/*.go",
		"*/application/projectors/*.go",
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

func (s *tableOwnershipScanner) checkFile(path string) ([]sqlViolation, map[string]bool, error) {
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
	node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, nil, err
	}

	return s.findTableViolations(node, relPath, ownerBC)
}

func (s *tableOwnershipScanner) findTableViolations(node *ast.File, relPath, ownerBC string) ([]sqlViolation, map[string]bool, error) {
	fileName := filepath.Base(relPath)
	var violations []sqlViolation
	usedEntries := make(map[string]bool)

	for _, literal := range extractStringLiterals(node) {
		for table := range extractTablesFromSQL(literal) {
			if !isCrossBCTableAccess(table, ownerBC) {
				continue
			}

			allowlistKey := ownerBC + "/" + fileName + " -> " + table
			if _, ok := allowedSQLCrossAccess[allowlistKey]; ok {
				usedEntries[allowlistKey] = true
				continue
			}

			violations = append(violations, sqlViolation{relPath, table, tableOwnership[table], ownerBC})
		}
	}

	return violations, usedEntries, nil
}

func isCrossBCTableAccess(table, ownerBC string) bool {
	tableOwner, known := tableOwnership[table]
	if !known {
		return false
	}
	if tableOwner == "infrastructure" || tableOwner == "shared" {
		return false
	}
	return tableOwner != ownerBC
}

func (s *tableOwnershipScanner) scan(files []string) sqlScanResult {
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

	scanner := &tableOwnershipScanner{internalDir: internalDir}

	files, err := scanner.collectFiles()
	if err != nil {
		t.Fatalf("failed to collect read model files: %v", err)
	}

	scan := scanner.scan(files)

	for _, e := range scan.errors {
		t.Errorf("failed to check: %s", e)
	}
	for _, v := range scan.violations {
		t.Errorf("SQL CROSS-BC VIOLATION: %s references table %s (owned by %s, file is in %s)", v.relPath, v.table, v.tableOwner, v.ownerBC)
	}

	t.Run("NoStaleSQLAllowlistEntries", func(t *testing.T) {
		for entry, spec := range allowedSQLCrossAccess {
			if !scan.usedAllowlistEntries[entry] {
				t.Errorf("STALE SQL ALLOWLIST ENTRY: %q (was for %s) — violation no longer exists, remove this entry", entry, spec)
			}
		}
	})
}
