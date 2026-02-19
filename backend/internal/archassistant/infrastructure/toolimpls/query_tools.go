package toolimpls

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"easi/backend/internal/archassistant/application/tools"
	"easi/backend/internal/archassistant/infrastructure/agenthttp"
)

const (
	defaultLimit = 20
	maxLimit     = 50
)

type query struct {
	client *agenthttp.Client
	ctx    context.Context
}

func (q query) fetch(path string) (*agenthttp.Response, *tools.ToolResult) {
	resp, err := q.client.Get(q.ctx, path)
	if err != nil {
		return nil, &tools.ToolResult{Content: "Failed to reach API: " + err.Error(), IsError: true}
	}
	if !resp.IsSuccess() {
		return nil, &tools.ToolResult{Content: "Failed: " + resp.ErrorMessage(), IsError: true}
	}
	return resp, nil
}

func (q query) fetchCollection(path string) ([]map[string]interface{}, *tools.ToolResult) {
	resp, errResult := q.fetch(path)
	if errResult != nil {
		return nil, errResult
	}
	return parseDataArray(resp.Body), nil
}

func (q query) fetchResource(path string) (map[string]interface{}, *tools.ToolResult) {
	resp, errResult := q.fetch(path)
	if errResult != nil {
		return nil, errResult
	}
	return parseSingleResource(resp.Body), nil
}

func (q query) getByID(args map[string]interface{}, basePath string, formatter func(map[string]interface{}) string) tools.ToolResult {
	id, errResult := requireUUID(args, "id")
	if errResult != nil {
		return *errResult
	}
	resource, errResult := q.fetchResource(basePath + "/" + id)
	if errResult != nil {
		return *errResult
	}
	return tools.ToolResult{Content: formatter(resource)}
}

func clampLimit(args map[string]interface{}) int {
	limit := defaultLimit
	if v, ok := args["limit"].(float64); ok {
		limit = int(v)
	}
	if limit < 1 {
		limit = 1
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	return limit
}

func capFilter(val string) string {
	if len(val) > maxStringLen {
		return val[:maxStringLen]
	}
	return val
}

func buildFilterPath(basePath string, args map[string]interface{}) string {
	params := url.Values{}
	if name, ok := args["name"].(string); ok && name != "" {
		params.Set("name", capFilter(name))
	}
	params.Set("limit", fmt.Sprintf("%d", clampLimit(args)))
	return basePath + "?" + params.Encode()
}

func parseDataArray(body []byte) []map[string]interface{} {
	var wrapper struct {
		Data []map[string]interface{} `json:"data"`
	}
	json.Unmarshal(body, &wrapper)
	return wrapper.Data
}

func parseSingleResource(body []byte) map[string]interface{} {
	var resource map[string]interface{}
	json.Unmarshal(body, &resource)
	return resource
}

func str(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		return fmt.Sprintf("%v", v)
	}
	return ""
}

func num(m map[string]interface{}, key string) int {
	if v, ok := m[key].(float64); ok {
		return int(v)
	}
	return 0
}

func ensureArgs(args map[string]interface{}) map[string]interface{} {
	if args == nil {
		return make(map[string]interface{})
	}
	return args
}

func writeOptionalField(b *strings.Builder, r map[string]interface{}, key, label string) {
	if val := str(r, key); val != "" {
		fmt.Fprintf(b, "%s: %s\n", label, val)
	}
}

func formatListItem(b *strings.Builder, i int, item map[string]interface{}, suffix string) {
	fmt.Fprintf(b, "%d. %s (id: %s)%s", i+1, str(item, "name"), str(item, "id"), suffix)
	if desc := str(item, "description"); desc != "" {
		fmt.Fprintf(b, " - %s", desc)
	}
	b.WriteString("\n")
}

func asMapSlice(raw interface{}) []map[string]interface{} {
	items, ok := raw.([]interface{})
	if !ok {
		return nil
	}
	result := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		if m, ok := item.(map[string]interface{}); ok {
			result = append(result, m)
		}
	}
	return result
}

type listApplicationsTool struct{ client *agenthttp.Client }

func (t *listApplicationsTool) Execute(ctx context.Context, args map[string]interface{}) tools.ToolResult {
	args = ensureArgs(args)
	items, errResult := query{t.client, ctx}.fetchCollection(buildFilterPath("/components", args))
	if errResult != nil {
		return *errResult
	}
	if len(items) == 0 {
		return tools.ToolResult{Content: "No applications found."}
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Found %d applications:\n", len(items))
	for i, item := range items {
		formatListItem(&b, i, item, "")
	}
	return tools.ToolResult{Content: b.String()}
}

type getApplicationDetailsTool struct{ client *agenthttp.Client }

func (t *getApplicationDetailsTool) Execute(ctx context.Context, args map[string]interface{}) tools.ToolResult {
	return query{t.client, ctx}.getByID(args, "/components", formatApplicationDetails)
}

func formatApplicationDetails(r map[string]interface{}) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Application: %s\n", str(r, "name"))
	fmt.Fprintf(&b, "ID: %s\n", str(r, "id"))
	writeOptionalField(&b, r, "description", "Description")
	for _, em := range asMapSlice(r["experts"]) {
		fmt.Fprintf(&b, "  - %s (%s)\n", str(em, "name"), str(em, "role"))
	}
	return b.String()
}

type listApplicationRelationsTool struct{ client *agenthttp.Client }

func (t *listApplicationRelationsTool) Execute(ctx context.Context, args map[string]interface{}) tools.ToolResult {
	id, errResult := requireUUID(args, "id")
	if errResult != nil {
		return *errResult
	}
	q := query{t.client, ctx}
	outgoing, errResult := q.fetchCollection("/relations/from/" + id)
	if errResult != nil {
		return *errResult
	}
	incoming, errResult := q.fetchCollection("/relations/to/" + id)
	if errResult != nil {
		return *errResult
	}
	if len(outgoing) == 0 && len(incoming) == 0 {
		return tools.ToolResult{Content: "No relations found for this application."}
	}
	var b strings.Builder
	writeRelationGroup(&b, "Outgoing", outgoing)
	writeRelationGroup(&b, "Incoming", incoming)
	return tools.ToolResult{Content: b.String()}
}

func writeRelationGroup(b *strings.Builder, direction string, relations []map[string]interface{}) {
	if len(relations) == 0 {
		return
	}
	fmt.Fprintf(b, "%s relations (%d):\n", direction, len(relations))
	for _, rel := range relations {
		label := str(rel, "name")
		if label == "" {
			label = str(rel, "relationType")
		}
		fmt.Fprintf(b, "  - %s: %s -> %s (type: %s, id: %s)\n",
			label, str(rel, "sourceComponentId"), str(rel, "targetComponentId"),
			str(rel, "relationType"), str(rel, "id"))
	}
}

type listCapabilitiesTool struct{ client *agenthttp.Client }

func (t *listCapabilitiesTool) Execute(ctx context.Context, args map[string]interface{}) tools.ToolResult {
	args = ensureArgs(args)
	items, errResult := query{t.client, ctx}.fetchCollection(buildFilterPath("/capabilities", args))
	if errResult != nil {
		return *errResult
	}
	if len(items) == 0 {
		return tools.ToolResult{Content: "No capabilities found."}
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Found %d capabilities:\n", len(items))
	for i, item := range items {
		suffix := fmt.Sprintf(", level: %s", str(item, "level"))
		if parent := str(item, "parentId"); parent != "" {
			suffix += fmt.Sprintf(" [parent: %s]", parent)
		}
		formatListItem(&b, i, item, suffix)
	}
	return tools.ToolResult{Content: b.String()}
}

type getCapabilityDetailsTool struct{ client *agenthttp.Client }

func (t *getCapabilityDetailsTool) Execute(ctx context.Context, args map[string]interface{}) tools.ToolResult {
	return query{t.client, ctx}.getByID(args, "/capabilities", formatCapabilityDetails)
}

func formatCapabilityDetails(r map[string]interface{}) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Capability: %s\n", str(r, "name"))
	fmt.Fprintf(&b, "ID: %s\n", str(r, "id"))
	fmt.Fprintf(&b, "Level: %s\n", str(r, "level"))
	writeOptionalField(&b, r, "description", "Description")
	writeOptionalField(&b, r, "parentId", "Parent ID")
	writeOptionalField(&b, r, "status", "Status")
	return b.String()
}

type listBusinessDomainsTool struct{ client *agenthttp.Client }

func (t *listBusinessDomainsTool) Execute(ctx context.Context, _ map[string]interface{}) tools.ToolResult {
	items, errResult := query{t.client, ctx}.fetchCollection("/business-domains")
	if errResult != nil {
		return *errResult
	}
	if len(items) == 0 {
		return tools.ToolResult{Content: "No business domains found."}
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Found %d business domains:\n", len(items))
	for i, item := range items {
		suffix := fmt.Sprintf(", capabilities: %d", num(item, "capabilityCount"))
		formatListItem(&b, i, item, suffix)
	}
	return tools.ToolResult{Content: b.String()}
}

type getBusinessDomainDetailsTool struct{ client *agenthttp.Client }

func (t *getBusinessDomainDetailsTool) Execute(ctx context.Context, args map[string]interface{}) tools.ToolResult {
	return query{t.client, ctx}.getByID(args, "/business-domains", formatBusinessDomainDetails)
}

func formatBusinessDomainDetails(r map[string]interface{}) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Business Domain: %s\n", str(r, "name"))
	fmt.Fprintf(&b, "ID: %s\n", str(r, "id"))
	writeOptionalField(&b, r, "description", "Description")
	fmt.Fprintf(&b, "Capabilities: %d\n", num(r, "capabilityCount"))
	writeOptionalField(&b, r, "domainArchitectId", "Domain Architect ID")
	return b.String()
}

type listValueStreamsTool struct{ client *agenthttp.Client }

func (t *listValueStreamsTool) Execute(ctx context.Context, _ map[string]interface{}) tools.ToolResult {
	items, errResult := query{t.client, ctx}.fetchCollection("/value-streams")
	if errResult != nil {
		return *errResult
	}
	if len(items) == 0 {
		return tools.ToolResult{Content: "No value streams found."}
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Found %d value streams:\n", len(items))
	for i, item := range items {
		suffix := fmt.Sprintf(", stages: %d", num(item, "stageCount"))
		formatListItem(&b, i, item, suffix)
	}
	return tools.ToolResult{Content: b.String()}
}

type getValueStreamDetailsTool struct{ client *agenthttp.Client }

func (t *getValueStreamDetailsTool) Execute(ctx context.Context, args map[string]interface{}) tools.ToolResult {
	return query{t.client, ctx}.getByID(args, "/value-streams", formatValueStreamDetails)
}

func formatValueStreamDetails(r map[string]interface{}) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Value Stream: %s\n", str(r, "name"))
	fmt.Fprintf(&b, "ID: %s\n", str(r, "id"))
	writeOptionalField(&b, r, "description", "Description")
	fmt.Fprintf(&b, "Stages: %d\n", num(r, "stageCount"))
	writeStageList(&b, asMapSlice(r["stages"]))
	writeCapabilityMappings(&b, asMapSlice(r["stageCapabilities"]))
	return b.String()
}

func writeStageList(b *strings.Builder, stages []map[string]interface{}) {
	if len(stages) == 0 {
		return
	}
	b.WriteString("Stage list:\n")
	for _, sm := range stages {
		fmt.Fprintf(b, "  %d. %s (id: %s)\n", num(sm, "position"), str(sm, "name"), str(sm, "id"))
	}
}

func writeCapabilityMappings(b *strings.Builder, mappings []map[string]interface{}) {
	if len(mappings) == 0 {
		return
	}
	b.WriteString("Capability mappings:\n")
	for _, mm := range mappings {
		fmt.Fprintf(b, "  - Stage %s -> %s (capability: %s)\n",
			str(mm, "stageId"), str(mm, "capabilityName"), str(mm, "capabilityId"))
	}
}

type searchArchitectureTool struct{ client *agenthttp.Client }

func (t *searchArchitectureTool) Execute(ctx context.Context, args map[string]interface{}) tools.ToolResult {
	searchQuery, errResult := requireString(args, "query")
	if errResult != nil {
		return *errResult
	}
	searchQuery = capFilter(searchQuery)
	encoded := url.QueryEscape(searchQuery)
	q := query{t.client, ctx}

	var sections []string
	sections = appendSearchSection(sections, "Applications", q, "/components?name="+encoded+"&limit=10", nil)
	sections = appendSearchSection(sections, "Capabilities", q, "/capabilities?name="+encoded+"&limit=10", nil)
	sections = appendSearchSection(sections, "Business Domains", q, "/business-domains", nameFilter(searchQuery))

	if len(sections) == 0 {
		return tools.ToolResult{Content: fmt.Sprintf("No results found for '%s'.", searchQuery)}
	}

	header := fmt.Sprintf("Search results for '%s':\n\n", searchQuery)
	return tools.ToolResult{Content: header + strings.Join(sections, "\n")}
}

func appendSearchSection(sections []string, label string, q query, path string, filter func([]map[string]interface{}) []map[string]interface{}) []string {
	items, _ := q.fetchCollection(path)
	if filter != nil {
		items = filter(items)
	}
	if len(items) == 0 {
		return sections
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%s (%d):\n", label, len(items))
	for _, item := range items {
		fmt.Fprintf(&b, "  - %s (id: %s)\n", str(item, "name"), str(item, "id"))
	}
	return append(sections, b.String())
}

func nameFilter(searchQuery string) func([]map[string]interface{}) []map[string]interface{} {
	lower := strings.ToLower(searchQuery)
	return func(items []map[string]interface{}) []map[string]interface{} {
		var matched []map[string]interface{}
		for _, item := range items {
			if strings.Contains(strings.ToLower(str(item, "name")), lower) {
				matched = append(matched, item)
			}
		}
		return matched
	}
}

type getPortfolioSummaryTool struct{ client *agenthttp.Client }

func (t *getPortfolioSummaryTool) Execute(ctx context.Context, _ map[string]interface{}) tools.ToolResult {
	q := query{t.client, ctx}
	apps, _ := q.fetchCollection("/components?limit=50")
	caps, _ := q.fetchCollection("/capabilities?limit=50")
	domains, _ := q.fetchCollection("/business-domains")
	streams, _ := q.fetchCollection("/value-streams")
	relations, _ := q.fetchCollection("/relations")

	var b strings.Builder
	b.WriteString("Architecture Portfolio Summary:\n")
	fmt.Fprintf(&b, "  Applications: %d\n", len(apps))
	fmt.Fprintf(&b, "  Capabilities: %d\n", len(caps))
	fmt.Fprintf(&b, "  Business Domains: %d\n", len(domains))
	fmt.Fprintf(&b, "  Value Streams: %d\n", len(streams))
	fmt.Fprintf(&b, "  Relations: %d\n", len(relations))
	return tools.ToolResult{Content: b.String()}
}

func RegisterQueryTools(registry *tools.Registry, client *agenthttp.Client) {
	registerComponentQueryTools(registry, client)
	registerCapabilityQueryTools(registry, client)
	registerDomainQueryTools(registry, client)
	registerValueStreamQueryTools(registry, client)
	registerCrossEntityQueryTools(registry, client)
}

func registerComponentQueryTools(registry *tools.Registry, client *agenthttp.Client) {
	registry.Register(tools.ToolDefinition{
		Name: "list_applications", Description: "List applications in the architecture portfolio. Optionally filter by name.",
		Permission: "components:read", Access: tools.AccessRead,
		Parameters: []tools.ParameterDef{
			{Name: "name", Type: "string", Description: "Filter by application name (partial match)"},
			{Name: "limit", Type: "integer", Description: "Max results (1-50, default 20)"},
		},
	}, &listApplicationsTool{client: client})

	registry.Register(tools.ToolDefinition{
		Name: "get_application_details", Description: "Get full details of an application by ID",
		Permission: "components:read", Access: tools.AccessRead,
		Parameters: []tools.ParameterDef{
			{Name: "id", Type: "string", Description: "Application ID (UUID)", Required: true},
		},
	}, &getApplicationDetailsTool{client: client})

	registry.Register(tools.ToolDefinition{
		Name: "list_application_relations", Description: "List all relations (incoming and outgoing) for an application",
		Permission: "components:read", Access: tools.AccessRead,
		Parameters: []tools.ParameterDef{
			{Name: "id", Type: "string", Description: "Application ID (UUID)", Required: true},
		},
	}, &listApplicationRelationsTool{client: client})
}

func registerCapabilityQueryTools(registry *tools.Registry, client *agenthttp.Client) {
	registry.Register(tools.ToolDefinition{
		Name: "list_capabilities", Description: "List business capabilities. Optionally filter by name.",
		Permission: "capabilities:read", Access: tools.AccessRead,
		Parameters: []tools.ParameterDef{
			{Name: "name", Type: "string", Description: "Filter by capability name (partial match)"},
			{Name: "limit", Type: "integer", Description: "Max results (1-50, default 20)"},
		},
	}, &listCapabilitiesTool{client: client})

	registry.Register(tools.ToolDefinition{
		Name: "get_capability_details", Description: "Get full details of a capability including realizations",
		Permission: "capabilities:read", Access: tools.AccessRead,
		Parameters: []tools.ParameterDef{
			{Name: "id", Type: "string", Description: "Capability ID (UUID)", Required: true},
		},
	}, &getCapabilityDetailsTool{client: client})
}

func registerDomainQueryTools(registry *tools.Registry, client *agenthttp.Client) {
	registry.Register(tools.ToolDefinition{
		Name: "list_business_domains", Description: "List all business domains",
		Permission: "domains:read", Access: tools.AccessRead,
	}, &listBusinessDomainsTool{client: client})

	registry.Register(tools.ToolDefinition{
		Name: "get_business_domain_details", Description: "Get details of a business domain with its capabilities",
		Permission: "domains:read", Access: tools.AccessRead,
		Parameters: []tools.ParameterDef{
			{Name: "id", Type: "string", Description: "Business domain ID (UUID)", Required: true},
		},
	}, &getBusinessDomainDetailsTool{client: client})
}

func registerValueStreamQueryTools(registry *tools.Registry, client *agenthttp.Client) {
	registry.Register(tools.ToolDefinition{
		Name: "list_value_streams", Description: "List all value streams",
		Permission: "valuestreams:read", Access: tools.AccessRead,
	}, &listValueStreamsTool{client: client})

	registry.Register(tools.ToolDefinition{
		Name: "get_value_stream_details", Description: "Get value stream details including stages and mapped capabilities",
		Permission: "valuestreams:read", Access: tools.AccessRead,
		Parameters: []tools.ParameterDef{
			{Name: "id", Type: "string", Description: "Value stream ID (UUID)", Required: true},
		},
	}, &getValueStreamDetailsTool{client: client})
}

func registerCrossEntityQueryTools(registry *tools.Registry, client *agenthttp.Client) {
	registry.Register(tools.ToolDefinition{
		Name: "search_architecture", Description: "Search across applications, capabilities, and business domains by name",
		Permission: "components:read", Access: tools.AccessRead,
		Parameters: []tools.ParameterDef{
			{Name: "query", Type: "string", Description: "Search query (name to search for)", Required: true},
		},
	}, &searchArchitectureTool{client: client})

	registry.Register(tools.ToolDefinition{
		Name: "get_portfolio_summary", Description: "Get aggregate statistics across the architecture portfolio",
		Permission: "components:read", Access: tools.AccessRead,
	}, &getPortfolioSummaryTool{client: client})
}
