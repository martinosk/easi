package repository

import (
	"fmt"
	"time"
)

func GetRequiredString(data map[string]interface{}, key string) (string, error) {
	val, exists := data[key]
	if !exists || val == nil {
		return "", NewMissingFieldError(key)
	}
	str, ok := val.(string)
	if !ok {
		return "", NewTypeError(key, "string", fmt.Sprintf("%T", val))
	}
	return str, nil
}

func GetOptionalString(data map[string]interface{}, key string, defaultVal string) (string, error) {
	val, exists := data[key]
	if !exists || val == nil {
		return defaultVal, nil
	}
	str, ok := val.(string)
	if !ok {
		return "", NewTypeError(key, "string", fmt.Sprintf("%T", val))
	}
	return str, nil
}

func GetRequiredInt(data map[string]interface{}, key string) (int, error) {
	val, exists := data[key]
	if !exists || val == nil {
		return 0, NewMissingFieldError(key)
	}
	switch v := val.(type) {
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	default:
		return 0, NewTypeError(key, "int", fmt.Sprintf("%T", val))
	}
}

func GetOptionalInt(data map[string]interface{}, key string, defaultVal int) (int, error) {
	val, exists := data[key]
	if !exists || val == nil {
		return defaultVal, nil
	}
	switch v := val.(type) {
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	default:
		return 0, NewTypeError(key, "int", fmt.Sprintf("%T", val))
	}
}

func GetRequiredFloat64(data map[string]interface{}, key string) (float64, error) {
	val, exists := data[key]
	if !exists || val == nil {
		return 0, NewMissingFieldError(key)
	}
	switch v := val.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	default:
		return 0, NewTypeError(key, "float64", fmt.Sprintf("%T", val))
	}
}

func GetOptionalFloat64(data map[string]interface{}, key string, defaultVal float64) (float64, error) {
	val, exists := data[key]
	if !exists || val == nil {
		return defaultVal, nil
	}
	switch v := val.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	default:
		return 0, NewTypeError(key, "float64", fmt.Sprintf("%T", val))
	}
}

func GetRequiredBool(data map[string]interface{}, key string) (bool, error) {
	val, exists := data[key]
	if !exists || val == nil {
		return false, NewMissingFieldError(key)
	}
	b, ok := val.(bool)
	if !ok {
		return false, NewTypeError(key, "bool", fmt.Sprintf("%T", val))
	}
	return b, nil
}

func GetOptionalBool(data map[string]interface{}, key string, defaultVal bool) (bool, error) {
	val, exists := data[key]
	if !exists || val == nil {
		return defaultVal, nil
	}
	b, ok := val.(bool)
	if !ok {
		return false, NewTypeError(key, "bool", fmt.Sprintf("%T", val))
	}
	return b, nil
}

func GetRequiredTime(data map[string]interface{}, key string) (time.Time, error) {
	val, exists := data[key]
	if !exists || val == nil {
		return time.Time{}, NewMissingFieldError(key)
	}
	str, ok := val.(string)
	if !ok {
		return time.Time{}, NewTypeError(key, "string (RFC3339)", fmt.Sprintf("%T", val))
	}
	t, err := time.Parse(time.RFC3339Nano, str)
	if err != nil {
		t, err = time.Parse(time.RFC3339, str)
		if err != nil {
			return time.Time{}, &FieldError{
				FieldName: key,
				Message:   fmt.Sprintf("invalid time format: %v", err),
			}
		}
	}
	return t, nil
}

func GetOptionalTime(data map[string]interface{}, key string, defaultVal time.Time) (time.Time, error) {
	val, exists := data[key]
	if !exists || val == nil {
		return defaultVal, nil
	}
	str, ok := val.(string)
	if !ok {
		return time.Time{}, NewTypeError(key, "string (RFC3339)", fmt.Sprintf("%T", val))
	}
	t, err := time.Parse(time.RFC3339Nano, str)
	if err != nil {
		t, err = time.Parse(time.RFC3339, str)
		if err != nil {
			return time.Time{}, &FieldError{
				FieldName: key,
				Message:   fmt.Sprintf("invalid time format: %v", err),
			}
		}
	}
	return t, nil
}

func GetRequiredMap(data map[string]interface{}, key string) (map[string]interface{}, error) {
	val, exists := data[key]
	if !exists || val == nil {
		return nil, NewMissingFieldError(key)
	}
	m, ok := val.(map[string]interface{})
	if !ok {
		return nil, NewTypeError(key, "map[string]interface{}", fmt.Sprintf("%T", val))
	}
	return m, nil
}

func GetOptionalMap(data map[string]interface{}, key string) (map[string]interface{}, error) {
	val, exists := data[key]
	if !exists || val == nil {
		return nil, nil
	}
	m, ok := val.(map[string]interface{})
	if !ok {
		return nil, NewTypeError(key, "map[string]interface{}", fmt.Sprintf("%T", val))
	}
	return m, nil
}

func GetRequiredMapSlice(data map[string]interface{}, key string) ([]map[string]interface{}, error) {
	val, exists := data[key]
	if !exists || val == nil {
		return nil, NewMissingFieldError(key)
	}
	raw, ok := val.([]interface{})
	if !ok {
		return nil, NewTypeError(key, "[]interface{}", fmt.Sprintf("%T", val))
	}
	result := make([]map[string]interface{}, 0, len(raw))
	for _, item := range raw {
		m, ok := item.(map[string]interface{})
		if !ok {
			return nil, NewTypeError(key, "[]map[string]interface{}", fmt.Sprintf("slice containing %T", item))
		}
		result = append(result, m)
	}
	return result, nil
}

func GetOptionalMapSlice(data map[string]interface{}, key string) ([]map[string]interface{}, error) {
	val, exists := data[key]
	if !exists || val == nil {
		return nil, nil
	}
	raw, ok := val.([]interface{})
	if !ok {
		return nil, NewTypeError(key, "[]interface{}", fmt.Sprintf("%T", val))
	}
	result := make([]map[string]interface{}, 0, len(raw))
	for _, item := range raw {
		m, ok := item.(map[string]interface{})
		if !ok {
			return nil, NewTypeError(key, "[]map[string]interface{}", fmt.Sprintf("slice containing %T", item))
		}
		result = append(result, m)
	}
	return result, nil
}

func GetRequiredStringSlice(data map[string]interface{}, key string) ([]string, error) {
	val, exists := data[key]
	if !exists || val == nil {
		return nil, NewMissingFieldError(key)
	}
	raw, ok := val.([]interface{})
	if !ok {
		return nil, NewTypeError(key, "[]interface{}", fmt.Sprintf("%T", val))
	}
	result := make([]string, 0, len(raw))
	for _, item := range raw {
		s, ok := item.(string)
		if !ok {
			return nil, NewTypeError(key, "[]string", fmt.Sprintf("slice containing %T", item))
		}
		result = append(result, s)
	}
	return result, nil
}

func GetOptionalStringSlice(data map[string]interface{}, key string) ([]string, error) {
	val, exists := data[key]
	if !exists || val == nil {
		return nil, nil
	}
	raw, ok := val.([]interface{})
	if !ok {
		return nil, NewTypeError(key, "[]interface{}", fmt.Sprintf("%T", val))
	}
	result := make([]string, 0, len(raw))
	for _, item := range raw {
		s, ok := item.(string)
		if !ok {
			return nil, NewTypeError(key, "[]string", fmt.Sprintf("slice containing %T", item))
		}
		result = append(result, s)
	}
	return result, nil
}
