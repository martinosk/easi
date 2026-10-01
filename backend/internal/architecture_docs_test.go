//go:build !integration

package internal_test

import (
	"bufio"
	"go/ast"
	"go/parser"
	"go/token"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
)

const crossContextEventsDoc = "../../docs/backend/cross-context-events.md"

type eventConstants map[string]map[string]bool

func (c eventConstants) add(context string, names []string) {
	if c[context] == nil {
		c[context] = map[string]bool{}
	}
	for _, name := range names {
		c[context][name] = true
	}
}

func (c eventConstants) sortedContexts() []string {
	return slices.Sorted(maps.Keys(c))
}

func (c eventConstants) missingFrom(other eventConstants) map[string][]string {
	missing := map[string][]string{}
	for context, names := range c {
		for name := range names {
			if !other[context][name] {
				missing[context] = append(missing[context], name)
			}
		}
		slices.Sort(missing[context])
	}
	return missing
}

func contextPackage(displayName string) string {
	return strings.ToLower(strings.ReplaceAll(strings.TrimSpace(displayName), " ", ""))
}

func stringConstants(file *ast.File) []string {
	var names []string
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.CONST {
			continue
		}
		for _, spec := range gen.Specs {
			names = append(names, stringConstantNames(spec.(*ast.ValueSpec))...)
		}
	}
	return names
}

func stringConstantNames(spec *ast.ValueSpec) []string {
	var names []string
	for i, name := range spec.Names {
		if i < len(spec.Values) && isStringLiteral(spec.Values[i]) {
			names = append(names, name.Name)
		}
	}
	return names
}

func isStringLiteral(expr ast.Expr) bool {
	literal, ok := expr.(*ast.BasicLit)
	return ok && literal.Kind == token.STRING
}

func publishedEventConstants(t *testing.T) eventConstants {
	t.Helper()
	files, err := filepath.Glob("*/publishedlanguage/*events.go")
	if err != nil {
		t.Fatalf("glob published language event files: %v", err)
	}
	constants := eventConstants{}
	for _, path := range files {
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
		constants.add(strings.Split(filepath.ToSlash(path), "/")[0], stringConstants(file))
	}
	return constants
}

type markdownLine string

func (l markdownLine) isTableRow() bool {
	return strings.HasPrefix(string(l), "|")
}

func (l markdownLine) cells() []string {
	cells := strings.Split(strings.Trim(strings.TrimSpace(string(l)), "|"), "|")
	for i := range cells {
		cells[i] = strings.TrimSpace(cells[i])
	}
	return cells
}

func (l markdownLine) boldLabel() (string, bool) {
	match := boldSupplier.FindStringSubmatch(string(l))
	if match == nil {
		return "", false
	}
	return match[1], true
}

func readDocSection(t *testing.T, heading string) []markdownLine {
	t.Helper()
	file, err := os.Open(crossContextEventsDoc)
	if err != nil {
		t.Fatalf("open %s: %v", crossContextEventsDoc, err)
	}
	defer func() { _ = file.Close() }()

	var lines []markdownLine
	inSection := false
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "## ") {
			inSection = line == heading
			continue
		}
		if inSection {
			lines = append(lines, markdownLine(line))
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("read %s: %v", crossContextEventsDoc, err)
	}
	if len(lines) == 0 {
		t.Fatalf("%s has no section %q", crossContextEventsDoc, heading)
	}
	return lines
}

var catalogueHeading = regexp.MustCompile("^### ([^(`]+)")

type catalogueReader struct {
	t         *testing.T
	context   string
	block     []string
	inBlock   bool
	constants eventConstants
}

func (r *catalogueReader) read(line markdownLine) {
	text := string(line)
	switch {
	case catalogueHeading.MatchString(text):
		r.context = contextPackage(catalogueHeading.FindStringSubmatch(text)[1])
		r.constants.add(r.context, nil)
	case text == "```go":
		r.inBlock, r.block = true, nil
	case r.inBlock && text == "```":
		r.inBlock = false
		r.constants.add(r.context, r.parseBlock())
	case r.inBlock:
		r.block = append(r.block, text)
	}
}

func (r *catalogueReader) parseBlock() []string {
	source := "package doc\n" + strings.Join(r.block, "\n")
	file, err := parser.ParseFile(token.NewFileSet(), crossContextEventsDoc, source, 0)
	if err != nil {
		r.t.Fatalf("catalogue block of %s is not valid Go: %v", r.context, err)
	}
	return stringConstants(file)
}

func documentedEventConstants(t *testing.T) eventConstants {
	t.Helper()
	reader := &catalogueReader{t: t, constants: eventConstants{}}
	for _, line := range readDocSection(t, "## Complete Event Constants Catalogue") {
		reader.read(line)
	}
	return reader.constants
}

func reportMissing(t *testing.T, missing map[string][]string, format string) {
	t.Helper()
	for _, context := range slices.Sorted(maps.Keys(missing)) {
		for _, name := range missing[context] {
			t.Errorf(format, name, context)
		}
	}
}

func TestEventCatalogueListsEveryPublishedEvent(t *testing.T) {
	published := publishedEventConstants(t)
	documented := documentedEventConstants(t)

	for _, context := range published.sortedContexts() {
		if _, ok := documented[context]; !ok {
			t.Errorf("DOC DRIFT: %s has publishedlanguage event constants but no section in the catalogue of %s", context, crossContextEventsDoc)
		}
	}
	reportMissing(t, published.missingFrom(documented), "DOC DRIFT: published event %s of %s is missing from the catalogue of "+crossContextEventsDoc)
}

func TestEventCatalogueListsOnlyPublishedEvents(t *testing.T) {
	published := publishedEventConstants(t)
	documented := documentedEventConstants(t)

	reportMissing(t, documented.missingFrom(published), "DOC DRIFT: the catalogue of "+crossContextEventsDoc+" lists %s under %s, but no such constant exists in its publishedlanguage *events.go files")
}

type registryReference struct {
	supplier string
	name     eventShorthand
}

var (
	boldSupplier   = regexp.MustCompile(`^\*\*([^*]+)\*\*`)
	backticked     = regexp.MustCompile("`([^`]+)`")
	camelCaseStart = regexp.MustCompile("[A-Z]")
)

type registryReader struct {
	supplier     string
	header       []string
	references   []registryReference
	tableRowSeen bool
}

func columnIndex(header []string, names ...string) int {
	return slices.IndexFunc(header, func(cell string) bool { return slices.Contains(names, cell) })
}

func (r *registryReader) read(line markdownLine) {
	if !line.isTableRow() {
		r.header = nil
		if label, ok := line.boldLabel(); ok {
			r.supplier = label
		}
		return
	}
	cells := line.cells()
	switch {
	case r.header == nil:
		r.header = cells
	case strings.HasPrefix(cells[0], "---"):
	default:
		r.readRow(cells)
	}
}

func (r *registryReader) readRow(cells []string) {
	eventColumn := columnIndex(r.header, "Event", "Events")
	if eventColumn < 0 {
		return
	}
	r.tableRowSeen = true
	supplier := r.supplier
	if supplierColumn := columnIndex(r.header, "Supplier"); supplierColumn >= 0 {
		supplier, _, _ = strings.Cut(cells[supplierColumn], " (")
	}
	cell := strings.ReplaceAll(cells[eventColumn], "` / `", "/")
	for _, match := range backticked.FindAllStringSubmatch(cell, -1) {
		r.references = append(r.references, registryReference{supplier: contextPackage(supplier), name: eventShorthand(match[1])})
	}
}

func documentedSubscriptions(t *testing.T) []registryReference {
	t.Helper()
	reader := &registryReader{}
	for _, line := range readDocSection(t, "## Cross-Context Subscription Registry") {
		reader.read(line)
	}
	if !reader.tableRowSeen {
		t.Fatalf("no subscription table with an Event or Events column found in %s", crossContextEventsDoc)
	}
	return reader.references
}

type eventShorthand string

func (e eventShorthand) parts() (string, []string) {
	parts := strings.Split(string(e), "/")
	return parts[0], parts[1:]
}

func (e eventShorthand) prefixes() []string {
	head, _ := e.parts()
	starts := camelCaseStart.FindAllStringIndex(head, -1)
	prefixes := make([]string, 0, len(starts))
	for _, start := range starts[1:] {
		prefixes = append(prefixes, head[:start[0]])
	}
	return prefixes
}

func (e eventShorthand) unresolvedIn(published map[string]bool) []string {
	head, suffixes := e.parts()
	var unresolved []string
	if !published[head] {
		unresolved = append(unresolved, head)
	}
	for _, suffix := range suffixes {
		resolves := slices.ContainsFunc(e.prefixes(), func(prefix string) bool { return published[prefix+suffix] })
		if !resolves {
			unresolved = append(unresolved, head+"/"+suffix)
		}
	}
	return unresolved
}

func TestSubscriptionRegistryNamesOnlyPublishedEvents(t *testing.T) {
	published := publishedEventConstants(t)

	for _, reference := range documentedSubscriptions(t) {
		supplierEvents, known := published[reference.supplier]
		if !known {
			t.Errorf("DOC DRIFT: the subscription registry of %s names supplier %q, which has no publishedlanguage event constants", crossContextEventsDoc, reference.supplier)
			continue
		}
		for _, name := range reference.name.unresolvedIn(supplierEvents) {
			t.Errorf("DOC DRIFT: the subscription registry of %s lists %s from %s, but %s publishes no such event", crossContextEventsDoc, name, reference.supplier, reference.supplier)
		}
	}
}

func TestEventShorthand_ExpandsAgainstTheSuppliersEvents(t *testing.T) {
	published := map[string]bool{
		"CapabilityCreated": true, "CapabilityUpdated": true, "CapabilityAssignedToDomain": true,
		"CapabilityUnassignedFromDomain": true, "CapabilityRealizationsInherited": true, "CapabilityRealizationsUninherited": true,
	}
	cases := []struct {
		shorthand eventShorthand
		want      []string
	}{
		{"CapabilityCreated", nil},
		{"CapabilityCreated/Updated", nil},
		{"CapabilityAssignedToDomain/UnassignedFromDomain", nil},
		{"CapabilityRealizationsInherited/Uninherited", nil},
		{"CapabilityCreated/Archived", []string{"CapabilityCreated/Archived"}},
		{"CapabilityRetired/Updated", []string{"CapabilityRetired"}},
	}
	for _, tc := range cases {
		t.Run(string(tc.shorthand), func(t *testing.T) {
			if got := tc.shorthand.unresolvedIn(published); !slices.Equal(got, tc.want) {
				t.Errorf("%q.unresolvedIn = %v, want %v", tc.shorthand, got, tc.want)
			}
		})
	}
}

func TestRegistryReader_ReadsOnlyTheEventColumnWithItsSupplier(t *testing.T) {
	reader := &registryReader{}
	for _, line := range []markdownLine{
		"**Capability Mapping** (`cmPL`):",
		"",
		"| Event | Projector | Wired In |",
		"|-------|-----------|----------|",
		"| `CapabilityCreated` / `Updated` | `CacheProjector` | `subscribe()` |",
		"",
		"| Supplier | Events | Projector |",
		"|----------|--------|-----------|",
		"| Architecture Modeling (`archPL`) | `VendorCreated/Deleted`, `VendorUpdated` | `IndexProjector` |",
		"",
		"| Supplier | Command | Consumer |",
		"|----------|---------|----------|",
		"| Auth | `EnsureInvitation` | Access Delegation |",
	} {
		reader.read(line)
	}

	want := []registryReference{
		{supplier: "capabilitymapping", name: "CapabilityCreated/Updated"},
		{supplier: "architecturemodeling", name: "VendorCreated/Deleted"},
		{supplier: "architecturemodeling", name: "VendorUpdated"},
	}
	if !slices.Equal(reader.references, want) {
		t.Errorf("references = %v, want %v", reader.references, want)
	}
}
