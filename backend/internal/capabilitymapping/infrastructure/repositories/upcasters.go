package repositories

type CapabilityMetadataUpdatedV1ToV2Upcaster struct{}

func (u CapabilityMetadataUpdatedV1ToV2Upcaster) EventType() string {
	return "CapabilityMetadataUpdated"
}

func (u CapabilityMetadataUpdatedV1ToV2Upcaster) Upcast(data map[string]interface{}) map[string]interface{} {
	if _, hasValue := data["maturityValue"]; hasValue {
		return data
	}

	if level, ok := data["maturityLevel"].(string); ok {
		data["maturityValue"] = float64(legacyStringToMaturityValue(level))
	}

	delete(data, "maturityLevel")

	return data
}

func legacyStringToMaturityValue(name string) int {
	switch name {
	case "Genesis":
		return 12
	case "Custom Build":
		return 37
	case "Product":
		return 62
	case "Commodity":
		return 87
	default:
		return 12
	}
}
