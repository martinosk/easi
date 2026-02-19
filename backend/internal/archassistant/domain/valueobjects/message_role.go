package valueobjects

import "errors"

var ErrInvalidMessageRole = errors.New("message role must be user, assistant, or tool")

type MessageRole string

const (
	MessageRoleUser      MessageRole = "user"
	MessageRoleAssistant MessageRole = "assistant"
	MessageRoleTool      MessageRole = "tool"
)

func NewMessageRole(s string) (MessageRole, error) {
	role := MessageRole(s)
	if !role.IsValid() {
		return "", ErrInvalidMessageRole
	}
	return role, nil
}

func (r MessageRole) IsValid() bool {
	switch r {
	case MessageRoleUser, MessageRoleAssistant, MessageRoleTool:
		return true
	}
	return false
}

func (r MessageRole) String() string { return string(r) }
