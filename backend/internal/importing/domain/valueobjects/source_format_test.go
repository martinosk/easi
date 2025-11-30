package valueobjects

import (
	"testing"
)

func TestNewSourceFormat_ValidArchiMateOpenExchange(t *testing.T) {
	sf, err := NewSourceFormat("archimate-openexchange")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if sf.Value() != "archimate-openexchange" {
		t.Errorf("expected 'archimate-openexchange', got %q", sf.Value())
	}
	if !sf.IsArchiMateOpenExchange() {
		t.Error("expected IsArchiMateOpenExchange() to return true")
	}
}

func TestNewSourceFormat_InvalidFormat(t *testing.T) {
	testCases := []string{
		"",
		"invalid",
		"archimate",
		"openexchange",
		"ARCHIMATE-OPENEXCHANGE",
	}

	for _, tc := range testCases {
		_, err := NewSourceFormat(tc)
		if err == nil {
			t.Errorf("expected error for source format %q, got nil", tc)
		}
		if err != ErrInvalidSourceFormat {
			t.Errorf("expected ErrInvalidSourceFormat, got %v", err)
		}
	}
}

func TestSourceFormat_Equals(t *testing.T) {
	sf1, _ := NewSourceFormat("archimate-openexchange")
	sf2, _ := NewSourceFormat("archimate-openexchange")

	if !sf1.Equals(sf2) {
		t.Error("expected equal source formats to return true")
	}
}

func TestSourceFormat_String(t *testing.T) {
	sf, _ := NewSourceFormat("archimate-openexchange")
	if sf.String() != "archimate-openexchange" {
		t.Errorf("expected 'archimate-openexchange', got %q", sf.String())
	}
}
