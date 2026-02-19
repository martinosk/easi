package valueobjects

import (
	"errors"
	"strings"
	"unicode"
)

const MaxUserMessageLength = 2000

var (
	ErrMessageContentEmpty   = errors.New("message content must not be empty")
	ErrMessageContentTooLong = errors.New("user message must not exceed 2000 characters")
)

type MessageContent struct {
	value string
}

func NewMessageContent(value string) (MessageContent, error) {
	if value == "" {
		return MessageContent{}, ErrMessageContentEmpty
	}
	if len(value) > MaxUserMessageLength {
		return MessageContent{}, ErrMessageContentTooLong
	}
	sanitized := stripControlChars(value)
	return MessageContent{value: sanitized}, nil
}

func ReconstructMessageContent(value string) MessageContent {
	return MessageContent{value: value}
}

func (c MessageContent) Value() string { return c.value }

func stripControlChars(s string) string {
	return strings.Map(func(r rune) rune {
		if isAllowedWhitespace(r) {
			return r
		}
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, s)
}

func isAllowedWhitespace(r rune) bool {
	return r == '\n' || r == '\t' || r == '\r'
}
