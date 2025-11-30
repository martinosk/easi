package valueobjects

import (
	"testing"
)

func TestNewImportProgress_ValidPhases(t *testing.T) {
	testCases := []string{
		"creating_components",
		"creating_capabilities",
		"creating_realizations",
		"assigning_domains",
	}

	for _, phase := range testCases {
		progress, err := NewImportProgress(phase, 100, 50)
		if err != nil {
			t.Fatalf("expected no error for phase %q, got %v", phase, err)
		}
		if progress.Phase() != phase {
			t.Errorf("expected phase %q, got %q", phase, progress.Phase())
		}
		if progress.TotalItems() != 100 {
			t.Errorf("expected 100 total items, got %d", progress.TotalItems())
		}
		if progress.CompletedItems() != 50 {
			t.Errorf("expected 50 completed items, got %d", progress.CompletedItems())
		}
	}
}

func TestNewImportProgress_InvalidPhase(t *testing.T) {
	_, err := NewImportProgress("invalid_phase", 100, 50)
	if err == nil {
		t.Error("expected error for invalid phase")
	}
	if err != ErrInvalidImportPhase {
		t.Errorf("expected ErrInvalidImportPhase, got %v", err)
	}
}

func TestNewImportProgress_InvalidCounts(t *testing.T) {
	_, err := NewImportProgress("creating_components", -1, 0)
	if err == nil {
		t.Error("expected error for negative total items")
	}

	_, err = NewImportProgress("creating_components", 10, -1)
	if err == nil {
		t.Error("expected error for negative completed items")
	}

	_, err = NewImportProgress("creating_components", 10, 15)
	if err == nil {
		t.Error("expected error for completed > total")
	}
}

func TestImportProgress_PercentComplete(t *testing.T) {
	progress, _ := NewImportProgress("creating_components", 100, 25)
	if progress.PercentComplete() != 25 {
		t.Errorf("expected 25%%, got %d%%", progress.PercentComplete())
	}

	progressZero, _ := NewImportProgress("creating_components", 0, 0)
	if progressZero.PercentComplete() != 0 {
		t.Errorf("expected 0%% for zero items, got %d%%", progressZero.PercentComplete())
	}
}

func TestImportProgress_WithIncrement(t *testing.T) {
	progress, _ := NewImportProgress("creating_components", 100, 50)
	updated := progress.WithIncrement(10)

	if updated.CompletedItems() != 60 {
		t.Errorf("expected 60 completed items, got %d", updated.CompletedItems())
	}
	if progress.CompletedItems() != 50 {
		t.Error("original progress should be unchanged (immutable)")
	}
}

func TestImportProgress_WithNextPhase(t *testing.T) {
	progress, _ := NewImportProgress("creating_components", 100, 100)
	nextPhase, _ := NewImportProgress("creating_capabilities", 50, 0)
	updated := progress.WithNextPhase(nextPhase.Phase(), nextPhase.TotalItems())

	if updated.Phase() != "creating_capabilities" {
		t.Errorf("expected phase 'creating_capabilities', got %q", updated.Phase())
	}
	if updated.TotalItems() != 50 {
		t.Errorf("expected 50 total items, got %d", updated.TotalItems())
	}
	if updated.CompletedItems() != 0 {
		t.Errorf("expected 0 completed items, got %d", updated.CompletedItems())
	}
}
