package valueobjects

import (
	"testing"
)

func TestNewImportError(t *testing.T) {
	ie := NewImportError("capability-xyz", "Some Capability", "Parent capability not found", "skipped")

	if ie.SourceElement() != "capability-xyz" {
		t.Errorf("expected 'capability-xyz', got %q", ie.SourceElement())
	}
	if ie.SourceName() != "Some Capability" {
		t.Errorf("expected 'Some Capability', got %q", ie.SourceName())
	}
	if ie.Error() != "Parent capability not found" {
		t.Errorf("expected 'Parent capability not found', got %q", ie.Error())
	}
	if ie.Action() != "skipped" {
		t.Errorf("expected 'skipped', got %q", ie.Action())
	}
}

func TestImportError_Equals(t *testing.T) {
	ie1 := NewImportError("cap-1", "Cap 1", "Error", "skipped")
	ie2 := NewImportError("cap-1", "Cap 1", "Error", "skipped")
	ie3 := NewImportError("cap-2", "Cap 2", "Error", "skipped")

	if !ie1.Equals(ie2) {
		t.Error("expected equal import errors to return true")
	}
	if ie1.Equals(ie3) {
		t.Error("expected different import errors to return false")
	}
}
