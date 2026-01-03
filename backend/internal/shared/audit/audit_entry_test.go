package audit

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatEventTypeDisplayName_SingleWord(t *testing.T) {
	result := FormatEventTypeDisplayName("capability.Created")
	assert.Equal(t, "Created", result)
}

func TestFormatEventTypeDisplayName_CamelCase(t *testing.T) {
	result := FormatEventTypeDisplayName("capability.NameChanged")
	assert.Equal(t, "Name Changed", result)
}

func TestFormatEventTypeDisplayName_MultipleCamelCaseWords(t *testing.T) {
	result := FormatEventTypeDisplayName("capability.BusinessValueUpdated")
	assert.Equal(t, "Business Value Updated", result)
}

func TestFormatEventTypeDisplayName_AllUpperCase(t *testing.T) {
	result := FormatEventTypeDisplayName("capability.DELETED")
	assert.Equal(t, "Deleted", result)
}

func TestFormatEventTypeDisplayName_MixedCase(t *testing.T) {
	result := FormatEventTypeDisplayName("component.TypeIDChanged")
	assert.Equal(t, "Type Idchanged", result)
}

func TestFormatEventTypeDisplayName_NestedNamespace(t *testing.T) {
	result := FormatEventTypeDisplayName("domain.subdomain.EntityCreated")
	assert.Equal(t, "Entity Created", result)
}

func TestFormatEventTypeDisplayName_NoNamespace(t *testing.T) {
	result := FormatEventTypeDisplayName("Created")
	assert.Equal(t, "Created", result)
}
