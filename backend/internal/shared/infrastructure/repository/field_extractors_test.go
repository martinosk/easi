package repository

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

type testResult struct {
	t           *testing.T
	funcName    string
	err         error
	wantErr     bool
	errContains string
}

func (r testResult) checkError() bool {
	r.t.Helper()
	if r.hasUnexpectedError() {
		return false
	}
	if r.hasMismatchedErrorContent() {
		return false
	}
	return !r.wantErr
}

func (r testResult) hasUnexpectedError() bool {
	if (r.err != nil) != r.wantErr {
		r.t.Errorf("%s() error = %v, wantErr %v", r.funcName, r.err, r.wantErr)
		return true
	}
	return false
}

func (r testResult) hasMismatchedErrorContent() bool {
	if !r.wantErr || r.errContains == "" {
		return false
	}
	if r.err == nil || !strings.Contains(r.err.Error(), r.errContains) {
		r.t.Errorf("%s() error = %v, want error containing %q", r.funcName, r.err, r.errContains)
		return true
	}
	return false
}

type valueTestCase[T any] struct {
	name        string
	data        map[string]interface{}
	key         string
	want        T
	wantErr     bool
	errContains string
}

func runValueTests[T any](t *testing.T, funcName string, tests []valueTestCase[T], extract func(map[string]interface{}, string) (T, error), equal func(T, T) bool) {
	t.Helper()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extract(tt.data, tt.key)
			r := testResult{t, funcName, err, tt.wantErr, tt.errContains}
			if r.checkError() && !equal(got, tt.want) {
				t.Errorf("%s() = %v, want %v", funcName, got, tt.want)
			}
		})
	}
}

func eq[T comparable](a, b T) bool { return a == b }

func deepEqual[T any](a, b T) bool { return reflect.DeepEqual(a, b) }

func withDefault[T any](fn func(map[string]interface{}, string, T) (T, error), def T) func(map[string]interface{}, string) (T, error) {
	return func(data map[string]interface{}, key string) (T, error) {
		return fn(data, key, def)
	}
}

func TestGetRequiredString(t *testing.T) {
	runValueTests(t, "GetRequiredString", []valueTestCase[string]{
		{"valid string", map[string]interface{}{"name": "test"}, "name", "test", false, ""},
		{"empty string is valid", map[string]interface{}{"name": ""}, "name", "", false, ""},
		{"missing field", map[string]interface{}{}, "name", "", true, "required field is missing"},
		{"null value", map[string]interface{}{"name": nil}, "name", "", true, "required field is missing"},
		{"wrong type", map[string]interface{}{"name": 123}, "name", "", true, "expected type string"},
	}, GetRequiredString, eq)
}

func TestGetOptionalString(t *testing.T) {
	runValueTests(t, "GetOptionalString", []valueTestCase[string]{
		{"valid string", map[string]interface{}{"name": "test"}, "name", "test", false, ""},
		{"missing field returns default", map[string]interface{}{}, "name", "default", false, ""},
		{"null value returns default", map[string]interface{}{"name": nil}, "name", "default", false, ""},
		{"wrong type returns error", map[string]interface{}{"name": 123}, "name", "", true, "expected type string"},
	}, withDefault(GetOptionalString, "default"), eq)
}

func TestGetRequiredInt(t *testing.T) {
	runValueTests(t, "GetRequiredInt", []valueTestCase[int]{
		{"valid int", map[string]interface{}{"count": 42}, "count", 42, false, ""},
		{"float64 converted to int", map[string]interface{}{"count": float64(42)}, "count", 42, false, ""},
		{"int64 converted to int", map[string]interface{}{"count": int64(42)}, "count", 42, false, ""},
		{"negative int", map[string]interface{}{"count": -100}, "count", -100, false, ""},
		{"missing field", map[string]interface{}{}, "count", 0, true, "required field is missing"},
		{"wrong type", map[string]interface{}{"count": "not a number"}, "count", 0, true, "expected type int"},
	}, GetRequiredInt, eq)
}

func TestGetOptionalInt(t *testing.T) {
	runValueTests(t, "GetOptionalInt", []valueTestCase[int]{
		{"valid int", map[string]interface{}{"count": 42}, "count", 42, false, ""},
		{"missing field returns default", map[string]interface{}{}, "count", 10, false, ""},
		{"null value returns default", map[string]interface{}{"count": nil}, "count", 10, false, ""},
		{"wrong type returns error", map[string]interface{}{"count": "not a number"}, "count", 0, true, "expected type int"},
	}, withDefault(GetOptionalInt, 10), eq)
}

func TestGetRequiredFloat64(t *testing.T) {
	runValueTests(t, "GetRequiredFloat64", []valueTestCase[float64]{
		{"valid float64", map[string]interface{}{"price": 42.5}, "price", 42.5, false, ""},
		{"int converted to float64", map[string]interface{}{"price": 42}, "price", 42.0, false, ""},
		{"missing field", map[string]interface{}{}, "price", 0, true, "required field is missing"},
		{"wrong type", map[string]interface{}{"price": "not a number"}, "price", 0, true, "expected type float64"},
	}, GetRequiredFloat64, eq)
}

func TestGetOptionalFloat64(t *testing.T) {
	runValueTests(t, "GetOptionalFloat64", []valueTestCase[float64]{
		{"valid float64", map[string]interface{}{"price": 42.5}, "price", 42.5, false, ""},
		{"missing field returns default", map[string]interface{}{}, "price", 10.5, false, ""},
		{"null value returns default", map[string]interface{}{"price": nil}, "price", 10.5, false, ""},
		{"wrong type returns error", map[string]interface{}{"price": "not a number"}, "price", 0, true, "expected type float64"},
	}, withDefault(GetOptionalFloat64, 10.5), eq)
}

func TestGetRequiredBool(t *testing.T) {
	runValueTests(t, "GetRequiredBool", []valueTestCase[bool]{
		{"valid true", map[string]interface{}{"active": true}, "active", true, false, ""},
		{"valid false", map[string]interface{}{"active": false}, "active", false, false, ""},
		{"missing field", map[string]interface{}{}, "active", false, true, "required field is missing"},
		{"wrong type", map[string]interface{}{"active": "true"}, "active", false, true, "expected type bool"},
	}, GetRequiredBool, eq)
}

func TestGetOptionalBool(t *testing.T) {
	runValueTests(t, "GetOptionalBool", []valueTestCase[bool]{
		{"valid bool", map[string]interface{}{"active": true}, "active", true, false, ""},
		{"missing field returns default", map[string]interface{}{}, "active", true, false, ""},
		{"null value returns default", map[string]interface{}{"active": nil}, "active", true, false, ""},
		{"wrong type returns error", map[string]interface{}{"active": "true"}, "active", false, true, "expected type bool"},
		{"int instead of bool returns error", map[string]interface{}{"active": 1}, "active", false, true, "expected type bool"},
	}, withDefault(GetOptionalBool, true), eq)
}

func TestGetRequiredTime(t *testing.T) {
	validTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	validTimeNano := time.Date(2024, 1, 15, 10, 30, 0, 123456789, time.UTC)
	runValueTests(t, "GetRequiredTime", []valueTestCase[time.Time]{
		{"valid RFC3339 time", map[string]interface{}{"createdAt": "2024-01-15T10:30:00Z"}, "createdAt", validTime, false, ""},
		{"valid RFC3339Nano time", map[string]interface{}{"createdAt": "2024-01-15T10:30:00.123456789Z"}, "createdAt", validTimeNano, false, ""},
		{"missing field", map[string]interface{}{}, "createdAt", time.Time{}, true, "required field is missing"},
		{"wrong type", map[string]interface{}{"createdAt": 12345}, "createdAt", time.Time{}, true, "expected type string"},
		{"invalid time format", map[string]interface{}{"createdAt": "not-a-time"}, "createdAt", time.Time{}, true, "invalid time format"},
	}, GetRequiredTime, func(a, b time.Time) bool { return a.Equal(b) })
}

func TestGetOptionalTime(t *testing.T) {
	defaultTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	validTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	runValueTests(t, "GetOptionalTime", []valueTestCase[time.Time]{
		{"valid time", map[string]interface{}{"createdAt": "2024-01-15T10:30:00Z"}, "createdAt", validTime, false, ""},
		{"missing field returns default", map[string]interface{}{}, "createdAt", defaultTime, false, ""},
		{"null value returns default", map[string]interface{}{"createdAt": nil}, "createdAt", defaultTime, false, ""},
		{"wrong type returns error", map[string]interface{}{"createdAt": 12345}, "createdAt", time.Time{}, true, "expected type string"},
		{"invalid format returns error", map[string]interface{}{"createdAt": "not-a-time"}, "createdAt", time.Time{}, true, "invalid time format"},
	}, withDefault(GetOptionalTime, defaultTime), func(a, b time.Time) bool { return a.Equal(b) })
}

func TestGetRequiredMap(t *testing.T) {
	runValueTests(t, "GetRequiredMap", []valueTestCase[map[string]interface{}]{
		{"valid map", map[string]interface{}{"metadata": map[string]interface{}{"key": "value"}}, "metadata", map[string]interface{}{"key": "value"}, false, ""},
		{"missing field", map[string]interface{}{}, "metadata", nil, true, "required field is missing"},
		{"wrong type", map[string]interface{}{"metadata": "not a map"}, "metadata", nil, true, "expected type map"},
	}, GetRequiredMap, deepEqual[map[string]interface{}])
}

func TestGetOptionalMap(t *testing.T) {
	runValueTests(t, "GetOptionalMap", []valueTestCase[map[string]interface{}]{
		{"valid map", map[string]interface{}{"metadata": map[string]interface{}{"key": "value"}}, "metadata", map[string]interface{}{"key": "value"}, false, ""},
		{"missing field returns nil", map[string]interface{}{}, "metadata", nil, false, ""},
		{"null value returns nil", map[string]interface{}{"metadata": nil}, "metadata", nil, false, ""},
		{"wrong type returns error", map[string]interface{}{"metadata": "not a map"}, "metadata", nil, true, "expected type map"},
	}, GetOptionalMap, deepEqual[map[string]interface{}])
}

func TestGetRequiredMapSlice(t *testing.T) {
	twoItems := []map[string]interface{}{{"id": "1"}, {"id": "2"}}
	runValueTests(t, "GetRequiredMapSlice", []valueTestCase[[]map[string]interface{}]{
		{
			"valid map slice",
			map[string]interface{}{"items": []interface{}{map[string]interface{}{"id": "1"}, map[string]interface{}{"id": "2"}}},
			"items", twoItems, false, "",
		},
		{"missing field", map[string]interface{}{}, "items", nil, true, "required field is missing"},
		{"wrong type - not a slice", map[string]interface{}{"items": "not a slice"}, "items", nil, true, "expected type"},
		{
			"slice with non-map items",
			map[string]interface{}{"items": []interface{}{"string item"}},
			"items", nil, true, "expected type",
		},
	}, GetRequiredMapSlice, deepEqual[[]map[string]interface{}])
}

func TestGetOptionalMapSlice(t *testing.T) {
	twoItems := []map[string]interface{}{{"id": "1"}, {"id": "2"}}
	runValueTests(t, "GetOptionalMapSlice", []valueTestCase[[]map[string]interface{}]{
		{
			"valid map slice",
			map[string]interface{}{"items": []interface{}{map[string]interface{}{"id": "1"}, map[string]interface{}{"id": "2"}}},
			"items", twoItems, false, "",
		},
		{"missing field returns nil", map[string]interface{}{}, "items", nil, false, ""},
		{"null value returns nil", map[string]interface{}{"items": nil}, "items", nil, false, ""},
		{"wrong type returns error", map[string]interface{}{"items": "not a slice"}, "items", nil, true, "expected type"},
		{
			"slice with non-map items returns error",
			map[string]interface{}{"items": []interface{}{"string item"}},
			"items", nil, true, "expected type",
		},
	}, GetOptionalMapSlice, deepEqual[[]map[string]interface{}])
}

func TestGetRequiredStringSlice(t *testing.T) {
	runValueTests(t, "GetRequiredStringSlice", []valueTestCase[[]string]{
		{"valid string slice", map[string]interface{}{"tags": []interface{}{"a", "b", "c"}}, "tags", []string{"a", "b", "c"}, false, ""},
		{"missing field", map[string]interface{}{}, "tags", nil, true, "required field is missing"},
		{"slice with non-string items", map[string]interface{}{"tags": []interface{}{1, 2, 3}}, "tags", nil, true, "expected type"},
	}, GetRequiredStringSlice, deepEqual[[]string])
}

func TestGetOptionalStringSlice(t *testing.T) {
	runValueTests(t, "GetOptionalStringSlice", []valueTestCase[[]string]{
		{"valid string slice", map[string]interface{}{"tags": []interface{}{"a", "b", "c"}}, "tags", []string{"a", "b", "c"}, false, ""},
		{"missing field returns nil", map[string]interface{}{}, "tags", nil, false, ""},
		{"null value returns nil", map[string]interface{}{"tags": nil}, "tags", nil, false, ""},
		{"wrong type returns error", map[string]interface{}{"tags": "not a slice"}, "tags", nil, true, "expected type"},
		{"slice with non-string items returns error", map[string]interface{}{"tags": []interface{}{1, 2, 3}}, "tags", nil, true, "expected type"},
	}, GetOptionalStringSlice, deepEqual[[]string])
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
