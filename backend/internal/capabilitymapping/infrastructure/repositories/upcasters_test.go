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

func TestCapabilityMetadataUpdatedV1ToV2Upcaster_Upcast_V2Event_PreservesMaturityValue(t *testing.T) {
	upcaster := CapabilityMetadataUpdatedV1ToV2Upcaster{}

	tests := []struct {
		name          string
		maturityValue float64
	}{
		{"non-zero value", 42},
		{"zero value", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := map[string]interface{}{
				"id":             "test-id",
				"strategyPillar": "Optimize",
				"pillarWeight":   float64(3),
				"maturityValue":  tt.maturityValue,
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
			if maturityValue != tt.maturityValue {
				t.Errorf("Expected maturityValue %v, got %v", tt.maturityValue, maturityValue)
			}
		})
	}
}

func TestCapabilityCreatedMaturityUpcaster_EventType(t *testing.T) {
	upcaster := CapabilityCreatedMaturityUpcaster{}
	if upcaster.EventType() != "CapabilityCreated" {
		t.Errorf("Expected EventType to be 'CapabilityCreated', got '%s'", upcaster.EventType())
	}
}

func TestCapabilityCreatedMaturityUpcaster_Upcast(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]interface{}
		expected float64
	}{
		{
			name: "missing field defaults to Genesis",
			data: map[string]interface{}{
				"id": "cap-1", "name": "Payments", "description": "", "parentId": "", "level": "L1",
			},
			expected: 12,
		},
		{
			name:     "present field is preserved",
			data:     map[string]interface{}{"id": "cap-1", "name": "Payments", "maturityValue": float64(62)},
			expected: 62,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			upcaster := CapabilityCreatedMaturityUpcaster{}

			result := upcaster.Upcast(tt.data)

			maturityValue, ok := result["maturityValue"].(float64)
			if !ok {
				t.Fatal("Expected maturityValue to be present after upcast")
			}
			if maturityValue != tt.expected {
				t.Errorf("Expected maturityValue %v, got %v", tt.expected, maturityValue)
			}
		})
	}
}
