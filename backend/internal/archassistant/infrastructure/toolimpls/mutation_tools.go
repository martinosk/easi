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

func requireString(args map[string]interface{}, key string) (string, *tools.ToolResult) {
	val, _ := args[key].(string)
	if val == "" {
		return "", &tools.ToolResult{Content: key + " is required", IsError: true}
	}
	if len(val) > maxStringLen {
		return "", &tools.ToolResult{Content: key + " must be at most 200 characters", IsError: true}
	}
	return val, nil
}

func requireUUID(args map[string]interface{}, key string) (string, *tools.ToolResult) {
	val, _ := args[key].(string)
	if val == "" {
		return "", &tools.ToolResult{Content: key + " is required", IsError: true}
	}
	if _, err := uuid.Parse(val); err != nil {
		return "", &tools.ToolResult{Content: key + " must be a valid UUID", IsError: true}
	}
	return val, nil
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

func apiError(err error) tools.ToolResult {
	return tools.ToolResult{Content: "Failed to reach API: " + err.Error(), IsError: true}
}

func responseError(resp *agenthttp.Response) tools.ToolResult {
	return tools.ToolResult{Content: "Failed: " + resp.ErrorMessage(), IsError: true}
}

func addOptionalStrings(args, body map[string]interface{}, keys ...string) *tools.ToolResult {
	for _, key := range keys {
		val, ok := args[key].(string)
		if !ok {
			continue
		}
		if len(val) > maxStringLen {
			return &tools.ToolResult{Content: key + " must be at most 200 characters", IsError: true}
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
	resp, err := m.client.Post(m.ctx, path, body)
	return m.checkResponse(resp, err)
}

func (m mutation) put(path string, body map[string]interface{}) (*agenthttp.Response, *tools.ToolResult) {
	resp, err := m.client.Put(m.ctx, path, body)
	return m.checkResponse(resp, err)
}

func (m mutation) del(path string) *tools.ToolResult {
	resp, err := m.client.Delete(m.ctx, path)
	_, errResult := m.checkResponse(resp, err)
	return errResult
}

func (m mutation) checkResponse(resp *agenthttp.Response, err error) (*agenthttp.Response, *tools.ToolResult) {
	if err != nil {
		return nil, &tools.ToolResult{Content: "Failed to reach API: " + err.Error(), IsError: true}
	}
	if !resp.IsSuccess() {
		return nil, &tools.ToolResult{Content: "Failed: " + resp.ErrorMessage(), IsError: true}
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
	if errResult := addOptionalStrings(args, body, s.optionalKeys...); errResult != nil {
		return *errResult
	}
	resp, errResult := m.post(s.basePath, body)
	if errResult != nil {
		return *errResult
	}
	return tools.ToolResult{Content: fmt.Sprintf("Created %s '%s' (id: %s)", s.entityType, name, extractField(resp.Body, "id"))}
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
	capSpec    = resourceSpec{"capability", "/capabilities", []string{"domainId", "description"}}
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
	return capSpec.create(mutation{t.client, ctx}, args)
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
