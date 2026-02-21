package publishedlanguage

type AccessClass string

const (
	AccessRead   AccessClass = "read"
	AccessWrite  AccessClass = "write"
	AccessCreate AccessClass = "create"
	AccessUpdate AccessClass = "update"
	AccessDelete AccessClass = "delete"
)

func (a AccessClass) IsWrite() bool {
	return a != AccessRead
}

type ParamSpec struct {
	Name        string
	Type        string
	Description string
	Required    bool
}

type AgentToolSpec struct {
	Name        string
	Description string
	Access      AccessClass
	Permission  string
	Method      string
	Path        string
	PathParams  []ParamSpec
	QueryParams []ParamSpec
	BodyParams  []ParamSpec
}

func UUIDParam(name, description string) ParamSpec {
	return ParamSpec{Name: name, Type: "uuid", Description: description, Required: true}
}

func StringParam(name, description string, required bool) ParamSpec {
	return ParamSpec{Name: name, Type: "string", Description: description, Required: required}
}

func IntParam(name, description string) ParamSpec {
	return ParamSpec{Name: name, Type: "integer", Description: description}
}
