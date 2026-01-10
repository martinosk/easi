package types

type Link struct {
	Href   string `json:"href"`
	Method string `json:"method"`
}

type Links map[string]Link
