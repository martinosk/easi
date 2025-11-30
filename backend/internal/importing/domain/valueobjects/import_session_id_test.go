package valueobjects

import (
	"testing"

	"github.com/google/uuid"
)

func TestNewImportSessionID_GeneratesValidUUID(t *testing.T) {
	id := NewImportSessionID()
	if id.Value() == "" {
		t.Error("expected non-empty ID")
	}
	if _, err := uuid.Parse(id.Value()); err != nil {
		t.Errorf("expected valid UUID, got %q", id.Value())
	}
}

func TestNewImportSessionIDFromString_ValidUUID(t *testing.T) {
	validUUID := uuid.New().String()
	id, err := NewImportSessionIDFromString(validUUID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if id.Value() != validUUID {
		t.Errorf("expected %q, got %q", validUUID, id.Value())
	}
}

func TestNewImportSessionIDFromString_InvalidUUID(t *testing.T) {
	testCases := []string{
		"",
		"not-a-uuid",
		"12345",
	}

	for _, tc := range testCases {
		_, err := NewImportSessionIDFromString(tc)
		if err == nil {
			t.Errorf("expected error for %q, got nil", tc)
		}
	}
}

func TestImportSessionID_Equals(t *testing.T) {
	id1, _ := NewImportSessionIDFromString(uuid.New().String())
	id2, _ := NewImportSessionIDFromString(id1.Value())
	id3 := NewImportSessionID()

	if !id1.Equals(id2) {
		t.Error("expected equal IDs to return true")
	}
	if id1.Equals(id3) {
		t.Error("expected different IDs to return false")
	}
}
