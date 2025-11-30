package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLayoutPreferences_Empty(t *testing.T) {
	prefs := NewLayoutPreferences(nil)
	assert.Equal(t, map[string]interface{}{}, prefs.ToMap())
}

func TestNewLayoutPreferences_WithData(t *testing.T) {
	data := map[string]interface{}{
		"colorScheme":     "pastel",
		"layoutDirection": "TB",
		"edgeType":        "bezier",
	}
	prefs := NewLayoutPreferences(data)
	assert.Equal(t, "pastel", prefs.Get("colorScheme"))
	assert.Equal(t, "TB", prefs.Get("layoutDirection"))
	assert.Equal(t, "bezier", prefs.Get("edgeType"))
}

func TestLayoutPreferences_Get_Missing(t *testing.T) {
	prefs := NewLayoutPreferences(nil)
	assert.Nil(t, prefs.Get("missing"))
}

func TestLayoutPreferences_WithUpdated(t *testing.T) {
	prefs := NewLayoutPreferences(map[string]interface{}{
		"colorScheme": "default",
	})

	newPrefs := prefs.WithUpdated(map[string]interface{}{
		"colorScheme":     "pastel",
		"layoutDirection": "LR",
	})

	assert.Equal(t, "pastel", newPrefs.Get("colorScheme"))
	assert.Equal(t, "LR", newPrefs.Get("layoutDirection"))
	assert.Equal(t, "default", prefs.Get("colorScheme"))
}

func TestLayoutPreferences_ToMap(t *testing.T) {
	data := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
	}
	prefs := NewLayoutPreferences(data)
	result := prefs.ToMap()

	assert.Equal(t, "value1", result["key1"])
	assert.Equal(t, 42, result["key2"])

	result["key1"] = "modified"
	assert.Equal(t, "value1", prefs.Get("key1"))
}

func TestLayoutPreferences_Equals(t *testing.T) {
	prefs1 := NewLayoutPreferences(map[string]interface{}{"key": "value"})
	prefs2 := NewLayoutPreferences(map[string]interface{}{"key": "value"})
	prefs3 := NewLayoutPreferences(map[string]interface{}{"key": "other"})

	assert.True(t, prefs1.Equals(prefs2))
	assert.False(t, prefs1.Equals(prefs3))
}
