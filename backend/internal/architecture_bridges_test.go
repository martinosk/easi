//go:build !integration

package internal_test

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

type contextName string

type sourceFile string

type importPath string

type bridgeEdge struct {
	Consumer contextName
	Supplier contextName
	Why      string
}

var declaredBridges = map[sourceFile][]bridgeEdge{
	"infrastructure/api/accessdelegation_bridges.go": {
		{Consumer: "accessdelegation", Supplier: "capabilitymapping", Why: "display names of capability and business-domain grant artifacts"},
		{Consumer: "accessdelegation", Supplier: "architecturemodeling", Why: "display names of component, vendor, acquired-entity and team grant artifacts"},
		{Consumer: "accessdelegation", Supplier: "architectureviews", Why: "display names of view grant artifacts"},
		{Consumer: "accessdelegation", Supplier: "auth", Why: "user existence, pending-invitation and domain-allowlist checks, and invitation requests for grantees without an account"},
	},
	"infrastructure/api/architecturedirection_bridges.go": {
		{Consumer: "architecturedirection", Supplier: "capabilitymapping", Why: "capability and business-domain existence, direct realization lookup, effective-domain check for journeys"},
		{Consumer: "architecturedirection", Supplier: "architecturemodeling", Why: "application component existence for standard applications and journeys"},
	},
	"infrastructure/api/architectureviews_bridges.go": {
		{Consumer: "architectureviews", Supplier: "auth", Why: "user role check when changing view visibility"},
	},
	"infrastructure/api/enterprisearchitecture_bridges.go": {
		{Consumer: "enterprisearchitecture", Supplier: "capabilitymapping", Why: "business-domain name at capability assignment time"},
	},
	"infrastructure/api/importing_bridges.go": {
		{Consumer: "importing", Supplier: "architecturemodeling", Why: "import gateway creating components and relations"},
		{Consumer: "importing", Supplier: "capabilitymapping", Why: "import gateway creating capabilities, realizations and domain assignments"},
		{Consumer: "importing", Supplier: "valuestreams", Why: "import gateway creating value streams and stages"},
	},
	"infrastructure/api/onepager_builtin_field_adapters.go": {
		{Consumer: "onepagers", Supplier: "architecturemodeling", Why: "built-in field values of application, vendor, acquired-entity and team subjects"},
		{Consumer: "onepagers", Supplier: "capabilitymapping", Why: "built-in field values of capability subjects"},
		{Consumer: "onepagers", Supplier: "enterprisearchitecture", Why: "built-in field values of enterprise-capability subjects"},
		{Consumer: "onepagers", Supplier: "metamodel", Why: "maturity scale sections for rendering maturity fields"},
	},
	"infrastructure/api/onepager_relation_adapters.go": {
		{Consumer: "onepagers", Supplier: "architecturemodeling", Why: "relation fields: component relations and origin links"},
		{Consumer: "onepagers", Supplier: "capabilitymapping", Why: "relation fields: realizations, dependencies and domain assignments"},
		{Consumer: "onepagers", Supplier: "enterprisearchitecture", Why: "relation fields: enterprise capability names"},
		{Consumer: "onepagers", Supplier: "architecturedirection", Why: "relation fields: enterprise capability composition"},
	},
	"infrastructure/api/onepager_subject_adapters.go": {
		{Consumer: "onepagers", Supplier: "architecturemodeling", Why: "subject existence for application, vendor, acquired-entity and team one-pagers"},
		{Consumer: "onepagers", Supplier: "capabilitymapping", Why: "subject existence for capability one-pagers"},
		{Consumer: "onepagers", Supplier: "enterprisearchitecture", Why: "subject existence for enterprise-capability one-pagers"},
	},
}

const (
	compositionRootDir            = "infrastructure/api"
	routerFile         sourceFile = "infrastructure/api/router.go"
	testSupportPackage            = "testing"
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

func (s contextSet) equals(other contextSet) bool {
	if len(s) != len(other) {
		return false
	}
	for name := range s {
		if !other[name] {
			return false
		}
	}
	return true
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

func (imp importPath) isPublishedLanguage() bool {
	return strings.Contains(string(imp), "/publishedlanguage")
}

func (imp importPath) isTestSupport() bool {
	return string(imp) == testSupportPackage || strings.HasPrefix(string(imp), testSupportPackage+"/")
}

func (file sourceFile) inCompositionRoot() bool {
	return filepath.ToSlash(filepath.Dir(string(file))) == compositionRootDir
}

func (file sourceFile) inNonContextPackage() bool {
	owner := firstSegment(string(file))
	return sharedPackages[owner] && owner != testSupportPackage
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

type compositionRootScan struct {
	reached          map[sourceFile]contextSet
	routerViolations []importPath
	leaks            []string
	testSupportUses  []string
}

func (scan *compositionRootScan) record(file sourceFile, imp importPath) {
	if imp.isTestSupport() {
		scan.testSupportUses = append(scan.testSupportUses, string(file)+" imports "+string(imp))
	}
	ctx, isContext := contextOf(imp)
	if !isContext {
		return
	}
	switch {
	case file == routerFile:
		scan.recordRouterImport(imp)
	case file.inCompositionRoot():
		scan.recordBridgeImport(file, ctx)
	case file.inNonContextPackage():
		scan.recordNonContextImport(file, imp)
	}
}

func (scan *compositionRootScan) recordRouterImport(imp importPath) {
	if !imp.isRoutesPackage() {
		scan.routerViolations = append(scan.routerViolations, imp)
	}
}

func (scan *compositionRootScan) recordBridgeImport(file sourceFile, ctx contextName) {
	if scan.reached[file] == nil {
		scan.reached[file] = contextSet{}
	}
	scan.reached[file][ctx] = true
}

func (scan *compositionRootScan) recordNonContextImport(file sourceFile, imp importPath) {
	if !imp.isPublishedLanguage() {
		scan.leaks = append(scan.leaks, string(file)+" imports "+string(imp))
	}
}

func scanCompositionRootOrFail(t *testing.T) compositionRootScan {
	t.Helper()
	internalDir, err := filepath.Abs(".")
	if err != nil {
		t.Fatalf("failed to get absolute path: %v", err)
	}
	scanner := architectureScanner{internalDir: internalDir}
	scan := compositionRootScan{reached: map[sourceFile]contextSet{}}
	err = scanner.walkProduction(func(file sourceFile, imports []importPath) error {
		for _, imp := range imports {
			scan.record(file, imp)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("failed to scan composition root: %v", err)
	}
	return scan
}

func sortedFiles(reached map[sourceFile]contextSet) []sourceFile {
	files := make([]sourceFile, 0, len(reached))
	for file := range reached {
		files = append(files, file)
	}
	sort.Slice(files, func(i, j int) bool { return files[i] < files[j] })
	return files
}

func declaredContexts(edges []bridgeEdge) contextSet {
	set := contextSet{}
	for _, e := range edges {
		set[e.Consumer] = true
		set[e.Supplier] = true
	}
	return set
}

func assertBridgeDeclared(t *testing.T, file sourceFile, reached contextSet) {
	t.Helper()
	if len(reached) < 2 {
		t.Errorf("SINGLE-CONTEXT ADAPTER: %s reaches only %v — an adapter of shared code for one context belongs inside that context, not in the composition root", file, reached.names())
		return
	}
	edges, declared := declaredBridges[file]
	if !declared {
		t.Errorf("UNDECLARED BRIDGE: %s reaches contexts %v — add it to declaredBridges with every consumer -> supplier edge", file, reached.names())
		return
	}
	if !declaredContexts(edges).equals(reached) {
		t.Errorf("INACCURATE BRIDGE DECLARATION: %s declares %v but imports %v", file, declaredContexts(edges).names(), reached.names())
	}
	assertEdgesWellFormed(t, file, edges)
}

func assertEdgesWellFormed(t *testing.T, file sourceFile, edges []bridgeEdge) {
	t.Helper()
	for _, e := range edges {
		if e.Consumer == e.Supplier {
			t.Errorf("INVALID BRIDGE EDGE: %s declares %s -> %s", file, e.Consumer, e.Supplier)
		}
		if strings.TrimSpace(e.Why) == "" {
			t.Errorf("UNJUSTIFIED BRIDGE EDGE: %s declares %s -> %s without a reason", file, e.Consumer, e.Supplier)
		}
	}
}

func TestRouterOnlyRegistersRoutes(t *testing.T) {
	scan := scanCompositionRootOrFail(t)
	for _, imp := range scan.routerViolations {
		t.Errorf("ROUTER VIOLATION: %s imports %s — router.go may only import a context's infrastructure/api package; move wiring into a declared *_bridges.go file", routerFile, imp)
	}
}

func TestEveryCompositionRootBridgeIsDeclaredExactly(t *testing.T) {
	scan := scanCompositionRootOrFail(t)
	for _, file := range sortedFiles(scan.reached) {
		assertBridgeDeclared(t, file, scan.reached[file])
	}
}

func TestNoStaleBridgeDeclarations(t *testing.T) {
	scan := scanCompositionRootOrFail(t)
	for file := range declaredBridges {
		if _, ok := scan.reached[file]; !ok {
			t.Errorf("STALE BRIDGE DECLARATION: %s no longer reaches any context (or does not exist) — remove its declaration", file)
		}
	}
}

func TestNonContextPackagesImportOnlyPublishedLanguage(t *testing.T) {
	scan := scanCompositionRootOrFail(t)
	for _, leak := range scan.leaks {
		t.Errorf("LEAK: %s — shared and infrastructure packages may import only a context's publishedlanguage", leak)
	}
}

func TestProductionCodeDoesNotImportTestSupport(t *testing.T) {
	scan := scanCompositionRootOrFail(t)
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
	for file, edges := range declaredBridges {
		for _, e := range edges {
			graph.add(e.Consumer, e.Supplier, "bridge "+string(file)+": "+e.Why)
		}
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
		sort.Slice(component, func(i, j int) bool { return component[i] < component[j] })
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
	internalDir, err := filepath.Abs(".")
	if err != nil {
		t.Fatalf("failed to get absolute path: %v", err)
	}
	scanner := architectureScanner{internalDir: internalDir}
	graph, err := scanner.buildDependencyGraph()
	if err != nil {
		t.Fatalf("failed to build dependency graph: %v", err)
	}
	for _, component := range graph.stronglyConnectedCycles() {
		t.Errorf("DEPENDENCY CYCLE between %s:\n%s", joinContexts(component), graph.describeCycle(component))
	}
}
