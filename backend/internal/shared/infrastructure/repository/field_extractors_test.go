package repository

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

func checkError(t *testing.T, funcName string, err error, wantErr bool, errContains string) bool {
	t.Helper()
	if (err != nil) != wantErr {
		t.Errorf("%s() error = %v, wantErr %v", funcName, err, wantErr)
		return false
	}
	if wantErr && errContains != "" && (err == nil || !strings.Contains(err.Error(), errContains)) {
		t.Errorf("%s() error = %v, want error containing %q", funcName, err, errContains)
		return false
	}
	return !wantErr
}

func TestGetRequiredString(t *testing.T) {
	tests := []struct {
		name        string
		data        map[string]interface{}
		key         string
		want        string
		wantErr     bool
		errContains string
	}{
		{"valid string", map[string]interface{}{"name": "test"}, "name", "test", false, ""},
		{"empty string is valid", map[string]interface{}{"name": ""}, "name", "", false, ""},
		{"missing field", map[string]interface{}{}, "name", "", true, "required field is missing"},
		{"null value", map[string]interface{}{"name": nil}, "name", "", true, "required field is missing"},
		{"wrong type", map[string]interface{}{"name": 123}, "name", "", true, "expected type string"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetRequiredString(tt.data, tt.key)
			if checkError(t, "GetRequiredString", err, tt.wantErr, tt.errContains) && got != tt.want {
				t.Errorf("GetRequiredString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetOptionalString(t *testing.T) {
	tests := []struct {
		name        string
		data        map[string]interface{}
		key         string
		defaultVal  string
		want        string
		wantErr     bool
		errContains string
	}{
		{"valid string", map[string]interface{}{"name": "test"}, "name", "default", "test", false, ""},
		{"missing field returns default", map[string]interface{}{}, "name", "default", "default", false, ""},
		{"null value returns default", map[string]interface{}{"name": nil}, "name", "default", "default", false, ""},
		{"wrong type returns error", map[string]interface{}{"name": 123}, "name", "default", "", true, "expected type string"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetOptionalString(tt.data, tt.key, tt.defaultVal)
			if checkError(t, "GetOptionalString", err, tt.wantErr, tt.errContains) && got != tt.want {
				t.Errorf("GetOptionalString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetRequiredInt(t *testing.T) {
	tests := []struct {
		name        string
		data        map[string]interface{}
		key         string
		want        int
		wantErr     bool
		errContains string
	}{
		{"valid int", map[string]interface{}{"count": 42}, "count", 42, false, ""},
		{"float64 converted to int", map[string]interface{}{"count": float64(42)}, "count", 42, false, ""},
		{"int64 converted to int", map[string]interface{}{"count": int64(42)}, "count", 42, false, ""},
		{"negative int", map[string]interface{}{"count": -100}, "count", -100, false, ""},
		{"missing field", map[string]interface{}{}, "count", 0, true, "required field is missing"},
		{"wrong type", map[string]interface{}{"count": "not a number"}, "count", 0, true, "expected type int"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetRequiredInt(tt.data, tt.key)
			if checkError(t, "GetRequiredInt", err, tt.wantErr, tt.errContains) && got != tt.want {
				t.Errorf("GetRequiredInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetOptionalInt(t *testing.T) {
	tests := []struct {
		name        string
		data        map[string]interface{}
		key         string
		defaultVal  int
		want        int
		wantErr     bool
		errContains string
	}{
		{"valid int", map[string]interface{}{"count": 42}, "count", 0, 42, false, ""},
		{"missing field returns default", map[string]interface{}{}, "count", 10, 10, false, ""},
		{"null value returns default", map[string]interface{}{"count": nil}, "count", 10, 10, false, ""},
		{"wrong type returns error", map[string]interface{}{"count": "not a number"}, "count", 10, 0, true, "expected type int"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetOptionalInt(tt.data, tt.key, tt.defaultVal)
			if checkError(t, "GetOptionalInt", err, tt.wantErr, tt.errContains) && got != tt.want {
				t.Errorf("GetOptionalInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetRequiredFloat64(t *testing.T) {
	tests := []struct {
		name        string
		data        map[string]interface{}
		key         string
		want        float64
		wantErr     bool
		errContains string
	}{
		{"valid float64", map[string]interface{}{"price": 42.5}, "price", 42.5, false, ""},
		{"int converted to float64", map[string]interface{}{"price": 42}, "price", 42.0, false, ""},
		{"missing field", map[string]interface{}{}, "price", 0, true, "required field is missing"},
		{"wrong type", map[string]interface{}{"price": "not a number"}, "price", 0, true, "expected type float64"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetRequiredFloat64(tt.data, tt.key)
			if checkError(t, "GetRequiredFloat64", err, tt.wantErr, tt.errContains) && got != tt.want {
				t.Errorf("GetRequiredFloat64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetOptionalFloat64(t *testing.T) {
	tests := []struct {
		name        string
		data        map[string]interface{}
		key         string
		defaultVal  float64
		want        float64
		wantErr     bool
		errContains string
	}{
		{"valid float64", map[string]interface{}{"price": 42.5}, "price", 0.0, 42.5, false, ""},
		{"missing field returns default", map[string]interface{}{}, "price", 10.5, 10.5, false, ""},
		{"null value returns default", map[string]interface{}{"price": nil}, "price", 10.5, 10.5, false, ""},
		{"wrong type returns error", map[string]interface{}{"price": "not a number"}, "price", 10.5, 0, true, "expected type float64"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetOptionalFloat64(tt.data, tt.key, tt.defaultVal)
			if checkError(t, "GetOptionalFloat64", err, tt.wantErr, tt.errContains) && got != tt.want {
				t.Errorf("GetOptionalFloat64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetRequiredBool(t *testing.T) {
	tests := []struct {
		name        string
		data        map[string]interface{}
		key         string
		want        bool
		wantErr     bool
		errContains string
	}{
		{"valid true", map[string]interface{}{"active": true}, "active", true, false, ""},
		{"valid false", map[string]interface{}{"active": false}, "active", false, false, ""},
		{"missing field", map[string]interface{}{}, "active", false, true, "required field is missing"},
		{"wrong type", map[string]interface{}{"active": "true"}, "active", false, true, "expected type bool"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetRequiredBool(tt.data, tt.key)
			if checkError(t, "GetRequiredBool", err, tt.wantErr, tt.errContains) && got != tt.want {
				t.Errorf("GetRequiredBool() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetOptionalBool(t *testing.T) {
	tests := []struct {
		name        string
		data        map[string]interface{}
		key         string
		defaultVal  bool
		want        bool
		wantErr     bool
		errContains string
	}{
		{"valid bool", map[string]interface{}{"active": true}, "active", false, true, false, ""},
		{"missing field returns default", map[string]interface{}{}, "active", true, true, false, ""},
		{"null value returns default", map[string]interface{}{"active": nil}, "active", true, true, false, ""},
		{"wrong type returns error", map[string]interface{}{"active": "true"}, "active", false, false, true, "expected type bool"},
		{"int instead of bool returns error", map[string]interface{}{"active": 1}, "active", false, false, true, "expected type bool"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetOptionalBool(tt.data, tt.key, tt.defaultVal)
			if checkError(t, "GetOptionalBool", err, tt.wantErr, tt.errContains) && got != tt.want {
				t.Errorf("GetOptionalBool() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetRequiredTime(t *testing.T) {
	validTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	validTimeNano := time.Date(2024, 1, 15, 10, 30, 0, 123456789, time.UTC)
	tests := []struct {
		name        string
		data        map[string]interface{}
		key         string
		want        time.Time
		wantErr     bool
		errContains string
	}{
		{"valid RFC3339 time", map[string]interface{}{"createdAt": "2024-01-15T10:30:00Z"}, "createdAt", validTime, false, ""},
		{"valid RFC3339Nano time", map[string]interface{}{"createdAt": "2024-01-15T10:30:00.123456789Z"}, "createdAt", validTimeNano, false, ""},
		{"missing field", map[string]interface{}{}, "createdAt", time.Time{}, true, "required field is missing"},
		{"wrong type", map[string]interface{}{"createdAt": 12345}, "createdAt", time.Time{}, true, "expected type string"},
		{"invalid time format", map[string]interface{}{"createdAt": "not-a-time"}, "createdAt", time.Time{}, true, "invalid time format"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetRequiredTime(tt.data, tt.key)
			if checkError(t, "GetRequiredTime", err, tt.wantErr, tt.errContains) && !got.Equal(tt.want) {
				t.Errorf("GetRequiredTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetOptionalTime(t *testing.T) {
	defaultTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	validTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	tests := []struct {
		name        string
		data        map[string]interface{}
		key         string
		defaultVal  time.Time
		want        time.Time
		wantErr     bool
		errContains string
	}{
		{"valid time", map[string]interface{}{"createdAt": "2024-01-15T10:30:00Z"}, "createdAt", defaultTime, validTime, false, ""},
		{"missing field returns default", map[string]interface{}{}, "createdAt", defaultTime, defaultTime, false, ""},
		{"null value returns default", map[string]interface{}{"createdAt": nil}, "createdAt", defaultTime, defaultTime, false, ""},
		{"wrong type returns error", map[string]interface{}{"createdAt": 12345}, "createdAt", defaultTime, time.Time{}, true, "expected type string"},
		{"invalid format returns error", map[string]interface{}{"createdAt": "not-a-time"}, "createdAt", defaultTime, time.Time{}, true, "invalid time format"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetOptionalTime(tt.data, tt.key, tt.defaultVal)
			if checkError(t, "GetOptionalTime", err, tt.wantErr, tt.errContains) && !got.Equal(tt.want) {
				t.Errorf("GetOptionalTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetRequiredMap(t *testing.T) {
	tests := []struct {
		name        string
		data        map[string]interface{}
		key         string
		wantNil     bool
		wantErr     bool
		errContains string
	}{
		{"valid map", map[string]interface{}{"metadata": map[string]interface{}{"key": "value"}}, "metadata", false, false, ""},
		{"missing field", map[string]interface{}{}, "metadata", true, true, "required field is missing"},
		{"wrong type", map[string]interface{}{"metadata": "not a map"}, "metadata", true, true, "expected type map"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetRequiredMap(tt.data, tt.key)
			if checkError(t, "GetRequiredMap", err, tt.wantErr, tt.errContains) && got == nil {
				t.Errorf("GetRequiredMap() returned nil for valid input")
			}
		})
	}
}

func TestGetOptionalMap(t *testing.T) {
	tests := []struct {
		name        string
		data        map[string]interface{}
		key         string
		wantNil     bool
		wantErr     bool
		errContains string
	}{
		{"valid map", map[string]interface{}{"metadata": map[string]interface{}{"key": "value"}}, "metadata", false, false, ""},
		{"missing field returns nil", map[string]interface{}{}, "metadata", true, false, ""},
		{"null value returns nil", map[string]interface{}{"metadata": nil}, "metadata", true, false, ""},
		{"wrong type returns error", map[string]interface{}{"metadata": "not a map"}, "metadata", true, true, "expected type map"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetOptionalMap(tt.data, tt.key)
			if !checkError(t, "GetOptionalMap", err, tt.wantErr, tt.errContains) {
				return
			}
			if tt.wantNil && got != nil {
				t.Errorf("GetOptionalMap() = %v, want nil", got)
			}
			if !tt.wantNil && got == nil {
				t.Errorf("GetOptionalMap() = nil, want non-nil")
			}
		})
	}
}

func TestGetRequiredMapSlice(t *testing.T) {
	tests := []struct {
		name        string
		data        map[string]interface{}
		key         string
		wantLen     int
		wantErr     bool
		errContains string
	}{
		{
			"valid map slice",
			map[string]interface{}{"items": []interface{}{map[string]interface{}{"id": "1"}, map[string]interface{}{"id": "2"}}},
			"items", 2, false, "",
		},
		{"missing field", map[string]interface{}{}, "items", 0, true, "required field is missing"},
		{"wrong type - not a slice", map[string]interface{}{"items": "not a slice"}, "items", 0, true, "expected type"},
		{
			"slice with non-map items",
			map[string]interface{}{"items": []interface{}{"string item"}},
			"items", 0, true, "expected type",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetRequiredMapSlice(tt.data, tt.key)
			if checkError(t, "GetRequiredMapSlice", err, tt.wantErr, tt.errContains) && len(got) != tt.wantLen {
				t.Errorf("GetRequiredMapSlice() returned %d items, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func TestGetOptionalMapSlice(t *testing.T) {
	tests := []struct {
		name        string
		data        map[string]interface{}
		key         string
		wantLen     int
		wantNil     bool
		wantErr     bool
		errContains string
	}{
		{
			"valid map slice",
			map[string]interface{}{"items": []interface{}{map[string]interface{}{"id": "1"}, map[string]interface{}{"id": "2"}}},
			"items", 2, false, false, "",
		},
		{"missing field returns nil", map[string]interface{}{}, "items", 0, true, false, ""},
		{"null value returns nil", map[string]interface{}{"items": nil}, "items", 0, true, false, ""},
		{"wrong type returns error", map[string]interface{}{"items": "not a slice"}, "items", 0, true, true, "expected type"},
		{
			"slice with non-map items returns error",
			map[string]interface{}{"items": []interface{}{"string item"}},
			"items", 0, true, true, "expected type",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetOptionalMapSlice(tt.data, tt.key)
			if !checkError(t, "GetOptionalMapSlice", err, tt.wantErr, tt.errContains) {
				return
			}
			if tt.wantNil && got != nil {
				t.Errorf("GetOptionalMapSlice() = %v, want nil", got)
			}
			if !tt.wantNil && len(got) != tt.wantLen {
				t.Errorf("GetOptionalMapSlice() returned %d items, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func TestGetRequiredStringSlice(t *testing.T) {
	tests := []struct {
		name        string
		data        map[string]interface{}
		key         string
		want        []string
		wantErr     bool
		errContains string
	}{
		{"valid string slice", map[string]interface{}{"tags": []interface{}{"a", "b", "c"}}, "tags", []string{"a", "b", "c"}, false, ""},
		{"missing field", map[string]interface{}{}, "tags", nil, true, "required field is missing"},
		{"slice with non-string items", map[string]interface{}{"tags": []interface{}{1, 2, 3}}, "tags", nil, true, "expected type"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetRequiredStringSlice(tt.data, tt.key)
			if checkError(t, "GetRequiredStringSlice", err, tt.wantErr, tt.errContains) && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetRequiredStringSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetOptionalStringSlice(t *testing.T) {
	tests := []struct {
		name        string
		data        map[string]interface{}
		key         string
		want        []string
		wantNil     bool
		wantErr     bool
		errContains string
	}{
		{"valid string slice", map[string]interface{}{"tags": []interface{}{"a", "b", "c"}}, "tags", []string{"a", "b", "c"}, false, false, ""},
		{"missing field returns nil", map[string]interface{}{}, "tags", nil, true, false, ""},
		{"null value returns nil", map[string]interface{}{"tags": nil}, "tags", nil, true, false, ""},
		{"wrong type returns error", map[string]interface{}{"tags": "not a slice"}, "tags", nil, true, true, "expected type"},
		{"slice with non-string items returns error", map[string]interface{}{"tags": []interface{}{1, 2, 3}}, "tags", nil, true, true, "expected type"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetOptionalStringSlice(tt.data, tt.key)
			if !checkError(t, "GetOptionalStringSlice", err, tt.wantErr, tt.errContains) {
				return
			}
			if tt.wantNil && got != nil {
				t.Errorf("GetOptionalStringSlice() = %v, want nil", got)
			}
			if !tt.wantNil && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetOptionalStringSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeserializationError(t *testing.T) {
	t.Run("error with field name", func(t *testing.T) {
		err := NewFieldDeserializationError("agg-123", "UserCreated", 5, "email", NewMissingFieldError("email"))
		errStr := err.Error()
		for _, want := range []string{"agg-123", "UserCreated", "5", "email"} {
			if !strings.Contains(errStr, want) {
				t.Errorf("error should contain %q, got: %s", want, errStr)
			}
		}
	})

	t.Run("error without field name", func(t *testing.T) {
		cause := NewTypeError("id", "string", "int")
		err := NewDeserializationError("agg-456", "OrderPlaced", 3, cause)
		errStr := err.Error()
		for _, want := range []string{"agg-456", "OrderPlaced"} {
			if !strings.Contains(errStr, want) {
				t.Errorf("error should contain %q, got: %s", want, errStr)
			}
		}
	})

	t.Run("unwrap returns cause", func(t *testing.T) {
		cause := NewMissingFieldError("name")
		err := NewDeserializationError("agg-789", "ItemAdded", 1, cause)
		if err.Unwrap() != cause {
			t.Errorf("Unwrap() should return the cause error")
		}
	})
}
