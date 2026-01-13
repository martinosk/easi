package repository

import (
	"fmt"
	"time"
)

type fieldData = map[string]interface{}

func getField(data fieldData, key string) (interface{}, bool) {
	val, exists := data[key]
	return val, exists && val != nil
}

func GetRequiredString(data fieldData, key string) (string, error) {
	val, exists := getField(data, key)
	if !exists {
		return "", NewMissingFieldError(key)
	}
	str, ok := val.(string)
	if !ok {
		return "", NewTypeError(key, "string", fmt.Sprintf("%T", val))
	}
	return str, nil
}

func GetOptionalString(data fieldData, key string, defaultVal string) (string, error) {
	val, exists := getField(data, key)
	if !exists {
		return defaultVal, nil
	}
	str, ok := val.(string)
	if !ok {
		return "", NewTypeError(key, "string", fmt.Sprintf("%T", val))
	}
	return str, nil
}

func convertToInt(key string, val interface{}) (int, error) {
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

func GetRequiredInt(data fieldData, key string) (int, error) {
	val, exists := getField(data, key)
	if !exists {
		return 0, NewMissingFieldError(key)
	}
	return convertToInt(key, val)
}

func GetOptionalInt(data fieldData, key string, defaultVal int) (int, error) {
	val, exists := getField(data, key)
	if !exists {
		return defaultVal, nil
	}
	return convertToInt(key, val)
}

func convertToFloat64(key string, val interface{}) (float64, error) {
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

func GetRequiredFloat64(data fieldData, key string) (float64, error) {
	val, exists := getField(data, key)
	if !exists {
		return 0, NewMissingFieldError(key)
	}
	return convertToFloat64(key, val)
}

func GetOptionalFloat64(data fieldData, key string, defaultVal float64) (float64, error) {
	val, exists := getField(data, key)
	if !exists {
		return defaultVal, nil
	}
	return convertToFloat64(key, val)
}

func GetRequiredBool(data fieldData, key string) (bool, error) {
	val, exists := getField(data, key)
	if !exists {
		return false, NewMissingFieldError(key)
	}
	b, ok := val.(bool)
	if !ok {
		return false, NewTypeError(key, "bool", fmt.Sprintf("%T", val))
	}
	return b, nil
}

func GetOptionalBool(data fieldData, key string, defaultVal bool) (bool, error) {
	val, exists := getField(data, key)
	if !exists {
		return defaultVal, nil
	}
	b, ok := val.(bool)
	if !ok {
		return false, NewTypeError(key, "bool", fmt.Sprintf("%T", val))
	}
	return b, nil
}

func parseTime(key string, val interface{}) (time.Time, error) {
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

func GetRequiredTime(data fieldData, key string) (time.Time, error) {
	val, exists := getField(data, key)
	if !exists {
		return time.Time{}, NewMissingFieldError(key)
	}
	return parseTime(key, val)
}

func GetOptionalTime(data fieldData, key string, defaultVal time.Time) (time.Time, error) {
	val, exists := getField(data, key)
	if !exists {
		return defaultVal, nil
	}
	return parseTime(key, val)
}

func GetRequiredMap(data fieldData, key string) (map[string]interface{}, error) {
	val, exists := getField(data, key)
	if !exists {
		return nil, NewMissingFieldError(key)
	}
	m, ok := val.(map[string]interface{})
	if !ok {
		return nil, NewTypeError(key, "map[string]interface{}", fmt.Sprintf("%T", val))
	}
	return m, nil
}

func GetOptionalMap(data fieldData, key string) (map[string]interface{}, error) {
	val, exists := getField(data, key)
	if !exists {
		return nil, nil
	}
	m, ok := val.(map[string]interface{})
	if !ok {
		return nil, NewTypeError(key, "map[string]interface{}", fmt.Sprintf("%T", val))
	}
	return m, nil
}

func convertToMapSlice(key string, val interface{}) ([]map[string]interface{}, error) {
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

func GetRequiredMapSlice(data fieldData, key string) ([]map[string]interface{}, error) {
	val, exists := getField(data, key)
	if !exists {
		return nil, NewMissingFieldError(key)
	}
	return convertToMapSlice(key, val)
}

func GetOptionalMapSlice(data fieldData, key string) ([]map[string]interface{}, error) {
	val, exists := getField(data, key)
	if !exists {
		return nil, nil
	}
	return convertToMapSlice(key, val)
}

func convertToStringSlice(key string, val interface{}) ([]string, error) {
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

func GetRequiredStringSlice(data fieldData, key string) ([]string, error) {
	val, exists := getField(data, key)
	if !exists {
		return nil, NewMissingFieldError(key)
	}
	return convertToStringSlice(key, val)
}

func GetOptionalStringSlice(data fieldData, key string) ([]string, error) {
	val, exists := getField(data, key)
	if !exists {
		return nil, nil
	}
	return convertToStringSlice(key, val)
}
