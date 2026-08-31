//go:build !integration

package internal_test

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"testing"
)

type contextName string

type sourceFile string

type importPath string

const (
	compositionRootDir       = "infrastructure/api"
	testSupportPackage       = "testing"
	migrationsDir            = "../deploy-scripts/migrations"
	firstEventsOnlyMigration = 139
)

type contextSet map[contextName]bool

func (s contextSet) names() []string {
	names := make([]string, 0, len(s))
	for name := range s {
		names = append(names, string(name))
	}
	sort.Strings(names)
	return names
}

func firstSegment(path string) string {
	return strings.SplitN(path, "/", 2)[0]
}

func contextOf(imp importPath) (contextName, bool) {
	name := firstSegment(string(imp))
	if sharedPackages[name] {
		return "", false
	}
	return contextName(name), true
}

func ownerOf(file sourceFile) (contextName, bool) {
	return contextOf(importPath(file))
}

func (imp importPath) isRoutesPackage() bool {
	parts := strings.Split(string(imp), "/")
	return len(parts) == 3 && parts[1] == "infrastructure" && parts[2] == "api"
}

func (imp importPath) isTestSupport() bool {
	return string(imp) == testSupportPackage || strings.HasPrefix(string(imp), testSupportPackage+"/")
}

func (file sourceFile) inCompositionRoot() bool {
	return strings.HasPrefix(string(file), compositionRootDir+"/")
}

func (file sourceFile) inNonContextPackage() bool {
	owner := firstSegment(string(file))
	return sharedPackages[owner] && owner != testSupportPackage && !file.inCompositionRoot()
}

func (s architectureScanner) internalImports(path string) ([]importPath, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
	if err != nil {
		return nil, err
	}
	var imports []importPath
	for _, imp := range node.Imports {
		full := strings.Trim(imp.Path.Value, `"`)
		if strings.HasPrefix(full, modulePrefix) {
			imports = append(imports, importPath(full[len(modulePrefix):]))
		}
	}
	return imports, nil
}

func (s architectureScanner) walkProduction(visit func(file sourceFile, imports []importPath) error) error {
	return filepath.Walk(s.internalDir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !isProductionGoFile(info, path) {
			return nil
		}
		relPath, err := filepath.Rel(s.internalDir, path)
		if err != nil {
			return err
		}
		imports, err := s.internalImports(path)
		if err != nil {
			return err
		}
		return visit(sourceFile(filepath.ToSlash(relPath)), imports)
	})
}

type integrationScan struct {
	compositionRootReaches []string
	nonContextReaches      []string
	testSupportUses        []string
}

func (scan *integrationScan) record(file sourceFile, imp importPath) {
	if imp.isTestSupport() {
		scan.testSupportUses = append(scan.testSupportUses, string(file)+" imports "+string(imp))
	}
	if _, isContext := contextOf(imp); !isContext {
		return
	}
	reach := string(file) + " imports " + string(imp)
	switch {
	case file.inCompositionRoot() && !imp.isRoutesPackage():
		scan.compositionRootReaches = append(scan.compositionRootReaches, reach)
	case file.inNonContextPackage():
		scan.nonContextReaches = append(scan.nonContextReaches, reach)
	}
}

func newScannerOrFail(t *testing.T) architectureScanner {
	t.Helper()
	internalDir, err := filepath.Abs(".")
	if err != nil {
		t.Fatalf("failed to get absolute path: %v", err)
	}
	return architectureScanner{internalDir: internalDir}
}

func scanIntegrationOrFail(t *testing.T) integrationScan {
	t.Helper()
	scan := integrationScan{}
	err := newScannerOrFail(t).walkProduction(func(file sourceFile, imports []importPath) error {
		for _, imp := range imports {
			scan.record(file, imp)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("failed to scan production code: %v", err)
	}
	return scan
}

func TestCompositionRootOnlyRegistersRoutes(t *testing.T) {
	scan := scanIntegrationOrFail(t)
	for _, reach := range scan.compositionRootReaches {
		t.Errorf("COMPOSITION ROOT REACH: %s — the composition root may import only a context's infrastructure/api package; a cache fed by published events replaces a query-time bridge (spec 209)", reach)
	}
}

func TestSharedAndInfrastructureImportNoContext(t *testing.T) {
	scan := scanIntegrationOrFail(t)
	for _, reach := range scan.nonContextReaches {
		t.Errorf("SHARED LEAK: %s — shared and infrastructure packages may not import any context; move the code into the context that owns it (spec 209)", reach)
	}
}

func TestProductionCodeDoesNotImportTestSupport(t *testing.T) {
	scan := scanIntegrationOrFail(t)
	for _, use := range scan.testSupportUses {
		t.Errorf("TEST SUPPORT LEAK: %s", use)
	}
}

type dependencyEdge struct {
	From contextName
	To   contextName
}

type dependencyGraph struct {
	evidence map[dependencyEdge][]string
}

func (g *dependencyGraph) add(from, to contextName, why string) {
	if from == to {
		return
	}
	edge := dependencyEdge{From: from, To: to}
	g.evidence[edge] = append(g.evidence[edge], why)
}

func (g *dependencyGraph) addImportEdges(file sourceFile, imports []importPath) {
	owner, isContext := ownerOf(file)
	if !isContext {
		return
	}
	for _, imp := range imports {
		if target, ok := contextOf(imp); ok {
			g.add(owner, target, string(file)+" imports "+string(imp))
		}
	}
}

func (s architectureScanner) buildDependencyGraph() (*dependencyGraph, error) {
	graph := &dependencyGraph{evidence: map[dependencyEdge][]string{}}
	err := s.walkProduction(func(file sourceFile, imports []importPath) error {
		graph.addImportEdges(file, imports)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return graph, nil
}

func (g *dependencyGraph) nodes() []contextName {
	set := contextSet{}
	for edge := range g.evidence {
		set[edge.From] = true
		set[edge.To] = true
	}
	return sortedContexts(set)
}

func (g *dependencyGraph) successors(node contextName) []contextName {
	set := contextSet{}
	for edge := range g.evidence {
		if edge.From == node {
			set[edge.To] = true
		}
	}
	return sortedContexts(set)
}

func sortedContexts(set contextSet) []contextName {
	names := set.names()
	contexts := make([]contextName, len(names))
	for i, name := range names {
		contexts[i] = contextName(name)
	}
	return contexts
}

type sccState struct {
	index   map[contextName]int
	lowlink map[contextName]int
	onStack map[contextName]bool
	stack   []contextName
	counter int
	cycles  [][]contextName
}

func newSCCState() *sccState {
	return &sccState{index: map[contextName]int{}, lowlink: map[contextName]int{}, onStack: map[contextName]bool{}}
}

func (st *sccState) visited(node contextName) bool {
	_, ok := st.index[node]
	return ok
}

func (st *sccState) push(node contextName) {
	st.index[node] = st.counter
	st.lowlink[node] = st.counter
	st.counter++
	st.stack = append(st.stack, node)
	st.onStack[node] = true
}

func (st *sccState) popComponent(root contextName) []contextName {
	var component []contextName
	for {
		top := st.stack[len(st.stack)-1]
		st.stack = st.stack[:len(st.stack)-1]
		st.onStack[top] = false
		component = append(component, top)
		if top == root {
			return component
		}
	}
}

func (g *dependencyGraph) stronglyConnectedCycles() [][]contextName {
	state := newSCCState()
	for _, node := range g.nodes() {
		if !state.visited(node) {
			g.strongConnect(node, state)
		}
	}
	return state.cycles
}

func (g *dependencyGraph) strongConnect(node contextName, st *sccState) {
	st.push(node)
	for _, next := range g.successors(node) {
		if !st.visited(next) {
			g.strongConnect(next, st)
			st.lowlink[node] = minInt(st.lowlink[node], st.lowlink[next])
		} else if st.onStack[next] {
			st.lowlink[node] = minInt(st.lowlink[node], st.index[next])
		}
	}
	if st.lowlink[node] != st.index[node] {
		return
	}
	component := st.popComponent(node)
	if len(component) > 1 {
		slices.Sort(component)
		st.cycles = append(st.cycles, component)
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (g *dependencyGraph) describeCycle(component []contextName) string {
	members := contextSet{}
	for _, node := range component {
		members[node] = true
	}
	edges := make([]dependencyEdge, 0, len(g.evidence))
	for edge := range g.evidence {
		if members[edge.From] && members[edge.To] {
			edges = append(edges, edge)
		}
	}
	sort.Slice(edges, func(i, j int) bool {
		if edges[i].From != edges[j].From {
			return edges[i].From < edges[j].From
		}
		return edges[i].To < edges[j].To
	})
	var lines []string
	for _, edge := range edges {
		evidence := g.evidence[edge]
		sort.Strings(evidence)
		lines = append(lines, fmt.Sprintf("  %s -> %s\n      %s", edge.From, edge.To, strings.Join(evidence, "\n      ")))
	}
	return strings.Join(lines, "\n")
}

func joinContexts(component []contextName) string {
	names := make([]string, len(component))
	for i, c := range component {
		names[i] = string(c)
	}
	return strings.Join(names, ", ")
}

func TestContextDependencyGraphIsAcyclic(t *testing.T) {
	graph, err := newScannerOrFail(t).buildDependencyGraph()
	if err != nil {
		t.Fatalf("failed to build dependency graph: %v", err)
	}
	for _, component := range graph.stronglyConnectedCycles() {
		t.Errorf("DEPENDENCY CYCLE between %s:\n%s", joinContexts(component), graph.describeCycle(component))
	}
}

func migrationNumber(name string) (int, bool) {
	number, err := strconv.Atoi(strings.SplitN(name, "_", 2)[0])
	return number, err == nil
}

func isBackfill(name string) bool {
	return strings.Contains(name, "backfill")
}

func isEventsOnlyMigration(name string) bool {
	number, numbered := migrationNumber(name)
	return numbered && number >= firstEventsOnlyMigration
}

func schemaNamesOrFail(t *testing.T) []string {
	t.Helper()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("failed to list internal packages: %v", err)
	}
	var names []string
	for _, entry := range entries {
		if entry.IsDir() && entry.Name() != testSupportPackage {
			names = append(names, entry.Name())
		}
	}
	return names
}

var sqlLineComment = regexp.MustCompile(`--[^\n]*`)

func referencedSchemas(sql string, schemas []string) []string {
	pattern := regexp.MustCompile(`\b(` + strings.Join(schemas, "|") + `)"?\."?[a-z_]+"?`)
	found := contextSet{}
	for _, match := range pattern.FindAllStringSubmatch(sqlLineComment.ReplaceAllString(sql, ""), -1) {
		found[contextName(match[1])] = true
	}
	return found.names()
}

var permanentSchemaObjectPattern = regexp.MustCompile(`(?i)\bCREATE\s+(?:OR\s+REPLACE\s+)?(?:MATERIALIZED\s+VIEW|VIEW|FUNCTION|TRIGGER)\b`)

func permanentSchemaObjects(sql string) []string {
	return permanentSchemaObjectPattern.FindAllString(sqlLineComment.ReplaceAllString(sql, ""), -1)
}

func TestReferencedSchemasHandlesQuotedIdentifiers(t *testing.T) {
	schemas := []string{"auth", "capabilitymapping"}
	tests := []struct {
		name string
		sql  string
		want []string
	}{
		{name: "BareIdentifiers", sql: `SELECT * FROM auth.users`, want: []string{"auth"}},
		{name: "BothIdentifiersQuoted", sql: `SELECT * FROM "auth"."users"`, want: []string{"auth"}},
		{name: "SchemaUnquotedTableQuoted", sql: `SELECT * FROM auth."users"`, want: []string{"auth"}},
		{name: "SchemaQuotedTableUnquoted", sql: `SELECT * FROM "auth".users`, want: []string{"auth"}},
		{name: "TwoSchemasOneQuoted", sql: `INSERT INTO capabilitymapping.capabilities SELECT * FROM "auth"."users"`, want: []string{"auth", "capabilitymapping"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := referencedSchemas(tt.sql, schemas)
			if !slices.Equal(got, tt.want) {
				t.Errorf("referencedSchemas() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPermanentSchemaObjectsDetectsViewsFunctionsAndTriggers(t *testing.T) {
	tests := []struct {
		name string
		sql  string
		want int
	}{
		{name: "PlainInsertSelectIsNotFlagged", sql: `INSERT INTO auth.tenant_cache (id) SELECT id FROM platform.tenants ON CONFLICT (id) DO UPDATE SET id = EXCLUDED.id;`, want: 0},
		{name: "CreateView", sql: `CREATE VIEW auth.combined AS SELECT * FROM platform.tenants;`, want: 1},
		{name: "CreateOrReplaceView", sql: `CREATE OR REPLACE VIEW auth.combined AS SELECT * FROM platform.tenants;`, want: 1},
		{name: "CreateMaterializedView", sql: `CREATE MATERIALIZED VIEW auth.combined AS SELECT * FROM platform.tenants;`, want: 1},
		{name: "CreateFunction", sql: `CREATE FUNCTION auth.sync_tenant() RETURNS trigger AS $$ BEGIN RETURN NEW; END; $$ LANGUAGE plpgsql;`, want: 1},
		{name: "CreateTrigger", sql: `CREATE TRIGGER sync_tenant AFTER INSERT ON platform.tenants FOR EACH ROW EXECUTE FUNCTION auth.sync_tenant();`, want: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := permanentSchemaObjects(tt.sql); len(got) != tt.want {
				t.Errorf("permanentSchemaObjects() = %v, want %d matches", got, tt.want)
			}
		})
	}
}

func TestNewMigrationsCrossSchemasOnlyInBackfills(t *testing.T) {
	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
	if err != nil {
		t.Fatalf("failed to list migrations: %v", err)
	}
	schemas := schemaNamesOrFail(t)
	for _, path := range files {
		name := filepath.Base(path)
		if !isEventsOnlyMigration(name) {
			continue
		}
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("failed to read %s: %v", name, err)
		}
		if isBackfill(name) {
			for _, stmt := range permanentSchemaObjects(string(content)) {
				t.Errorf("BACKFILL CREATES PERMANENT OBJECT: %s declares %q — a backfill may only seed data with INSERT/UPDATE, never create a permanent view, function or trigger (spec 209)", name, stmt)
			}
			continue
		}
		if referenced := referencedSchemas(string(content), schemas); len(referenced) > 1 {
			t.Errorf("CROSS-SCHEMA MIGRATION: %s references schemas %v — only *backfill*.sql migrations may read another context's schema (spec 209)", name, referenced)
		}
	}
}
