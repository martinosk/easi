package orchestrator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testRegisteredNames = []string{
	"list_applications",
	"get_application_details",
	"list_capabilities",
	"search_architecture",
	"get_portfolio_summary",
}

func TestParseTextToolCalls_ToolCallJSON(t *testing.T) {
	content := `I'll look up the applications.
<tool_call>
{"name": "list_applications", "arguments": {"name": "test"}}
</tool_call>`

	calls, cleaned := parseTextToolCalls(content, testRegisteredNames)

	require.Len(t, calls, 1)
	assert.Equal(t, "text-tc-0", calls[0].ID)
	assert.Equal(t, "list_applications", calls[0].Name)
	assert.JSONEq(t, `{"name":"test"}`, calls[0].Arguments)
	assert.Equal(t, "I'll look up the applications.", cleaned)
}

func TestParseTextToolCalls_FunctionCallsInvokeNameFormat(t *testing.T) {
	content := `I'll look up the applications in your enterprise right away.
<function_calls> <invoke name="get_applications"> <parameter name="tenant">acme</parameter> </invoke> </function_calls>`

	calls, cleaned := parseTextToolCalls(content, testRegisteredNames)

	require.Len(t, calls, 1)
	assert.Equal(t, "list_applications", calls[0].Name)
	assert.JSONEq(t, `{"tenant":"acme"}`, calls[0].Arguments)
	assert.Equal(t, "I'll look up the applications in your enterprise right away.", cleaned)
}

func TestParseTextToolCalls_FunctionCallsToolNameFormat(t *testing.T) {
	content := `<function_calls>
<invoke>
<tool_name>list_applications</tool_name>
<parameters>
<workspace_id>123</workspace_id>
</parameters>
</invoke>
</function_calls>`

	calls, cleaned := parseTextToolCalls(content, testRegisteredNames)

	require.Len(t, calls, 1)
	assert.Equal(t, "list_applications", calls[0].Name)
	assert.JSONEq(t, `{"workspace_id":"123"}`, calls[0].Arguments)
	assert.Equal(t, "", cleaned)
}

func TestParseTextToolCalls_FunctionCallsWithHallucinatedResponse(t *testing.T) {
	content := `I'll look that up.
<function_calls> <invoke name="get_applications"> <parameter name="tenant">acme</parameter> </invoke> </function_calls> <function_result> <invoke name="get_applications"> {"applications": [{"name": "FakeApp"}]} </invoke> </function_calls>
Based on the results, you have one application.`

	calls, cleaned := parseTextToolCalls(content, testRegisteredNames)

	require.Len(t, calls, 1)
	assert.Equal(t, "list_applications", calls[0].Name)
	assert.Equal(t, "I'll look that up.", cleaned)
	assert.NotContains(t, cleaned, "FakeApp")
	assert.NotContains(t, cleaned, "function_result")
}

func TestParseTextToolCalls_NoToolCalls(t *testing.T) {
	content := "Here are your applications: App1, App2"

	calls, cleaned := parseTextToolCalls(content, testRegisteredNames)

	assert.Empty(t, calls)
	assert.Equal(t, content, cleaned)
}

func TestParseTextToolCalls_MultipleToolCalls(t *testing.T) {
	content := `Let me check both.
<tool_call>
{"name": "list_applications", "arguments": {}}
</tool_call>
<tool_call>
{"name": "get_portfolio_summary", "arguments": {}}
</tool_call>`

	calls, cleaned := parseTextToolCalls(content, testRegisteredNames)

	require.Len(t, calls, 2)
	assert.Equal(t, "text-tc-0", calls[0].ID)
	assert.Equal(t, "list_applications", calls[0].Name)
	assert.Equal(t, "text-tc-1", calls[1].ID)
	assert.Equal(t, "get_portfolio_summary", calls[1].Name)
	assert.Equal(t, "Let me check both.", cleaned)
}

func TestParseTextToolCalls_MalformedJSON(t *testing.T) {
	content := `<tool_call>not json</tool_call>`

	calls, cleaned := parseTextToolCalls(content, testRegisteredNames)

	assert.Empty(t, calls)
	assert.Equal(t, content, cleaned)
}

func TestParseTextToolCalls_StripsToolResponse(t *testing.T) {
	content := `Preamble.
<tool_call>
{"name": "list_applications", "arguments": {}}
</tool_call>
<tool_response>
{"fake": "data"}
</tool_response>`

	calls, cleaned := parseTextToolCalls(content, testRegisteredNames)

	require.Len(t, calls, 1)
	assert.NotContains(t, cleaned, "tool_response")
	assert.NotContains(t, cleaned, "fake")
}

func TestParseTextToolCalls_EmptyArguments(t *testing.T) {
	content := `<tool_call>{"name": "list_applications", "arguments": {}}</tool_call>`

	calls, _ := parseTextToolCalls(content, testRegisteredNames)

	require.Len(t, calls, 1)
	assert.Equal(t, "{}", calls[0].Arguments)
}

func TestParseTextToolCalls_MissingName(t *testing.T) {
	content := `<tool_call>{"arguments": {"key": "value"}}</tool_call>`

	calls, cleaned := parseTextToolCalls(content, testRegisteredNames)

	assert.Empty(t, calls)
	assert.Equal(t, content, cleaned)
}

func TestResolveToolName_ExactMatch(t *testing.T) {
	assert.Equal(t, "list_applications", resolveToolName("list_applications", testRegisteredNames))
}

func TestResolveToolName_PrefixSwap(t *testing.T) {
	assert.Equal(t, "list_applications", resolveToolName("get_applications", testRegisteredNames))
	assert.Equal(t, "list_applications", resolveToolName("fetch_applications", testRegisteredNames))
	assert.Equal(t, "list_applications", resolveToolName("show_applications", testRegisteredNames))
}

func TestResolveToolName_SubstringMatch(t *testing.T) {
	assert.Equal(t, "search_architecture", resolveToolName("search_architecture_data", testRegisteredNames))
}

func TestResolveToolName_NoMatch(t *testing.T) {
	assert.Equal(t, "completely_unknown", resolveToolName("completely_unknown", testRegisteredNames))
}

func TestResolveToolName_GetDetails(t *testing.T) {
	assert.Equal(t, "get_application_details", resolveToolName("fetch_application_details", testRegisteredNames))
}

func TestParseTextToolCalls_MultipleXMLParams(t *testing.T) {
	content := `<function_calls> <invoke name="search_architecture"> <parameter name="query">billing</parameter> <parameter name="limit">10</parameter> </invoke> </function_calls>`

	calls, _ := parseTextToolCalls(content, testRegisteredNames)

	require.Len(t, calls, 1)
	assert.Equal(t, "search_architecture", calls[0].Name)
	assert.JSONEq(t, `{"query":"billing","limit":"10"}`, calls[0].Arguments)
}

func TestParseTextToolCalls_NoParamsInvoke(t *testing.T) {
	content := `<function_calls> <invoke name="get_portfolio_summary"> </invoke> </function_calls>`

	calls, _ := parseTextToolCalls(content, testRegisteredNames)

	require.Len(t, calls, 1)
	assert.Equal(t, "get_portfolio_summary", calls[0].Name)
	assert.Equal(t, "{}", calls[0].Arguments)
}
