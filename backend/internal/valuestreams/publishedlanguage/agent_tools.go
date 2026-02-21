package publishedlanguage

import (
	pl "easi/backend/internal/archassistant/publishedlanguage"
)

func AgentTools() []pl.AgentToolSpec {
	return []pl.AgentToolSpec{
		{
			Name: "list_value_streams", Description: "List all value streams",
			Access: pl.AccessRead, Permission: "valuestreams:read",
			Method: "GET", Path: "/value-streams",
		},
		{
			Name: "get_value_stream_details", Description: "Get value stream details including stages and mapped capabilities",
			Access: pl.AccessRead, Permission: "valuestreams:read",
			Method: "GET", Path: "/value-streams/{id}",
			PathParams: []pl.ParamSpec{pl.UUIDParam("id", "Value stream ID (UUID)")},
		},
	}
}
