package repository

import (
	"fmt"
	"time"
)

type fieldData = map[string]interface{}
type converter[T any] func(key string, val interface{}) (T, error)

func getField(data fieldData, key string) (interface{}, bool) {
	val, exists := data[key]
	return val, exists && val != nil
}

func extractRequired[T any](data fieldData, key string, convert converter[T]) (T, error) {
	var zero T
	val, exists := getField(data, key)
	if !exists {
		return zero, NewMissingFieldError(key)
	}
	return convert(key, val)
}

func extractOptional[T any](data fieldData, key string, defaultVal T, convert converter[T]) (T, error) {
	val, exists := getField(data, key)
	if !exists {
		return defaultVal, nil
	}
	return convert(key, val)
}

func newSimpleConverter[T any](typeName string) converter[T] {
	return func(key string, val interface{}) (T, error) {
		var zero T
		v, ok := val.(T)
		if !ok {
			return zero, NewTypeError(key, typeName, fmt.Sprintf("%T", val))
		}
		return v, nil
	}
}

var convertString = newSimpleConverter[string]("string")

func extractNumeric(val interface{}) (intVal int, floatVal float64, ok bool) {
	switch v := val.(type) {
	case int:
		return v, float64(v), true
	case int64:
		return int(v), float64(v), true
	case float64:
		return int(v), v, true
	default:
		return 0, 0, false
	}
}

func convertInt(key string, val interface{}) (int, error) {
	intVal, _, ok := extractNumeric(val)
	if !ok {
		return 0, NewTypeError(key, "int", fmt.Sprintf("%T", val))
	}
	return intVal, nil
}

func convertFloat64(key string, val interface{}) (float64, error) {
	_, floatVal, ok := extractNumeric(val)
	if !ok {
		return 0, NewTypeError(key, "float64", fmt.Sprintf("%T", val))
	}
	return floatVal, nil
}

var convertBool = newSimpleConverter[bool]("bool")

func convertTime(key string, val interface{}) (time.Time, error) {
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

var convertMap = newSimpleConverter[map[string]interface{}]("map[string]interface{}")

type elementConverter[T any] func(interface{}) (T, bool)

func convertSlice[T any](key string, val interface{}, convert elementConverter[T], typeName string) ([]T, error) {
	raw, ok := val.([]interface{})
	if !ok {
		return nil, NewTypeError(key, "[]interface{}", fmt.Sprintf("%T", val))
	}
	result := make([]T, 0, len(raw))
	for _, item := range raw {
		elem, ok := convert(item)
		if !ok {
			return nil, NewTypeError(key, typeName, fmt.Sprintf("slice containing %T", item))
		}
		result = append(result, elem)
	}
	return result, nil
}

func convertMapSlice(key string, val interface{}) ([]map[string]interface{}, error) {
	return convertSlice(key, val, func(v interface{}) (map[string]interface{}, bool) {
		m, ok := v.(map[string]interface{})
		return m, ok
	}, "[]map[string]interface{}")
}

func convertStringSlice(key string, val interface{}) ([]string, error) {
	return convertSlice(key, val, func(v interface{}) (string, bool) {
		s, ok := v.(string)
		return s, ok
	}, "[]string")
}

func GetRequiredString(data fieldData, key string) (string, error) {
	return extractRequired(data, key, convertString)
}

func GetOptionalString(data fieldData, key string, defaultVal string) (string, error) {
	return extractOptional(data, key, defaultVal, convertString)
}

func GetRequiredInt(data fieldData, key string) (int, error) {
	return extractRequired(data, key, convertInt)
}

func GetOptionalInt(data fieldData, key string, defaultVal int) (int, error) {
	return extractOptional(data, key, defaultVal, convertInt)
}

func GetRequiredFloat64(data fieldData, key string) (float64, error) {
	return extractRequired(data, key, convertFloat64)
}

func GetOptionalFloat64(data fieldData, key string, defaultVal float64) (float64, error) {
	return extractOptional(data, key, defaultVal, convertFloat64)
}

func GetRequiredBool(data fieldData, key string) (bool, error) {
	return extractRequired(data, key, convertBool)
}

func GetOptionalBool(data fieldData, key string, defaultVal bool) (bool, error) {
	return extractOptional(data, key, defaultVal, convertBool)
}

func GetRequiredTime(data fieldData, key string) (time.Time, error) {
	return extractRequired(data, key, convertTime)
}

func GetOptionalTime(data fieldData, key string, defaultVal time.Time) (time.Time, error) {
	return extractOptional(data, key, defaultVal, convertTime)
}

func GetRequiredMap(data fieldData, key string) (map[string]interface{}, error) {
	return extractRequired(data, key, convertMap)
}

func GetOptionalMap(data fieldData, key string) (map[string]interface{}, error) {
	return extractOptional(data, key, nil, convertMap)
}

func GetRequiredMapSlice(data fieldData, key string) ([]map[string]interface{}, error) {
	return extractRequired(data, key, convertMapSlice)
}

func GetOptionalMapSlice(data fieldData, key string) ([]map[string]interface{}, error) {
	return extractOptional(data, key, nil, convertMapSlice)
}

func GetRequiredStringSlice(data fieldData, key string) ([]string, error) {
	return extractRequired(data, key, convertStringSlice)
}

func GetOptionalStringSlice(data fieldData, key string) ([]string, error) {
	return extractOptional(data, key, nil, convertStringSlice)
}
