package parsers

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

var (
	sampleResult     *ParseResult
	sampleResultOnce sync.Once
	sampleResultErr  error
	sampleSkipMsg    string
)

func parseSampleModel() (*ParseResult, error, string) {
	sampleResultOnce.Do(func() {
		sampleFilePath := filepath.Join("..", "..", "..", "..", "..", "docs", "sample-model.xml")
		file, err := os.Open(sampleFilePath)
		if err != nil {
			sampleSkipMsg = "Sample file not found at " + sampleFilePath + ": " + err.Error()
			return
		}
		defer file.Close()

		parser := NewArchiMateParser()
		result, err := parser.Parse(file)
		if err != nil {
			sampleResultErr = err
			return
		}
		sampleResult = result
	})
	return sampleResult, sampleResultErr, sampleSkipMsg
}

func loadSampleResult(t *testing.T) *ParseResult {
	t.Helper()
	result, err, skipMsg := parseSampleModel()
	if skipMsg != "" {
		t.Skip(skipMsg)
	}
	if err != nil {
		t.Fatalf("Failed to parse sample model: %v", err)
	}
	return result
}

func TestArchiMateParser_ParseSampleModel_Capabilities(t *testing.T) {
	result := loadSampleResult(t)

	expectedCapabilities := map[string]bool{
		"Invoicing":                    true,
		"Sales":                        true,
		"Aftermarket":                  true,
		"Support":                      true,
		"Customer Service":             true,
		"Strategic customer relations": true,
		"This is level 4":              true,
		"This is level 5":              true,
	}

	if len(result.Capabilities) != len(expectedCapabilities) {
		t.Errorf("expected %d capabilities, got %d", len(expectedCapabilities), len(result.Capabilities))
	}

	for _, cap := range result.Capabilities {
		if !expectedCapabilities[cap.Name] {
			t.Errorf("unexpected capability: %s", cap.Name)
		}
	}
}

func TestArchiMateParser_ParseSampleModel_ApplicationComponents(t *testing.T) {
	result := loadSampleResult(t)

	expectedComponents := map[string]bool{
		"App 1": true,
		"App 2": true,
	}

	if len(result.Components) != len(expectedComponents) {
		t.Errorf("expected %d components, got %d", len(expectedComponents), len(result.Components))
	}

	for _, comp := range result.Components {
		if !expectedComponents[comp.Name] {
			t.Errorf("unexpected component: %s", comp.Name)
		}
	}
}

func TestArchiMateParser_ParseSampleModel_UnsupportedElements(t *testing.T) {
	result := loadSampleResult(t)

	expectedUnsupported := map[string]int{
		"Resource":             1,
		"BusinessActor":        1,
		"ApplicationInterface": 1,
		"SystemSoftware":       1,
	}

	for elemType, expectedCount := range expectedUnsupported {
		if result.UnsupportedElements[elemType] != expectedCount {
			t.Errorf("expected %d unsupported %s, got %d", expectedCount, elemType, result.UnsupportedElements[elemType])
		}
	}
}

func countRelationshipsByType(relationships []ParsedRelationship, relType string) int {
	count := 0
	for _, rel := range relationships {
		if rel.Type == relType {
			count++
		}
	}
	return count
}

func TestArchiMateParser_ParseSampleModel_HierarchyRelationships(t *testing.T) {
	result := loadSampleResult(t)

	compositionCount := countRelationshipsByType(result.Relationships, "Composition")
	aggregationCount := countRelationshipsByType(result.Relationships, "Aggregation")

	if compositionCount != 5 {
		t.Errorf("expected 5 composition relationships, got %d", compositionCount)
	}
	if aggregationCount != 1 {
		t.Errorf("expected 1 aggregation relationship, got %d", aggregationCount)
	}
}

func TestArchiMateParser_ParseSampleModel_RealizationRelationships(t *testing.T) {
	result := loadSampleResult(t)

	realizationCount := countRelationshipsByType(result.Relationships, "Realization")

	if realizationCount != 1 {
		t.Errorf("expected 1 realization relationship (only Component→Capability are valid), got %d", realizationCount)
	}
}

func TestArchiMateParser_ParseSampleModel_ComponentRelationships(t *testing.T) {
	result := loadSampleResult(t)

	triggeringCount := countRelationshipsByType(result.Relationships, "Triggering")
	servingCount := countRelationshipsByType(result.Relationships, "Serving")

	if triggeringCount != 0 {
		t.Errorf("expected 0 Triggering relationships (sample has none between Components), got %d", triggeringCount)
	}
	if servingCount != 1 {
		t.Errorf("expected 1 Serving relationship, got %d", servingCount)
	}
}

func sumMapValues(m map[string]int) int {
	total := 0
	for _, count := range m {
		total += count
	}
	return total
}

func TestArchiMateParser_ParseSampleModel_SupportedPreview(t *testing.T) {
	result := loadSampleResult(t)
	supported := result.GetPreview().Supported()

	if supported.Capabilities != 8 {
		t.Errorf("expected 8 capabilities in preview, got %d", supported.Capabilities)
	}
	if supported.Components != 2 {
		t.Errorf("expected 2 components in preview, got %d", supported.Components)
	}
	if supported.ParentChildRelationships != 6 {
		t.Errorf("expected 6 parent-child relationships in preview, got %d", supported.ParentChildRelationships)
	}
	if supported.Realizations != 1 {
		t.Errorf("expected 1 realization in preview (only Component→Capability), got %d", supported.Realizations)
	}
	if supported.ComponentRelationships != 1 {
		t.Errorf("expected 1 component relationship in preview (only Component→Component), got %d", supported.ComponentRelationships)
	}
}

func TestArchiMateParser_ParseSampleModel_UnsupportedPreview(t *testing.T) {
	result := loadSampleResult(t)
	unsupported := result.GetPreview().Unsupported()

	totalElements := sumMapValues(unsupported.Elements)
	if totalElements != 4 {
		t.Errorf("expected 4 unsupported elements, got %d", totalElements)
	}

	totalRels := sumMapValues(unsupported.Relationships)
	if totalRels != 3 {
		t.Errorf("expected 3 unsupported relationships (invalid source/target types), got %d", totalRels)
	}
}

func indexCapabilities(capabilities []ParsedElement) (byID map[string]ParsedElement, byName map[string]ParsedElement) {
	byID = make(map[string]ParsedElement, len(capabilities))
	byName = make(map[string]ParsedElement, len(capabilities))
	for _, cap := range capabilities {
		byID[cap.SourceID] = cap
		byName[cap.Name] = cap
	}
	return byID, byName
}

func buildParentMap(relationships []ParsedRelationship) map[string]string {
	parentMap := make(map[string]string)
	for _, rel := range relationships {
		if rel.Type == "Composition" || rel.Type == "Aggregation" {
			parentMap[rel.TargetRef] = rel.SourceRef
		}
	}
	return parentMap
}

func TestArchiMateParser_ParseSampleModel_RootCapabilities(t *testing.T) {
	result := loadSampleResult(t)

	_, capByName := indexCapabilities(result.Capabilities)
	parentMap := buildParentMap(result.Relationships)
	strategicCustRelID := capByName["Strategic customer relations"].SourceID

	rootCount := 0
	for _, cap := range result.Capabilities {
		if _, hasParent := parentMap[cap.SourceID]; hasParent {
			continue
		}
		rootCount++
		if cap.SourceID != strategicCustRelID && cap.Name != "Invoicing" {
			t.Errorf("unexpected root capability: %s", cap.Name)
		}
	}

	if rootCount != 2 {
		t.Errorf("expected 2 root capabilities (Strategic customer relations, Invoicing), got %d", rootCount)
	}
}

func TestArchiMateParser_ParseSampleModel_ParentChildRelationships(t *testing.T) {
	result := loadSampleResult(t)

	capByID, capByName := indexCapabilities(result.Capabilities)
	parentMap := buildParentMap(result.Relationships)

	expectedParentChild := map[string]string{
		"Support":          "Strategic customer relations",
		"Aftermarket":      "Strategic customer relations",
		"Customer Service": "Support",
		"Sales":            "Support",
		"This is level 4":  "Customer Service",
		"This is level 5":  "This is level 4",
	}

	for childName, expectedParentName := range expectedParentChild {
		child := capByName[childName]
		parentSourceID, hasParent := parentMap[child.SourceID]
		if !hasParent {
			t.Errorf("expected %s to have parent %s, but it has no parent", childName, expectedParentName)
			continue
		}
		parent := capByID[parentSourceID]
		if parent.Name != expectedParentName {
			t.Errorf("expected %s parent to be %s, got %s", childName, expectedParentName, parent.Name)
		}
	}
}
