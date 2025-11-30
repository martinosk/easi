package valueobjects

import (
	"testing"
)

func TestNewImportPreview(t *testing.T) {
	supported := SupportedCounts{
		Capabilities:             45,
		Components:               12,
		ParentChildRelationships: 38,
		Realizations:             15,
	}
	unsupported := UnsupportedCounts{
		Elements: map[string]int{
			"BusinessProcess": 5,
			"DataObject":      3,
		},
		Relationships: map[string]int{
			"Flow":    12,
			"Serving": 4,
		},
	}

	preview := NewImportPreview(supported, unsupported)

	if preview.Supported().Capabilities != 45 {
		t.Errorf("expected 45 capabilities, got %d", preview.Supported().Capabilities)
	}
	if preview.Supported().Components != 12 {
		t.Errorf("expected 12 components, got %d", preview.Supported().Components)
	}
	if preview.Unsupported().Elements["BusinessProcess"] != 5 {
		t.Errorf("expected 5 BusinessProcess, got %d", preview.Unsupported().Elements["BusinessProcess"])
	}
}

func TestImportPreview_TotalSupportedItems(t *testing.T) {
	supported := SupportedCounts{
		Capabilities:             10,
		Components:               5,
		ParentChildRelationships: 8,
		Realizations:             3,
	}
	unsupported := UnsupportedCounts{}

	preview := NewImportPreview(supported, unsupported)

	expected := 10 + 5 + 8 + 3
	if preview.TotalSupportedItems() != expected {
		t.Errorf("expected %d total items, got %d", expected, preview.TotalSupportedItems())
	}
}

func TestImportPreview_HasUnsupportedElements(t *testing.T) {
	supported := SupportedCounts{}
	unsupportedWithElements := UnsupportedCounts{
		Elements: map[string]int{"BusinessProcess": 1},
	}
	unsupportedEmpty := UnsupportedCounts{
		Elements: map[string]int{},
	}

	previewWith := NewImportPreview(supported, unsupportedWithElements)
	previewWithout := NewImportPreview(supported, unsupportedEmpty)

	if !previewWith.HasUnsupportedElements() {
		t.Error("expected HasUnsupportedElements() to return true")
	}
	if previewWithout.HasUnsupportedElements() {
		t.Error("expected HasUnsupportedElements() to return false")
	}
}

func TestImportPreview_Equals(t *testing.T) {
	supported := SupportedCounts{Capabilities: 10}
	unsupported := UnsupportedCounts{}

	p1 := NewImportPreview(supported, unsupported)
	p2 := NewImportPreview(supported, unsupported)

	if !p1.Equals(p2) {
		t.Error("expected equal previews to return true")
	}
}
