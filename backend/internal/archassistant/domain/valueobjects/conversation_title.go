package valueobjects

import (
	"errors"
	"strings"
	"unicode"
)

const MaxTitleLength = 100

var ErrTitleTooLong = errors.New("title must not exceed 100 characters")

type ConversationTitle struct {
	value string
}

func NewConversationTitle(value string) (ConversationTitle, error) {
	if len(value) > MaxTitleLength {
		return ConversationTitle{}, ErrTitleTooLong
	}
	return ConversationTitle{value: sanitizeTitle(value)}, nil
}

func ReconstructConversationTitle(value string) ConversationTitle {
	return ConversationTitle{value: value}
}

func GenerateTitleFromMessage(firstMessage string) ConversationTitle {
	title := sanitizeTitle(firstMessage)
	if len(title) > MaxTitleLength-3 {
		title = title[:MaxTitleLength-3] + "..."
	}
	return ConversationTitle{value: title}
}

func DefaultTitle() ConversationTitle {
	return ConversationTitle{value: "New conversation"}
}

func (t ConversationTitle) Value() string   { return t.value }
func (t ConversationTitle) IsDefault() bool { return t.value == "New conversation" }

func sanitizeTitle(s string) string {
	return strings.Map(func(r rune) rune {
		if r == '<' || r == '>' {
			return -1
		}
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, s)
}
