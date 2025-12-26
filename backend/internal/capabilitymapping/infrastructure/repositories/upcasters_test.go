package repositories

import (
	"testing"
)

func TestCapabilityMetadataUpdatedV1ToV2Upcaster_EventType(t *testing.T) {
	upcaster := CapabilityMetadataUpdatedV1ToV2Upcaster{}
	if upcaster.EventType() != "CapabilityMetadataUpdated" {
		t.Errorf("Expected EventType to be 'CapabilityMetadataUpdated', got '%s'", upcaster.EventType())
	}
}

func TestCapabilityMetadataUpdatedV1ToV2Upcaster_Upcast_V1Event(t *testing.T) {
	upcaster := CapabilityMetadataUpdatedV1ToV2Upcaster{}

	tests := []struct {
		name          string
		maturityLevel string
		expectedValue int
	}{
		{"Genesis", "Genesis", 12},
		{"Custom Build", "Custom Build", 37},
		{"Product", "Product", 62},
		{"Commodity", "Commodity", 87},
		{"Unknown defaults to Genesis", "Unknown", 12},
		{"Empty defaults to Genesis", "", 12},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := map[string]interface{}{
				"id":             "test-id",
				"strategyPillar": "Optimize",
				"pillarWeight":   float64(3),
				"maturityLevel":  tt.maturityLevel,
				"ownershipModel": "Centralized",
				"primaryOwner":   "John Doe",
				"eaOwner":        "Jane Doe",
				"status":         "Active",
			}

			result := upcaster.Upcast(data)

			if _, hasMaturityLevel := result["maturityLevel"]; hasMaturityLevel {
				t.Error("Expected maturityLevel to be removed after upcast")
			}

			maturityValue, ok := result["maturityValue"].(float64)
			if !ok {
				t.Fatal("Expected maturityValue to be present after upcast")
			}

			if int(maturityValue) != tt.expectedValue {
				t.Errorf("Expected maturityValue %d, got %d", tt.expectedValue, int(maturityValue))
			}
		})
	}
}

func TestCapabilityMetadataUpdatedV1ToV2Upcaster_Upcast_V2Event(t *testing.T) {
	upcaster := CapabilityMetadataUpdatedV1ToV2Upcaster{}

	data := map[string]interface{}{
		"id":             "test-id",
		"strategyPillar": "Optimize",
		"pillarWeight":   float64(3),
		"maturityValue":  float64(42),
		"ownershipModel": "Centralized",
		"primaryOwner":   "John Doe",
		"eaOwner":        "Jane Doe",
		"status":         "Active",
	}

	result := upcaster.Upcast(data)

	maturityValue, ok := result["maturityValue"].(float64)
	if !ok {
		t.Fatal("Expected maturityValue to be preserved")
	}

	if int(maturityValue) != 42 {
		t.Errorf("Expected maturityValue 42, got %d", int(maturityValue))
	}
}

func TestCapabilityMetadataUpdatedV1ToV2Upcaster_Upcast_V2EventWithZeroValue(t *testing.T) {
	upcaster := CapabilityMetadataUpdatedV1ToV2Upcaster{}

	data := map[string]interface{}{
		"id":             "test-id",
		"strategyPillar": "Optimize",
		"pillarWeight":   float64(3),
		"maturityValue":  float64(0),
		"ownershipModel": "Centralized",
		"primaryOwner":   "John Doe",
		"eaOwner":        "Jane Doe",
		"status":         "Active",
	}

	result := upcaster.Upcast(data)

	maturityValue, ok := result["maturityValue"].(float64)
	if !ok {
		t.Fatal("Expected maturityValue to be preserved even when zero")
	}

	if int(maturityValue) != 0 {
		t.Errorf("Expected maturityValue 0, got %d", int(maturityValue))
	}
}
