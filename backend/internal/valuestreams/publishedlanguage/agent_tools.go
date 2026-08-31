package publishedlanguage

import (
	"easi/backend/internal/shared/agenttools"
)

func AgentTools() []agenttools.AgentToolSpec {
	var specs []agenttools.AgentToolSpec
	specs = append(specs, valueStreamTools()...)
	specs = append(specs, stageTools()...)
	return specs
}

func valueStreamTools() []agenttools.AgentToolSpec {
	return []agenttools.AgentToolSpec{
		{
			Name: "list_value_streams", Description: "List all value streams. Value streams are ordered sequences of stages representing end-to-end business value delivery. Each stage maps to capabilities showing which business functions support that step.",
			Access: agenttools.AccessRead, Permission: "valuestreams:read",
			Method: "GET", Path: "/value-streams",
		},
		{
			Name: "get_value_stream_details", Description: "Get value stream details including its ordered stages and the capabilities mapped to each stage. Use to trace which capabilities (and their realizing systems) support each step of value delivery.",
			Access: agenttools.AccessRead, Permission: "valuestreams:read",
			Method: "GET", Path: "/value-streams/{id}",
			PathParams: []agenttools.ParamSpec{agenttools.UUIDParam("id", "Value stream ID (UUID)")},
		},
		{
			Name: "create_value_stream", Description: "Create a new value stream. Value streams model end-to-end business value delivery. After creation, add stages using create_value_stream_stage.",
			Access: agenttools.AccessCreate, Permission: "valuestreams:write",
			Method: "POST", Path: "/value-streams",
			BodyParams: []agenttools.ParamSpec{
				agenttools.StringParam("name", "Value stream name", true),
				agenttools.StringParam("description", "Value stream description", false),
			},
		},
		{
			Name: "update_value_stream", Description: "Update a value stream's name or description. Does not affect its stages or capability mappings.",
			Access: agenttools.AccessUpdate, Permission: "valuestreams:write",
			Method: "PUT", Path: "/value-streams/{id}",
			PathParams: []agenttools.ParamSpec{agenttools.UUIDParam("id", "Value stream ID (UUID)")},
			BodyParams: []agenttools.ParamSpec{
				agenttools.StringParam("name", "New value stream name", true),
				agenttools.StringParam("description", "New value stream description", false),
			},
		},
		{
			Name: "get_value_stream_capabilities", Description: "Get all capabilities mapped to a value stream's stages. Returns a flat list of all capabilities involved in this value stream, useful for understanding which business functions support this end-to-end flow.",
			Access: agenttools.AccessRead, Permission: "valuestreams:read",
			Method: "GET", Path: "/value-streams/{id}/capabilities",
			PathParams: []agenttools.ParamSpec{agenttools.UUIDParam("id", "Value stream ID (UUID)")},
		},
	}
}

func stageTools() []agenttools.AgentToolSpec {
	return []agenttools.AgentToolSpec{
		{
			Name: "create_value_stream_stage", Description: "Add a new stage to a value stream. Stages are ordered steps in the value delivery process. Optionally specify a position; otherwise the stage is added at the end.",
			Access: agenttools.AccessCreate, Permission: "valuestreams:write",
			Method: "POST", Path: "/value-streams/{id}/stages",
			PathParams: []agenttools.ParamSpec{agenttools.UUIDParam("id", "Value stream ID (UUID)")},
			BodyParams: []agenttools.ParamSpec{
				agenttools.StringParam("name", "Stage name", true),
				agenttools.StringParam("description", "Stage description", false),
			},
		},
		{
			Name: "update_value_stream_stage", Description: "Update a stage's name or description within a value stream. Does not affect its position or capability mappings.",
			Access: agenttools.AccessUpdate, Permission: "valuestreams:write",
			Method: "PUT", Path: "/value-streams/{id}/stages/{stageId}",
			PathParams: []agenttools.ParamSpec{
				agenttools.UUIDParam("id", "Value stream ID (UUID)"),
				agenttools.UUIDParam("stageId", "Stage ID (UUID)"),
			},
			BodyParams: []agenttools.ParamSpec{
				agenttools.StringParam("name", "New stage name", true),
				agenttools.StringParam("description", "New stage description", false),
			},
		},
		{
			Name: "reorder_value_stream_stages", Description: "Reorder stages within a value stream by specifying new positions for each stage. All stages must be included with their new position values.",
			Access: agenttools.AccessUpdate, Permission: "valuestreams:write",
			Method: "PUT", Path: "/value-streams/{id}/stages/positions",
			PathParams: []agenttools.ParamSpec{agenttools.UUIDParam("id", "Value stream ID (UUID)")},
		},
		{
			Name: "add_stage_capability", Description: "Map a business capability to a value stream stage, indicating this capability is needed at this step of value delivery. One stage can have multiple capabilities.",
			Access: agenttools.AccessCreate, Permission: "valuestreams:write",
			Method: "POST", Path: "/value-streams/{id}/stages/{stageId}/capabilities",
			PathParams: []agenttools.ParamSpec{
				agenttools.UUIDParam("id", "Value stream ID (UUID)"),
				agenttools.UUIDParam("stageId", "Stage ID (UUID)"),
			},
			BodyParams: []agenttools.ParamSpec{
				{Name: "capabilityId", Type: "uuid", Description: "Capability ID (UUID) to map to this stage", Required: true},
			},
		},
	}
}
