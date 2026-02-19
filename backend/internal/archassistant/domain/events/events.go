package events

import "time"

type ConversationStarted struct {
	ConversationID string
	TenantID       string
	UserID         string
	OccurredAt     time.Time
}

type UserMessageSent struct {
	MessageID      string
	ConversationID string
	TenantID       string
	UserID         string
	OccurredAt     time.Time
}

type AssistantMessageReceived struct {
	MessageID      string
	ConversationID string
	TenantID       string
	TokensUsed     int
	OccurredAt     time.Time
}
