package repository

import (
	"strings"
	"testing"
	"time"
)

func TestGetRequiredString(t *testing.T) {
	tests := []struct {
		name        string
		data        map[string]interface{}
		key         string
		want        string
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid string",
			data:    map[string]interface{}{"name": "test"},
			key:     "name",
			want:    "test",
			wantErr: false,
		},
		{
			name:    "empty string is valid",
			data:    map[string]interface{}{"name": ""},
			key:     "name",
			want:    "",
			wantErr: false,
		},
		{
			name:        "missing field",
			data:        map[string]interface{}{},
			key:         "name",
			wantErr:     true,
			errContains: "required field is missing",
		},
		{
			name:        "null value",
			data:        map[string]interface{}{"name": nil},
			key:         "name",
			wantErr:     true,
			errContains: "required field is missing",
		},
		{
			name:        "wrong type",
			data:        map[string]interface{}{"name": 123},
			key:         "name",
			wantErr:     true,
			errContains: "expected type string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetRequiredString(tt.data, tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRequiredString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errContains != "" {
				if err == nil || !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("GetRequiredString() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}
			if got != tt.want {
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
		{
			name:       "valid string",
			data:       map[string]interface{}{"name": "test"},
			key:        "name",
			defaultVal: "default",
			want:       "test",
			wantErr:    false,
		},
		{
			name:       "missing field returns default",
			data:       map[string]interface{}{},
			key:        "name",
			defaultVal: "default",
			want:       "default",
			wantErr:    false,
		},
		{
			name:       "null value returns default",
			data:       map[string]interface{}{"name": nil},
			key:        "name",
			defaultVal: "default",
			want:       "default",
			wantErr:    false,
		},
		{
			name:        "wrong type returns error",
			data:        map[string]interface{}{"name": 123},
			key:         "name",
			defaultVal:  "default",
			wantErr:     true,
			errContains: "expected type string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetOptionalString(tt.data, tt.key, tt.defaultVal)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetOptionalString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errContains != "" {
				if err == nil || !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("GetOptionalString() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}
			if got != tt.want {
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
		{
			name:    "valid int",
			data:    map[string]interface{}{"count": 42},
			key:     "count",
			want:    42,
			wantErr: false,
		},
		{
			name:    "float64 converted to int",
			data:    map[string]interface{}{"count": float64(42)},
			key:     "count",
			want:    42,
			wantErr: false,
		},
		{
			name:    "int64 converted to int",
			data:    map[string]interface{}{"count": int64(42)},
			key:     "count",
			want:    42,
			wantErr: false,
		},
		{
			name:    "negative int",
			data:    map[string]interface{}{"count": -100},
			key:     "count",
			want:    -100,
			wantErr: false,
		},
		{
			name:        "missing field",
			data:        map[string]interface{}{},
			key:         "count",
			wantErr:     true,
			errContains: "required field is missing",
		},
		{
			name:        "wrong type",
			data:        map[string]interface{}{"count": "not a number"},
			key:         "count",
			wantErr:     true,
			errContains: "expected type int",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetRequiredInt(tt.data, tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRequiredInt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errContains != "" {
				if err == nil || !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("GetRequiredInt() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}
			if got != tt.want {
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
		{
			name:       "valid int",
			data:       map[string]interface{}{"count": 42},
			key:        "count",
			defaultVal: 0,
			want:       42,
			wantErr:    false,
		},
		{
			name:       "missing field returns default",
			data:       map[string]interface{}{},
			key:        "count",
			defaultVal: 10,
			want:       10,
			wantErr:    false,
		},
		{
			name:       "null value returns default",
			data:       map[string]interface{}{"count": nil},
			key:        "count",
			defaultVal: 10,
			want:       10,
			wantErr:    false,
		},
		{
			name:        "wrong type returns error",
			data:        map[string]interface{}{"count": "not a number"},
			key:         "count",
			defaultVal:  10,
			wantErr:     true,
			errContains: "expected type int",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetOptionalInt(tt.data, tt.key, tt.defaultVal)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetOptionalInt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errContains != "" {
				if err == nil || !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("GetOptionalInt() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}
			if got != tt.want {
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
		{
			name:    "valid float64",
			data:    map[string]interface{}{"price": 42.5},
			key:     "price",
			want:    42.5,
			wantErr: false,
		},
		{
			name:    "int converted to float64",
			data:    map[string]interface{}{"price": 42},
			key:     "price",
			want:    42.0,
			wantErr: false,
		},
		{
			name:        "missing field",
			data:        map[string]interface{}{},
			key:         "price",
			wantErr:     true,
			errContains: "required field is missing",
		},
		{
			name:        "wrong type",
			data:        map[string]interface{}{"price": "not a number"},
			key:         "price",
			wantErr:     true,
			errContains: "expected type float64",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetRequiredFloat64(tt.data, tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRequiredFloat64() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errContains != "" {
				if err == nil || !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("GetRequiredFloat64() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}
			if got != tt.want {
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
		{
			name:       "valid float64",
			data:       map[string]interface{}{"price": 42.5},
			key:        "price",
			defaultVal: 0.0,
			want:       42.5,
			wantErr:    false,
		},
		{
			name:       "missing field returns default",
			data:       map[string]interface{}{},
			key:        "price",
			defaultVal: 10.5,
			want:       10.5,
			wantErr:    false,
		},
		{
			name:       "null value returns default",
			data:       map[string]interface{}{"price": nil},
			key:        "price",
			defaultVal: 10.5,
			want:       10.5,
			wantErr:    false,
		},
		{
			name:        "wrong type returns error",
			data:        map[string]interface{}{"price": "not a number"},
			key:         "price",
			defaultVal:  10.5,
			wantErr:     true,
			errContains: "expected type float64",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetOptionalFloat64(tt.data, tt.key, tt.defaultVal)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetOptionalFloat64() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errContains != "" {
				if err == nil || !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("GetOptionalFloat64() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}
			if got != tt.want {
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
		{
			name:    "valid true",
			data:    map[string]interface{}{"active": true},
			key:     "active",
			want:    true,
			wantErr: false,
		},
		{
			name:    "valid false",
			data:    map[string]interface{}{"active": false},
			key:     "active",
			want:    false,
			wantErr: false,
		},
		{
			name:        "missing field",
			data:        map[string]interface{}{},
			key:         "active",
			wantErr:     true,
			errContains: "required field is missing",
		},
		{
			name:        "wrong type",
			data:        map[string]interface{}{"active": "true"},
			key:         "active",
			wantErr:     true,
			errContains: "expected type bool",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetRequiredBool(tt.data, tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRequiredBool() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errContains != "" {
				if err == nil || !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("GetRequiredBool() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}
			if got != tt.want {
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
		{
			name:       "valid bool",
			data:       map[string]interface{}{"active": true},
			key:        "active",
			defaultVal: false,
			want:       true,
			wantErr:    false,
		},
		{
			name:       "missing field returns default",
			data:       map[string]interface{}{},
			key:        "active",
			defaultVal: true,
			want:       true,
			wantErr:    false,
		},
		{
			name:       "null value returns default",
			data:       map[string]interface{}{"active": nil},
			key:        "active",
			defaultVal: true,
			want:       true,
			wantErr:    false,
		},
		{
			name:        "wrong type returns error",
			data:        map[string]interface{}{"active": "true"},
			key:         "active",
			defaultVal:  false,
			wantErr:     true,
			errContains: "expected type bool",
		},
		{
			name:        "int instead of bool returns error",
			data:        map[string]interface{}{"active": 1},
			key:         "active",
			defaultVal:  false,
			wantErr:     true,
			errContains: "expected type bool",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetOptionalBool(tt.data, tt.key, tt.defaultVal)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetOptionalBool() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errContains != "" {
				if err == nil || !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("GetOptionalBool() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}
			if got != tt.want {
				t.Errorf("GetOptionalBool() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetRequiredTime(t *testing.T) {
	validTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	tests := []struct {
		name        string
		data        map[string]interface{}
		key         string
		want        time.Time
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid RFC3339 time",
			data:    map[string]interface{}{"createdAt": "2024-01-15T10:30:00Z"},
			key:     "createdAt",
			want:    validTime,
			wantErr: false,
		},
		{
			name:    "valid RFC3339Nano time",
			data:    map[string]interface{}{"createdAt": "2024-01-15T10:30:00.123456789Z"},
			key:     "createdAt",
			want:    time.Date(2024, 1, 15, 10, 30, 0, 123456789, time.UTC),
			wantErr: false,
		},
		{
			name:        "missing field",
			data:        map[string]interface{}{},
			key:         "createdAt",
			wantErr:     true,
			errContains: "required field is missing",
		},
		{
			name:        "wrong type",
			data:        map[string]interface{}{"createdAt": 12345},
			key:         "createdAt",
			wantErr:     true,
			errContains: "expected type string",
		},
		{
			name:        "invalid time format",
			data:        map[string]interface{}{"createdAt": "not-a-time"},
			key:         "createdAt",
			wantErr:     true,
			errContains: "invalid time format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetRequiredTime(tt.data, tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRequiredTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errContains != "" {
				if err == nil || !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("GetRequiredTime() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}
			if !got.Equal(tt.want) {
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
		{
			name:       "valid time",
			data:       map[string]interface{}{"createdAt": "2024-01-15T10:30:00Z"},
			key:        "createdAt",
			defaultVal: defaultTime,
			want:       validTime,
			wantErr:    false,
		},
		{
			name:       "missing field returns default",
			data:       map[string]interface{}{},
			key:        "createdAt",
			defaultVal: defaultTime,
			want:       defaultTime,
			wantErr:    false,
		},
		{
			name:       "null value returns default",
			data:       map[string]interface{}{"createdAt": nil},
			key:        "createdAt",
			defaultVal: defaultTime,
			want:       defaultTime,
			wantErr:    false,
		},
		{
			name:        "wrong type returns error",
			data:        map[string]interface{}{"createdAt": 12345},
			key:         "createdAt",
			defaultVal:  defaultTime,
			wantErr:     true,
			errContains: "expected type string",
		},
		{
			name:        "invalid format returns error",
			data:        map[string]interface{}{"createdAt": "not-a-time"},
			key:         "createdAt",
			defaultVal:  defaultTime,
			wantErr:     true,
			errContains: "invalid time format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetOptionalTime(tt.data, tt.key, tt.defaultVal)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetOptionalTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errContains != "" {
				if err == nil || !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("GetOptionalTime() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}
			if !got.Equal(tt.want) {
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
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid map",
			data:    map[string]interface{}{"metadata": map[string]interface{}{"key": "value"}},
			key:     "metadata",
			wantErr: false,
		},
		{
			name:        "missing field",
			data:        map[string]interface{}{},
			key:         "metadata",
			wantErr:     true,
			errContains: "required field is missing",
		},
		{
			name:        "wrong type",
			data:        map[string]interface{}{"metadata": "not a map"},
			key:         "metadata",
			wantErr:     true,
			errContains: "expected type map",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetRequiredMap(tt.data, tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRequiredMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errContains != "" {
				if err == nil || !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("GetRequiredMap() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}
			if !tt.wantErr && got == nil {
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
		{
			name:    "valid map",
			data:    map[string]interface{}{"metadata": map[string]interface{}{"key": "value"}},
			key:     "metadata",
			wantNil: false,
			wantErr: false,
		},
		{
			name:    "missing field returns nil",
			data:    map[string]interface{}{},
			key:     "metadata",
			wantNil: true,
			wantErr: false,
		},
		{
			name:    "null value returns nil",
			data:    map[string]interface{}{"metadata": nil},
			key:     "metadata",
			wantNil: true,
			wantErr: false,
		},
		{
			name:        "wrong type returns error",
			data:        map[string]interface{}{"metadata": "not a map"},
			key:         "metadata",
			wantErr:     true,
			errContains: "expected type map",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetOptionalMap(tt.data, tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetOptionalMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errContains != "" {
				if err == nil || !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("GetOptionalMap() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}
			if !tt.wantErr && tt.wantNil && got != nil {
				t.Errorf("GetOptionalMap() = %v, want nil", got)
			}
			if !tt.wantErr && !tt.wantNil && got == nil {
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
			name: "valid map slice",
			data: map[string]interface{}{
				"items": []interface{}{
					map[string]interface{}{"id": "1"},
					map[string]interface{}{"id": "2"},
				},
			},
			key:     "items",
			wantLen: 2,
			wantErr: false,
		},
		{
			name:        "missing field",
			data:        map[string]interface{}{},
			key:         "items",
			wantErr:     true,
			errContains: "required field is missing",
		},
		{
			name:        "wrong type - not a slice",
			data:        map[string]interface{}{"items": "not a slice"},
			key:         "items",
			wantErr:     true,
			errContains: "expected type",
		},
		{
			name: "slice with non-map items",
			data: map[string]interface{}{
				"items": []interface{}{"string item"},
			},
			key:         "items",
			wantErr:     true,
			errContains: "expected type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetRequiredMapSlice(tt.data, tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRequiredMapSlice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errContains != "" {
				if err == nil || !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("GetRequiredMapSlice() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}
			if !tt.wantErr && len(got) != tt.wantLen {
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
			name: "valid map slice",
			data: map[string]interface{}{
				"items": []interface{}{
					map[string]interface{}{"id": "1"},
					map[string]interface{}{"id": "2"},
				},
			},
			key:     "items",
			wantLen: 2,
			wantErr: false,
		},
		{
			name:    "missing field returns nil",
			data:    map[string]interface{}{},
			key:     "items",
			wantNil: true,
			wantErr: false,
		},
		{
			name:    "null value returns nil",
			data:    map[string]interface{}{"items": nil},
			key:     "items",
			wantNil: true,
			wantErr: false,
		},
		{
			name:        "wrong type returns error",
			data:        map[string]interface{}{"items": "not a slice"},
			key:         "items",
			wantErr:     true,
			errContains: "expected type",
		},
		{
			name: "slice with non-map items returns error",
			data: map[string]interface{}{
				"items": []interface{}{"string item"},
			},
			key:         "items",
			wantErr:     true,
			errContains: "expected type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetOptionalMapSlice(tt.data, tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetOptionalMapSlice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errContains != "" {
				if err == nil || !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("GetOptionalMapSlice() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}
			if !tt.wantErr && tt.wantNil && got != nil {
				t.Errorf("GetOptionalMapSlice() = %v, want nil", got)
			}
			if !tt.wantErr && !tt.wantNil && len(got) != tt.wantLen {
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
		{
			name:    "valid string slice",
			data:    map[string]interface{}{"tags": []interface{}{"a", "b", "c"}},
			key:     "tags",
			want:    []string{"a", "b", "c"},
			wantErr: false,
		},
		{
			name:        "missing field",
			data:        map[string]interface{}{},
			key:         "tags",
			wantErr:     true,
			errContains: "required field is missing",
		},
		{
			name:        "slice with non-string items",
			data:        map[string]interface{}{"tags": []interface{}{1, 2, 3}},
			key:         "tags",
			wantErr:     true,
			errContains: "expected type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetRequiredStringSlice(tt.data, tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRequiredStringSlice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errContains != "" {
				if err == nil || !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("GetRequiredStringSlice() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}
			if !tt.wantErr {
				if len(got) != len(tt.want) {
					t.Errorf("GetRequiredStringSlice() = %v, want %v", got, tt.want)
				}
				for i := range got {
					if got[i] != tt.want[i] {
						t.Errorf("GetRequiredStringSlice()[%d] = %v, want %v", i, got[i], tt.want[i])
					}
				}
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
		{
			name:    "valid string slice",
			data:    map[string]interface{}{"tags": []interface{}{"a", "b", "c"}},
			key:     "tags",
			want:    []string{"a", "b", "c"},
			wantErr: false,
		},
		{
			name:    "missing field returns nil",
			data:    map[string]interface{}{},
			key:     "tags",
			wantNil: true,
			wantErr: false,
		},
		{
			name:    "null value returns nil",
			data:    map[string]interface{}{"tags": nil},
			key:     "tags",
			wantNil: true,
			wantErr: false,
		},
		{
			name:        "wrong type returns error",
			data:        map[string]interface{}{"tags": "not a slice"},
			key:         "tags",
			wantErr:     true,
			errContains: "expected type",
		},
		{
			name:        "slice with non-string items returns error",
			data:        map[string]interface{}{"tags": []interface{}{1, 2, 3}},
			key:         "tags",
			wantErr:     true,
			errContains: "expected type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetOptionalStringSlice(tt.data, tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetOptionalStringSlice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errContains != "" {
				if err == nil || !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("GetOptionalStringSlice() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}
			if !tt.wantErr && tt.wantNil && got != nil {
				t.Errorf("GetOptionalStringSlice() = %v, want nil", got)
			}
			if !tt.wantErr && !tt.wantNil {
				if len(got) != len(tt.want) {
					t.Errorf("GetOptionalStringSlice() = %v, want %v", got, tt.want)
				}
				for i := range got {
					if got[i] != tt.want[i] {
						t.Errorf("GetOptionalStringSlice()[%d] = %v, want %v", i, got[i], tt.want[i])
					}
				}
			}
		})
	}
}

func TestDeserializationError(t *testing.T) {
	t.Run("error with field name", func(t *testing.T) {
		err := NewFieldDeserializationError("agg-123", "UserCreated", 5, "email", NewMissingFieldError("email"))
		errStr := err.Error()
		if !strings.Contains(errStr, "agg-123") {
			t.Errorf("error should contain aggregate ID, got: %s", errStr)
		}
		if !strings.Contains(errStr, "UserCreated") {
			t.Errorf("error should contain event type, got: %s", errStr)
		}
		if !strings.Contains(errStr, "5") {
			t.Errorf("error should contain sequence number, got: %s", errStr)
		}
		if !strings.Contains(errStr, "email") {
			t.Errorf("error should contain field name, got: %s", errStr)
		}
	})

	t.Run("error without field name", func(t *testing.T) {
		cause := NewTypeError("id", "string", "int")
		err := NewDeserializationError("agg-456", "OrderPlaced", 3, cause)
		errStr := err.Error()
		if !strings.Contains(errStr, "agg-456") {
			t.Errorf("error should contain aggregate ID, got: %s", errStr)
		}
		if !strings.Contains(errStr, "OrderPlaced") {
			t.Errorf("error should contain event type, got: %s", errStr)
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

