package parsers

import (
	"os"
	"path/filepath"
	"testing"
)

func TestArchiMateParser_ParseSampleModel(t *testing.T) {
	sampleFilePath := filepath.Join("..", "..", "..", "..", "..", "docs", "sample-model.xml")

	file, err := os.Open(sampleFilePath)
	if err != nil {
		t.Skipf("Sample file not found at %s: %v", sampleFilePath, err)
	}
	defer file.Close()

	parser := NewArchiMateParser()
	result, err := parser.Parse(file)
	if err != nil {
		t.Fatalf("Failed to parse sample model: %v", err)
	}

	t.Run("parses capabilities", func(t *testing.T) {
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
	})

	t.Run("parses application components", func(t *testing.T) {
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
	})

	t.Run("identifies unsupported elements", func(t *testing.T) {
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
	})

	t.Run("parses hierarchy relationships", func(t *testing.T) {
		compositionCount := 0
		aggregationCount := 0
		for _, rel := range result.Relationships {
			if rel.Type == "Composition" {
				compositionCount++
			}
			if rel.Type == "Aggregation" {
				aggregationCount++
			}
		}

		if compositionCount != 5 {
			t.Errorf("expected 5 composition relationships, got %d", compositionCount)
		}
		if aggregationCount != 1 {
			t.Errorf("expected 1 aggregation relationship, got %d", aggregationCount)
		}
	})

	t.Run("parses realization relationships", func(t *testing.T) {
		realizationCount := 0
		for _, rel := range result.Relationships {
			if rel.Type == "Realization" {
				realizationCount++
			}
		}

		if realizationCount != 1 {
			t.Errorf("expected 1 realization relationship (only Component→Capability are valid), got %d", realizationCount)
		}
	})

	t.Run("parses component relationships", func(t *testing.T) {
		triggeringCount := 0
		servingCount := 0
		for _, rel := range result.Relationships {
			if rel.Type == "Triggering" {
				triggeringCount++
			}
			if rel.Type == "Serving" {
				servingCount++
			}
		}

		if triggeringCount != 0 {
			t.Errorf("expected 0 Triggering relationships (sample has none between Components), got %d", triggeringCount)
		}
		if servingCount != 1 {
			t.Errorf("expected 1 Serving relationship, got %d", servingCount)
		}
	})

	t.Run("generates correct preview", func(t *testing.T) {
		preview := result.GetPreview()

		if preview.Supported().Capabilities != 8 {
			t.Errorf("expected 8 capabilities in preview, got %d", preview.Supported().Capabilities)
		}
		if preview.Supported().Components != 2 {
			t.Errorf("expected 2 components in preview, got %d", preview.Supported().Components)
		}
		if preview.Supported().ParentChildRelationships != 6 {
			t.Errorf("expected 6 parent-child relationships in preview, got %d", preview.Supported().ParentChildRelationships)
		}
		if preview.Supported().Realizations != 1 {
			t.Errorf("expected 1 realization in preview (only Component→Capability), got %d", preview.Supported().Realizations)
		}
		if preview.Supported().ComponentRelationships != 1 {
			t.Errorf("expected 1 component relationship in preview (only Component→Component), got %d", preview.Supported().ComponentRelationships)
		}

		totalUnsupportedElements := 0
		for _, count := range preview.Unsupported().Elements {
			totalUnsupportedElements += count
		}
		if totalUnsupportedElements != 4 {
			t.Errorf("expected 4 unsupported elements, got %d", totalUnsupportedElements)
		}

		totalUnsupportedRels := 0
		for _, count := range preview.Unsupported().Relationships {
			totalUnsupportedRels += count
		}
		if totalUnsupportedRels != 3 {
			t.Errorf("expected 3 unsupported relationships (invalid source/target types), got %d", totalUnsupportedRels)
		}
	})

	t.Run("builds correct capability hierarchy", func(t *testing.T) {
		capByID := make(map[string]ParsedElement)
		capByName := make(map[string]ParsedElement)
		for _, cap := range result.Capabilities {
			capByID[cap.SourceID] = cap
			capByName[cap.Name] = cap
		}

		parentMap := make(map[string]string)
		for _, rel := range result.Relationships {
			if rel.Type == "Composition" || rel.Type == "Aggregation" {
				parentMap[rel.TargetRef] = rel.SourceRef
			}
		}

		strategicCustRelID := ""
		for _, cap := range result.Capabilities {
			if cap.Name == "Strategic customer relations" {
				strategicCustRelID = cap.SourceID
				break
			}
		}

		rootCount := 0
		for _, cap := range result.Capabilities {
			if _, hasParent := parentMap[cap.SourceID]; !hasParent {
				rootCount++
				if cap.SourceID != strategicCustRelID && cap.Name != "Invoicing" {
					t.Errorf("unexpected root capability: %s", cap.Name)
				}
			}
		}

		if rootCount != 2 {
			t.Errorf("expected 2 root capabilities (Strategic customer relations, Invoicing), got %d", rootCount)
		}

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
	})
}
