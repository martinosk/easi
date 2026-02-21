package toolimpls

import (
	"easi/backend/internal/archassistant/application/tools"

	"github.com/google/uuid"
)

const maxStringLen = 200

func toolErr(msg string) *tools.ToolResult {
	return &tools.ToolResult{Content: msg, IsError: true}
}

type argValidator func(val, key string) *tools.ToolResult

var validateStringLen argValidator = func(val, key string) *tools.ToolResult {
	if len(val) > maxStringLen {
		return toolErr(key + " must be at most 200 characters")
	}
	return nil
}

var validateUUIDFormat argValidator = func(val, key string) *tools.ToolResult {
	if _, err := uuid.Parse(val); err != nil {
		return toolErr(key + " must be a valid UUID")
	}
	return nil
}

func requireArg(args map[string]interface{}, key string, validate argValidator) (string, *tools.ToolResult) {
	val, _ := args[key].(string)
	if val == "" {
		return "", toolErr(key + " is required")
	}
	if errResult := validate(val, key); errResult != nil {
		return "", errResult
	}
	return val, nil
}

func requireString(args map[string]interface{}, key string) (string, *tools.ToolResult) {
	return requireArg(args, key, validateStringLen)
}

func requireUUID(args map[string]interface{}, key string) (string, *tools.ToolResult) {
	return requireArg(args, key, validateUUIDFormat)
}

func capFilter(val string) string {
	if len(val) > maxStringLen {
		return val[:maxStringLen]
	}
	return val
}
