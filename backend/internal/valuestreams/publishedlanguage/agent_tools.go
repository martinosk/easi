package publishedlanguage

import (
	pl "easi/backend/internal/archassistant/publishedlanguage"
)

func AgentTools() []pl.AgentToolSpec {
	return []pl.AgentToolSpec{
		{
			Name: "list_value_streams", Description: "List all value streams. Value streams are ordered sequences of stages representing end-to-end business value delivery. Each stage maps to capabilities showing which business functions support that step.",
			Access: pl.AccessRead, Permission: "valuestreams:read",
			Method: "GET", Path: "/value-streams",
		},
		{
			Name: "get_value_stream_details", Description: "Get value stream details including its ordered stages and the capabilities mapped to each stage. Use to trace which capabilities (and their realizing systems) support each step of value delivery.",
			Access: pl.AccessRead, Permission: "valuestreams:read",
			Method: "GET", Path: "/value-streams/{id}",
			PathParams: []pl.ParamSpec{pl.UUIDParam("id", "Value stream ID (UUID)")},
		},
	}
}
