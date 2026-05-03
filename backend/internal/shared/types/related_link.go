package types

import (
	"bytes"
	"encoding/json"
)

type RelatedLink struct {
	Href         string   `json:"href"`
	Methods      []string `json:"methods"`
	Title        string   `json:"title"`
	TargetType   string   `json:"targetType"`
	RelationType string   `json:"relationType"`
}

type RelationEndpoint struct {
	Path   string
	Method string
}

var relationEndpoints = map[string]RelationEndpoint{
	"component-triggers":     {Path: "/api/v1/relations", Method: "POST"},
	"component-serves":       {Path: "/api/v1/relations", Method: "POST"},
	"capability-parent":      {Path: "/api/v1/capabilities/{id}/parent", Method: "PATCH"},
	"capability-realization": {Path: "/api/v1/capabilities/{id}/systems", Method: "POST"},
	"origin-acquired-via":    {Path: "/api/v1/components/{id}/origin/acquired-via", Method: "PUT"},
	"origin-purchased-from":  {Path: "/api/v1/components/{id}/origin/purchased-from", Method: "PUT"},
	"origin-built-by":        {Path: "/api/v1/components/{id}/origin/built-by", Method: "PUT"},
}

func LookupRelationEndpoint(relationType string) (RelationEndpoint, bool) {
	e, ok := relationEndpoints[relationType]
	return e, ok
}

func SpliceXRelated(dtoJSON []byte, related []RelatedLink) ([]byte, error) {
	if len(related) == 0 {
		return dtoJSON, nil
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(dtoJSON, &raw); err != nil {
		return nil, err
	}
	links := map[string]json.RawMessage{}
	if existing, ok := raw["_links"]; ok {
		if err := json.Unmarshal(existing, &links); err != nil {
			return nil, err
		}
	}
	encoded, err := marshalNoEscape(related)
	if err != nil {
		return nil, err
	}
	links["x-related"] = encoded
	mergedLinks, err := marshalNoEscape(links)
	if err != nil {
		return nil, err
	}
	raw["_links"] = mergedLinks
	return marshalNoEscape(raw)
}

func marshalNoEscape(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return bytes.TrimRight(buf.Bytes(), "\n"), nil
}
