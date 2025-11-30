package parsers

import (
	"strings"
	"testing"
)

var testXML = `<?xml version="1.0" encoding="UTF-8"?>
<model xmlns="http://www.opengroup.org/xsd/archimate/3.0/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" identifier="test-model">
  <name>Test Model</name>
  <elements>
    <element identifier="cap-1" xsi:type="Capability">
      <name>Customer Management</name>
      <documentation>Manages customer relationships</documentation>
    </element>
    <element identifier="cap-2" xsi:type="Capability">
      <name>Customer Onboarding</name>
    </element>
    <element identifier="comp-1" xsi:type="ApplicationComponent">
      <name>CRM System</name>
      <documentation>Customer relationship management system</documentation>
    </element>
    <element identifier="svc-1" xsi:type="ApplicationService">
      <name>Customer Service</name>
    </element>
    <element identifier="bp-1" xsi:type="BusinessProcess">
      <name>Sales Process</name>
    </element>
  </elements>
  <relationships>
    <relationship identifier="rel-1" xsi:type="Aggregation" source="cap-1" target="cap-2">
      <name>Contains</name>
    </relationship>
    <relationship identifier="rel-2" xsi:type="Composition" source="cap-1" target="cap-2"/>
    <relationship identifier="rel-3" xsi:type="Realization" source="comp-1" target="cap-1">
      <name>Supports</name>
      <documentation>CRM supports customer management</documentation>
    </relationship>
    <relationship identifier="rel-4" xsi:type="Realization" source="svc-1" target="cap-2"/>
    <relationship identifier="rel-5" xsi:type="Flow" source="comp-1" target="svc-1"/>
  </relationships>
</model>`

func TestArchiMateParser_Parse(t *testing.T) {
	parser := NewArchiMateParser()
	result, err := parser.Parse(strings.NewReader(testXML))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(result.Capabilities) != 2 {
		t.Errorf("expected 2 capabilities, got %d", len(result.Capabilities))
	}
	if len(result.Components) != 2 {
		t.Errorf("expected 2 components (ApplicationComponent + ApplicationService), got %d", len(result.Components))
	}

	cap1 := findElement(result.Capabilities, "cap-1")
	if cap1 == nil {
		t.Fatal("expected to find cap-1")
	}
	if cap1.Name != "Customer Management" {
		t.Errorf("expected name 'Customer Management', got %q", cap1.Name)
	}
	if cap1.Description != "Manages customer relationships" {
		t.Errorf("expected description 'Manages customer relationships', got %q", cap1.Description)
	}

	comp1 := findElement(result.Components, "comp-1")
	if comp1 == nil {
		t.Fatal("expected to find comp-1")
	}
	if comp1.Name != "CRM System" {
		t.Errorf("expected name 'CRM System', got %q", comp1.Name)
	}
}

func TestArchiMateParser_ParseRelationships(t *testing.T) {
	parser := NewArchiMateParser()
	result, err := parser.Parse(strings.NewReader(testXML))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	aggregations := filterRelationships(result.Relationships, "Aggregation")
	if len(aggregations) != 1 {
		t.Errorf("expected 1 aggregation, got %d", len(aggregations))
	}

	compositions := filterRelationships(result.Relationships, "Composition")
	if len(compositions) != 1 {
		t.Errorf("expected 1 composition, got %d", len(compositions))
	}

	realizations := filterRelationships(result.Relationships, "Realization")
	if len(realizations) != 2 {
		t.Errorf("expected 2 realizations, got %d", len(realizations))
	}
}

func TestArchiMateParser_Preview(t *testing.T) {
	parser := NewArchiMateParser()
	result, err := parser.Parse(strings.NewReader(testXML))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	preview := result.GetPreview()

	if preview.Supported().Capabilities != 2 {
		t.Errorf("expected 2 capabilities in preview, got %d", preview.Supported().Capabilities)
	}
	if preview.Supported().Components != 2 {
		t.Errorf("expected 2 components in preview, got %d", preview.Supported().Components)
	}
	if preview.Supported().ParentChildRelationships != 2 {
		t.Errorf("expected 2 parent-child relationships (aggregation + composition), got %d", preview.Supported().ParentChildRelationships)
	}
	if preview.Supported().Realizations != 2 {
		t.Errorf("expected 2 realizations, got %d", preview.Supported().Realizations)
	}

	if preview.Unsupported().Elements["BusinessProcess"] != 1 {
		t.Errorf("expected 1 unsupported BusinessProcess, got %d", preview.Unsupported().Elements["BusinessProcess"])
	}
	if preview.Unsupported().Relationships["Flow"] != 1 {
		t.Errorf("expected 1 unsupported Flow, got %d", preview.Unsupported().Relationships["Flow"])
	}
}

func TestArchiMateParser_InvalidXML(t *testing.T) {
	parser := NewArchiMateParser()
	_, err := parser.Parse(strings.NewReader("not valid xml"))
	if err == nil {
		t.Error("expected error for invalid XML")
	}
}

func TestArchiMateParser_EmptyModel(t *testing.T) {
	emptyXML := `<?xml version="1.0" encoding="UTF-8"?>
<model xmlns="http://www.opengroup.org/xsd/archimate/3.0/">
  <name>Empty Model</name>
  <elements/>
  <relationships/>
</model>`

	parser := NewArchiMateParser()
	result, err := parser.Parse(strings.NewReader(emptyXML))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(result.Capabilities) != 0 {
		t.Errorf("expected 0 capabilities, got %d", len(result.Capabilities))
	}
	if len(result.Components) != 0 {
		t.Errorf("expected 0 components, got %d", len(result.Components))
	}
}

func findElement(elements []ParsedElement, sourceID string) *ParsedElement {
	for i := range elements {
		if elements[i].SourceID == sourceID {
			return &elements[i]
		}
	}
	return nil
}

func filterRelationships(rels []ParsedRelationship, relType string) []ParsedRelationship {
	var result []ParsedRelationship
	for _, r := range rels {
		if r.Type == relType {
			result = append(result, r)
		}
	}
	return result
}
