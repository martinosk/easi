package valueobjects

import (
	"testing"
)

func TestNewImportSessionIDFromString_Empty(t *testing.T) {
	_, err := NewImportSessionIDFromString("")
	if err == nil {
		t.Error("expected error for empty string, got nil")
	}
}

func TestImportSessionID_Equals(t *testing.T) {
	id1, _ := NewImportSessionIDFromString("550e8400-e29b-41d4-a716-446655440000")
	id2, _ := NewImportSessionIDFromString("550e8400-e29b-41d4-a716-446655440000")
	id3, _ := NewImportSessionIDFromString("660e8400-e29b-41d4-a716-446655440000")

	if !id1.Equals(id2) {
		t.Error("expected equal IDs to return true")
	}
	if id1.Equals(id3) {
		t.Error("expected different IDs to return false")
	}
}
