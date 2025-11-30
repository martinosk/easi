package valueobjects

import (
	"testing"
)

func TestNewImportStatus_ValidStatuses(t *testing.T) {
	testCases := []struct {
		status   string
		isPending    bool
		isImporting  bool
		isCompleted  bool
		isFailed     bool
	}{
		{"pending", true, false, false, false},
		{"importing", false, true, false, false},
		{"completed", false, false, true, false},
		{"failed", false, false, false, true},
	}

	for _, tc := range testCases {
		is, err := NewImportStatus(tc.status)
		if err != nil {
			t.Fatalf("expected no error for status %q, got %v", tc.status, err)
		}
		if is.Value() != tc.status {
			t.Errorf("expected %q, got %q", tc.status, is.Value())
		}
		if is.IsPending() != tc.isPending {
			t.Errorf("status %q: IsPending() expected %v, got %v", tc.status, tc.isPending, is.IsPending())
		}
		if is.IsImporting() != tc.isImporting {
			t.Errorf("status %q: IsImporting() expected %v, got %v", tc.status, tc.isImporting, is.IsImporting())
		}
		if is.IsCompleted() != tc.isCompleted {
			t.Errorf("status %q: IsCompleted() expected %v, got %v", tc.status, tc.isCompleted, is.IsCompleted())
		}
		if is.IsFailed() != tc.isFailed {
			t.Errorf("status %q: IsFailed() expected %v, got %v", tc.status, tc.isFailed, is.IsFailed())
		}
	}
}

func TestNewImportStatus_InvalidStatus(t *testing.T) {
	testCases := []string{
		"",
		"invalid",
		"PENDING",
		"Completed",
	}

	for _, tc := range testCases {
		_, err := NewImportStatus(tc)
		if err == nil {
			t.Errorf("expected error for status %q, got nil", tc)
		}
		if err != ErrInvalidImportStatus {
			t.Errorf("expected ErrInvalidImportStatus, got %v", err)
		}
	}
}

func TestImportStatusPending_Constructor(t *testing.T) {
	is := ImportStatusPending()
	if !is.IsPending() {
		t.Error("expected IsPending() to return true")
	}
	if is.Value() != "pending" {
		t.Errorf("expected 'pending', got %q", is.Value())
	}
}

func TestImportStatusImporting_Constructor(t *testing.T) {
	is := ImportStatusImporting()
	if !is.IsImporting() {
		t.Error("expected IsImporting() to return true")
	}
}

func TestImportStatusCompleted_Constructor(t *testing.T) {
	is := ImportStatusCompleted()
	if !is.IsCompleted() {
		t.Error("expected IsCompleted() to return true")
	}
}

func TestImportStatusFailed_Constructor(t *testing.T) {
	is := ImportStatusFailed()
	if !is.IsFailed() {
		t.Error("expected IsFailed() to return true")
	}
}

func TestImportStatus_CanTransitionTo(t *testing.T) {
	pending := ImportStatusPending()
	importing := ImportStatusImporting()
	completed := ImportStatusCompleted()
	failed := ImportStatusFailed()

	if !pending.CanTransitionTo(importing) {
		t.Error("pending should transition to importing")
	}
	if pending.CanTransitionTo(completed) {
		t.Error("pending should not transition directly to completed")
	}
	if pending.CanTransitionTo(failed) {
		t.Error("pending should not transition directly to failed")
	}

	if !importing.CanTransitionTo(completed) {
		t.Error("importing should transition to completed")
	}
	if !importing.CanTransitionTo(failed) {
		t.Error("importing should transition to failed")
	}
	if importing.CanTransitionTo(pending) {
		t.Error("importing should not transition back to pending")
	}

	if completed.CanTransitionTo(pending) {
		t.Error("completed should not transition to pending")
	}

	if failed.CanTransitionTo(pending) {
		t.Error("failed should not transition to pending")
	}
}

func TestImportStatus_Equals(t *testing.T) {
	is1, _ := NewImportStatus("pending")
	is2 := ImportStatusPending()

	if !is1.Equals(is2) {
		t.Error("expected equal import statuses to return true")
	}
}
