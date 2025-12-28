package valueobjects

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"reflect"
)

type LayoutPreferences struct {
	data map[string]interface{}
}

func NewLayoutPreferences(data map[string]interface{}) LayoutPreferences {
	if data == nil {
		data = make(map[string]interface{})
	}
	copyData := make(map[string]interface{}, len(data))
	for k, v := range data {
		copyData[k] = v
	}
	return LayoutPreferences{data: copyData}
}

func (l LayoutPreferences) Get(key string) interface{} {
	return l.data[key]
}

func (l LayoutPreferences) WithUpdated(updates map[string]interface{}) LayoutPreferences {
	newData := make(map[string]interface{}, len(l.data)+len(updates))
	for k, v := range l.data {
		newData[k] = v
	}
	for k, v := range updates {
		newData[k] = v
	}
	return LayoutPreferences{data: newData}
}

func (l LayoutPreferences) ToMap() map[string]interface{} {
	copyData := make(map[string]interface{}, len(l.data))
	for k, v := range l.data {
		copyData[k] = v
	}
	return copyData
}

func (l LayoutPreferences) Equals(other domain.ValueObject) bool {
	if otherPrefs, ok := other.(LayoutPreferences); ok {
		return reflect.DeepEqual(l.data, otherPrefs.data)
	}
	return false
}
