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

func parseDataArray(body []byte) []map[string]interface{} {
	var wrapper struct {
		Data []map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(body, &wrapper); err != nil {
		return nil
	}
	return wrapper.Data
}

func str(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		return fmt.Sprintf("%v", v)
	}
	return ""
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
