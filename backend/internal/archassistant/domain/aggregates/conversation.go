package aggregates

import (
	"time"

	vo "easi/backend/internal/archassistant/domain/valueobjects"

	"github.com/google/uuid"
)

type Message struct {
	id             string
	conversationID string
	role           vo.MessageRole
	content        vo.MessageContent
	tokensUsed     *int
	createdAt      time.Time
}

func (m *Message) ID() string              { return m.id }
func (m *Message) ConversationID() string   { return m.conversationID }
func (m *Message) Role() vo.MessageRole     { return m.role }
func (m *Message) Content() string          { return m.content.Value() }
func (m *Message) TokensUsed() *int         { return m.tokensUsed }
func (m *Message) CreatedAt() time.Time     { return m.createdAt }

type ReconstructMessageParams struct {
	ID             string
	ConversationID string
	Role           vo.MessageRole
	Content        string
	TokensUsed     *int
	CreatedAt      time.Time
}

func ReconstructMessage(p ReconstructMessageParams) *Message {
	return &Message{
		id:             p.ID,
		conversationID: p.ConversationID,
		role:           p.Role,
		content:        vo.ReconstructMessageContent(p.Content),
		tokensUsed:     p.TokensUsed,
		createdAt:      p.CreatedAt,
	}
}

type Conversation struct {
	id            string
	tenantID      string
	userID        string
	title         vo.ConversationTitle
	createdAt     time.Time
	lastMessageAt time.Time
}

func NewConversation(tenantID, userID string) *Conversation {
	now := time.Now()
	return &Conversation{
		id:            uuid.New().String(),
		tenantID:      tenantID,
		userID:        userID,
		title:         vo.DefaultTitle(),
		createdAt:     now,
		lastMessageAt: now,
	}
}

type ReconstructConversationParams struct {
	ID            string
	TenantID      string
	UserID        string
	Title         string
	CreatedAt     time.Time
	LastMessageAt time.Time
}

func ReconstructConversation(p ReconstructConversationParams) *Conversation {
	return &Conversation{
		id:            p.ID,
		tenantID:      p.TenantID,
		userID:        p.UserID,
		title:         vo.ReconstructConversationTitle(p.Title),
		createdAt:     p.CreatedAt,
		lastMessageAt: p.LastMessageAt,
	}
}

func (c *Conversation) ID() string              { return c.id }
func (c *Conversation) TenantID() string         { return c.tenantID }
func (c *Conversation) UserID() string           { return c.userID }
func (c *Conversation) Title() string            { return c.title.Value() }
func (c *Conversation) CreatedAt() time.Time     { return c.createdAt }
func (c *Conversation) LastMessageAt() time.Time { return c.lastMessageAt }

func (c *Conversation) AddUserMessage(content string) (*Message, error) {
	mc, err := vo.NewMessageContent(content)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	c.lastMessageAt = now

	if c.title.IsDefault() {
		c.title = vo.GenerateTitleFromMessage(mc.Value())
	}

	return &Message{
		id:             uuid.New().String(),
		conversationID: c.id,
		role:           vo.MessageRoleUser,
		content:        mc,
		createdAt:      now,
	}, nil
}

func (c *Conversation) AddAssistantMessage(content string, tokensUsed int) *Message {
	now := time.Now()
	c.lastMessageAt = now
	tc := vo.NewTokenCount(tokensUsed)

	return &Message{
		id:             uuid.New().String(),
		conversationID: c.id,
		role:           vo.MessageRoleAssistant,
		content:        vo.ReconstructMessageContent(content),
		tokensUsed:     tc.Pointer(),
		createdAt:      now,
	}
}

func (c *Conversation) IsOwnedBy(userID string) bool {
	return c.userID == userID
}
