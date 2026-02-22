package toolimpls

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"easi/backend/internal/archassistant/application/tools"
	"easi/backend/internal/archassistant/infrastructure/agenthttp"
	pl "easi/backend/internal/archassistant/publishedlanguage"

	"github.com/google/uuid"
)

type ParamSpec = pl.ParamSpec

type AgentToolSpec = pl.AgentToolSpec

type GenericAPIToolExecutor struct {
	spec   AgentToolSpec
	client *agenthttp.Client
}

func NewGenericExecutor(spec AgentToolSpec, client *agenthttp.Client) *GenericAPIToolExecutor {
	return &GenericAPIToolExecutor{spec: spec, client: client}
}

func (e *GenericAPIToolExecutor) Execute(ctx context.Context, args map[string]interface{}) tools.ToolResult {
	if args == nil {
		args = make(map[string]interface{})
	}
	if errResult := validateAllParams(e.spec, args); errResult != nil {
		return *errResult
	}
	return e.dispatch(ctx, args)
}

func validateAllParams(spec AgentToolSpec, args map[string]interface{}) *tools.ToolResult {
	for _, group := range [][]ParamSpec{spec.PathParams, spec.QueryParams, spec.BodyParams} {
		for _, p := range group {
			if errResult := validateParam(args, p); errResult != nil {
				return errResult
			}
		}
	}
	return nil
}

func (e *GenericAPIToolExecutor) dispatch(ctx context.Context, args map[string]interface{}) tools.ToolResult {
	path := substitutePath(e.spec.Path, e.spec.PathParams, args)
	path += buildQueryString(e.spec.QueryParams, args)

	resp, err := e.callAPI(ctx, path, args)
	if err != nil {
		return tools.ToolResult{Content: "Failed to reach API: " + err.Error(), IsError: true}
	}
	if !resp.IsSuccess() {
		return tools.ToolResult{Content: "Failed: " + resp.ErrorMessage(), IsError: true}
	}
	return tools.ToolResult{Content: string(resp.Body)}
}

func (e *GenericAPIToolExecutor) callAPI(ctx context.Context, path string, args map[string]interface{}) (*agenthttp.Response, error) {
	if e.spec.Method == "POST" || e.spec.Method == "PUT" {
		return e.client.DoWithBody(ctx, e.spec.Method, path, buildBody(e.spec.BodyParams, args))
	}
	return e.client.Do(ctx, e.spec.Method, path)
}

func substitutePath(path string, params []ParamSpec, args map[string]interface{}) string {
	for _, p := range params {
		val, _ := args[p.Name].(string)
		path = strings.ReplaceAll(path, fmt.Sprintf("{%s}", p.Name), val)
	}
	return path
}

func buildQueryString(params []ParamSpec, args map[string]interface{}) string {
	qv := url.Values{}
	for _, p := range params {
		if val := extractStringValue(args, p); val != "" {
			qv.Set(p.Name, val)
		}
	}
	if len(qv) == 0 {
		return ""
	}
	return "?" + qv.Encode()
}

func buildBody(params []ParamSpec, args map[string]interface{}) map[string]interface{} {
	body := make(map[string]interface{})
	for _, p := range params {
		if val := extractStringValue(args, p); val != "" {
			body[p.Name] = val
		}
	}
	return body
}

func extractStringValue(args map[string]interface{}, p ParamSpec) string {
	switch p.Type {
	case "integer":
		if v, ok := args[p.Name].(float64); ok {
			return fmt.Sprintf("%d", int(v))
		}
	case "boolean":
		if v, ok := args[p.Name].(bool); ok {
			return fmt.Sprintf("%t", v)
		}
	default:
		if v, ok := args[p.Name].(string); ok {
			return v
		}
	}
	return ""
}

func isStringTyped(p ParamSpec) bool {
	return p.Type != "integer" && p.Type != "boolean"
}

func validateParam(args map[string]interface{}, p ParamSpec) *tools.ToolResult {
	val, _ := args[p.Name].(string)
	if val == "" && isStringTyped(p) {
		if p.Required {
			return toolErr(p.Name + " is required")
		}
		return nil
	}
	return validateParamValue(val, p)
}

func validateParamValue(val string, p ParamSpec) *tools.ToolResult {
	switch p.Type {
	case "uuid":
		if _, err := uuid.Parse(val); err != nil {
			return toolErr(p.Name + " must be a valid UUID")
		}
	case "string":
		if len(val) > maxStringLen {
			return toolErr(p.Name + " must be at most 200 characters")
		}
	}
	return nil
}
