package toolimpls

import (
	"context"
	"encoding/json"
	"fmt"

	"easi/backend/internal/archassistant/application/tools"
	"easi/backend/internal/archassistant/infrastructure/agenthttp"

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

func requireTwoUUIDs(args map[string]interface{}, key1, key2 string) (string, string, *tools.ToolResult) {
	id1, errResult := requireUUID(args, key1)
	if errResult != nil {
		return "", "", errResult
	}
	id2, errResult := requireUUID(args, key2)
	if errResult != nil {
		return "", "", errResult
	}
	return id1, id2, nil
}

func extractField(body []byte, field string) string {
	var parsed map[string]interface{}
	if err := json.Unmarshal(body, &parsed); err == nil {
		if val, ok := parsed[field]; ok {
			return fmt.Sprintf("%v", val)
		}
	}
	return ""
}

func addOptionalStrings(args, body map[string]interface{}, keys ...string) *tools.ToolResult {
	for _, key := range keys {
		val, ok := args[key].(string)
		if !ok {
			continue
		}
		if len(val) > maxStringLen {
			return toolErr(key + " must be at most 200 characters")
		}
		body[key] = val
	}
	return nil
}

type mutation struct {
	client *agenthttp.Client
	ctx    context.Context
}

func (m mutation) post(path string, body map[string]interface{}) (*agenthttp.Response, *tools.ToolResult) {
	return m.send(m.client.Post(m.ctx, path, body))
}

func (m mutation) put(path string, body map[string]interface{}) (*agenthttp.Response, *tools.ToolResult) {
	return m.send(m.client.Put(m.ctx, path, body))
}

func (m mutation) del(path string) *tools.ToolResult {
	_, errResult := m.send(m.client.Delete(m.ctx, path))
	return errResult
}

func (m mutation) send(resp *agenthttp.Response, err error) (*agenthttp.Response, *tools.ToolResult) {
	if err != nil {
		return nil, toolErr("Failed to reach API: " + err.Error())
	}
	if !resp.IsSuccess() {
		return nil, toolErr("Failed: " + resp.ErrorMessage())
	}
	return resp, nil
}

type resourceSpec struct {
	entityType   string
	basePath     string
	optionalKeys []string
}

func (s resourceSpec) create(m mutation, args map[string]interface{}) tools.ToolResult {
	name, errResult := requireString(args, "name")
	if errResult != nil {
		return *errResult
	}
	body := map[string]interface{}{"name": name}
	resp, errResult := s.postWithOptionals(m, args, body)
	if errResult != nil {
		return *errResult
	}
	return tools.ToolResult{Content: fmt.Sprintf("Created %s '%s' (id: %s)", s.entityType, name, extractField(resp.Body, "id"))}
}

func (s resourceSpec) postWithOptionals(m mutation, args, body map[string]interface{}) (*agenthttp.Response, *tools.ToolResult) {
	if errResult := addOptionalStrings(args, body, s.optionalKeys...); errResult != nil {
		return nil, errResult
	}
	return m.post(s.basePath, body)
}

func (s resourceSpec) update(m mutation, args map[string]interface{}) tools.ToolResult {
	id, errResult := requireUUID(args, "id")
	if errResult != nil {
		return *errResult
	}
	body := make(map[string]interface{})
	if errResult := addOptionalStrings(args, body, "name", "description"); errResult != nil {
		return *errResult
	}
	resp, errResult := m.put(s.basePath+"/"+id, body)
	if errResult != nil {
		return *errResult
	}
	return tools.ToolResult{Content: fmt.Sprintf("Updated %s '%s' (id: %s)", s.entityType, extractField(resp.Body, "name"), id)}
}

func (s resourceSpec) deleteOne(m mutation, args map[string]interface{}) tools.ToolResult {
	id, errResult := requireUUID(args, "id")
	if errResult != nil {
		return *errResult
	}
	if errResult := m.del(s.basePath + "/" + id); errResult != nil {
		return *errResult
	}
	return tools.ToolResult{Content: fmt.Sprintf("Deleted %s (id: %s)", s.entityType, id)}
}

var (
	appSpec    = resourceSpec{"application", "/components", []string{"description"}}
	capSpec    = resourceSpec{"capability", "/capabilities", nil}
	domainSpec = resourceSpec{"business domain", "/business-domains", []string{"description"}}
)

type createApplicationTool struct{ client *agenthttp.Client }

func (t *createApplicationTool) Execute(ctx context.Context, args map[string]interface{}) tools.ToolResult {
	return appSpec.create(mutation{t.client, ctx}, args)
}

type updateApplicationTool struct{ client *agenthttp.Client }

func (t *updateApplicationTool) Execute(ctx context.Context, args map[string]interface{}) tools.ToolResult {
	return appSpec.update(mutation{t.client, ctx}, args)
}

type deleteApplicationTool struct{ client *agenthttp.Client }

func (t *deleteApplicationTool) Execute(ctx context.Context, args map[string]interface{}) tools.ToolResult {
	return appSpec.deleteOne(mutation{t.client, ctx}, args)
}

type createCapabilityTool struct{ client *agenthttp.Client }

func (t *createCapabilityTool) Execute(ctx context.Context, args map[string]interface{}) tools.ToolResult {
	name, errResult := requireString(args, "name")
	if errResult != nil {
		return *errResult
	}
	level, errResult := requireString(args, "level")
	if errResult != nil {
		return *errResult
	}
	body := map[string]interface{}{"name": name, "level": level}
	if errResult := addOptionalStrings(args, body, "parentId", "description"); errResult != nil {
		return *errResult
	}
	m := mutation{t.client, ctx}
	resp, errResult := m.post("/capabilities", body)
	if errResult != nil {
		return *errResult
	}
	return tools.ToolResult{Content: fmt.Sprintf("Created capability '%s' (id: %s, level: %s)", name, extractField(resp.Body, "id"), level)}
}

type updateCapabilityTool struct{ client *agenthttp.Client }

func (t *updateCapabilityTool) Execute(ctx context.Context, args map[string]interface{}) tools.ToolResult {
	return capSpec.update(mutation{t.client, ctx}, args)
}

type deleteCapabilityTool struct{ client *agenthttp.Client }

func (t *deleteCapabilityTool) Execute(ctx context.Context, args map[string]interface{}) tools.ToolResult {
	return capSpec.deleteOne(mutation{t.client, ctx}, args)
}

type createBusinessDomainTool struct{ client *agenthttp.Client }

func (t *createBusinessDomainTool) Execute(ctx context.Context, args map[string]interface{}) tools.ToolResult {
	return domainSpec.create(mutation{t.client, ctx}, args)
}

type updateBusinessDomainTool struct{ client *agenthttp.Client }

func (t *updateBusinessDomainTool) Execute(ctx context.Context, args map[string]interface{}) tools.ToolResult {
	return domainSpec.update(mutation{t.client, ctx}, args)
}

type createApplicationRelationTool struct{ client *agenthttp.Client }

func (t *createApplicationRelationTool) Execute(ctx context.Context, args map[string]interface{}) tools.ToolResult {
	m := mutation{t.client, ctx}
	sourceID, errResult := requireUUID(args, "sourceId")
	if errResult != nil {
		return *errResult
	}
	targetID, errResult := requireUUID(args, "targetId")
	if errResult != nil {
		return *errResult
	}
	relType, errResult := requireString(args, "type")
	if errResult != nil {
		return *errResult
	}
	body := map[string]interface{}{"targetId": targetID, "type": relType}
	if errResult := addOptionalStrings(args, body, "description"); errResult != nil {
		return *errResult
	}
	resp, errResult := m.post("/components/"+sourceID+"/relations", body)
	if errResult != nil {
		return *errResult
	}
	return tools.ToolResult{Content: fmt.Sprintf("Created relation from %s to %s (type: %s, id: %s)", sourceID, targetID, relType, extractField(resp.Body, "id"))}
}

type deleteApplicationRelationTool struct{ client *agenthttp.Client }

func (t *deleteApplicationRelationTool) Execute(ctx context.Context, args map[string]interface{}) tools.ToolResult {
	componentID, relationID, errResult := requireTwoUUIDs(args, "componentId", "relationId")
	if errResult != nil {
		return *errResult
	}
	if errResult := (mutation{t.client, ctx}).del(fmt.Sprintf("/components/%s/relations/%s", componentID, relationID)); errResult != nil {
		return *errResult
	}
	return tools.ToolResult{Content: fmt.Sprintf("Deleted relation (id: %s)", relationID)}
}

type realizeCapabilityTool struct{ client *agenthttp.Client }

func (t *realizeCapabilityTool) Execute(ctx context.Context, args map[string]interface{}) tools.ToolResult {
	capabilityID, applicationID, errResult := requireTwoUUIDs(args, "capabilityId", "applicationId")
	if errResult != nil {
		return *errResult
	}
	body := map[string]interface{}{"applicationId": applicationID}
	resp, errResult := (mutation{t.client, ctx}).post("/capabilities/"+capabilityID+"/realizations", body)
	if errResult != nil {
		return *errResult
	}
	return tools.ToolResult{Content: fmt.Sprintf("Linked application %s to capability %s (id: %s)", applicationID, capabilityID, extractField(resp.Body, "id"))}
}

type unrealizeCapabilityTool struct{ client *agenthttp.Client }

func (t *unrealizeCapabilityTool) Execute(ctx context.Context, args map[string]interface{}) tools.ToolResult {
	capabilityID, realizationID, errResult := requireTwoUUIDs(args, "capabilityId", "realizationId")
	if errResult != nil {
		return *errResult
	}
	if errResult := (mutation{t.client, ctx}).del("/capabilities/" + capabilityID + "/realizations/" + realizationID); errResult != nil {
		return *errResult
	}
	return tools.ToolResult{Content: fmt.Sprintf("Unlinked realization (id: %s) from capability %s", realizationID, capabilityID)}
}

type assignCapabilityToDomainTool struct{ client *agenthttp.Client }

func (t *assignCapabilityToDomainTool) Execute(ctx context.Context, args map[string]interface{}) tools.ToolResult {
	domainID, capabilityID, errResult := requireTwoUUIDs(args, "domainId", "capabilityId")
	if errResult != nil {
		return *errResult
	}
	body := map[string]interface{}{"capabilityId": capabilityID}
	_, errResult = (mutation{t.client, ctx}).post("/business-domains/"+domainID+"/capabilities", body)
	if errResult != nil {
		return *errResult
	}
	return tools.ToolResult{Content: fmt.Sprintf("Assigned capability %s to business domain %s", capabilityID, domainID)}
}

type removeCapabilityFromDomainTool struct{ client *agenthttp.Client }

func (t *removeCapabilityFromDomainTool) Execute(ctx context.Context, args map[string]interface{}) tools.ToolResult {
	domainID, capabilityID, errResult := requireTwoUUIDs(args, "domainId", "capabilityId")
	if errResult != nil {
		return *errResult
	}
	if errResult := (mutation{t.client, ctx}).del("/business-domains/" + domainID + "/capabilities/" + capabilityID); errResult != nil {
		return *errResult
	}
	return tools.ToolResult{Content: fmt.Sprintf("Removed capability %s from business domain %s", capabilityID, domainID)}
}
