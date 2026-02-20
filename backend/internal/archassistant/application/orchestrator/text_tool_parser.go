package orchestrator

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"
)

var (
	toolCallPattern      = regexp.MustCompile(`(?s)<tool_call>\s*(.*?)\s*</tool_call>`)
	functionCallsPattern = regexp.MustCompile(`(?s)<function_calls>\s*(.*?)\s*</function_calls>`)
	invokeNamePattern    = regexp.MustCompile(`(?s)<invoke\s+name="([^"]+)"[^>]*>(.*?)</invoke>`)
	invokeToolPattern    = regexp.MustCompile(`(?s)<invoke>(.*?)</invoke>`)
	toolNameTagPattern   = regexp.MustCompile(`(?s)<tool_name>\s*(.*?)\s*</tool_name>`)
	xmlParamPattern      = regexp.MustCompile(`(?s)<parameter\s+name="([^"]+)"[^>]*>(.*?)</parameter>`)
	namedTagParamPattern = regexp.MustCompile(`(?s)<(\w+)>(.*?)</(\w+)>`)
	nestedParamPattern   = regexp.MustCompile(`(?s)<parameters>(.*?)</parameters>`)

	hallucinationStart = regexp.MustCompile(`<tool_call>|<function_calls>|<tool_response>|<function_result>`)
	hallucinationBlock = regexp.MustCompile(`(?s)<tool_response>.*?</tool_response>|<function_result>.*?(?:</function_result>|</function_calls>|$)`)

	actionPrefixes = []string{
		"get_", "list_", "search_", "find_", "fetch_",
		"query_", "retrieve_", "show_", "read_", "lookup_",
		"create_", "update_", "delete_", "add_", "remove_",
		"set_", "put_", "post_",
	}
)

func parseTextToolCalls(content string, registeredNames []string) ([]ChatToolCall, string) {
	var calls []ChatToolCall

	for _, match := range toolCallPattern.FindAllStringSubmatch(content, -1) {
		tc, ok := parseToolCallJSON([]byte(match[1]))
		if ok {
			calls = append(calls, tc)
		}
	}
	for _, block := range functionCallsPattern.FindAllStringSubmatch(content, -1) {
		calls = append(calls, parseInvokeBlocks(block[1])...)
	}

	if len(calls) == 0 {
		return nil, content
	}

	for i := range calls {
		calls[i].ID = fmt.Sprintf("text-tc-%d", i)
		calls[i].Name = resolveToolName(calls[i].Name, registeredNames)
	}

	return calls, cleanContent(content)
}

func parseInvokeBlocks(blockContent string) []ChatToolCall {
	var calls []ChatToolCall

	for _, match := range invokeNamePattern.FindAllStringSubmatch(blockContent, -1) {
		params := xmlParamPattern.FindAllStringSubmatch(match[2], -1)
		calls = append(calls, ChatToolCall{Name: match[1], Arguments: marshalParams(params)})
	}

	if len(calls) > 0 {
		return calls
	}

	for _, match := range invokeToolPattern.FindAllStringSubmatch(blockContent, -1) {
		nameMatch := toolNameTagPattern.FindStringSubmatch(match[1])
		if nameMatch == nil {
			continue
		}
		tags := extractNestedTags(match[1])
		calls = append(calls, ChatToolCall{Name: nameMatch[1], Arguments: marshalParams(tags)})
	}

	return calls
}

func marshalParams(matches [][]string) string {
	if len(matches) == 0 {
		return "{}"
	}
	m := make(map[string]interface{}, len(matches))
	for _, p := range matches {
		m[p[1]] = strings.TrimSpace(p[2])
	}
	data, _ := json.Marshal(m)
	return string(data)
}

func extractNestedTags(invokeBody string) [][]string {
	paramBlock := nestedParamPattern.FindStringSubmatch(invokeBody)
	if paramBlock == nil {
		return nil
	}
	var result [][]string
	for _, t := range namedTagParamPattern.FindAllStringSubmatch(paramBlock[1], -1) {
		if t[1] == t[3] && t[1] != "parameters" {
			result = append(result, t)
		}
	}
	return result
}

func parseToolCallJSON(raw []byte) (ChatToolCall, bool) {
	var wrapper struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	}
	if err := json.Unmarshal(raw, &wrapper); err != nil {
		return ChatToolCall{}, false
	}
	if wrapper.Name == "" {
		return ChatToolCall{}, false
	}
	return ChatToolCall{Name: wrapper.Name, Arguments: normalizeArguments(wrapper.Arguments)}, true
}

func normalizeArguments(raw json.RawMessage) string {
	if len(raw) == 0 {
		return "{}"
	}
	var obj map[string]interface{}
	if err := json.Unmarshal(raw, &obj); err != nil {
		return "{}"
	}
	data, err := json.Marshal(obj)
	if err != nil {
		return "{}"
	}
	return string(data)
}

func resolveToolName(hallucinated string, registered []string) string {
	entity := stripActionPrefix(hallucinated)
	strategies := []func(string) bool{
		func(r string) bool { return r == hallucinated },
		func(r string) bool { return stripActionPrefix(r) == entity },
		func(r string) bool {
			return strings.Contains(r, entity) || strings.Contains(entity, stripActionPrefix(r))
		},
	}
	for _, matches := range strategies {
		for _, r := range registered {
			if matches(r) {
				return r
			}
		}
	}
	return hallucinated
}

func stripActionPrefix(name string) string {
	for _, prefix := range actionPrefixes {
		if strings.HasPrefix(name, prefix) {
			return name[len(prefix):]
		}
	}
	return name
}

func cleanContent(content string) string {
	loc := hallucinationStart.FindStringIndex(content)
	if loc == nil {
		return content
	}
	preamble := content[:loc[0]]
	preamble = hallucinationBlock.ReplaceAllString(preamble, "")
	return strings.TrimSpace(preamble)
}

func applyTextToolCallFallback(result *agentStreamResult, registeredNames []string) {
	if len(result.toolCalls) > 0 {
		log.Printf("[archassistant] LLM returned %d native tool calls, skipping text fallback", len(result.toolCalls))
		return
	}
	textToolCalls, cleanedContent := parseTextToolCalls(result.content, registeredNames)
	if len(textToolCalls) > 0 {
		log.Printf("[archassistant] text fallback parsed %d tool calls from LLM text output", len(textToolCalls))
		for _, tc := range textToolCalls {
			log.Printf("[archassistant]   -> tool: %s, args: %s", tc.Name, tc.Arguments)
		}
		result.toolCalls = textToolCalls
		result.content = cleanedContent
	} else {
		log.Printf("[archassistant] no tool calls found (native or text), returning LLM response as-is")
	}
}
