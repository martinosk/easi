package events

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"easi/backend/internal/capabilitymapping/domain/valueobjects"
)

func TestNewCapabilityCreated_CarriesDefaultMaturityValue(t *testing.T) {
	event := NewCapabilityCreated("cap-1", "Payments", "desc", "", "L1")

	assert.Equal(t, valueobjects.DefaultMaturityValue, event.MaturityValue)
}

func TestCapabilityCreated_EventData_IncludesMaturityValue(t *testing.T) {
	event := NewCapabilityCreated("cap-1", "Payments", "desc", "", "L1")

	data := event.EventData()

	assert.Equal(t, valueobjects.DefaultMaturityValue, data["maturityValue"])
}
